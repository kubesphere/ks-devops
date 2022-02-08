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
	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/go-logr/logr"
	testing2 "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
)

func Test_getRepo(t *testing.T) {
	type args struct {
		repo *v1alpha1.GitRepository
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "not supported provider",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{Provider: "fake"}},
		},
		want: "",
	}, {
		name: "provider is emtpy",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{
				URL: "https://github.com/linuxsuren/test",
			}},
		},
		want: "",
	}, {
		name: "github as the provider",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{
				Provider: "github",
				URL:      "https://github.com/linuxsuren/test",
			}},
		},
		want: "linuxsuren/test",
	}, {
		name: "gitlab as the provider",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{
				Provider: "gitlab",
				URL:      "https://gitlab.com/linuxsuren/test",
			}},
		},
		want: "linuxsuren/test",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRepo(tt.args.repo); got != tt.want {
				t.Errorf("getRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exist(t *testing.T) {
	type args struct {
		server string
		hooks  []*scm.Hook
	}
	tests := []struct {
		name      string
		args      args
		wantExist bool
		wantId    string
	}{{
		name: "not exist from empty",
		args: args{
			server: "fake",
			hooks:  nil,
		},
		wantExist: false,
		wantId:    "",
	}, {
		name: "not exist",
		args: args{
			server: "fake",
			hooks: []*scm.Hook{{
				Target: "random",
			}},
		},
		wantExist: false,
		wantId:    "",
	}, {
		name: "exist",
		args: args{
			server: "fake",
			hooks: []*scm.Hook{{
				ID:     "fake-id",
				Target: "fake",
			}, {
				ID:     "random-id",
				Target: "random",
			}},
		},
		wantExist: true,
		wantId:    "fake-id",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExist, gotId := exist(tt.args.server, tt.args.hooks)
			if gotExist != tt.wantExist {
				t.Errorf("exist() gotExist = %v, want %v", gotExist, tt.wantExist)
			}
			if gotId != tt.wantId {
				t.Errorf("exist() gotId = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func Test_linkToWebhook(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	repo := &v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "name",
		},
	}
	webhook := &v1alpha1.Webhook{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "webhook",
		},
	}
	webhookA := &v1alpha1.Webhook{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "webhook-a",
			Annotations: map[string]string{
				v1alpha1.AnnotationKeyGitRepos: "name",
			},
		},
	}

	type args struct {
		webhook v1.LocalObjectReference
		repo    *v1alpha1.GitRepository
		client  client.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		check   func(client client.Client) bool
	}{{
		name: "normal case",
		args: args{
			webhook: v1.LocalObjectReference{Name: "webhook"},
			repo: &v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "name",
				},
			},
			client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy(), webhook.DeepCopy()),
		},
		check: func(client client.Client) bool {
			webh := &v1alpha1.Webhook{}
			if err := client.Get(context.TODO(), types.NamespacedName{
				Namespace: "test",
				Name:      "webhook",
			}, webh); err != nil {
				return false
			}
			val := webh.Annotations[v1alpha1.AnnotationKeyGitRepos]
			return val == "name"
		},
		wantErr: false,
	}, {
		name: "not found the desired webhook",
		args: args{
			webhook: v1.LocalObjectReference{Name: "not-found"},
			repo: &v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "name",
				},
			},
			client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy(), webhook.DeepCopy()),
		},
		check: func(client client.Client) bool {
			webh := &v1alpha1.Webhook{}
			if err := client.Get(context.TODO(), types.NamespacedName{
				Namespace: "test",
				Name:      "webhook",
			}, webh); err != nil {
				return false
			}
			val := webh.Annotations[v1alpha1.AnnotationKeyGitRepos]
			return val == "name"
		},
		wantErr: true,
	}, {
		name: "has the same repo in the annotations",
		args: args{
			webhook: v1.LocalObjectReference{Name: "webhook-a"},
			repo: &v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "name",
				},
			},
			client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy(), webhookA.DeepCopy()),
		},
		check: func(client client.Client) (result bool) {
			webh := &v1alpha1.Webhook{}
			if err := client.Get(context.TODO(), types.NamespacedName{
				Namespace: "test",
				Name:      "webhook-a",
			}, webh); err != nil {
				return false
			}
			val := webh.Annotations[v1alpha1.AnnotationKeyGitRepos]
			result = val == "name"
			if !result {
				fmt.Printf("expect 'name', the actual value is '%s'\n", val)
			}
			return
		},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := linkToWebhook(tt.args.webhook, tt.args.repo, tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("linkToWebhook() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if !tt.check(tt.args.client) {
					t.Errorf("failed to do the check work: %s", tt.name)
				}
			}
		})
	}
}

