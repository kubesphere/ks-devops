package v1alpha4

import (
	"context"
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"kubesphere.io/devops/pkg/apiserver/query"
	resourcesV1alpha3 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type apiHandlerOption struct {
	client client.Client
}

// apiHandler contains functions to handle coming request and give a response.
type apiHandler struct {
	apiHandlerOption
}

// newAPIHandler creates an APIHandler.
func newAPIHandler(o apiHandlerOption) *apiHandler {
	return &apiHandler{o}
}

func (h *apiHandler) listPipelineRuns() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		nsName := request.PathParameter("namespace")
		pipName := request.PathParameter("pipeline")
		branchName := request.QueryParameter("branch")
		backward, err := strconv.ParseBool(request.QueryParameter("backward"))
		if err != nil {
			// by default, we have to guarantee backward compatibility
			backward = true
		}

		queryParam := query.ParseQueryParameter(request)

		// validate the Pipeline
		var pip v1alpha3.Pipeline
		err = h.client.Get(context.Background(), client.ObjectKey{Namespace: nsName, Name: pipName}, &pip)
		if err != nil {
			api.HandleError(request, response, err)
			return
		}

		// build label selector
		labelSelector, err := buildLabelSelector(queryParam, pip.Name, branchName)
		if err != nil {
			api.HandleError(request, response, err)
			return
		}

		var prs v1alpha4.PipelineRunList
		// fetch PipelineRuns
		if err := h.client.List(context.Background(), &prs,
			client.InNamespace(pip.Namespace),
			client.MatchingLabelsSelector{Selector: labelSelector}); err != nil {
			api.HandleError(request, response, err)
			return
		}

		var listHandler resourcesV1alpha3.ListHandler
		if backward {
			listHandler = backwardListHandler{}
		}
		apiResult := resourcesV1alpha3.ToListResult(convertPipelineRunsToObject(prs.Items), queryParam, listHandler)
		_ = response.WriteAsJson(apiResult)
	}
}
