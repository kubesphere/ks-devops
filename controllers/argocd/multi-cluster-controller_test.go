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
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_ignore(t *testing.T) {
	sampleCluster := &unstructured.Unstructured{}
	sampleCluster.SetKind("Cluster")
	sampleCluster.SetAPIVersion("cluster.kubesphere.io/v1alpha1")

	type args struct {
		cluster func() *unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "cluster is nil",
		want: true,
	}, {
		name: "cluster without the particular label",
		args: args{cluster: func() *unstructured.Unstructured {
			return sampleCluster.DeepCopy()
		}},
		want: false,
	}, {
		name: "with the particular label",
		args: args{cluster: func() *unstructured.Unstructured {
			cluster := sampleCluster.DeepCopy()
			cluster.SetLabels(map[string]string{
				"cluster-role.kubesphere.io/host": "",
			})
			return cluster
		}},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cluster *unstructured.Unstructured
			if tt.args.cluster != nil {
				cluster = tt.args.cluster()
			}
			assert.Equalf(t, tt.want, ignore(cluster), "ignore(%v)", tt.args.cluster)
		})
	}
}

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
			client:         fake.NewFakeClientWithScheme(schema, cluster.DeepCopy()),
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
			client:         fake.NewFakeClientWithScheme(schema),
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
			Client: fake.NewFakeClientWithScheme(schema),
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
			Client: fake.NewFakeClientWithScheme(schema, defaultArgoCluster.DeepCopy()),
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
			Client: fake.NewFakeClientWithScheme(schema),
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
			Client: fake.NewFakeClientWithScheme(schema, ignoreCluster.DeepCopy(), argoSecret.DeepCopy()),
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
			Client: fake.NewFakeClientWithScheme(schema, defaultCluster.DeepCopy(), argoSecret.DeepCopy()),
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
			gotResult, err := r.Reconcile(tt.args.req)
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
			Client: fake.NewFakeClientWithScheme(schema),
		},
		want: "",
	}, {
		name: "normal case, have the expected secret",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, argoSecret.DeepCopy()),
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
    server: server
  name: name
users:
- name: name
  user:
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
		want: []byte(`{"bearerToken":"token","tlsClientConfig":{"insecure":true}}`),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, string(tt.want), string(getArgoClusterConfigData(tt.args.config)), "getArgoClusterConfigFormat(%v)", tt.args.config)
		})
	}
}
