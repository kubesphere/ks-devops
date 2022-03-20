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
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/kapis"
	resourcev1alpha3 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *handler) handleQueryClusterTemplates(request *restful.Request, response *restful.Response) {
	commonQuery := query.ParseQueryParameter(request)
	kapis.ResponseWriter{Response: response}.WriteEntityOrError(h.queryClusterTemplates(commonQuery))
}

func (h *handler) handleRenderClusterTemplate(request *restful.Request, response *restful.Response) {
	templateName := request.PathParameter(ClusterTemplatePathParameter.Data().Name)

	//var parameters []Parameter
	var renderBody RenderBody
	if err := request.ReadEntity(&renderBody); err != nil && err != io.EOF {
		kapis.HandleError(request, response, err)
		return
	}

	kapis.ResponseWriter{Response: response}.WriteEntityOrError(h.renderClusterTemplate(templateName, renderBody.Parameters))
}

func (h *handler) queryClusterTemplates(commonQuery *query.Query) (*api.ListResult, error) {
	templateList := &v1alpha3.ClusterTemplateList{}
	if err := h.List(context.Background(),
		templateList,
		client.MatchingLabelsSelector{
			Selector: commonQuery.Selector(),
		}); err != nil {
		return nil, err
	}
	return resourcev1alpha3.ToListResult(clusterTemplatesToObjects(templateList.Items), commonQuery, nil), nil
}

func (h *handler) getClusterTemplate(templateName string) (*v1alpha3.ClusterTemplate, error) {
	template := &v1alpha3.ClusterTemplate{}
	if err := h.Get(context.Background(), client.ObjectKey{Name: templateName}, template); err != nil {
		return nil, err
	}
	return template, nil
}

func (h *handler) renderClusterTemplate(templateName string, parameters []Parameter) (v1alpha3.TemplateObject, error) {
	template, err := h.getClusterTemplate(templateName)
	if err != nil {
		return nil, err
	}

	return render(template, parameters)
}

func clusterTemplatesToObjects(templates []v1alpha3.ClusterTemplate) []runtime.Object {
	var objects []runtime.Object
	for i := range templates {
		objects = append(objects, &templates[i])
	}
	return objects
}
