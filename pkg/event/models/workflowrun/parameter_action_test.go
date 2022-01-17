package workflowrun

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParameterAction_Kind(t *testing.T) {
	tests := []struct {
		name            string
		parameterAction *ParameterAction
		want            string
	}{{
		name:            "Should return correct kind if the action is nil",
		parameterAction: nil,
		want:            "hudson.model.ParametersAction",
	}, {
		name:            "Should return correct kind if the action is not nil",
		parameterAction: &ParameterAction{},
		want:            "hudson.model.ParametersAction",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.parameterAction.Kind(); got != tt.want {
				t.Errorf("ParameterAction.Kind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetParameters(t *testing.T) {
	type args struct {
		actionsJSON string
	}
	tests := []struct {
		name    string
		args    args
		want    []Parameter
		wantErr bool
	}{{
		name: "Should return nil if no parameters",
		args: args{
			actionsJSON: `
				[{
                    "_class": "hudson.model.ParametersAction",
                    "fake_parameters": [{
                        "_class": "hudson.model.BooleanParameterValue",
                        "name": "skip",
                        "value": false
                    }]
                }]`,
		},
		want: nil,
	}, {
		name: "Should return nil if kind is not hudson.model.ParametersAction",
		args: args{
			actionsJSON: `
				[{
                    "_class": "fake.class",
                    "parameters": [{
                        "_class": "hudson.model.BooleanParameterValue",
                        "name": "skip",
                        "value": false
                    }]
                }]`,
		},
		want: nil,
	}, {
		name: "Should return parse parameter correctly if parameters are here",
		args: args{
			actionsJSON: `
				[{
                    "_class": "hudson.model.ParametersAction",
                    "parameters": [{
                        "_class": "hudson.model.BooleanParameterValue",
                        "name": "skip",
                        "value": false
                    }]
                }]`,
		},
		want: []Parameter{{
			Kind:  "hudson.model.BooleanParameterValue",
			Name:  "skip",
			Value: false,
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := Actions{}
			if err := json.Unmarshal([]byte(tt.args.actionsJSON), &actions); err != nil {
				t.Errorf("failed to unmarshal action JSON, err: %v", err)
				return
			}
			got, err := actions.GetParameters()
			if err != nil != tt.wantErr {
				t.Errorf("Want an error, err: %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
