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

// Data contains WorkflowJob breif information and WorkflowRun detail.
type Data struct {
	ParentFullName string      `json:"parentFullName"`
	ProjectName    string      `json:"projectName"`
	IsMultiBranch  bool        `json:"multiBranch"`
	Run            WorkflowRun `json:"run"`
}

// WorkflowRun contains WorkflowRun detail.
type WorkflowRun struct {
	Actions           Actions `json:"actions"`
	Building          bool    `json:"building"`
	Description       string  `json:"description"`
	DisplayName       string  `json:"displayName"`
	Duration          int     `json:"duration"`
	EstimatedDuration int     `json:"estimatedDuration"`
	FullDisplayName   string  `json:"fullDisplayName"`
	ID                string  `json:"id"`
	KeepLog           bool    `json:"keepLog"`
	Number            int     `json:"number"`
	QueueID           int     `json:"queueId"`
	Result            string  `json:"result"`
	Timestamp         int64   `json:"timestamp"`
}

// Funcs is a collection of handlers for various event type.
type Funcs struct {
	HandleInitialize func(*Data) error
	HandleStarted    func(*Data) error
	HandleFinalized  func(*Data) error
	HandleCompleted  func(*Data) error
	HandleDeleted    func(*Data) error
}
