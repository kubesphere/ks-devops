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
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
)

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=gitrepositories,verbs=get;list;watch;update

// AmendReconciler reconciles a GitRepository object
// See the main reason for this controller, https://github.com/kubesphere/ks-devops/issues/567
type AmendReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

func (r *AmendReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	r.log.Info(fmt.Sprintf("start to AmendReconciler: %s", req.String()))

	repo := &v1alpha3.GitRepository{}
	if err = r.Get(ctx, req.NamespacedName, repo); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if amendGitlabURL(repo) {
		err = r.Update(ctx, repo)
	}
	return
}

func amendGitlabURL(repo *v1alpha3.GitRepository) (changed bool) {
	if strings.ToLower(repo.Spec.Provider) != "gitlab" {
		return
	}

	if !strings.HasSuffix(repo.Spec.URL, ".git") {
		repo.Spec.URL += ".git"
		changed = true
	}
	return
}

func (r *AmendReconciler) GetName() string {
	return "git-repository-amend"
}

func (r *AmendReconciler) GetGroupName() string {
	return groupName
}

// SetupWithManager sets up the controller with the Manager.
func (r *AmendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	r.log = ctrl.Log.WithName(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha3.GitRepository{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
