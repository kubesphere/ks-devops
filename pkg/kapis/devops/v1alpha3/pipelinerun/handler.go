/*
Copyright 2022 The KubeSphere Authors.

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
	"io"
	"strconv"

	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	apiserverrequest "kubesphere.io/devops/pkg/apiserver/request"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/models/pipelinerun"
	resourcesV1alpha3 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// apiHandlerOption holds some useful tools for API handler.
type apiHandlerOption struct {
	client client.Client
}

// apiHandler contains functions to handle coming request and give a response.
type apiHandler struct {
	apiHandlerOption
}

// newAPIHandler creates an APIHandler.
func newAPIHandler(o apiHandlerOption) *apiHandler {
	return &apiHandler{o}
}

func (h *apiHandler) listPipelineRuns(request *restful.Request, response *restful.Response) {
	nsName := request.PathParameter("namespace")
	pipName := request.PathParameter("pipeline")
	branchName := request.QueryParameter("branch")
	backward, err := strconv.ParseBool(request.QueryParameter("backward"))
	if err != nil {
		// by default, we have to guarantee backward compatibility
		backward = true
	}

	queryParam := query.ParseQueryParameter(request)

	// validate the Pipeline
	pipeline := &v1alpha3.Pipeline{}
	err = h.client.Get(context.Background(), client.ObjectKey{Namespace: nsName, Name: pipName}, pipeline)
	if err != nil {
		api.HandleError(request, response, err)
		return
	}

	// build label selector
	labelSelector, err := buildLabelSelector(queryParam, pipeline.Name)
	if err != nil {
		api.HandleError(request, response, err)
		return
	}

	opts := make([]client.ListOption, 0, 3)
	opts = append(opts, client.InNamespace(pipeline.Namespace))
	opts = append(opts, client.MatchingLabelsSelector{Selector: labelSelector})
	if branchName != "" {
		opts = append(opts, client.MatchingFields{v1alpha3.PipelineRunSCMRefNameField: branchName})
	}

	var prs v1alpha3.PipelineRunList
	// fetch PipelineRuns
	if err := h.client.List(context.Background(), &prs, opts...); err != nil {
		api.HandleError(request, response, err)
		return
	}

	var listHandler resourcesV1alpha3.ListHandler = listHandler{}
	if backward {
		listHandler = backwardListHandler{}
	}
	apiResult := resourcesV1alpha3.ToListResult(convertPipelineRunsToObject(prs.Items), queryParam, listHandler)
	_ = response.WriteAsJson(apiResult)

	go func() {
		err := h.requestSyncPipelineRun(client.ObjectKey{Namespace: pipeline.Namespace, Name: pipeline.Name})
		if err != nil {
			klog.Errorf("failed to request to synchronize PipelineRuns. Pipeline = %s", pipeline.Name)
			return
		}
	}()
}

func (h *apiHandler) requestSyncPipelineRun(key client.ObjectKey) error {
	// get latest Pipeline
	pipelineToUpdate := &v1alpha3.Pipeline{}
	err := h.client.Get(context.Background(), key, pipelineToUpdate)
	if err != nil {
		return err
	}
	// update pipeline annotations
	if pipelineToUpdate.Annotations == nil {
		pipelineToUpdate.Annotations = make(map[string]string)
	}
	pipelineToUpdate.Annotations[v1alpha3.PipelineRequestToSyncRunsAnnoKey] = "true"
	// update the pipeline
	if err := h.client.Update(context.Background(), pipelineToUpdate); err != nil && !apierrors.IsConflict(err) {
		// we allow the conflict error here
		return err
	}
	return nil
}

func (h *apiHandler) createPipelineRun(request *restful.Request, response *restful.Response) {
	nsName := request.PathParameter("namespace")
	pipName := request.PathParameter("pipeline")
	branch := request.QueryParameter("branch")
	payload := devops.RunPayload{}
	if err := request.ReadEntity(&payload); err != nil && err != io.EOF {
		api.HandleBadRequest(response, request, err)
		return
	}
	// validate the Pipeline
	var pipeline v1alpha3.Pipeline
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: nsName, Name: pipName}, &pipeline); err != nil {
		api.HandleError(request, response, err)
		return
	}

	var (
		scm *v1alpha3.SCM
		err error
	)
	if scm, err = CreateScm(&pipeline.Spec, branch); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	// get current login user from request context
	user, ok := apiserverrequest.UserFrom(request.Request.Context())
	if !ok || user == nil {
		// should never happen
		err := fmt.Errorf("unauthenticated user entered to create PipelineRun for Pipeline '%s/%s'", nsName, pipName)
		api.HandleUnauthorized(response, request, err)
		return
	}
	// create PipelineRun
	pr := CreatePipelineRun(&pipeline, &payload, scm)
	if user != nil && user.GetName() != "" {
		pr.GetAnnotations()[v1alpha3.PipelineRunCreatorAnnoKey] = user.GetName()
	}
	if err := h.client.Create(context.Background(), pr); err != nil {
		api.HandleError(request, response, err)
		return
	}

	_ = response.WriteEntity(pr)
}

func (h *apiHandler) getPipelineRun(request *restful.Request, response *restful.Response) {
	nsName := request.PathParameter("namespace")
	prName := request.PathParameter("pipelinerun")

	// get pipelinerun
	var pr v1alpha3.PipelineRun
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: nsName, Name: prName}, &pr); err != nil {
		api.HandleError(request, response, err)
		return
	}
	_ = response.WriteEntity(&pr)
}

func (h *apiHandler) getNodeDetails(request *restful.Request, response *restful.Response) {
	namespaceName := request.PathParameter("namespace")
	pipelineRunName := request.PathParameter("pipelinerun")

	// get pipelinerun
	pr := &v1alpha3.PipelineRun{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: namespaceName, Name: pipelineRunName}, pr); err != nil {
		api.HandleError(request, response, err)
		return
	}

	// get stage status
	stagesJSON, ok := pr.Annotations[v1alpha3.JenkinsPipelineRunStagesStatusAnnoKey]
	if !ok {
		// If the stages status dose not exist, set it as an empty array
		stagesJSON = "[]"
	}
	stages := []pipelinerun.NodeDetail{}
	if err := json.Unmarshal([]byte(stagesJSON), &stages); err != nil {
		api.HandleError(request, response, err)
		return
	}

	// TODO(johnniang): Check current user Handle the approvable field of NodeDetail
	// this is a temporary solution of approvable
	for i := range stages {
		for j := range stages[i].Steps {
			stages[i].Steps[j].Approvable = true
		}
	}

	_ = response.WriteEntity(&stages)
}
