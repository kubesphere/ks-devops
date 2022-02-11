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

package template

import (
	"context"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/kapis"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type handler struct {
	genericClient client.Client
}

func newHandler(options *kapisv1alpha1.Options) *handler {
	return &handler{
		genericClient: options.GenericClient,
	}
}

func (h *handler) handleQuery(request *restful.Request, response *restful.Response) {
	if data, err := h.queryTemplate(request); err != nil {
		kapis.HandleError(request, response, err)
	} else {
		_ = response.WriteEntity(data)
	}
}

func (h *handler) queryTemplate(request *restful.Request) (*api.ListResult, error) {
	devopsName := request.PathParameter(DevopsPathParameter.Data().Name)
	queryParam := query.ParseQueryParameter(request)
	templateList := &v1alpha1.TemplateList{}
	if err := h.genericClient.List(context.Background(), templateList, client.InNamespace(devopsName)); err != nil {
		return nil, err
	}
	return v1alpha3.ToListResult(toObjects(templateList.Items), queryParam, nil), nil
}

func (h *handler) handleGet(request *restful.Request, response *restful.Response) {
	if template, err := h.getTemplate(request); err != nil {
		kapis.HandleError(request, response, err)
	} else {
		_ = response.WriteEntity(template)
	}
}

func (h *handler) getTemplate(request *restful.Request) (*v1alpha1.Template, error) {
	devopsName := request.PathParameter(DevopsPathParameter.Data().Name)
	templateName := request.PathParameter(TemplatePathParameter.Data().Name)
	template := &v1alpha1.Template{}
	if err := h.genericClient.Get(context.Background(), client.ObjectKey{Namespace: devopsName, Name: templateName}, template); err != nil {
		return nil, err
	}
	return template, nil
}

func (h *handler) handleRender(request *restful.Request, response *restful.Response) {
	if template, err := h.renderTemplate(request); err != nil {
		kapis.HandleError(request, response, err)
	} else {
		_ = response.WriteEntity(template)
	}
}

func (h *handler) renderTemplate(request *restful.Request) (*v1alpha1.Template, error) {
	tmpl, err := h.getTemplate(request)
	if err != nil {
		return nil, err
	}

	// TODO render template using parameters

	tmplCopy := tmpl.DeepCopy()
	if tmplCopy.GetAnnotations() == nil {
		tmplCopy.SetAnnotations(map[string]string{})
	}

	tmplCopy.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey] = tmplCopy.Spec.Template

	// get parameter
	return tmplCopy, nil
}

func toObjects(templates []v1alpha1.Template) []runtime.Object {
	var objects []runtime.Object
	for i := range templates {
		objects = append(objects, &templates[i])
	}
	return objects
}
