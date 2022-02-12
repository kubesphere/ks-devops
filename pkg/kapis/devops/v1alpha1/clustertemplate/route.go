package clustertemplate

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
)

func RegisterRoutes(service *restful.WebService, options *common.Options) {
	service.Route(service.GET("/clustertemplates"))
}
