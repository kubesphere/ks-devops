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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPipeline_IsMultiBranch(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *Pipeline
		want     bool
	}{{
		name:     "Should return false if the Pipeline is nil",
		pipeline: nil,
		want:     false,
	}, {
		name: "Should return false if the type of Pipeline is empty",
		pipeline: &Pipeline{
			Spec: PipelineSpec{
				Type: "",
			},
		},
		want: false,
	}, {name: "Should return false if the type of Pipeline is pipeline",
		pipeline: &Pipeline{
			Spec: PipelineSpec{
				Type: NoScmPipelineType,
			},
		},
		want: false,
	}, {
		name: "Should return true if the type of Pipeline is multi-branch-pipeline",
		pipeline: &Pipeline{
			Spec: PipelineSpec{
				Type: MultiBranchPipelineType,
			},
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pipeline.IsMultiBranch(); got != tt.want {
				t.Errorf("Pipeline.IsMultiBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiBranchPipeline_GetGitURL(t *testing.T) {
	type fields struct {
		SourceType            string
		GitSource             *GitSource
		GitHubSource          *GithubSource
		GitlabSource          *GitlabSource
		BitbucketServerSource *BitbucketServerSource
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "github",
		fields: fields{
			SourceType:   SourceTypeGithub,
			GitHubSource: &GithubSource{Owner: "linuxsuren", Repo: "tools"},
		},
		want: "https://github.com/linuxsuren/tools",
	}, {
		name: "gitlab",
		fields: fields{
			SourceType:   SourceTypeGitlab,
			GitlabSource: &GitlabSource{Owner: "linuxsuren", Repo: "tools"},
		},
		want: "https://gitlab.com/linuxsuren/tools",
	}, {
		name: "git",
		fields: fields{
			SourceType: SourceTypeGit,
			GitSource:  &GitSource{Url: "https://fake.com"},
		},
		want: "https://fake.com",
	}, {
		name: "bitbucket",
		fields: fields{
			SourceType:            SourceTypeBitbucket,
			BitbucketServerSource: &BitbucketServerSource{Owner: "linuxsuren", Repo: "tools"},
		},
		want: "https://bitbucket.org/linuxsuren/tools",
	}, {
		name: "fake",
		fields: fields{
			SourceType: "fake",
			GitSource:  &GitSource{Url: "https://fake.com"},
		},
		want: "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &MultiBranchPipeline{
				SourceType:            tt.fields.SourceType,
				GitSource:             tt.fields.GitSource,
				GitHubSource:          tt.fields.GitHubSource,
				GitlabSource:          tt.fields.GitlabSource,
				BitbucketServerSource: tt.fields.BitbucketServerSource,
			}
			assert.Equalf(t, tt.want, b.GetGitURL(), "GetGitURL()")
		})
	}
}
