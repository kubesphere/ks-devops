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
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/api/gitops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/utils/k8sutil"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=gitrepositories,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="source.toolkit.fluxcd.io",resources=gitrepositories,verbs=get;list;create;update;delete

// GitRepositoryReconciler is the reconciler of the FluxCDGitRepository
type GitRepositoryReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile maintains the FluxCDGitRepository against to the GitRepository
func (r *GitRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	repo := &v1alpha3.GitRepository{}

	if err = r.Get(ctx, req.NamespacedName, repo); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	err = r.reconcileFluxGitRepo(repo)
	return
}

func (r *GitRepositoryReconciler) reconcileFluxGitRepo(repo *v1alpha3.GitRepository) (err error) {
	ctx := context.Background()
	fluxGitRepo := createBareFluxGitRepoObject()
	// FluxGitRepo's namespace = v1alpha3.GitRepository's namespace
	// FluxGitRepo's name = "fluxcd-" + v1alpha3.GitRepository's name
	ns, name := repo.GetNamespace(), getFluxRepoName(repo.GetName())

	if !isArtifactRepo(repo) || !repo.ObjectMeta.DeletionTimestamp.IsZero() {
		if err = r.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, fluxGitRepo); err != nil {
			err = client.IgnoreNotFound(err)
		} else {
			r.log.Info("delete FluxCDGitRepository", "name", fluxGitRepo.GetName())
			err = r.Delete(ctx, fluxGitRepo)
		}

		if err == nil {
			k8sutil.RemoveFinalizer(&repo.ObjectMeta, v1alpha3.GitRepoFinalizerName)
			err = r.Update(ctx, repo)
		}
		return
	}
	if k8sutil.AddFinalizer(&repo.ObjectMeta, v1alpha3.GitRepoFinalizerName) {
		err = r.Update(ctx, repo)
		if err != nil {
			return
		}
	}

	if err = r.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, fluxGitRepo); err != nil {
		if !apierrors.IsNotFound(err) {
			return
		}
		// flux git repo did not existed
		// create
		newFluxGitRepo := createUnstructuredFluxGitRepo(repo)
		if err = r.Create(ctx, newFluxGitRepo); err != nil {
			r.recorder.Eventf(newFluxGitRepo, v1.EventTypeWarning, "FailedWithFluxCD",
				"failed to create FluxCDGitRepository, error is: %v", err)
		}
	} else {
		// flux git repo existed
		// update
		newFluxGitRepo := createUnstructuredFluxGitRepo(repo)
		fluxGitRepo.Object["spec"] = newFluxGitRepo.Object["spec"]
		err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
			latestGitRepo := createBareFluxGitRepoObject()
			if err = r.Get(ctx, types.NamespacedName{
				Namespace: fluxGitRepo.GetNamespace(),
				Name:      fluxGitRepo.GetName(),
			}, latestGitRepo); err != nil {
				return
			}

			fluxGitRepo.SetResourceVersion(latestGitRepo.GetResourceVersion())
			r.log.Info("update FluxCDGitRepository", "name", fluxGitRepo.GetName())
			err = r.Update(ctx, fluxGitRepo)
			return
		})

	}
	return
}

func createUnstructuredFluxGitRepo(repo *v1alpha3.GitRepository) *unstructured.Unstructured {
	newFluxGitRepo := createBareFluxGitRepoObject()
	newFluxGitRepo.SetNamespace(repo.GetNamespace())
	newFluxGitRepo.SetName(getFluxRepoName(repo.GetName()))

	// set url
	_ = unstructured.SetNestedField(newFluxGitRepo.Object, repo.Spec.URL, "spec", "url")
	// set interval
	_ = unstructured.SetNestedField(newFluxGitRepo.Object, "1m", "spec", "interval")
	// set secretRef
	if repo.Spec.Secret != nil && repo.Spec.Secret.Name != "" {
		_ = unstructured.SetNestedField(newFluxGitRepo.Object, repo.Spec.Secret.Name, "spec", "secretRef", "name")
	}

	newFluxGitRepo.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": v1alpha1.GroupName,
	})

	return newFluxGitRepo
}

func getFluxRepoName(name string) string {
	return fmt.Sprintf("fluxcd-%s", name)
}

// GetName returns the name of this reconciler
func (r *GitRepositoryReconciler) GetName() string {
	return "FluxGitRepositoryReconciler"
}

func createBareFluxGitRepoObject() *unstructured.Unstructured {
	fluxGitRepo := &unstructured.Unstructured{}
	fluxGitRepo.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "source.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "GitRepository",
	})
	return fluxGitRepo
}

// isArtifactRepo check whether the repo is ArtifactRepo
func isArtifactRepo(repo *v1alpha3.GitRepository) bool {
	if v, ok := repo.GetLabels()[v1alpha1.ArtifactRepoLabelKey]; ok {
		return v == "true"
	}
	return false
}

// SetupWithManager setups the reconciler with a manager
// setup the logger, recorder
func (r *GitRepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		Named("fluxcd_git_repository_controller").
		For(&v1alpha3.GitRepository{}).
		Complete(r)
}
