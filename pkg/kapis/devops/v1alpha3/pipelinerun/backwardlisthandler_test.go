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
	"encoding/json"
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
)

func Test_compatibleTransform(t *testing.T) {
	tests := []struct {
		name string
		obj  runtime.Object
		want interface{}
	}{{
		name: "With run status",
		obj: &v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunStatusAnnoKey: `{"id": "123"}`,
				},
			},
		},
		want: json.RawMessage(`{"id": "123"}`),
	}, {
		name: "Without annotations",
		obj: &v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{},
		},
		want: json.RawMessage("{}"),
	}, {
		name: "Nil PipelineRun",
		obj:  (*v1alpha3.PipelineRun)(nil),
		want: json.RawMessage("{}"),
	}, {
		name: "Nil object",
		obj:  nil,
		want: json.RawMessage("{}"),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := backwardListHandler{}
			if got := handler.Transformer()(tt.obj); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("backwardTransform() = %v, want %v", got, tt.want)
			} else if !reflect.TypeOf(got).AssignableTo(reflect.TypeOf((*json.Marshaler)(nil)).Elem()) {
				t.Errorf("backwardTransform() should return an instance of json.Marshaler, current type is %s", reflect.TypeOf(got))
			}
		})
	}
}

func Test_backwardFilter(t *testing.T) {
	type args struct {
		obj    runtime.Object
		filter query.Filter
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "Nil object",
		args: args{
			obj: nil,
		},
		want: false,
	}, {
		name: "Nil PipelineRun",
		args: args{
			obj: (*v1alpha3.PipelineRun)(nil),
		},
		want: false,
	}, {
		name: "PipelineRun has started but without Jenkins run status",
		args: args{
			obj: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
					},
				},
			},
		},
		want: false,
	}, {
		name: "PipelineRun hasn't started but with Jenkins run status",
		args: args{
			obj: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunStatusAnnoKey: `{"id": "123"}`,
					},
				},
			},
		},
		want: true,
	}, {
		name: "PipelineRun has started and with Jenkins run status",
		args: args{
			obj: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "abc",
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunStatusAnnoKey: `{"id": "123"}`,
						v1alpha3.JenkinsPipelineRunIDAnnoKey:     "123",
					},
				},
			},
			filter: query.Filter{
				Field: "name",
				Value: "abc",
			},
		},
		want: true,
	}, {
		name: "PipelineRun has started and with Jenkins run status but failed with default filter",
		args: args{
			obj: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "abc",
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunStatusAnnoKey: `{"id": "123"}`,
						v1alpha3.JenkinsPipelineRunIDAnnoKey:     "123",
					},
				},
			},
			filter: query.Filter{
				Field: "name",
				Value: "def",
			},
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := backwardListHandler{}
			if got := handler.Filter()(tt.args.obj, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("backwardFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