func Test_addToSlick(t *testing.T) {
	type args struct {
		array    map[string]string
		key, val string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{{
		name: "the map is nil",
		args: args{
			array: nil,
			key:   "key",
			val:   "val",
		},
		want: map[string]string{"key": "val"},
	}, {
		name: "the map is empty",
		args: args{
			array: map[string]string{},
			key:   "key",
			val:   "val",
		},
		want: map[string]string{"key": "val"},
	}, {
		name: "have the same item in the map",
		args: args{
			array: map[string]string{"key": "value"},
			key:   "key",
			val:   "val",
		},
		want: map[string]string{"key": "value,val"},
	}, {
		name: "have not the same item in the map",
		args: args{
			array: map[string]string{"key": "val"},
			key:   "left",
			val:   "right",
		},
		want: map[string]string{"key": "val", "left": "right"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addToArrayInAnnotations(tt.args.array, tt.args.key, tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addToArrayInAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconciler_linkToWebhooks(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	repo := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "name",
		},
		Spec: v1alpha1.GitRepositorySpec{
			Webhooks: []v1.LocalObjectReference{{
				Name: "webhook",
			}},
		},
	}
	webhook := v1alpha1.Webhook{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "webhook",
		},
	}

	type fields struct {
		Client client.Client
		log    logr.Logger
	}
	type args struct {
		repo *v1alpha1.GitRepository
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		check   func(client client.Client) bool
	}{{
		name: "normal case",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy(), webhook.DeepCopy()),
			log:    testing2.NullLogger{},
		},
		args: args{
			repo: &repo,
		},
		wantErr: false,
		check: func(client client.Client) bool {
			webh := &v1alpha1.Webhook{}
			if err := client.Get(context.TODO(), types.NamespacedName{
				Namespace: "test",
				Name:      "webhook",
			}, webh); err != nil {
				return false
			}
			val := webh.Annotations[v1alpha1.AnnotationKeyGitRepos]
			return val == "name"
		},
	}, {
		name: "has errors",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, repo.DeepCopy()),
			log:    testing2.NullLogger{},
		},
		args: args{
			repo: &repo,
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client: tt.fields.Client,
				log:    tt.fields.log,
			}
			if err := r.linkToWebhooks(tt.args.repo); (err != nil) != tt.wantErr {
				t.Errorf("linkToWebhooks() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if !tt.check(r.Client) {
					t.Errorf("failed to do the check work: %s", tt.name)
				}
			}
		})
	}
}

func TestReconciler_getTokenFromSecret(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	basicSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basicSecret",
			Namespace: "ns",
		},
		Type: v1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			v1.BasicAuthPasswordKey: []byte("token"),
		},
	}
	opaqueSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "opaqueSecret",
			Namespace: "ns",
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			v1.ServiceAccountTokenKey: []byte("token"),
		},
	}

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		secretRef        *v1.SecretReference
		defaultNamespace string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantToken string
		wantErr   assert.ErrorAssertionFunc
	}{{
		name: "normal case, basic auth secret",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, basicSecret.DeepCopy()),
		},
		args: args{
			secretRef: &v1.SecretReference{
				Name:      "basicSecret",
				Namespace: "ns",
			},
			defaultNamespace: "ns",
		},
		wantToken: "token",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return false
		},
	}, {
		name: "normal case, opaque secret",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, opaqueSecret.DeepCopy()),
		},
		args: args{
			secretRef: &v1.SecretReference{
				Name:      "opaqueSecret",
				Namespace: "ns",
			},
			defaultNamespace: "ns",
		},
		wantToken: "token",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return false
		},
	}, {
		name: "normal case, no namespace in the SecretReference",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, opaqueSecret.DeepCopy()),
		},
		args: args{
			secretRef: &v1.SecretReference{
				Name: "opaqueSecret",
			},
			defaultNamespace: "ns",
		},
		wantToken: "token",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return false
		},
	}, {
		name: "error case, not exist secret",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema),
		},
		args: args{
			secretRef: &v1.SecretReference{
				Name:      "opaqueSecret",
				Namespace: "ns",
			},
			defaultNamespace: "ns",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			gotToken, err := r.getTokenFromSecret(tt.args.secretRef, tt.args.defaultNamespace)
			if !tt.wantErr(t, err, fmt.Sprintf("getTokenFromSecret(%v, %v)", tt.args.secretRef, tt.args.defaultNamespace)) {
				return
			}
			assert.Equalf(t, tt.wantToken, gotToken, "getTokenFromSecret(%v, %v)", tt.args.secretRef, tt.args.defaultNamespace)
		})
	}
}

func TestReconciler_getGitClient(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	basicSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basicSecret",
			Namespace: "ns",
		},
		Type: v1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			v1.BasicAuthPasswordKey: []byte("token"),
		},
	}
	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		repo *v1alpha1.GitRepository
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantClient *scm.Client
		wantErr    assert.ErrorAssertionFunc
	}{{
		name: "not support git provider",
		args: args{
			repo: &v1alpha1.GitRepository{
				Spec: v1alpha1.GitRepositorySpec{
					Provider: "not-support",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err, i)
			assert.Equal(t, strings.HasPrefix(err.Error(), "not support git provider: "), true, i)
			return true
		},
	}, {
		name: "no secret found",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema),
		},
		args: args{
			repo: &v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns"},
				Spec: v1alpha1.GitRepositorySpec{
					Provider: "github",
					Secret: &v1.SecretReference{
						Name: "basicSecret",
					},
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err, i)
			return true
		},
		wantClient: github.NewDefault(),
	}, {
		name: "github provider",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, basicSecret.DeepCopy()),
		},
		args: args{
			repo: &v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns"},
				Spec: v1alpha1.GitRepositorySpec{
					Provider: "github",
					Secret: &v1.SecretReference{
						Name: "basicSecret",
					},
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err, i)
			return false
		},
	}, {
		name: "gitlab provider",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, basicSecret.DeepCopy()),
		},
		args: args{
			repo: &v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns"},
				Spec: v1alpha1.GitRepositorySpec{
					Provider: "gitlab",
					Secret: &v1.SecretReference{
						Name: "basicSecret",
					},
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err, i)
			return false
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			gotClient, err := r.getGitClient(tt.args.repo)
			if !tt.wantErr(t, err, fmt.Sprintf("getGitClient(%v)", tt.args.repo)) {
				return
			}
			assert.Equalf(t, tt.wantClient, gotClient, "getGitClient(%v)", tt.args.repo)
		})
	}
}
