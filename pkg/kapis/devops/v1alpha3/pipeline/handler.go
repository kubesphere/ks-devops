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

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"kubesphere.io/devops/pkg/kapis"

	"github.com/emicklei/go-restful"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	modelpipeline "kubesphere.io/devops/pkg/models/pipeline"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type apiHandlerOption struct {
	client client.Client
}

type apiHandler struct {
	apiHandlerOption
}

func newAPIHandler(option apiHandlerOption) *apiHandler {
	return &apiHandler{
		apiHandlerOption: option,
	}
}

func (h *apiHandler) getBranches(request *restful.Request, response *restful.Response) {
	namespaceName := request.PathParameter("namespace")
	pipelineName := request.PathParameter("pipeline")
	filter := job.Filter(request.QueryParameter("filter"))

	// get pipelinerun
	pipeline := &v1alpha3.Pipeline{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: namespaceName, Name: pipelineName}, pipeline); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	if pipeline.Spec.Type != v1alpha3.MultiBranchPipelineType {
		kapis.HandleBadRequest(response, request, fmt.Errorf("invalid multi-branch Pipeline provided"))
		return
	}

	branchesJSON := pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey]
	var branches []modelpipeline.Branch
	if err := json.Unmarshal([]byte(branchesJSON), &branches); err != nil {
		// ignore this error
		klog.Errorf("unable to unmarshal branches JSON: %s, and err = %v", branchesJSON, err)
	}

	// filter branches with filter
	branches = filterBranches(branches, filter)

	queryParam := query.ParseQueryParameter(request)
	total := len(branches)
	startIndex, endIndex := queryParam.Pagination.GetValidPagination(total)
	_ = response.WriteEntity(api.NewListResult(branchSlice(branches[startIndex:endIndex]).toGenericSlice(), total))
}

func (h *apiHandler) getBranch(request *restful.Request, response *restful.Response) {
	namespaceName := request.PathParameter("namespace")
	pipelineName := request.PathParameter("pipeline")
	branch := request.PathParameter("branch")

	// get pipelinerun
	pipeline := &v1alpha3.Pipeline{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: namespaceName, Name: pipelineName}, pipeline); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	if pipeline.Spec.Type != v1alpha3.MultiBranchPipelineType {
		kapis.HandleBadRequest(response, request, fmt.Errorf("invalid multi-branch Pipeline provided"))
		return
	}

	branchesJSON := pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey]
	branches := []modelpipeline.Branch{}
	if err := json.Unmarshal([]byte(branchesJSON), &branches); err != nil {
		// ignore this error
		klog.Errorf("unable to unmarshal branches JSON: %s, and err = %v", branchesJSON, err)
	}

	exist, searchedBranch := modelpipeline.BranchSlice(branches).SearchByName(branch)
	if !exist {
		// branch was not found
		kapis.HandleNotFound(response, request, fmt.Errorf("Branch %s was not found", branch))
		return
	}
	_ = response.WriteEntity(searchedBranch)
}
