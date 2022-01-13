package workflowrun

import (
	"encoding/json"
	"testing"

	"kubesphere.io/devops/pkg/event/models/common"
)

func TestWorkflowRunEvent_IsExpectedData(t *testing.T) {
	eventJSON := `
{
    "data":     {
            "_class": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
            "actions": [],
            "artifacts": [],
            "building": false,
            "description": null,
            "displayName": "#1",
            "duration": 1535,
            "estimatedDuration": 1535,
            "executor": {"_class": "hudson.model.OneOffExecutor"},
            "fullDisplayName": "my-devops-project » example-pipeline #1",
            "id": "1",
            "keepLog": false,
            "number": 1,
            "queueId": 1,
            "result": "SUCCESS",
            "timestamp": 1642059183172,
            "changeSets": [],
            "culprits": [],
            "nextBuild": null,
            "previousBuild": null
    },
    "dataType": "io.jenkins.plugins.pipeline.event.data.WorkflowRunData",
    "id": "2e42b4a8-dc8d-46da-8456-f6f4936fe3f5",
    "source": "job/my-devops-project/job/example-pipeline/",
    "time": "2022-01-13T15:33:04.774+0800",
    "type": "run.finalized"
}
	`

	var workflowRunEvent Event
	if err := json.Unmarshal([]byte(eventJSON), &workflowRunEvent); err != nil {
		t.Errorf("failed to unmarshal event JSON, err: %v", err)
		return
	}

	type fields struct {
		Event common.Event
		Data  *Data
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{}

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
