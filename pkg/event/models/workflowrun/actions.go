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
