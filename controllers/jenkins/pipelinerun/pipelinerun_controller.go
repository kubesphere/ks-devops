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
	"errors"
	"github.com/go-logr/logr"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"net/http"
	"net/url"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

// Reconciler reconciles a PipelineRun object
type Reconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	DevOpsClient devopsClient.Interface
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

	// get pipeline
	pipelineRef := v1.GetControllerOf(&pr)

	if pipelineRef == nil {
		log.Error(nil, "skipped to reconcile this PipelineRun due to not found pipeline reference in owner references of PipelineRun.")
		return ctrl.Result{}, nil
	}

	var pipeline v1alpha3.Pipeline
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: pr.GetNamespace(), Name: pipelineRef.Name}, &pipeline); err != nil {
		log.Error(err, "unable to get pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// set Pipeline name to the PipelineRun labels
	if pr.Labels == nil {
		pr.Labels = make(map[string]string)
	}
	pr.Labels[devopsv1alpha4.JenkinsPipelineRunPipelineKey] = pipeline.GetName()

	var projectName = pipeline.Namespace

	// check PipelineRun status
	if pr.HasStarted() {
		log.V(5).Info("pipeline has already started, and we are retrieving run data from Jenkins.")

		runResult, err := r.getPipelineRunResult(projectName, pipeline.GetName(), &pr)
		if err != nil {
			log.Error(err, "unable get PipelineRun data.")
			return ctrl.Result{}, err
		}

		// set latest run result into annotations
		runResultJSON, err := json.Marshal(runResult)
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
			return ctrl.Result{}, err
		}

		if runResult.EndTime != "" {
			// this means the PipelineRun has ended
			if err := r.completePipelineRun(&pr, runResult); err != nil {
				return ctrl.Result{}, err
			}
			// update PipelineRun
			if err := r.Status().Update(ctx, &pr); err != nil {
				log.Error(err, "unable to update PipelineRun status.")
				return ctrl.Result{}, err
			}
			log.Info("PipelineRun has been finished")
			return ctrl.Result{}, nil
		}

		// update updateTime
		now := v1.Now()
		pr.Status.UpdateTime = &now

		if err := r.Status().Update(ctx, &pr); err != nil {
			log.Error(err, "unable to update PipelineRun status.")
			return ctrl.Result{}, err
		}

		// until the status is okay
		// TODO make the RequeueAfter configurable
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// first run
	parameters := pr.Spec.Parameters
	if parameters == nil {
		parameters = make([]devopsv1alpha4.Parameter, 0)
	}

	// build http parameters
	httpParameters, err := buildHttpParametersForRunning(&pr)
	if err != nil {
		log.Error(err, "unable to create http parameters for running pipeline.")
		return ctrl.Result{}, err
	}
	// run pipeline
	var runResponse *devopsClient.RunPipeline
	if pr.IsMultiBranchPipeline() {
		// run multi branch pipeline
		runResponse, err = r.DevOpsClient.RunBranchPipeline(projectName, pipeline.GetName(), pr.Spec.SCM.RefName, httpParameters)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// run normal pipeline
		runResponse, err = r.DevOpsClient.RunPipeline(projectName, pipeline.GetName(), httpParameters)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// set Jenkins PipelineRun id
	if pr.Annotations == nil {
		pr.Annotations = make(map[string]string)
	}
	pr.Annotations[devopsv1alpha4.JenkinsPipelineRunIdKey] = runResponse.ID

	// label PipelineRun as running
	pr.Status.MarkPending()

	// The Update method only updates fields except status
	if err := r.Client.Update(ctx, &pr); err != nil {
		log.Error(err, "unable to update PipelineRun.")
		return ctrl.Result{}, err
	}
	if err := r.Client.Status().Update(ctx, &pr); err != nil {
		log.Error(err, "unable to update PipelineRun status.")
		return ctrl.Result{}, err
	}

	// requeue immediately
	return ctrl.Result{Requeue: true}, nil
}

func (r *Reconciler) getPipelineRunResult(projectName, pipelineName string, pr *devopsv1alpha4.PipelineRun) (runResult *devopsClient.PipelineRun, err error) {
	// get PipelineRun id
	runId, _ := pr.GetPipelineRunId()
	// get latest runs data from Jenkins
	if pr.IsMultiBranchPipeline() {
		runResult, err = r.DevOpsClient.GetBranchPipelineRun(projectName, pipelineName, pr.Spec.SCM.RefName, runId, &devopsClient.HttpParameters{
			Url:    getStubUrl(),
			Method: http.MethodGet,
		})
	} else {
		runResult, err = r.DevOpsClient.GetPipelineRun(projectName, pipelineName, runId, &devopsClient.HttpParameters{
			Url:    getStubUrl(),
			Method: http.MethodGet,
		})
	}
	return
}

func (r *Reconciler) completePipelineRun(pr *devopsv1alpha4.PipelineRun, jenkinsRunResult *devopsClient.PipelineRun) error {
	// time sample: 2021-08-18T16:36:47.236+0000
	endTime, err := time.Parse(time.RFC3339, jenkinsRunResult.EndTime)
	if err != nil {
		return err
	}

	pr.Status.MarkCompleted(endTime)

	// resolve status
	jenkinsRunData := JenkinsRunData{jenkinsRunResult}
	if err := jenkinsRunData.resolveStatus(pr); err != nil {
		return err
	}
	return nil
}

func buildHttpParametersForRunning(pr *devopsv1alpha4.PipelineRun) (*devopsClient.HttpParameters, error) {
	if pr == nil {
		return nil, errors.New("invalid PipelineRun")
	}
	// first run
	parameters := pr.Spec.Parameters
	if parameters == nil {
		parameters = make([]devopsv1alpha4.Parameter, 0)
	}

	// build http parameters
	var body = map[string]interface{}{
		"parameters": parameters,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &devopsClient.HttpParameters{
		Url:    getStubUrl(),
		Method: http.MethodPost,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(string(bodyJSON))),
	}, nil
}

func getStubUrl() *url.URL {
	// this url isn't meaningful.
	const stubDevOpsHost = "https://devops.kubesphere.io/"
	stubUrl, err := url.Parse(stubDevOpsHost)
	if err != nil {
		// should never happen
		panic("invalid stub url: " + stubDevOpsHost)
	}
	return stubUrl
}

type JenkinsRunState string

const (
	Queued        JenkinsRunState = "QUEUED"
	Running       JenkinsRunState = "RUNNING"
	Paused        JenkinsRunState = "PAUSED"
	Skipped       JenkinsRunState = "SKIPPED"
	NotBuiltState JenkinsRunState = "NOT_BUILT"
	Finished      JenkinsRunState = "FINISHED"
)

type JenkinsRunResult string

const (
	Success        JenkinsRunResult = "SUCCESS"
	Unstable       JenkinsRunResult = "UNSTABLE"
	Failure        JenkinsRunResult = "FAILURE"
	NotBuiltResult JenkinsRunResult = "NOT_BUILT"
	Unknown        JenkinsRunResult = "UNKNOWN"
	Aborted        JenkinsRunResult = "ABORTED"
)

type JenkinsRunData struct {
	*devopsClient.PipelineRun
}

func (jrs *JenkinsRunData) resolveStatus(pr *devopsv1alpha4.PipelineRun) error {
	// TODO Complete the conversion
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1alpha4.PipelineRun{}).
		Complete(r)
}
