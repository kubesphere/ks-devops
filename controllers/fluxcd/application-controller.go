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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/controllers/argocd"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=watch;get;list
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="kustomize.toolkit.fluxcd.io",resources=kustomizations,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups="helm.toolkit.fluxcd.io",resources=helmreleases,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups="source.toolkit.fluxcd.io/v1beta2",resources=helmcharts,verbs=get;list

// ApplicationReconciler is the reconciler of the FluxCD HelmRelease and FluxCD Kustomization
type ApplicationReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile sync the Application with underlying FluxCD HelmRelease and FluxCD Kustomization CRD
func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
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

	if wantSaveHelmTemplate(app) && app.Spec.FluxApp.Spec.Config.HelmRelease.Chart != nil {
		if err = r.saveTemplate(ctx, app); err != nil {
			return
		}
	}

	var helmChart *unstructured.Unstructured
	if helmTemplateName := app.Spec.FluxApp.Spec.Config.HelmRelease.Template; helmTemplateName != "" {
		// use template
		helmTemplateNS := app.GetNamespace()
		helmTemplateList := createBareFluxHelmTemplateObjectList()

		if err = r.List(ctx, helmTemplateList, client.InNamespace(helmTemplateNS), client.MatchingLabels{
			v1alpha1.HelmTemplateName: helmTemplateName,
		}); err != nil {
			return
		}
		// there is a helmtemplate that user specified
		helmChart = &helmTemplateList.Items[0]
	} else {
		helmChart = createUnstructuredFluxHelmTemplate(app)
	}

	if err = r.reconcileHelmReleaseList(app, helmChart); err != nil {
		return
	}

	return
}

func (r *ApplicationReconciler) reconcileHelmReleaseList(app *v1alpha1.Application, helmChart *unstructured.Unstructured) (err error) {
	ctx := context.Background()

	fluxApp := app.Spec.FluxApp.DeepCopy()
	appNS, appName := app.GetNamespace(), app.GetName()

	fluxHelmReleaseList := createBareFluxHelmReleaseListObject()
	if err = r.List(ctx, fluxHelmReleaseList, client.InNamespace(appNS), client.MatchingLabels{
		"app.kubernetes.io/managed-by": getHelmTemplateName(appNS, appName),
	}); err != nil {
		return
	}

	hrMap := make(map[string]*unstructured.Unstructured, len(fluxHelmReleaseList.Items))
	for _, fluxHelmRelease := range fluxHelmReleaseList.Items {
		hrMap[fluxHelmRelease.GetName()] = fluxHelmRelease.DeepCopy()
	}

	helmReleaseNum := len(fluxApp.Spec.Config.HelmRelease.Deploy)
	for i := 0; i < helmReleaseNum; i++ {
		deploy := fluxApp.Spec.Config.HelmRelease.Deploy[i]
		helmReleaseName := getHelmReleaseName(deploy.Destination.TargetNamespace)
		if hr, ok := hrMap[helmReleaseName]; !ok {
			// there is no matching helmRelease
			// create
			if err = r.createHelmRelease(ctx, app, helmChart, deploy); err != nil {
				return
			}
		} else {
			// there is a matching helmRelease
			// update the helmRelease
			if err = r.updateHelmRelease(ctx, hr, helmChart, deploy); err != nil {
				return
			}
		}
	}
	return
}

func (r *ApplicationReconciler) createHelmRelease(ctx context.Context, app *v1alpha1.Application, helmChart *unstructured.Unstructured, deploy *v1alpha1.Deploy) (err error) {
	appNS, appName := app.GetNamespace(), app.GetName()
	hrNS, hrName := appNS, getHelmReleaseName(deploy.Destination.TargetNamespace)

	newFluxHelmRelease := createBareFluxHelmReleaseObject()
	setFluxHelmReleaseFields(newFluxHelmRelease, helmChart, deploy)
	newFluxHelmRelease.SetNamespace(hrNS)
	newFluxHelmRelease.SetName(hrName)
	newFluxHelmRelease.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": getHelmTemplateName(appNS, appName),
	})
	newFluxHelmRelease.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "gitops.kubesphere.io/v1alpha1",
			Kind:       "Application",
			Name:       appName,
			UID:        app.GetUID(),
		},
	})
	if err = r.Create(ctx, newFluxHelmRelease); err != nil {
		return
	}
	r.log.Info("Create FluxCD HelmRelease", "name", newFluxHelmRelease.GetName())
	r.recorder.Eventf(app, corev1.EventTypeNormal, "Created", "Created HelmRelease %s", newFluxHelmRelease.GetName())
	return
}

