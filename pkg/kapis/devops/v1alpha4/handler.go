package v1alpha4

import (
	"context"
	"encoding/json"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"kubesphere.io/devops/pkg/apiserver/query"
	v1alpha32 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
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
		var transformFunc = v1alpha32.NoTransformFunc()
		if backward {
			transformFunc = backwardTransform()
		}

		apiResult := v1alpha32.DefaultList(convertPipelineRunsToObject(prs.Items), queryParam, v1alpha32.DefaultCompare(), v1alpha32.DefaultFilter(), transformFunc)
		_ = response.WriteAsJson(apiResult)
	}
}

// backwardTransform transforms PipelineRun into JSON raw message of Jenkins run status.
func backwardTransform() v1alpha32.TransformFunc {
	return func(object runtime.Object) interface{} {
		pr := object.(*v1alpha4.PipelineRun)
		runStatusJSON := pr.Annotations[v1alpha4.JenkinsPipelineRunStatusKey]
		rawRunStatus := json.RawMessage{}
		// the error will never happen due to raw message won't be nil, so we ignore the error explicitly
		_ = rawRunStatus.UnmarshalJSON([]byte(runStatusJSON))
		return rawRunStatus
	}
}

func buildLabelSelector(queryParam *query.Query, pipelineName, branchName string) (labels.Selector, error) {
	labelSelector := queryParam.Selector()
	rq, err := labels.NewRequirement(v1alpha4.PipelineNameLabelKey, selection.Equals, []string{pipelineName})
	if err != nil {
		// should never happen
		return nil, err
	}
	labelSelector = labelSelector.Add(*rq)
	if branchName != "" {
		rq, err = labels.NewRequirement(v1alpha4.SCMRefNameLabelKey, selection.Equals, []string{branchName})
		if err != nil {
			// should never happen
			return nil, err
		}
		labelSelector = labelSelector.Add(*rq)
	}
	return labelSelector, nil
}

func convertPipelineRunsToObject(prs []v1alpha4.PipelineRun) []runtime.Object {
	var result []runtime.Object
	for i := range prs {
		result = append(result, &prs[i])
	}
	return result
}
