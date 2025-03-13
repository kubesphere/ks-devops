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

	"github.com/go-logr/logr"
	mgrcore "github.com/kubesphere/ks-devops/controllers/core"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	}, {
		name: "gitlab, have owner and repo, but without URL",
		args: args{
			repo: &v1alpha3.GitRepository{
				Spec: v1alpha3.GitRepositorySpec{
					Provider: "gitlab",
					Owner:    "linuxsuren",
					Repo:     "test",
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
			amend := gitlabPublicAmend{}
			if !amend.Match(tt.args.repo) {
				return
			}
			assert.Equalf(t, tt.wantChanged, amend.Amend(tt.args.repo), "amendGitlabURL(%v)", tt.args.repo)
			tt.verify(t, tt.args.repo)
		})
	}
}

func TestAmendReconciler_SetupWithManager(t *testing.T) {
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
			r := &AmendReconciler{
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

func TestAmendReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	repo := &v1alpha3.GitRepository{}
	repo.SetName("fake")
	repo.SetNamespace("ns")
	repo.Spec.Provider = "gitlab"
	repo.Spec.Owner = "linuxsuren"
	repo.Spec.Repo = "test"
	repo.Spec.URL = "https://gitlab.com/linuxsuren/test"

	ghRepo := repo.DeepCopy()
	ghRepo.Spec.Provider = "github"
	ghRepo.Spec.URL = ""

	bitbucket := ghRepo.DeepCopy()
	bitbucket.Spec.Provider = "bitbucket"

	type fields struct {
		Client client.Client
	}
	type args struct {
		req controllerruntime.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult controllerruntime.Result
		wantErr    assert.ErrorAssertionFunc
		verify     func(tt *testing.T, c client.Client)
	}{{
		name: "not found",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).Build(),
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name: "gitlab case",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(repo.DeepCopy()).Build(),
		},
		args: args{
			req: controllerruntime.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name: "github case",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(ghRepo.DeepCopy()).Build(),
		},
		args: args{
			req: controllerruntime.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
		verify: func(tt *testing.T, c client.Client) {
			repo := &v1alpha3.GitRepository{}
			err := c.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "fake",
			}, repo)
			assert.Nil(tt, err)
			assert.Equal(tt, "https://github.com/linuxsuren/test", repo.Spec.URL)
		},
	}, {
		name: "bitbucket case",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(bitbucket.DeepCopy()).Build(),
		},
		args: args{
			req: controllerruntime.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
		verify: func(tt *testing.T, c client.Client) {
			repo := &v1alpha3.GitRepository{}
			err := c.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "fake",
			}, repo)
			assert.Nil(tt, err)
			assert.Equal(tt, "https://bitbucket.org/linuxsuren/test", repo.Spec.URL)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AmendReconciler{
				Client: tt.fields.Client,
				log:    logr.New(log.NullLogSink{}),
			}
			gotResult, err := r.Reconcile(context.Background(), tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("Reconcile(%v)", tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.wantResult, gotResult, "Reconcile(%v)", tt.args.req)
			if tt.verify != nil {
				tt.verify(t, tt.fields.Client)
			}
		})
	}
}
