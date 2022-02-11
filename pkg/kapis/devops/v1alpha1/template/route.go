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
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/constants"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"net/http"
)

var (
	// DevopsPathParameter is a path parameter definition for devops.
	DevopsPathParameter = restful.PathParameter("devops", "DevOps project name")
	// TemplatePathParameter is a path parameter definition for template.
	TemplatePathParameter = restful.PathParameter("template", "Template name")
)

// RegisterRoutes is for registering template routes into WebService.
func RegisterRoutes(service *restful.WebService, options *kapisv1alpha1.Options) {
	handler := newHandler(options)
	service.Route(service.GET("/devops/{devops}/templates").
		To(handler.handleQuery).
		Param(DevopsPathParameter).
		Doc("Query templates for a DevOps Project.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsTemplateTag}))

	service.Route(service.GET("/devops/{devops}/templates/{template}").
		To(handler.handleGet).
		Param(DevopsPathParameter).
		Param(TemplatePathParameter).
		Doc("Get template").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Template{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsTemplateTag}))

	service.Route(service.POST("/devops/{devops}/templates/{template}/render").
		To(handler.handleRender).
		Param(DevopsPathParameter).
		Param(TemplatePathParameter).
		Doc(fmt.Sprintf("Render template and return render result into annotations (%s/%s) inside template", devops.GroupName, devops.RenderResultAnnoKey)).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Template{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsTemplateTag}))
}
