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

// EventType represents the current type of event. e.g.: run.initialize, run.finished, and so on.
type EventType string

const (
	// RunInitialize represents Jenkins run is initializing.
	RunInitialize EventType = "run.initialize"
	// RunStarted represents Jenkins run has started.
	RunStarted EventType = "run.started"
	// RunFinalized represents Jenkins run has finalized.
	RunFinalized EventType = "run.finalized"
	// RunCompleted represents Jenkins run has completed.
	RunCompleted EventType = "run.completed"
	// RunDeleted represents Jenkins run has been deleted.
	RunDeleted EventType = "run.deleted"
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

// TypeEquals checks the given type is equal to type in event.
// Return true if event is not nil and types are equal, false otherwise.
func (event *Event) TypeEquals(eventType EventType) bool {
	return event != nil && event.Type == string(eventType)
}
