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
)

// Action is an item of actions, which contains kind and raw data of action for further unserialization.
type Action struct {
	// Raw data for futher unmarshal.
	Raw json.RawMessage
	// Kind represents what action would look like.
	Kind string
}

var _ json.Unmarshaler = (*Action)(nil)

// UnmarshalJSON will extract Kind field for action and store the raw data to action.
func (action *Action) UnmarshalJSON(data []byte) error {
	// get the _class field
	kind := &struct {
		Kind string `json:"_class"`
	}{}
	if err := json.Unmarshal(data, kind); err != nil {
		return err
	}
	action.Kind = kind.Kind
	action.Raw = data
	return nil
}

// Actions are set of action carried by WorkflowRun.
type Actions []Action

// GetAction gets action from actions in WorkflowRun by full qualified class name.
// Return nil if not found.
func (actions Actions) GetAction(class string) *Action {
	if class == "" {
		return nil
	}
	for i := range actions {
		action := &actions[i]
		if action.Kind == class {
			return action
		}
	}
	return nil
}
