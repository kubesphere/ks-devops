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

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/kubesphere/ks-devops/pkg/config"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/gitops"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/api"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/apiserver/runtime"
	dclient "github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/client/k8s"
	"github.com/kubesphere/ks-devops/pkg/constants"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/common"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/pipeline"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/pipelinerun"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/scm"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/steptemplate"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/template"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/webhook"
	"github.com/kubesphere/ks-devops/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

// TODO perhaps we can find a better way to declaim the permission needs of the apiserver
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=devopsprojects,verbs=get;list;update;delete;create;watch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;update;delete;create;watch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns,verbs=get;list;update;delete;create;watch

// GroupVersion describes CRD group and its version.
var GroupVersion = schema.GroupVersion{Group: api.GroupName, Version: "v1alpha3"}

// AddToContainer adds web service into container.
func AddToContainer(container *restful.Container, devopsClient dclient.Interface, k8sClient k8s.Client,
	client client.Client, runtimeCache cache.Cache, jenkins core.JenkinsCore, cfg *config.Config) (wss []*restful.WebService) {

	services := []*restful.WebService{
		runtime.NewWebService(v1alpha3.GroupVersion),
	}

	for _, service := range services {
		registerRoutes(cfg, devopsClient, k8sClient, client, runtimeCache, service)
		pipelinerun.RegisterRoutes(service, devopsClient, client)
		pipeline.RegisterRoutes(service, client)
		template.RegisterRoutes(service, &common.Options{
			GenericClient: client,
		})
		steptemplate.RegisterRoutes(service, &common.Options{
			GenericClient: client,
		})
		webhook.RegisterWebhooks(client, service, jenkins)
		container.Add(service)
	}
	return services
}

func registerRoutes(cfg *config.Config, devopsClient dclient.Interface, k8sClient k8s.Client, client client.Client, runtimeCache cache.Cache, ws *restful.WebService) {
	handler := newDevOpsHandler(client, devopsClient, k8sClient, runtimeCache)
	registerRoutersForCredentials(handler, ws)
	registerRoutersForPipelines(handler, ws)
	registerRoutersForWorkspace(handler, ws)
	scm.RegisterRoutersForSCM(client, ws)
	registerRoutersForCI(handler, ws)

	gitopsHandler := gitops.NewHandler(client, cfg.GitOpsOptions)
	gitops.RegisterRouters(ws, gitopsHandler)
}

func registerRoutersForCredentials(handler *devopsHandler, ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{devops}/credentials").
		To(handler.ListCredential).
		Param(ws.PathParameter("devops", "devops name")).
		Param(ws.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. ascending=false").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Doc("list the credentials of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsCredentialTags))

	ws.Route(ws.POST("/namespaces/{devops}/credentials").
		To(handler.CreateCredential).
		Param(ws.PathParameter("devops", "devops name")).
		Reads(corev1.Secret{}).
		Doc("create the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, corev1.Secret{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsCredentialTags))

	ws.Route(ws.GET("/namespaces/{devops}/credentials/{credential}").
		To(handler.GetCredential).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("credential", "pipeline name")).
		Doc("get the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, corev1.Secret{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsCredentialTags))

	ws.Route(ws.PUT("/namespaces/{devops}/credentials/{credential}").
		To(handler.UpdateCredential).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("credential", "credential name")).
		Reads(corev1.Secret{}).
		Doc("put the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, corev1.Secret{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsCredentialTags))

	ws.Route(ws.DELETE("/namespaces/{devops}/credentials/{credential}").
		To(handler.DeleteCredential).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("credential", "credential name")).
		Doc("delete the credential of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, corev1.Secret{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsCredentialTags))
}

func registerRoutersForPipelines(handler *devopsHandler, ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{devops}/pipelines").
		To(handler.ListPipeline).
		Param(ws.PathParameter("devops", "devops name")).
		Param(ws.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Doc("list the pipelines of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags))

	ws.Route(ws.POST("/namespaces/{devops}/pipelines").
		To(handler.CreatePipeline).
		Param(ws.PathParameter("devops", "devops name")).
		Reads(v1alpha3.Pipeline{}).
		Doc("create the pipeline of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags))

	ws.Route(ws.GET("/namespaces/{devops}/pipelines/{pipeline}").
		To(handler.GetPipeline).
		Operation("getPipelineByName").
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("pipeline", "pipeline name")).
		Doc("get the pipeline of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags))

	ws.Route(ws.PUT("/namespaces/{devops}/pipelines/{pipeline}").
		To(handler.UpdatePipeline).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("pipeline", "pipeline name")).
		Doc("put the pipeline of the specified devops for the current user").
		Reads(v1alpha3.Pipeline{}).
		Returns(http.StatusOK, api.StatusOK, v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags))

	ws.Route(ws.PUT("/namespaces/{devops}/pipelines/{pipeline}/jenkinsfile").
		To(handler.UpdateJenkinsfile).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("pipeline", "pipeline name")).
		Param(ws.QueryParameter("mode", "the mode(json or raw) that you expect to update the Jenkinsfile")).
		Reads(GenericPayload{}, "The Jenkinsfile content should be in the 'data' field").
		Doc("Update the Jenkinsfile of a Pipeline").
		Returns(http.StatusOK, api.StatusOK, GenericResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsJenkinsTags))

	ws.Route(ws.DELETE("/namespaces/{devops}/pipelines/{pipeline}").
		To(handler.DeletePipeline).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.PathParameter("pipeline", "pipeline name")).
		Doc("delete the pipeline of the specified devops for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.Pipeline{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags))
}

func registerRoutersForWorkspace(handler *devopsHandler, ws *restful.WebService) {
	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(handler.ListDevOpsProject).
		Operation("ListDevOpsProject").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).Doc("List the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsProjectTags))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}/namespaces").
		To(handler.ListDevOpsProject).
		Operation("ListDevOpsProjectForWorkspaceMember").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspace member name")).
		Param(ws.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).Doc("List the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsProjectTags))

	ws.Route(ws.POST("/workspaces/{workspace}/namespaces").
		To(handler.CreateDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Reads(v1alpha3.DevOpsProject{}).
		Doc("Create the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsProjectTags))

	ws.Route(ws.GET("/workspaces/{workspace}/namespaces/{devops}").
		To(handler.GetDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "project name")).
		Param(ws.QueryParameter("generateName", "use '{devops}` as a generatName if 'generateName=true', or as a regular name")).
		Doc("Get the devops project of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsProjectTags))

	ws.Route(ws.PUT("/workspaces/{workspace}/namespaces/{devops}").
		To(handler.UpdateDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "project name")).
		Reads(v1alpha3.DevOpsProject{}).
		Doc("Put the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsProjectTags))

	ws.Route(ws.DELETE("/workspaces/{workspace}/namespaces/{devops}").
		To(handler.DeleteDevOpsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "project name")).
		Doc("Get the devopsproject of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsProjectTags))

}

func registerRoutersForCI(handler *devopsHandler, ws *restful.WebService) {
	ws.Route(ws.GET("/ci/nodelabels").
		To(handler.getJenkinsLabels).
		Doc("Get the all labels of the Jenkins").
		Returns(http.StatusOK, api.StatusOK, GenericArrayResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsJenkinsTags))
}
