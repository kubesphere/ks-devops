/*
Copyright 2022 The KubeSphere Authors.

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

package scm

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/git"
	"kubesphere.io/devops/pkg/kapis/common"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	pathParameterSCM          = restful.PathParameter("scm", "the SCM type")
	pathParameterOrganization = restful.PathParameter("organization",
		"The git provider organization. For a GitHub repository address: https://github.com/kubesphere/ks-devops. kubesphere is the organization name")
	pathParameterGitRepository    = restful.PathParameter("gitrepository", "The GitRepository customs resource")
	queryParameterSecret          = restful.QueryParameter("secret", "the secret name")
	queryParameterSecretNamespace = restful.QueryParameter("secretNamespace", "the namespace of target secret")
)

// RegisterRoutersForSCM registers the APIs which related to scm
func RegisterRoutersForSCM(k8sClient client.Client, ws *restful.WebService) {
	h := newHandler(k8sClient)
	registerSCMAPIs(ws, h)
	registerGitRepositoryAPIs(ws, h)
}

func registerSCMAPIs(ws *restful.WebService, h *handler) {
	ws.Route(ws.POST("/scms/{scm}/verify").
		To(h.verify).
		Param(pathParameterSCM).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Doc("verify the token of different git providers").
		Returns(http.StatusOK, api.StatusOK, git.VerifyResponse{}))

	ws.Route(ws.GET("/scms/{scm}/organizations").
		To(h.listOrganizations).
		Param(pathParameterSCM).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Doc("List all the readable organizations").
		Returns(http.StatusOK, api.StatusOK, []organization{}))

	ws.Route(ws.GET("/scms/{scm}/organizations/{organization}/repositories").
		To(h.listRepositories).
		Param(pathParameterSCM).
		Param(pathParameterOrganization).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Doc("List all the readable repositories").
		Returns(http.StatusOK, api.StatusOK, []repository{}))
}

func registerGitRepositoryAPIs(ws *restful.WebService, h *handler) {
	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories").
		To(h.listGitRepositories).
		Param(common.NamespacePathParameter).
		Doc("List all the GitRepositories").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.GitRepository{}))

	ws.Route(ws.POST("/namespaces/{namespace}/gitrepositories/").
		To(h.createGitRepositories).
		Param(common.NamespacePathParameter).
		Reads(v1alpha3.GitRepository{}).
		Doc("List all the GitRepositories").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.GitRepository{}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/gitrepositories/{gitrepository}").
		To(h.deleteGitRepositories).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Doc("Delete a GitRepository by name").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.GitRepository{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}").
		To(h.getGitRepository).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Doc("Get a GitRepository by name").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.GitRepository{}))

	ws.Route(ws.PUT("/namespaces/{namespace}/gitrepositories/{gitrepository}").
		To(h.updateGitRepositories).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Reads(v1alpha3.GitRepository{}).
		Doc("Update a GitRepositories").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.GitRepository{}))
}
