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
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=get;list;update
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// ApplicationReconciler is the reconciler of the Application
type ApplicationReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile makes sure the Application and ArgoCD application works well
// Consider all the ArgoCD application need to be in one particular namespace. But the Application from this project can be in different namespaces.
// In order to avoid the naming conflict, we will check if the name existing the target namespace. Take the original name
// as the generatedName of the ArgoCD application if there is a potential conflict. In the most cases, we can keep the original name
// same to the ArgoCD Application name.
func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	r.log.Info(fmt.Sprintf("start to reconcile application: %s", req.String()))

	app := &v1alpha1.Application{}
	if err = r.Client.Get(ctx, req.NamespacedName, app); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if argo := app.Spec.ArgoApp; argo != nil {
		err = r.reconcileArgoApplication(app)
	}
	return
}

func (r *ApplicationReconciler) reconcileArgoApplication(app *v1alpha1.Application) (err error) {
	ctx := context.Background()

	argoApp := createBareArgoCDApplicationObject()
	argoCDNamespace, ok := app.Labels[v1alpha1.ArgoCDLocationLabelKey]
	if !ok || "" == argoCDNamespace {
		r.recorder.Eventf(app, corev1.EventTypeWarning, "Invalid",
			"Cannot find the namespace of the Argo CD instance from key: %s", v1alpha1.ArgoCDLocationLabelKey)
		return
	}
	argoCDAppName, hasArgoName := app.Labels[v1alpha1.ArgoCDAppNameLabelKey]
	if argoCDAppName == "" {
		argoCDAppName = app.GetName()
	}

	// the application was deleted
	if !app.ObjectMeta.DeletionTimestamp.IsZero() {
		if err = r.Get(ctx, types.NamespacedName{
			Namespace: argoCDNamespace,
			Name:      argoCDAppName,
		}, argoApp); err != nil {
			if !apierrors.IsNotFound(err) {
				return
			}
			err = nil
		} else {
			err = r.Delete(ctx, argoApp)
		}

		if err == nil {
			k8sutil.RemoveFinalizer(&app.ObjectMeta, v1alpha1.ApplicationFinalizerName)
			k8sutil.RemoveFinalizer(&app.ObjectMeta, v1alpha1.ArgoCDResourcesFinalizer)
			err = r.Update(context.TODO(), app)
		}
		return
	}

	if app.Spec.ArgoApp.Spec.Project != app.Namespace {
		app.Spec.ArgoApp.Spec.Project = app.Namespace // update the cache as well
	}

	if err = r.Get(ctx, types.NamespacedName{
		Namespace: argoCDNamespace,
		Name:      argoCDAppName,
	}, argoApp); err != nil {
		if !apierrors.IsNotFound(err) {
			return
		}

		if argoApp, err = createUnstructuredApplication(app); err != nil {
			return
		}

		argoApp.SetName(app.Name)
		argoApp.SetNamespace(argoCDNamespace)
		if err = r.Create(ctx, argoApp); err == nil {
			err = r.addArgoAppNameIntoLabels(app.GetNamespace(), app.GetName(), argoApp.GetName())
		} else {
			data, _ := argoApp.MarshalJSON()
			r.log.Error(err, "failed to create application to ArgoCD", "data", string(data))
			r.recorder.Eventf(app, corev1.EventTypeWarning, "FailedWithArgoCD",
				"failed to create application to ArgoCD, error is: %v", err)
		}
	} else {
		if !hasArgoName {
			// naming conflict happened
			if argoApp, err = createUnstructuredApplication(app); err != nil {
				return
			}
			argoApp.SetGenerateName(app.GetName())
			argoApp.SetName("")
			argoApp.SetNamespace(argoCDNamespace)

			if err = r.Create(ctx, argoApp); err == nil {
				err = r.addArgoAppNameIntoLabels(app.GetNamespace(), app.GetName(), argoApp.GetName())
			} else {
				data, _ := argoApp.MarshalJSON()
				r.log.Error(err, "failed to create application to ArgoCD", "data", string(data))
				r.recorder.Eventf(app, corev1.EventTypeWarning, "FailedWithArgoCD",
					"failed to create application to ArgoCD, error is: %v", err)
			}
		} else {
			var newArgoApp *unstructured.Unstructured
			if newArgoApp, err = createUnstructuredApplication(app); err == nil {
				argoApp.Object["spec"] = newArgoApp.Object["spec"]
				argoApp.Object["operation"] = newArgoApp.Object["operation"]

				// append annotations and labels
				copyArgoAnnotationsAndLabels(newArgoApp, argoApp)

				argoApp.SetFinalizers(newArgoApp.GetFinalizers())
				err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
					latestArgoApp := createBareArgoCDApplicationObject()
					if err = r.Get(context.Background(), types.NamespacedName{
						Namespace: argoApp.GetNamespace(),
						Name:      argoApp.GetName(),
					}, latestArgoApp); err != nil {
						return
					}

					argoApp.SetResourceVersion(latestArgoApp.GetResourceVersion())
					err = r.Update(ctx, argoApp)
					return
				})
			}
		}
	}

	if err == nil {
		if err = r.setArgoProject(app); err != nil {
			return
		}
	}
	return
}

