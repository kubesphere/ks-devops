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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	helmv2 "kubesphere.io/devops/pkg/external/fluxcd/helm/v2beta1"
	kusv1 "kubesphere.io/devops/pkg/external/fluxcd/kustomize/v1beta2"
	apimeta "kubesphere.io/devops/pkg/external/fluxcd/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=get;update
//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications/status,verbs=get;update
//+kubebuilder:rbac:groups="kustomize.toolkit.fluxcd.io",resources=kustomizations,verbs=get;list;watch
//+kubebuilder:rbac:groups="helm.toolkit.fluxcd.io",resources=helmreleases,verbs=get;list;watch

// ApplicationStatusReconciler represents a controller to sync the status of
// FluxApplication (HelmRelease and Kustomization) to Kubesphere GitOps Application
type ApplicationStatusReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile is the entrypoint of the controller
func (r *ApplicationStatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	r.log.Info(fmt.Sprintf("start to reconcile application: %s", req.String()))
	hr := &helmv2.HelmRelease{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, hr); err == nil {
		return r.reconcileHelmRelease(ctx, hr)
	} else if !apierrors.IsNotFound(err) {
		return
	}

	kus := &kusv1.Kustomization{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, kus); err == nil {
		return r.reconcileKustomization(ctx, kus)
	} else if !apierrors.IsNotFound(err) {
		return
	}
	err = nil
	return
}

func (r *ApplicationStatusReconciler) reconcileHelmRelease(ctx context.Context, hr *helmv2.HelmRelease) (result ctrl.Result, err error) {
	appName, appNs := hr.GetLabels()["app.kubernetes.io/managed-by"], hr.GetNamespace()
	// Only reconcile Kubesphere managed HelmRelease not standalone HelmRelease
	if appName == "" {
		return
	}
	app := &v1alpha1.Application{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: appNs,
		Name:      appName,
	}, app); err != nil {
		return
	}

	readyHRNum, totalHRNum := 0, len(app.Spec.FluxApp.Spec.Config.HelmRelease.Deploy)
	if app.Status.FluxApp.HelmReleaseStatus == nil {
		app.Status.FluxApp.HelmReleaseStatus = make(map[string]*helmv2.HelmReleaseStatus, totalHRNum)
	}
	app.Status.FluxApp.HelmReleaseStatus[hr.GetAnnotations()["app.kubernetes.io/name"]] = hr.Status.DeepCopy()
	// Update status
	if err = r.Status().Update(ctx, app); err != nil {
		return
	}

	if app.GetLabels() == nil {
		app.SetLabels(map[string]string{})
	}
	if app.GetAnnotations() == nil {
		app.SetAnnotations(map[string]string{})
	}
	// should aggregate status from all the HelmRelease that FluxApp managed
	for _, status := range app.Status.FluxApp.HelmReleaseStatus {
		if meta.IsStatusConditionTrue(status.Conditions, apimeta.ReadyCondition) {
			readyHRNum++
		}
		if app.GetAnnotations()[FluxAppLastRevision] != status.LastAppliedRevision {
			app.GetAnnotations()[FluxAppLastRevision] = status.LastAppliedRevision
		}
	}
	app.GetLabels()[FluxAppReadyNumKey] = strconv.Itoa(readyHRNum) + "-" + strconv.Itoa(totalHRNum)
	// TODO: should find a better way to add AppType
	app.GetLabels()[FluxAppTypeKey] = string(HelmRelease)

	// update label
	if err = r.Update(ctx, app); err != nil {
		return
	}
	return
}

func (r *ApplicationStatusReconciler) reconcileKustomization(ctx context.Context, kus *kusv1.Kustomization) (result ctrl.Result, err error) {
	appName, appNs := kus.GetLabels()["app.kubernetes.io/managed-by"], kus.GetNamespace()
	// only reconcile Kubesphere managed Kustomization not standalone Kustomization
	app := &v1alpha1.Application{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: appNs,
		Name:      appName,
	}, app); err != nil {
		return
	}

	readyKusNum, totalKusNum := 0, len(app.Spec.FluxApp.Spec.Config.Kustomization)
	if app.Status.FluxApp.KustomizationStatus == nil {
		app.Status.FluxApp.KustomizationStatus = make(map[string]*kusv1.KustomizationStatus, totalKusNum)
	}
	app.Status.FluxApp.KustomizationStatus[kus.GetAnnotations()["app.kubernetes.io/name"]] = kus.Status.DeepCopy()
	// Update status
	if err = r.Status().Update(ctx, app); err != nil {
		return
	}

	if app.GetLabels() == nil {
		app.SetLabels(map[string]string{})
	}
	if app.GetAnnotations() == nil {
		app.SetAnnotations(map[string]string{})
	}
	// should aggregate status from all the Kustomization that FluxApp managed
	for _, status := range app.Status.FluxApp.KustomizationStatus {
		if meta.IsStatusConditionTrue(status.Conditions, apimeta.ReadyCondition) {
			readyKusNum++
		}
		if app.GetAnnotations()[FluxAppLastRevision] != status.LastAppliedRevision {
			app.GetAnnotations()[FluxAppLastRevision] = status.LastAppliedRevision
		}
	}
	app.GetLabels()[FluxAppReadyNumKey] = strconv.Itoa(readyKusNum) + "-" + strconv.Itoa(totalKusNum)
	// TODO: should find a better way to add AppType
	app.GetLabels()[FluxAppTypeKey] = string(Kustomization)
	// update label
	if err = r.Update(ctx, app); err != nil {
		return
	}
	return
}

// GetName returns the name of this controller
func (r *ApplicationStatusReconciler) GetName() string {
	return "FluxCDApplicationStatusController"
}

// SetupWithManager init the logger, recorder and filters
func (r *ApplicationStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		Watches(&source.Kind{Type: &kusv1.Kustomization{}}, &handler.EnqueueRequestForObject{}).
		For(&helmv2.HelmRelease{}).
		Complete(r)
}
