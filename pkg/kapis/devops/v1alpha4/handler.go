package v1alpha4

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
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
		var transformFunc = resourcesV1alpha3.NoTransformFunc()
		var filterFunc = resourcesV1alpha3.DefaultFilter()
		if backward {
			transformFunc = backwardTransform()
			filterFunc = backwardFilter()
		}
		apiResult := resourcesV1alpha3.DefaultList(convertPipelineRunsToObject(prs.Items), queryParam, resourcesV1alpha3.DefaultCompare(), filterFunc, transformFunc)
		_ = response.WriteAsJson(apiResult)
	}
}

// backwardFilter is used to filter PipelineRuns that have started and have Jenkins run status.
func backwardFilter() resourcesV1alpha3.FilterFunc {
	return resourcesV1alpha3.DefaultFilter().And(func(object runtime.Object, filter query.Filter) bool {
		pr, ok := object.(*v1alpha4.PipelineRun)
		if !ok || pr == nil {
			return false
		}
		return pr.HasStarted() && pr.Annotations[v1alpha4.JenkinsPipelineRunStatusKey] != ""
	})
}

// backwardTransform transforms PipelineRun into JSON raw message of Jenkins run status.
func backwardTransform() resourcesV1alpha3.TransformFunc {
	return func(object runtime.Object) interface{} {
		return func(object runtime.Object) json.Marshaler {
			pr, ok := object.(*v1alpha4.PipelineRun)
			if !ok || pr == nil {
				return json.RawMessage("{}")
			}
			runStatusJSON := pr.Annotations[v1alpha4.JenkinsPipelineRunStatusKey]
			rawRunStatus := json.RawMessage(runStatusJSON)
			// check if the run status is a valid JSON
			valid := json.Valid(rawRunStatus)
			if !valid {
				klog.ErrorS(nil, "invalid Jenkins run status",
					"PipelineRun", fmt.Sprintf("%s/%s", pr.GetNamespace(), pr.GetName()), "runStatusJSON", runStatusJSON)
				rawRunStatus = []byte("{}")
			}
			return rawRunStatus
		}(object)
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
