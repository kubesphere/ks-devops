package v1alpha4

import (
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"testing"
)

func TestPipelineRunSpec_IsMultiBranchPipeline(t *testing.T) {
	type fields struct {
		PipelineRef  *v1.ObjectReference
		PipelineSpec *v1alpha3.PipelineSpec
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
			PipelineSpec: &v1alpha3.PipelineSpec{
				Type: v1alpha3.NoScmPipelineType,
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
			PipelineSpec: &v1alpha3.PipelineSpec{
				Type: v1alpha3.MultiBranchPipelineType,
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
