package v1alpha1

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha1/template"
)

// AddToContainer adds web services into web service container.
func AddToContainer(container *restful.Container, options *common.Options) []*restful.WebService {
	var services []*restful.WebService
	services = append(services, runtime.NewWebService(v1alpha1.GroupVersion))
	services = append(services, runtime.NewWebServiceWithoutGroup(v1alpha1.GroupVersion))
	for _, service := range services {
		template.RegisterRoutes(service, options)
		container.Add(service)
	}
	return services
}
