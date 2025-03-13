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

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/api"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/client/devops"
	dclient "github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/constants"
	"github.com/kubesphere/ks-devops/pkg/models/pipelinerun"
)

// RegisterRoutes register routes into web service.
func RegisterRoutes(ws *restful.WebService, devopsClient dclient.Interface, c client.Client) {
	handler := newAPIHandler(apiHandlerOption{
		devopsClient: devopsClient,
		client:       c,
	})

	ws.Route(ws.GET("/namespaces/{namespace}/pipelines/{pipeline}/pipelineruns").
		To(handler.listPipelineRuns).
		Doc("Get all runs of the specified pipeline").
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags).
		Param(ws.PathParameter("namespace", "Namespace of the pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the pipeline")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter("branch", "The name of SCM reference")).
		Param(ws.QueryParameter("backward", "Backward compatibility for v1alpha2 API "+
			"`/devops/{devops}/pipelines/{pipeline}/runs`. By default, the backward is true. If you want to list "+
			"full data of PipelineRuns, just set the parameters to false.").
			DataType("boolean").
			DefaultValue("true")).
		Returns(http.StatusOK, api.StatusOK, v1alpha3.PipelineRunList{}))

	ws.Route(ws.POST("/namespaces/{namespace}/pipelines/{pipeline}/pipelineruns").
		To(handler.createPipelineRun).
		Doc("Create a PipelineRun for the specified pipeline").
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags).
		Param(ws.PathParameter("namespace", "Namespace of the pipeline")).
		Param(ws.PathParameter("pipeline", "Name of the pipeline")).
		Param(ws.QueryParameter("branch", "The name of SCM reference, only for multi-branch pipeline")).
		Reads(devops.RunPayload{}).
		Returns(http.StatusCreated, api.StatusOK, v1alpha3.PipelineRun{}))

	ws.Route(ws.GET("/namespaces/{namespace}/pipelineruns/{pipelinerun}").
		To(handler.getPipelineRun).
		Doc("Get a PipelineRun for a specified pipeline").
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags).
		Param(ws.PathParameter("namespace", "Namespace of the PipelineRun")).
		Param(ws.PathParameter("pipelinerun", "Name of the PipelineRun")).
		Returns(http.StatusOK, api.StatusOK, v1alpha3.PipelineRun{}))

	ws.Route(ws.GET("/namespaces/{namespace}/pipelineruns/{pipelinerun}/nodedetails").
		To(handler.getNodeDetails).
		Doc("Get node details including steps and approvable for a given Pipeline").
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags).
		Param(ws.PathParameter("namespace", "Namespace of the PipelineRun")).
		Param(ws.PathParameter("pipelinerun", "Name of the PipelineRun")).
		Returns(http.StatusOK, api.StatusOK, []pipelinerun.NodeDetail{}))

	// download PipelineRun artifact
	ws.Route(ws.GET("/namespaces/{namespace}/pipelineruns/{pipelinerun}/artifacts/download").
		Param(ws.PathParameter("namespace", "Namespace of the PipelineRun")).
		Param(ws.PathParameter("pipelinerun", "Name of the PipelineRun")).
		Param(ws.QueryParameter("filename", "artifact filename. e.g. artifact:v1.0.1")).
		To(handler.downloadArtifact).
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags))
}
