package pipelinerun

import (
	"net/http"

	"kubesphere.io/devops/pkg/api/devops/v1alpha3"

	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/client/devops"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Option holds some useful tools when creating web service.
type Option struct {
	Container *restful.Container
	Client    client.Client
}

// RegisterRoutes register routes into web service.
func RegisterRoutes(ws *restful.WebService, c client.Client) {
	handler := newAPIHandler(apiHandlerOption{
		client: c,
	})
	ws.Route(ws.GET("/namespaces/{namespace}/pipelines/{pipeline}/pipelineruns").
		To(handler.listPipelineRuns).
		Doc("Get all runs of the specified pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the pipeline")).
		Param(ws.QueryParameter("branch", "The name of SCM reference")).
		Param(ws.QueryParameter("backward", "Backward compatibility for v1alpha2 API "+
			"`/devops/{devops}/pipelines/{pipeline}/runs`. By default, the backward is true. If you want to list "+
			"full data of PipelineRuns, just set the parameters to false.").
			DataType("bool").
			DefaultValue("true")).
		Returns(http.StatusOK, api.StatusOK, v1alpha3.PipelineRunList{}),
	)
	ws.Route(ws.POST("/namespaces/{namespace}/pipelines/{pipeline}/pipelineruns").
		To(handler.createPipelineRuns).
		Doc("Create a PipelineRun for the specified pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the pipeline")).
		Param(ws.QueryParameter("branch", "The name of SCM reference, only for multi-branch pipeline")).
		Reads(devops.RunPayload{}).
		Returns(http.StatusCreated, api.StatusOK, v1alpha3.PipelineRun{}),
	)
	ws.Route(ws.GET("/namespaces/{namespace}/pipelineruns/{pipelinerun}").
		To(handler.getPipelineRun).
		Doc("Get a PipelineRun for a specified pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the PipelineRun")).
		Param(ws.PathParameter("pipelinerun", "Name of the PipelineRun")).
		Returns(http.StatusOK, api.StatusOK, v1alpha3.PipelineRun{}))

}
