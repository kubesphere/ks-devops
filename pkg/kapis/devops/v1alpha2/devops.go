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

package v1alpha2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"kubesphere.io/devops/pkg/kapis"

	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/apiserver/request"

	"github.com/emicklei/go-restful"
	log "k8s.io/klog"
	"k8s.io/klog/v2"

	"kubesphere.io/devops/pkg/api/devops/v1alpha3"

	"kubesphere.io/devops/pkg/api"
	clientDevOps "kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/devops/jenkins"

	//"kubesphere.io/devops/pkg/apiserver/authorization/authorizer"
	//"kubesphere.io/devops/pkg/apiserver/request"
	"kubesphere.io/devops/pkg/constants"
	"kubesphere.io/devops/pkg/models/devops"
)

const jenkinsHeaderPre = "X-"

func (h *ProjectPipelineHandler) GetPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.GetPipeline(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) getPipelinesByRequest(req *restful.Request) (api.ListResult, error) {
	// this is a very trick way, but don't have a better solution for now
	var namespace string

	// parse query from the request
	nameReg, namespace := parseNameFilterFromQuery(req.QueryParameter("q"))

	// compatible
	queryParam := buildPipelineSearchQueryParam(req, nameReg)

	// make sure we have an appropriate value
	return h.devopsOperator.ListPipelineObj(namespace, queryParam)
}

func buildPipelineSearchQueryParam(req *restful.Request, nameReg string) (q *query.Query) {
	startStr := req.QueryParameter(query.ParameterStart)
	if req.Request.Form.Get(query.ParameterPage) == "" && startStr != "" {
		// for pagination compatibility
		req.Request.Form.Set(query.ParameterStart, startStr)
	}

	q = query.ParseQueryParameter(req)

	// for filter compatibility
	if _, ok := q.Filters[query.FieldName]; !ok {
		// set name filter by default
		q.Filters[query.FieldName] = query.Value(nameReg)
	}

	// for compare compatibility
	if len(req.QueryParameter(query.ParameterOrderBy)) == 0 {
		// set sort by as name by default
		q.SortBy = query.FieldName
	}

	// for ascending compatibility
	if len(req.QueryParameter(query.ParameterAscending)) == 0 {
		// set ascending as true by default
		q.Ascending = true
	}

	return
}

func parseNameFilterFromQuery(query string) (nameReg, namespace string) {
	for _, val := range strings.Split(query, ";") {
		if strings.HasPrefix(val, "pipeline:") {
			nsAndName := strings.TrimPrefix(val, "pipeline:")
			filterMeta := strings.Split(nsAndName, "/")
			if len(filterMeta) >= 2 {
				namespace = filterMeta[0]
				nameReg = filterMeta[1] // the format is '*keyword*'
				nameReg = strings.TrimSuffix(nameReg, "*")
				nameReg = strings.TrimPrefix(nameReg, "*")
			} else if len(filterMeta) > 0 {
				namespace = filterMeta[0]
			}
		}
	}
	return
}

func (h *ProjectPipelineHandler) ListPipelines(req *restful.Request, resp *restful.Response) {
	objs, err := h.getPipelinesByRequest(req)
	if err != nil {
		parseErr(err, resp)
		return
	}

	// get all pipelines which come from ks
	pipelineList := &clientDevOps.PipelineList{
		Total: objs.TotalItems,
		Items: make([]clientDevOps.Pipeline, 0, len(objs.Items)),
	}
	pipelineMap := make(map[string]int)
	for i := range objs.Items {
		if pipeline, ok := objs.Items[i].(*v1alpha3.Pipeline); ok {
			pipelineMap[pipeline.Name] = i
			pipelineList.Items = append(pipelineList.Items, clientDevOps.Pipeline{
				Name:        pipeline.Name,
				Annotations: pipeline.Annotations,
			})
		}
	}

	// get all pipelines which come from Jenkins
	// fill out the rest fields
	if jenkinsQuery, err := jenkins.ParseJenkinsQuery(req.Request.URL.RawQuery); err == nil {
		jenkinsQuery.Set("limit", "10000")
		jenkinsQuery.Set("start", "0")
		req.Request.URL.RawQuery = jenkinsQuery.Encode()
	}
	res, err := h.devopsOperator.ListPipelines(req.Request)
	if err != nil {
		log.Error(err)
	} else {
		for i, _ := range res.Items {
			if index, ok := pipelineMap[res.Items[i].Name]; ok {
				// keep annotations field of pipelineList
				annotations := pipelineList.Items[index].Annotations
				pipelineList.Items[index] = res.Items[i]
				pipelineList.Items[index].Annotations = annotations
			}
		}
	}

	_ = resp.WriteEntity(pipelineList)
}

