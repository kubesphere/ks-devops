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
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/common"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/pipeline"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/pipelinerun"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/scm"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/template"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/webhook"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/constants"
	"kubesphere.io/devops/pkg/server/params"
)

// TODO perhaps we can find a better way to declaim the permission needs of the apiserver
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=devopsprojects,verbs=get;list;update;delete;create;watch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;update;delete;create;watch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns,verbs=get;list;update;delete;create;watch

// GroupVersion describes CRD group and its version.
var GroupVersion = schema.GroupVersion{Group: api.GroupName, Version: "v1alpha3"}

// AddToContainer adds web service into container.
func AddToContainer(container *restful.Container, devopsClient devopsClient.Interface,
	k8sClient k8s.Client, client client.Client) (wss []*restful.WebService) {

	services := []*restful.WebService{
		runtime.NewWebService(v1alpha3.GroupVersion),
		runtime.NewWebServiceWithoutGroup(v1alpha3.GroupVersion),
	}

	for _, service := range services {
		registerRoutes(devopsClient, k8sClient, client, service)
		pipelinerun.RegisterRoutes(service, devopsClient, client)
		pipeline.RegisterRoutes(service, client)
		template.RegisterRoutes(service, &common.Options{
			GenericClient: client,
		})
		webhook.RegisterWebhooks(client, service)
		container.Add(service)
	}
	return services
}

func registerRoutes(devopsClient devopsClient.Interface, k8sClient k8s.Client, client client.Client, ws *restful.WebService) {
	handler := newDevOpsHandler(devopsClient, k8sClient)
	registerRoutersForCredentials(handler, ws)
	registerRoutersForPipelines(handler, ws)
	registerRoutersForWorkspace(handler, ws)
	scm.RegisterRoutersForSCM(client, ws)
}

func registerRoutersForCredentials(handler *devopsHandler, ws *restful.WebService) {
	ws.Route(ws.GET("/devops/{devops}/credentials").
		To(handler.ListCredential).
		Param(ws.PathParameter("devops", "devops name")).
		Param(ws.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. ascending=false").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Doc("list the credentials of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.POST("/devops/{devops}/credentials").
		To(handler.CreateCredential).
		Param(ws.PathParameter("devops", "devops name")).
		Doc("create the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.GET("/devops/{devops}/credentials/{credential}").
		To(handler.GetCredential).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("credential", "pipeline name")).
		Doc("get the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1.Secret{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.PUT("/devops/{devops}/credentials/{credential}").
		To(handler.UpdateCredential).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("credential", "credential name")).
		Doc("put the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1.Secret{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.DELETE("/devops/{devops}/credentials/{credential}").
		To(handler.DeleteCredential).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("credential", "credential name")).
		Doc("delete the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1.Secret{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}))
}

func registerRoutersForPipelines(handler *devopsHandler, ws *restful.WebService) {
	ws.Route(ws.GET("/devops/{devops}/pipelines").
		To(handler.ListPipeline).
		Param(ws.PathParameter("devops", "devops name")).
		Param(ws.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Doc("list the pipelines of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.POST("/devops/{devops}/pipelines").
		To(handler.CreatePipeline).
		Param(ws.PathParameter("devops", "devops name")).
		Doc("create the pipeline of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.GET("/devops/{devops}/pipelines/{pipeline}").
		To(handler.GetPipeline).
		Operation("getPipelineByName").
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("pipeline", "pipeline name")).
		Doc("get the pipeline of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.PUT("/devops/{devops}/pipelines/{pipeline}").
		To(handler.UpdatePipeline).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("pipeline", "pipeline name")).
		Doc("put the pipeline of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.DELETE("/devops/{devops}/pipelines/{pipeline}").
		To(handler.DeletePipeline).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("pipeline", "pipeline name")).
		Doc("delete the pipeline of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}))
}

func registerRoutersForWorkspace(handler *devopsHandler, ws *restful.WebService) {
	ws.Route(ws.GET("/workspaces/{workspace}/devops").
		To(handler.ListDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).Doc("List the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.POST("/workspaces/{workspace}/devops").
		To(handler.CreateDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/devops/{devops}").
		To(handler.GetDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.QueryParameter("generateName", "use '{devops}` as a generatName if 'generateName=true', or as a regular name")).
		Doc("Get the devops project of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.PUT("/workspaces/{workspace}/devops/{devops}").
		To(handler.UpdateDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "project name")).
		Doc("Put the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.DELETE("/workspaces/{workspace}/devops/{devops}").
		To(handler.DeleteDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "project name")).
		Doc("Get the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))
}
