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
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/bitbucket"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"github.com/jenkins-x/go-scm/scm/driver/gitlab"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"net/http"
	"testing"
)

func Test_getSCMClient(t *testing.T) {
	type args struct {
		request func() *http.Request
	}
	tests := []struct {
		name string
		args args
		want *scm.Client
	}{{
		name: "github",
		args: args{
			request: func() *http.Request {
				defaultRequest := &http.Request{}
				defaultRequest.Header = map[string][]string{}
				defaultRequest.Header.Add("X-GitHub-Event", "event")
				return defaultRequest
			},
		},
		want: github.NewDefault(),
	}, {
		name: "gtilab",
		args: args{
			request: func() *http.Request {
				defaultRequest := &http.Request{}
				defaultRequest.Header = map[string][]string{}
				defaultRequest.Header.Add("X-Gitlab-Event", "event")
				return defaultRequest
			},
		},
		want: gitlab.NewDefault(),
	}, {
		name: "bitbucket",
		args: args{
			request: func() *http.Request {
				defaultRequest := &http.Request{}
				defaultRequest.Header = map[string][]string{}
				defaultRequest.Header.Add("User-Agent", "Bitbucket-Webhooks")
				return defaultRequest
			},
		},
		want: bitbucket.NewDefault(),
	}, {
		name: "unknown SCM provider",
		args: args{
			request: func() *http.Request {
				return &http.Request{}
			},
		},
		want: nil,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getSCMClient(tt.args.request()), "getSCMClient(%v)", tt.args.request())
		})
	}
}

func Test_branchMatch(t *testing.T) {
	type args struct {
		pipeline v1alpha3.Pipeline
		branch   string
	}
	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{{
		name: "no any annotations",
		args: args{
			pipeline: v1alpha3.Pipeline{},
			branch:   "master",
		},
		wantOk: true,
	}, {
		name: "branch name equal literally",
		args: args{
			pipeline: v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						scmRefAnnotationKey: `["master", "good"]`,
					},
				},
			},
			branch: "master",
		},
		wantOk: true,
	}, {
		name: "branch name equal in regex",
		args: args{
			pipeline: v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						scmRefAnnotationKey: `["feat-.*", "good"]`,
					},
				},
			},
			branch: "feat-login",
		},
		wantOk: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantOk, branchMatch(tt.args.pipeline, tt.args.branch), "branchMatch(%v, %v)", tt.args.pipeline, tt.args.branch)
		})
	}
}
