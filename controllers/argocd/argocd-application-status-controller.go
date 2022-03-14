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
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=get;update
//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications/status,verbs=get;update
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list

// ApplicationStatusReconciler represents a controller to sync cluster to ArgoCD cluster
type ApplicationStatusReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile is the entrypoint of the controller
func (r *ApplicationStatusReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	var argoCDApp *unstructured.Unstructured
	if argoCDApp, err = getArgoCDApplication(r.Client, req.NamespacedName); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	ctx := context.Background()
	app := &v1alpha1.Application{}
	if err = r.Get(ctx, req.NamespacedName, app); err != nil {
		err = nil
		return
	}

	var status map[string]interface{}
	if status, _, err = unstructured.NestedMap(argoCDApp.Object, "status"); err == nil {
		var statusData []byte
		if statusData, err = json.Marshal(status); err == nil {
			app.Status.ArgoApp = string(statusData)
			err = r.Status().Update(ctx, app.DeepCopy())
		}
	}
	return
}

func getArgoCDApplication(client client.Reader, namespacedName types.NamespacedName) (app *unstructured.Unstructured, err error) {
	app = getArgoCDApplicationObject()

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

func getArgoCDApplicationObject() *unstructured.Unstructured {
	cluster := &unstructured.Unstructured{}
	cluster.SetKind("Application")
	cluster.SetAPIVersion("argoproj.io/v1alpha1")
	return cluster.DeepCopy()
}

// SetupWithManager init the logger, recorder and filters
func (r *ApplicationStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cluster := getArgoCDApplicationObject()
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		For(cluster).
		Complete(r)
}
