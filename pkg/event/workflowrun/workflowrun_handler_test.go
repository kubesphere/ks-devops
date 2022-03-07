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
	"errors"
	"kubesphere.io/devops/pkg/event/common"
	"testing"
)

func TestEvent_HandleWorkflowRun(t *testing.T) {
	createEvent := func(eventType, dataType string, data *Data) *common.Event {
		dataBytes, _ := json.Marshal(data)
		return &common.Event{
			ID:       "fake.id",
			Type:     eventType,
			Time:     "fake-time",
			DataType: dataType,
			Data:     dataBytes,
		}
	}
	type args struct {
		handlers Handlers
	}
	errInitialize := errors.New("initialized")
	errStarted := errors.New("started")
	errCompleted := errors.New("completed")
	errFinalized := errors.New("finalized")
	errDeleted := errors.New("deleted")
	initializeHandler := func(*Data) error {
		return errInitialize
	}
	startedHandler := func(*Data) error {
		return errStarted
	}
	completedHandler := func(*Data) error {
		return errCompleted
	}
	finalizedHandler := func(*Data) error {
		return errFinalized
	}
	deletedHandler := func(*Data) error {
		return errDeleted
	}
	tests := []struct {
		name    string
		event   *common.Event
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
		event: createEvent(string(common.RunInitialize), Type, &Data{}),
		args: args{
			handlers: Handlers{
				HandleInitialize: initializeHandler,
			},
		},
		wantErr: errInitialize,
	}, {
		name:  "Should invoke started handler",
		event: createEvent(string(common.RunStarted), Type, &Data{}),
		args: args{
			handlers: Handlers{
				HandleStarted: startedHandler,
			},
		},
		wantErr: errStarted,
	}, {
		name:  "Should invoke finalized handler",
		event: createEvent(string(common.RunFinalized), Type, &Data{}),
		args: args{
			handlers: Handlers{
				HandleFinalized: finalizedHandler,
			},
		},
		wantErr: errFinalized,
	}, {
		name:  "Should invoke completed handler",
		event: createEvent(string(common.RunCompleted), Type, &Data{}),
		args: args{
			handlers: Handlers{
				HandleCompleted: completedHandler,
			},
		},
		wantErr: errCompleted,
	}, {
		name:  "Should invoke deleted handler",
		event: createEvent(string(common.RunDeleted), Type, &Data{}),
		args: args{
			handlers: Handlers{
				HandleDeleted: deletedHandler,
			},
		},
		wantErr: errDeleted,
	}, {
		name:  "Should return nil if event type is out of range",
		event: createEvent("fake.event", Type, &Data{}),
		args: args{
			handlers: Handlers{
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
		event: createEvent(string(common.RunInitialize), "fake.data.type", &Data{}),
		args: args{
			handlers: Handlers{
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
			if err := tt.args.handlers.Handle(tt.event); err != tt.wantErr {
				t.Errorf("Event.Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
