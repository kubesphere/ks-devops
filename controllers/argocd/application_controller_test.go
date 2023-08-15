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
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/controllers/core"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
			Annotations: map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx",
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
	appWithLabel := app.DeepCopy()
	appWithLabel.SetLabels(map[string]string{
		v1alpha1.ArgoCDAppNameLabelKey:  "fake",
		v1alpha1.ArgoCDLocationLabelKey: "fake",
	})

	argoApp := &unstructured.Unstructured{}
	argoApp.SetKind("Application")
	argoApp.SetAPIVersion("argoproj.io/v1alpha1")
	argoApp.SetName("fake")
	argoApp.SetNamespace("fake")

	argoAppWithLabel := argoApp.DeepCopy()
	argoAppWithLabel.SetLabels(map[string]string{
		v1alpha1.ArgoCDAppNameLabelKey: "fake",
	})

	argoAppList := &unstructured.UnstructuredList{}
	argoAppList.SetKind("ApplicationList")
	argoAppList.SetAPIVersion("argoproj.io/v1alpha1")
	argoAppList.Items = append(argoAppList.Items, *argoApp)

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
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(app.DeepCopy()).Build(),
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
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(app.DeepCopy()).Build(),
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
			annotations, _, _ := unstructured.NestedStringMap(app.Object, "metadata", "annotations")
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx",
			}, annotations)
		},
	}, {
		name: "update the existing argo application which not have labels",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(app.DeepCopy()).WithLists(argoAppList.DeepCopy()).Build(),
		},
		args: args{
			app: app.DeepCopy(),
		},
		verify: func(t *testing.T, Client client.Client, err error) {
			assert.Nil(t, err)
		},
	}, {
		name: "update the existing argo application which has labels",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(appWithLabel.DeepCopy()).WithObjects(argoAppWithLabel.DeepCopy()).Build(),
		},
		args: args{
			app: appWithLabel.DeepCopy(),
		},
		verify: func(t *testing.T, Client client.Client, err error) {
			assert.Nil(t, err)

			app := argoApp.DeepCopy()

			err = Client.Get(context.TODO(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake",
			}, app)
			assert.Nil(t, err)
			annotations, _, _ := unstructured.NestedStringMap(app.Object, "metadata", "annotations")
			assert.Equal(t, map[string]string{
				"argocd-image-updater.argoproj.io/image-list": "nginx",
			}, annotations)
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
				ObjectOld: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"b"},
					},
				},
				ObjectNew: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"a"},
					},
				},
			},
		},
		want: true,
	}, {
		name: "different order with the finalizers",
		args: args{
			e: event.UpdateEvent{
				ObjectOld: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"b", "a"},
					},
				},
				ObjectNew: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"a", "b"},
					},
				},
			},
		},
		want: true,
	}, {
		name: "ObjectOld is nil",
		args: args{
			e: event.UpdateEvent{
				ObjectOld: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{},
				},
				ObjectNew: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{},
				},
			},
		},
		want: false,
	}, {
		name: "MetaNew is nil",
		args: args{
			e: event.UpdateEvent{
				ObjectOld: &v1alpha1.Application{},
				ObjectNew: &v1alpha1.Application{},
			},
		},
		want: false,
	}, {
		name: "ObjectNew is nil",
		args: args{
			e: event.UpdateEvent{
				ObjectOld: &v1alpha1.Application{},
				ObjectNew: &v1alpha1.Application{},
			},
		},
		want: false,
	}, {
		name: "metaOld is nil",
		args: args{
			e: event.UpdateEvent{
				ObjectOld: &v1alpha1.Application{},
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

func Test_mapKeysContains(t *testing.T) {
	type args struct {
		filter       string
		annotations  map[string]string
		annotations2 map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantHas bool
	}{{
		name: "do not contain the key string in one map",
		args: args{
			filter: "fake",
			annotations: map[string]string{
				"name": "linuxsuren",
			},
		},
		wantHas: false,
	}, {
		name: "do not contain the key string in two maps",
		args: args{
			filter: "fake",
			annotations: map[string]string{
				"name": "linuxsuren",
			},
			annotations2: map[string]string{
				"name": "linuxsuren-bot",
			},
		},
		wantHas: false,
	}, {
		name: "contain the key string in one map",
		args: args{
			filter: "fake",
			annotations: map[string]string{
				"name.fake.com": "linuxsuren",
				"age":           "10",
			},
		},
		wantHas: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantHas, mapKeysContains(tt.args.filter, tt.args.annotations, tt.args.annotations2), "mapKeysContains(%v, %v, %v)", tt.args.filter, tt.args.annotations, tt.args.annotations2)
		})
	}
}

func Test_specificAnnotationsOrLabelsChangedPredicate_Update(t *testing.T) {
	type fields struct {
		Funcs  predicate.Funcs
		filter string
	}

	defaultPredicate := fields{filter: "fake"}
	type args struct {
		e event.UpdateEvent
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantChanged bool
	}{{
		name:   "everything is same",
		fields: defaultPredicate,
		args: args{
			e: event.UpdateEvent{
				ObjectNew: &v1alpha1.Application{},
				ObjectOld: &v1alpha1.Application{},
			},
		},
		wantChanged: false,
	}, {
		name:   "annotations are different, but do not have key",
		fields: defaultPredicate,
		args: args{
			e: event.UpdateEvent{
				ObjectNew: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"name": "fake",
						},
					},
				},
				ObjectOld: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"name": "linuxsuren",
						},
					},
				},
			},
		},
		wantChanged: false,
	}, {
		name:   "labels are different, but do not have key",
		fields: defaultPredicate,
		args: args{
			e: event.UpdateEvent{
				ObjectNew: &v1alpha1.Application{

					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "fake",
						},
					},
				},
				ObjectOld: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "linuxsuren",
						},
					}},
			},
		},
		wantChanged: false,
	}, {
		name:   "old meta's annotations has key",
		fields: defaultPredicate,
		args: args{
			e: event.UpdateEvent{
				ObjectNew: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{},
				},
				ObjectOld: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"this.is.fake": "",
						},
					},
				},
			},
		},
		wantChanged: true,
	}, {
		name:   "new meta's annotations has key",
		fields: defaultPredicate,
		args: args{
			e: event.UpdateEvent{
				ObjectNew: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"this.is.fake": "",
						},
					},
				},
				ObjectOld: &v1alpha1.Application{},
			},
		},
		wantChanged: true,
	}, {
		name:   "old meta's labels has key",
		fields: defaultPredicate,
		args: args{
			e: event.UpdateEvent{
				ObjectNew: &v1alpha1.Application{},
				ObjectOld: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"this.is.fake": "",
						},
					},
				},
			},
		},
		wantChanged: true,
	}, {
		name:   "new meta's labels has key",
		fields: defaultPredicate,
		args: args{
			e: event.UpdateEvent{
				ObjectNew: &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"this.is.fake": "",
						},
					},
				},
				ObjectOld: &v1alpha1.Application{},
			},
		},
		wantChanged: true,
	}}
	t.Logf("there are [%d] test cases\n", len(tests))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := specificAnnotationsOrLabelsChangedPredicate{
				Funcs:  tt.fields.Funcs,
				filter: tt.fields.filter,
			}
			assert.Equalf(t, tt.wantChanged, p.Update(tt.args.e), "Update(%v)", tt.args.e)
		})
	}
}

func TestApplicationReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)
	err = v1alpha1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type fields struct {
		Client   client.Client
		recorder record.EventRecorder
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
			r := &ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      logr.Logger{},
				recorder: tt.fields.recorder,
			}
			tt.wantErr(t, r.SetupWithManager(tt.args.mgr), fmt.Sprintf("SetupWithManager(%v)", tt.args.mgr))
		})
	}
}

func TestApplicationReconciler_deleteArgoApp(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	ctx := context.Background()
	baseArgoApp := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      "test-app",
				"namespace": "argocd",
			},
		},
	}

	tests := []struct {
		name   string
		verify func(*testing.T)
	}{
		{
			name: "cascade",
			verify: func(t *testing.T) {
				argoApp := baseArgoApp.DeepCopy()
				app := &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"finalizer-1", "resources-finalizer.argocd.argoproj.io", "finalizer-2"},
					},
				}
				expectedFinalizers := []string{"resources-finalizer.argocd.argoproj.io"}
				r := &ApplicationReconciler{
					Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(baseArgoApp.DeepCopy()).Build(),
					log:    logr.Logger{},
				}
				err = r.DeleteArgoApp(app, argoApp)
				assert.NoError(t, err)

				err = r.Get(ctx, types.NamespacedName{
					Namespace: argoApp.GetNamespace(),
					Name:      argoApp.GetName(),
				}, argoApp)

				assert.NoError(t, err)
				assert.NotNil(t, argoApp.GetDeletionTimestamp())
				assert.Equal(t, argoApp.GetFinalizers(), expectedFinalizers)
			},
		},
		{
			name: "cascade with other finalizer",
			verify: func(t *testing.T) {
				argoApp := baseArgoApp.DeepCopy()
				argoApp.SetFinalizers([]string{"test-finalizer.argocd.argoproj.io"})
				app := &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"finalizer-1", "resources-finalizer.argocd.argoproj.io", "finalizer-2"},
					},
				}
				expectedFinalizers := []string{"resources-finalizer.argocd.argoproj.io"}
				r := &ApplicationReconciler{
					Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(baseArgoApp.DeepCopy()).Build(),
					log:    logr.Logger{},
				}
				err = r.DeleteArgoApp(app, argoApp)
				assert.NoError(t, err)

				err = r.Get(ctx, types.NamespacedName{
					Namespace: argoApp.GetNamespace(),
					Name:      argoApp.GetName(),
				}, argoApp)

				assert.NoError(t, err)
				assert.NotNil(t, argoApp.GetDeletionTimestamp())
				assert.Equal(t, argoApp.GetFinalizers(), expectedFinalizers)
			},
		},
		{
			name: "normal",
			verify: func(t *testing.T) {
				argoApp := baseArgoApp.DeepCopy()
				app := &v1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"finalizer-1", "finalizer-2"},
					},
				}
				r := &ApplicationReconciler{
					Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(baseArgoApp.DeepCopy()).Build(),
					log:    logr.Logger{},
				}
				err = r.DeleteArgoApp(app, argoApp)
				assert.NoError(t, err)

				err = r.Get(ctx, types.NamespacedName{
					Namespace: argoApp.GetNamespace(),
					Name:      argoApp.GetName(),
				}, argoApp)
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(t)
		})
	}
}

