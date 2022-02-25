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
	"encoding/base64"
	"encoding/json"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/utils/k8sutil"
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

	argocdNamespace string
}

// Reconcile is the entrypoint of the controller
func (r *MultiClusterReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	var cluster *unstructured.Unstructured
	if cluster, err = getCluster(r.Client, req.NamespacedName); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// find the namespace of ArgoCD, and cache it
	if r.argocdNamespace == "" {
		if r.argocdNamespace = r.findArgoCDNamespace(); r.argocdNamespace == "" {
			r.log.V(6).Info("cannot found the namespace of ArgoCD")
			return
		}
	}

	// ignore the host cluster
	if ignore(cluster) {
		return
	}

	ownerRef := getOwnerReference(cluster)
	argoCluster := createArgoCluster(cluster)
	argoCluster.SetNamespace(r.argocdNamespace)
	err = r.updateOrCreate(argoCluster, ownerRef)
	return
}

func getOwnerReference(object *unstructured.Unstructured) (ref metav1.OwnerReference) {
	ref = metav1.OwnerReference{
		APIVersion: object.GetAPIVersion(),
		Kind:       object.GetKind(),
		Name:       object.GetName(),
		UID:        object.GetUID(),
	}
	return
}

func (r *MultiClusterReconciler) updateOrCreate(cluster *v1.Secret, ownerRef metav1.OwnerReference) (err error) {
	ctx := context.Background()
	existingArgoCluster := cluster.DeepCopy()
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}, existingArgoCluster); err != nil {
		if apierrors.IsNotFound(err) {
			// create the argo cluster
			k8sutil.SetOwnerReference(cluster, ownerRef)
			err = r.Create(ctx, cluster)
		}
		return
	}
	existingArgoCluster.Data = cluster.Data
	k8sutil.SetOwnerReference(existingArgoCluster, ownerRef)
	err = r.Update(ctx, existingArgoCluster)
	return
}

// let's assume there is only one ArgoCD existing
// TODO consider to have a better solution to figure the namespace of ArgoCD
func (r *MultiClusterReconciler) findArgoCDNamespace() string {
	secretList := &v1.SecretList{}
	if err := r.List(context.Background(), secretList); err == nil {
		for i := range secretList.Items {
			secret := secretList.Items[i]
			if secret.Name == "argocd-secret" {
				return secret.Namespace
			}
		}
	}
	return ""
}

// createArgoCluster creates an object which represents the argo cluster\
// see also https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#clusters
func createArgoCluster(cluster *unstructured.Unstructured) (secret *v1.Secret) {
	name := cluster.GetName()
	secret = &v1.Secret{
		Type: "Opaque",
	}
	secret.SetName(name)
	secret.SetLabels(map[string]string{
		"argocd.argoproj.io/secret-type": "cluster",
	})

	config, _, _ := unstructured.NestedString(cluster.Object, "spec", "connection", "kubeconfig")
	server, _, _ := unstructured.NestedString(cluster.Object, "spec", "connection", "kubernetesAPIEndpoint")

	secret.Data = map[string][]byte{
		"name":   []byte(name),
		"config": getArgoClusterConfigData(config),
		"server": []byte(server),
	}
	return
}

// getArgoClusterConfigFormat returns a ArgoCD cluster format config from a KubeSpher cluster config format
// the parameter `config` is base64 encoded
func getArgoClusterConfigData(config string) (result []byte) {
	clusterConfigObject := getArgoClusterConfigObject(config)
	result, _ = json.Marshal(clusterConfigObject)
	return
}

// only support one config
func getArgoClusterConfigObject(configText string) (object *ClusterConfig) {
	rawDecodedconfig, _ := base64.StdEncoding.DecodeString(configText)

	var kubeconfig *config
	if kubeconfig = parseKubeConfig(rawDecodedconfig); kubeconfig == nil {
		return
	}

	if len(kubeconfig.Clusters) == 0 || len(kubeconfig.Users) == 0 {
		return
	}

	cluster := kubeconfig.Clusters[0]
	user := kubeconfig.Users[0]

	object = &ClusterConfig{
		BearerToken: user.Auth.Token,
		TLSClientConfig: TLSClientConfig{
			Insecure: cluster.Cluster.SkipTLS,
			CAData:   []byte(cluster.Cluster.CA),
			CertData: []byte(user.Auth.ClientCert),
			KeyData:  []byte(user.Auth.ClientKey),
		},
	}
	return
}

func getCluster(client client.Reader, namespacedName types.NamespacedName) (cluster *unstructured.Unstructured, err error) {
	cluster = getClusterObject()

	if err = client.Get(context.Background(), namespacedName, cluster); err != nil {
		cluster = nil
	}
	return
}

func ignore(cluster *unstructured.Unstructured) bool {
	if cluster != nil {
		_, ok := cluster.GetLabels()["cluster-role.kubesphere.io/host"]
		return ok
	}
	return true
}

// GetName returns the name of this controller
func (r *MultiClusterReconciler) GetName() string {
	return "ArgoCDMultiClusterController"
}

// GetGroupName returns the group name of this controller
func (r *MultiClusterReconciler) GetGroupName() string {
	return controllerGroupName
}

func getClusterObject() *unstructured.Unstructured {
	cluster := &unstructured.Unstructured{}
	cluster.SetKind("Cluster")
	cluster.SetAPIVersion("cluster.kubesphere.io/v1alpha1")
	return cluster.DeepCopy()
}

// SetupWithManager init the logger, recorder and filters
func (r *MultiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cluster := getClusterObject()
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		For(cluster).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
