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
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kubesphere/ks-devops/controllers/core"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_createArgoCluster(t *testing.T) {
	defaultCluster := &unstructured.Unstructured{}
	defaultCluster.SetKind("Cluster")
	defaultCluster.SetAPIVersion("cluster.kubesphere.io/v1alpha1")
	defaultCluster.SetNamespace("ns")
	defaultCluster.SetName("name")
	_ = unstructured.SetNestedField(defaultCluster.Object, base64.StdEncoding.EncodeToString([]byte(`clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: server
  name: name
users:
- name: name
  user:
    token: token`)), "spec", "connection", "kubeconfig")
	_ = unstructured.SetNestedField(defaultCluster.Object, "server", "spec", "connection", "kubernetesAPIEndpoint")

	defaultArgoCluster := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "name",
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
			},
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"name":   []byte("name"),
			"config": []byte(`{"bearerToken":"token","tlsClientConfig":{"insecure":true}}`),
			"server": []byte("server"),
		},
	}

	type args struct {
		cluster *unstructured.Unstructured
	}
	tests := []struct {
		name       string
		args       args
		wantSecret *v1.Secret
	}{{
		name: "normal",
		args: args{
			cluster: defaultCluster.DeepCopy(),
		},
		wantSecret: defaultArgoCluster.DeepCopy(),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantSecret, createArgoCluster(tt.args.cluster), "createArgoCluster(%v)", tt.args.cluster)
		})
	}
}

func Test_getCluster(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	cluster := &unstructured.Unstructured{}
	cluster.SetKind("Cluster")
	cluster.SetResourceVersion("999")
	cluster.SetAPIVersion("cluster.kubesphere.io/v1alpha1")
	cluster.SetName("name")
	cluster.SetNamespace("namespace")

	defaultNamespacedName := types.NamespacedName{
		Namespace: "namespace",
		Name:      "name",
	}

	type args struct {
		client         client.Reader
		namespacedName types.NamespacedName
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, obj *unstructured.Unstructured, err error)
	}{{
		name: "normal case",
		args: args{
			client:         fake.NewClientBuilder().WithScheme(schema).WithObjects(cluster.DeepCopy()).Build(),
			namespacedName: defaultNamespacedName,
		},
		verify: func(t *testing.T, obj *unstructured.Unstructured, err error) {
			assert.Nil(t, err)
			assert.NotNil(t, obj)
			assert.Equal(t, cluster, obj)
		},
	}, {
		name: "not found",
		args: args{
			client:         fake.NewClientBuilder().WithScheme(schema).Build(),
			namespacedName: defaultNamespacedName,
		},
		verify: func(t *testing.T, obj *unstructured.Unstructured, err error) {
			assert.NotNil(t, err)
			assert.Nil(t, obj)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := getCluster(tt.args.client, tt.args.namespacedName)
			tt.verify(t, gotResult, err)
		})
	}
}

func TestMultiClusterReconciler_updateOrCreate(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	defaultArgoCluster := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
		},
		StringData: map[string]string{
			"key": "value",
		},
	}

	defaultOwnerRef := metav1.OwnerReference{}

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		cluster *v1.Secret
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(t *testing.T, c client.Client, err error)
	}{{
		name: "create when the cluster does not exist",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).Build(),
		},
		args: args{
			cluster: defaultArgoCluster.DeepCopy(),
		},
		verify: func(t *testing.T, c client.Client, err error) {
			assert.Nil(t, err)

			secret := &v1.Secret{}
			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "name",
			}, secret)
			assert.Nil(t, err)
			assert.Equal(t, "value", secret.StringData["key"])
			assert.True(t, len(secret.GetOwnerReferences()) == 1)
			assert.Equal(t, defaultOwnerRef, secret.GetOwnerReferences()[0])
		},
	}, {
		name: "update when the corresponding cluster does not exist",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(defaultArgoCluster.DeepCopy()).Build(),
		},
		args: args{
			cluster: defaultArgoCluster.DeepCopy(),
		},
		verify: func(t *testing.T, c client.Client, err error) {
			assert.Nil(t, err)

			secret := &v1.Secret{}
			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "name",
			}, secret)
			assert.Nil(t, err)
			assert.Equal(t, "value", secret.StringData["key"])
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MultiClusterReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			err := r.updateOrCreate(tt.args.cluster, defaultOwnerRef)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func Test_getOwnerReference(t *testing.T) {
	defaultObject := &unstructured.Unstructured{}
	defaultObject.SetName("name")
	defaultObject.SetKind("kind")
	defaultObject.SetUID("uid")
	defaultObject.SetAPIVersion("apiversion")

	type args struct {
		object *unstructured.Unstructured
	}
	tests := []struct {
		name    string
		args    args
		wantRef metav1.OwnerReference
	}{{
		name: "normal",
		args: args{
			object: defaultObject.DeepCopy(),
		},
		wantRef: metav1.OwnerReference{
			APIVersion: "apiversion",
			Kind:       "kind",
			Name:       "name",
			UID:        "uid",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRef, getOwnerReference(tt.args.object), "getOwnerReference(%v)", tt.args.object)
		})
	}
}

func TestMultiClusterReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	defaultRequest := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "ns",
			Name:      "name",
		},
	}

	defaultCluster := &unstructured.Unstructured{}
	defaultCluster.SetKind("Cluster")
	defaultCluster.SetAPIVersion("cluster.kubesphere.io/v1alpha1")
	defaultCluster.SetName("name")
	defaultCluster.SetNamespace("ns")

	ignoreCluster := defaultCluster.DeepCopy()
	ignoreCluster.SetLabels(map[string]string{
		"cluster-role.kubesphere.io/host": "",
	})

	argoSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "argocd-secret",
		},
	}

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
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
		verify     func(t *testing.T, c client.Client)
	}{{
		name: "not found",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).Build(),
		},
		args: args{
			req: defaultRequest,
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return true
		},
		wantResult: controllerruntime.Result{},
	}, {
		name: "ignore changes case",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(ignoreCluster.DeepCopy(), argoSecret.DeepCopy()).Build(),
		},
		args: args{
			req: defaultRequest,
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
		wantResult: controllerruntime.Result{},
	}, {
		name: "no argo cluster existing",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(defaultCluster.DeepCopy(), argoSecret.DeepCopy()).Build(),
		},
		args: args{
			req: defaultRequest,
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
		wantResult: controllerruntime.Result{},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MultiClusterReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
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

func TestMultiClusterReconciler_findArgoCDNamespace(t *testing.T) {
	const expectedNamespace = "namespace"
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	argoSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: expectedNamespace,
			Name:      "argocd-secret",
		},
	}

	type fields struct {
		Client client.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "no expected secret",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).Build(),
		},
		want: "",
	}, {
		name: "normal case, have the expected secret",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(argoSecret.DeepCopy()).Build(),
		},
		want: expectedNamespace,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MultiClusterReconciler{
				Client: tt.fields.Client,
			}
			assert.Equalf(t, tt.want, r.findArgoCDNamespace(), "findArgoCDNamespace()")
		})
	}
}

func Test_getArgoClusterConfigFormat(t *testing.T) {
	configData := base64.StdEncoding.EncodeToString([]byte(`clusters:
- cluster:
    insecure-skip-tls-verify: true
    certificate-authority-data: LeRuIGZha2UK
    server: server
  name: name
users:
- name: name
  user:
    client-certificate-data: LeRuIGZha2UK
    client-key-data: LeRuIGZha2UK
    token: token`))
	type args struct {
		config string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{{
		name: "normal",
		args: args{
			config: configData,
		},
		want: []byte(`{"bearerToken":"token","tlsClientConfig":{"insecure":true,"certData":"LeRuIGZha2UK","keyData":"LeRuIGZha2UK","caData":"LeRuIGZha2UK"}}`),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, string(tt.want), string(getArgoClusterConfigData(tt.args.config)), "getArgoClusterConfigFormat(%v)", tt.args.config)
		})
	}
}

func TestMultiClusterReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type fields struct {
		Client          client.Client
		log             logr.Logger
		recorder        record.EventRecorder
		argocdNamespace string
	}
	type args struct {
		mgr controllerruntime.Manager
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
			r := &MultiClusterReconciler{
				Client:          tt.fields.Client,
				log:             tt.fields.log,
				recorder:        tt.fields.recorder,
				argocdNamespace: tt.fields.argocdNamespace,
			}
			tt.wantErr(t, r.SetupWithManager(tt.args.mgr), fmt.Sprintf("SetupWithManager(%v)", tt.args.mgr))
		})
	}
}
