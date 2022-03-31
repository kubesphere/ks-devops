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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=gitrepositories,verbs=get;list;watch;update;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// GitRepositoryController is the reconciler of the GitRepository
type GitRepositoryController struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder

	// ArgoNamespace is the namespace of the ArgoCD instance
	ArgoNamespace string
}

// Reconcile maintains the Argo CD git repository secrets against to the GitRepository
func (c *GitRepositoryController) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	c.log.Info(fmt.Sprintf("start to git repository: %s", req.String()))

	repo := &v1alpha3.GitRepository{}
	if err = c.Get(ctx, req.NamespacedName, repo); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	var deleted bool
	if deleted, err = c.handleFinalizer(repo); deleted || err != nil {
		return
	}

	err = c.handleArgoGitRepo(repo)
	return
}

func (c *GitRepositoryController) handleFinalizer(repo *v1alpha3.GitRepository) (deleted bool, err error) {
	if repo.ObjectMeta.DeletionTimestamp.IsZero() {
		if k8sutil.AddFinalizer(&repo.ObjectMeta, v1alpha3.GitRepoFinalizerName) {
			err = c.Update(context.TODO(), repo)
		}
		return
	}

	// finalize the Argo Git Repositories aka the special secrets
	deleted = true
	ctx := context.Background()
	ns, name := c.ArgoNamespace, getSecretName(repo.Name)

	secret := &v1.Secret{}
	if err = c.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, secret); err != nil {
		err = client.IgnoreNotFound(err)
	} else {
		err = c.Delete(ctx, secret)
	}

	if err == nil {
		k8sutil.RemoveFinalizer(&repo.ObjectMeta, v1alpha3.GitRepoFinalizerName)
		err = c.Update(context.TODO(), repo)
	}
	return
}

func (c *GitRepositoryController) handleArgoGitRepo(repo *v1alpha3.GitRepository) (err error) {
	ctx := context.Background()
	ns, name := c.ArgoNamespace, getSecretName(repo.Name)

	secret := &v1.Secret{}
	if err = c.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, secret); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return
		}

		// no existing secret, create a new one
		secret.SetNamespace(ns)
		secret.SetName(name)
		c.setArgoGitRepoFields(repo, secret)
		c.log.Info(fmt.Sprintf("create secret for ArgoCD: %s/%s", ns, name))
		err = c.Create(ctx, secret)
	} else {
		c.setArgoGitRepoFields(repo, secret)
		err = c.Update(ctx, secret)
	}
	return
}

func (c *GitRepositoryController) setArgoGitRepoFields(repo *v1alpha3.GitRepository, secret *v1.Secret) {
	if secret.Labels == nil {
		secret.Labels = map[string]string{}
	}
	secret.Labels["argocd.argoproj.io/secret-type"] = "repository"

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	secret.Data["type"] = []byte("git")
	secret.Data["url"] = []byte(repo.Spec.URL)

	c.setArgoGitRepoAuth(secret, repo.Spec.Secret)
}

func (c *GitRepositoryController) setArgoGitRepoAuth(secret *v1.Secret, ref *v1.SecretReference) {
	if ref == nil {
		return
	}

	authSecret := &v1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}, authSecret); err == nil {
		switch authSecret.Type {
		case v1.SecretTypeBasicAuth, v1alpha3.SecretTypeBasicAuth:
			secret.Data["username"] = authSecret.Data[v1.BasicAuthUsernameKey]
			secret.Data["password"] = authSecret.Data[v1.BasicAuthPasswordKey]
		default:
			c.log.V(4).Info("not support auth secret", "type", authSecret.Type)
		}
	}
}

func getSecretName(name string) string {
	return fmt.Sprintf("%s-repo", name)
}

// GetName returns the name of this reconciler
func (c *GitRepositoryController) GetName() string {
	return "GitRepositoryController"
}

// GetGroupName returns the group name of this reconciler
func (c *GitRepositoryController) GetGroupName() string {
	return controllerGroupName
}

// SetupWithManager setups the reconciler with a manager
// setup the logger, recorder
func (c *GitRepositoryController) SetupWithManager(mgr ctrl.Manager) error {
	c.log = ctrl.Log.WithName(c.GetName())
	c.recorder = mgr.GetEventRecorderFor(c.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		For(&v1alpha3.GitRepository{}).
		Complete(c)
}
