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

package workflowrun

import (
	"encoding/json"
	"kubesphere.io/devops/pkg/event/common"
)

// Type is a full qualified class name in Java.
const Type = "org.jenkinsci.plugins.workflow.job.WorkflowRun"

// Handler is a function definition of handling event.
type Handler func(*Data) error

// Handlers are collection of handlers for various event type.
type Handlers struct {
	HandleInitialize Handler
	HandleStarted    Handler
	HandleFinalized  Handler
	HandleCompleted  Handler
	HandleDeleted    Handler
}

// discardHandler will give up handling any data.
var discardHandler Handler = func(data *Data) error {
	// do nothing and return no error
	return nil
}

func (handlers Handlers) getHandler(eventType string) Handler {
	handlerMap := map[string]Handler{
		common.RunInitialize: handlers.HandleInitialize,
		common.RunStarted:    handlers.HandleStarted,
		common.RunFinalized:  handlers.HandleFinalized,
		common.RunCompleted:  handlers.HandleCompleted,
		common.RunDeleted:    handlers.HandleDeleted,
	}
	handler, exist := handlerMap[eventType]
	if !exist || handler == nil {
		return discardHandler
	}
	return handler
}

// Handle handles WorkflowRun event.
func (handlers Handlers) Handle(event *common.Event) error {
	if event == nil || len(event.Data) == 0 || event.DataType != Type {
		return nil
	}
	// unmarshal data to WorkflowRunData
	data := &Data{}
	if err := json.Unmarshal(event.Data, data); err != nil {
		return err
	}
	return handlers.getHandler(event.Type)(data)
}
