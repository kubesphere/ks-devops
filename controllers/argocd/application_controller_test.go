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
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_createUnstructuredApplication(t *testing.T) {
	type args struct {
		app *v1alpha1.Application
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, gotResult *unstructured.Unstructured, gotErr error)
	}{{
		name: "without argo",
		args: args{
			app: &v1alpha1.Application{},
		},
		verify: func(t *testing.T, gotResult *unstructured.Unstructured, gotErr error) {
			assert.NotNil(t, gotErr)
		},
	}, {
		name: "empty Argo Application with the default value",
		args: args{
			app: &v1alpha1.Application{
				Spec: v1alpha1.ApplicationSpec{
					ArgoApp: &v1alpha1.ArgoApplication{},
				},
			},
		},
		verify: func(t *testing.T, gotResult *unstructured.Unstructured, gotErr error) {
			assert.Nil(t, gotErr)

			// make sure it has the ownerReference
			refs, _, _ := unstructured.NestedSlice(gotResult.Object, "metadata", "ownerReferences")
			assert.NotNil(t, refs)
			assert.Equal(t, 1, len(refs))
		},
	}, {
		name: "with some specific fields, with default values",
		args: args{
			app: &v1alpha1.Application{
				Spec: v1alpha1.ApplicationSpec{
					ArgoApp: &v1alpha1.ArgoApplication{
						Destination: v1alpha1.ApplicationDestination{
							Server:    "server",
							Namespace: "namespace",
						},
					},
				},
			},
		},
		verify: func(t *testing.T, gotResult *unstructured.Unstructured, gotErr error) {
			assert.Nil(t, gotErr)

			project, _, _ := unstructured.NestedString(gotResult.Object, "spec", "project")
			assert.Equal(t, "default", project)
			destServer, _, _ := unstructured.NestedString(gotResult.Object, "spec", "destination", "server")
			assert.Equal(t, "server", destServer)
			destNs, _, _ := unstructured.NestedString(gotResult.Object, "spec", "destination", "namespace")
			assert.Equal(t, "namespace", destNs)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := createUnstructuredApplication(tt.args.app)
			tt.verify(t, gotResult, err)
		})
	}
}

func TestApplicationReconciler_reconcileArgoApplication(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	app := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake",
			Namespace: "fake",
		},
		Spec: v1alpha1.ApplicationSpec{
			ArgoApp: &v1alpha1.ArgoApplication{
				Project: "project",
				SyncPolicy: &v1alpha1.SyncPolicy{
					Automated: &v1alpha1.SyncPolicyAutomated{
						Prune: true,
					},
				},
			},
		},
	}

	argoApp := &unstructured.Unstructured{}
	argoApp.SetKind("Application")
	argoApp.SetAPIVersion("argoproj.io/v1alpha1")
	argoApp.SetName("fake")
	argoApp.SetNamespace("fake")

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		app *v1alpha1.Application
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(t *testing.T, Client client.Client, err error)
	}{{
		name: "without Argo Application",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema),
		},
		args: args{
			app: app.DeepCopy(),
		},
		verify: func(t *testing.T, Client client.Client, err error) {
			assert.Nil(t, err)

			app := argoApp.DeepCopy()

			err = Client.Get(context.TODO(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake",
			}, app)
			assert.Nil(t, err)
			project, _, _ := unstructured.NestedString(app.Object, "spec", "project")
			assert.Equal(t, "project", project)
			prune, _, _ := unstructured.NestedBool(app.Object, "spec", "syncPolicy", "automated", "prune")
			assert.True(t, prune)
		},
	}, {
		name: "with Argo Application",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, argoApp.DeepCopy()),
		},
		args: args{
			app: app.DeepCopy(),
		},
		verify: func(t *testing.T, Client client.Client, err error) {
			assert.Nil(t, err)

			app := argoApp.DeepCopy()

			err = Client.Get(context.TODO(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake",
			}, app)
			assert.Nil(t, err)
			project, _, _ := unstructured.NestedString(app.Object, "spec", "project")
			assert.Equal(t, "project", project)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			err := r.reconcileArgoApplication(tt.args.app)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}
