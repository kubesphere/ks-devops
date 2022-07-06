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
	"kubesphere.io/devops/pkg/models/pipeline"
	"reflect"
	"testing"
	"time"

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

func Test_convertCauses(t *testing.T) {
	type args struct {
		jobCauses []job.Cause
	}
	tests := []struct {
		name string
		args args
		want []pipeline.Cause
	}{{
		name: "normal",
		args: args{
			jobCauses: []job.Cause{{
				"shortDescription": "shortDescription",
			}},
		},
		want: []pipeline.Cause{{
			ShortDescription: "shortDescription",
		}},
	}, {
		name: "parameter and result is nil",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertCauses(tt.args.jobCauses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertCauses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertBranches(t *testing.T) {
	type args struct {
		jobBranches []job.PipelineBranch
	}
	tests := []struct {
		name string
		args args
		want []pipeline.Branch
	}{{
		name: "parameter and result is nil",
		want: []pipeline.Branch{},
	}, {
		name: "normal",
		args: args{
			jobBranches: []job.PipelineBranch{{
				BluePipelineItem: job.BluePipelineItem{
					Name:        "name",
					DisplayName: "displayName",
					Disabled:    true,
				},
				BlueRunnableItem: job.BlueRunnableItem{
					WeatherScore: 100,
				},
			}},
		},
		want: []pipeline.Branch{{
			Name:         "name",
			RawName:      "displayName",
			WeatherScore: 100,
			Disabled:     true,
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertBranches(tt.args.jobBranches); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertBranches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertLatestRun(t *testing.T) {
	now := job.Time{
		Time: time.Now(),
	}
	var durationInMillis int64

	type args struct {
		jobLatestRun *job.PipelineRunSummary
	}
	tests := []struct {
		name string
		args args
		want *pipeline.LatestRun
	}{{
		name: "parameter and result is nil",
	}, {
		name: "normal",
		args: args{
			jobLatestRun: &job.PipelineRunSummary{
				BlueItemRun: job.BlueItemRun{
					DurationInMillis: &durationInMillis,
					EndTime:          now,
					StartTime:        now,
					ID:               "id",
					Name:             "name",
					Result:           "result",
					State:            "state",
				},
			},
		},
		want: &pipeline.LatestRun{
			EndTime:          now,
			DurationInMillis: &durationInMillis,
			StartTime:        now,
			ID:               "id",
			Name:             "name",
			Result:           "result",
			State:            "state",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertLatestRun(tt.args.jobLatestRun); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertLatestRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertPipeline(t *testing.T) {
	var durationInMillis int64

	type args struct {
		jobPipeline *job.Pipeline
	}
	tests := []struct {
		name string
		args args
		want *pipeline.Metadata
	}{{
		name: "normal",
		args: args{
			jobPipeline: &job.Pipeline{
				BlueMultiBranchPipeline: job.BlueMultiBranchPipeline{
					BluePipelineItem: job.BluePipelineItem{
						Name:     "name",
						Disabled: true,
					},
					BlueRunnableItem: job.BlueRunnableItem{
						WeatherScore:              100,
						EstimatedDurationInMillis: durationInMillis,
					},
					BlueContainerItem: job.BlueContainerItem{
						NumberOfPipelines: 100,
						NumberOfFolders:   100,
					},
					BlueMultiBranchItem: job.BlueMultiBranchItem{
						BranchNames:                    []string{"master"},
						NumberOfFailingBranches:        100,
						NumberOfSuccessfulBranches:     100,
						NumberOfSuccessfulPullRequests: 100,
						TotalNumberOfBranches:          100,
						TotalNumberOfPullRequests:      100,
					},
					ScriptPath: "Jenkinsfile",
				},
			},
		},
		want: &pipeline.Metadata{
			WeatherScore:                   100,
			EstimatedDurationInMillis:      durationInMillis,
			Name:                           "name",
			Disabled:                       true,
			NumberOfPipelines:              100,
			NumberOfFolders:                100,
			TotalNumberOfBranches:          100,
			NumberOfSuccessfulBranches:     100,
			TotalNumberOfPullRequests:      100,
			NumberOfFailingBranches:        100,
			NumberOfSuccessfulPullRequests: 100,
			BranchNames:                    []string{"master"},
			ScriptPath:                     "Jenkinsfile",
			Parameters:                     []job.ParameterDefinition{},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertPipeline(tt.args.jobPipeline); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertPipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}
