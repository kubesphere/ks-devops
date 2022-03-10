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
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/kapis/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *handler) getGitRepository(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	repoName := common.GetPathParameter(req, pathParameterGitRepository)

	repo := &v1alpha3.GitRepository{}
	err := h.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      repoName,
	}, repo)
	common.Response(req, res, repo, err)
}

func (h *handler) createGitRepositories(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	repo := &v1alpha3.GitRepository{}
	err := req.ReadEntity(repo)
	ctx := context.Background()

	if err == nil {
		repo.Namespace = namespace
		err = h.Create(ctx, repo)
	}
	common.Response(req, res, repo, err)
}

func (h *handler) listGitRepositories(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	repoList := &v1alpha3.GitRepositoryList{}
	err := h.List(context.Background(), repoList, client.InNamespace(namespace))
	common.Response(req, res, repoList, err)
}

func (h *handler) updateGitRepositories(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	repoName := common.GetPathParameter(req, pathParameterGitRepository)
	ctx := context.Background()

	patchRepo := &v1alpha3.GitRepository{}
	repo := &v1alpha3.GitRepository{}

	err := req.ReadEntity(patchRepo)
	if err == nil {
		err = h.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      repoName,
		}, repo)
		if err == nil {
			repo.Spec = patchRepo.Spec
			err = h.Update(ctx, repo)
		}
	}
	common.Response(req, res, repo, err)
}

func (h *handler) deleteGitRepositories(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	repoName := common.GetPathParameter(req, pathParameterGitRepository)
	ctx := context.Background()

	repo := &v1alpha3.GitRepository{}
	err := h.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      repoName,
	}, repo)
	if err == nil {
		err = h.Delete(ctx, repo)
	}
	common.Response(req, res, repo, err)
}