func (h *ProjectPipelineHandler) GetPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetPipelineRun(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ListPipelineRuns(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.ListPipelineRuns(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) StopPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.StopPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ReplayPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.ReplayPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) RunPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.RunPipeline(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetArtifacts(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetRunLog(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, header, err := h.devopsOperator.GetStepLog(projectName, pipelineName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	for k, v := range header {
		if strings.HasPrefix(k, jenkinsHeaderPre) {
			resp.AddHeader(k, v[0])
		}
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")

	res, err := h.devopsOperator.GetNodeSteps(projectName, pipelineName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetPipelineRunNodes(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetPipelineRunNodes(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

// approvableCheck requires the users who have PipelineRun management permission to
// approve a step. If the particular submitters exist, we also restrict the users
// who are Pipeline creator or in the particular submitters can be able to approve or reject a step.
func (h *ProjectPipelineHandler) approvableCheck(nodes []clientDevOps.NodesDetail, pipe pipelineParam) {
	userInfo, ok := request.UserFrom(pipe.Context)
	if !ok {
		klog.V(6).Infof("cannot get the current user when checking the approvable with pipeline '%s/%s'",
			pipe.ProjectName, pipe.Name)
		return
	}

	// check every input steps if it's approvable
	for i := range nodes {
		node := &nodes[i]
		if node.State != clientDevOps.StatePaused {
			continue
		}

		for j := range node.Steps {
			step := &node.Steps[j]
			if step.State != clientDevOps.StatePaused || step.Input == nil {
				continue
			}
			if len(step.Input.GetSubmitters()) == 0 {
				// TODO: no particular submitters, we make the PipelineRun approvable by default.
				// Until we can obtain the permissions of the currently logged in user
				step.Approvable = true
			} else {
				isCreator := h.createdBy(pipe.ProjectName, pipe.Name, userInfo.GetName())
				step.Approvable = isCreator || step.Input.Approvable(userInfo.GetName())
			}
		}
	}
}

func (h *ProjectPipelineHandler) createdBy(projectName string, pipelineName string, currentUserName string) bool {
	if pipeline, err := h.devopsOperator.GetPipelineObj(projectName, pipelineName); err == nil {
		if creator, ok := pipeline.Annotations[constants.CreatorAnnotationKey]; ok {
			return creator == currentUserName
		}
	} else {
		log.V(4).Infof("cannot get pipeline %s/%s, error %#v", projectName, pipelineName, err)
	}
	return false
}

func (h *ProjectPipelineHandler) hasSubmitPermission(req *restful.Request) (hasPermit bool, err error) {
	pipeParam := parsePipelineParam(req)
	httpReq := &http.Request{
		URL:      req.Request.URL,
		Header:   req.Request.Header,
		Form:     req.Request.Form,
		PostForm: req.Request.PostForm,
	}

	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")
	branchName := req.PathParameter("branch")

	// check if current user can approve this input
	var res []clientDevOps.NodesDetail

	if branchName == "" {
		res, err = h.devopsOperator.GetNodesDetail(pipeParam.ProjectName, pipeParam.Name, runId, httpReq)
	} else {
		res, err = h.devopsOperator.GetBranchNodesDetail(pipeParam.ProjectName, pipeParam.Name, branchName, runId, httpReq)
	}

	if err == nil {
		h.approvableCheck(res, parsePipelineParam(req))

		for _, node := range res {
			if node.ID != nodeId {
				continue
			}

			for _, step := range node.Steps {
				if step.ID != stepId {
					continue
				}
				if step.Input == nil {
					err = fmt.Errorf("the step '%s' is not approvable", step.DisplayName)
				}
				hasPermit = step.Approvable
				break
			}
			break
		}
	} else {
		log.V(4).Infof("cannot get nodes detail, error: %v", err)
		err = errors.New("cannot get the submitters of current step")
		return
	}
	if !hasPermit && err == nil {
		err = fmt.Errorf("you have no permission to approve this step")
	}
	return
}

func (h *ProjectPipelineHandler) SubmitInputStep(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	var response []byte
	var err error
	var ok bool

	if ok, err = h.hasSubmitPermission(req); !ok || err != nil {
		msg := map[string]string{
			"allow":   "false",
			"message": fmt.Sprintf("%v", err),
		}

		response, _ = json.Marshal(msg)
	} else {
		response, err = h.devopsOperator.SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId, req.Request)
		if err != nil {
			parseErr(err, resp)
			return
		}
	}
	resp.Write(response)
}

func (h *ProjectPipelineHandler) GetNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetNodesDetail(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	h.approvableCheck(res, parsePipelineParam(req))

	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")

	res, err := h.devopsOperator.GetBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchPipelineRun(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) StopBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.StopBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ReplayBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.ReplayBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) RunBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")

	res, err := h.devopsOperator.RunBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchArtifacts(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchRunLog(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetBranchStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, header, err := h.devopsOperator.GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)

	if err != nil {
		parseErr(err, resp)
		return
	}
	for k, v := range header {
		if strings.HasPrefix(k, jenkinsHeaderPre) {
			resp.AddHeader(k, v[0])
		}
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetBranchNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")

	res, err := h.devopsOperator.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchPipelineRunNodes(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) SubmitBranchInputStep(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	var response []byte
	var err error
	var ok bool

	if ok, err = h.hasSubmitPermission(req); !ok || err != nil {
		msg := map[string]string{
			"allow":   "false",
			"message": fmt.Sprintf("%v", err),
		}

		response, _ = json.Marshal(msg)
	} else {
		response, err = h.devopsOperator.SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)
		if err != nil {
			parseErr(err, resp)
			return
		}
	}

	resp.Write(response)
}

func (h *ProjectPipelineHandler) GetBranchNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchNodesDetail(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	h.approvableCheck(res, parsePipelineParam(req))
	resp.WriteAsJson(res)
}

func parsePipelineParam(req *restful.Request) pipelineParam {
	return pipelineParam{
		Workspace:   req.PathParameter("workspace"),
		ProjectName: req.PathParameter("devops"),
		Name:        req.PathParameter("pipeline"),
		Context:     req.Request.Context(),
	}
}

func (h *ProjectPipelineHandler) GetPipelineBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.GetPipelineBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ScanBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.ScanBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetConsoleLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.GetConsoleLog(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetCrumb(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GetCrumb(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetSCMServers(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.GetSCMServers(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetSCMOrg(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.GetSCMOrg(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetOrgRepo(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")
	organizationId := req.PathParameter("organization")

	res, err := h.devopsOperator.GetOrgRepo(scmId, organizationId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) CreateSCMServers(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.CreateSCMServers(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) Validate(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.Validate(scmId, req.Request)
	if err != nil {
		log.Error(err)
		if jErr, ok := err.(*devops.JkError); ok {
			if jErr.Code != http.StatusUnauthorized {
				resp.WriteError(jErr.Code, err)
			} else {
				resp.WriteHeader(http.StatusPreconditionRequired)
			}
		} else {
			resp.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetNotifyCommit(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GetNotifyCommit(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) PostNotifyCommit(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GetNotifyCommit(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) GithubWebhook(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GithubWebhook(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) genericWebhook(req *restful.Request, resp *restful.Response) {
	if res, err := h.devopsOperator.GenericWebhook(req.Request); err != nil {
		parseErr(err, resp)
		return
	} else {
		_, _ = resp.Write(res)
	}
}

func (h *ProjectPipelineHandler) CheckScriptCompile(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	resBody, err := h.devopsOperator.CheckScriptCompile(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.WriteAsJson(resBody)
}

func (h *ProjectPipelineHandler) CheckCron(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")

	res, err := h.devopsOperator.CheckCron(projectName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ToJenkinsfile(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.ToJenkinsfile(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ToJSON(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.ToJSON(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetProjectCredentialUsage(req *restful.Request, resp *restful.Response) {
	projectId := req.PathParameter("devops")
	credentialId := req.PathParameter("credential")
	response, err := h.projectCredentialGetter.GetProjectCredentialUsage(projectId, credentialId)
	if err != nil {
		log.Errorf("%+v", err)
		kapis.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteAsJson(response)
	return

}

func parseErr(err error, resp *restful.Response) {
	log.Error(err)
	if jErr, ok := err.(*devops.JkError); ok {
		resp.WriteError(jErr.Code, err)
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
	return
}