func (r *ApplicationReconciler) updateHelmRelease(ctx context.Context, hr, helmChart *unstructured.Unstructured, deploy *v1alpha1.Deploy) (err error) {
	setFluxHelmReleaseFields(hr, helmChart, deploy)
	if err = r.Update(ctx, hr); err != nil {
		return
	}
	return
}

func setFluxHelmReleaseFields(hr, helmTemplate *unstructured.Unstructured, deploy *v1alpha1.Deploy) {
	helmTemplateSpec, _, _ := unstructured.NestedFieldNoCopy(helmTemplate.Object, "spec")
	_ = argocd.SetNestedField(hr.Object, helmTemplateSpec, "spec", "chart", "spec")

	configs, _ := argocd.InterfaceToMap(*deploy)
	for k, config := range configs {
		if k != "destination" {
			_ = unstructured.SetNestedField(hr.Object, config, "spec", k)
		}
	}

	_ = argocd.SetNestedField(hr.Object, deploy.Destination.KubeConfig, "spec", "kubeConfig")
	_ = unstructured.SetNestedField(hr.Object, deploy.Destination.TargetNamespace, "spec", "targetNamespace")
}

func (r *ApplicationReconciler) reconcileKustomization(app *v1alpha1.Application) (err error) {
	ctx := context.Background()

	fluxApp := app.Spec.FluxApp.DeepCopy()
	appNS, appName := app.GetNamespace(), app.GetName()

	kusList := createBareFluxKustomizationListObject()
	if err = r.List(ctx, kusList, client.InNamespace(appNS), client.MatchingLabels{
		"app.kubernetes.io/managed-by": getKusGroupName(appNS, appName),
	}); err != nil {
		return
	}

	kusMap := make(map[string]*unstructured.Unstructured, len(kusList.Items))
	for _, kus := range kusList.Items {
		kusMap[kus.GetName()] = kus.DeepCopy()
	}

	kusNum := len(app.Spec.FluxApp.Spec.Config.Kustomization)
	for i := 0; i < kusNum; i++ {
		kusDeploy := fluxApp.Spec.Config.Kustomization[i]
		kusName := getKustomizationName(kusDeploy.Destination.TargetNamespace)

		if kus, ok := kusMap[kusName]; !ok {
			// not found
			// create
			err = r.createKustomization(ctx, app, kusDeploy)
		} else {
			// found
			// update this kus
			err = r.updateKustomization(ctx, kus, app, kusDeploy)
		}
	}
	return
}

func (r *ApplicationReconciler) createKustomization(ctx context.Context, app *v1alpha1.Application, deploy *v1alpha1.KustomizationSpec) (err error) {
	appNS, appName := app.GetNamespace(), app.GetName()
	kusNS, kusName := appNS, getKustomizationName(deploy.Destination.TargetNamespace)

	kus := createBareFluxKustomizationObject()
	setFluxKustomizationFields(kus, app, deploy)
	kus.SetNamespace(kusNS)
	kus.SetName(kusName)
	kus.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "gitops.kubesphere.io/v1alpha1",
			Kind:       "Application",
			Name:       appName,
			UID:        app.GetUID(),
		},
	})
	kus.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": getKusGroupName(appNS, appName),
	})
	err = r.Create(ctx, kus)
	return
}

func (r *ApplicationReconciler) updateKustomization(ctx context.Context, kus *unstructured.Unstructured, app *v1alpha1.Application, deploy *v1alpha1.KustomizationSpec) (err error) {
	setFluxKustomizationFields(kus, app, deploy)
	err = r.Update(ctx, kus)
	return
}

