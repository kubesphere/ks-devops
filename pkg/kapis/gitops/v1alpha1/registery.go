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

package v1alpha1

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/kapis/common"
	"kubesphere.io/devops/pkg/kapis/gitops/v1alpha1/argocd"
)

// AddToContainer adds web services into web service container.
func AddToContainer(container *restful.Container, options *common.Options) []*restful.WebService {
	services := []*restful.WebService{
		runtime.NewWebService(v1alpha1.GroupVersion),
		runtime.NewWebServiceWithoutGroup(v1alpha1.GroupVersion),
	}
	for _, service := range services {
		argocd.RegisterRoutes(service, options)
		container.Add(service)
	}
	return services
}
