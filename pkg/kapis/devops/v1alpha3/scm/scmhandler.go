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
	"context"
	"fmt"
	"github.com/emicklei/go-restful"
	goscm "github.com/jenkins-x/go-scm/scm"
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/client/git"
	"kubesphere.io/devops/pkg/kapis"
	"kubesphere.io/devops/pkg/kapis/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// handler holds all the API handlers of SCM
type handler struct {
	client.Client
}

// NewHandler creates the instance of the SCM handler
func newHandler(k8sClient client.Client) *handler {
	return &handler{
		Client: k8sClient,
	}
}

// verify checks a SCM auth
func (h *handler) verify(request *restful.Request, response *restful.Response) {
	scm := request.PathParameter("scm")
	secretName := request.QueryParameter("secret")
	secretNamespace := request.QueryParameter("secretNamespace")

	_, code, err := h.getOrganizations(scm, secretName, secretNamespace, 1, 1, false)

	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	verifyResult := git.VerifyResult(err, code)
	verifyResult.CredentialID = secretName
	_ = response.WriteAsJson(verifyResult)
}

func (h *handler) getOrganizations(scm, secret, namespace string, page, size int, includeUser bool) (orgs []*goscm.Organization, code int, err error) {
	factory := git.NewClientFactory(scm, &v1.SecretReference{
		Namespace: namespace, Name: secret,
	}, h.Client)

	ctx := context.Background()
	var c *goscm.Client
	if c, err = factory.GetClient(); err == nil {
		var resp *goscm.Response

		if orgs, resp, err = c.Organizations.List(ctx, goscm.ListOptions{Size: size, Page: page}); err == nil {
			code = resp.Status
		} else {
			code = 101
		}

		if includeUser {
			var user string
			if user, err = h.getCurrentUsername(c); err == nil {
				orgs = append(orgs, &goscm.Organization{
					Name:   user,
					Avatar: fmt.Sprintf("https://avatars.githubusercontent.com/%s", user),
				})
			}
		}
	} else {
		code = 100
	}
	return
}

func (h *handler) getRepositories(scm, org, secret, namespace string, page, size int) (repos []*goscm.Repository, code int, err error) {
	factory := git.NewClientFactory(scm, &v1.SecretReference{
		Namespace: namespace, Name: secret,
	}, h.Client)

	var c *goscm.Client
	if c, err = factory.GetClient(); err == nil {
		// check if the org name is an user account name
		var user string
		var listRepositoryFunc listRepository
		if user, err = h.getCurrentUsername(c); err == nil {
			if user == org {
				listRepositoryFunc = c.Repositories.ListUser
			} else {
				listRepositoryFunc = c.Repositories.ListOrganisation
			}
		} else {
			return
		}

		ctx := context.Background()
		if repos, _, err = listRepositoryFunc(ctx, org, goscm.ListOptions{
			Page: page,
			Size: size,
		}); err != nil {
			code = 101
		}
	} else {
		code = 100
	}
	return
}

type listRepository func(context.Context, string, goscm.ListOptions) ([]*goscm.Repository, *goscm.Response, error)

func (h *handler) listOrganizations(req *restful.Request, rsp *restful.Response) {
	scm := req.PathParameter("scm")
	secretName := req.QueryParameter("secret")
	secretNamespace := req.QueryParameter("secretNamespace")
	includeUser := common.GetQueryParameter(req, queryParameterIncludeUser) == "true"
	pageNumber, pageSize := common.GetPageParameters(req)

	orgs, _, err := h.getOrganizations(scm, secretName, secretNamespace, pageNumber, pageSize, includeUser)
	if err != nil {
		kapis.HandleError(req, rsp, err)
	} else {
		_ = rsp.WriteEntity(transformOrganizations(orgs))
	}
}

func (h *handler) listRepositories(req *restful.Request, rsp *restful.Response) {
	scm := req.PathParameter("scm")
	organization := req.PathParameter("organization")
	secretName := req.QueryParameter("secret")
	secretNamespace := req.QueryParameter("secretNamespace")
	pageNumber, pageSize := common.GetPageParameters(req)

	repos, _, err := h.getRepositories(scm, organization, secretName, secretNamespace, pageNumber, pageSize)
	if err != nil {
		kapis.HandleError(req, rsp, err)
	} else {
		_ = rsp.WriteEntity(transformRepositories(repos))
	}
}

func (h *handler) getCurrentUsername(c *goscm.Client) (username string, err error) {
	var user *goscm.User
	user, _, err = c.Users.Find(context.Background())
	if err != nil {
		return
	}

	username = user.Login
	return
}

func transformOrganizations(orgs []*goscm.Organization) (result []organization) {
	if orgs != nil {
		result = make([]organization, len(orgs))
		for i := range orgs {
			result[i] = organization{
				Name:   orgs[i].Name,
				Avatar: orgs[i].Avatar,
			}
		}
	}
	return
}

func transformRepositories(goSCMRepos []*goscm.Repository) (result *repositoryListResult) {
	result = &repositoryListResult{}
	if goSCMRepos != nil {
		repos := make([]repository, len(goSCMRepos))
		for i := range goSCMRepos {
			repos[i] = repository{
				Name:          goSCMRepos[i].Name,
				DefaultBranch: goSCMRepos[i].Branch,
			}
		}
		result.Repositories.Items = repos
	}
	return
}
