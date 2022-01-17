package common

import (
	"encoding/json"

	"kubesphere.io/devops/pkg/event/models/workflowrun"
)

// WorkflowRunDataType is a full qualified class name in Java.
const WorkflowRunDataType = "io.jenkins.plugins.pipeline.event.data.WorkflowRunData"

// HandleWorkflowRun handles WorkflowRun event.
func (event *Event) HandleWorkflowRun(funcs workflowrun.Funcs) error {
	if event == nil || len(event.Data) == 0 || event.DataType != WorkflowRunDataType {
		return nil
	}
	// unmarshal data to WorkflowRunData
	workflowRunData := &workflowrun.Data{}
	if err := json.Unmarshal(event.Data, workflowRunData); err != nil {
		return err
	}
	/// handle various event
	if event.TypeEquals(RunInitialize) && funcs.HandleInitialize != nil {
		return funcs.HandleInitialize(workflowRunData)
	}
	if event.TypeEquals(RunStarted) && funcs.HandleStarted != nil {
		return funcs.HandleStarted(workflowRunData)
	}
	if event.TypeEquals(RunFinalized) && funcs.HandleFinalized != nil {
		return funcs.HandleFinalized(workflowRunData)
	}
	if event.TypeEquals(RunCompleted) && funcs.HandleCompleted != nil {
		return funcs.HandleCompleted(workflowRunData)
	}
	if event.TypeEquals(RunDeleted) && funcs.HandleDeleted != nil {
		return funcs.HandleDeleted(workflowRunData)
	}
	return nil
}

// TODO Handle other events
// func (event *Event) HandleWorkflowJob(fucns workflowjob.Funcs) error {}
