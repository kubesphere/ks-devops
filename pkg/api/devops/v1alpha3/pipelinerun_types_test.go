package v1alpha3

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestPipelineRunSpec_IsMultiBranchPipeline(t *testing.T) {
	type fields struct {
		PipelineRef  *v1.ObjectReference
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
