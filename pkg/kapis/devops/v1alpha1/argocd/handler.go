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
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/kapis"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *handler) applicationList(req *restful.Request, res *restful.Response) {
	namespace := getPathParameter(req, kapisv1alpha1.DevopsPathParameter)

	applicationList := &v1alpha1.ApplicationList{}
	err := h.genericClient.List(context.Background(), applicationList, client.InNamespace(namespace))
	response(req, res, applicationList, err)
}

func (h *handler) getApplication(req *restful.Request, res *restful.Response) {
	namespace := getPathParameter(req, kapisv1alpha1.DevopsPathParameter)
	name := getPathParameter(req, pathParameterApplication)

	application := &v1alpha1.Application{}
	err := h.genericClient.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, application)
	response(req, res, application, err)
}

func (h *handler) delApplication(req *restful.Request, res *restful.Response) {
	namespace := getPathParameter(req, kapisv1alpha1.DevopsPathParameter)
	name := getPathParameter(req, pathParameterApplication)

	application := &v1alpha1.Application{}
	err := h.genericClient.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, application)
	if err == nil {
		err = h.genericClient.Delete(context.Background(), application)
	}
	response(req, res, application, err)
}

func (h *handler) createApplication(req *restful.Request, res *restful.Response) {
	var err error
	namespace := getPathParameter(req, kapisv1alpha1.DevopsPathParameter)

	application := &v1alpha1.Application{}
	if err = req.ReadEntity(application); err == nil {
		application.Namespace = namespace
		err = h.genericClient.Create(context.Background(), application)
	}

	response(req, res, application, err)
}

func (h *handler) updateApplication(req *restful.Request, res *restful.Response) {
	namespace := getPathParameter(req, kapisv1alpha1.DevopsPathParameter)
	name := getPathParameter(req, pathParameterApplication)

	application := &v1alpha1.Application{}
	err := h.genericClient.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, application)
	if err == nil {
		if err = req.ReadEntity(application); err == nil {
			err = h.genericClient.Update(context.Background(), application)
		}
	}
	response(req, res, application, err)
}

func getPathParameter(req *restful.Request, param *restful.Parameter) string {
	return req.PathParameter(param.Data().Name)
}

func response(req *restful.Request, res *restful.Response, object interface{}, err error) {
	if err != nil {
		kapis.HandleError(req, res, err)
	} else {
		_ = res.WriteEntity(object)
	}
}

type handler struct {
	genericClient client.Client
}

func newHandler(options *kapisv1alpha1.Options) *handler {
	return &handler{
		genericClient: options.GenericClient,
	}
}
