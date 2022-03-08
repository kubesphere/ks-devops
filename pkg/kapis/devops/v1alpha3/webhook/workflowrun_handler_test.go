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

package webhook

import (
	"context"
	"kubesphere.io/devops/pkg/event/workflowrun"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_pipelineRunIdentifier_String(t *testing.T) {
	tests := []struct {
		name       string
		identifier *pipelineRunIdentifier
		want       string
	}{{
		name: "Should return identifier not containing namespace name",
		identifier: &pipelineRunIdentifier{
			namespaceName: "fake.namespace",
			pipelineName:  "fake.pipeline",
			scmRefName:    "fake.ref",
			buildNumber:   "1",
		},
		want: "fake.pipeline-fake.ref-1",
	}, {
		name:       "Should return empty string if PipelineRunIdentifier is nil",
		identifier: nil,
		want:       "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.identifier.String(); got != tt.want {
				t.Errorf("pipelineRunIdentifier.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertParameters(t *testing.T) {
	type args struct {
		parameters []workflowrun.Parameter
	}
	tests := []struct {
		name string
		args args
		want []v1alpha3.Parameter
	}{{
		name: "Nil parameters",
		args: args{
			parameters: nil,
		},
		want: nil,
	}, {
		name: "Single parameter",
		args: args{
			parameters: []workflowrun.Parameter{{
				Name:  "aname",
				Value: "avalue",
			}},
		},
		want: []v1alpha3.Parameter{{
			Name:  "aname",
			Value: "avalue",
		}},
	}, {
		name: "Empty parameter",
		args: args{
			parameters: []workflowrun.Parameter{{
				Name:  "",
				Value: "",
			}},
		},
		want: nil,
	}, {
		name: "Empty value only",
		args: args{
			parameters: []workflowrun.Parameter{{
				Name:  "fakeName",
				Value: "",
			}},
		},
		want: []v1alpha3.Parameter{{
			Name:  "fakeName",
			Value: "",
		}},
	}, {
		name: "Two parameters",
		args: args{
			parameters: []workflowrun.Parameter{{
				Name:  "aname",
				Value: "avalue",
			}, {
				Name:  "bname",
				Value: "bvalue",
			}},
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
			if got := convertParameters(tt.args.parameters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
func createWorkflowRun(parentFullName, projectName, buildNumber string, isMultiBranch bool) *workflowrun.Data {
	return &workflowrun.Data{
		ParentFullName: parentFullName,
		ProjectName:    projectName,
		IsMultiBranch:  isMultiBranch,
		ID:             buildNumber,
	}
}
func Test_extractPipelineRunIdentifier(t *testing.T) {

	type args struct {
		workflowRunData *workflowrun.Data
	}
	tests := []struct {
		name string
		args args
		want *pipelineRunIdentifier
	}{{
		name: "Should return nil if workflowRunData is nil",
		args: args{
			workflowRunData: nil,
		},
		want: nil,
	}, {
		name: "Should return nil if parent full name is empty",
		args: args{
			workflowRunData: createWorkflowRun("", "main", "1", true),
		},
		want: nil,
	}, {
		name: "Should return nil if parent full name contains more than one slash and multi branch is true",
		args: args{
			workflowRunData: createWorkflowRun("a/b/c", "main", "1", true),
		},
		want: nil,
	}, {
		name: "Should return nil if parent full name contains slash and multi branch is false",
		args: args{
			workflowRunData: createWorkflowRun("a/b", "c", "1", false),
		},
		want: nil,
	}, {
		name: "Should return identifier if parent full name does not contain slash and multi branch is false",
		args: args{
			workflowRunData: createWorkflowRun("a", "b", "1", false),
		},
		want: &pipelineRunIdentifier{
			namespaceName: "a",
			pipelineName:  "b",
			buildNumber:   "1",
			scmRefName:    "",
		},
	}, {
		name: "Should return identifier if parent full name contains one slash and multi branch is true",
		args: args{
			workflowRunData: createWorkflowRun("a/b", "main", "1", true),
		},
		want: &pipelineRunIdentifier{
			namespaceName: "a",
			pipelineName:  "b",
			buildNumber:   "1",
			scmRefName:    "main",
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractPipelineRunIdentifier(tt.args.workflowRunData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractPipelineRunIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_handleWorkflowRunInitialize(t *testing.T) {
	createPipeline := func(namespace, name string) *v1alpha3.Pipeline {
		return &v1alpha3.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}
	type args struct {
		workflowRunData *workflowrun.Data
		initObjs        []runtime.Object
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		assertion func(*testing.T, client.Client)
	}{{
		name: "Should create a new PipelineRun",
		args: args{
			workflowRunData: createWorkflowRun("fake-namespace", "fake-pipeline", "1", false),
			initObjs: []runtime.Object{
				createPipeline("fake-namespace", "fake-pipeline"),
			},
		},
		wantErr: false,
		assertion: func(t *testing.T, c client.Client) {
			pipelineRuns := &v1alpha3.PipelineRunList{}
			_ = c.List(context.Background(), pipelineRuns)
			assert.Equal(t, 1, len(pipelineRuns.Items))
		},
	}, {
		name: "Should return an not found error if Pipeline not found",
		args: args{
			workflowRunData: createWorkflowRun("fake-namepsace", "fake-pipeline", "1", false),
		},
		wantErr: true,
	}, {
		name: "Should create nothing if WorkflowRunData is invalid",
		args: args{
			workflowRunData: createWorkflowRun("", "", "", false),
		},
		wantErr: false,
		assertion: func(t *testing.T, c client.Client) {
			pipelineRuns := &v1alpha3.PipelineRunList{}
			_ = c.List(context.Background(), pipelineRuns)
			assert.Equal(t, 0, len(pipelineRuns.Items))
		},
	}}
	for _, tt := range tests {
		scheme := runtime.NewScheme()
		_ = v1alpha3.AddToScheme(scheme)
		fakeClient := fake.NewFakeClientWithScheme(scheme, tt.args.initObjs...)

		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{
				Client: fakeClient,
			}
			if err := handler.handleWorkflowRunInitialize(tt.args.workflowRunData); (err != nil) != tt.wantErr {
				t.Errorf("Handler.handleWorkflowRunInitialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.assertion != nil {
				// check the result
				tt.assertion(t, fakeClient)
			}
		})
	}
}
