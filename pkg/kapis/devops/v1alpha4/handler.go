package v1alpha4

import (
	"context"
	"errors"
	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/client/devops"
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

func (h *apiHandler) createPipelineRuns() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		nsName := request.PathParameter("namespace")
		pipName := request.PathParameter("pipeline")
		branch := request.QueryParameter("branch")
		payload := devops.RunPayload{}
		if err := request.ReadEntity(&payload); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
		// validate the Pipeline
		var pipeline v1alpha3.Pipeline
		if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: nsName, Name: pipName}, &pipeline); err != nil {
			api.HandleError(request, response, err)
			return
		}

		var (
			scm *v1alpha4.SCM
			err error
		)
		if scm, err = getScm(&pipeline.Spec, branch); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}

		// create PipelineRun
		pr := createPipelineRun(&pipeline, &payload, scm)
		if err := h.client.Create(context.Background(), pr); err != nil {
			api.HandleError(request, response, err)
			return
		}

		_ = response.WriteEntity(pr)
	}
}

func getScm(ps *v1alpha3.PipelineSpec, branch string) (*v1alpha4.SCM, error) {
	var scm *v1alpha4.SCM
	if ps.Type == v1alpha3.MultiBranchPipelineType {
		if branch == "" {
			return nil, errors.New("missing branch name for running a multi-branch Pipeline")
		}
		// TODO validate if the branch dose exist
		// we can not determine what is reference type here. So we set reference name only for now
		scm = &v1alpha4.SCM{
			RefName: branch,
			RefType: "",
		}
	}
	return scm, nil
}

func getPipelineRef(pipeline *v1alpha3.Pipeline) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:      pipeline.Kind,
		Name:      pipeline.GetName(),
		Namespace: pipeline.GetNamespace(),
	}
}

func createPipelineRun(pipeline *v1alpha3.Pipeline, payload *devops.RunPayload, scm *v1alpha4.SCM) *v1alpha4.PipelineRun {
	controllerRef := metav1.NewControllerRef(pipeline, pipeline.GroupVersionKind())
	return &v1alpha4.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    pipeline.GetName() + "-runs-",
			Namespace:       pipeline.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{*controllerRef},
		},
		Spec: v1alpha4.PipelineRunSpec{
			PipelineRef:  getPipelineRef(pipeline),
			PipelineSpec: &pipeline.Spec,
			Parameters:   convertParameters(payload),
			SCM:          scm,
		},
	}
}
