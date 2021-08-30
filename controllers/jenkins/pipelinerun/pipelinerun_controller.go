/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipelinerun

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	time "time"
)

// Reconciler reconciles a PipelineRun object
type Reconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	DevOpsClient devopsClient.Interface
	JenkinsCore  core.JenkinsCore
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("PipelineRun", req.NamespacedName)

	// get PipelineRun
	var pr devopsv1alpha4.PipelineRun
	if err := r.Client.Get(ctx, req.NamespacedName, &pr); err != nil {
		log.Error(err, "unable to fetch PipelineRun")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// the PipelineRun cannot allow building
	if !pr.Buildable() {
		return ctrl.Result{}, nil
	}

	// don't modify the cache in other places, like informer cache.
	pr = *pr.DeepCopy()

	// check PipelineRef
	if pr.Spec.PipelineRef == nil || pr.Spec.PipelineRef.Name == "" {
		// make the PipelineRun as orphan
		return ctrl.Result{}, r.makePipelineRunOrphan(ctx, &pr)
	}

	// get pipeline
	var pipeline v1alpha3.Pipeline
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: pr.Namespace, Name: pr.Spec.PipelineRef.Name}, &pipeline); err != nil {
		log.Error(err, "unable to get pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	devopsProjectName := pipeline.Namespace
	pipelineName := pipeline.GetName()

	log = log.WithValues("DevOpsProject", devopsProjectName, "Pipeline", pipelineName)

	// check PipelineRun status
	if pr.HasStarted() {
		log.V(5).Info("pipeline has already started, and we are retrieving run data from Jenkins.")

		pipelineBuild, err := r.getPipelineRunResult(devopsProjectName, pipelineName, &pr)
		if err != nil {
			log.Error(err, "unable get PipelineRun data.")
			return ctrl.Result{}, err
		}

		// set the latest run result into annotations
		runResultJSON, err := json.Marshal(pipelineBuild)
		if err != nil {
			return ctrl.Result{}, err
		}
		if pr.Annotations == nil {
			pr.Annotations = make(map[string]string)
		}
		pr.Annotations[devopsv1alpha4.JenkinsPipelineRunDataKey] = string(runResultJSON)
		// update PipelineRun
		if err := r.updateLabelsAndAnnotations(ctx, &pr); err != nil {
			log.Error(err, "unable to update PipelineRun labels and annotations.")
			return ctrl.Result{RequeueAfter: time.Second}, err
		}

		status := pr.Status.DeepCopy()
		pbApplier := pipelineBuildApplier{pipelineBuild}
		pbApplier.apply(status)

		// Because the status is a subresource of PipelineRun, we have to update status separately.
		// See also: https://book-v1.book.kubebuilder.io/basics/status_subresource.html
		if err := r.updateStatus(ctx, status, req.NamespacedName); err != nil {
			log.Error(err, "unable to update PipelineRun status.")
			return ctrl.Result{RequeueAfter: time.Second}, err
		}
		// until the status is okay
		// TODO make the RequeueAfter configurable
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// first run
	pipelineBuild, err := r.triggerJenkinsJob(devopsProjectName, pipelineName, &pr.Spec)
	if err != nil {
		log.Error(err, "unable to run pipeline", "devopsProjectName", devopsProjectName, "pipeline", pipeline.Name)
		return ctrl.Result{}, err
	}
	log.Info("Triggered a PipelineRun", "runID", pipelineBuild.ID)

	// set Jenkins run ID
	if pr.Annotations == nil {
		pr.Annotations = make(map[string]string)
	}
	pr.Annotations[devopsv1alpha4.JenkinsPipelineRunIDKey] = pipelineBuild.ID

	// set Pipeline name to the PipelineRun labels
	if pr.Labels == nil {
		pr.Labels = make(map[string]string)
	}

	// the Update method only updates fields except subresource: status
	if err := r.updateLabelsAndAnnotations(ctx, &pr); err != nil {
		log.Error(err, "unable to update PipelineRun labels and annotations.")
		return ctrl.Result{}, err
	}

	pr.Status.StartTime = &v1.Time{Time: time.Now()}
	pr.Status.UpdateTime = &v1.Time{Time: time.Now()}
	// due to the status is subresource of PipelineRun, we have to update status separately.
	// see also: https://book-v1.book.kubebuilder.io/basics/status_subresource.html

	if err := r.updateStatus(ctx, &pr.Status, req.NamespacedName); err != nil {
		log.Error(err, "unable to update PipelineRun status.")
		return ctrl.Result{}, err
	}

	// requeue immediately
	return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
}

func (r *Reconciler) triggerJenkinsJob(devopsProjectName, pipelineName string, prSpec *devopsv1alpha4.PipelineRunSpec) (*job.PipelineBuild, error) {
	boClient := job.BlueOceanClient{JenkinsCore: r.JenkinsCore, Organization: "jenkins"}
	return boClient.Build(job.BuildOption{
		Pipelines:  []string{devopsProjectName, pipelineName},
		Parameters: parameterConverter{parameters: prSpec.Parameters}.convert(),
	})
}

func (r *Reconciler) getPipelineRunResult(projectName, pipelineName string, pr *devopsv1alpha4.PipelineRun) (*job.PipelineBuild, error) {
	runIDStr, exists := pr.GetPipelineRunID()
	if !exists {
		return nil, fmt.Errorf("unable to get PipelineRun result due to not found run ID")
	}
	boClient := job.BlueOceanClient{JenkinsCore: r.JenkinsCore, Organization: "jenkins"}
	return boClient.GetBuild(runIDStr, projectName, pipelineName)
}

func (r *Reconciler) updateLabelsAndAnnotations(ctx context.Context, pr *devopsv1alpha4.PipelineRun) error {
	// get pipeline
	prToUpdate := devopsv1alpha4.PipelineRun{}
	err := r.Get(ctx, client.ObjectKey{Namespace: pr.Namespace, Name: pr.Name}, &prToUpdate)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(pr.Labels, prToUpdate.Labels) && reflect.DeepEqual(pr.Annotations, prToUpdate.Annotations) {
		return nil
	}
	prToUpdate = *prToUpdate.DeepCopy()
	prToUpdate.Labels = pr.Labels
	prToUpdate.Annotations = pr.Annotations
	return r.Update(ctx, &prToUpdate)
}

func (r *Reconciler) updateStatus(ctx context.Context, desiredStatus *devopsv1alpha4.PipelineRunStatus, prKey client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		prToUpdate := devopsv1alpha4.PipelineRun{}
		err := r.Get(ctx, prKey, &prToUpdate)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(*desiredStatus, prToUpdate.Status) {
			return nil
		}
		prToUpdate = *prToUpdate.DeepCopy()
		prToUpdate.Status = *desiredStatus
		return r.Status().Update(ctx, &prToUpdate)
	})
}

func (r *Reconciler) makePipelineRunOrphan(ctx context.Context, pr *devopsv1alpha4.PipelineRun) error {
	// make the PipelineRun as orphan
	prToUpdate := pr.DeepCopy()
	prToUpdate.LabelAsAnOrphan()
	if err := r.updateLabelsAndAnnotations(ctx, prToUpdate); err != nil {
		return err
	}
	condition := devopsv1alpha4.Condition{
		Type:               devopsv1alpha4.ConditionSucceeded,
		Status:             devopsv1alpha4.ConditionUnknown,
		Reason:             "SKIPPED",
		Message:            "skipped to reconcile this PipelineRun due to not found Pipeline reference in PipelineRun.",
		LastTransitionTime: v1.Now(),
		LastProbeTime:      v1.Now(),
	}
	prToUpdate.Status.AddCondition(&condition)
	prToUpdate.Status.Phase = devopsv1alpha4.Unknown
	return r.updateStatus(ctx, &pr.Status, client.ObjectKey{Namespace: pr.Namespace, Name: pr.Name})
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1alpha4.PipelineRun{}).
		Complete(r)
}
