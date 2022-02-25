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
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/client/git"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	pathParameterSCM          = restful.PathParameter("scm", "the SCM type")
	pathParameterOrganization = restful.PathParameter("organization",
		"The git provider organization. For a GitHub repository address: https://github.com/kubesphere/ks-devops. kubesphere is the organization name")
	queryParameterSecret          = restful.QueryParameter("secret", "the secret name")
	queryParameterSecretNamespace = restful.QueryParameter("secretNamespace", "the namespace of target secret")
)

// RegisterRoutersForSCM registers the APIs which related to scm
func RegisterRoutersForSCM(k8sClient client.Client, ws *restful.WebService) {
	handler := newHandler(k8sClient)

	ws.Route(ws.POST("/scms/{scm}/verify").
		To(handler.verify).
		Param(pathParameterSCM).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Doc("verify the token of different git providers").
		Returns(http.StatusOK, api.StatusOK, git.VerifyResponse{}))

	ws.Route(ws.GET("/scms/{scm}/organizations").
		To(handler.listOrganizations).
		Param(pathParameterSCM).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Doc("List all the readable organizations").
		Returns(http.StatusOK, api.StatusOK, []organization{}))

	ws.Route(ws.GET("/scms/{scm}/organizations/{organization}/repositories").
		To(handler.listRepositories).
		Param(pathParameterSCM).
		Param(pathParameterOrganization).
		Param(queryParameterSecret).
		Param(queryParameterSecretNamespace).
		Doc("List all the readable repositories").
		Returns(http.StatusOK, api.StatusOK, []repository{}))
}
