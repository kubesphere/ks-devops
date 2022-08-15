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

package indexers

import (
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"testing"
)

func TestCreatePipelineRunSCMRefNameIndexer(t *testing.T) {
	type args struct {
		runtimeCache cache.Cache
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{
		name: "normal",
		args: args{
			runtimeCache: &informertest.FakeInformers{},
		},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreatePipelineRunSCMRefNameIndexer(tt.args.runtimeCache); (err != nil) != tt.wantErr {
				t.Errorf("CreatePipelineRunSCMRefNameIndexer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreatePipelineRunIdentityIndexer(t *testing.T) {
	type args struct {
		runtimeCache cache.Cache
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{
		name:    "normal",
		args:    args{runtimeCache: &informertest.FakeInformers{}},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreatePipelineRunIdentityIndexer(tt.args.runtimeCache); (err != nil) != tt.wantErr {
				t.Errorf("CreatePipelineRunIdentityIndexer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_extractSCMFunc(t *testing.T) {
	type args struct {
		o runtime.Object
	}
	tests := []struct {
		name string
		args args
		want []string
	}{{
		name: "not expect Kind",
		args: args{
			o: &v1.ConfigMap{},
		},
		want: []string{},
	}, {
		name: "scm is nil",
		args: args{
			o: &v1alpha3.PipelineRun{},
		},
		want: []string{},
	}, {
		name: "have valid scm",
		args: args{
			o: &v1alpha3.PipelineRun{
				Spec: v1alpha3.PipelineRunSpec{
					SCM: &v1alpha3.SCM{
						RefName: "master",
					},
				},
			},
		},
		want: []string{"master"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractSCMFunc(tt.args.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractSCMFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractPipelineRunIdentifier(t *testing.T) {
	type args struct {
		o runtime.Object
	}
	tests := []struct {
		name string
		args args
		want []string
	}{{
		name: "not expect kind",
		args: args{
			o: &v1.ConfigMap{},
		},
		want: []string{},
	}, {
		name: "valid PipelineRun",
		args: args{
			o: &v1alpha3.PipelineRun{
				ObjectMeta: v12.ObjectMeta{
					Labels: map[string]string{
						"devops.kubesphere.io/pipeline": "fake",
					},
				},
			},
		},
		want: []string{"fake"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractPipelineRunIdentifier(tt.args.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractPipelineRunIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}
