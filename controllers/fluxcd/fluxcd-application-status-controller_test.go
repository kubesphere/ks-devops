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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

func TestApplicationStatusReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = helmv2.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	err = kusv1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	fluxHelmApp := &v1alpha1.Application{
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
							},
						},
					},
				},
			},
		},
	}

	fluxKusApp := &v1alpha1.Application{
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
						},
					},
				},
			},
		},
	}

	hr := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-appabcde",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "fake-app",
			},
			Annotations: map[string]string{
				"app.kubernetes.io/name": "aliyun-kubeconfig-fake-targetNamespace",
			},
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						APIVersion: "source.toolkit.fluxcd.io/v1beta2",
						Kind:       "GitRepository",
						Name:       "fake-repo",
						Namespace:  "fake-ns",
					},
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
			},
			Interval: metav1.Duration{},
			KubeConfig: &helmv2.KubeConfig{
				SecretRef: meta.SecretKeyReference{
					Name: "aliyun-kubeconfig",
					Key:  "kubeconfig",
				},
			},
			TargetNamespace: "fake-targetNamespace",
		},
	}

	kus := &kusv1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-appabcde",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "fake-app",
			},
			Annotations: map[string]string{
				"app.kubernetes.io/name": "aliyun-kubeconfig-fake-targetNamespace",
			},
		},
		Spec: kusv1.KustomizationSpec{
			Interval: metav1.Duration{Duration: time.Minute},
			KubeConfig: &kusv1.KubeConfig{
				SecretRef: meta.SecretKeyReference{
					Name: "aliyun-kubeconfig",
					Key:  "kubeconfig",
				},
			},
			Path:  "kustomization",
			Prune: true,
			SourceRef: kusv1.CrossNamespaceSourceReference{
				APIVersion: "source.toolkit.fluxcd.io/v1beta2",
				Kind:       "GitRepository",
				Name:       "fake-repo",
				Namespace:  "fake-ns",
			},
			TargetNamespace: "fake-targetNamespace",
		},
	}

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
			name: "found a HelmRelease",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, hr.DeepCopy(), fluxHelmApp.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-appabcde",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "found a Kustomization",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, kus.DeepCopy(), fluxKusApp.DeepCopy()),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-appabcde",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found a HelmRelease neither a Kustomization",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "fake-ns",
						Name:      "fake-name",
					},
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationStatusReconciler{
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

func TestApplicationStatusReconciler_reconcileHelmRelease(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = helmv2.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	fluxHelmApp := &v1alpha1.Application{
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
							},
						},
					},
				},
			},
		},
	}

	t1, _ := time.Parse("2006-01-02T15:04:05Z", "2022-09-05T13:57:03Z")
	unKnownHelmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-appabcde",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "fake-app",
			},
			Annotations: map[string]string{
				"app.kubernetes.io/name": "aliyun-kubeconfig-fake-targetNamespace",
			},
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						APIVersion: "source.toolkit.fluxcd.io/v1beta2",
						Kind:       "GitRepository",
						Name:       "fake-repo",
						Namespace:  "fake-ns",
					},
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
			},
			Interval: metav1.Duration{},
			KubeConfig: &helmv2.KubeConfig{
				SecretRef: meta.SecretKeyReference{
					Name: "aliyun-kubeconfig",
					Key:  "kubeconfig",
				},
			},
			TargetNamespace: "fake-targetNamespace",
		},
		Status: helmv2.HelmReleaseStatus{
			ObservedGeneration: 1,
			Conditions: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             "Unknown",
					ObservedGeneration: 0,
					LastTransitionTime: metav1.Time{Time: t1},
					Reason:             meta.ProgressingReason,
					Message:            "Reconciliation in progress",
				},
			},
			LastAttemptedRevision:       "0.1.0+1",
			LastAttemptedValuesChecksum: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			HelmChart:                   "my-devops-project5s2r8/my-devops-project5s2r8-test-fluxcdk9dbc",
		},
	}

	t2, _ := time.Parse("2006-01-02T15:04:05Z", "2022-09-06T05:14:44Z")
	t3, _ := time.Parse("2006-01-02T15:04:05Z", "2022-09-06T05:14:44Z")
	readyHelmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-appabcde",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "fake-app",
			},
			Annotations: map[string]string{
				"app.kubernetes.io/name": "aliyun-kubeconfig-fake-targetNamespace",
			},
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						APIVersion: "source.toolkit.fluxcd.io/v1beta2",
						Kind:       "GitRepository",
						Name:       "fake-repo",
						Namespace:  "fake-ns",
					},
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
			},
			Interval: metav1.Duration{},
			KubeConfig: &helmv2.KubeConfig{
				SecretRef: meta.SecretKeyReference{
					Name: "aliyun-kubeconfig",
					Key:  "kubeconfig",
				},
			},
			TargetNamespace: "fake-targetNamespace",
		},
		Status: helmv2.HelmReleaseStatus{
			ObservedGeneration: 1,
			Conditions: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: t2},
					Reason:             helmv2.ReconciliationSucceededReason,
					Message:            "Release reconciliation succeeded",
				},
				{
					Type:               helmv2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: t3},
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			LastAppliedRevision:         "0.1.0+1",
			LastAttemptedRevision:       "0.1.0+1",
			LastAttemptedValuesChecksum: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			LastReleaseRevision:         1,
			HelmChart:                   "my-devops-project5s2r8/my-devops-project5s2r8-test-fluxcd2hbs2",
		},
	}

	type fields struct {
		Client client.Client
	}

	type args struct {
		hr *helmv2.HelmRelease
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(t *testing.T, Client client.Client, err error)
	}{
		{
			name: "update Application's status (a Unknown HelmRelease)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, fluxHelmApp.DeepCopy()),
			},
			args: args{
				hr: unKnownHelmRelease.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()
				app := &v1alpha1.Application{}
				appNS, appName := fluxHelmApp.GetNamespace(), fluxHelmApp.GetName()

				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: appName}, app)
				assert.Nil(t, err)

				// status
				assert.Equal(t, 1, len(app.Status.FluxApp.HelmReleaseStatus))
				status := app.Status.FluxApp.HelmReleaseStatus["aliyun-kubeconfig-fake-targetNamespace"]
				assert.Equal(t, 1, int(status.ObservedGeneration))
				c := status.Conditions[0]
				assert.Equal(t, meta.ReadyCondition, c.Type)
				assert.Equal(t, "Unknown", string(c.Status))
				assert.Equal(t, 0, int(c.ObservedGeneration))
				assert.True(t, t1.Equal(c.LastTransitionTime.Time))
				assert.Equal(t, "Progressing", c.Reason)
				assert.Equal(t, "Reconciliation in progress", c.Message)
				assert.Equal(t, "0.1.0+1", status.LastAttemptedRevision)
				assert.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", status.LastAttemptedValuesChecksum)
				assert.Equal(t, "my-devops-project5s2r8/my-devops-project5s2r8-test-fluxcdk9dbc", status.HelmChart)

				// labels
				assert.Equal(t, string(HelmRelease), app.GetLabels()[FluxAppTypeKey])
				assert.Equal(t, "0-1", app.GetLabels()[FluxAppReadyNumKey])
			},
		},
		{
			name: "update Application's status (a Ready HelmRelease)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, fluxHelmApp.DeepCopy()),
			},
			args: args{
				hr: readyHelmRelease.DeepCopy(),
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()
				app := &v1alpha1.Application{}
				appNS, appName := fluxHelmApp.GetNamespace(), fluxHelmApp.GetName()

				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: appName}, app)
				assert.Nil(t, err)

				assert.Equal(t, 1, len(app.Status.FluxApp.HelmReleaseStatus))
				status := app.Status.FluxApp.HelmReleaseStatus["aliyun-kubeconfig-fake-targetNamespace"]
				assert.Equal(t, 1, int(status.ObservedGeneration))
				c1, c2 := status.Conditions[0], status.Conditions[1]
				assert.Equal(t, meta.ReadyCondition, c1.Type)
				assert.Equal(t, metav1.ConditionTrue, c1.Status)
				assert.Equal(t, helmv2.ReconciliationSucceededReason, c1.Reason)
				assert.Equal(t, "Release reconciliation succeeded", c1.Message)
				assert.True(t, t2.Equal(c1.LastTransitionTime.Time))

				assert.Equal(t, helmv2.ReleasedCondition, c2.Type)
				assert.Equal(t, metav1.ConditionTrue, c2.Status)
				assert.True(t, t3.Equal(c2.LastTransitionTime.Time))
				assert.Equal(t, helmv2.InstallSucceededReason, c2.Reason)
				assert.Equal(t, "Helm install succeeded", c2.Message)

				assert.Equal(t, "0.1.0+1", status.LastAppliedRevision)
				assert.Equal(t, "0.1.0+1", status.LastAttemptedRevision)
				assert.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", status.LastAttemptedValuesChecksum)
				assert.Equal(t, 1, status.LastReleaseRevision)
				assert.Equal(t, "my-devops-project5s2r8/my-devops-project5s2r8-test-fluxcd2hbs2", status.HelmChart)

				// labels
				assert.Equal(t, string(HelmRelease), app.GetLabels()[FluxAppTypeKey])
				assert.Equal(t, "1-1", app.GetLabels()[FluxAppReadyNumKey])
			},
		},
		{
			name: "not found a app that manage this HelmRelease",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				hr: readyHelmRelease,
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ApplicationStatusReconciler{
				Client:   tt.fields.Client,
				log:      logr.New(log.NullLogSink{}),
				recorder: &record.FakeRecorder{},
			}
			_, err := r.reconcileHelmRelease(context.Background(), tt.args.hr)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func TestApplicationStatusReconciler_reconcileKustomization(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = kusv1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	fluxKusApp := &v1alpha1.Application{
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
						},
					},
				},
			},
		},
	}

	t1, _ := time.Parse("2006-01-02T15:04:05Z", "2022-09-06T07:46:47Z")
	readyKus := &kusv1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-ns",
			Name:      "fake-appabcde",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "fake-app",
			},
			Annotations: map[string]string{
				"app.kubernetes.io/name": "aliyun-kubeconfig-fake-targetNamespace",
			},
		},
		Spec: kusv1.KustomizationSpec{
			SourceRef: kusv1.CrossNamespaceSourceReference{
				APIVersion: "source.toolkit.fluxcd.io/v1beta2",
				Kind:       "GitRepository",
				Name:       "fake-repo",
				Namespace:  "fake-ns",
			},
			KubeConfig: &kusv1.KubeConfig{
				SecretRef: meta.SecretKeyReference{
					Name: "aliyun-kubeconfig",
					Key:  "kubeconfig",
				},
			},
			TargetNamespace: "fake-targetNamespace",
			Path:            "kustomization",
			Prune:           true,
		},
		Status: kusv1.KustomizationStatus{
			ObservedGeneration: 1,
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: t1},
					Message:            "Applied revision: master/4b8497c8c3dc8c7ad5c9cb66160dbd3bcc1cfd4f",
					Reason:             kusv1.ReconciliationSucceededReason,
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionTrue,
				},
			},
			LastAppliedRevision:   "master/4b8497c8c3dc8c7ad5c9cb66160dbd3bcc1cfd4f",
			LastAttemptedRevision: "master/4b8497c8c3dc8c7ad5c9cb66160dbd3bcc1cfd4f",
			Inventory: &kusv1.ResourceInventory{
				Entries: []kusv1.ResourceRef{
					{
						ID:      "default_nginx-svc__Service",
						Version: "v1",
					},
					{
						ID:      "default_nginx-deployment_apps_Deployment",
						Version: "v1",
					},
				},
			},
		},
	}

	type fields struct {
		Client client.Client
	}

	type args struct {
		kus *kusv1.Kustomization
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(t *testing.T, Client client.Client, err error)
	}{
		{
			name: "update Application's status (a Kustomization)",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema, fluxKusApp.DeepCopy()),
			},
			args: args{
				kus: readyKus,
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.Nil(t, err)
				ctx := context.Background()
				app := &v1alpha1.Application{}
				appNS, appName := fluxKusApp.GetNamespace(), fluxKusApp.GetName()

				err = Client.Get(ctx, types.NamespacedName{Namespace: appNS, Name: appName}, app)
				assert.Nil(t, err)

				// status
				assert.Equal(t, 1, len(app.Status.FluxApp.KustomizationStatus))
				status := app.Status.FluxApp.KustomizationStatus["aliyun-kubeconfig-fake-targetNamespace"]
				assert.Equal(t, 1, int(status.ObservedGeneration))

				assert.Equal(t, 1, int(status.ObservedGeneration))
				c := status.Conditions[0]
				assert.True(t, t1.Equal(c.LastTransitionTime.Time))
				assert.Equal(t, "Applied revision: master/4b8497c8c3dc8c7ad5c9cb66160dbd3bcc1cfd4f", c.Message)
				assert.Equal(t, kusv1.ReconciliationSucceededReason, c.Reason)
				assert.Equal(t, meta.ReadyCondition, c.Type)
				assert.Equal(t, metav1.ConditionTrue, c.Status)
				assert.Equal(t, "master/4b8497c8c3dc8c7ad5c9cb66160dbd3bcc1cfd4f", status.LastAppliedRevision)
				assert.Equal(t, "master/4b8497c8c3dc8c7ad5c9cb66160dbd3bcc1cfd4f", status.LastAttemptedRevision)
				assert.Equal(t, "default_nginx-svc__Service", status.Inventory.Entries[0].ID)
				assert.Equal(t, "v1", status.Inventory.Entries[0].Version)
				assert.Equal(t, "default_nginx-deployment_apps_Deployment", status.Inventory.Entries[1].ID)
				assert.Equal(t, "v1", status.Inventory.Entries[1].Version)

				// labels
				assert.Equal(t, string(Kustomization), app.GetLabels()[FluxAppTypeKey])
				assert.Equal(t, "1-1", app.GetLabels()[FluxAppReadyNumKey])
			},
		},
		{
			name: "not found a app that manage this Kustomization",
			fields: fields{
				Client: fake.NewFakeClientWithScheme(schema),
			},
			args: args{
				kus: readyKus,
			},
			verify: func(t *testing.T, Client client.Client, err error) {
				assert.True(t, apierrors.IsNotFound(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ApplicationStatusReconciler{
				Client:   tt.fields.Client,
				log:      logr.New(log.NullLogSink{}),
				recorder: &record.FakeRecorder{},
			}
			_, err := r.reconcileKustomization(context.Background(), tt.args.kus)
			tt.verify(t, tt.fields.Client, err)
		})
	}
}

func TestApplicationStatusReconciler_GetName(t *testing.T) {
	t.Run("get ApplicationStatusReconciler name", func(t *testing.T) {

		r := &ApplicationStatusReconciler{}
		assert.Equal(t, "FluxCDApplicationStatusController", r.GetName())
	})
}

func TestApplicationStatusReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = helmv2.AddToScheme(schema)
	assert.Nil(t, err)

	err = kusv1.AddToScheme(schema)
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
			r := &ApplicationStatusReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			tt.wantErr(t, r.SetupWithManager(tt.args.mgr), fmt.Sprintf("SetupWithManager(%v)", tt.args.mgr))
		})
	}
}
