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
	"encoding/json"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/api"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/client/git"
	"github.com/kubesphere/ks-devops/pkg/constants"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
)

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=gitrepositories,verbs=get;list;update;delete;create;watch;patch

var (
	pathParameterSCM          = restful.PathParameter("scm", "the SCM type")
	pathParameterOrganization = restful.PathParameter("organization",
		"The git provider organization. For a GitHub repository address: https://github.com/kubesphere/ks-devops. kubesphere is the organization name")
	pathParameterGitRepository    = restful.PathParameter("gitrepository", "The GitRepository customs resource")
	queryParameterServer          = restful.QueryParameter("server", "The address of a self-hosted scm provider")
	queryParameterSecret          = restful.QueryParameter("secret", "the secret name")
	queryParameterSecretNamespace = restful.QueryParameter("secretNamespace", "the namespace of target secret")
	queryParameterIncludeUser     = restful.QueryParameter("includeUser", "Indicate if you want to include the current user")
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
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(pathParameterSCM).
		Param(queryParameterServer).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Doc("verify the token of different git providers").
		Reads(json.RawMessage{}).
		Returns(http.StatusOK, api.StatusOK, git.VerifyResponse{}))

	ws.Route(ws.GET("/scms/{scm}/organizations").
		To(h.listOrganizations).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(pathParameterSCM).
		Param(queryParameterServer).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Param(queryParameterIncludeUser.DataType("boolean").DefaultValue("true")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. ascending=false").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Doc("List all the readable organizations").
		Returns(http.StatusOK, api.StatusOK, []organization{}))

	ws.Route(ws.GET("/scms/{scm}/organizations/{organization}/repositories").
		To(h.listRepositories).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(pathParameterSCM).
		Param(queryParameterServer).
		Param(pathParameterOrganization).
		Param(queryParameterSecret.Required(true)).
		Param(queryParameterSecretNamespace.Required(true)).
		Param(common.PageNumberQueryParameter).
		Param(common.PageSizeQueryParameter).
		Doc("List all the readable Repositories").
		Returns(http.StatusOK, api.StatusOK, repositoryListResult{}))
}

func registerGitRepositoryAPIs(ws *restful.WebService, h *handler) {
	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories").
		To(h.listGitRepositories).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(common.NamespacePathParameter).
		Param(common.PageQueryParameter).
		Param(common.LimitQueryParameter).
		Doc("List all the GitRepositories").
		Returns(http.StatusOK, api.StatusOK, GitRepositoryPageResult{}))

	ws.Route(ws.POST("/namespaces/{namespace}/gitrepositories/").
		To(h.createGitRepositories).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(common.NamespacePathParameter).
		Reads(v1alpha3.GitRepository{}).
		Doc("List all the GitRepositories").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.GitRepository{}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/gitrepositories/{gitrepository}").
		To(h.deleteGitRepositories).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Doc("Delete a GitRepository by name").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.GitRepository{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}").
		To(h.getGitRepository).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Doc("Get a GitRepository by name").
		Returns(http.StatusOK, api.StatusOK, v1alpha3.GitRepository{}))

	ws.Route(ws.PUT("/namespaces/{namespace}/gitrepositories/{gitrepository}").
		To(h.updateGitRepositories).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsScmTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Reads(v1alpha3.GitRepository{}).
		Doc("Update a GitRepositories").
		Returns(http.StatusOK, api.StatusOK, []v1alpha3.GitRepository{}))
}
