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
	"encoding/base64"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//+kubebuilder:rbac:groups=cluster.kubesphere.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;create;update

// MultiClusterReconciler represents a controller to sync cluster to ArgoCD cluster
type MultiClusterReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile sync the cluster.kubesphere.io/clusters to kubeconfig secret
// in DevopsProject Namespaces where Flux Application(HelmRelease and Kustomization) exist.
func (r *MultiClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	r.log.Info(fmt.Sprintf("start to reconcile member cluster kubeconfig: %s", req.String()))

	cluster := createBareClusterObject()
	if err = r.Get(ctx, req.NamespacedName, cluster); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if err = r.reconcileCluster(ctx, cluster); err != nil {
		return
	}

	return
}

func (r *MultiClusterReconciler) reconcileCluster(ctx context.Context, cluster *unstructured.Unstructured) (err error) {
	name := cluster.GetName()
	nsList := &v1.NamespaceList{}
	if err = r.getFluxAppNsList(ctx, nsList); err != nil {
		return
	}

	for _, ns := range nsList.Items {
		secret := &v1.Secret{}
		if err = r.Get(ctx, types.NamespacedName{Namespace: ns.GetName(), Name: name}, secret); err != nil {
			if !apierrors.IsNotFound(err) {
				return
			}
			// not found kubeconfig
			// create
			newSecret := &v1.Secret{
				Type: "Opaque",
			}
			setKubeConfig(newSecret, cluster)
			newSecret.SetNamespace(ns.GetName())
			newSecret.SetName(name)
			newSecret.SetLabels(map[string]string{
				"app.kubernetes.io/managed-by": "fluxcd-controller",
			})
			newSecret.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: "cluster.kubesphere.io/v1alpha1",
					Kind:       "Cluster",
					Name:       cluster.GetName(),
					UID:        cluster.GetUID(),
				},
			})
			if err = r.Create(ctx, newSecret); err != nil {
				return
			}
			r.log.Info("Create member cluster kubeconfig", "name", newSecret.GetName())
			r.recorder.Eventf(newSecret, v1.EventTypeNormal, "Created", "Created Secret %s", newSecret.GetName())
		} else {
			// found kubeconfig
			// update
			newSecret := &v1.Secret{}
			setKubeConfig(newSecret, cluster)
			secret.StringData = newSecret.StringData
			if err = r.Update(ctx, secret); err != nil {
				return
			}
			r.log.Info("Update member cluster kubeconfig", "name", secret.GetName())
			r.recorder.Eventf(secret, v1.EventTypeNormal, "Updated", "Created Secret %s", secret.GetName())
		}
	}
	return
}

func setKubeConfig(secret *v1.Secret, cluster *unstructured.Unstructured) {
	kubeconfig, _, _ := unstructured.NestedString(cluster.Object, "spec", "connection", "kubeconfig")
	secret.StringData = map[string]string{
		// set the default key that fluxApp(HelmRelease and Kustomization) need.
		DefaultKubeConfigKey: base64DecodeWithoutErrorCheck(kubeconfig),
	}
}

func base64DecodeWithoutErrorCheck(str string) string {
	data, _ := base64.StdEncoding.DecodeString(str)
	return string(data)
}

// getFluxAppNs get the namespaces where the fluxApp (HelmRelease or Kustomization) existed.
// The secret must be in the same namespace as the fluxApp.
func (r *MultiClusterReconciler) getFluxAppNsList(ctx context.Context, nsList *v1.NamespaceList) (err error) {
	devops, _ := labels.NewRequirement("kubesphere.io/devopsproject", selection.Exists, nil)
	selector := labels.NewSelector().Add(*devops)
	if err = r.List(ctx, nsList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return
	}
	return
}

// GetName returns the name of this controller
func (r *MultiClusterReconciler) GetName() string {
	return "FluxCDMultiClusterController"
}

func createBareClusterObject() *unstructured.Unstructured {
	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.kubesphere.io",
		Version: "v1alpha1",
		Kind:    "Cluster",
	})
	return cluster
}

// SetupWithManager init the logger, recorder and filters
func (r *MultiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cluster := createBareClusterObject()
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		Named("fluxcd_multi_cluster_controller").
		For(cluster).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
