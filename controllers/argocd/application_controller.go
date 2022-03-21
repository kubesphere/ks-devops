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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=get;list;update
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;create;update

// ApplicationReconciler is the reconciler of the Application
type ApplicationReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile makes sure the Application and ArgoCD application works well
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

	argoApp := &unstructured.Unstructured{}
	argoApp.SetKind("Application")
	argoApp.SetAPIVersion("argoproj.io/v1alpha1")

	if err = r.Get(ctx, types.NamespacedName{
		Namespace: app.GetNamespace(),
		Name:      app.GetName(),
	}, argoApp); err != nil {
		if !apierrors.IsNotFound(err) {
			return
		}

		if argoApp, err = createUnstructuredApplication(app); err != nil {
			return
		}

		err = r.Create(ctx, argoApp)
	} else {
		var newArgoApp *unstructured.Unstructured
		if newArgoApp, err = createUnstructuredApplication(app); err == nil {
			argoApp.Object["spec"] = newArgoApp.Object["spec"]
			k8sutil.AddOwnerReference(argoApp, app.TypeMeta, app.ObjectMeta)
			err = r.Update(ctx, argoApp)
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

	newArgoApp := &unstructured.Unstructured{}
	newArgoApp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
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

	// set owner reference
	k8sutil.AddOwnerReference(newArgoApp, app.TypeMeta, app.ObjectMeta)

	return newArgoApp, nil
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
		For(&v1alpha1.Application{}).
		Complete(r)
}
