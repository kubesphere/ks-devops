package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/event/models/common"
	"kubesphere.io/devops/pkg/event/models/workflowrun"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/pipelinerun"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler handles requests from webhooks.
type Handler struct {
	genericClient client.Client
}

func NewHandler(genericClient client.Client) *Handler {
	return &Handler{
		genericClient: genericClient,
	}
}

// ReceiveEventsFromJenkins receives events from Jenkins
func (handler *Handler) ReceiveEventsFromJenkins(request *restful.Request, response *restful.Response) {
	workflowRunEvent := &workflowrun.Event{}
	if err := request.ReadEntity(workflowRunEvent); err != nil {
		errorHandle(request, response, nil, err)
		return
	}

	if workflowRunEvent.DataTypeMatch() {
		// we are only interested in WorkflowRun event here
		var err error
		if workflowRunEvent.TypeEquals(common.RunInitialize) {
			err = handleWorkflowRunInitialize(workflowRunEvent, handler.genericClient)
		}

		// TODO Handle other event type, like run.started, run.finalized and so on.

		if err != nil {
			errorHandle(request, response, nil, err)
		}
		return
	}

	// TODO Handle other data type, like WorkflowJob.
}

func handleWorkflowRunInitialize(workflowRunEvent *workflowrun.Event, genericClient client.Client) error {
	workflowRunData := workflowRunEvent.Data
	namespaceName := ""
	pipelineName := ""
	scmRefName := ""
	buildNumber := workflowRunData.Run.ID

	fullName := workflowRunData.ParentFullName
	names := strings.Split(fullName, "/")
	if workflowRunData.IsMultiBranch {
		if len(names) != 2 {
			// return if this is not a standard multi-branch Pipeline in ks-devops
			return nil
		}
		namespaceName = names[0]
		pipelineName = names[1]
		scmRefName = workflowRunData.ProjectName
	} else {
		if len(names) != 1 {
			// return if this is not a standard Pipeline in ks-devops
			return nil
		}
		namespaceName = workflowRunData.ParentFullName
		pipelineName = workflowRunData.ProjectName
	}

	// TODO Execute process below asynchronously

	pipelineRunIdentifier := v1alpha3.BuildPipelineRunIdentifier(pipelineName, scmRefName, fmt.Sprint(buildNumber))
	pipelineRunList := &v1alpha3.PipelineRunList{}
	if err := genericClient.List(context.Background(), pipelineRunList,
		client.InNamespace(namespaceName),
		client.MatchingFields{v1alpha3.PipelineRunIdentifierIndexerName: pipelineRunIdentifier}); err != nil {
		return err
	}

	if len(pipelineRunList.Items) == 0 {
		// If no PipelineRun found, it indicates that a fresh PipelineRun was created from Jenkins.
		pipeline := &v1alpha3.Pipeline{}
		if err := genericClient.Get(context.Background(), client.ObjectKey{Namespace: namespaceName, Name: pipelineName}, pipeline); err != nil {
			return err
		}
		scm, err := pipelinerun.CreateScm(&pipeline.Spec, scmRefName)
		if err != nil {
			return err
		}

		parameters, err := workflowRunData.Run.Actions.GetParameters()
		if err != nil {
			return err
		}
		pipelineRun := pipelinerun.CreateBarePipelineRun(pipeline, pipelinerun.ConvertParameters2(parameters), scm)

		// Set the RunID manually
		pipelineRun.GetAnnotations()[v1alpha3.JenkinsPipelineRunIDAnnoKey] = fmt.Sprint(buildNumber)
		if err := genericClient.Create(context.Background(), pipelineRun); err != nil {
			return err
		}
		klog.Infof("Created a PipelineRun: %s/%s", pipelineRun.Namespace, pipelineRun.Name)
	}
	return nil
}
