package pipeline

import (
	"reflect"
	"testing"

	"github.com/jenkins-zh/jenkins-client/pkg/job"
)

func Test_convertParameterDefinitions(t *testing.T) {
	type args struct {
		paramDefs []job.ParameterDefinition
	}
	tests := []struct {
		name string
		args args
		want []job.ParameterDefinition
	}{{
		name: "Convert nil parameter definitions",
		args: args{},
		want: []job.ParameterDefinition{},
	}, {
		name: "Convert empty parameter definitions",
		args: args{
			paramDefs: []job.ParameterDefinition{},
		},
		want: []job.ParameterDefinition{},
	}, {
		name: "Convert string parameter definition",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "string-name",
				Type: "StringParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "string-name",
			Type: "string",
		}},
	}, {
		name: "Convert choice parameter definition",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "choice-name",
				Type: "ChoiceParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "choice-name",
			Type: "choice",
		}},
	}, {
		name: "Convert text parameter definition",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "text-name",
				Type: "TextParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "text-name",
			Type: "text",
		}},
	}, {
		name: "Convert boolean parameter definition",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "boolean-name",
				Type: "BooleanParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "boolean-name",
			Type: "boolean",
		}},
	}, {
		name: "Convert file parameter definition",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "file-name",
				Type: "FileParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "file-name",
			Type: "file",
		}},
	}, {
		name: "Convert password parameter definition",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "password-name",
				Type: "PasswordParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "password-name",
			Type: "password",
		}},
	}, {
		name: "Convert multi parameter definitions",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "password-name",
				Type: "PasswordParameterDefinition",
			}, {
				Name: "file-name",
				Type: "FileParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "password-name",
			Type: "password",
		}, {
			Name: "file-name",
			Type: "file",
		}},
	}, {
		name: "Convert invalid parameter definition",
		args: args{
			paramDefs: []job.ParameterDefinition{{
				Name: "invalid-name",
				Type: "InvalidParameterDefinition",
			}},
		},
		want: []job.ParameterDefinition{{
			Name: "invalid-name",
			Type: "InvalidParameterDefinition",
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertParameterDefinitions(tt.args.paramDefs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertParameterDefinitions() = %v, want %v", got, tt.want)
			}
		})
	}
}
