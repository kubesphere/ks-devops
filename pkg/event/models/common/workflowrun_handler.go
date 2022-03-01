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

package common

import (
	"encoding/json"

	"kubesphere.io/devops/pkg/event/models/workflowrun"
)

// WorkflowRunType is a full qualified class name in Java.
const WorkflowRunType = "org.jenkinsci.plugins.workflow.job.WorkflowRun"

// HandleWorkflowRun handles WorkflowRun event.
func (event *Event) HandleWorkflowRun(funcs workflowrun.Funcs) error {
	if event == nil || len(event.Data) == 0 || event.DataType != WorkflowRunType {
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
