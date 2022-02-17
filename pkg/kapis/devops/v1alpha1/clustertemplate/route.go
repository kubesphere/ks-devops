package clustertemplate

import (
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/constants"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"net/http"
)

// PageResult is the model of a ClusterTemplate page result.
type PageResult struct {
	Items []v1alpha1.ClusterTemplate `json:"items"`
	Total int                        `json:"total"`
}

var (
	// PathParam is path parameter definition of ClusterTemplate.
	PathParam = restful.PathParameter("clustertemplate", "Name of ClusterTemplate.")
)

// RegisterRoutes register routes for ClusterTemplate.
func RegisterRoutes(service *restful.WebService, options *common.Options) {
	handler := newHandler(options)
	service.Route(service.GET("/clustertemplates").
		To(handler.handleQuery).
		Doc("Query cluster templates.").
		Returns(http.StatusOK, api.StatusOK, PageResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsClusterTemplateTag}))

	service.Route(service.POST("/clustertemplates/{clustertemplate}/render").
		To(handler.handleRender).
		Param(PathParam).
		Doc("Render cluster template.").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.ClusterTemplate{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsClusterTemplateTag}))
}
