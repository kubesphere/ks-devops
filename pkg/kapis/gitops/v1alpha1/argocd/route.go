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

package argocd

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/common"
	"net/http"
)

var (
	// pathParameterApplication is a path parameter definition for application.
	pathParameterApplication = restful.PathParameter("application", "The application name")
)

// RegisterRoutes is for registering Argo CD Application routes into WebService.
func RegisterRoutes(service *restful.WebService, options *kapisv1alpha1.Options) {
	handler := newHandler(options)
	service.Route(service.GET("/namespaces/{namespace}/applications").
		To(handler.applicationList).
		Param(kapisv1alpha1.NamespacePathParameter).
		Doc("Get the application list").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.ApplicationList{}))

	service.Route(service.POST("/namespaces/{namespace}/applications").
		To(handler.createApplication).
		Param(kapisv1alpha1.NamespacePathParameter).
		Reads(v1alpha1.Application{}).
		Doc("Create an application").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.GET("/namespaces/{namespace}/applications/{application}").
		To(handler.getApplication).
		Param(kapisv1alpha1.NamespacePathParameter).
		Param(pathParameterApplication).
		Doc("Get a particular application").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.DELETE("/namespaces/{namespace}/applications/{application}").
		To(handler.delApplication).
		Param(kapisv1alpha1.NamespacePathParameter).
		Param(pathParameterApplication).
		Doc("Delete a particular application").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.PUT("/namespaces/{namespace}/applications/{application}").
		To(handler.updateApplication).
		Param(kapisv1alpha1.NamespacePathParameter).
		Param(pathParameterApplication).
		Reads(v1alpha1.Application{}).
		Doc("Update a particular application").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.GET("/clusters").
		To(handler.getClusters).
		Doc("Get the clusters list").
		Returns(http.StatusOK, api.StatusOK, []v1alpha1.ApplicationDestination{}))
}