func setFluxKustomizationFields(kus *unstructured.Unstructured, app *v1alpha1.Application, deploy *v1alpha1.KustomizationSpec) {
	_ = argocd.SetNestedField(kus.Object, app.Spec.FluxApp.Spec.Source.SourceRef, "spec", "sourceRef")

	configs, _ := argocd.InterfaceToMap(*deploy)
	for k, config := range configs {
		if k != "destination" {
			_ = unstructured.SetNestedField(kus.Object, config, "spec", k)
		}
	}
	_ = argocd.SetNestedField(kus.Object, deploy.Destination.KubeConfig, "spec", "kubeConfig")
	_ = unstructured.SetNestedField(kus.Object, deploy.Destination.TargetNamespace, "spec", "targetNamespace")
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

// TODO: it will be better to use hash name for naming conflict reason
func getHelmReleaseName(targetNamespace string) string {
	return targetNamespace
}

func getKustomizationName(targetNamespace string) string {
	return targetNamespace
}

func getHelmTemplateName(ns, name string) string {
	return ns + "-" + name
}

func getKusGroupName(ns, name string) string {
	return ns + "-" + name
}

func (r *ApplicationReconciler) saveTemplate(ctx context.Context, app *v1alpha1.Application) (err error) {
	helmTemplate := createBareFluxHelmTemplateObject()

	appNS, appName := app.GetNamespace(), app.GetName()
	helmTemplateNS, helmTemplateName := appNS, getHelmTemplateName(appNS, appName)

	if err = r.Get(ctx, types.NamespacedName{Namespace: helmTemplateNS, Name: helmTemplateName}, helmTemplate); err != nil {
		if !apierrors.IsNotFound(err) {
			return
		}
		// not found
		// create
		newFluxHelmTemplate := createUnstructuredFluxHelmTemplate(app)
		newFluxHelmTemplate.SetName(helmTemplateName)
		newFluxHelmTemplate.SetNamespace(helmTemplateNS)
		newFluxHelmTemplate.SetLabels(map[string]string{
			v1alpha1.HelmTemplateName: helmTemplateName,
		})
		if err = r.Create(ctx, newFluxHelmTemplate); err != nil {
			return
		}
		r.log.Info("Create HelmTemplate", "name", newFluxHelmTemplate.GetName())
		r.recorder.Eventf(app, corev1.EventTypeNormal, "Created", "Created HelmTemplate %s", newFluxHelmTemplate.GetName())
	}
	return
}

func wantSaveHelmTemplate(app *v1alpha1.Application) bool {
	if v, ok := app.GetLabels()[v1alpha1.SaveTemplateLabelKey]; ok {
		return v == "true"
	}
	return false
}

func createUnstructuredFluxHelmTemplate(app *v1alpha1.Application) (newFluxHelmTemplate *unstructured.Unstructured) {
	newFluxHelmTemplate = createBareFluxHelmTemplateObject()
	newFluxHelmTemplate.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": v1alpha1.GroupName,
	})

	_ = argocd.SetNestedField(newFluxHelmTemplate.Object, app.Spec.FluxApp.Spec.Config.HelmRelease.Chart, "spec")
	_ = argocd.SetNestedField(newFluxHelmTemplate.Object, app.Spec.FluxApp.Spec.Source.SourceRef, "spec", "sourceRef")
	return
}

func createBareFluxHelmTemplateObject() *unstructured.Unstructured {
	fluxHelmChart := &unstructured.Unstructured{}
	fluxHelmChart.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "source.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "HelmChart",
	})
	return fluxHelmChart
}

func createBareFluxHelmTemplateObjectList() *unstructured.UnstructuredList {
	fluxHelmTemplateList := &unstructured.UnstructuredList{}
	fluxHelmTemplateList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "source.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "HelmChartList",
	})
	return fluxHelmTemplateList
}

func createBareFluxHelmReleaseObject() *unstructured.Unstructured {
	fluxHelmRelease := &unstructured.Unstructured{}
	fluxHelmRelease.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "helm.toolkit.fluxcd.io",
		Version: "v2beta1",
		Kind:    "HelmRelease",
	})
	return fluxHelmRelease
}

func createBareFluxHelmReleaseListObject() *unstructured.UnstructuredList {
	fluxHelmReleaseList := &unstructured.UnstructuredList{}
	fluxHelmReleaseList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "helm.toolkit.fluxcd.io",
		Version: "v2beta1",
		Kind:    "HelmReleaseList",
	})
	return fluxHelmReleaseList
}

func createBareFluxKustomizationObject() *unstructured.Unstructured {
	fluxKustomization := &unstructured.Unstructured{}
	fluxKustomization.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kustomize.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "Kustomization",
	})
	return fluxKustomization
}

func createBareFluxKustomizationListObject() *unstructured.UnstructuredList {
	fluxHelmReleaseList := &unstructured.UnstructuredList{}
	fluxHelmReleaseList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kustomize.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "KustomizationList",
	})
	return fluxHelmReleaseList
}

// SetupWithManager setups the reconciler with a manager
// setup the logger, recorder
func (r *ApplicationReconciler) SetupWithManager(mgr manager.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Application{}).
		Complete(r)
}
