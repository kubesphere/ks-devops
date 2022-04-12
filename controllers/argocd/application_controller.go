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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
			err = r.Update(context.TODO(), app)
		}
		return
	}
	if k8sutil.AddFinalizer(&app.ObjectMeta, v1alpha1.ApplicationFinalizerName) {
		if err = r.Update(context.TODO(), app); err != nil {
			return
		}
	}

	if err = r.setArgoProject(app); err != nil {
		return
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
				err = r.Update(ctx, argoApp)
			}
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

	// there is a appProject in the same namespace
	if app.Spec.ArgoApp.Spec.Project != app.Namespace {
		app.Spec.ArgoApp.Spec.Project = app.Namespace // update the cache as well
		ctx := context.Background()
		latestApp := &v1alpha1.Application{}
		if err = r.Get(ctx, types.NamespacedName{
			Namespace: app.Namespace,
			Name:      app.Name,
		}, latestApp); err != nil {
			err = client.IgnoreNotFound(err)
		} else {
			latestApp.Spec.ArgoApp.Spec.Project = app.Namespace
			err = r.Update(ctx, latestApp)
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

// SetupWithManager setups the reconciler with a manager
// setup the logger, recorder
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		For(&v1alpha1.Application{}).
		Complete(r)
}
