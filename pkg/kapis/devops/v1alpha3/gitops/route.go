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

package gitops

import (
	"encoding/json"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/ks-devops/pkg/api"
	"github.com/kubesphere/ks-devops/pkg/constants"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=gitrepositories,verbs=get;list;update;delete;create;watch;patch

var (
	pathParameterSCM          = restful.PathParameter("scm", "the SCM type")
	pathParameterOrganization = restful.PathParameter("organization",
		"The git provider organization. For a GitHub repository address: https://github.com/kubesphere/ks-devops. kubesphere is the organization name")
	pathParameterGitRepository   = restful.PathParameter("gitrepository", "The GitRepository customs resource").DataType("string")
	pathParameterBranch          = restful.PathParameter("branch", "The branch of git repository").DataType("string")
	pathParameterCommit          = restful.PathParameter("commit", "The commit hash").DataType("string")
	pathParameterFile            = restful.PathParameter("file", "base64 encoded file path").DataType("string")
	queryParameterFile           = restful.QueryParameter("file", "the relative path of the file or directory in the git repository").DataType("string")
	queryParameterMessage        = restful.QueryParameter("message", "the commit message").DataType("string")
	queryParameterWithContent    = restful.QueryParameter("withContent", "whether get the base64 encoded content of the file").DataType("boolean")
	queryParameterWithHead       = restful.QueryParameter("withHead", "whether get the head reference").DataType("boolean")
	queryParameterWithLastCommit = restful.QueryParameter("withLastCommit", "whether get the last commit for file").DataType("boolean")
)

func RegisterRouters(ws *restful.WebService, h Handler) {
	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/configs/default").
		To(h.GetConfig).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Doc("get git config (.git/config)").
		Returns(http.StatusOK, api.StatusOK, GetConfigOutput{}))

	ws.Route(ws.PUT("/namespaces/{namespace}/gitrepositories/{gitrepository}/configs/default").
		To(h.UpdateConfig).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Reads(UpdateConfigInput{}).
		Doc("update git config").
		Returns(http.StatusOK, api.StatusOK, UpdateConfigOutput{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches").
		To(h.ListBranches).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(queryParameterWithHead).
		Doc("list branches for GitRepository like 'git branch -r'").
		Returns(http.StatusOK, api.StatusOK, ListBranchesOutput{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}").
		To(h.GetBranch).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Doc("get branch info").
		Returns(http.StatusOK, api.StatusOK, BranchInfo{}))

	ws.Route(ws.POST("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/checkouts").
		To(h.CheckOutBranch).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Reads(json.RawMessage{}).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Doc("checkout branch").
		Returns(http.StatusOK, api.StatusOK, CheckOutBranchOutput{}))

	ws.Route(ws.POST("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/pulls").
		To(h.CleanAndPullBranch).
		Reads(json.RawMessage{}).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Doc("clean and pull branch").
		Returns(http.StatusOK, api.StatusOK, CleanAndPullOutput{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/commits").
		To(h.ListCommits).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(queryParameterFile).
		Doc("list commits like 'git log'").
		Returns(http.StatusOK, api.StatusOK, ListCommitsOutput{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/commits/{commit}").
		To(h.GetCommit).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterCommit).
		Doc("get commit like 'git show <hash>'").
		Returns(http.StatusOK, api.StatusOK, CommitInfo{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/files").
		To(h.ListFiles).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(queryParameterFile.Required(true).Description("must be a directory and ends with '/'")).
		Param(queryParameterWithContent).
		Param(queryParameterWithLastCommit).
		Doc("list files in the branch").
		Returns(http.StatusOK, api.StatusOK, ListFilesOutput{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/files/{file}").
		To(h.GetFile).
		Operation("GetFileFromBranch").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Param(pathParameterFile).
		Param(queryParameterWithContent).
		Doc("get file from branch").
		Returns(http.StatusOK, api.StatusOK, FileInfo{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/rawfiles/{file}").
		To(h.DownloadFile).
		Operation("DownloadFileFromBranch").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Param(pathParameterFile).
		Doc("download file from branch").
		Returns(http.StatusOK, api.StatusOK, []byte{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/commits/{commit}/files/{file}").
		To(h.GetFile).
		Operation("GetFileFromCommit").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterCommit).
		Param(pathParameterFile).
		Param(queryParameterWithContent).
		Doc("get file from commit").
		Returns(http.StatusOK, api.StatusOK, FileInfo{}))

	ws.Route(ws.GET("/namespaces/{namespace}/gitrepositories/{gitrepository}/commits/{commit}/rawfiles/{file}").
		To(h.DownloadFile).
		Operation("DownloadFileFromCommit").
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterCommit).
		Param(pathParameterFile).
		Doc("download file from commit").
		Returns(http.StatusOK, api.StatusOK, []byte{}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/files").
		To(h.DeleteFiles).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Param(queryParameterFile.AllowMultiple(true).Required(true)).
		Param(queryParameterMessage.Required(true)).
		Doc("delete files in the branch").
		Returns(http.StatusOK, api.StatusOK, DeleteFilesOutput{}))

	ws.Route(ws.POST("/namespaces/{namespace}/gitrepositories/{gitrepository}/branches/{branch}/files").
		To(h.AddFiles).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(pathParameterBranch).
		Reads(AddFilesInput{}).
		Doc("add files in the branch, file content can either be passed by the request payload, or use the uploaded files by specify uploaded=true").
		Notes("when unpack is true, only the first file of files will be used and it must be a tar gzip archive.").
		Returns(http.StatusOK, api.StatusOK, AddFilesOutput{}))

	ws.Route(ws.POST("/namespaces/{namespace}/gitrepositories/{gitrepository}/uploads").
		To(h.UploadFiles).
		Metadata(restfulspec.KeyOpenAPITags, constants.GitOpsTags).
		Consumes("multipart/form-data").
		Param(common.NamespacePathParameter).
		Param(pathParameterGitRepository).
		Param(restful.FormParameter("fileX", "the file of fileX, X starts from 0").DataType("file")).
		Param(restful.FormParameter("fileX_name", "the target path in repository, X starts from 0").DataType("string")).
		Doc("upload files by multipart/form-data, up to 10 files can be passed in the form, you must call AddFiles later to commit the files").
		Returns(http.StatusOK, api.StatusOK, UploadFilesOutput{}))
}
