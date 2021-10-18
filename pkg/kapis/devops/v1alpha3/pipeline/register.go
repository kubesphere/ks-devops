package pipeline

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/models/pipeline"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RegisterRoutes register routes into web service.
func RegisterRoutes(ws *restful.WebService, c client.Client) {
	handler := newAPIHandler(apiHandlerOption{
		client: c,
	})

	ws.Route(ws.GET("/namespaces/{namespace}/pipelines/{pipeline}/branches").
		To(handler.getBranches).
		Doc("Paging query branches of multi branch Pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the Pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the Pipeline")).
		Param(ws.PathParameter("filter", "Pipeline filter, allowed values: origin, pull_requests and no-folders")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}))

	ws.Route(ws.GET("/namespaces/{namespace}/pipelines/{pipeline}/branches/{branch}").
		To(handler.getBranch).
		Doc("Paging query branches of multi branch Pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the Pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the Pipeline")).
		Param(ws.PathParameter("branch", "Name of branch, tag or pull request")).
		Returns(http.StatusOK, api.StatusOK, pipeline.Branch{}))
}
