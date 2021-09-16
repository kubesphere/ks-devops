package pipelinerun

import (
	prv1alpha3 "kubesphere.io/devops/pkg/api/devops/pipelinerun/v1alpha3"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"testing"
)

func Test_getBranch(t *testing.T) {
	type args struct {
		prSpec *prv1alpha3.PipelineRunSpec
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{{
		name: "No SCM Pipeline",
		args: args{
			prSpec: &prv1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.NoScmPipelineType,
				},
			},
		},
		want: "",
	}, {
		name: "No SCM Pipeline but SCM set",
		args: args{
			prSpec: &prv1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.NoScmPipelineType,
				},
				SCM: &prv1alpha3.SCM{
					RefName: "main",
					RefType: "branch",
				},
			},
		},
		want: "",
	}, {
		name: "Multi-branch Pipeline but not SCM set",
		args: args{
			prSpec: &prv1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
			},
		},
		wantErr: true,
	}, {
		name: "Multi-branch Pipeline and SCM set",
		args: args{
			prSpec: &prv1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
				SCM: &prv1alpha3.SCM{
					RefName: "main",
					RefType: "branch",
				},
			},
		},
		want: "main",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSCMRefName(tt.args.prSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSCMRefName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSCMRefName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
