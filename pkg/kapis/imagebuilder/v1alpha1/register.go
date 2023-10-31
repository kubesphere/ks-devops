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
	//shbuild: shipwright-io/build
	shbuild "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
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

	ws.Route(ws.GET("imagebuildStrategies").
		To(handler.listImagebuildStrategies).
		Doc("Get all imagebuildStrategies").
		Param(ws.QueryParameter("language", "Kind of the language, one of nodejs/go/java/..")).
		Returns(http.StatusOK, api.StatusOK, shbuild.ClusterBuildStrategyList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.GET("/imagebuildStrategies/{imagebuildStrategy}").
		To(handler.getImagebuildStrategy).
		Doc("Get an imagebuildStrategy").
		Param(ws.PathParameter("imagebuildStrategy", "Name of the imagebuildStrategy")).
		Returns(http.StatusOK, api.StatusOK, shbuild.ClusterBuildStrategy{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.POST("/namespaces/{namespace}/imagebuilds").
		To(handler.createImagebuild).
		Doc("Create an imagebuild").
		Param(ws.PathParameter("namespace", "Namespace of the imagebuild")).
		Reads(shbuild.Build{}).
		Returns(http.StatusCreated, api.StatusOK, shbuild.Build{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.GET("/namespaces/{namespace}/imagebuilds").
		To(handler.listImagebuilds).
		Doc("Get all imagebuilds").
		Param(ws.PathParameter("namespace", "Namespace of the imagebuilds")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.GET("/namespaces/{namespace}/imagebuilds/{imagebuild}").
		To(handler.getImagebuild).
		Doc("Get an imagebuild").
		Param(ws.PathParameter("namespace", "Namespace of the imagebuild")).
		Param(ws.PathParameter("imagebuild", "Name of the imagebuild")).
		Returns(http.StatusOK, api.StatusOK, shbuild.Build{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/imagebuilds/{imagebuild}").
		To(handler.deleteImagebuild).
		Doc("Delete an imagebuild").
		Param(ws.PathParameter("namespace", "Namespace of the imagebuildRun")).
		Param(ws.PathParameter("imagebuild", "Name of the imagebuildRun")).
		Returns(http.StatusOK, api.StatusOK, shbuild.BuildRun{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.PUT("/namespaces/{namespace}/imagebuilds/{imagebuild}").
		To(handler.updateImagebuild).
		Doc("Update an imagebuild").
		Param(ws.PathParameter("namespace", "Namespace of the imagebuild")).
		Param(ws.PathParameter("imagebuild", "Name of the imagebuild")).
		Reads(shbuild.Build{}).
		Returns(http.StatusCreated, api.StatusOK, shbuild.Build{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.POST("/namespaces/{namespace}/imagebuilds/{imagebuild}/imagebuildRuns").
		To(handler.createImagebuildRun).
		Doc("Create an imagebuildRun").
		Param(ws.PathParameter("namespace", "Namespace of the imagebuildRun")).
		Param(ws.PathParameter("imagebuild", "Name of the imagebuildRun")).
		Returns(http.StatusCreated, api.StatusOK, shbuild.BuildRun{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.GET("/namespaces/{namespace}/imagebuilds/{imagebuild}/imagebuildRuns").
		To(handler.listImagebuildRuns).
		Doc("Get all imagebuildRuns of the imagebuild").
		Param(ws.PathParameter("namespace", "Namespace of imagebuildRuns")).
		Param(ws.PathParameter("imagebuild", "Imagebuild of imagebuildRuns")).
		Returns(http.StatusOK, api.StatusOK, shbuild.BuildRunList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.GET("/namespaces/{namespace}/imagebuildRuns/{imagebuildRun}").
		To(handler.getImagebuildRun).
		Doc("Get an imagebuildRun").
		Param(ws.PathParameter("namespace", "Namespace of imagebuildRun")).
		Param(ws.PathParameter("imagebuildRun", "Name of the imagebuildRun")).
		Returns(http.StatusOK, api.StatusOK, shbuild.BuildRun{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/imagebuildRuns/{imagebuildRun}").
		To(handler.deleteImagebuildRun).
		Doc("Delete an imagebuildRun").
		Param(ws.PathParameter("namespace", "Namespace of the imagebuildRun")).
		Param(ws.PathParameter("imagebuildRun", "Name of the imagebuildRun")).
		Returns(http.StatusOK, api.StatusOK, shbuild.BuildRun{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsImageBuilder}))

}
