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

package clustertemplate

import (
	"context"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/kapis"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type handler struct {
	genericClient client.Client
}

func (h *handler) handleQuery(request *restful.Request, response *restful.Response) {
	commonQuery := query.ParseQueryParameter(request)
	if listResult, err := h.query(commonQuery); err != nil {
		kapis.HandleError(request, response, err)
	} else {
		_ = response.WriteEntity(listResult)
	}
}

func (h *handler) query(commonQuery *query.Query) (*api.ListResult, error) {
	templateList := &v1alpha1.ClusterTemplateList{}
	if err := h.genericClient.List(context.Background(),
		templateList,
		client.MatchingLabelsSelector{
			Selector: commonQuery.Selector(),
		}); err != nil {
		return nil, err
	}
	return v1alpha3.ToListResult(toObjects(templateList.Items), commonQuery, nil), nil
}

func (h *handler) handleRender(request *restful.Request, response *restful.Response) {
	templateName := request.PathParameter(PathParam.Data().Name)

	if clusterTemplate, err := h.render(templateName); err != nil {
		kapis.HandleError(request, response, err)
	} else {
		_ = response.WriteEntity(clusterTemplate)
	}
}

func (h *handler) render(templateName string) (*v1alpha1.ClusterTemplate, error) {
	template := &v1alpha1.ClusterTemplate{}
	if err := h.genericClient.Get(context.Background(), client.ObjectKey{Name: templateName}, template); err != nil {
		return nil, err
	}
	templateCopy := template.DeepCopy()
	if templateCopy.GetAnnotations() == nil {
		templateCopy.SetAnnotations(map[string]string{})
	}

	// TODO render template using parameters
	templateCopy.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey] = templateCopy.Spec.Template
	return templateCopy, nil
}

func toObjects(templates []v1alpha1.ClusterTemplate) []runtime.Object {
	var objects []runtime.Object
	for i := range templates {
		objects = append(objects, &templates[i])
	}
	return objects
}

func newHandler(options *common.Options) *handler {
	if options == nil {
		return &handler{}
	}
	return &handler{
		genericClient: options.GenericClient,
	}
}
