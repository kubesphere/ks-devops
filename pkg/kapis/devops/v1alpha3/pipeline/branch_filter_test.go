package pipeline

import (
	"reflect"
	"testing"

	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/apimachinery/pkg/util/rand"
	"kubesphere.io/devops/pkg/models/pipeline"
)

func Test_filterBranches(t *testing.T) {
	type args struct {
		branches []pipeline.PipelineBranch
		filter   string
	}
	tests := []struct {
		name string
		args args
		want []pipeline.PipelineBranch
	}{{
		name: "Without filter",
		args: args{
			branches: []pipeline.PipelineBranch{{
				Name: "main1",
			}, {
				Name: "PR1",
				PullRequest: &job.PullRequest{
					ID: "1",
				},
			}},
			filter: "",
		},
		want: []pipeline.PipelineBranch{{
			Name: "main1",
		}, {
			Name: "PR1",
			PullRequest: &job.PullRequest{
				ID: "1",
			},
		}},
	}, {
		name: "With filter: origin",
		args: args{
			branches: []pipeline.PipelineBranch{{
				Name:        "main1",
				PullRequest: nil,
			}, {
				Name:        "main2",
				PullRequest: &job.PullRequest{},
			}, {
				Name: "PR1",
				PullRequest: &job.PullRequest{
					ID: "1",
				},
			}},
			filter: "origin",
		},
		want: []pipeline.PipelineBranch{{
			Name:        "main1",
			PullRequest: nil,
		}, {
			Name:        "main2",
			PullRequest: &job.PullRequest{},
		}},
	}, {
		name: "With filter: pull-requests",
		args: args{
			branches: []pipeline.PipelineBranch{{
				Name:        "main1",
				PullRequest: nil,
			}, {
				Name:        "main2",
				PullRequest: &job.PullRequest{},
			}, {
				Name: "PR1",
				PullRequest: &job.PullRequest{
					ID: "1",
				},
			}},
			filter: "pull-requests",
		},
		want: []pipeline.PipelineBranch{{
			Name: "PR1",
			PullRequest: &job.PullRequest{
				ID: "1",
			},
		}},
	}, {
		name: "With other filter",
		args: args{
			branches: []pipeline.PipelineBranch{{
				Name: "main1",
			}, {
				Name: "PR1",
				PullRequest: &job.PullRequest{
					ID: "1",
				},
			}},
			filter: rand.String(10),
		},
		want: []pipeline.PipelineBranch{{
			Name: "main1",
		}, {
			Name: "PR1",
			PullRequest: &job.PullRequest{
				ID: "1",
			},
		}},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterBranches(tt.args.branches, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterBranches() = %v, want %v", got, tt.want)
			}
		})
	}
}
