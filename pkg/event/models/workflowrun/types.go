package workflowrun

import "kubesphere.io/devops/pkg/event/models/common"

var _ common.DataTypeMatcher = &Event{}

// Event is an event with WorkflowRun detail.
type Event struct {
	common.Event `json:",inline"`
	Data         *Data `json:"data"`
}

// Data contains WorkflowJob breif information and WorkflowRun detail.
type Data struct {
	ParentFullName string      `json:"parentFullName"`
	ProjectName    string      `json:"projectName"`
	IsMultiBranch  bool        `json:"multiBranch"`
	Run            WorkflowRun `json:"run"`
}

// DataTypeMatch returns true if the DataType in event is equal to real type of event data, false otherwise.
func (event *Event) DataTypeMatch() bool {
	return event != nil && event.Data != nil && event.DataType == "io.jenkins.plugins.pipeline.event.data.WorkflowRunData"
}

// WorkflowRun contains WorkflowRun detail.
type WorkflowRun struct {
	Actions           Actions                  `json:"actions"`
	Artifacts         []map[string]interface{} `json:"artifacts"`
	Building          bool                     `json:"building"`
	Description       string                   `json:"description"`
	DisplayName       string                   `json:"displayName"`
	Duration          int                      `json:"duration"`
	EstimatedDuration int                      `json:"estimatedDuration"`
	FullDisplayName   string                   `json:"fullDisplayName"`
	ID                string                   `json:"id"`
	KeepLog           bool                     `json:"keepLog"`
	Number            int                      `json:"number"`
	QueueID           int                      `json:"queueId"`
	Result            string                   `json:"result"`
	Timestamp         int64                    `json:"timestamp"`
	ChangeSets        []map[string]interface{} `json:"changeSets"`
	Culprits          []map[string]interface{} `json:"culprits"`
}
