/*

  Copyright 2023 The KubeSphere Authors.

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

package v1alpha1

import (
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/constants"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"kubesphere.io/devops/pkg/api"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GroupVersion describes CRD group and its version.
var GroupVersion = schema.GroupVersion{Group: "builder.kubesphere.io", Version: "v1alpha1"}

// AddToContainer adds web service into container.
func AddToContainer(container *restful.Container, client client.Client, devopsClient devopsClient.Interface) (ws *restful.WebService) {
	ws = runtime.NewWebService(GroupVersion)
	registerRoutes(ws, devopsClient, client)
	container.Add(ws)
	return
}

// RegisterRoutes register routes into web service.
func registerRoutes(ws *restful.WebService, devopsClient devopsClient.Interface, c client.Client) {
	handler := newAPIHandler(apiHandlerOption{
		devopsClient: devopsClient,
		client:       c,
	})

	ws.Route(ws.POST("/namespaces/{namespace}/imageBuilds/{imageBuild}").
		To(handler.createImageBuild).
		Doc("Create an ImageBuild").
		Param(ws.PathParameter("namespace", "Namespace of the ImageBuild")).
		Param(ws.PathParameter("imageBuild", "Name of the ImageBuild")).
		Param(ws.QueryParameter("codeUrl", "URL for the code")).
		Param(ws.QueryParameter("languageKind", "Kind of the language")).
		Param(ws.QueryParameter("outputImageUrl", "Output image url")).
		Returns(http.StatusCreated, api.StatusOK, v1alpha1.Build{}))

	ws.Route(ws.GET("/namespaces/{namespace}/imageBuilds").
		To(handler.listImageBuilds).
		Doc("Get all imageBuilds").
		Param(ws.PathParameter("namespace", "Namespace of the imageBuilds")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.BuildList{}))

	ws.Route(ws.GET("/namespaces/{namespace}/imageBuilds/{imageBuild}").
		To(handler.getImageBuild).
		Doc("Get an ImageBuild").
		Param(ws.PathParameter("namespace", "Namespace of the ImageBuild")).
		Param(ws.PathParameter("imageBuild", "Name of the ImageBuild")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Build{}))

	ws.Route(ws.GET("/namespaces/{namespace}/imageBuilds/{imageBuild}").
		To(handler.deleteImageBuild).
		Doc("Delete an ImageBuild").
		Param(ws.PathParameter("namespace", "Namespace of the ImageBuildRun")).
		Param(ws.PathParameter("imageBuild", "Name of the ImageBuildRun")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.BuildRun{}))

	ws.Route(ws.POST("/namespaces/{namespace}/imageBuilds/{imageBuild}").
		To(handler.updateImageBuild).
		Doc("Update an ImageBuild").
		Param(ws.PathParameter("namespace", "Namespace of the ImageBuild")).
		Param(ws.PathParameter("imageBuild", "Name of the ImageBuild")).
		Param(ws.QueryParameter("codeUrl", "URL for the code")).
		Param(ws.QueryParameter("languageKind", "Kind of the language")).
		Param(ws.QueryParameter("outputImageUrl", "Output image url")).
		Returns(http.StatusCreated, api.StatusOK, v1alpha1.Build{}))

	ws.Route(ws.GET("imageBuildStrategies").
		To(handler.listImageBuildStrategies).
		Doc("Get all imageBuildStrategies").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.ClusterBuildStrategyList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.GET("/imageBuildStrategies/{imageBuildStrategy}").
		To(handler.getImageBuildStrategy).
		Doc("Get an imageBuildStrategy").
		Param(ws.PathParameter("imageBuildStrategy", "Name of the imageBuildStrategy")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.ClusterBuildStrategy{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.POST("/namespaces/{namespace}/ImageBuildRuns/{ImageBuildRun}").
		To(handler.createImageBuildRun).
		Doc("Create an ImageBuildRun").
		Param(ws.PathParameter("namespace", "Namespace of the ImageBuildRun")).
		Param(ws.PathParameter("imageBuildRun", "Name of the ImageBuildRun for imageBuild")).
		Param(ws.QueryParameter("imageBuild", "Name of Build for the buildRun")).
		Returns(http.StatusCreated, api.StatusOK, v1alpha1.BuildRun{}))

	ws.Route(ws.GET("/namespace/{namespace}/ImageBuildRuns").
		To(handler.listImageBuildRuns).
		Doc("Get all imageBuildRuns").
		Param(ws.PathParameter("namespace", "Namespace of imageBuildRuns")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.BuildRunList{}))

	ws.Route(ws.GET("/namespace/{namespace}/ImageBuildRuns/{ImageBuildRun}").
		To(handler.getImageBuildRun).
		Doc("Get an imageBuildRun").
		Param(ws.PathParameter("namespace", "Namespace of imageBuildRun")).
		Param(ws.PathParameter("imageBuildRun", "Name of the ImageBuildRun")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.BuildRun{}))

	ws.Route(ws.GET("/namespaces/{namespace}/ImageBuildRuns/{ImageBuildRun}").
		To(handler.deleteImageBuildRun).
		Doc("Delete an ImageBuildRun").
		Param(ws.PathParameter("namespace", "Namespace of the ImageBuildRun")).
		Param(ws.PathParameter("imageBuildRun", "Name of the ImageBuildRun")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.BuildRun{}))

}
