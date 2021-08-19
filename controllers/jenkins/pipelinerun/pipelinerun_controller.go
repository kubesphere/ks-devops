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
	log := r.Log.WithValues("pipelinerun", req.NamespacedName)

	// get pipeline run
	var pr devopsv1alpha4.PipelineRun
	if err := r.Client.Get(ctx, req.NamespacedName, &pr); err != nil {
		log.Error(err, "unable to fetch PipelineRun")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// get pipeline
	pipelineRef := v1.GetControllerOf(&pr)

	if pipelineRef == nil {
		log.Error(nil, "skipped to reconcile this pipeline run due to not found pipeline reference in pipelinerun owner references.")
		return ctrl.Result{}, nil
	}

	var pipeline v1alpha3.Pipeline
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: pr.GetNamespace(), Name: pipelineRef.Name}, &pipeline); err != nil {
		log.Error(err, "unable to get pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var projectName = pipeline.Namespace

	// check pipeline run status
	if pr.HasStarted() {
		log.V(5).Info("pipeline has already started, and we are retrieving pipeline run data from Jenkins.")

		runResult, err := r.getPipelineRunResult(projectName, pipeline.GetName(), &pr)
		if err != nil {
			log.Error(err, "unable get pipeline run data.")
			return ctrl.Result{}, err
		}

		// set latest run result into annotations
		if pr.Annotations == nil {
			pr.Annotations = make(map[string]string)
		}
		runResultJSON, err := json.Marshal(runResult)
		if err != nil {
			return ctrl.Result{}, err
		}
		pr.Annotations[devopsv1alpha4.JenkinsPipelineRunDataKey] = string(runResultJSON)

		// update pipeline run
		if err := r.Update(ctx, &pr); err != nil {
			log.Error(err, "unable to update pipeline run.")
			return ctrl.Result{}, err
		}

		// TODO Update pipeline run status according to run result
		if runResult.EndTime != "" {
			// this means the pipeline run has ended
			return ctrl.Result{}, nil
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

	// label pipelinerun as running
	if pr.Annotations == nil {
		pr.Annotations = make(map[string]string)
	}
	pr.Annotations[devopsv1alpha4.JenkinsPipelineRunIdKey] = runResponse.ID
	pr.Status.MarkPending()
	now := v1.NewTime(time.Now())
	pr.Status.StartTime = &now
	pr.Status.UpdateTime = &now

	if err := r.Client.Update(ctx, &pr); err != nil {
		log.Error(err, "unable to update pipeline run")
		return ctrl.Result{}, err
	}
	if err := r.Client.Status().Update(ctx, &pr); err != nil {
		log.Error(err, "unable to update pipeline run status.")
		return ctrl.Result{}, err
	}

	// requeue immediately
	return ctrl.Result{Requeue: true}, nil
}

func (r *Reconciler) getPipelineRunResult(projectName, pipelineName string, pr *devopsv1alpha4.PipelineRun) (runResult *devopsClient.PipelineRun, err error) {
	// get pipeline run id
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

func buildHttpParametersForRunning(pr *devopsv1alpha4.PipelineRun) (*devopsClient.HttpParameters, error) {
	if pr == nil {
		return nil, errors.New("invalid pipeline run")
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

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1alpha4.PipelineRun{}).
		Complete(r)
}
