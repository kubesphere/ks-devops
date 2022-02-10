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

package addon

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestReconciler_supportedStrategy(t *testing.T) {
	type args struct {
		strategy *v1alpha1.AddonStrategy
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "argument is nil",
		want: false,
	}, {
		name: "empty object",
		args: args{strategy: &v1alpha1.AddonStrategy{}},
		want: false,
	}, {
		name: "fake type",
		args: args{strategy: &v1alpha1.AddonStrategy{
			Spec: v1alpha1.AddStrategySpec{Type: "fake"},
		}},
		want: false,
	}, {
		name: "simple-operator",
		args: args{strategy: &v1alpha1.AddonStrategy{
			Spec: v1alpha1.AddStrategySpec{Type: "simple-operator"},
		}},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{}
			assert.Equalf(t, tt.want, r.supportedStrategy(tt.args.strategy), "supportedStrategy(%v)", tt.args.strategy)
		})
	}
}

func Test_getTemplate(t *testing.T) {
	type args struct {
		tpl   string
		addon *v1alpha1.Addon
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
		wantErr    assert.ErrorAssertionFunc
	}{{
		name: "addon is nil",
		args: args{
			tpl: "this is a template",
		},
		wantResult: "this is a template",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}, {
		name: "addon with version",
		args: args{
			tpl:   "{{.spec.version}}",
			addon: &v1alpha1.Addon{Spec: v1alpha1.AddonSpec{Version: "v1alpha1"}},
		},
		wantResult: "v1alpha1",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := getTemplate(tt.args.tpl, tt.args.addon)
			if !tt.wantErr(t, err, fmt.Sprintf("getTemplate(%v, %v)", tt.args.tpl, tt.args.addon)) {
				return
			}
			assert.Equalf(t, tt.wantResult, gotResult, "getTemplate(%v, %v)", tt.args.tpl, tt.args.addon)
		})
	}
}

func TestReconciler_addonHandle(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	strategy := &v1alpha1.AddonStrategy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "simple-operator",
		},
		Spec: v1alpha1.AddStrategySpec{
			Available:      true,
			Type:           v1alpha1.AddonInstallStrategySimpleOperator,
			SimpleOperator: v1.ObjectReference{},
			Template: `apiVersion: devops.kubesphere.io/v1alpha1
kind: ReleaserController
spec:
  image: "ghcr.io/kubesphere-sigs/ks-releaser"
  version: {{.Spec.Version}}
  webhook: false`,
		},
	}
	addon := &v1alpha1.Addon{
		ObjectMeta: metav1.ObjectMeta{Name: "ks-releaser", Namespace: "default"},
		Spec: v1alpha1.AddonSpec{
			Version: "v0.0.1",
			Strategy: v1.LocalObjectReference{
				Name: "simple-operator",
			},
		},
	}

	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx   context.Context
		addon *v1alpha1.Addon
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
		verify  func(t *testing.T, Client client.Client)
	}{{
		name: "normal case",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, strategy.DeepCopy(), addon.DeepCopy()),
		},
		args: args{
			ctx:   context.TODO(),
			addon: addon.DeepCopy(),
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return err == nil
		},
		verify: func(t *testing.T, c client.Client) {
			var err error
			addon := &v1alpha1.Addon{}
			err = c.Get(context.Background(), types.NamespacedName{Name: "ks-releaser", Namespace: "default"}, addon)
			assert.Nil(t, err)
			assert.ElementsMatch(t, []string{v1alpha1.AddonFinalizerName}, addon.Finalizers)

			obj := &unstructured.Unstructured{}
			obj.SetKind("ReleaserController")
			obj.SetAPIVersion("devops.kubesphere.io/v1alpha1")

			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "ks-releaser",
			}, obj)
			assert.Nil(t, err)
			assert.Equal(t, "ks-releaser", obj.GetName())

			var image string
			var ok bool
			image, ok, err = unstructured.NestedString(obj.Object, "spec", "image")
			assert.Nil(t, err)
			assert.True(t, ok)
			assert.Equal(t, "ghcr.io/kubesphere-sigs/ks-releaser", image)
		},
	}, {
		name: "update existing addon",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, strategy.DeepCopy(), &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "devops.kubesphere.io/v1alpha1",
				"kind":       "ReleaserController",
				"metadata": map[string]interface{}{
					"name":      "ks-releaser",
					"namespace": "default",
				},
			}}),
		},
		args: args{
			ctx: context.TODO(),
			addon: &v1alpha1.Addon{
				ObjectMeta: metav1.ObjectMeta{Name: "ks-releaser"},
				Spec: v1alpha1.AddonSpec{
					Version: "v0.0.1",
					Strategy: v1.LocalObjectReference{
						Name: "simple-operator",
					},
				}},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return err == nil
		},
		verify: func(t *testing.T, c client.Client) {
			obj := &unstructured.Unstructured{}
			obj.SetKind("ReleaserController")
			obj.SetAPIVersion("devops.kubesphere.io/v1alpha1")

			err := c.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "ks-releaser",
			}, obj)
			assert.Nil(t, err)
			assert.Equal(t, "ks-releaser", obj.GetName())

			var image string
			var ok bool
			image, ok, err = unstructured.NestedString(obj.Object, "spec", "image")
			assert.Nil(t, err)
			assert.True(t, ok)
			assert.Equal(t, "ghcr.io/kubesphere-sigs/ks-releaser", image)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client:   tt.fields.Client,
				log:      logr.Discard(),
				recorder: record.NewFakeRecorder(100),
			}
			tt.wantErr(t, r.addonHandle(tt.args.ctx, tt.args.addon), fmt.Sprintf("addonHandle(%v, %v)", tt.args.ctx, tt.args.addon))
			if tt.verify != nil {
				tt.verify(t, r.Client)
			}
		})
	}
}

func Test_beingDeleting(t *testing.T) {
	nowTime := metav1.Now()

	type args struct {
		addon *v1alpha1.Addon
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "empty struct",
		args: args{
			addon: &v1alpha1.Addon{},
		},
		want: false,
	}, {
		name: "has deletionTimestamp",
		args: args{
			addon: &v1alpha1.Addon{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &nowTime,
				},
			},
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, beingDeleting(tt.args.addon), "beingDeleting(%v)", tt.args.addon)
		})
	}
}
