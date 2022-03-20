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

package template

import (
	"fmt"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/constants"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/common"
	"net/http"
)

var (
	// TemplatePathParameter is path parameter definition of template.
	TemplatePathParameter = restful.PathParameter("template", "Template name")
	// ClusterTemplatePathParameter is path parameter definition of ClusterTemplate.
	ClusterTemplatePathParameter = restful.PathParameter("clustertemplate", "Name of ClusterTemplate.")
)

// PageResult is the model of Template page result.
type PageResult struct {
	Items []v1alpha3.Template `json:"items"`
	Total int                 `json:"total"`
}

// RenderBody is the model of request body of render API.
type RenderBody struct {
	Parameters []Parameter `json:"parameters"`
}

// RegisterRoutes is for registering template routes into WebService.
func RegisterRoutes(service *restful.WebService, options *common.Options) {
	handler := newHandler(options)
	// Template
	service.Route(service.GET("/devops/{devops}/templates").
		To(handler.handleQuery).
		Param(common.DevopsPathParameter).
		Doc("Query templates for a DevOps Project.").
		Returns(http.StatusOK, api.StatusOK, PageResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsTemplateTag}))

	service.Route(service.GET("/devops/{devops}/templates/{template}").
		To(handler.handleGetTemplate).
		Param(common.DevopsPathParameter).
		Param(TemplatePathParameter).
		Doc("Get template").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.Template{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsTemplateTag}))

	service.Route(service.POST("/devops/{devops}/templates/{template}/render").
		To(handler.handleRenderTemplate).
		Param(common.DevopsPathParameter).
		Param(TemplatePathParameter).
		Reads(RenderBody{}).
		Doc(fmt.Sprintf("Render template and return render result into annotations (%s/%s) inside template", devops.GroupName, devops.RenderResultAnnoKey)).
		Returns(http.StatusOK, api.StatusOK, v1alpha3.Template{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsTemplateTag}))

	// ClusterTemplate
	service.Route(service.GET("/clustertemplates").
		To(handler.handleQueryClusterTemplates).
		Doc("Query cluster templates.").
		Returns(http.StatusOK, api.StatusOK, PageResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsClusterTemplateTag}))

	service.Route(service.POST("/clustertemplates/{clustertemplate}/render").
		To(handler.handleRenderClusterTemplate).
		Param(ClusterTemplatePathParameter).
		Reads(RenderBody{}).
		Doc("Render cluster template.").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.ClusterTemplate{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsClusterTemplateTag}))
}
