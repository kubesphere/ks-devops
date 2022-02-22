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

package pipelinerun

import (
	"fmt"
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/client/devops"
)

func Test_buildLabelSelector(t *testing.T) {
	parseSelector := func(selector string) labels.Selector {
		parsedSelector, err := labels.Parse(selector)
		if err != nil {
			t.Fatalf("unable to parse labele selector, err = %v", err)
		}
		return parsedSelector
	}

	type args struct {
		queryParam   *query.Query
		pipelineName string
	}
	tests := []struct {
		name    string
		args    args
		want    labels.Selector
		wantErr bool
	}{{
		name: "No label selector was provided",
		args: args{
			queryParam:   &query.Query{},
			pipelineName: "pipelineA",
		},
		want: parseSelector(fmt.Sprintf("%s=pipelineA", v1alpha3.PipelineNameLabelKey)),
	}, {
		name: "Label selector was provided",
		args: args{
			queryParam: &query.Query{
				LabelSelector: "a=b",
			},
			pipelineName: "pipelineA",
		},
		want: parseSelector(fmt.Sprintf("%s=pipelineA,a=b", v1alpha3.PipelineNameLabelKey)),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildLabelSelector(tt.args.queryParam, tt.args.pipelineName)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildLabelSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildLabelSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertPipelineRunsToObject(t *testing.T) {
	type args struct {
		prs []v1alpha3.PipelineRun
	}
	tests := []struct {
		name string
		args args
		want []runtime.Object
	}{{
		name: "Make sure the sequence is correct",
		args: args{
			prs: []v1alpha3.PipelineRun{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-run-a",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-run-b",
					},
				},
			},
		},
		want: []runtime.Object{
			&v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-run-a",
				},
			},
			&v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-run-b",
				},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertPipelineRunsToObject(tt.args.prs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertPipelineRunsToObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertParameters(t *testing.T) {
	type args struct {
		payload *devops.RunPayload
	}
	tests := []struct {
		name string
		args args
		want []v1alpha3.Parameter
	}{{
		name: "Nil payload",
		args: args{
			payload: nil,
		},
		want: nil,
	}, {
		name: "Nil parameters",
		args: args{
			payload: &devops.RunPayload{
				Parameters: nil,
			},
		},
		want: nil,
	}, {
		name: "Single parameter",
		args: args{
			payload: &devops.RunPayload{
				Parameters: []devops.Parameter{{
					Name:  "aname",
					Value: "avalue",
				}},
			},
		},
		want: []v1alpha3.Parameter{{
			Name:  "aname",
			Value: "avalue",
		}},
	}, {
		name: "Empty parameter",
		args: args{
			payload: &devops.RunPayload{
				Parameters: []devops.Parameter{{
					Name:  "",
					Value: "",
				}},
			},
		},
		want: nil,
	}, {
		name: "Empty value only",
		args: args{
			payload: &devops.RunPayload{
				Parameters: []devops.Parameter{{
					Name:  "fakeName",
					Value: "",
				}},
			},
		},
		want: []v1alpha3.Parameter{{
			Name:  "fakeName",
			Value: "",
		}},
	}, {
		name: "Two parameters",
		args: args{
			payload: &devops.RunPayload{
				Parameters: []devops.Parameter{{
					Name:  "aname",
					Value: "avalue",
				}, {
					Name:  "bname",
					Value: "bvalue",
				}},
			},
		},
		want: []v1alpha3.Parameter{{
			Name:  "aname",
			Value: "avalue",
		}, {
			Name:  "bname",
			Value: "bvalue",
		}},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertParameters(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
