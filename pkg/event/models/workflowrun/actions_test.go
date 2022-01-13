package workflowrun

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_GetParameters(t *testing.T) {
	type args struct {
		actionsJSON string
	}
	tests := []struct {
		name string
		args args
		want []Parameter
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
		name: "Should return nil if _class is not hudson.model.ParametersAction",
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
			Name:  "skip",
			Value: "false",
		}},
	}, {
		name: "Should skip parsing parameters if no name field",
		args: args{
			actionsJSON: `
				[{
                    "_class": "hudson.model.ParametersAction",
                    "parameters": [{
                        "_class": "hudson.model.BooleanParameterValue",
                        "fake_name": "skip",
                        "value": false
                    }, {
                        "_class": "hudson.model.BooleanParameterValue",
                        "name": "tag",
                        "value": "main"
                    }]
                }]`,
		},
		want: []Parameter{{
			Name:  "tag",
			Value: "main",
		}},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := Actions{}
			if err := json.Unmarshal([]byte(tt.args.actionsJSON), &actions); err != nil {
				t.Errorf("failed to unmarshal action JSON, err: %v", err)
				return
			}
			if got := actions.GetParameters(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
