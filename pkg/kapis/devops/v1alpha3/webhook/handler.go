package webhook

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/event/models/common"
	"kubesphere.io/devops/pkg/event/models/workflowrun"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler handles requests from webhooks.
type Handler struct {
	genericClient client.Client
}

// NewHandler creates a new handler for handling webhooks.
func NewHandler(genericClient client.Client) *Handler {
	return &Handler{
		genericClient: genericClient,
	}
}

// ReceiveEventsFromJenkins receives events from Jenkins
func (handler *Handler) ReceiveEventsFromJenkins(request *restful.Request, response *restful.Response) {
	// concrete event body
	event := &common.Event{}
	if err := request.ReadEntity(event); err != nil {
		api.HandleError(request, response, err)
		return
	}

	// handle event
	if err := event.HandleWorkflowRun(workflowrun.Funcs{
		HandleInitialize: handler.handleWorkflowRunInitialize,
		// TODO Handler others
		HandleStarted:   nil,
		HandleFinalized: nil,
		HandleCompleted: nil,
		HandleDeleted:   nil,
	}); err != nil {
		api.HandleError(request, response, err)
	}

	// TODO Register other event handlers here
	// event.HandleWorkflowJob(workflowjob.Funcs{ HandleCreated: handler.handleWorkflowJobCreated })
}
