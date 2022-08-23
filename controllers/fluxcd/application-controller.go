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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	helmv2 "kubesphere.io/devops/pkg/external/fluxcd/helm/v2beta1"
	kusv1 "kubesphere.io/devops/pkg/external/fluxcd/kustomize/v1beta2"
	sourcev1 "kubesphere.io/devops/pkg/external/fluxcd/source/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=watch;get;list
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="kustomize.toolkit.fluxcd.io",resources=kustomizations,verbs=watch;get;list;create;update;delete
//+kubebuilder:rbac:groups="helm.toolkit.fluxcd.io",resources=helmreleases,verbs=watch;get;list;create;update;delete
//+kubebuilder:rbac:groups="source.toolkit.fluxcd.io",resources=helmcharts,verbs=watch;get;list;create

// ApplicationReconciler is the reconciler of the FluxCD HelmRelease and FluxCD Kustomization
type ApplicationReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile sync the Application with underlying FluxCD HelmRelease and FluxCD Kustomization CRD
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	r.log.Info(fmt.Sprintf("start to reconcile application: %s", req.String()))

	app := &v1alpha1.Application{}
	if err = r.Get(ctx, req.NamespacedName, app); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if isFluxApp(app) {
		err = r.reconcileFluxApp(app)
	}
	return
}

func isFluxApp(app *v1alpha1.Application) bool {
	return app.Spec.Kind == v1alpha1.FluxCD && app.Spec.FluxApp != nil
}

func (r *ApplicationReconciler) reconcileFluxApp(app *v1alpha1.Application) (err error) {
	var at AppType
	if at, err = isHelmOrKustomize(app); err != nil {
		return err
	}
	if at == HelmRelease {
		return r.reconcileHelm(app)
	}
	if at == Kustomization {
		return r.reconcileKustomization(app)
	}
	return
}

func (r *ApplicationReconciler) reconcileHelm(app *v1alpha1.Application) (err error) {
	ctx := context.Background()

	var helmChart *sourcev1.HelmChart
	// 1. get helmChart by searching existed helm template
	if helmTemplateName := app.Spec.FluxApp.Spec.Config.HelmRelease.Template; helmTemplateName != "" {
		helmTemplateNS := app.GetNamespace()

		helmChart = &sourcev1.HelmChart{}
		if err = r.Get(ctx, types.NamespacedName{Namespace: helmTemplateNS, Name: helmTemplateName}, helmChart); err != nil {
			return
		}
	} else {
		// 2. get helmChart by building from app
		helmChart = buildTemplateFromApp(app)

		if wantSaveHelmTemplate(app) {
			if err = r.saveTemplate(ctx, helmChart); err != nil {
				return
			}
		}
	}

	if err = r.reconcileHelmReleaseList(ctx, app, helmChart); err != nil {
		return
	}

	return
}

func (r *ApplicationReconciler) reconcileHelmReleaseList(ctx context.Context, app *v1alpha1.Application, helmChart *sourcev1.HelmChart) (err error) {
	fluxApp := app.Spec.FluxApp.DeepCopy()
	appNS, appName := app.GetNamespace(), app.GetName()

	fluxHelmReleaseList := &helmv2.HelmReleaseList{}
	if err = r.List(ctx, fluxHelmReleaseList, client.InNamespace(appNS), client.MatchingLabels{
		"app.kubernetes.io/managed-by": appName,
	}); err != nil {
		return
	}

	hrMap := make(map[string]*helmv2.HelmRelease, len(fluxHelmReleaseList.Items))
	for _, fluxHelmRelease := range fluxHelmReleaseList.Items {
		name := fluxHelmRelease.GetAnnotations()["app.kubernetes.io/name"]
		hrMap[name] = fluxHelmRelease.DeepCopy()
	}

	for _, deploy := range fluxApp.Spec.Config.HelmRelease.Deploy {
		name := getHelmReleaseName(deploy)
		if hr, ok := hrMap[name]; !ok {
			// there is no matching helmRelease
			// create
			if err = r.createHelmRelease(ctx, app, helmChart, deploy); err != nil {
				return
			}
		} else {
			// there is a matching helmRelease
			// update the helmRelease
			// TODO: determine whether this helmrelease should update by ResourceVersion
			if err = r.updateHelmRelease(ctx, hr, helmChart, deploy); err != nil {
				return
			}
		}

	}
	return
}

