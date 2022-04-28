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

package gitrepository

import (
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"testing"
)

func Test_amendGitlabURL(t *testing.T) {
	type args struct {
		repo *v1alpha3.GitRepository
	}
	tests := []struct {
		name        string
		args        args
		wantChanged bool
		verify      func(t *testing.T, repo *v1alpha3.GitRepository)
	}{{
		name: "not gitlab",
		args: args{
			repo: &v1alpha3.GitRepository{},
		},
		wantChanged: false,
	}, {
		name: "gitlab, URL without suffix .git",
		args: args{
			repo: &v1alpha3.GitRepository{
				Spec: v1alpha3.GitRepositorySpec{
					Provider: "Gitlab",
					URL:      "https://gitlab.com/linuxsuren/test",
				},
			},
		},
		wantChanged: true,
		verify: func(t *testing.T, repo *v1alpha3.GitRepository) {
			assert.Equal(t, "https://gitlab.com/linuxsuren/test.git", repo.Spec.URL)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.verify == nil {
				tt.verify = func(t *testing.T, repo *v1alpha3.GitRepository) {
				}
			}
			assert.Equalf(t, tt.wantChanged, amendGitlabURL(tt.args.repo), "amendGitlabURL(%v)", tt.args.repo)
			tt.verify(t, tt.args.repo)
		})
	}
}
