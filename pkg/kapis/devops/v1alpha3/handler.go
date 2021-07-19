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
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/constants"

	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/apiserver/query"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/models/devops"
	servererr "kubesphere.io/devops/pkg/server/errors"
	"kubesphere.io/devops/pkg/server/params"
)

type devopsHandler struct {
	k8sClient    k8s.Client
	devopsClient devopsClient.Interface
}

func newDevOpsHandler(devopsClient devopsClient.Interface, k8sClient k8s.Client) *devopsHandler {

	return &devopsHandler{
		k8sClient:    k8sClient,
		devopsClient: devopsClient,
	}
}

// devopsproject handler about get/list/post/put/delete
func (h *devopsHandler) GetDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	devops := request.PathParameter("devops")

	if client, err := h.getDevOps(request); err == nil {
		project, err := client.GetDevOpsProject(workspace, devops)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(project)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) ListDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	limit, offset := params.ParsePaging(request)

	if client, err := h.getDevOps(request); err == nil {
		projectList, err := client.ListDevOpsProject(workspace, limit, offset)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(projectList)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) CreateDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var devOpsProject v1alpha3.DevOpsProject
	err := request.ReadEntity(&devOpsProject)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if client, err := h.getDevOps(request); err == nil {
		created, err := client.CreateDevOpsProject(workspace, &devOpsProject)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			} else if errors.IsConflict(err) {
				api.HandleConflict(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(created)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) UpdateDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var devOpsProject v1alpha3.DevOpsProject
	err := request.ReadEntity(&devOpsProject)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if client, err := h.getDevOps(request); err == nil {
		project, err := client.UpdateDevOpsProject(workspace, &devOpsProject)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(project)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) DeleteDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	devops := request.PathParameter("devops")

	if client, err := h.getDevOps(request); err == nil {
		err := client.DeleteDevOpsProject(workspace, devops)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(servererr.None)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

// pipeline handler about get/list/post/put/delete
func (h *devopsHandler) GetPipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	pipeline := request.PathParameter("pipeline")

	if client, err := h.getDevOps(request); err == nil {
		obj, err := client.GetPipelineObj(devops, pipeline)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(obj)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) ListPipeline(request *restful.Request, response *restful.Response) {
	devopsProject := request.PathParameter("devops")
	limit, offset := params.ParsePaging(request)

	if client, err := h.getDevOps(request); err == nil {
		objs, err := client.ListPipelineObj(devopsProject, nil, nil, limit, offset)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}

		_ = response.WriteEntity(objs)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) CreatePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var pipeline v1alpha3.Pipeline
	err := request.ReadEntity(&pipeline)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if client, err := h.getDevOps(request); err == nil {
		created, err := client.CreatePipelineObj(devops, &pipeline)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(created)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) UpdatePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")

	var pipeline v1alpha3.Pipeline
	err := request.ReadEntity(&pipeline)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if client, err := h.getDevOps(request); err == nil {
		obj, err := client.UpdatePipelineObj(devops, &pipeline)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(obj)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) DeletePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	pipeline := request.PathParameter("pipeline")

	klog.V(8).Infof("ready to delete pipeline %s/%s", devops, pipeline)

	if client, err := h.getDevOps(request); err == nil {
		err := client.DeletePipelineObj(devops, pipeline)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(servererr.None)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

//credential handler about get/list/post/put/delete
func (h *devopsHandler) GetCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	credential := request.PathParameter("credential")

	if client, err := h.getDevOps(request); err == nil {
		obj, err := client.GetCredentialObj(devops, credential)

		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}

		_ = response.WriteEntity(obj)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) ListCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	query := query.ParseQueryParameter(request)

	if client, err := h.getDevOps(request); err == nil {
		objs, err := client.ListCredentialObj(devops, query)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(objs)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) CreateCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var obj v1.Secret
	err := request.ReadEntity(&obj)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if client, err := h.getDevOps(request); err == nil {
		created, err := client.CreateCredentialObj(devops, &obj)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(created)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) UpdateCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var obj v1.Secret
	err := request.ReadEntity(&obj)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if client, err := h.getDevOps(request); err == nil {
		updated, err := client.UpdateCredentialObj(devops, &obj)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(updated)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) DeleteCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	credential := request.PathParameter("credential")

	if client, err := h.getDevOps(request); err == nil {
		err := client.DeleteCredentialObj(devops, credential)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleBadRequest(response, request, err)
			return
		}
		_ = response.WriteEntity(servererr.None)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}

func (h *devopsHandler) getDevOps(request *restful.Request) (devops.DevopsOperator, error) {
	ctx := request.Request.Context()
	token := ctx.Value(constants.K8SToken).(string)

	klog.V(9).Infof("get DevOps client with  token: %s", token)
	kubernetesClient, err := k8s.NewKubernetesClientWithToken(token, h.k8sClient.Config().Host)
	if err != nil {
		return nil, err
	}

	return devops.NewDevopsOperator(h.devopsClient, kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere()), nil
}
