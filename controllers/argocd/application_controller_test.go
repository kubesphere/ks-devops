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
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		},
	}, {
		name: "with some specific fields, with default values",
		args: args{
			app: &v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"argocd-image-updater.argoproj.io/image-list": "nginx",
						"other": "other",
					},
					Labels: map[string]string{
						"argocd.argoproj.io/instance": "instance",
						"other":                       "other",
					},
				},
				Spec: v1alpha1.ApplicationSpec{
					ArgoApp: &v1alpha1.ArgoApplication{
						Spec: v1alpha1.ArgoApplicationSpec{
							Destination: v1alpha1.ApplicationDestination{
								Server:    "server",
								Namespace: "namespace",
							},
						},
						Operation: &v1alpha1.Operation{
							Sync: &v1alpha1.SyncOperation{
								Revision:    "master",
								Prune:       true,
								DryRun:      true,
								SyncOptions: []string{"a=b"},
							},
							InitiatedBy: v1alpha1.OperationInitiator{
								Username: "admin",
							},
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
			revision, _, _ := unstructured.NestedString(gotResult.Object, "operation", "sync", "revision")
			assert.Equal(t, "master", revision)
			syncOptions, _, _ := unstructured.NestedStringSlice(gotResult.Object, "operation", "sync", "syncOptions")
			assert.Equal(t, []string{"a=b"}, syncOptions)
			prune, _, _ := unstructured.NestedBool(gotResult.Object, "operation", "sync", "prune")
			assert.Equal(t, true, prune)
			dryRun, _, _ := unstructured.NestedBool(gotResult.Object, "operation", "sync", "dryRun")
			assert.Equal(t, true, dryRun)
			initiatedBy, _, _ := unstructured.NestedString(gotResult.Object, "operation", "initiatedBy", "username")
			assert.Equal(t, "admin", initiatedBy)

			// check annotations
			annotations, _, _ := unstructured.NestedStringMap(gotResult.Object, "metadata", "annotations")
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx",
			}, annotations)
			// check labels
			labels, _, _ := unstructured.NestedStringMap(gotResult.Object, "metadata", "labels")
			assert.Equal(t, map[string]string{
				"argocd.argoproj.io/instance":                        "instance",
				"gitops.kubesphere.io/application-name":              "",
				"gitops.kubesphere.io/application-namespace":         "",
				"gitops.kubesphere.io/argocd-application-control-by": "ks-devops",
			}, labels)
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
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	app := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake",
			Namespace: "fake",
			Labels: map[string]string{
				v1alpha1.ArgoCDLocationLabelKey: "fake",
			},
		},
		Spec: v1alpha1.ApplicationSpec{
			ArgoApp: &v1alpha1.ArgoApplication{
				Spec: v1alpha1.ArgoApplicationSpec{
					Project: "project",
					SyncPolicy: &v1alpha1.SyncPolicy{
						Automated: &v1alpha1.SyncPolicyAutomated{
							Prune: true,
						},
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
		Client client.Client
		log    logr.Logger
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
			Client: fake.NewFakeClientWithScheme(schema, app.DeepCopy()),
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
			assert.Equal(t, "fake", project)
			prune, _, _ := unstructured.NestedBool(app.Object, "spec", "syncPolicy", "automated", "prune")
			assert.True(t, prune)
		},
	}, {
		name: "with Argo Application",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, app.DeepCopy()),
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
			assert.Equal(t, "fake", project)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: &record.FakeRecorder{},
			}
			err := r.reconcileArgoApplication(tt.args.app)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func Test_finalizersChangedPredicate_Update(t *testing.T) {
	type fields struct {
		Funcs predicate.Funcs
	}
	type args struct {
		e event.UpdateEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{{
		name: "different finalizers",
		args: args{
			e: event.UpdateEvent{
				MetaOld: &metav1.ObjectMeta{
					Finalizers: []string{"b"},
				},
				ObjectOld: &v1alpha1.Application{},
				MetaNew: &metav1.ObjectMeta{
					Finalizers: []string{"a"},
				},
				ObjectNew: &v1alpha1.Application{},
			},
		},
		want: true,
	}, {
		name: "different order with the finalizers",
		args: args{
			e: event.UpdateEvent{
				MetaOld: &metav1.ObjectMeta{
					Finalizers: []string{"b", "a"},
				},
				ObjectOld: &v1alpha1.Application{},
				MetaNew: &metav1.ObjectMeta{
					Finalizers: []string{"a", "b"},
				},
				ObjectNew: &v1alpha1.Application{},
			},
		},
		want: true,
	}, {
		name: "ObjectOld is nil",
		args: args{
			e: event.UpdateEvent{
				MetaOld:   &metav1.ObjectMeta{},
				MetaNew:   &metav1.ObjectMeta{},
				ObjectNew: &v1alpha1.Application{},
			},
		},
		want: false,
	}, {
		name: "MetaNew is nil",
		args: args{
			e: event.UpdateEvent{
				ObjectOld: &v1alpha1.Application{},
				MetaOld:   &metav1.ObjectMeta{},
				ObjectNew: &v1alpha1.Application{},
			},
		},
		want: false,
	}, {
		name: "ObjectNew is nil",
		args: args{
			e: event.UpdateEvent{
				MetaOld:   &metav1.ObjectMeta{},
				ObjectOld: &v1alpha1.Application{},
				MetaNew:   &metav1.ObjectMeta{},
			},
		},
		want: false,
	}, {
		name: "metaOld is nil",
		args: args{
			e: event.UpdateEvent{
				ObjectOld: &v1alpha1.Application{},
				MetaNew:   &metav1.ObjectMeta{},
				ObjectNew: &v1alpha1.Application{},
			},
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi := finalizersChangedPredicate{
				Funcs: tt.fields.Funcs,
			}
			assert.Equalf(t, tt.want, fi.Update(tt.args.e), "Update(%v)", tt.args.e)
		})
	}
}

func Test_copyArgoAnnotationsAndLabels(t *testing.T) {
	type args struct {
		app     *v1alpha1.Application
		argoApp *unstructured.Unstructured
	}
	tests := []struct {
		name   string
		args   args
		verify func(*testing.T, *unstructured.Unstructured)
	}{{
		name: "normal",
		args: args{
			app: &v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"other":           "other",
						"a.argoproj.io/b": "ab",
					},
					Labels: map[string]string{
						"other":           "other",
						"a.argoproj.io/b": "ab",
					},
				},
			},
			argoApp: &unstructured.Unstructured{},
		},
		verify: func(t *testing.T, u *unstructured.Unstructured) {
			annotations, _, _ := unstructured.NestedStringMap(u.Object, "metadata", "annotations")
			assert.Equal(t, map[string]string{
				"a.argoproj.io/b": "ab",
			}, annotations)
			labels, _, _ := unstructured.NestedStringMap(u.Object, "metadata", "annotations")
			assert.Equal(t, map[string]string{
				"a.argoproj.io/b": "ab",
			}, labels)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyArgoAnnotationsAndLabels(tt.args.app, tt.args.argoApp)
		})
	}
}
