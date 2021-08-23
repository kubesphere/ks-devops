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
	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
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

	if pr.Spec.PipelineRef == nil || pr.Spec.PipelineRef.Name == "" {
		// ignore this PipelineRun
		log.Info("skipped to reconcile this PipelineRun due to not found Pipeline reference in PipelineRun.")
		return ctrl.Result{}, nil
	}

	var projectName = pr.Labels[devopsv1alpha4.JenkinsPipelineRunDevOpsProjectKey]
	var pipelineName = pr.Labels[devopsv1alpha4.JenkinsPipelineRunPipelineKey]

	if pr.HasCompleted() {
		// Pipeline has been completed
		return ctrl.Result{}, nil
	}

	// check PipelineRun status
	if pr.HasStarted() {
		log.V(5).Info("pipeline has already started, and we are retrieving run data from Jenkins.")

		buildResult, err := r.getPipelineRunResult(projectName, pipelineName, &pr)
		if err != nil {
			log.Error(err, "unable get PipelineRun data.")
			return ctrl.Result{}, err
		}

		// set the latest run result into annotations
		runResultJSON, err := json.Marshal(buildResult)
		if err != nil {
			return ctrl.Result{}, err
		}
		if pr.Annotations == nil {
			pr.Annotations = make(map[string]string)
		}
		pr.Annotations[devopsv1alpha4.JenkinsPipelineRunDataKey] = string(runResultJSON)
		// update PipelineRun
		if err := r.Update(ctx, &pr); err != nil {
			log.Error(err, "unable to update PipelineRun.")
			return ctrl.Result{RequeueAfter: time.Second}, err
		}

		if err := r.apply(buildResult, &pr.Status); err != nil {
			log.Error(err, "unable to apply Jenkins run result to PipelineRun", "jenkinsRunData", buildResult)
			return ctrl.Result{}, err
		}

		// Because the status is a subresource of PipelineRun, we have to update status separately.
		// See also: https://book-v1.book.kubebuilder.io/basics/status_subresource.html
		if err := r.Status().Update(ctx, &pr); err != nil {
			log.Error(err, "unable to update PipelineRun status.")
			return ctrl.Result{RequeueAfter: time.Second}, err
		}
		// until the status is okay
		// TODO make the RequeueAfter configurable
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// first run
	// get pipeline
	pipelineNamespace := pr.Spec.PipelineRef.Namespace
	if pipelineNamespace == "" {
		pipelineNamespace = pr.Namespace
	}
	var pipeline v1alpha3.Pipeline
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: pipelineNamespace, Name: pr.Spec.PipelineRef.Name}, &pipeline); err != nil {
		log.Error(err, "unable to get pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	projectName = pipeline.Namespace
	pipelineName = pipeline.GetName()

	// set Pipeline name to the PipelineRun labels
	if pr.Labels == nil {
		pr.Labels = make(map[string]string)
	}

	pr.Labels[devopsv1alpha4.JenkinsPipelineRunPipelineKey] = pipelineName
	pr.Labels[devopsv1alpha4.JenkinsPipelineRunDevOpsProjectKey] = projectName

	jobBuild, err := r.RunPipeline(projectName, pipelineName, &pr.Spec)
	if err != nil {
		log.Error(err, "unable to run pipeline", "projectName", projectName, "pipeline", pipeline.Name)
		return ctrl.Result{}, err
	}

	// set Jenkins PipelineRun id
	if pr.Annotations == nil {
		pr.Annotations = make(map[string]string)
	}
	pr.Annotations[devopsv1alpha4.JenkinsPipelineRunIDKey] = jobBuild.ID
	pr.Status.StartTime = &v1.Time{Time: time.Now()}

	// the Update method only updates fields except status
	if err := r.Client.Update(ctx, &pr); err != nil {
		log.Error(err, "unable to update PipelineRun.")
		return ctrl.Result{}, err
	}
	// due to the status is subresource of PipelineRun, we have to update status separately.
	// see also: https://book-v1.book.kubebuilder.io/basics/status_subresource.html
	if err := r.Client.Status().Update(ctx, &pr); err != nil {
		log.Error(err, "unable to update PipelineRun status.")
		return ctrl.Result{}, err
	}

	// requeue immediately
	return ctrl.Result{Requeue: true}, nil
}

func (r *Reconciler) RunPipeline(projectName, pipelineName string, prSpec *devopsv1alpha4.PipelineRunSpec) (*job.Build, error) {
	jobName := generateJobName(projectName, pipelineName, prSpec)
	jobClient := job.Client{JenkinsCore: r.JenkinsCore} // run normal pipeline
	//*runResult, err = jobClient.BuildAndReturn(fmt.Sprintf("%s %s", projectName, pipelineName), "Triggered by PipelineRun reconciler", 0, 0)
	// build pipeline with params
	err := jobClient.BuildWithParams(jobName, convertParam(prSpec.Parameters))
	if err != nil {
		return nil, err
	}
	// get latest build
	identityBuild, err := jobClient.GetBuild(jobName, -1)
	if err != nil {
		return nil, err
	}
	return identityBuild, nil
}

func (r *Reconciler) getPipelineRunResult(projectName, pipelineName string, pr *devopsv1alpha4.PipelineRun) (runResult *job.Build, err error) {
	runIDStr, _ := pr.GetPipelineRunID()
	runID, err := strconv.Atoi(runIDStr)
	if err != nil {
		return nil, err
	}
	jobName := generateJobName(projectName, pipelineName, &pr.Spec)
	jobClient := job.Client{JenkinsCore: r.JenkinsCore}
	return jobClient.GetBuild(jobName, runID)
}

func (r *Reconciler) apply(jobBuild *job.Build, prStatus *devopsv1alpha4.PipelineRunStatus) error {
	condition := devopsv1alpha4.Condition{
		Type:          devopsv1alpha4.ConditionReady,
		LastProbeTime: v1.Now(),
		Reason:        jobBuild.Result,
	}

	var phase = devopsv1alpha4.Unknown

	if jobBuild.Building {
		// when pipeline is building
		condition.Status = devopsv1alpha4.ConditionUnknown
		phase = devopsv1alpha4.Running
	} else {
		// when pipeline is not building
		switch jobBuild.Result {
		case Success.String():
			condition.Type = devopsv1alpha4.ConditionSucceeded
			condition.Status = devopsv1alpha4.ConditionTrue
			phase = devopsv1alpha4.Succeeded
		case Unstable.String():
			condition.Status = devopsv1alpha4.ConditionFalse
			phase = devopsv1alpha4.Failed
		case Failure.String():
			condition.Status = devopsv1alpha4.ConditionFalse
			phase = devopsv1alpha4.Failed
		case NotBuiltResult.String():
			condition.Status = devopsv1alpha4.ConditionUnknown
			phase = devopsv1alpha4.Unknown
		case Aborted.String():
			condition.Status = devopsv1alpha4.ConditionFalse
			phase = devopsv1alpha4.Failed
		default:
			condition.Status = devopsv1alpha4.ConditionUnknown
		}

		// set completion time
		completionTime := time.UnixMilli(jobBuild.Timestamp + jobBuild.Duration)
		prStatus.CompletionTime = &v1.Time{Time: completionTime}
	}

	prStatus.Phase = phase
	prStatus.AddCondition(&condition)
	prStatus.UpdateTime = &v1.Time{Time: time.Now()}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1alpha4.PipelineRun{}).
		Complete(r)
}
