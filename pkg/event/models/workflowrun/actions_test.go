package workflowrun

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAction_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		actionJSON string
		wantErr    bool
		wantAction *Action
	}{{
		name:       "Should return an action with kind while having _class field",
		actionJSON: `{"_class": "fake.class"}`,
		wantAction: &Action{
			Raw:  json.RawMessage(`{"_class": "fake.class"}`),
			Kind: "fake.class",
		},
	}, {
		name:       "Should return an action without kind while having no _class field",
		actionJSON: `{"_fake_class": "fake.class"}`,
		wantAction: &Action{
			Raw:  json.RawMessage(`{"_fake_class": "fake.class"}`),
			Kind: "",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &Action{}
			if err := json.Unmarshal([]byte(tt.actionJSON), action); err != nil != tt.wantErr {
				t.Errorf("Action.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(action, tt.wantAction) {
				t.Errorf("failed to unmarshal action, got = %v, want = %v", action, tt.wantAction)
			}
		})
	}
}
