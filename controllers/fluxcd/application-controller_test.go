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
	"fmt"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/controllers/core"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	helmv2 "kubesphere.io/devops/pkg/external/fluxcd/helm/v2beta1"
	kusv1 "kubesphere.io/devops/pkg/external/fluxcd/kustomize/v1beta2"
	"kubesphere.io/devops/pkg/external/fluxcd/meta"
	sourcev1 "kubesphere.io/devops/pkg/external/fluxcd/source/v1beta2"
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

	err = helmv2.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	err = sourcev1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	err = kusv1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	argoApp := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-app",
		},
		Spec: v1alpha1.ApplicationSpec{
			Kind: v1alpha1.ArgoCD,
		},
	}

	fluxApp := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-app",
		},
		Spec: v1alpha1.ApplicationSpec{
			Kind: "fluxcd",
			FluxApp: &v1alpha1.FluxApplication{
				Spec: v1alpha1.FluxApplicationSpec{
					Source: &v1alpha1.FluxApplicationSource{
						SourceRef: helmv2.CrossNamespaceObjectReference{
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
										KubeConfig: &helmv2.KubeConfig{
											SecretRef: meta.SecretKeyReference{
												Name: "aliyun-kubeconfig",
												Key:  "kubeconfig",
											},
										},
										TargetNamespace: "fake-targetNamespace",
									},
								},
								{
									Destination: v1alpha1.FluxApplicationDestination{
										KubeConfig: &helmv2.KubeConfig{
											SecretRef: meta.SecretKeyReference{
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

	invalidFluxApp1 := fluxApp.DeepCopy()
	invalidFluxApp1.Spec.FluxApp = nil

	invalidFluxApp2 := fluxApp.DeepCopy()
	invalidFluxApp2.Spec.FluxApp.Spec.Config = nil

	invalidFluxApp3 := fluxApp.DeepCopy()
	invalidFluxApp3.Spec.FluxApp.Spec.Source = nil

	invalidFluxApp4 := fluxApp.DeepCopy()
	invalidFluxApp4.Spec.FluxApp.Spec.Config.HelmRelease = nil

	invalidFluxApp5 := fluxApp.DeepCopy()
	invalidFluxApp5.Spec.FluxApp.Spec.Config.HelmRelease.Chart = nil

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
						Namespace: "fake-ns",
						Name:      "fake-app",
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
						Namespace: "fake-ns",
						Name:      "fake-app",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "found an invalid flux application that have no FluxApplication field",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, invalidFluxApp1.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-app",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "found an invalid flux application that have no Config",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, invalidFluxApp2.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-app",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "found an invalid flux application that have no Source",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, invalidFluxApp3.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-app",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "found an invalid flux application that have no HelmRelease",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, invalidFluxApp4.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-app",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "found an invalid flux application (HelmRelease) that have no chart",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, invalidFluxApp5.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-app",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "found a flux application",
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
				log:      logr.New(log.NullLogSink{}),
				recorder: &record.FakeRecorder{},
			}
			gotResult, err := r.Reconcile(context.Background(), tt.args.req)
			if tt.wantErr(t, err) {
				assert.Equal(t, tt.wantResult, gotResult)
			}
		})
	}
}

func TestApplicationReconciler_reconcileApp(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = helmv2.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	err = sourcev1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	err = kusv1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	helmApp := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-app",
			Namespace: "fake-ns",
		},
		Spec: v1alpha1.ApplicationSpec{
			Kind: "fluxcd",
			FluxApp: &v1alpha1.FluxApplication{
				Spec: v1alpha1.FluxApplicationSpec{
					Source: &v1alpha1.FluxApplicationSource{
						SourceRef: helmv2.CrossNamespaceObjectReference{
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
										KubeConfig: &helmv2.KubeConfig{
											SecretRef: meta.SecretKeyReference{
												Name: "aliyun-kubeconfig",
												Key:  "kubeconfig",
											},
										},
										TargetNamespace: "fake-targetNamespace",
									},
								},
								{
									Destination: v1alpha1.FluxApplicationDestination{
										KubeConfig: &helmv2.KubeConfig{
											SecretRef: meta.SecretKeyReference{
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
	helmAppWithUpdate.Spec.FluxApp.Spec.Config.HelmRelease.Deploy[0].ValuesFrom = []helmv2.ValuesReference{
		{
			Kind:      "ConfigMap",
			Name:      "fake-cm",
			ValuesKey: "fake-key",
		},
	}

	fluxHelmChart, err := buildTemplateFromApp(helmApp.Spec.FluxApp.DeepCopy())
	assert.Nil(t, err)
	fluxHelmChart.SetNamespace(helmApp.GetNamespace())
	fluxHelmChart.SetName(helmApp.GetName())
	fluxHelmChart.SetAnnotations(map[string]string{
		v1alpha1.HelmTemplateName: helmApp.GetName(),
	})

	helmDeploy := helmApp.Spec.FluxApp.Spec.Config.HelmRelease.Deploy[0]
	fluxHR, err := buildHelmRelease(fluxHelmChart, helmDeploy.DeepCopy())
	assert.Nil(t, err)
	fluxHR.SetNamespace(helmApp.GetNamespace())
	fluxHR.SetName(helmApp.GetName() + "abcde")
	fluxHR.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": helmApp.GetName(),
	})
	fluxHR.SetAnnotations(map[string]string{
		"app.kubernetes.io/name": getHelmReleaseName(helmDeploy),
	})

	helmAppWithTemplate := helmApp.DeepCopy()
	helmAppWithTemplate.Spec.FluxApp.Spec.Source = nil
	helmAppWithTemplate.Spec.FluxApp.Spec.Config.HelmRelease.Chart = nil
	helmAppWithTemplate.Spec.FluxApp.Spec.Config.HelmRelease.Template = "fake-app"

	helmAppWithNoInterval := helmAppWithLabel.DeepCopy()
	helmAppWithNoInterval.Spec.FluxApp.Spec.Config.HelmRelease.Chart.Interval = nil

	kusApp := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-app",
		},
		Spec: v1alpha1.ApplicationSpec{
			Kind: "fluxcd",
			FluxApp: &v1alpha1.FluxApplication{
				Spec: v1alpha1.FluxApplicationSpec{
					Source: &v1alpha1.FluxApplicationSource{
						SourceRef: helmv2.CrossNamespaceObjectReference{
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
									KubeConfig: &helmv2.KubeConfig{
										SecretRef: meta.SecretKeyReference{
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
									KubeConfig: &helmv2.KubeConfig{
										SecretRef: meta.SecretKeyReference{
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

	kusDeploy := kusApp.Spec.FluxApp.Spec.Config.Kustomization[0]
	fluxKus, err := buildKustomization(kusApp.Spec.FluxApp.DeepCopy(), kusDeploy.DeepCopy())
	assert.Nil(t, err)

	fluxKus.SetNamespace(kusApp.GetNamespace())
	fluxKus.SetName(kusApp.GetName() + "abcde")
	fluxKus.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": kusApp.GetName(),
	})
	fluxKus.SetAnnotations(map[string]string{
		"app.kubernetes.io/name": getKustomizationName(kusDeploy),
	})

	kusAppWithUpdate := kusApp.DeepCopy()
	kusAppWithUpdate.Spec.FluxApp.Spec.Config.Kustomization[0].Path = "another-kustomization"

	fakeApp := &v1alpha1.Application{
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
					Config: &v1alpha1.FluxApplicationConfig{
						HelmRelease:   nil,
						Kustomization: nil,
					},
				},
			},
		},
	}

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
			name: "reconcile a invalid fluxApp",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				app: fakeApp.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "create a Multi-Clusters FluxApp(HelmRelease) without saving Template",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				app: helmApp.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()

				fluxHRList := &helmv2.HelmReleaseList{}
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": appName,
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					// same settings
					assert.Equal(t, "./helm-chart", hr.Spec.Chart.Spec.Chart)
					assert.Equal(t, "0.1.0", hr.Spec.Chart.Spec.Version)
					assert.Equal(t, time.Minute, hr.Spec.Chart.Spec.Interval.Duration)
					assert.Equal(t, "Revision", hr.Spec.Chart.Spec.ReconcileStrategy)
					chartValueFiles := hr.Spec.Chart.Spec.ValuesFiles
					assert.Equal(t, 1, len(hr.Spec.Chart.Spec.ValuesFiles))
					assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
					switch hr.GetName() {
					case "fake-targetNamespace":
						assert.Equal(t, "aliyun-kubeconfig", hr.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", hr.Spec.KubeConfig.SecretRef.Key)
					case "another-fake-targetNamespace":
						assert.Equal(t, "tencentcloud-kubeconfig", hr.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", hr.Spec.KubeConfig.SecretRef.Key)
					}
				}

				fluxChart := &sourcev1.HelmChart{}
				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: appName}, fluxChart)
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

				fluxHRList := &helmv2.HelmReleaseList{}
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": appName,
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					// same settings
					assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", hr.Spec.Chart.Spec.SourceRef.APIVersion)
					assert.Equal(t, "GitRepository", hr.Spec.Chart.Spec.SourceRef.Kind)
					assert.Equal(t, "fake-repo", hr.Spec.Chart.Spec.SourceRef.Name)
					assert.Equal(t, "./helm-chart", hr.Spec.Chart.Spec.Chart)
					assert.Equal(t, "0.1.0", hr.Spec.Chart.Spec.Version)
					assert.Equal(t, time.Minute, hr.Spec.Chart.Spec.Interval.Duration)
					assert.Equal(t, "Revision", hr.Spec.Chart.Spec.ReconcileStrategy)
					chartValueFiles := hr.Spec.Chart.Spec.ValuesFiles
					assert.Equal(t, 1, len(chartValueFiles))
					assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
					switch hr.GetName() {
					case "fake-targetNamespace":
						assert.Equal(t, "aliyun-kubeconfig", hr.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", hr.Spec.KubeConfig.SecretRef.Key)
					case "another-fake-targetNamespace":
						assert.Equal(t, "tencentcloud-kubeconfig", hr.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", hr.Spec.KubeConfig.SecretRef.Key)
					}
				}

				fluxChart := &sourcev1.HelmChart{}
				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: appName}, fluxChart)
				assert.Nil(t, err)

				assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", fluxChart.Spec.SourceRef.APIVersion)
				assert.Equal(t, "GitRepository", fluxChart.Spec.SourceRef.Kind)
				assert.Equal(t, "fake-repo", fluxChart.Spec.SourceRef.Name)
				assert.Equal(t, "./helm-chart", fluxChart.Spec.Chart)
				assert.Equal(t, "0.1.0", fluxChart.Spec.Version)
				assert.Equal(t, time.Minute, fluxChart.Spec.Interval.Duration)
				assert.Equal(t, "Revision", fluxChart.Spec.ReconcileStrategy)
				chartValueFiles := fluxChart.Spec.ValuesFiles
				assert.Equal(t, 1, len(chartValueFiles))
				assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
			},
		},
		{
			name: "create the Multi-Clusters FluxApp(HelmRelease) but with no Interval for HelmChart",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				app: helmAppWithNoInterval.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				fluxChart := &sourcev1.HelmChart{}
				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: appName}, fluxChart)
				assert.Nil(t, err)

				assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", fluxChart.Spec.SourceRef.APIVersion)
				assert.Equal(t, "GitRepository", fluxChart.Spec.SourceRef.Kind)
				assert.Equal(t, "fake-repo", fluxChart.Spec.SourceRef.Name)
				assert.Equal(t, "./helm-chart", fluxChart.Spec.Chart)
				assert.Equal(t, "0.1.0", fluxChart.Spec.Version)
				// The default interval is 10m0s
				assert.Equal(t, 10*time.Minute, fluxChart.Spec.Interval.Duration)
				assert.Equal(t, "Revision", fluxChart.Spec.ReconcileStrategy)
				chartValueFiles := fluxChart.Spec.ValuesFiles
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

				fluxHRList := &helmv2.HelmReleaseList{}
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": appName,
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					name := hr.GetAnnotations()["app.kubernetes.io/name"]
					if name == "fake-targetNamespace" {
						valuesFrom := hr.Spec.ValuesFrom[0]

						assert.Equal(t, "ConfigMap", valuesFrom.Kind)
						assert.Equal(t, "fake-cm", valuesFrom.Name)
						assert.Equal(t, "fake-key", valuesFrom.ValuesKey)
					} else if name == "another-fake-targetNamespace" {
						valuesFromSlice := hr.Spec.ValuesFrom
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

				fluxHRList := &helmv2.HelmReleaseList{}
				appNS, appName := helmApp.GetNamespace(), helmApp.GetName()

				err = Client.List(ctx, fluxHRList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": appName,
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(fluxHRList.Items))

				for _, hr := range fluxHRList.Items {
					// same settings
					assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", hr.Spec.Chart.Spec.SourceRef.APIVersion)
					assert.Equal(t, "GitRepository", hr.Spec.Chart.Spec.SourceRef.Kind)
					assert.Equal(t, "fake-repo", hr.Spec.Chart.Spec.SourceRef.Name)
					assert.Equal(t, "./helm-chart", hr.Spec.Chart.Spec.Chart)
					assert.Equal(t, "0.1.0", hr.Spec.Chart.Spec.Version)
					assert.Equal(t, time.Minute, hr.Spec.Chart.Spec.Interval.Duration)
					assert.Equal(t, "Revision", hr.Spec.Chart.Spec.ReconcileStrategy)
					chartValueFiles := hr.Spec.Chart.Spec.ValuesFiles
					assert.Equal(t, 1, len(chartValueFiles))
					assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
					switch hr.GetName() {
					case "fake-targetNamespace":
						assert.Equal(t, "aliyun-kubeconfig", hr.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", hr.Spec.KubeConfig.SecretRef.Key)
					case "another-fake-targetNamespace":
						assert.Equal(t, "tencentcloud-kubeconfig", hr.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", hr.Spec.KubeConfig.SecretRef.Key)
					}
				}

				fluxChart := &sourcev1.HelmChart{}
				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: appName}, fluxChart)
				assert.Nil(t, err)

				assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", fluxChart.Spec.SourceRef.APIVersion)
				assert.Equal(t, "GitRepository", fluxChart.Spec.SourceRef.Kind)
				assert.Equal(t, "fake-repo", fluxChart.Spec.SourceRef.Name)
				assert.Equal(t, "./helm-chart", fluxChart.Spec.Chart)
				assert.Equal(t, "0.1.0", fluxChart.Spec.Version)
				assert.Equal(t, time.Minute, fluxChart.Spec.Interval.Duration)
				assert.Equal(t, "Revision", fluxChart.Spec.ReconcileStrategy)
				chartValueFiles := fluxChart.Spec.ValuesFiles
				assert.Equal(t, 1, len(chartValueFiles))
				assert.Equal(t, "./helm-chart/values.yaml", chartValueFiles[0])
			},
		},
		{
			name: "create a Multi-Clusters FluxApp(HelmRelease) by using a Template",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				app: helmAppWithTemplate.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.NotNil(t, err)
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

				kusList := &kusv1.KustomizationList{}
				appNS, appName := kusApp.GetNamespace(), kusApp.GetName()
				err = Client.List(ctx, kusList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": appName,
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(kusList.Items))

				for _, kus := range kusList.Items {
					// same settings
					assert.Equal(t, "source.toolkit.fluxcd.io/v1beta2", kus.Spec.SourceRef.APIVersion)
					assert.Equal(t, "GitRepository", kus.Spec.SourceRef.Kind)
					assert.Equal(t, "fake-repo", kus.Spec.SourceRef.Name)
					assert.Equal(t, "fake-ns", kus.Spec.SourceRef.Namespace)
					assert.Equal(t, "kustomization", kus.Spec.Path)
					assert.True(t, kus.Spec.Prune)

					switch kus.GetAnnotations()["app.kubernetes.io/name"] {
					case "fake-targetNamespace":
						assert.Equal(t, "aliyun-kubeconfig", kus.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", kus.Spec.KubeConfig.SecretRef.Key)
					case "another-fake-targetNamespace":
						assert.Equal(t, "tencentcloud-kubeconfig", kus.Spec.KubeConfig.SecretRef.Name)
						assert.Equal(t, "kubeconfig", kus.Spec.KubeConfig.SecretRef.Key)
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

				kusList := &kusv1.KustomizationList{}
				appNS, appName := kusApp.GetNamespace(), kusApp.GetName()
				err = Client.List(ctx, kusList, client.InNamespace(appNS), client.MatchingLabels{
					"app.kubernetes.io/managed-by": appName,
				})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(kusList.Items))

				for _, kus := range kusList.Items {
					if kus.GetAnnotations()["app.kubernetes.io/name"] == "fake-targetNamespace" {
						assert.Equal(t, "another-kustomization", kus.Spec.Path)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      logr.New(log.NullLogSink{}),
				recorder: &record.FakeRecorder{},
			}
			err := r.reconcileFluxApp(tt.args.app)
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

func TestApplicationReconciler_getHelmReleaseName(t *testing.T) {
	hostDeploy := &v1alpha1.Deploy{
		Destination: v1alpha1.FluxApplicationDestination{
			TargetNamespace: "default",
		},
	}

	memberDeploy := &v1alpha1.Deploy{
		Destination: v1alpha1.FluxApplicationDestination{
			KubeConfig: &helmv2.KubeConfig{SecretRef: meta.SecretKeyReference{
				Name: "qingcloud",
				Key:  "value",
			}},
			TargetNamespace: "default",
		},
	}

	tests := []struct {
		name   string
		args   *v1alpha1.Deploy
		expect string
	}{
		{
			name:   "in the host cluster (no kubeconfig)",
			args:   hostDeploy,
			expect: "default",
		},
		{
			name:   "in the member cluster (have kubeconfig)",
			args:   memberDeploy,
			expect: "qingcloud-default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, getHelmReleaseName(tt.args))
		})
	}
}

func TestApplicationReconciler_getKustomizationName(t *testing.T) {
	hostDeploy := &v1alpha1.KustomizationSpec{
		Destination: v1alpha1.FluxApplicationDestination{
			TargetNamespace: "default",
		},
	}

	memberDeploy := &v1alpha1.KustomizationSpec{
		Destination: v1alpha1.FluxApplicationDestination{
			KubeConfig: &helmv2.KubeConfig{SecretRef: meta.SecretKeyReference{
				Name: "qingcloud",
				Key:  "value",
			}},
			TargetNamespace: "default",
		},
	}

	tests := []struct {
		name   string
		args   *v1alpha1.KustomizationSpec
		expect string
	}{
		{
			name:   "in the host cluster (no kubeconfig)",
			args:   hostDeploy,
			expect: "default",
		},
		{
			name:   "in the member cluster (have kubeconfig)",
			args:   memberDeploy,
			expect: "qingcloud-default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, getKustomizationName(tt.args))
		})
	}
}

func TestApplicationReconciler_convertKubeconfig(t *testing.T) {
	tests := []struct {
		name   string
		args   *helmv2.KubeConfig
		verify func(kubeconfig *kusv1.KubeConfig)
	}{
		{
			name: "host cluster",
			args: nil,
			verify: func(kubeconfig *kusv1.KubeConfig) {
				assert.Nil(t, kubeconfig)
			},
		},
		{
			name: "member cluster",
			args: &helmv2.KubeConfig{SecretRef: meta.SecretKeyReference{
				Name: "qingcloud",
				Key:  "value",
			}},
			verify: func(kubeconfig *kusv1.KubeConfig) {
				assert.Equal(t, "qingcloud", kubeconfig.SecretRef.Name)
				assert.Equal(t, "value", kubeconfig.SecretRef.Key)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(convertKubeconfig(tt.args))
		})
	}
}

func TestApplicationReconciler_convertInterval(t *testing.T) {
	tests := []struct {
		name   string
		args   *metav1.Duration
		expect metav1.Duration
	}{
		{
			name:   "did not set interval",
			args:   nil,
			expect: metav1.Duration{Duration: 10 * time.Minute},
		},
		{
			name:   "set interval",
			args:   &metav1.Duration{Duration: 5 * time.Minute},
			expect: metav1.Duration{Duration: 5 * time.Minute},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect.Duration, convertInterval(tt.args).Duration)
		})
	}
}

func TestApplicationReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
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
				Client: fake.NewFakeClientWithScheme(schema),
			},
		},
		wantErr: core.NoErrors,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			tt.wantErr(t, r.SetupWithManager(tt.args.mgr), fmt.Sprintf("SetupWithManager(%v)", tt.args.mgr))
		})
	}
}
