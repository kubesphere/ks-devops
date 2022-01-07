/*
Copyright 2021 The KubeSphere Authors.

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

package gitrepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport/oauth2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles a GitRepository object
type Reconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=webhooks,verbs=get;list;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=secrets,verbs=get
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=gitrepositories,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO currently, this controller only can add a new webhook to a git repository. It might be
//   possible to add more and more wenhooks, and there's no way to clean them.
//   Second problem is that this controller cannot update the webhook when a webhook changed.
func (r *Reconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	log := r.log.WithValues("GitRepository", req.NamespacedName)

	repo := &v1alpha1.GitRepository{}
	if err = r.Client.Get(ctx, req.NamespacedName, repo); err != nil {
		log.Error(err, "unable to fetch GitRepository")
		result = ctrl.Result{}
		err = client.IgnoreNotFound(err)
		return
	}

	webhooks := repo.Spec.Webhooks
	if len(webhooks) == 0 {
		// do nothing if there are not any webhooks
		return
	}

	// make links between the webhook and git repositories
	if err = r.linkToWebhooks(repo); err != nil {
		return
	}

	secret := repo.Spec.Secret
	if secret == nil {
		result = ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}
		return
	}

	err = r.createOrUpdateWebhook(repo)
	return
}

func (r *Reconciler) createOrUpdateWebhook(repo *v1alpha1.GitRepository) (err error) {
	var gitClient *scm.Client
	if gitClient, err = r.getGitClient(repo); err != nil {
		return
	}

	repoAddress := getRepo(repo)
	if repoAddress == "" {
		err = fmt.Errorf("not support: %v", repo.Spec.URL)
		return
	}

	var hooks []*scm.Hook
	if hooks, _, err = gitClient.Repositories.ListHooks(context.TODO(), repoAddress, scm.ListOptions{
		Page: 1,
		Size: 30,
	}); err != nil {
		err = fmt.Errorf("failed to list the existing webhooks, error: %v", err)
		return
	}

	for index := range repo.Spec.Webhooks {
		webhookRef := repo.Spec.Webhooks[index]
		webhook := &v1alpha1.Webhook{}
		if err = r.Client.Get(context.TODO(), types.NamespacedName{
			Namespace: repo.Namespace,
			Name:      webhookRef.Name,
		}, webhook); err != nil {
			// TODO need to print the error log output
			continue
		}

		// the token is optional, we can ignore the error
		webhookToken, _ := r.getTokenFromSecret(repo.Spec.Secret, repo.Namespace)

		// TODO users need to add every single event of target git provider if they want to add all of them
		//   it's possible to have a solution to allow users add all events in an easy way.
		//   For instance, we can use 'all' represents it.
		hookInput := &scm.HookInput{
			Name:         webhookRef.Name,
			Target:       webhook.Spec.Server,
			Secret:       webhookToken,
			SkipVerify:   webhook.Spec.SkipVerify,
			NativeEvents: webhook.Spec.Events,
		}

		if ok, id := exist(webhook.Spec.Server, hooks); ok {
			// update the existing webhooks
			_, _, err = gitClient.Repositories.UpdateHook(context.TODO(), repoAddress, id, hookInput)
		} else {
			// create the webhook
			_, _, err = gitClient.Repositories.CreateHook(context.TODO(), repoAddress, hookInput)
		}
	}
	return
}

func exist(server string, hooks []*scm.Hook) (exist bool, id string) {
	for _, hook := range hooks {
		if hook.Target == server {
			id = hook.ID
			exist = true
			break
		}
	}
	return
}

func (r *Reconciler) getGitClient(repo *v1alpha1.GitRepository) (client *scm.Client, err error) {
	provider := repo.Spec.Provider
	switch provider {
	case "github":
		client = github.NewDefault()
	case "gitlab":
		client = gitlab.NewDefault()
	default:
		err = errors.New("not support git provider: " + provider)
		return
	}

	var gitToken string
	secretRef := repo.Spec.Secret
	if gitToken, err = r.getTokenFromSecret(secretRef, repo.Namespace); err != nil {
		return
	}

	client.Client = &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&scm.Token{
					Token: gitToken,
				},
			),
		},
	}
	return
}

func (r *Reconciler) getTokenFromSecret(secretRef *v1.SecretReference, defaultNamespace string) (token string, err error) {
	var gitSecret *v1.Secret
	if gitSecret, err = r.getSecret(secretRef, defaultNamespace); err != nil {
		return
	}

	switch gitSecret.Type {
	case v1.SecretTypeBasicAuth:
		token = string(gitSecret.Data[v1.BasicAuthPasswordKey])
	case v1.SecretTypeOpaque:
		token = string(gitSecret.Data[v1.ServiceAccountTokenKey])
	}
	return
}

// getSecret returns the secret, taking the namespace from GitRepository if it is empty
func (r *Reconciler) getSecret(ref *v1.SecretReference, defaultNamespace string) (secret *v1.Secret, err error) {
	secret = &v1.Secret{}
	ns := ref.Namespace
	if ns == "" {
		ns = defaultNamespace
	}

	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: ns, Name: ref.Name,
	}, secret); err != nil {
		err = fmt.Errorf("cannot get secret %v, error is: %v", secret, err)
	}
	return
}

func getRepo(repo *v1alpha1.GitRepository) string {
	if repo == nil || repo.Spec.Provider == "" {
		return ""
	}

	address := repo.Spec.URL
	switch repo.Spec.Provider {
	case "github":
		return strings.ReplaceAll(address, "https://github.com/", "")
	case "gitlab":
		return strings.ReplaceAll(address, "https://gitlab.com/", "")
	}
	return ""
}

func (r *Reconciler) linkToWebhooks(repo *v1alpha1.GitRepository) (err error) {
	var failedLinks []string
	for i := range repo.Spec.Webhooks {
		webhookRef := repo.Spec.Webhooks[i]
		if err = linkToWebhook(webhookRef, repo, r.Client); err != nil {
			r.log.V(6).Info("failed to link to webhook: %v, error: %v", webhookRef, err)
			failedLinks = append(failedLinks, webhookRef.Name)
		}
	}

	if len(failedLinks) > 0 {
		err = fmt.Errorf("failed to link to webhooks: %v", failedLinks)
	}
	return
}

func linkToWebhook(webhookRef v1.LocalObjectReference, repo *v1alpha1.GitRepository, client client.Client) (err error) {
	webhook := &v1alpha1.Webhook{}
	if err = client.Get(context.TODO(), types.NamespacedName{Namespace: repo.Namespace, Name: webhookRef.Name}, webhook); err != nil {
		err = fmt.Errorf("cannot find webhook '%v', error： %v", webhookRef, err)
		return
	}

	webhook.Annotations = addToArrayInAnnotations(webhook.Annotations, v1alpha1.AnnotationKeyGitRepos, repo.Name)
	err = client.Update(context.TODO(), webhook)
	return
}

func addToArrayInAnnotations(array map[string]string, key, value string) map[string]string {
	if array == nil {
		return map[string]string{key: value}
	}

	if val, ok := array[key]; ok {
		if val == value || strings.Contains(val, ","+value) || strings.Contains(val, value+",") {
			return array
		}

		array[key] = val + "," + value
	} else {
		array[key] = value
	}
	return array
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// the server should obey Kubernetes naming convention: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	r.recorder = mgr.GetEventRecorderFor("gitrepository-controller")
	r.log = ctrl.Log.WithName("gitrepository-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GitRepository{}).
		Complete(r)
}
