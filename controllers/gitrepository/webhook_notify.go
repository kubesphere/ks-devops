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

package gitrepository

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// WebhookReconciler notifies the GitRepositories if there are corresponding webhooks changed
type WebhookReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=webhooks,verbs=get;list;watch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=gitrepositories,verbs=get;update

// Reconcile handles the update events of Webhook, then send the notification to a GitRepository
func (r *WebhookReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	log := r.log.WithValues("webhook", req.NamespacedName)

	webhook := &v1alpha1.Webhook{}
	if err = r.Client.Get(ctx, req.NamespacedName, webhook); err != nil {
		log.Error(err, "unable to fetch webhook")
		err = client.IgnoreNotFound(err)
		return
	}

	// skip those don't have the desired annotation
	repos, ok := webhook.Annotations[v1alpha1.AnnotationKeyGitRepos]
	if !ok {
		return
	}

	if err = r.notifyGitRepos(webhook.Namespace, repos); err != nil {
		result = ctrl.Result{
			Requeue: true,
		}
	}
	return
}

func (r *WebhookReconciler) notifyGitRepos(ns, repos string) (err error) {
	var errs []error
	repoArray := strings.Split(repos, ",")
	for index := range repoArray {
		if e := r.notifyGitRepo(ns, repoArray[index]); e != nil {
			errs = append(errs, e)
		}
	}

	if len(errs) > 0 {
		err = fmt.Errorf("error happend when notifying git repos, error: %v", errs)
	}
	return
}

func (r *WebhookReconciler) notifyGitRepo(ns, name string) (err error) {
	gitRepo := &v1alpha1.GitRepository{}

	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, gitRepo); err != nil {
		err = client.IgnoreNotFound(err)
	}

	if gitRepo.Annotations == nil {
		gitRepo.Annotations = map[string]string{}
	}
	gitRepo.Annotations[v1alpha1.AnnotationKeyWebhookUpdates] = names.SimpleNameGenerator.GenerateName("")
	err = r.Client.Update(context.TODO(), gitRepo)
	return
}

func (r *WebhookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("webhook-notify")
	r.log = ctrl.Log.WithName("webhook-notify")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Webhook{}).
		Complete(r)
}
