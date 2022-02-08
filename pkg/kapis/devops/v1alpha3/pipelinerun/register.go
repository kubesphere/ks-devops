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

package pipelinerun

import (
	"net/http"

	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/models/pipelinerun"

	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/client/devops"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
		Returns(http.StatusOK, api.StatusOK, v1alpha3.PipelineRunList{}))

	ws.Route(ws.POST("/namespaces/{namespace}/pipelines/{pipeline}/pipelineruns").
		To(handler.createPipelineRun).
		Doc("Create a PipelineRun for the specified pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the pipeline")).
		Param(ws.QueryParameter("branch", "The name of SCM reference, only for multi-branch pipeline")).
		Reads(devops.RunPayload{}).
		Returns(http.StatusCreated, api.StatusOK, v1alpha3.PipelineRun{}))

	ws.Route(ws.GET("/namespaces/{namespace}/pipelineruns/{pipelinerun}").
		To(handler.getPipelineRun).
		Doc("Get a PipelineRun for a specified pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the PipelineRun")).
		Param(ws.PathParameter("pipelinerun", "Name of the PipelineRun")).
		Returns(http.StatusOK, api.StatusOK, v1alpha3.PipelineRun{}))

	ws.Route(ws.GET("/namespaces/{namespace}/pipelineruns/{pipelinerun}/nodedetails").
		To(handler.getNodeDetails).
		Doc("Get node details including steps and approvable for a given Pipeline").
		Param(ws.PathParameter("namespace", "Namespace of the PipelineRun")).
		Param(ws.PathParameter("pipelinerun", "Name of the PipelineRun")).
		Returns(http.StatusOK, api.StatusOK, []pipelinerun.NodeDetail{}))
}