func (r *ApplicationReconciler) createHelmRelease(ctx context.Context, app *v1alpha1.Application, helmChart *sourcev1.HelmChart, deploy *v1alpha1.Deploy) (err error) {
	appNS, appName := app.GetNamespace(), app.GetName()
	hrNS := appNS

	hr := buildHelmRelease(helmChart, deploy)
	hr.SetNamespace(hrNS)
	hr.SetGenerateName(appName)
	hr.SetName("")
	hr.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": appName,
	})
	// put the name that can uniquely identifies the HelmRelease in annotations
	// and use generateName to avoid naming conflict
	hr.SetAnnotations(map[string]string{
		"app.kubernetes.io/name": getHelmReleaseName(deploy),
	})
	hr.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "gitops.kubesphere.io/v1alpha1",
			Kind:       "Application",
			Name:       appName,
			UID:        app.GetUID(),
		},
	})
	if err = r.Create(ctx, hr); err != nil {
		return
	}
	r.log.Info("Create FluxCD HelmRelease", "name", hr.GetAnnotations()["app.kubernetes.io/name"])
	r.recorder.Eventf(hr, corev1.EventTypeNormal, "Created", "Created FluxCD HelmRelease %s", hr.GetAnnotations()["app.kubernetes.io/name"])
	return
}

func (r *ApplicationReconciler) updateHelmRelease(ctx context.Context, hr *helmv2.HelmRelease, helmChart *sourcev1.HelmChart, deploy *v1alpha1.Deploy) (err error) {
	newHR := buildHelmRelease(helmChart, deploy)
	hr.Spec = newHR.Spec
	if err = r.Update(ctx, hr); err != nil {
		return
	}
	r.log.Info("Update FluxCD HelmRelease", "name", hr.GetAnnotations()["app.kubernetes.io/name"])
	r.recorder.Eventf(hr, corev1.EventTypeNormal, "Updated", "Update FluxCD HelmRelease %s", hr.GetAnnotations()["app.kubernetes.io/name"])
	return
}

func (r *ApplicationReconciler) reconcileKustomization(app *v1alpha1.Application) (err error) {
	ctx := context.Background()
	appNS, appName := app.GetNamespace(), app.GetName()

	kusList := &kusv1.KustomizationList{}
	if err = r.List(ctx, kusList, client.InNamespace(appNS), client.MatchingLabels{
		"app.kubernetes.io/managed-by": appName,
	}); err != nil {
		return
	}

	kusMap := make(map[string]*kusv1.Kustomization, len(kusList.Items))
	for _, kus := range kusList.Items {
		name := kus.GetAnnotations()["app.kubernetes.io/name"]
		kusMap[name] = kus.DeepCopy()
	}

	for _, kusDeploy := range app.Spec.FluxApp.Spec.Config.Kustomization {
		name := getKustomizationName(kusDeploy)

		if kus, ok := kusMap[name]; !ok {
			// not found
			// create
			err = r.createKustomization(ctx, app, kusDeploy)
		} else {
			// found
			// update this kus
			err = r.updateKustomization(ctx, app, kus, kusDeploy)
		}
	}
	return
}

