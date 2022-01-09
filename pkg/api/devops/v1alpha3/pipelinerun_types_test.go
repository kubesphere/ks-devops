package v1alpha3

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	corev1 "k8s.io/api/core/v1"
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
