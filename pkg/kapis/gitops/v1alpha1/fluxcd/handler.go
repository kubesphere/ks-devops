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

package fluxcd

import (
	"context"
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/kapis/common"
	"kubesphere.io/devops/pkg/kapis/gitops/v1alpha1/gitops"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *handler) createApplication(req *restful.Request, res *restful.Response) {
	var err error
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	application := &v1alpha1.Application{}
	if err = req.ReadEntity(application); err == nil {
		application.Namespace = namespace
		err = h.Create(context.Background(), application)
	}
	common.Response(req, res, application, err)
}

func (h *handler) getClusters(req *restful.Request, res *restful.Response) {
	ctx := context.Background()

	secrets := &v1.SecretList{}
	err := h.List(ctx, secrets, client.MatchingLabels{
		"app.kubernetes.io/managed-by": "fluxcd-controller",
	})

	// Compatibility with the front end for now
	//TODO: change to FluxApplicationDestination
	fluxClusters := []v1alpha1.ApplicationDestination{{
		Name: "in-cluster",
	}}
	for _, s := range secrets.Items {
		fluxClusters = append(fluxClusters, v1alpha1.ApplicationDestination{
			Name: s.Name,
		})
	}
	common.Response(req, res, fluxClusters, err)
}

type handler struct {
	*gitops.Handler
}

func newHandler(options *common.Options, fluxOption *config.FluxCDOption) *handler {
	return &handler{
		gitops.NewHandler(options),
	}
}
