package pipelinerun

import (
	"context"
	"github.com/emicklei/go-restful"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api"
	prv1alpha3 "kubesphere.io/devops/pkg/api/devops/pipelinerun/v1alpha3"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/client/devops"
	resourcesV1alpha3 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

// apiHandlerOption holds some useful tools for API handler.
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

func (h *apiHandler) listPipelineRuns(request *restful.Request, response *restful.Response) {
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
	pipeline := &v1alpha3.Pipeline{}
	err = h.client.Get(context.Background(), client.ObjectKey{Namespace: nsName, Name: pipName}, pipeline)
	if err != nil {
		api.HandleError(request, response, err)
		return
	}

	// build label selector
	labelSelector, err := buildLabelSelector(queryParam, pipeline.Name, branchName)
	if err != nil {
		api.HandleError(request, response, err)
		return
	}

	var prs prv1alpha3.PipelineRunList
	// fetch PipelineRuns
	if err := h.client.List(context.Background(), &prs,
		client.InNamespace(pipeline.Namespace),
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

	go func() {
		err := h.requestSyncPipelineRun(client.ObjectKey{Namespace: pipeline.Namespace, Name: pipeline.Name})
		if err != nil {
			klog.Errorf("failed to request to synchronize PipelineRuns. Pipeline = %s", pipeline.Name)
			return
		}
	}()
}

func (h *apiHandler) requestSyncPipelineRun(key client.ObjectKey) error {
	// get latest Pipeline
	pipelineToUpdate := &v1alpha3.Pipeline{}
	err := h.client.Get(context.Background(), key, pipelineToUpdate)
	if err != nil {
		return err
	}
	// update pipeline annotations
	if pipelineToUpdate.Annotations == nil {
		pipelineToUpdate.Annotations = make(map[string]string)
	}
	pipelineToUpdate.Annotations[v1alpha3.PipelineRequestToSyncRunsAnnoKey] = "true"
	// update the pipeline
	if err := h.client.Update(context.Background(), pipelineToUpdate); err != nil && !apierrors.IsConflict(err) {
		// we allow the conflict error here
		return err
	}
	return nil
}

func (h *apiHandler) createPipelineRuns(request *restful.Request, response *restful.Response) {
	nsName := request.PathParameter("namespace")
	pipName := request.PathParameter("pipeline")
	branch := request.QueryParameter("branch")
	payload := devops.RunPayload{}
	if err := request.ReadEntity(&payload); err != nil && err != io.EOF {
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
		scm *prv1alpha3.SCM
		err error
	)
	if scm, err = CreateScm(&pipeline.Spec, branch); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	// create PipelineRun
	pr := CreatePipelineRun(&pipeline, &payload, scm)
	if err := h.client.Create(context.Background(), pr); err != nil {
		api.HandleError(request, response, err)
		return
	}

	_ = response.WriteEntity(pr)
}

func (h *apiHandler) getPipelineRun(request *restful.Request, response *restful.Response) {
	nsName := request.PathParameter("namespace")
	prName := request.PathParameter("pipelinerun")

	// get pipelinerun
	var pr prv1alpha3.PipelineRun
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: nsName, Name: prName}, &pr); err != nil {
		api.HandleError(request, response, err)
		return
	}
	_ = response.WriteEntity(&pr)
}
