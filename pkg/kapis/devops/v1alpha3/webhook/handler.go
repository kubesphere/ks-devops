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

package webhook

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/event/models/common"
	"kubesphere.io/devops/pkg/event/models/workflowrun"
	"kubesphere.io/devops/pkg/kapis"
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
		kapis.HandleError(request, response, err)
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
		kapis.HandleError(request, response, err)
	}

	// TODO Register other event handlers here
	// event.HandleWorkflowJob(workflowjob.Funcs{ HandleCreated: handler.handleWorkflowJobCreated })
}
