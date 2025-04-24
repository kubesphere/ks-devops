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
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kubesphere/ks-devops/controllers/core"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestMultiClusterReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "fake-cluster",
		},
	}

	memberCluster := createBareClusterObject()
	memberCluster.SetName("fake-cluster")

	hostCluster := memberCluster.DeepCopy()
	hostCluster.SetLabels(map[string]string{
		"cluster-role.kubesphere.io/host": "",
	})

	devopsNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-devops-project-ns",
			Labels: map[string]string{
				"github.com/kubesphere/ks-devopsproject": "fake-devops-project-ns",
			},
		},
	}

	type fields struct {
		Client client.Client
		log    logr.Logger
	}
	type args struct {
		req ctrl.Request
	}

	tests := []struct {
		name string
		fields
		args
		verify func(t *testing.T, c client.Client, err error)
	}{
		{
			name: "not found cluster",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(nil).Build(),
				log:    logr.New(log.NullLogSink{}),
			},
			args: args{
				req: req,
			},
			verify: func(t *testing.T, c client.Client, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "found a hostCluster",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(nil).WithObjects(hostCluster.DeepCopy()).Build(),
				log:    logr.New(log.NullLogSink{}),
			},
			args: args{
				req: req,
			},
			verify: func(t *testing.T, c client.Client, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "found a memberCluster",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(memberCluster.DeepCopy(), devopsNs.DeepCopy()).Build(),
				log:    logr.New(log.NullLogSink{}),
			},
			args: args{
				req: req,
			},
			verify: func(t *testing.T, c client.Client, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MultiClusterReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: &record.FakeRecorder{},
			}
			_, err := r.Reconcile(context.Background(), tt.args.req)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func TestMultiClusterReconciler_reconcileCluster(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	fakeKubeConfig := "fake-kubeconfig-data"
	base64Str := base64.StdEncoding.EncodeToString([]byte(fakeKubeConfig))
	memberCluster := createBareClusterObject()
	_ = unstructured.SetNestedField(memberCluster.Object, base64Str, "spec", "connection", "kubeconfig")

	memberCluster.SetName("fake-cluster")

	newFakeKubeConfig := "new-fake-kubeconfig-data"
	NewBase64Str := base64.StdEncoding.EncodeToString([]byte(newFakeKubeConfig))
	newMemberCluster := memberCluster.DeepCopy()
	_ = unstructured.SetNestedField(newMemberCluster.Object, NewBase64Str, "spec", "connection", "kubeconfig")

	oldSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-cluster",
			Namespace: "fake-devops-project-ns",
		},
		Type: "Opaque",
		StringData: map[string]string{
			DefaultKubeConfigKey: base64.StdEncoding.EncodeToString([]byte("fake-kubeconfig-data")),
		},
	}

	hostCluster := memberCluster.DeepCopy()
	hostCluster.SetLabels(map[string]string{
		"cluster-role.kubesphere.io/host": "",
	})

	type fields struct {
		Client client.Client
		log    logr.Logger
	}
	type args struct {
		cluster *unstructured.Unstructured
	}

	devopsProjectNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-devops-project-ns",
			Labels: map[string]string{
				"github.com/kubesphere/ks-devopsproject": "fake-devops-project-ns",
			},
		},
	}

	anotherDevopsProjectNs := devopsProjectNs.DeepCopy()
	anotherDevopsProjectNs.SetName("another-fake-devops-project-ns")

	tests := []struct {
		name string
		fields
		args
		verify func(t *testing.T, c client.Client, err error)
	}{
		{
			name: "there is no devops project namespace",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).Build(),
				log:    logr.New(log.NullLogSink{}),
			},
			args: args{
				cluster: memberCluster.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "create kubeconfig secret in devops project namespaces",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(devopsProjectNs.DeepCopy(), anotherDevopsProjectNs.DeepCopy()).Build(),
				log:    logr.New(log.NullLogSink{}),
			},
			args: args{
				cluster: memberCluster.DeepCopy(),
			},
			verify: func(t *testing.T, c client.Client, err error) {
				assert.Nil(t, err)
				secret := &v1.Secret{}
				err = c.Get(context.TODO(), types.NamespacedName{
					Namespace: "fake-devops-project-ns",
					Name:      "fake-cluster",
				}, secret)
				assert.Nil(t, err)
				assert.Equal(t, fakeKubeConfig, secret.StringData[DefaultKubeConfigKey])

				anotherSecret := &v1.Secret{}
				err = c.Get(context.TODO(), types.NamespacedName{
					Namespace: "another-fake-devops-project-ns",
					Name:      "fake-cluster",
				}, anotherSecret)
				assert.Nil(t, err)
				assert.Equal(t, fakeKubeConfig, anotherSecret.StringData[DefaultKubeConfigKey])
			},
		},
		{
			name: "update a kubeconfig secret",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(memberCluster.DeepCopy(), devopsProjectNs.DeepCopy(), anotherDevopsProjectNs.DeepCopy(), oldSecret.DeepCopy()).Build(),
				log:    logr.New(log.NullLogSink{}),
			},
			args: args{
				cluster: newMemberCluster.DeepCopy(),
			},
			verify: func(t *testing.T, c client.Client, err error) {
				assert.Nil(t, err)
				secret := &v1.Secret{}
				err = c.Get(context.TODO(), types.NamespacedName{
					Namespace: "fake-devops-project-ns",
					Name:      "fake-cluster",
				}, secret)
				assert.Nil(t, err)
				assert.Equal(t, newFakeKubeConfig, secret.StringData[DefaultKubeConfigKey])

				anotherSecret := &v1.Secret{}
				err = c.Get(context.TODO(), types.NamespacedName{
					Namespace: "another-fake-devops-project-ns",
					Name:      "fake-cluster",
				}, anotherSecret)
				assert.Nil(t, err)
				assert.Equal(t, newFakeKubeConfig, anotherSecret.StringData[DefaultKubeConfigKey])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MultiClusterReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: &record.FakeRecorder{},
			}

			err := r.reconcileCluster(context.Background(), tt.args.cluster)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func TestMultiClusterReconciler_GetName(t *testing.T) {
	r := &MultiClusterReconciler{}
	assert.Equal(t, "FluxCDMultiClusterController", r.GetName())
}

func TestMultiClusterReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	type fields struct {
		Client client.Client
		log    logr.Logger
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
				Scheme: schema,
				Client: fake.NewClientBuilder().WithScheme(schema).Build(),
			},
		},
		wantErr: core.NoErrors,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MultiClusterReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: &record.FakeRecorder{},
			}
			tt.wantErr(t, r.SetupWithManager(tt.args.mgr), fmt.Sprintf("SetupWithManager(%v)", tt.args.mgr))
		})
	}
}