func (r *ApplicationReconciler) createKustomization(ctx context.Context, app *v1alpha1.Application, deploy *v1alpha1.KustomizationSpec) (err error) {
	appNS, appName := app.GetNamespace(), app.GetName()
	kusNS := appNS

	kus := buildKustomization(app, deploy)
	kus.SetNamespace(kusNS)
	kus.SetGenerateName(appName)
	kus.SetName("")

	kus.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": appName,
	})
	kus.SetAnnotations(map[string]string{
		"app.kubernetes.io/name": getKustomizationName(deploy),
	})
	kus.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "gitops.kubesphere.io/v1alpha1",
			Kind:       "Application",
			Name:       appName,
			UID:        app.GetUID(),
		},
	})

	if err = r.Create(ctx, kus); err != nil {
		return
	}
	r.log.Info("Create FluxCD Kustomization", "name", kus.GetAnnotations()["app.kubernetes.io/name"])
	r.recorder.Eventf(kus, corev1.EventTypeNormal, "Created", "Created FluxCD Kustomization %s", kus.GetAnnotations()["app.kubernetes.io/name"])
	return
}
func (r *ApplicationReconciler) updateKustomization(ctx context.Context, app *v1alpha1.Application, kus *kusv1.Kustomization, deploy *v1alpha1.KustomizationSpec) (err error) {
	newKus := buildKustomization(app, deploy)
	kus.Spec = newKus.Spec
	if err = r.Update(ctx, kus); err != nil {
		return
	}
	r.log.Info("Update FluxCD HelmRelease", "name", kus.GetAnnotations()["app.kubernetes.io/name"])
	r.recorder.Eventf(kus, corev1.EventTypeNormal, "Updated", "Update FluxCD HelmRelease %s", kus.GetAnnotations()["app.kubernetes.io/name"])
	return
}

func isHelmOrKustomize(app *v1alpha1.Application) (AppType, error) {
	conf := app.Spec.FluxApp.Spec.Config.DeepCopy()
	helm, kus := conf.HelmRelease != nil, conf.Kustomization != nil
	if helm && !kus {
		return HelmRelease, nil
	}
	if !helm && kus {
		return Kustomization, nil
	}
	return "", fmt.Errorf("should has one and only one FluxApplicationConfig (HelmRelease or Kustomization)")
}

// GetGroupName returns the group name
func (r *ApplicationReconciler) GetGroupName() string {
	return controllerGroupName
}

// GetName returns the name of this reconciler
func (r *ApplicationReconciler) GetName() string {
	return "FluxApplicationReconciler"
}

func getHelmReleaseName(deploy *v1alpha1.Deploy) string {
	// host cluster
	if deploy.Destination.KubeConfig == nil {
		return deploy.Destination.TargetNamespace
	}
	// member cluster
	return deploy.Destination.KubeConfig.SecretRef.Name + "-" + deploy.Destination.TargetNamespace
}

func getKustomizationName(deploy *v1alpha1.KustomizationSpec) string {
	// host cluster
	if deploy.Destination.KubeConfig == nil {
		return deploy.Destination.TargetNamespace
	}
	// member cluster
	return deploy.Destination.KubeConfig.SecretRef.Name + "-" + deploy.Destination.TargetNamespace
}

func (r *ApplicationReconciler) saveTemplate(ctx context.Context, helmChart *sourcev1.HelmChart) (err error) {
	ns, name := helmChart.GetNamespace(), helmChart.GetName()

	if err = r.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, helmChart); err != nil {
		if !apierrors.IsNotFound(err) {
			return
		}
		// not found
		// create
		if err = r.Create(ctx, helmChart); err != nil {
			return
		}
		r.log.Info("Create HelmTemplate", "name", helmChart.GetName())
		r.recorder.Eventf(helmChart, corev1.EventTypeNormal, "Created", "Created HelmTemplate %s", helmChart.GetName())
	}
	return
}

func wantSaveHelmTemplate(app *v1alpha1.Application) bool {
	if v, ok := app.GetLabels()[v1alpha1.SaveTemplateLabelKey]; ok {
		return v == "true"
	}
	return false
}

