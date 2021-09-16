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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	prv1alpha3 "kubesphere.io/devops/pkg/api/devops/pipelinerun/v1alpha3"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// Reconciler reconciles a PipelineRun object
type Reconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	DevOpsClient devopsClient.Interface
	JenkinsCore  core.JenkinsCore
	recorder     record.EventRecorder
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("PipelineRun", req.NamespacedName)

	// get PipelineRun
	var pr prv1alpha3.PipelineRun
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

	namespaceName := pipeline.Namespace
	pipelineName := pipeline.GetName()

	// set Pipeline name and SCM ref name into labels
	if pr.Labels == nil {
		pr.Labels = make(map[string]string)
	}
	pr.Labels[prv1alpha3.PipelineNameLabelKey] = pipelineName
	if refName, err := getSCMRefName(&pr.Spec); err == nil && refName != "" {
		pr.Labels[prv1alpha3.SCMRefNameLabelKey] = refName
	}

	log = log.WithValues("namespace", namespaceName, "Pipeline", pipelineName)

	// check PipelineRun status
	if pr.HasStarted() {
		log.V(5).Info("pipeline has already started, and we are retrieving run data from Jenkins.")
		pipelineBuild, err := r.getPipelineRunResult(namespaceName, pipelineName, &pr)
		if err != nil {
			log.Error(err, "unable get PipelineRun data.")
			r.recorder.Eventf(&pr, corev1.EventTypeWarning, prv1alpha3.RetrieveFailed, "Failed to retrieve running data from Jenkins, and error was %s", err)
			return ctrl.Result{}, err
		}

		prNodes, err := r.getPipelineNodes(namespaceName, pipelineName, &pr)
		if err != nil {
			log.Error(err, "unable to get PipelineRun nodes detail")
			r.recorder.Eventf(&pr, corev1.EventTypeWarning, prv1alpha3.RetrieveFailed, "Failed to retrieve nodes detail from Jenkins, and error was %s", err)
			return ctrl.Result{}, err
		}

		// set the latest run result into annotations
		runResultJSON, err := json.Marshal(pipelineBuild)
		if err != nil {
			return ctrl.Result{}, err
		}
		prNodesJSON, err := json.Marshal(prNodes)
		if err != nil {
			return ctrl.Result{}, err
		}
		if pr.Annotations == nil {
			pr.Annotations = make(map[string]string)
		}
		pr.Annotations[prv1alpha3.JenkinsPipelineRunStatusKey] = string(runResultJSON)
		pr.Annotations[prv1alpha3.JenkinsPipelineRunStagesStatusKey] = string(prNodesJSON)
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
		r.recorder.Eventf(&pr, corev1.EventTypeNormal, prv1alpha3.Updated, "Updated running data for PipelineRun %s", req.NamespacedName)
		// until the status is okay
		// TODO make the RequeueAfter configurable
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// first run
	pipelineBuild, err := r.triggerJenkinsJob(namespaceName, pipelineName, &pr.Spec)
	if err != nil {
		log.Error(err, "unable to run pipeline", "namespace", namespaceName, "pipeline", pipeline.Name)
		r.recorder.Eventf(&pr, corev1.EventTypeWarning, prv1alpha3.TriggerFailed, "Failed to trigger PipelineRun %s, and error was %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	}
	log.Info("Triggered a PipelineRun", "runID", pipelineBuild.ID)

	// set Jenkins run ID
	if pr.Annotations == nil {
		pr.Annotations = make(map[string]string)
	}
	pr.Annotations[prv1alpha3.JenkinsPipelineRunIDKey] = pipelineBuild.ID

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
	r.recorder.Eventf(&pr, corev1.EventTypeNormal, prv1alpha3.Started, "Started PipelineRun %s", req.NamespacedName)
	// requeue after 1 second
	return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
}

func (r *Reconciler) triggerJenkinsJob(devopsProjectName, pipelineName string, prSpec *prv1alpha3.PipelineRunSpec) (*job.PipelineBuild, error) {
	c := job.BlueOceanClient{JenkinsCore: r.JenkinsCore, Organization: "jenkins"}

	branch, err := getSCMRefName(prSpec)
	if err != nil {
		return nil, err
	}

	return c.Build(job.BuildOption{
		Pipelines:  []string{devopsProjectName, pipelineName},
		Parameters: parameterConverter{parameters: prSpec.Parameters}.convert(),
		Branch:     branch,
	})
}

func getSCMRefName(prSpec *prv1alpha3.PipelineRunSpec) (string, error) {
	var branch = ""
	if prSpec.IsMultiBranchPipeline() {
		if prSpec.SCM == nil || prSpec.SCM.RefName == "" {
			return "", fmt.Errorf("failed to obtain SCM reference name for multi-branch Pipeline")
		}
		branch = prSpec.SCM.RefName
	}
	return branch, nil
}

func (r *Reconciler) getPipelineRunResult(devopsProjectName, pipelineName string, pr *prv1alpha3.PipelineRun) (*job.PipelineBuild, error) {
	runID, exists := pr.GetPipelineRunID()
	if !exists {
		return nil, fmt.Errorf("unable to get PipelineRun result due to not found run ID")
	}
	c := job.BlueOceanClient{JenkinsCore: r.JenkinsCore, Organization: "jenkins"}

	branch, err := getSCMRefName(&pr.Spec)
	if err != nil {
		return nil, err
	}
	return c.GetBuild(job.GetBuildOption{
		RunID:     runID,
		Pipelines: []string{devopsProjectName, pipelineName},
		Branch:    branch,
	})
}

func (r *Reconciler) getPipelineNodes(devopsProjectName, pipelineName string, pr *prv1alpha3.PipelineRun) ([]job.Node, error) {
	runID, exists := pr.GetPipelineRunID()
	if !exists {
		return nil, fmt.Errorf("unable to get PipelineRun result due to not found run ID")
	}
	c := job.BlueOceanClient{JenkinsCore: r.JenkinsCore, Organization: "jenkins"}
	branch, err := getSCMRefName(&pr.Spec)
	if err != nil {
		return nil, err
	}
	return c.GetNodes(job.GetNodesOption{
		Pipelines: []string{devopsProjectName, pipelineName},
		Branch:    branch,
		RunID:     runID,
	})
}

func (r *Reconciler) updateLabelsAndAnnotations(ctx context.Context, pr *prv1alpha3.PipelineRun) error {
	// get pipeline
	prToUpdate := prv1alpha3.PipelineRun{}
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

func (r *Reconciler) updateStatus(ctx context.Context, desiredStatus *prv1alpha3.PipelineRunStatus, prKey client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		prToUpdate := prv1alpha3.PipelineRun{}
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

func (r *Reconciler) makePipelineRunOrphan(ctx context.Context, pr *prv1alpha3.PipelineRun) error {
	// make the PipelineRun as orphan
	prToUpdate := pr.DeepCopy()
	prToUpdate.LabelAsAnOrphan()
	if err := r.updateLabelsAndAnnotations(ctx, prToUpdate); err != nil {
		return err
	}
	condition := prv1alpha3.Condition{
		Type:               prv1alpha3.ConditionSucceeded,
		Status:             prv1alpha3.ConditionUnknown,
		Reason:             "SKIPPED",
		Message:            "skipped to reconcile this PipelineRun due to not found Pipeline reference in PipelineRun.",
		LastTransitionTime: v1.Now(),
		LastProbeTime:      v1.Now(),
	}
	prToUpdate.Status.AddCondition(&condition)
	prToUpdate.Status.Phase = prv1alpha3.Unknown
	return r.updateStatus(ctx, &pr.Status, client.ObjectKey{Namespace: pr.Namespace, Name: pr.Name})
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// the name should obey Kubernetes naming convention: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	r.recorder = mgr.GetEventRecorderFor("pipelinerun-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&prv1alpha3.PipelineRun{}).
		Complete(r)
}
