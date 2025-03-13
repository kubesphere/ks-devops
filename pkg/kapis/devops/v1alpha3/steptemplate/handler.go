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

package steptemplate

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/ks-devops/pkg/api"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

func (h *handler) clusterStepTemplates(req *restful.Request, resp *restful.Response) {
	ctx := context.TODO()

	clusterStepTemplateList := &v1alpha3.ClusterStepTemplateList{}
	err := h.List(ctx, clusterStepTemplateList)

	queryParam := query.ParseQueryParameter(req)
	apiResult := resourcesv1alpha3.DefaultList(convertToObject(clusterStepTemplateList.Items), queryParam, api.DefaultCompareFunc, api.DefaultFilterFunc)

	writeResponse(apiResult, err, resp)
}

func convertToObject(prs []v1alpha3.ClusterStepTemplate) []runtime.Object {
	var result []runtime.Object
	for i := range prs {
		result = append(result, &prs[i])
	}
	return result
}

func (h *handler) getClusterStepTemplate(req *restful.Request, resp *restful.Response) {
	ctx := context.TODO()
	name := req.PathParameter(ClusterStepTemplate.Data().Name)

	clusterStepTemplate := &v1alpha3.ClusterStepTemplate{}
	err := h.Get(ctx, types.NamespacedName{Name: name}, clusterStepTemplate)
	writeResponse(clusterStepTemplate, err, resp)
}

func (h *handler) renderClusterStepTemplate(req *restful.Request, resp *restful.Response) {
	ctx := context.TODO()
	name := req.PathParameter(ClusterStepTemplate.Data().Name)

	var err error
	clusterStepTemplate := &v1alpha3.ClusterStepTemplate{}
	if err = h.Get(ctx, types.NamespacedName{Name: name}, clusterStepTemplate); err != nil {
		_ = resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	var secret *v1.Secret
	if secret, err = h.getSecret(req); err != nil {
		// TODO considering have logger output instead of the std output
		fmt.Printf("something goes wrong when getting secret, error: %v\n", err)
	}

	param := map[string]interface{}{}
	// get the parameters from request
	if err = req.ReadEntity(&param); err != nil {
		// TODO considering have logger output instead of the std output
		fmt.Printf("something goes wrong when getting parameter from request body, error: %v\n", err)
	}

	var output string
	output, err = clusterStepTemplate.Spec.Render(param, secret)
	writeResponse(map[string]string{
		"data": output,
	}, err, resp)
}

func (h *handler) getSecret(req *restful.Request) (secret *v1.Secret, err error) {
	secretName := req.QueryParameter(SecretNameQueryParameter.Data().Name)
	secretNamespace := req.QueryParameter(SecretNamespaceQueryParameter.Data().Name)
	if secretName != "" || secretNamespace != "" {
		secret = &v1.Secret{}
		err = h.Get(context.Background(), types.NamespacedName{
			Namespace: secretNamespace,
			Name:      secretName,
		}, secret)
	}
	return
}

func writeResponse(object interface{}, err error, resp *restful.Response) {
	if err == nil {
		_ = resp.WriteAsJson(object)
	} else {
		_ = resp.WriteError(http.StatusInternalServerError, err)
	}
}
