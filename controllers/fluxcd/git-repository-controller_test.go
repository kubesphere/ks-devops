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

package fluxcd

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestGitRepositoryReconciler_reconcileFluxGitRepo(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	repo := v1alpha3.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "devops-project",
		},
		Spec: v1alpha3.GitRepositorySpec{
			Provider: "git",
			URL:      "https://example.com/fake/fake",
			Secret:   nil,
		},
	}

	repoWithAuth := repo.DeepCopy()
	repoWithAuth.Spec.Secret = &v1.SecretReference{
		Name:      "fake-auth",
		Namespace: "devops-project",
	}

	now := metav1.Now()
	repoWithDeletion := repoWithAuth.DeepCopy()
	repoWithDeletion.DeletionTimestamp = &now

	repoWithNewURL := repoWithAuth.DeepCopy()
	repoWithNewURL.Spec.URL = "https://example.com/fake/real"

	FluxGitRepo := createBareFluxGitRepoObject()
	preDelFluxGitRepo := createUnstructuredFluxGitRepo(repoWithAuth)

	type fields struct {
		Client client.Client
	}
	type args struct {
		repo *v1alpha3.GitRepository
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(t *testing.T, Client client.Client, err error)
	}{
		{
			name: "create a git repository without Secret",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy()),
			},
			args: args{
				repo: repo.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(repo.GetName()),
					Namespace: repo.GetNamespace(),
				}, gitrepo)
				assert.Nil(t, err)
				url, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "url")
				assert.Equal(t, "https://example.com/fake/fake", url)
				secretName, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "secretRef", "name")
				assert.Equal(t, "", secretName)
			},
		},
		{
			name: "create a git repository with Secret",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, repoWithAuth.DeepCopy()),
			},
			args: args{
				repo: repoWithAuth.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(repo.GetName()),
					Namespace: repo.GetNamespace(),
				}, gitrepo)
				assert.Nil(t, err)
				url, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "url")
				assert.Equal(t, "https://example.com/fake/fake", url)
				secretName, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "secretRef", "name")
				assert.Equal(t, "fake-auth", secretName)
			},
		},
		{
			name: "delete a git repository with Secret",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, repoWithAuth.DeepCopy(), preDelFluxGitRepo.DeepCopy()),
			},
			args: args{
				repo: repoWithDeletion.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(repo.GetName()),
					Namespace: repo.GetNamespace(),
				}, gitrepo)
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
		{
			name: "delete a git repository but not found a corresponding flux git repository",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, repoWithAuth.DeepCopy()),
			},
			args: args{
				repo: repoWithDeletion.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(repo.GetName()),
					Namespace: repo.GetNamespace(),
				}, gitrepo)
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
		{
			name: "update a git repository with Secret",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, repoWithAuth.DeepCopy(), preDelFluxGitRepo.DeepCopy()),
			},
			args: args{
				repo: repoWithNewURL.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(repo.GetName()),
					Namespace: repo.GetNamespace(),
				}, gitrepo)
				assert.Nil(t, err)
				url, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "url")
				assert.Equal(t, "https://example.com/fake/real", url)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &GitRepositoryReconciler{
				Client:   tt.fields.Client,
				log:      log.NullLogger{},
				recorder: &record.FakeRecorder{},
			}
			err := r.reconcileFluxGitRepo(tt.args.repo)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func TestGitRepositoryReconciler_getFluxRepoName(t *testing.T) {
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
			name: "fake-git-repo",
		},
		want: "fluxcd-fake-git-repo",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getFluxRepoName(tt.args.name), "getSecretName(%v)", tt.args.name)
		})
	}
}

func TestGitRepositoryReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	repo := v1alpha3.GitRepository{}
	repo.SetName("fake-git-repo")
	repo.SetNamespace("devops-project")

	type fields struct {
		Client client.Client
	}
	type args struct {
		req ctrl.Request
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult ctrl.Result
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "not found",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "fake-git-repo",
						Namespace: "another-devops-project",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "found",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "fake-git-repo",
						Namespace: "devops-project",
					},
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &GitRepositoryReconciler{
				Client:   tt.fields.Client,
				recorder: &record.FakeRecorder{},
			}
			gotResult, err := r.Reconcile(tt.args.req)
			if tt.wantErr(t, err) {
				assert.Equal(t, tt.wantResult, gotResult)
			}
		})
	}
}

func TestGitRepositoryReconciler_GetName(t *testing.T) {
	t.Run("get GitRepositoryReconciler name", func(t *testing.T) {

		r := &GitRepositoryReconciler{}
		assert.Equal(t, "FluxGitRepositoryReconciler", r.GetName())
	})
}
