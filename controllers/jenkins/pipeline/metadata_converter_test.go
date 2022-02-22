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
