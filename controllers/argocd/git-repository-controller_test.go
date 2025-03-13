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

package argocd

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kubesphere/ks-devops/controllers/core"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func Test_getSecretName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "normal case",
		args: args{
			name: "name",
		},
		want: "name-repo",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getSecretName(tt.args.name), "getSecretName(%v)", tt.args.name)
		})
	}
}

func TestGitRepositoryController_Reconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	repo := v1alpha3.GitRepository{}
	repo.SetNamespace("ns")
	repo.SetName("fake")

	repoWithAuth := repo.DeepCopy()
	repoWithAuth.Spec.Secret = &v1.SecretReference{
		Namespace: "argocd",
		Name:      "fake-repo",
	}

	now := metav1.Now()
	repoWithDeletion := repo.DeepCopy()
	repoWithDeletion.DeletionTimestamp = &now
	repoWithDeletion.SetFinalizers([]string{"fake"})

	secret := v1.Secret{}
	secret.SetNamespace("argocd")
	secret.SetName("fake-repo")
	secret.Type = v1.SecretTypeBasicAuth

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
	}{{
		name: "not found a git repository",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).Build(),
		},
		args: args{
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}, {
		name: "create a git repository, without argo secret",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(repo.DeepCopy()).Build(),
		},
		args: args{
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}, {
		name: "create a git repository, with an argo secret",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(repo.DeepCopy(), secret.DeepCopy()).Build(),
		},
		args: args{
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}, {
		name: "create a git repository with auth secret, with an argo secret",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(repoWithAuth.DeepCopy(), secret.DeepCopy()).Build(),
		},
		args: args{
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}, {
		name: "delete a git repository, without argo secret",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(repoWithDeletion.DeepCopy()).Build(),
		},
		args: args{
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}, {
		name: "delete a git repository, with argo secret",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(repoWithDeletion.DeepCopy(), secret.DeepCopy()).Build(),
		},
		args: args{
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "ns",
					Name:      "fake",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GitRepositoryController{
				Client:        tt.fields.Client,
				log:           logr.New(log.NullLogSink{}),
				recorder:      &record.FakeRecorder{},
				ArgoNamespace: "argocd",
			}
			gotResult, err := c.Reconcile(context.Background(), tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("Reconcile(%v)", tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.wantResult, gotResult, "Reconcile(%v)", tt.args.req)
		})
	}
}

func TestGitRepositoryController_SetupWithManager(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type fields struct {
		Client        client.Client
		log           logr.Logger
		recorder      record.EventRecorder
		ArgoNamespace string
	}
	type args struct {
		mgr ctrl.Manager
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "normal",
		args: args{
			mgr: &core.FakeManager{
				Client: fake.NewClientBuilder().WithScheme(schema).Build(),
				Scheme: schema,
			},
		},
		wantErr: core.NoErrors,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GitRepositoryController{
				Client:        tt.fields.Client,
				log:           tt.fields.log,
				recorder:      tt.fields.recorder,
				ArgoNamespace: tt.fields.ArgoNamespace,
			}
			tt.wantErr(t, c.SetupWithManager(tt.args.mgr), fmt.Sprintf("SetupWithManager(%v)", tt.args.mgr))
		})
	}
}