func createUnstructuredApplication(app *v1alpha1.Application) (result *unstructured.Unstructured, err error) {
	if app.Spec.ArgoApp == nil {
		err = fmt.Errorf("no argo found from the spec")
		return
	}
	argoApp := app.Spec.ArgoApp.DeepCopy()
	// TODO set some default values
	if argoApp.Spec.Project == "" {
		argoApp.Spec.Project = "default"
	}

	// application destination can't have both name and server defined
	if argoApp.Spec.Destination.Name != "" && argoApp.Spec.Destination.Server != "" {
		argoApp.Spec.Destination.Server = ""
	}

	newArgoApp := createBareArgoCDApplicationObject()
	newArgoApp.SetLabels(map[string]string{
		v1alpha1.ArgoCDAppControlByLabelKey: "ks-devops",
		v1alpha1.AppNamespaceLabelKey:       app.Namespace,
		v1alpha1.AppNameLabelKey:            app.Name,
	})
	newArgoApp.SetName(app.GetName())
	newArgoApp.SetNamespace(app.GetNamespace())

	// make sure all Argo CD supported annotations and labels exist
	copyArgoAnnotationsAndLabels(app, newArgoApp)

	// copy all potential finalizers
	finalizers := app.GetFinalizers()
	targetFinalizers := make([]string, 0)
	for i := range finalizers {
		finalizer := finalizers[i]
		if strings.HasSuffix(finalizer, "argocd.argoproj.io") {
			targetFinalizers = append(targetFinalizers, finalizer)
		}
	}
	newArgoApp.SetFinalizers(targetFinalizers)

	if err := SetNestedField(newArgoApp.Object, argoApp.Spec, "spec"); err != nil {
		return nil, err
	}

	if argoApp.Operation != nil {
		if err := SetNestedField(newArgoApp.Object, argoApp.Operation, "operation"); err != nil {
			return nil, err
		}
	}
	return newArgoApp, nil
}

func copyArgoAnnotationsAndLabels(app metav1.Object, argoApp metav1.Object) {
	var annotations map[string]string
	if annotations = argoApp.GetAnnotations(); annotations == nil {
		annotations = map[string]string{}
	}
	for k, v := range app.GetAnnotations() {
		if strings.Contains(k, "argoproj.io") {
			annotations[k] = v
		}
	}
	argoApp.SetAnnotations(annotations)

	var labels map[string]string
	if labels = argoApp.GetLabels(); labels == nil {
		labels = map[string]string{}
	}
	for k, v := range app.GetLabels() {
		if strings.Contains(k, "argoproj.io") {
			labels[k] = v
		}
	}
	argoApp.SetLabels(labels)
}

func (r *ApplicationReconciler) addArgoAppNameIntoLabels(namespace, name, argoAppName string) (err error) {
	app := &v1alpha1.Application{}
	ctx := context.Background()
	if err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, app); err == nil {
		if app.Labels == nil {
			app.Labels = make(map[string]string)
		}
		app.Labels[v1alpha1.ArgoCDAppNameLabelKey] = argoAppName
		err = r.Update(ctx, app)
	}
	return
}

// TODO it might be better to do it as a mutating hook
func (r *ApplicationReconciler) setArgoProject(app *v1alpha1.Application) (err error) {
	if app.Spec.ArgoApp == nil {
		return
	}

	latestApp := &v1alpha1.Application{}
	ctx := context.Background()
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, latestApp); err != nil {
		return client.IgnoreNotFound(err)
	}
	err = r.Update(ctx, latestApp)

	// there is a appProject in the same namespace
	needUpdate := false
	if latestApp.Spec.ArgoApp.Spec.Project != latestApp.Namespace {
		latestApp.Spec.ArgoApp.Spec.Project = latestApp.Namespace
		needUpdate = true
	}

	if k8sutil.AddFinalizer(&latestApp.ObjectMeta, v1alpha1.ApplicationFinalizerName) || needUpdate {
		if err = r.Update(context.TODO(), latestApp); err != nil {
			return
		}
	}
	return
}

// GetName returns the name of this reconciler
func (r *ApplicationReconciler) GetName() string {
	return "ApplicationController"
}

// GetGroupName returns the group name of this reconciler
func (r *ApplicationReconciler) GetGroupName() string {
	return controllerGroupName
}

type finalizersChangedPredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating finalizers change
func (finalizersChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.MetaOld == nil {
		return false
	}
	if e.ObjectOld == nil {
		return false
	}
	if e.ObjectNew == nil {
		return false
	}
	if e.MetaNew == nil {
		return false
	}
	return !reflect.DeepEqual(e.MetaNew.GetFinalizers(), e.MetaOld.GetFinalizers())
}

type specificAnnotationsOrLabelsChangedPredicate struct {
	predicate.Funcs
	filter string
}

func (p specificAnnotationsOrLabelsChangedPredicate) Update(e event.UpdateEvent) (changed bool) {
	changed = !reflect.DeepEqual(e.MetaNew.GetAnnotations(), e.MetaOld.GetAnnotations())
	if !changed {
		if changed = !reflect.DeepEqual(e.MetaNew.GetLabels(), e.MetaOld.GetLabels()); changed {
			changed = mapKeysContains(p.filter, e.MetaNew.GetLabels(), e.MetaOld.GetLabels())
		}
	} else {
		changed = mapKeysContains(p.filter, e.MetaNew.GetAnnotations(), e.MetaOld.GetAnnotations())
	}
	return
}

func mapKeysContains(filter string, annotations ...map[string]string) (has bool) {
	for _, anno := range annotations {
		for k, _ := range anno {
			if has = strings.Contains(k, filter); has {
				return
			}
		}
	}
	return
}

// SetupWithManager setups the reconciler with a manager
// setup the logger, recorder
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			finalizersChangedPredicate{},
			specificAnnotationsOrLabelsChangedPredicate{filter: "argoproj.io"})).
		For(&v1alpha1.Application{}).
		Complete(r)
}
