package workflowrun

import (
	"testing"

	"kubesphere.io/devops/pkg/event/models/common"
)

func TestWorkflowRunEvent_IsExpectedData(t *testing.T) {
	type fields struct {
		Event common.Event
		Data  *Data
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{{
		name: "Should return true if data type in event is 'io.jenkins.plugins.pipeline.event.data.WorkflowRunData'",
		fields: fields{
			Event: common.Event{
				DataType: "io.jenkins.plugins.pipeline.event.data.WorkflowRunData",
			},
			Data: &Data{},
		},
		want: true,
	}, {
		name: "Should return false if data type in event is not 'io.jenkins.plugins.pipeline.event.data.WorkflowRunData'",
		fields: fields{
			Event: common.Event{
				DataType: "fake.data.type",
			},
			Data: &Data{},
		},
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{
				Event: tt.fields.Event,
				Data:  tt.fields.Data,
			}
			if got := event.DataTypeMatch(); got != tt.want {
				t.Errorf("WorkflowRunEvent.IsExpectedData() = %v, want %v", got, tt.want)
			}
		})
	}
}
