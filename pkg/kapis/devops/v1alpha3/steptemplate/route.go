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
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type handler struct {
	client.Client
}

var (
	// ClusterStepTemplate is path parameter definition of clustersteptemplate.
	ClusterStepTemplate = restful.PathParameter("clustersteptemplate", "The name of clustersteptemplate")
	// SecretNameQueryParameter is a query parameter of secret
	SecretNameQueryParameter = restful.QueryParameter("secret", "The name of a secret")
	// SecretNamespaceQueryParameter is a query parameter of the secret namespace
	SecretNamespaceQueryParameter = restful.QueryParameter("secretNamespace", "The namespace of a secret")
)

// TODO perhaps we can find a better way to declaim the permission needs of the apiserver
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=clustersteptemplates,verbs=get;list;update;delete;create;watch

// RegisterRoutes registry the handlers of the stepTemplates
func RegisterRoutes(service *restful.WebService, options *common.Options) {
	h := &handler{options.GenericClient}
	service.Route(service.GET("/clustersteptemplates").
		To(h.clusterStepTemplates).
		Doc("Return the cluster level stepTemplate list"))
	service.Route(service.GET("/clustersteptemplates/{clustersteptemplate}").
		To(h.getClusterStepTemplate).
		Param(ClusterStepTemplate).
		Doc("Return a specific ClusterStepTemplate"))
	service.Route(service.POST("/clustersteptemplates/{clustersteptemplate}/render").
		To(h.renderClusterStepTemplate).
		Param(ClusterStepTemplate).
		Param(SecretNameQueryParameter).
		Param(SecretNamespaceQueryParameter).
		Reads(map[string]string{}, "The parameters of the ClusterStepTemplate").
		Doc("Render a specific ClusterStepTemplate, then return it"))
}
