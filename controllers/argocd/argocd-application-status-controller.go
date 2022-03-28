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
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=get;update
//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications/status,verbs=get;update
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch

// ApplicationStatusReconciler represents a controller to sync cluster to ArgoCD cluster
type ApplicationStatusReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile is the entrypoint of the controller
func (r *ApplicationStatusReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	var argoCDApp *unstructured.Unstructured
	r.log.Info(fmt.Sprintf("start to reconcile ArgoCD application: %s", req.String()))

	if argoCDApp, err = getArgoCDApplication(r.Client, req.NamespacedName); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	appNs := argoCDApp.GetLabels()[v1alpha1.AppNamespaceLabelKey]
	appName := argoCDApp.GetLabels()[v1alpha1.AppNameLabelKey]
	if appName == "" || appNs == "" {
		return
	}

	ctx := context.Background()
	app := &v1alpha1.Application{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: appNs,
		Name:      appName,
	}, app); err != nil {
		r.log.Error(err, "cannot find application with namespace: %s, name: %s", appNs, appName)
		err = nil
		return
	}

	var status map[string]interface{}
	if status, _, err = unstructured.NestedMap(argoCDApp.Object, "status"); err == nil {
		var statusData []byte
		if statusData, err = json.Marshal(status); err == nil && !reflect.DeepEqual([]byte(app.Status.ArgoApp), statusData) {
			//app = app.DeepCopy()
			if app.GetLabels() == nil {
				// make sure the labels are not nil
				app.SetLabels(map[string]string{})
			}
			// set sync status into labels for filtering
			if syncStatus, found, _ := unstructured.NestedString(status, "sync", "status"); found {
				app.GetLabels()[v1alpha1.SyncStatusLabelKey] = syncStatus
			}
			// set health status into labels for filtering
			if healthStatus, found, _ := unstructured.NestedString(status, "health", "status"); found {
				app.GetLabels()[v1alpha1.HealthStatusLabelKey] = healthStatus
			}

			// unset operation field if it was absent
			if _, found, err := unstructured.NestedMap(argoCDApp.Object, "operation"); err != nil {
				return ctrl.Result{}, err
			} else if !found && app.Spec.ArgoApp != nil {
				app.Spec.ArgoApp.Operation = nil
			}

			// update labels
			if err = r.Update(ctx, app); err != nil {
				return
			}

			app.Status.ArgoApp = string(statusData)
			err = r.Status().Update(ctx, app)
		}
	}
	return
}

func getArgoCDApplication(client client.Reader, namespacedName types.NamespacedName) (app *unstructured.Unstructured, err error) {
	app = createBareArgoCDApplicationObject()

	if err = client.Get(context.Background(), namespacedName, app); err != nil {
		app = nil
	}
	return
}

// GetName returns the name of this controller
func (r *ApplicationStatusReconciler) GetName() string {
	return "ArgoCDApplicationStatusController"
}

// GetGroupName returns the group name of this controller
func (r *ApplicationStatusReconciler) GetGroupName() string {
	return controllerGroupName
}

func createBareArgoCDApplicationObject() *unstructured.Unstructured {
	argoApp := &unstructured.Unstructured{}
	argoApp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	})
	return argoApp
}

// SetupWithManager init the logger, recorder and filters
func (r *ApplicationStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	argoApp := createBareArgoCDApplicationObject()
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	var withLabelPredicate = predicate.NewPredicateFuncs(func(meta metav1.Object, object runtime.Object) (ok bool) {
		_, ok = meta.GetLabels()[v1alpha1.ArgoCDAppControlByLabelKey]
		return
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(argoApp).
		WithEventFilter(withLabelPredicate).
		Complete(r)
}
