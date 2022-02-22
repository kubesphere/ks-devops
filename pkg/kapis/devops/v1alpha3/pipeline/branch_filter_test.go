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
		branches []pipeline.Branch
		filter   job.Filter
	}
	tests := []struct {
		name string
		args args
		want []pipeline.Branch
	}{{
		name: "Without filter",
		args: args{
			branches: []pipeline.Branch{{
				Name: "main1",
			}, {
				Name: "PR1",
				PullRequest: &job.PullRequest{
					ID: "1",
				},
			}},
			filter: "",
		},
		want: []pipeline.Branch{{
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
			branches: []pipeline.Branch{{
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
		want: []pipeline.Branch{{
			Name:        "main1",
			PullRequest: nil,
		}, {
			Name:        "main2",
			PullRequest: &job.PullRequest{},
		}},
	}, {
		name: "With filter: origin, but name is written in Chinese",
		args: args{
			branches: []pipeline.Branch{{
				Name:        "main1",
				PullRequest: nil,
			}, {
				Name:        "主分支2",
				PullRequest: &job.PullRequest{},
			}, {
				Name: "PR1",
				PullRequest: &job.PullRequest{
					ID: "1",
				},
			}},
			filter: "origin",
		},
		want: []pipeline.Branch{{
			Name:        "main1",
			PullRequest: nil,
		}, {
			Name:        "主分支2",
			PullRequest: &job.PullRequest{},
		}},
	}, {
		name: "With filter: pull-requests",
		args: args{
			branches: []pipeline.Branch{{
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
		want: []pipeline.Branch{{
			Name: "PR1",
			PullRequest: &job.PullRequest{
				ID: "1",
			},
		}},
	}, {
		name: "With filter: pull-requests, but name is written in Chinese",
		args: args{
			branches: []pipeline.Branch{{
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
			}, {
				Name: "分支2",
				PullRequest: &job.PullRequest{
					ID: "2",
				},
			}},
			filter: "pull-requests",
		},
		want: []pipeline.Branch{{
			Name: "PR1",
			PullRequest: &job.PullRequest{
				ID: "1",
			},
		}, {
			Name: "分支2",
			PullRequest: &job.PullRequest{
				ID: "2",
			},
		}},
	}, {
		name: "With other filter",
		args: args{
			branches: []pipeline.Branch{{
				Name: "main1",
			}, {
				Name: "PR1",
				PullRequest: &job.PullRequest{
					ID: "1",
				},
			}},
			filter: job.Filter(rand.String(10)),
		},
		want: []pipeline.Branch{{
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

func Test_branchSlice_toGenericSlice(t *testing.T) {
	tests := []struct {
		name     string
		branches branchSlice
		want     []interface{}
	}{{
		name:     "Empty branches",
		branches: branchSlice{},
		want:     []interface{}{},
	}, {
		name: "Non-empty branches and sequence kept",
		branches: branchSlice{{
			Name: "main",
		}, {
			Name: "dev",
		}, {
			Name: "release",
		}},
		want: []interface{}{pipeline.Branch{
			Name: "main",
		}, pipeline.Branch{
			Name: "dev",
		}, pipeline.Branch{
			Name: "release",
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.branches.toGenericSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("branchSlice.toGenericSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
