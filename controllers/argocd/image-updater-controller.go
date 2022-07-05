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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=applications,verbs=get;list;update
//+kubebuilder:rbac:groups=gitops.kubesphere.io,resources=imageupdaters,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// ImageUpdaterReconciler is the reconciler of the ImageUpdater
type ImageUpdaterReconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile makes sure the Application has the expected annotations
func (r *ImageUpdaterReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	r.log.Info(fmt.Sprintf("start to reconcile imageUpdater: %s", req.String()))

	updater := &v1alpha1.ImageUpdater{}
	if err = r.Get(ctx, req.NamespacedName, updater); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// skip if kind is not argocd
	if updater.Spec.Kind != "argocd" {
		r.log.V(7).Info(fmt.Sprintf("skip %s due to the spec.kind value is not argocd", req.String()))
		return
	}

	argo := updater.Spec.Argo
	if argo == nil {
		r.log.V(7).Info(fmt.Sprintf("skip %s due to the Argo is nil", req.String()))
		return
	}

	appNs := req.Namespace
	appName := argo.App.Name
	if appName == "" {
		r.recorder.Eventf(updater, corev1.EventTypeWarning, "Missing", "application name is required")
		return
	}

	app := &v1alpha1.Application{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: appNs,
		Name:      appName,
	}, app); err != nil {
		result = ctrl.Result{RequeueAfter: time.Minute}
		return
	}

	if app.Annotations == nil {
		app.Annotations = map[string]string{}
	}
	setImagePreference(argo, app.Annotations)

	updateImageList(updater.Spec.Images, app.Annotations)
	err = r.Update(ctx, app)
	return
}

func setImagePreference(argo *v1alpha1.ArgoImageUpdater, annotations map[string]string) {
	// set write method
	if argo.Write.GetValue() != "" {
		annotations["argocd-image-updater.argoproj.io/write-back-method"] = argo.Write.GetValue()
	}

	for name, secret := range argo.Secrets {
		annotations[fmt.Sprintf("argocd-image-updater.argoproj.io/%s.pull-secret", name)] =
			fmt.Sprintf("pullsecret:%s", secret)
	}

	for name, tagReg := range argo.AllowTags {
		annotations[fmt.Sprintf("argocd-image-updater.argoproj.io/%s.allow-tags", name)] = fmt.Sprintf("regexp:%s", tagReg)
	}

	for name, strategy := range argo.UpdateStrategy {
		annotations[fmt.Sprintf("argocd-image-updater.argoproj.io/%s.update-strategy", name)] = strategy
	}

	for name, platform := range argo.Platforms {
		annotations[fmt.Sprintf("argocd-image-updater.argoproj.io/%s.platforms", name)] = platform
	}

	for name, ignoreTag := range argo.IgnoreTags {
		annotations[fmt.Sprintf("argocd-image-updater.argoproj.io/%s.ignore-tags", name)] = ignoreTag
	}
}

func updateImageList(images []string, annotations map[string]string) {
	if len(images) == 0 {
		if _, ok := annotations["argocd-image-updater.argoproj.io/image-list"]; !ok {
			return
		}
	}
	if len(images) == 0 {
		delete(annotations, "argocd-image-updater.argoproj.io/image-list")
	} else {
		annotations["argocd-image-updater.argoproj.io/image-list"] = strings.Join(images, ",")
	}
}

// GetName returns the name of this controller
func (r *ImageUpdaterReconciler) GetName() string {
	return "ImageUpdaterController"
}

// GetGroupName returns the group name of this controller
func (r *ImageUpdaterReconciler) GetGroupName() string {
	return controllerGroupName
}

// SetupWithManager setups the log and recorder
func (r *ImageUpdaterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ImageUpdater{}).
		Complete(r)
}
