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
	"errors"
	"testing"

	"kubesphere.io/devops/pkg/event/models/workflowrun"
)

func TestEvent_HandleWorkflowRun(t *testing.T) {
	createEvent := func(eventType, dataType string, data *workflowrun.Data) *Event {
		dataBytes, _ := json.Marshal(data)
		return &Event{
			ID:       "fake.id",
			Type:     eventType,
			Time:     "fake-time",
			DataType: dataType,
			Data:     dataBytes,
		}
	}
	type args struct {
		funcs workflowrun.Funcs
	}
	errInitialize := errors.New("Initialized")
	errStarted := errors.New("Started")
	errCompleted := errors.New("Completed")
	errFinalized := errors.New("Finalized")
	errDeleted := errors.New("Deleted")
	initializeHandler := func(*workflowrun.Data) error {
		return errInitialize
	}
	startedHandler := func(*workflowrun.Data) error {
		return errStarted
	}
	completedHandler := func(*workflowrun.Data) error {
		return errCompleted
	}
	finalizedHandler := func(*workflowrun.Data) error {
		return errFinalized
	}
	deletedHandler := func(*workflowrun.Data) error {
		return errDeleted
	}
	tests := []struct {
		name    string
		event   *Event
		args    args
		wantErr error
	}{{
		name:    "Should not return error if event is nil",
		event:   nil,
		wantErr: nil,
	}, {
		name:    "Should not return error if data is nil",
		event:   createEvent("", "", nil),
		wantErr: nil,
	}, {
		name:  "Should invoke initialize handler",
		event: createEvent(string(RunInitialize), WorkflowRunType, &workflowrun.Data{}),
		args: args{
			funcs: workflowrun.Funcs{
				HandleInitialize: initializeHandler,
			},
		},
		wantErr: errInitialize,
	}, {
		name:  "Should invoke started handler",
		event: createEvent(string(RunStarted), WorkflowRunType, &workflowrun.Data{}),
		args: args{
			funcs: workflowrun.Funcs{
				HandleStarted: startedHandler,
			},
		},
		wantErr: errStarted,
	}, {
		name:  "Should invoke finalized handler",
		event: createEvent(string(RunFinalized), WorkflowRunType, &workflowrun.Data{}),
		args: args{
			funcs: workflowrun.Funcs{
				HandleFinalized: finalizedHandler,
			},
		},
		wantErr: errFinalized,
	}, {
		name:  "Should invoke completed handler",
		event: createEvent(string(RunCompleted), WorkflowRunType, &workflowrun.Data{}),
		args: args{
			funcs: workflowrun.Funcs{
				HandleCompleted: completedHandler,
			},
		},
		wantErr: errCompleted,
	}, {
		name:  "Should invoke deleted handler",
		event: createEvent(string(RunDeleted), WorkflowRunType, &workflowrun.Data{}),
		args: args{
			funcs: workflowrun.Funcs{
				HandleDeleted: deletedHandler,
			},
		},
		wantErr: errDeleted,
	}, {
		name:  "Should return nil if event type is out of range",
		event: createEvent("fake.event", WorkflowRunType, &workflowrun.Data{}),
		args: args{
			funcs: workflowrun.Funcs{
				HandleInitialize: initializeHandler,
				HandleStarted:    startedHandler,
				HandleFinalized:  finalizedHandler,
				HandleCompleted:  completedHandler,
				HandleDeleted:    deletedHandler,
			},
		},
		wantErr: nil,
	}, {
		name:  "Should return nil if data type is invalid",
		event: createEvent(string(RunInitialize), "fake.data.type", &workflowrun.Data{}),
		args: args{
			funcs: workflowrun.Funcs{
				HandleInitialize: initializeHandler,
				HandleStarted:    startedHandler,
				HandleFinalized:  finalizedHandler,
				HandleCompleted:  completedHandler,
				HandleDeleted:    deletedHandler,
			},
		},
		wantErr: nil,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.event.HandleWorkflowRun(tt.args.funcs); err != tt.wantErr {
				t.Errorf("Event.HandleWorkflowRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
