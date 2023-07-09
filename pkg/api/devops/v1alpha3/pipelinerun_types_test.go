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

package v1alpha3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPipelineRunSpec_IsMultiBranchPipeline(t *testing.T) {
	type fields struct {
		PipelineRef  *corev1.ObjectReference
		PipelineSpec *PipelineSpec
		Parameters   []Parameter
		SCM          *SCM
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{{
		name: "No SCM Pipeline",
		fields: fields{
			PipelineSpec: &PipelineSpec{
				Type: NoScmPipelineType,
			},
		},
		want: false,
	}, {
		name:   "No PipelineSpec",
		fields: fields{},
		want:   false,
	}, {
		name: "SCM Pipeline",
		fields: fields{
			PipelineSpec: &PipelineSpec{
				Type: MultiBranchPipelineType,
			},
		},
		want: true,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prSpec := &PipelineRunSpec{
				PipelineRef:  tt.fields.PipelineRef,
				PipelineSpec: tt.fields.PipelineSpec,
				Parameters:   tt.fields.Parameters,
				SCM:          tt.fields.SCM,
			}
			if got := prSpec.IsMultiBranchPipeline(); got != tt.want {
				t.Errorf("IsMultiBranchPipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineRun_GetPipelineRunID(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       PipelineRunSpec
		Status     PipelineRunStatus
	}
	tests := []struct {
		name              string
		fields            fields
		wantPipelineRunID string
		wantExist         bool
	}{{
		name: "normal case",
		fields: fields{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{
					JenkinsPipelineRunIDAnnoKey: "11",
				},
			},
		},
		wantPipelineRunID: "11",
		wantExist:         true,
	}, {
		name: "no build id exist",
		fields: fields{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{},
			},
		},
		wantPipelineRunID: "",
		wantExist:         false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PipelineRun{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			gotPipelineRunID, gotExist := pr.GetPipelineRunID()
			assert.Equalf(t, tt.wantPipelineRunID, gotPipelineRunID, "GetPipelineRunID()")
			assert.Equalf(t, tt.wantExist, gotExist, "GetPipelineRunID()")
		})
	}
}

func TestPipelineRun_Buildable(t *testing.T) {
	now := v1.Now()

	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       PipelineRunSpec
		Status     PipelineRunStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{{
		name: "not buildable due to it was completed",
		fields: fields{
			Status: PipelineRunStatus{
				CompletionTime: &now,
			},
		},
		want: false,
	}, {
		name:   "not completed yet",
		fields: fields{},
		want:   true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PipelineRun{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			assert.Equalf(t, tt.want, pr.Buildable(), "Buildable()")
		})
	}
}

func TestPipelineRun_HasStarted(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       PipelineRunSpec
		Status     PipelineRunStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{{
		name: "started",
		fields: fields{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{
					JenkinsPipelineRunIDAnnoKey: "11",
				},
			},
			Status: PipelineRunStatus{},
			Spec:   PipelineRunSpec{},
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PipelineRun{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			assert.NotNil(t, pr.DeepCopy())
			assert.NotNil(t, pr.DeepCopyObject())
			assert.NotNil(t, pr.Status.DeepCopy())
			assert.NotNil(t, pr.Spec.DeepCopy())
			assert.Equalf(t, tt.want, pr.HasStarted(), "HasStarted()")
		})
	}
}

func TestBuildPipelineRunIdentifier(t *testing.T) {
	type args struct {
		pipelineName string
		scmRefName   string
		runID        string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "Without all of arguments",
		args: args{},
		want: "",
	}, {
		name: "Without SCM reference name",
		args: args{
			pipelineName: "fake-pipeline",
			scmRefName:   "",
			runID:        "1",
		},
		want: "fake-pipeline-1",
	}, {
		name: "With SCM reference name",
		args: args{
			pipelineName: "fake-pipeline",
			scmRefName:   "main",
			runID:        "1",
		},
		want: "fake-pipeline-main-1",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildPipelineRunIdentifier(tt.args.pipelineName, tt.args.scmRefName, tt.args.runID); got != tt.want {
				t.Errorf("BuildPipelineRunIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineRun_GetRefName(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       PipelineRunSpec
		Status     PipelineRunStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test with nil scm",
			fields: fields{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec: PipelineRunSpec{
					PipelineSpec: &PipelineSpec{
						Type: MultiBranchPipelineType,
					},
					SCM: nil,
				},
				Status: PipelineRunStatus{},
			},
			want: "",
		},
		{
			name: "test with noScmPipelineType",
			fields: fields{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec: PipelineRunSpec{
					PipelineSpec: &PipelineSpec{
						Type: NoScmPipelineType,
					},
				},
				Status: PipelineRunStatus{},
			},
			want: "",
		},
		{
			name: "test with multiBranchPipelineType",
			fields: fields{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec: PipelineRunSpec{
					PipelineSpec: &PipelineSpec{
						Type: MultiBranchPipelineType,
					},
					SCM: &SCM{
						RefType: "",
						RefName: "main",
					},
				},
				Status: PipelineRunStatus{},
			},
			want: "main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PipelineRun{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := pr.GetRefName(); got != tt.want {
				t.Errorf("GetRefName() = %v, want %v", got, tt.want)
			}
		})
	}
}
