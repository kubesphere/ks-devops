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
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestGitRepositoryReconciler_reconcileFluxGitRepo(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	NonArtifactRepo := &v1alpha3.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-repo",
			Namespace: "fake-ns",
			Labels: map[string]string{
				v1alpha1.ArtifactRepoLabelKey: "false",
			},
			ResourceVersion: "999",
		},
		Spec: v1alpha3.GitRepositorySpec{
			Provider: "git",
			URL:      "https://fakeGitHub.com/faker/fake-project",
			Secret: &v1.SecretReference{
				Name:      "fake-secret",
				Namespace: "fake-ns",
			},
		},
	}

	ArtifactRepo := NonArtifactRepo.DeepCopy()
	ArtifactRepo.SetLabels(map[string]string{
		v1alpha1.ArtifactRepoLabelKey: "true",
	})

	now := metav1.Now()
	ArtifactRepoWithDeletion := ArtifactRepo.DeepCopy()
	ArtifactRepoWithDeletion.DeletionTimestamp = &now

	ArtifactRepoWithNewURL := ArtifactRepo.DeepCopy()
	ArtifactRepoWithNewURL.Spec.URL = "https://fakeGitHub.com/faker/another-fake-project"

	FluxGitRepo := createBareFluxGitRepoObject()
	preDelFluxGitRepo := createUnstructuredFluxGitRepo(ArtifactRepo)

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
			name: "create a Non-Artifact git repository",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, NonArtifactRepo.DeepCopy()),
			},
			args: args{
				repo: NonArtifactRepo.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(ArtifactRepo.GetName()),
					Namespace: ArtifactRepo.GetNamespace(),
				}, gitrepo)
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
		{
			name: "create a Artifact git repository",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, ArtifactRepo.DeepCopy()),
			},
			args: args{
				repo: ArtifactRepo.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(ArtifactRepo.GetName()),
					Namespace: ArtifactRepo.GetNamespace(),
				}, gitrepo)
				assert.Nil(t, err)
				url, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "url")
				assert.Equal(t, "https://fakeGitHub.com/faker/fake-project", url)
				secretName, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "secretRef", "name")
				assert.Equal(t, "fake-secret", secretName)
				ok := gitrepo.GetLabels()[v1alpha1.GroupName]
				assert.True(t, true, ok)
			},
		},
		{
			name: "delete a Artifact git repository by click the delete button",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, ArtifactRepo.DeepCopy(), preDelFluxGitRepo.DeepCopy()),
			},
			args: args{
				repo: ArtifactRepoWithDeletion.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(ArtifactRepo.GetName()),
					Namespace: ArtifactRepo.GetNamespace(),
				}, gitrepo)
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
		{
			name: "update a Artifact git repository (delete the ArtifactRepo Label)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, ArtifactRepo.DeepCopy(), preDelFluxGitRepo.DeepCopy()),
			},
			args: args{
				repo: NonArtifactRepo.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(ArtifactRepo.GetName()),
					Namespace: ArtifactRepo.GetNamespace(),
				}, gitrepo)
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
		{
			name: "update a Artifact git repository (normal case)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, ArtifactRepo.DeepCopy(), preDelFluxGitRepo.DeepCopy()),
			},
			args: args{
				repo: ArtifactRepoWithNewURL.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(ArtifactRepo.GetName()),
					Namespace: ArtifactRepo.GetNamespace(),
				}, gitrepo)
				assert.Nil(t, err)
				url, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "url")
				assert.Equal(t, "https://fakeGitHub.com/faker/another-fake-project", url)
			},
		},
		{
			name: "update a Non-Artifact git repository (add the ArtifactRepo Label)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, NonArtifactRepo.DeepCopy()),
			},
			args: args{
				repo: ArtifactRepo.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				gitrepo := FluxGitRepo.DeepCopy()
				err = Client.Get(context.TODO(), types.NamespacedName{
					Name:      getFluxRepoName(NonArtifactRepo.GetName()),
					Namespace: NonArtifactRepo.GetNamespace(),
				}, gitrepo)
				assert.Nil(t, err)
				url, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "url")
				assert.Equal(t, "https://fakeGitHub.com/faker/fake-project", url)
				secretName, _, _ := unstructured.NestedString(gitrepo.Object, "spec", "secretRef", "name")
				assert.Equal(t, "fake-secret", secretName)
				ok := gitrepo.GetLabels()[v1alpha1.GroupName]
				assert.True(t, true, ok)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &GitRepositoryReconciler{
				Client:   tt.fields.Client,
				log:      logr.New(log.NullLogSink{}),
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

	repo := v1alpha3.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-repo",
			Namespace: "fake-ns",
		},
	}

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
						Name:      "fake-repo",
						Namespace: "another-fake-ns",
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
						Name:      "fake-repo",
						Namespace: "fake-ns",
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
			gotResult, err := r.Reconcile(context.Background(), tt.args.req)
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
