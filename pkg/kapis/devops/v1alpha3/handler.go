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

package v1alpha3

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful/v3"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	apiserverRequest "github.com/kubesphere/ks-devops/pkg/apiserver/request"
	devopsClient "github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/client/k8s"
	"github.com/kubesphere/ks-devops/pkg/constants"
	"github.com/kubesphere/ks-devops/pkg/kapis"
	devopsModel "github.com/kubesphere/ks-devops/pkg/models/devops"
	servererr "github.com/kubesphere/ks-devops/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	resourcesV1beta1 "kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

type devopsHandler struct {
	k8sClient    k8s.Client
	devopsClient devopsClient.Interface

	am am.AccessManagementInterface
}

func newDevOpsHandler(client client.Client, devopsClient devopsClient.Interface, k8sClient k8s.Client, runtimeCache cache.Cache) *devopsHandler {
	resourceMgr, err := resourcesV1beta1.New(context.Background(), client, runtimeCache)
	if err != nil {
		klog.Fatalf("failed to create resource manager, error: %+v", err)
		return nil
	}
	return &devopsHandler{
		k8sClient:    k8sClient,
		devopsClient: devopsClient,
		am:           am.NewOperator(resourceMgr),
	}
}

// GetDevOpsProject handler about get/list/post/put/delete
func (h *devopsHandler) GetDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	devopsProject := request.PathParameter("devops")
	generateNameFlag := request.QueryParameter("generateName")
	check := request.QueryParameter("check")

	if check == "true" {
		h.CheckDevopsName(request, response, workspace, devopsProject, generateNameFlag)
		return
	}

	if devopsOperator, err := h.getDevOps(request); err == nil {
		var project *v1alpha3.DevOpsProject
		var err error

		switch generateNameFlag {
		case "true":
			project, err = devopsOperator.GetDevOpsProjectByGenerateName(workspace, devopsProject)
		default:
			project, err = devopsOperator.GetDevOpsProject(workspace, devopsProject)
		}
		errorHandle(request, response, project, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) ListDevOpsProject(request *restful.Request, response *restful.Response) {
	if devopsOperator, err := h.getDevOps(request); err == nil {
		workspace := request.PathParameter("workspace")
		queryParam := query.ParseQueryParameter(request)

		var requestUser user.Info
		if username := request.PathParameter("workspacemember"); username != "" {
			requestUser = &user.DefaultInfo{
				Name: username,
			}
		} else {
			var ok bool
			requestUser, ok = apiserverRequest.UserFrom(request.Request.Context())
			if !ok {
				err := fmt.Errorf("cannot obtain user info")
				klog.Errorln(err)
				kapis.HandleForbidden(response, nil, err)
				return
			}
		}
		projectList, err := devopsOperator.ListDevOpsProject(requestUser, workspace, queryParam)
		errorHandle(request, response, projectList, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) CreateDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var devOpsProject v1alpha3.DevOpsProject
	err := request.ReadEntity(&devOpsProject)

	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	if devopsOperator, err := h.getDevOps(request); err == nil {
		created, err := devopsOperator.CreateDevOpsProject(workspace, &devOpsProject)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				kapis.HandleNotFound(response, request, err)
				return
			} else if errors.IsConflict(err) {
				kapis.HandleConflict(response, request, err)
				return
			}
			kapis.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(created)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) UpdateDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var devOpsProject v1alpha3.DevOpsProject
	err := request.ReadEntity(&devOpsProject)

	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	if devopsOperator, err := h.getDevOps(request); err == nil {
		project, err := devopsOperator.UpdateDevOpsProject(workspace, &devOpsProject)
		errorHandle(request, response, project, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) DeleteDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	devops := request.PathParameter("devops")

	if devopsOperator, err := h.getDevOps(request); err == nil {
		err := devopsOperator.DeleteDevOpsProject(workspace, devops)
		errorHandle(request, response, nil, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

// GetPipeline pipeline handler about get/list/post/put/delete
func (h *devopsHandler) GetPipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	pipeline := request.PathParameter("pipeline")

	if devopsOperator, err := h.getDevOps(request); err == nil {
		obj, err := devopsOperator.GetPipelineObj(devops, pipeline)
		errorHandle(request, response, obj, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) ListPipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	queryParam := query.ParseQueryParameter(request)

	if devopsOperator, err := h.getDevOps(request); err == nil {
		objs, err := devopsOperator.ListPipelineObj(devops, queryParam)
		errorHandle(request, response, objs, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) CreatePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var pipeline v1alpha3.Pipeline
	err := request.ReadEntity(&pipeline)

	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	if devopsOperator, err := h.getDevOps(request); err == nil {
		created, err := devopsOperator.CreatePipelineObj(devops, &pipeline)
		errorHandle(request, response, created, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) UpdatePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")

	var pipeline v1alpha3.Pipeline
	err := request.ReadEntity(&pipeline)

	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	if devopsOperator, err := h.getDevOps(request); err == nil {
		obj, err := devopsOperator.UpdatePipelineObj(devops, &pipeline)
		errorHandle(request, response, obj, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

// GenericPayload represents a generic HTTP request payload data structure
type GenericPayload struct {
	Data string `json:"data"`
}

// GenericResponse represents a generic HTTP response data structure
type GenericResponse struct {
	Result string `json:"result"`
}

// NewSuccessResponse creates a response for the success case
func NewSuccessResponse() *GenericResponse {
	return &GenericResponse{
		Result: "success",
	}
}

// GenericArrayResponse represents a generic array response
type GenericArrayResponse struct {
	Status  string   `json:"status"`
	Data    []string `json:"data"`
	Message string   `json:"message"`
}

// NewSuccessGenericArrayResponse creates a generic array response
func NewSuccessGenericArrayResponse(data []string) *GenericArrayResponse {
	return &GenericArrayResponse{
		Status: "success",
		Data:   data,
	}
}

func (h *devopsHandler) UpdateJenkinsfile(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter("devops")
	pipelineName := request.PathParameter("pipeline")
	mode := request.QueryParameter("mode")

	var err error
	payload := &GenericPayload{}
	if err = request.ReadEntity(payload); err != nil {
		kapis.HandleBadRequest(response, request, err)
		return
	}

	var devopsOperator devopsModel.DevopsOperator
	if devopsOperator, err = h.getDevOps(request); err == nil {
		err = devopsOperator.UpdateJenkinsfile(projectName, pipelineName, mode, payload.Data)
	}
	errorHandle(request, response, NewSuccessResponse(), err)
}

func (h *devopsHandler) DeletePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	pipeline := request.PathParameter("pipeline")

	klog.V(8).Infof("ready to delete pipeline %s/%s", devops, pipeline)

	if devopsOperator, err := h.getDevOps(request); err == nil {
		err := devopsOperator.DeletePipelineObj(devops, pipeline)
		errorHandle(request, response, nil, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

// GetCredential handler about get/list/post/put/delete
func (h *devopsHandler) GetCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	credential := request.PathParameter("credential")

	if devopsOperator, err := h.getDevOps(request); err == nil {
		obj, err := devopsOperator.GetCredentialObj(devops, credential)
		errorHandle(request, response, obj, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) ListCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	queryParam := query.ParseQueryParameter(request)

	if devopsOperator, err := h.getDevOps(request); err == nil && devopsOperator != nil {
		objs, err := devopsOperator.ListCredentialObj(devops, queryParam)
		errorHandle(request, response, objs, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) CreateCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var obj v1.Secret
	err := request.ReadEntity(&obj)

	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	if devopsOperator, err := h.getDevOps(request); err == nil {
		created, err := devopsOperator.CreateCredentialObj(devops, &obj)
		errorHandle(request, response, created, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) UpdateCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var obj v1.Secret
	err := request.ReadEntity(&obj)

	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	if devopsOperator, err := h.getDevOps(request); err == nil {
		updated, err := devopsOperator.UpdateCredentialObj(devops, &obj)
		errorHandle(request, response, updated, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func errorHandle(request *restful.Request, response *restful.Response, obj interface{}, err error) {
	if obj == nil {
		obj = servererr.None
	}

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			kapis.HandleNotFound(response, request, err)
			return
		}
		kapis.HandleBadRequest(response, request, err)
		return
	}
	_ = response.WriteEntity(obj)
}

func (h *devopsHandler) DeleteCredential(request *restful.Request, response *restful.Response) {
	devopsProject := request.PathParameter("devops")
	credential := request.PathParameter("credential")

	if devopsOperator, err := h.getDevOps(request); err == nil {
		err := devopsOperator.DeleteCredentialObj(devopsProject, credential)
		errorHandle(request, response, servererr.None, err)
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) getJenkinsLabels(request *restful.Request, response *restful.Response) {
	devopsOperator, err := h.getDevOps(request)
	if err != nil {
		kapis.HandleBadRequest(response, request, err)
		return
	}

	var labels []string
	if labels, err = devopsOperator.GetJenkinsAgentLabels(); err != nil {
		kapis.HandleBadRequest(response, request, err)
	} else {
		errorHandle(request, response, NewSuccessGenericArrayResponse(labels), nil)
	}
}

// TODO: delete this func and add DevopsOperator to devopsHandler
func (h *devopsHandler) getDevOps(request *restful.Request) (operator devopsModel.DevopsOperator, err error) {
	ctx := request.Request.Context()
	token := ctx.Value(constants.K8SToken).(constants.ContextKeyK8SToken)

	var k8sClient k8s.Client
	if token == "" {
		k8sClient = h.k8sClient
	} else {
		klog.V(9).Infof("get DevOps client with token: %s", token)
		k8sClient, err = k8s.NewKubernetesClientWithToken(string(token), h.k8sClient.Config().Host)
	}

	if err == nil {
		operator = devopsModel.NewDevopsOperator(h.devopsClient, k8sClient.Kubernetes(), k8sClient.KubeSphere(), h.am)
	}
	return
}

func (h *devopsHandler) CheckDevopsName(request *restful.Request, response *restful.Response, workspace, devopsName string, generateNameFlag string) {

	var result map[string]interface{}
	if devopsOperator, err := h.getDevOps(request); err == nil {
		switch generateNameFlag {
		case "true":
			result, err = devopsOperator.CheckDevopsProject(workspace, devopsName)
			if err != nil {
				errorHandle(request, response, result, err)
			}
			errorHandle(request, response, result, nil)
		default:
			errorHandle(request, response, nil, errors.NewBadRequest("generateNameFlag can not be false"))
		}
	} else {
		kapis.HandleBadRequest(response, request, err)
	}
}
