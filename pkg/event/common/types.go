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
)

const (
	// RunInitialize represents Jenkins run is initializing.
	RunInitialize string = "run.initialize"
	// RunStarted represents Jenkins run has started.
	RunStarted string = "run.started"
	// RunFinalized represents Jenkins run has finalized.
	RunFinalized string = "run.finalized"
	// RunCompleted represents Jenkins run has completed.
	RunCompleted string = "run.completed"
	// RunDeleted represents Jenkins run has been deleted.
	RunDeleted string = "run.deleted"
)

// Event contains common fields of event except event data.
type Event struct {
	Type     string          `json:"type"`
	Source   string          `json:"source"`
	ID       string          `json:"id"`
	Time     string          `json:"time"`
	DataType string          `json:"dataType"`
	Data     json.RawMessage `json:"data"`
}
