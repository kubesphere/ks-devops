// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package fluxcd

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"

	"github.com/kubesphere/ks-devops/pkg/api"
	"github.com/kubesphere/ks-devops/pkg/api/gitops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/config"
	"github.com/kubesphere/ks-devops/pkg/constants"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
)

var (
	// pathParameterApplication is a path parameter definition for application.
	pathParameterApplication = restful.PathParameter("application", "The application name")
	syncStatusQueryParam     = restful.QueryParameter("syncStatus", `Filter by sync status. Available values: "Unknown", "Synced" and "OutOfSync"`)
	healthStatusQueryParam   = restful.QueryParameter("healthStatus", `Filter by health status. Available values: "Unknown", "Progressing", "Healthy", "Suspended", "Degraded" and "Missing"`)
	cascadeQueryParam        = restful.QueryParameter("cascade",
		"Delete both the app and its resources, rather than only the application if cascade is true").
		DefaultValue("false").DataType("boolean")
)

// ApplicationPageResult is the model of page result of Applications.
type ApplicationPageResult struct {
	Items      []v1alpha1.Application `json:"items"`
	TotalItems int                    `json:"totalItems"`
}

// RegisterRoutes is for registering Argo CD Application routes into WebService.
func RegisterRoutes(service *restful.WebService, options *common.Options, fluxOption *config.FluxCDOption) {
	handler := newHandler(options, fluxOption)

	// public
	service.Route(service.GET("/namespaces/{namespace}/applications").
		To(handler.ApplicationList).
		Param(common.NamespacePathParameter).
		Param(common.PageQueryParameter).
		Param(common.LimitQueryParameter).
		Param(common.NameQueryParameter).
		Param(common.SortByQueryParameter).
		Param(common.AscendingQueryParameter).
		Param(syncStatusQueryParam).
		Param(healthStatusQueryParam).
		Doc("Search applications").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Returns(http.StatusOK, api.StatusOK, ApplicationPageResult{}))

	service.Route(service.GET("/namespaces/{namespace}/applications/{application}").
		To(handler.GetApplication).
		Param(common.NamespacePathParameter).
		Param(pathParameterApplication).
		Doc("Get a particular application").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.DELETE("/namespaces/{namespace}/applications/{application}").
		To(handler.DelApplication).
		Param(common.NamespacePathParameter).
		Param(pathParameterApplication).
		Param(cascadeQueryParam).
		Doc("Delete a particular application").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.PUT("/namespaces/{namespace}/applications/{application}").
		To(handler.UpdateApplication).
		Param(common.NamespacePathParameter).
		Param(pathParameterApplication).
		Reads(v1alpha1.Application{}).
		Doc("Update a particular application").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	// fluxcd
	service.Route(service.POST("/namespaces/{namespace}/applications").
		To(handler.createApplication).
		Param(common.NamespacePathParameter).
		Reads(v1alpha1.Application{}).
		Doc("Create an application").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.GET("/clusters").
		To(handler.getClusters).
		Doc("Get the clusters list").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Returns(http.StatusOK, api.StatusOK, []v1alpha1.ApplicationDestination{}))
}
