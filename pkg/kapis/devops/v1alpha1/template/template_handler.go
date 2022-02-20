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
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/kapis"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *handler) handleQuery(request *restful.Request, response *restful.Response) {
	devopsName := request.PathParameter(kapisv1alpha1.DevopsPathParameter.Data().Name)
	commonQuery := query.ParseQueryParameter(request)

	kapis.ResponseWriter{Response: response}.WriteEntityOrError(h.queryTemplate(devopsName, commonQuery))
}

func (h *handler) queryTemplate(devopsName string, commonQuery *query.Query) (*api.ListResult, error) {
	templateList := &v1alpha1.TemplateList{}
	if err := h.genericClient.List(context.Background(),
		templateList,
		client.InNamespace(devopsName),
		client.MatchingLabelsSelector{
			Selector: commonQuery.Selector(),
		}); err != nil {
		return nil, err
	}
	return v1alpha3.ToListResult(templatesToObjects(templateList.Items), commonQuery, nil), nil
}

func (h *handler) handleGetTemplate(request *restful.Request, response *restful.Response) {
	devopsName := request.PathParameter(kapisv1alpha1.DevopsPathParameter.Data().Name)
	templateName := request.PathParameter(TemplatePathParameter.Data().Name)

	kapis.ResponseWriter{Response: response}.WriteEntityOrError(h.getTemplate(devopsName, templateName))
}

func (h *handler) getTemplate(devopsName, templateName string) (*v1alpha1.Template, error) {
	template := &v1alpha1.Template{}
	if err := h.genericClient.Get(context.Background(), client.ObjectKey{Namespace: devopsName, Name: templateName}, template); err != nil {
		return nil, err
	}
	return template, nil
}

func (h *handler) handleRenderTemplate(request *restful.Request, response *restful.Response) {
	devopsName := request.PathParameter(kapisv1alpha1.DevopsPathParameter.Data().Name)
	templateName := request.PathParameter(TemplatePathParameter.Data().Name)

	kapis.ResponseWriter{Response: response}.WriteEntityOrError(h.renderTemplate(devopsName, templateName))
}

func (h *handler) renderTemplate(devopsName, templateName string) (v1alpha1.TemplateObject, error) {
	tmpl, err := h.getTemplate(devopsName, templateName)
	if err != nil {
		return nil, err
	}
	return render(tmpl), nil
}

func templatesToObjects(templates []v1alpha1.Template) []runtime.Object {
	var objects []runtime.Object
	for i := range templates {
		objects = append(objects, &templates[i])
	}
	return objects
}
