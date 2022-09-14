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

package store

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DataKeyAllLog is the key of all log
	DataKeyAllLog = "log-all"
	// DataKeyStage is the key of stage data
	DataKeyStage = "stage"
	// DataKeyStatus is the key of status
	DataKeyStatus = "status"
)

// StepLogKey generates a unique key by stage and step number
func StepLogKey(stage, step int) string {
	return fmt.Sprintf("log-step-%d-%d", stage, step)
}

// KeyValueStore represents a key-value store
type KeyValueStore interface {
	Get(key string) string
	Set(key, value string)
	Save() error
}

// PipelineRunDataStore represents a PipelineRun data store
type PipelineRunDataStore interface {
	KeyValueStore

	GetStages() string
	SetStages(stages string)
	GetStatus() string
	SetStatus(status string)
	GetStepLog(stage, step int) string
	SetStepLog(stage, step int, log string)
	GetAllLog() string
	SetAllLog(log string)
}

// ConfigMapStore represents a store base on a ConfigMap
type ConfigMapStore interface {
	PipelineRunDataStore
	SetOwnerReference(owner metav1.OwnerReference)
}
