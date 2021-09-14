package v1alpha4

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/client/devops"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Option holds some useful tools when creating web service.
type Option struct {
	Container *restful.Container
	Client    client.Client
}

// AddToContainer registers web services with some options.
func AddToContainer(o Option) {
	// create API handler
	h := newAPIHandler(apiHandlerOption{
		client: o.Client,
	})
	// create web service without group
	ws := runtime.NewWebServiceWithoutGroup(v1alpha4.GroupVersion)
	addToContainer(o, ws, h)
	// create web service with group
	ws = runtime.NewWebService(v1alpha4.GroupVersion)
	addToContainer(o, ws, h)
}

func addToContainer(o Option, ws *restful.WebService, handler *apiHandler) {
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
		Returns(http.StatusOK, api.StatusOK, v1alpha4.PipelineRunList{}),
	)
	ws.Route(ws.POST("/namespaces/{namespace}/pipelines/{pipeline}/pipelineruns").
		To(handler.createPipelineRuns).
		Doc("Create a PipelineRun for the specified pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the pipeline")).
		Param(ws.QueryParameter("branch", "The name of SCM reference, only for multi-branch pipeline")).
		Reads(devops.RunPayload{}).
		Returns(http.StatusCreated, api.StatusOK, v1alpha4.PipelineRun{}),
	)
	o.Container.Add(ws)
}
