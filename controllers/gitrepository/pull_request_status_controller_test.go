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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	mgrcore "kubesphere.io/devops/controllers/core"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var mockHeaders = map[string]string{
	"X-GitHub-Request-Id":   "DD0E:6011:12F21A8:1926790:5A2064E2",
	"X-RateLimit-Limit":     "60",
	"X-RateLimit-Remaining": "59",
	"X-RateLimit-Reset":     "1512076018",
}

func TestCreatePullRequestStatus(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/pulls/1347").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr.json")

	gock.New("https://api.github.com").
		Post("/repos/octocat/hello-world/statuses/6dcb09b5b57875f334f61aebed695e2e4193db5e").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/status.json")

	maker := NewStatusMaker("octocat/hello-world", "")
	maker.WithTarget("https://ci.example.com/1000/output").WithPR(1347)

	err := maker.Create(context.Background(), scm.StateSuccess, "continuous-integration/drone", "Build has completed successfully")
	assert.Nil(t, err)

	maker.WithProvider("fake-provider").WithServer("fake-server")
	err = maker.Create(context.Background(), scm.StateSuccess, "continuous-integration/drone", "Build has completed successfully")
	assert.NotNil(t, err)
}

func TestPullRequestStatusReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "normal",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PullRequestStatusReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			mgr := &mgrcore.FakeManager{
				Client: tt.fields.Client,
				Scheme: schema,
			}
			tt.wantErr(t, r.SetupWithManager(mgr), fmt.Sprintf("SetupWithManager(%v)", mgr))
		})
	}
}

func TestPullRequestStatusReconciler(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type request struct {
		name, namespace string
	}
	defaultReq := request{namespace: "ns", name: "fake"}

	secret := &v1.Secret{}
	secret.SetName("token")
	secret.SetNamespace(defaultReq.namespace)
	secret.Type = v1.SecretTypeBasicAuth
	secret.Data = map[string][]byte{
		v1.BasicAuthPasswordKey: []byte(""),
	}

	pipRun := &v1alpha3.PipelineRun{}
	pipRun.SetName(defaultReq.name)
	pipRun.SetNamespace(defaultReq.namespace)
	pipRun.Spec = v1alpha3.PipelineRunSpec{
		SCM: &v1alpha3.SCM{
			RefName: "PR-1347",
		},
		PipelineRef: &v1.ObjectReference{
			Name: "pipeline",
		},
		PipelineSpec: &v1alpha3.PipelineSpec{
			Type: "multi-branch-pipeline",
			MultiBranchPipeline: &v1alpha3.MultiBranchPipeline{
				SourceType: v1alpha3.SourceTypeGithub,
				GitHubSource: &v1alpha3.GithubSource{
					CredentialId: "token",
					Owner:        "octocat",
					Repo:         "hello-world",
				},
			},
		},
	}
	pipRun.Status.Phase = "Succeeded"

	finishedTime := "2022-10-11T15:16:13Z"
	timeLayout := "2006-01-02T15:04:05Z"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, finishedTime, loc)
	finalTime := metav1.NewTime(theTime)
	pipRun.Status.CompletionTime = &finalTime

	project := &v1alpha3.DevOpsProject{}
	project.SetName(defaultReq.namespace)
	project.Labels = map[string]string{
		"kubesphere.io/workspace": "ws",
	}

	tests := []struct {
		name       string
		request    request
		prepare    func(*testing.T)
		k8sClient  client.Client
		wantResult ctrl.Result
		wantErr    bool
	}{{
		name:      "not found pipelinerun",
		request:   defaultReq,
		k8sClient: fake.NewClientBuilder().WithScheme(schema).Build(),

		wantErr: false,
	}, {
		name:    "pipeline with github",
		request: defaultReq,
		prepare: func(t *testing.T) {
			gock.New("https://api.github.com").
				Get("/repos/octocat/hello-world/pulls/1347").
				Reply(200).
				Type("application/json").
				SetHeaders(mockHeaders).
				File("testdata/pr.json")

			gock.New("https://api.github.com").
				Post("/repos/octocat/hello-world/statuses/6dcb09b5b57875f334f61aebed695e2e4193db5e").
				Reply(201).
				Type("application/json").
				SetHeaders(mockHeaders).
				File("testdata/status.json")
		},
		k8sClient: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(pipRun.DeepCopy(), secret.DeepCopy(), project.DeepCopy()).Build(),
		wantErr:   false,
	}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			if tt.prepare != nil {
				tt.prepare(t)
			}
			recon := &PullRequestStatusReconciler{
				log:    logr.New(log.NullLogSink{}),
				Client: tt.k8sClient,
			}

			result, err := recon.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: tt.request.namespace,
					Name:      tt.request.name,
				},
			})

			assert.Equal(t, tt.wantResult, result)
			if tt.wantErr {
				assert.NotNil(t, err, "should have error in case [%s]-[%d]", tt.name, i)
			} else {
				assert.Nil(t, err, "should not have error in case [%s]-[%d]", tt.name, i)
			}
		})
	}
}

