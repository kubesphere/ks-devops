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

package addon

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"html/template"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Reconciler takes the responsible for addon lifecycle
type Reconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

const defaultNamespace = "default"

// Reconcile responsible for addon lifecycle
func (r *Reconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	r.log.Info(fmt.Sprintf("start to reconcile addon: %s", req.String()))

	addon := &v1alpha1.Addon{}
	if err = r.Client.Get(ctx, req.NamespacedName, addon); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if beingDeleting(addon) {
		return r.cleanup(addon)
	}

	err = r.addonHandle(ctx, addon)
	return
}

func beingDeleting(addon *v1alpha1.Addon) bool {
	return addon != nil && !addon.DeletionTimestamp.IsZero()
}

func (r *Reconciler) cleanup(addon *v1alpha1.Addon) (result ctrl.Result, err error) {
	var obj *unstructured.Unstructured
	if obj, _, err = r.findTemplateInstance(context.Background(), addon); err != nil && !apierrors.IsNotFound(err) {
		return
	}

	if err = r.Client.Delete(context.Background(), obj); err == nil {
		k8sutil.RemoveFinalizer(&addon.ObjectMeta, v1alpha1.AddonFinalizerName)
		err = r.Client.Update(context.Background(), addon)
	}
	return
}

const (
	// EventReasonMissing represents the reason because of missing something
	EventReasonMissing = "Missing"
)

func (r *Reconciler) addonHandle(ctx context.Context, addon *v1alpha1.Addon) (err error) {
	var tpl string
	var obj *unstructured.Unstructured
	if obj, tpl, err = r.findTemplateInstance(ctx, addon); err != nil {
		if apierrors.IsNotFound(err) {
			err = r.Client.Create(ctx, obj)
		}
	} else {
		resVersion := obj.GetResourceVersion()

		if err = yaml.Unmarshal([]byte(tpl), obj); err != nil {
			err = fmt.Errorf("failed parse template to addon, error is %v", err)
			return
		}
		obj.SetResourceVersion(resVersion)
		obj.SetName(addon.Name)
		obj.SetNamespace(defaultNamespace)
		err = r.Client.Update(ctx, obj)
	}

	// add finalizer
	if err == nil {
		k8sutil.AddFinalizer(&addon.ObjectMeta, v1alpha1.AddonFinalizerName)
		err = r.Update(ctx, addon)
	}
	return
}

func (r *Reconciler) findTemplateInstance(ctx context.Context, addon *v1alpha1.Addon) (instance *unstructured.Unstructured, tpl string, err error) {
	strategy := addon.Spec.Strategy

	addonStrategy := &v1alpha1.AddonStrategy{}
	if err = r.Client.Get(ctx, types.NamespacedName{Name: strategy.Name}, addonStrategy); err != nil {
		r.recorder.Eventf(addon, corev1.EventTypeWarning, EventReasonMissing, "failed to get AddonStrategy with: %s", strategy.Name)
		return
	}

	if !r.supportedStrategy(addonStrategy) {
		err = fmt.Errorf("not supported addon strategy: %s", strategy.Name)
		return
	}

	if addonStrategy.Spec.Template == "" {
		r.recorder.Eventf(addon, corev1.EventTypeWarning, EventReasonMissing, "no template found from %s", strategy.Name)
		err = fmt.Errorf("no template found from %s", strategy.Name)
		return
	}

	if tpl, err = getTemplate(addonStrategy.Spec.Template, addon); err != nil {
		err = fmt.Errorf("failed to render template: %s, error: %v", addonStrategy.Spec.Template, err)
		return
	}

	instance = &unstructured.Unstructured{}
	if err = yaml.Unmarshal([]byte(tpl), instance); err != nil {
		err = fmt.Errorf("failed parse template to addon, error is %v", err)
		return
	}
	instance.SetName(addon.Name)
	instance.SetNamespace(defaultNamespace)

	err = r.Client.Get(ctx, types.NamespacedName{Name: addon.Name, Namespace: defaultNamespace}, instance)
	return
}

func getTemplate(tpl string, addon *v1alpha1.Addon) (result string, err error) {
	var addonTpl *template.Template
	if addonTpl, err = template.New("addon").Parse(tpl); err == nil {
		buf := bytes.NewBuffer([]byte{})
		if err = addonTpl.Execute(buf, addon); err == nil {
			result = buf.String()
		}
	}
	return
}

func (r *Reconciler) supportedStrategy(strategy *v1alpha1.AddonStrategy) bool {
	// TODO support more types in the future
	if strategy != nil {
		return strategy.Spec.Type == v1alpha1.AddonInstallStrategySimpleOperator
	}
	return false
}

// SetupWithManager set the reconcilers
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName("AddonReconciler")
	r.recorder = mgr.GetEventRecorderFor("addon-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Addon{}).
		Complete(r)
}