func buildTemplateFromApp(app *v1alpha1.Application) *sourcev1.HelmChart {
	s := app.Spec.FluxApp.Spec.Source.SourceRef.DeepCopy()
	c := app.Spec.FluxApp.Spec.Config.HelmRelease.Chart.DeepCopy()

	return &sourcev1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.GetName(),
			Namespace: app.GetNamespace(),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": v1alpha1.GroupName,
			},
			Annotations: map[string]string{
				v1alpha1.HelmTemplateName: app.GetName(),
			},
		},
		Spec: sourcev1.HelmChartSpec{
			SourceRef: sourcev1.LocalHelmChartSourceReference{
				APIVersion: s.APIVersion,
				Kind:       s.Kind,
				Name:       s.Name,
			},
			Chart:             c.Chart,
			Version:           c.Version,
			Interval:          *c.Interval,
			ReconcileStrategy: c.ReconcileStrategy,
			ValuesFiles:       c.ValuesFiles,
		},
	}
}

func buildHelmRelease(helmChart *sourcev1.HelmChart, deploy *v1alpha1.Deploy) *helmv2.HelmRelease {
	ht := helmChart.DeepCopy()
	dp := deploy.DeepCopy()
	return &helmv2.HelmRelease{
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						APIVersion: ht.Spec.SourceRef.APIVersion,
						Kind:       ht.Spec.SourceRef.Kind,
						Name:       ht.Spec.SourceRef.Name,
					},
					Chart:             ht.Spec.Chart,
					Version:           ht.Spec.Version,
					Interval:          &ht.Spec.Interval,
					ReconcileStrategy: ht.Spec.ReconcileStrategy,
					ValuesFiles:       ht.Spec.ValuesFiles,
				},
			},
			KubeConfig:         dp.Destination.KubeConfig,
			TargetNamespace:    dp.Destination.TargetNamespace,
			Interval:           dp.Interval,
			Suspend:            dp.Suspend,
			ReleaseName:        dp.ReleaseName,
			StorageNamespace:   dp.StorageNamespace,
			DependsOn:          dp.DependsOn,
			Timeout:            dp.Timeout,
			MaxHistory:         dp.MaxHistory,
			ServiceAccountName: dp.ServiceAccountName,
			Install:            dp.Install,
			Upgrade:            dp.Upgrade,
			Test:               dp.Test,
			Rollback:           dp.Rollback,
			Uninstall:          dp.Uninstall,
			ValuesFrom:         dp.ValuesFrom,
			Values:             dp.Values,
			PostRenderers:      dp.PostRenderers,
		},
	}
}
func buildKustomization(app *v1alpha1.Application, deploy *v1alpha1.KustomizationSpec) *kusv1.Kustomization {
	s := app.Spec.FluxApp.Spec.Source.SourceRef
	dp := deploy.DeepCopy()

	return &kusv1.Kustomization{
		Spec: kusv1.KustomizationSpec{
			SourceRef: kusv1.CrossNamespaceSourceReference{
				APIVersion: s.APIVersion,
				Kind:       s.Kind,
				Name:       s.Name,
				Namespace:  s.Namespace,
			},
			DependsOn:          dp.DependsOn,
			Decryption:         dp.Decryption,
			Interval:           dp.Interval,
			RetryInterval:      dp.RetryInterval,
			KubeConfig:         convertKubeconfig(dp.Destination.KubeConfig),
			TargetNamespace:    dp.Destination.TargetNamespace,
			Path:               dp.Path,
			PostBuild:          dp.PostBuild,
			Prune:              dp.Prune,
			HealthChecks:       dp.HealthChecks,
			Patches:            dp.Patches,
			Images:             dp.Images,
			ServiceAccountName: dp.ServiceAccountName,
			Suspend:            dp.Suspend,
			Timeout:            dp.Timeout,
			Force:              dp.Force,
			Wait:               dp.Wait,
		},
	}
}

func convertKubeconfig(kubeconfig *helmv2.KubeConfig) *kusv1.KubeConfig {
	if kubeconfig == nil {
		return nil
	}
	return &kusv1.KubeConfig{SecretRef: kubeconfig.SecretRef}
}

// SetupWithManager setups the reconciler with a manager
func (r *ApplicationReconciler) SetupWithManager(mgr manager.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Application{}).
		Complete(r)
}