func Test(t *testing.T) {
	tests := []struct {
		name    string
		pr      string
		wantNum int
		wantErr bool
	}{{
		name:    "lower case",
		pr:      "pr-1",
		wantNum: 1,
		wantErr: false,
	}, {
		name:    "lower and upper case",
		pr:      "Pr-1",
		wantNum: 1,
		wantErr: false,
	}, {
		name:    "upper case",
		pr:      "PR-1",
		wantNum: 1,
		wantErr: false,
	}, {
		name:    "upper case with MR",
		pr:      "MR-3",
		wantNum: 3,
	}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, err := getPRNumber(tt.pr)
			assert.Equal(t, tt.wantNum, num)
			if tt.wantErr {
				assert.NotNil(t, err, "should have error in case [%d]-[%s]", i, tt.name)
			} else {

				assert.Nil(t, err, "should not have error in case [%d]-[%s]", i, tt.name)
			}
		})
	}
}

func TestConvertPipelineRunPhaseToSCMStatus(t *testing.T) {
	tests := []struct {
		name       string
		phase      v1alpha3.RunPhase
		wantStatus scm.State
	}{{
		name:       "success",
		phase:      v1alpha3.Succeeded,
		wantStatus: scm.StateSuccess,
	}, {
		name:       "cancelled",
		phase:      v1alpha3.Cancelled,
		wantStatus: scm.StateCanceled,
	}, {
		name:       "failure",
		phase:      v1alpha3.Failed,
		wantStatus: scm.StateFailure,
	}, {
		name:       "pendding",
		phase:      v1alpha3.Pending,
		wantStatus: scm.StatePending,
	}, {
		name:       "unknown",
		phase:      v1alpha3.Unknown,
		wantStatus: scm.StateUnknown,
	}, {
		name:       "running",
		phase:      v1alpha3.Running,
		wantStatus: scm.StateRunning,
	}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := convertPipelineRunPhaseToSCMStatus(tt.phase)
			assert.Equal(t, tt.wantStatus, status, "failed in case [%d]", i)
		})
	}
}

func TestGetRepoInfo(t *testing.T) {
	emptyRepoInfo := repoInfo{}

	tests := []struct {
		name     string
		repo     *v1alpha3.MultiBranchPipeline
		wantInfo repoInfo
	}{{
		name:     "repo is nil",
		repo:     nil,
		wantInfo: emptyRepoInfo,
	}, {
		name: "github",
		repo: &v1alpha3.MultiBranchPipeline{
			SourceType: v1alpha3.SourceTypeGithub,
			GitHubSource: &v1alpha3.GithubSource{
				Owner:        "owner",
				Repo:         "repo",
				CredentialId: "token",
			},
		},
		wantInfo: repoInfo{owner: "owner", repo: "repo", tokenId: "token", provider: "github"},
	}, {
		name: "gitlab",
		repo: &v1alpha3.MultiBranchPipeline{
			SourceType: v1alpha3.SourceTypeGitlab,
			GitlabSource: &v1alpha3.GitlabSource{
				Owner:        "owner",
				Repo:         "repo",
				CredentialId: "token",
			},
		},
		wantInfo: repoInfo{owner: "owner", repo: "repo", tokenId: "token", provider: "gitlab"},
	}, {
		name: "bitbucket",
		repo: &v1alpha3.MultiBranchPipeline{
			SourceType: v1alpha3.SourceTypeBitbucket,
			BitbucketServerSource: &v1alpha3.BitbucketServerSource{
				ApiUri:       "https://bitbucket.org",
				Owner:        "owner",
				Repo:         "repo",
				CredentialId: "token",
			},
		},
		wantInfo: repoInfo{owner: "owner", repo: "repo", tokenId: "token", provider: "bitbucketcloud"},
	}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := getRepoInfo(tt.repo)
			assert.Equal(t, tt.wantInfo, info, "failed in case [%d]", i)
		})
	}
}
