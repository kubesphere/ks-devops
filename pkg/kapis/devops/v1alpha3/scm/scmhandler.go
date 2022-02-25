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
	goscm "github.com/drone/go-scm/scm"
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/client/git"
	"kubesphere.io/devops/pkg/kapis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// handler holds all the API handlers of SCM
type handler struct {
	k8sClient client.Client
}

// NewHandler creates the instance of the SCM handler
func newHandler(k8sClient client.Client) *handler {
	return &handler{k8sClient: k8sClient}
}

// verify checks a SCM auth
func (h *handler) verify(request *restful.Request, response *restful.Response) {
	scm := request.PathParameter("scm")
	secretName := request.QueryParameter("secret")
	secretNamespace := request.QueryParameter("secretNamespace")

	_, code, err := h.getOrganizations(scm, secretName, secretNamespace, 1)

	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	verifyResult := git.VerifyResult(err, code)
	verifyResult.CredentialID = secretName
	_ = response.WriteAsJson(verifyResult)
}

func (h *handler) getOrganizations(scm, secret, namespace string, size int) (orgs []*goscm.Organization, code int, err error) {
	factory := git.NewClientFactory(scm, &v1.SecretReference{
		Namespace: namespace, Name: secret,
	}, h.k8sClient)

	var c *goscm.Client
	if c, err = factory.GetClient(); err == nil {
		var resp *goscm.Response

		if orgs, resp, err = c.Organizations.List(context.TODO(), goscm.ListOptions{Size: size, Page: 1}); err == nil {
			code = resp.Status
		} else {
			code = 101
		}
	} else {
		code = 100
	}
	return
}

func (h *handler) getRepositories(scm, org, secret, namespace string, size int) (repos []*goscm.Repository, code int, err error) {
	factory := git.NewClientFactory(scm, &v1.SecretReference{
		Namespace: namespace, Name: secret,
	}, h.k8sClient)

	var c *goscm.Client
	if c, err = factory.GetClient(); err == nil {
		var resp *goscm.Response

		var allRepos []*goscm.Repository
		if allRepos, resp, err = c.Repositories.List(context.TODO(), goscm.ListOptions{
			Page: 1,
			Size: size,
		}); err == nil {
			code = resp.Status

			for i := range allRepos {
				repo := allRepos[i]
				if repo.Namespace == org {
					repos = append(repos, repo)
				}
			}
		} else {
			code = 101
		}
	} else {
		code = 100
	}
	return
}

func (h *handler) listOrganizations(req *restful.Request, rsp *restful.Response) {
	scm := req.PathParameter("scm")
	secretName := req.QueryParameter("secret")
	secretNamespace := req.QueryParameter("secretNamespace")

	orgs, _, err := h.getOrganizations(scm, secretName, secretNamespace, 1000)
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

	repos, _, err := h.getRepositories(scm, organization, secretName, secretNamespace, 10000)
	if err != nil {
		kapis.HandleError(req, rsp, err)
	} else {
		_ = rsp.WriteEntity(transformRepositories(repos))
	}
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

func transformRepositories(repos []*goscm.Repository) (result []repository) {
	if repos != nil {
		result = make([]repository, len(repos))
		for i := range repos {
			result[i] = repository{
				Name:          repos[i].Name,
				DefaultBranch: repos[i].Branch,
			}
		}
	}
	return
}
