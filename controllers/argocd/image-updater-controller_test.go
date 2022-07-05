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
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

func Test_updateImageList(t *testing.T) {
	type args struct {
		images      []string
		annotations map[string]string
	}
	tests := []struct {
		name   string
		args   args
		verify func(*testing.T, map[string]string)
	}{{
		name: "empty images and annotations",
		args: args{
			images:      []string{},
			annotations: map[string]string{},
		},
		verify: func(t *testing.T, m map[string]string) {
			assert.Equal(t, map[string]string{}, m)
		},
	}, {
		name: "clean the image list in the annotations",
		args: args{
			images: nil,
			annotations: map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx",
			},
		},
		verify: func(t *testing.T, m map[string]string) {
			assert.Equal(t, map[string]string{}, m)
		},
	}, {
		name: "have different items",
		args: args{
			images: []string{"nginx", "alpine"},
			annotations: map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "a, b",
			},
		},
		verify: func(t *testing.T, m map[string]string) {
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx,alpine",
			}, m)
		},
	}, {
		name: "put an image to the empty annotations",
		args: args{
			images:      []string{"nginx"},
			annotations: map[string]string{},
		},
		verify: func(t *testing.T, m map[string]string) {
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx",
			}, m)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateImageList(tt.args.images, tt.args.annotations)
			tt.verify(t, tt.args.annotations)
		})
	}
}

func TestReconcile(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	app := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "fake",
		},
	}

	updater := &v1alpha1.ImageUpdater{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "updater",
			Namespace: "fake",
		},
		Spec: v1alpha1.ImageUpdaterSpec{
			Kind:   "argocd",
			Images: []string{"nginx"},
			Argo: &v1alpha1.ArgoImageUpdater{
				App:   v1.LocalObjectReference{Name: "app"},
				Write: v1alpha1.WriteMethodBuiltIn,
			},
		},
	}
	notArgoKindUpdater := updater.DeepCopy()
	notArgoKindUpdater.Spec.Kind = "fake"
	noArgoSettingUpdater := updater.DeepCopy()
	noArgoSettingUpdater.Spec.Argo = nil
	emptyAppName := updater.DeepCopy()
	emptyAppName.Spec.Argo.App.Name = ""
	invalidWriteMethod := updater.DeepCopy()
	invalidWriteMethod.Spec.Argo.Write = "invalid"
	withSecretUpdater := updater.DeepCopy()
	withSecretUpdater.Spec.Argo.Secrets = map[string]string{
		"nginx": "ns/secret",
	}

	defaultReq := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "fake",
			Name:      "updater",
		},
	}

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
		name: "no imageUpdaters found",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, app.DeepCopy(), updater.DeepCopy()),
		},
		args: args{
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "fake",
					Name:      "fake",
				},
			},
		},
		wantResult: ctrl.Result{},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name: "kind is not argocd",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, notArgoKindUpdater.DeepCopy()),
		},
		args: args{
			req: defaultReq,
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return true
		},
		wantResult: ctrl.Result{},
	}, {
		name: "argo setting is nil",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, noArgoSettingUpdater),
		},
		args: args{req: defaultReq},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return true
		},
		wantResult: ctrl.Result{},
	}, {
		name: "argo app name is empty",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, emptyAppName),
		},
		args: args{req: defaultReq},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return true
		},
		wantResult: ctrl.Result{},
	}, {
		name: "cannot found app",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, updater),
		},
		args: args{req: defaultReq},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return true
		},
		wantResult: ctrl.Result{RequeueAfter: time.Minute},
	}, {
		name: "normal case",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, updater, app),
		},
		args: args{req: defaultReq},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			c := i[1].(client.Client)
			assert.Nil(t, err)

			resultApp := &v1alpha1.Application{}
			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "app",
			}, resultApp)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list":        "nginx",
				"argocd-image-updater.argoproj.io/write-back-method": "argocd",
			}, resultApp.Annotations)

			return true
		},
		wantResult: ctrl.Result{},
	}, {
		name: "invalid write method",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, invalidWriteMethod, app),
		},
		args: args{req: defaultReq},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			c := i[1].(client.Client)
			assert.Nil(t, err)

			resultApp := &v1alpha1.Application{}
			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "app",
			}, resultApp)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx",
			}, resultApp.Annotations)

			return true
		},
		wantResult: ctrl.Result{},
	}, {
		name: "with image secret",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, withSecretUpdater, app),
		},
		args: args{req: defaultReq},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			c := i[1].(client.Client)
			assert.Nil(t, err)

			resultApp := &v1alpha1.Application{}
			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "app",
			}, resultApp)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list":        "nginx",
				"argocd-image-updater.argoproj.io/write-back-method": "argocd",
				"argocd-image-updater.argoproj.io/nginx.pull-secret": "pullsecret:ns/secret",
			}, resultApp.Annotations)

			return true
		},
		wantResult: ctrl.Result{},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ImageUpdaterReconciler{
				Client:   tt.fields.Client,
				log:      log.NullLogger{},
				recorder: &record.FakeRecorder{},
			}
			gotResult, err := r.Reconcile(tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("Reconcile(%v)", tt.args.req), tt.fields.Client) {
				return
			}
			assert.Equalf(t, tt.wantResult, gotResult, "Reconcile(%v)", tt.args.req)
		})
	}
}

func Test_setImagePreference(t *testing.T) {
	type args struct {
		argo        *v1alpha1.ArgoImageUpdater
		annotations map[string]string
	}
	tests := []struct {
		name   string
		args   args
		verify func(*testing.T, map[string]string)
	}{{
		name: "normal case",
		args: args{
			argo: &v1alpha1.ArgoImageUpdater{
				Secrets: map[string]string{
					"nginx": "ns/name",
				},
				AllowTags: map[string]string{
					"nginx": "v1.1.1",
				},
				IgnoreTags: map[string]string{
					"nginx": "latest",
				},
				UpdateStrategy: map[string]string{
					"nginx": "semver",
				},
				Platforms: map[string]string{
					"nginx": "linux/amd64",
				},
			},
			annotations: map[string]string{},
		},
		verify: func(t *testing.T, m map[string]string) {
			assert.Equal(t, "pullsecret:ns/name", m["argocd-image-updater.argoproj.io/nginx.pull-secret"])
			assert.Equal(t, "v1.1.1", m["argocd-image-updater.argoproj.io/nginx.allow-tags"])
			assert.Equal(t, "latest", m["argocd-image-updater.argoproj.io/nginx.ignore-tags"])
			assert.Equal(t, "semver", m["argocd-image-updater.argoproj.io/nginx.update-strategy"])
			assert.Equal(t, "linux/amd64", m["argocd-image-updater.argoproj.io/nginx.platforms"])
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setImagePreference(tt.args.argo, tt.args.annotations)
		})
	}
}