func TestApplicationReconciler_addArgoAppNameIntoLabels(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.NoError(t, err)
	ctx := context.Background()

	baseApp := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "argocd",
			Name:      "test-app",
		},
	}

	labelApp := baseApp.DeepCopy()
	labelApp.SetLabels(map[string]string{
		"test-label": "test",
	})

	sameLabelApp := baseApp.DeepCopy()
	sameLabelApp.SetLabels(map[string]string{
		v1alpha1.ArgoCDAppNameLabelKey: "test",
	})

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		namespace      string
		name           string
		argoAppName    string
		expectedLabels map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "empty label",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(baseApp.DeepCopy()).Build(),
				log:    logr.Logger{},
			},
			args: args{
				namespace:   "argocd",
				name:        "test-app",
				argoAppName: "test-argo-app",
				expectedLabels: map[string]string{
					v1alpha1.ArgoCDAppNameLabelKey: "test-argo-app",
				},
			},
			wantErr: false,
		},
		{
			name: "with label",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(labelApp.DeepCopy()).Build(),
				log:    logr.Logger{},
			},
			args: args{
				namespace:   "argocd",
				name:        "test-app",
				argoAppName: "test-argo-app",
				expectedLabels: map[string]string{
					v1alpha1.ArgoCDAppNameLabelKey: "test-argo-app",
					"test-label":                   "test",
				},
			},
			wantErr: false,
		},
		{
			name: "same label",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(sameLabelApp.DeepCopy()).Build(),
				log:    logr.Logger{},
			},
			args: args{
				namespace:   "argocd",
				name:        "test-app",
				argoAppName: "test-argo-app",
				expectedLabels: map[string]string{
					v1alpha1.ArgoCDAppNameLabelKey: "test-argo-app",
				},
			},
			wantErr: false,
		},
		{
			name: "not exist app",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(sameLabelApp.DeepCopy()).Build(),
				log:    logr.Logger{},
			},
			args: args{
				namespace: "not-exist",
				name:      "not-exist",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			if err := r.addArgoAppNameIntoLabels(tt.args.namespace, tt.args.name, tt.args.argoAppName); (err != nil) != tt.wantErr {
				t.Errorf("addArgoAppNameIntoLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				app := &v1alpha1.Application{}
				err = r.Client.Get(ctx, types.NamespacedName{
					Namespace: tt.args.namespace,
					Name:      tt.args.name,
				}, app)
				assert.NoError(t, err)
				assert.Equal(t, app.Labels, tt.args.expectedLabels)
			}
		})
	}
}

func TestApplicationReconciler_RemoveAppFinalizer(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.NoError(t, err)
	ctx := context.Background()

	baseApp := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "argocd",
			Name:      "test-app",
		},
	}

	noFinalizersApp := baseApp.DeepCopy()

	finalizersApp := baseApp.DeepCopy()
	finalizersApp.SetFinalizers([]string{v1alpha1.ApplicationFinalizerName, v1alpha1.ArgoCDResourcesFinalizer})

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		app *v1alpha1.Application
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "empty finalizer",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(noFinalizersApp).Build(),
				log:    logr.Logger{},
			},
			args: args{
				app: noFinalizersApp,
			},
			wantErr: false,
		},
		{
			name: "with finalizer",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(finalizersApp).Build(),
				log:    logr.Logger{},
			},
			args: args{
				app: finalizersApp,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			if err := r.RemoveAppFinalizer(tt.args.app); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAppFinalizer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				app := &v1alpha1.Application{}
				err = r.Client.Get(ctx, types.NamespacedName{
					Namespace: tt.args.app.Namespace,
					Name:      tt.args.app.Name,
				}, app)
				assert.NoError(t, err)
				assert.Empty(t, app.Finalizers)
			}
		})
	}
}
