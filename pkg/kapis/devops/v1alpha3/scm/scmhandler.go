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
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	goscm "github.com/jenkins-x/go-scm/scm"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/client/git"
	"github.com/kubesphere/ks-devops/pkg/kapis"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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
	server := common.GetQueryParameter(request, queryParameterServer)

	code, err := 0, error(nil)
	switch scm {
	case "git":
		if server == "" {
			err := fmt.Errorf("server is required")
			kapis.HandleError(request, response, err)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
			return
		}
		code, err = h.checkRepoAccess(server, secretName, secretNamespace)

	default:
		_, code, err = h.getOrganizations(scm, server, secretName, secretNamespace, 1, 1, false)
	}

	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	verifyResult := git.VerifyResult(err, code)
	verifyResult.CredentialID = secretName
	_ = response.WriteAsJson(verifyResult)
}

func (h *handler) checkRepoAccess(repourl, secretName, secretNamespace string) (int, error) {
	user, token, privateKey, err := h.getTokenWithSShFromSecret(
		context.Background(), &v1.SecretReference{
			Namespace: secretNamespace, Name: secretName,
		})
	if err != nil {
		return http.StatusOK, err
	}

	storage := memory.NewStorage()
	remote := gogit.NewRemote(storage, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repourl},
	})

	auth, err := getAuthMethod(repourl, user, token, privateKey)
	if err != nil {
		return http.StatusBadRequest, err
	}

	_, err = remote.List(&gogit.ListOptions{Auth: auth})
	if err != nil {
		return http.StatusForbidden, handleGitError(err)
	}

	return http.StatusOK, nil
}

func (h *handler) getTokenWithSShFromSecret(ctx context.Context, ref *v1.SecretReference) (username, token string, privateKey []byte, err error) {
	secret := &v1.Secret{}
	if err := h.Get(ctx, types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}, secret); err != nil {
		return "", "", nil, err
	}

	switch secret.Type {
	case v1.SecretTypeBasicAuth, v1alpha3.SecretTypeBasicAuth:
		token = string(secret.Data[v1.BasicAuthPasswordKey])
		username = string(secret.Data[v1.BasicAuthUsernameKey])
	case v1.SecretTypeOpaque:
		token = string(secret.Data[v1.ServiceAccountTokenKey])
	case v1alpha3.SecretTypeSecretText:
		token = string(secret.Data[v1alpha3.SecretTextSecretKey])
		username = string(secret.Data[v1.BasicAuthUsernameKey])
	case v1alpha3.SecretTypeSSHAuth:
		privateKey = secret.Data[v1alpha3.SSHAuthPrivateKey]
		token = string(secret.Data[v1alpha3.SSHAuthPassphraseKey])
		username = string(secret.Data[v1alpha3.SSHAuthUsernameKey])
	}

	return username, token, privateKey, err
}

func (h *handler) getOrganizations(scm, server, secret, namespace string, page, size int, includeUser bool) (orgs []*goscm.Organization, code int, err error) {
	factory := git.NewClientFactory(scm, &v1.SecretReference{
		Namespace: namespace, Name: secret,
	}, h.Client)
	factory.Server = server

	ctx := context.Background()
	var c *goscm.Client
	if c, err = factory.GetClient(); err == nil {
		var resp *goscm.Response

		if orgs, resp, err = c.Organizations.List(ctx, &goscm.ListOptions{Size: size, Page: page}); err == nil {
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

func (h *handler) getRepositories(scm, server, org, secret, namespace string, page, size int) (repos []*goscm.Repository, code int, err error) {
	factory := git.NewClientFactory(scm, &v1.SecretReference{
		Namespace: namespace, Name: secret,
	}, h.Client)
	factory.Server = server

	var c *goscm.Client
	if c, err = factory.GetClient(); err == nil {
		// check if the org name is a user account name
		var user string
		var listRepositoryFunc listRepository
		if user, err = h.getCurrentUsername(c); err == nil {
			if user == org && !strings.HasPrefix(scm, "bitbucket") {
				listRepositoryFunc = func(ctx context.Context, s string, options *goscm.ListOptions) ([]*goscm.Repository, *goscm.Response, error) {
					return c.Repositories.List(ctx, options)
				}
			} else {
				listRepositoryFunc = c.Repositories.ListOrganisation
			}
		} else {
			return
		}

		ctx := context.Background()
		if repos, _, err = listRepositoryFunc(ctx, org, &goscm.ListOptions{
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

type listRepository func(context.Context, string, *goscm.ListOptions) ([]*goscm.Repository, *goscm.Response, error)

func (h *handler) listOrganizations(req *restful.Request, rsp *restful.Response) {
	scm := req.PathParameter("scm")
	secretName := req.QueryParameter("secret")
	secretNamespace := req.QueryParameter("secretNamespace")
	server := common.GetQueryParameter(req, queryParameterServer)
	includeUser := common.GetQueryParameter(req, queryParameterIncludeUser) == "true"
	pageNumber, pageSize := common.GetPageParameters(req)

	orgs, _, err := h.getOrganizations(scm, server, secretName, secretNamespace, pageNumber, pageSize, includeUser)
	if err != nil {
		kapis.HandleError(req, rsp, err)
	} else {
		_ = rsp.WriteEntity(transformOrganizations(orgs))
	}
}

func (h *handler) listRepositories(req *restful.Request, rsp *restful.Response) {
	scm := req.PathParameter("scm")
	server := common.GetQueryParameter(req, queryParameterServer)
	organization := req.PathParameter("organization")
	secretName := req.QueryParameter("secret")
	secretNamespace := req.QueryParameter("secretNamespace")
	pageNumber, pageSize := common.GetPageParameters(req)

	repos, _, err := h.getRepositories(scm, server, organization, secretName, secretNamespace, pageNumber, pageSize)
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
