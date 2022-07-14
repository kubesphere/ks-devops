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
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apischema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

func TestApplicationReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	argoApp := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-app",
			Namespace: "fake-ns",
		},
		Spec: v1alpha1.ApplicationSpec{
			Kind: v1alpha1.ArgoCD,
		},
	}

	fluxApp := argoApp.DeepCopy()
	fluxApp.Spec.Kind = v1alpha1.FluxCD

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
			name: "not found an application",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-app",
						Name:      "fake-ns",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "found an argocd application",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, argoApp.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-app",
						Name:      "fake-ns",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "found a fluxcd application",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, fluxApp.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-app",
					},
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      log.NullLogger{},
				recorder: &record.FakeRecorder{},
			}
			gotResult, err := r.Reconcile(tt.args.req)
			if tt.wantErr(t, err) {
				assert.Equal(t, tt.wantResult, gotResult)
			}
		})
	}
}

func TestApplicationReconciler_reconcileApp(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	schema.AddKnownTypeWithName(apischema.GroupVersionKind{
		Group:   "helm.toolkit.fluxcd.io",
		Version: "v2beta1",
		Kind:    "HelmReleaseList",
	}, &unstructured.UnstructuredList{})

	schema.AddKnownTypeWithName(apischema.GroupVersionKind{
		Group:   "kustomize.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "KustomizationList",
	}, &unstructured.UnstructuredList{})

	schema.AddKnownTypeWithName(apischema.GroupVersionKind{
		Group:   "source.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "HelmChartList",
	}, &unstructured.UnstructuredList{})

	helmApp := &v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gitops.kubesphere.io/v1alpha1",
			Kind:       "Application",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-app",
			Namespace: "fake-ns",
		},
		Spec: v1alpha1.ApplicationSpec{
			Kind: "fluxcd",
			FluxApp: &v1alpha1.FluxApplication{
				Spec: v1alpha1.FluxApplicationSpec{
					Source: &v1alpha1.FluxApplicationSource{
						SourceRef: v1alpha1.CrossNamespaceObjectReference{
							APIVersion: "source.toolkit.fluxcd.io/v1beta2",
							Kind:       "GitRepository",
							Name:       "fake-repo",
							Namespace:  "fake-ns",
						},
					},
					Config: &v1alpha1.FluxApplicationConfig{
						HelmRelease: &v1alpha1.HelmReleaseSpec{
							Chart: &v1alpha1.HelmChartTemplateSpec{
								Chart:   "./helm-chart",
								Version: "0.1.0",
								Interval: &metav1.Duration{
									Duration: time.Minute,
								},
								ReconcileStrategy: "Revision",
								ValuesFiles: []string{
									"./helm-chart/values.yaml",
								},
							},
							Deploy: []*v1alpha1.Deploy{
								{
									Destination: v1alpha1.FluxApplicationDestination{
										KubeConfig: &v1alpha1.KubeConfig{
											SecretRef: v1alpha1.SecretKeyReference{
												Name: "aliyun-kubeconfig",
												Key:  "kubeconfig",
											},
										},
										TargetNamespace: "fake-targetNamespace",
									},
								},
								{
									Destination: v1alpha1.FluxApplicationDestination{
										KubeConfig: &v1alpha1.KubeConfig{
											SecretRef: v1alpha1.SecretKeyReference{
												Name: "tencentcloud-kubeconfig",
												Key:  "kubeconfig",
											},
										},
										TargetNamespace: "another-fake-targetNamespace",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	helmAppWithLabel := helmApp.DeepCopy()
	helmAppWithLabel.SetLabels(map[string]string{
		v1alpha1.SaveTemplateLabelKey: "true",
	})

	helmAppWithUpdate := helmApp.DeepCopy()
	helmAppWithUpdate.Spec.FluxApp.Spec.Config.HelmRelease.Deploy[0].ValuesFrom = []v1alpha1.ValuesReference{
		{
			Kind:      "ConfigMap",
			Name:      "fake-cm",
			ValuesKey: "fake-key",
		},
	}

	fluxHelmChart := createUnstructuredFluxHelmTemplate(helmApp)
	fluxHelmChart.SetNamespace(helmApp.GetNamespace())
	fluxHelmChart.SetName(getHelmTemplateName(helmApp.GetNamespace(), helmApp.GetName()))
	fluxHelmChart.SetLabels(map[string]string{
		v1alpha1.HelmTemplateName: getHelmTemplateName(helmApp.GetNamespace(), helmApp.GetName()),
	})

	fluxHR := createBareFluxHelmReleaseObject()
	helmDeploy := helmApp.Spec.FluxApp.Spec.Config.HelmRelease.Deploy[0]
	setFluxHelmReleaseFields(fluxHR, fluxHelmChart, helmDeploy)
	fluxHR.SetNamespace(helmApp.GetNamespace())
	fluxHR.SetName(getHelmReleaseName(helmDeploy.Destination.TargetNamespace))
	fluxHR.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": getHelmTemplateName(helmApp.GetNamespace(), helmApp.GetName()),
	})

	helmAppWithTemplate := helmApp.DeepCopy()
	helmAppWithTemplate.Spec.FluxApp.Spec.Source = nil
	helmAppWithTemplate.Spec.FluxApp.Spec.Config.HelmRelease.Chart = nil
	helmAppWithTemplate.Spec.FluxApp.Spec.Config.HelmRelease.Template = "fake-ns-fake-app"

	kusApp := &v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			Kind: "fluxcd",
			FluxApp: &v1alpha1.FluxApplication{
				Spec: v1alpha1.FluxApplicationSpec{
					Source: &v1alpha1.FluxApplicationSource{
						SourceRef: v1alpha1.CrossNamespaceObjectReference{
							APIVersion: "source.toolkit.fluxcd.io/v1beta2",
							Kind:       "GitRepository",
							Name:       "fake-repo",
							Namespace:  "fake-ns",
						},
					},
					Config: &v1alpha1.FluxApplicationConfig{
						Kustomization: []*v1alpha1.KustomizationSpec{
							{
								Destination: v1alpha1.FluxApplicationDestination{
									KubeConfig: &v1alpha1.KubeConfig{
										SecretRef: v1alpha1.SecretKeyReference{
											Name: "aliyun-kubeconfig",
											Key:  "kubeconfig",
										},
									},
									TargetNamespace: "fake-targetNamespace",
								},
								Path:  "kustomization",
								Prune: true,
							},
							{
								Destination: v1alpha1.FluxApplicationDestination{
									KubeConfig: &v1alpha1.KubeConfig{
										SecretRef: v1alpha1.SecretKeyReference{
											Name: "tencentcloud-kubeconfig",
											Key:  "kubeconfig",
										},
									},
									TargetNamespace: "another-fake-targetNamespace",
								},
								Path:  "kustomization",
								Prune: true,
							},
						},
					},
				},
			},
		},
	}

	fluxKus := createBareFluxKustomizationObject()
	KusDeploy := kusApp.Spec.FluxApp.Spec.Config.Kustomization[0]
	setFluxKustomizationFields(fluxKus, kusApp, KusDeploy)
	fluxKus.SetNamespace(kusApp.GetNamespace())
	fluxKus.SetName(getKustomizationName(KusDeploy.Destination.TargetNamespace))
	fluxKus.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": getKusManagerName(kusApp.GetNamespace(), kusApp.GetName()),
	})

	kusAppWithUpdate := kusApp.DeepCopy()
	kusAppWithUpdate.Spec.FluxApp.Spec.Config.Kustomization[0].Path = "another-kustomization"

	type fields struct {
		Client client.Client
	}
	type args struct {
		app *v1alpha1.Application
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(t *testing.T, Client client.Client, err error)
	}{
		{
			name: "create a Multi-Clusters FluxApp(HelmRelease) without saving Template",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				app: helmApp.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				ctx := context.Background()
				assert.Nil(t, err)

				fluxHRList := createBareFluxHelmReleaseListObject()
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": getHelmTemplateName(appNS, appName),
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					// same settings
					chart, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "chart")
					assert.Equal(t, "./helm-chart", chart)
					chartVersion, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "version")
					assert.Equal(t, "0.1.0", chartVersion)
					chartInterval, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "interval")
					assert.Equal(t, "1m0s", chartInterval)
					chartReconcileStrategy, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "reconcileStrategy")
					assert.Equal(t, "Revision", chartReconcileStrategy)
					chartValueFiles, _, _ := unstructured.NestedStringSlice(hr.Object, "spec", "chart", "spec", "valuesFiles")
					assert.Equal(t, 1, len(chartValueFiles))
					assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
					switch hr.GetName() {
					case "fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "aliyun-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					case "another-fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "tencentcloud-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					}
				}

				fluxChart := createBareFluxHelmTemplateObject()
				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: getHelmTemplateName(appNS, appName)}, fluxChart)
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
		{
			name: "create a Multi-Clusters FluxApp(HelmRelease) and save Template",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				app: helmAppWithLabel.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()

				fluxHRList := createBareFluxHelmReleaseListObject()
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": getHelmTemplateName(appNS, appName),
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					// same settings
					sourceAPIVersion, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "apiVersion")
					assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", sourceAPIVersion)
					sourceKind, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "kind")
					assert.Equal(t, "GitRepository", sourceKind)
					sourceName, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "name")
					assert.Equal(t, "fake-repo", sourceName)
					sourceNS, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "namespace")
					assert.Equal(t, "fake-ns", sourceNS)
					chart, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "chart")
					assert.Equal(t, "./helm-chart", chart)
					chartVersion, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "version")
					assert.Equal(t, "0.1.0", chartVersion)
					chartInterval, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "interval")
					assert.Equal(t, "1m0s", chartInterval)
					chartReconcileStrategy, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "reconcileStrategy")
					assert.Equal(t, "Revision", chartReconcileStrategy)
					chartValueFiles, _, _ := unstructured.NestedStringSlice(hr.Object, "spec", "chart", "spec", "valuesFiles")
					assert.Equal(t, 1, len(chartValueFiles))
					assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
					switch hr.GetName() {
					case "fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "aliyun-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					case "another-fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "tencentcloud-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					}
				}

				fluxChart := createBareFluxHelmTemplateObject()
				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: getHelmTemplateName(appNS, appName)}, fluxChart)
				assert.Nil(t, err)

				sourceAPIVersion, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "apiVersion")
				assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", sourceAPIVersion)
				sourceKind, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "kind")
				assert.Equal(t, "GitRepository", sourceKind)
				sourceName, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "name")
				assert.Equal(t, "fake-repo", sourceName)
				sourceNS, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "namespace")
				assert.Equal(t, "fake-ns", sourceNS)
				chart, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "chart")
				assert.Equal(t, "./helm-chart", chart)
				chartVersion, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "version")
				assert.Equal(t, "0.1.0", chartVersion)
				chartInterval, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "interval")
				assert.Equal(t, "1m0s", chartInterval)
				chartReconcileStrategy, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "reconcileStrategy")
				assert.Equal(t, "Revision", chartReconcileStrategy)
				chartValueFiles, _, _ := unstructured.NestedStringSlice(fluxChart.Object, "spec", "valuesFiles")
				assert.Equal(t, 1, len(chartValueFiles))
				assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
			},
		},
		{
			name: "update the Multi-Clusters FluxApp(HelmRelease)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, fluxHR.DeepCopy()),
			},
			args: args{
				app: helmAppWithUpdate.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()

				fluxHRList := createBareFluxHelmReleaseListObject()
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": getHelmTemplateName(appNS, appName),
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					if hr.GetName() == "fake-targetNamespace" {
						valuesFromSlice, _, _ := unstructured.NestedSlice(hr.Object, "spec", "valuesFrom")
						valuesFrom := valuesFromSlice[0].(map[string]interface{})
						assert.Equal(t, "ConfigMap", valuesFrom["kind"].(string))
						assert.Equal(t, "fake-cm", valuesFrom["name"].(string))
						assert.Equal(t, "fake-key", valuesFrom["valuesKey"].(string))
					} else if hr.GetName() == "another-fake-targetNamespace" {
						valuesFromSlice, _, _ := unstructured.NestedSlice(hr.Object, "spec", "valuesFrom")
						assert.Nil(t, valuesFromSlice)
					}
				}
			},
		},
		{
			name: "create a Multi-Clusters FluxApp(HelmRelease) by using a Template",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, fluxHelmChart.DeepCopy()),
			},
			args: args{
				app: helmAppWithTemplate.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)

				ctx := context.Background()

				fluxHRList := createBareFluxHelmReleaseListObject()
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": getHelmTemplateName(appNS, appName),
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					// same settings
					sourceAPIVersion, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "apiVersion")
					assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", sourceAPIVersion)
					sourceKind, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "kind")
					assert.Equal(t, "GitRepository", sourceKind)
					sourceName, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "name")
					assert.Equal(t, "fake-repo", sourceName)
					sourceNS, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "namespace")
					assert.Equal(t, "fake-ns", sourceNS)
					chart, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "chart")
					assert.Equal(t, "./helm-chart", chart)
					chartVersion, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "version")
					assert.Equal(t, "0.1.0", chartVersion)
					chartInterval, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "interval")
					assert.Equal(t, "1m0s", chartInterval)
					chartReconcileStrategy, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "reconcileStrategy")
					assert.Equal(t, "Revision", chartReconcileStrategy)
					chartValueFiles, _, _ := unstructured.NestedStringSlice(hr.Object, "spec", "chart", "spec", "valuesFiles")
					assert.Equal(t, 1, len(chartValueFiles))
					assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
					switch hr.GetName() {
					case "fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "aliyun-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					case "another-fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "tencentcloud-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(hr.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					}
				}

				fluxChart := createBareFluxHelmTemplateObject()
				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: getHelmTemplateName(appNS, appName)}, fluxChart)
				assert.Nil(t, err)

				sourceAPIVersion, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "apiVersion")
				assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", sourceAPIVersion)
				sourceKind, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "kind")
				assert.Equal(t, "GitRepository", sourceKind)
				sourceName, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "name")
				assert.Equal(t, "fake-repo", sourceName)
				sourceNS, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "sourceRef", "namespace")
				assert.Equal(t, "fake-ns", sourceNS)
				chart, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "chart")
				assert.Equal(t, "./helm-chart", chart)
				chartVersion, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "version")
				assert.Equal(t, "0.1.0", chartVersion)
				chartInterval, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "interval")
				assert.Equal(t, "1m0s", chartInterval)
				chartReconcileStrategy, _, _ := unstructured.NestedString(fluxChart.Object, "spec", "reconcileStrategy")
				assert.Equal(t, "Revision", chartReconcileStrategy)
				chartValueFiles, _, _ := unstructured.NestedStringSlice(fluxChart.Object, "spec", "valuesFiles")
				assert.Equal(t, 1, len(chartValueFiles))
				assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
			},
		},
		{
			name: "create a  Multi-Clusters FluxApp(Kustomization)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				app: kusApp.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()

				kusList := createBareFluxKustomizationListObject()
				appNS, appName := kusApp.GetNamespace(), kusApp.GetName()
				err = Client.List(ctx, kusList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": getKusManagerName(appNS, appName),
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(kusList.Items))

				for _, kus := range kusList.Items {
					// same settings
					sourceAPIVersion, _, _ := unstructured.NestedString(kus.Object, "spec", "sourceRef", "apiVersion")
					assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", sourceAPIVersion)
					sourceKind, _, _ := unstructured.NestedString(kus.Object, "spec", "sourceRef", "kind")
					assert.Equal(t, "GitRepository", sourceKind)
					sourceName, _, _ := unstructured.NestedString(kus.Object, "spec", "sourceRef", "name")
					assert.Equal(t, "fake-repo", sourceName)
					sourceNS, _, _ := unstructured.NestedString(kus.Object, "spec", "sourceRef", "namespace")
					assert.Equal(t, "fake-ns", sourceNS)
					path, _, _ := unstructured.NestedString(kus.Object, "spec", "path")
					assert.Equal(t, "kustomization", path)
					prune, _, _ := unstructured.NestedBool(kus.Object, "spec", "prune")
					assert.True(t, prune)

					switch kus.GetName() {
					case "fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(kus.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "aliyun-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(kus.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					case "another-fake-targetNamespace":
						kubeconfigName, _, _ := unstructured.NestedString(kus.Object, "spec", "kubeConfig", "secretRef", "name")
						assert.Equal(t, "tencentcloud-kubeconfig", kubeconfigName)
						kubeconfigKey, _, _ := unstructured.NestedString(kus.Object, "spec", "kubeConfig", "secretRef", "key")
						assert.Equal(t, "kubeconfig", kubeconfigKey)
					}
				}
			},
		},
		{
			name: "update the Multi-Clusters FluxApp(Kustomization)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, fluxKus.DeepCopy()),
			},
			args: args{
				app: kusAppWithUpdate.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()

				kusList := createBareFluxKustomizationListObject()
				appNS, appName := kusApp.GetNamespace(), kusApp.GetName()
				err = Client.List(ctx, kusList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": getKusManagerName(appNS, appName),
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(kusList.Items))

				for _, kus := range kusList.Items {
					if kus.GetName() == "fake-targetNamespace" {
						path, _, _ := unstructured.NestedString(kus.Object, "spec", "path")
						assert.Equal(t, "another-kustomization", path)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      log.NullLogger{},
				recorder: &record.FakeRecorder{},
			}
			err := r.reconcileApp(tt.args.app)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func TestApplicationReconciler_GetName(t *testing.T) {
	t.Run("get FluxApplicationReconciler name", func(t *testing.T) {

		r := &ApplicationReconciler{}
		assert.Equal(t, "FluxApplicationReconciler", r.GetName())
	})
}

func TestApplicationReconciler_GetGroupName(t *testing.T) {
	t.Run("get Flux Controllers GroupName", func(t *testing.T) {
		r := &ApplicationReconciler{}
		assert.Equal(t, "fluxcd", r.GetGroupName())
	})
}
