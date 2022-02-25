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
	"context"
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/kapis/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *handler) applicationList(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)

	applicationList := &v1alpha1.ApplicationList{}
	err := h.List(context.Background(), applicationList, client.InNamespace(namespace))
	common.Response(req, res, applicationList, err)
}

func (h *handler) getApplication(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	name := common.GetPathParameter(req, pathParameterApplication)

	application := &v1alpha1.Application{}
	err := h.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, application)
	common.Response(req, res, application, err)
}

func (h *handler) delApplication(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	name := common.GetPathParameter(req, pathParameterApplication)

	application := &v1alpha1.Application{}
	err := h.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, application)
	if err == nil {
		err = h.Delete(context.Background(), application)
	}
	common.Response(req, res, application, err)
}

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

func (h *handler) updateApplication(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	name := common.GetPathParameter(req, pathParameterApplication)

	application := &v1alpha1.Application{}
	err := h.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, application)
	if err == nil {
		if err = req.ReadEntity(application); err == nil {
			err = h.Update(context.Background(), application)
		}
	}
	common.Response(req, res, application, err)
}

func (h *handler) getClusters(req *restful.Request, res *restful.Response) {
	ctx := context.Background()

	secrets := &v1.SecretList{}
	err := h.List(ctx, secrets, client.MatchingLabels{
		"argocd.argoproj.io/secret-type": "cluster",
	})

	argoClusters := make([]v1alpha1.ApplicationDestination, 0)
	if err == nil && secrets != nil && len(secrets.Items) > 0 {
		for i := range secrets.Items {
			secret := secrets.Items[i]

			name := string(secret.Data["name"])
			server := string(secret.Data["server"])

			if name != "" && server != "" {
				argoClusters = append(argoClusters, v1alpha1.ApplicationDestination{
					Server: server,
					Name:   name,
				})
			}
		}
	}
	common.Response(req, res, argoClusters, err)
}

type handler struct {
	client.Client
}

func newHandler(options *common.Options) *handler {
	return &handler{
		Client: options.GenericClient,
	}
}
