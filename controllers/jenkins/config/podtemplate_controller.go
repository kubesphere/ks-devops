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

package config

import (
	"context"
	"github.com/go-logr/logr"
	k8s "github.com/jenkins-zh/jenkins-client/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/controllers/predicate"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	"kubesphere.io/devops/pkg/utils/stringutils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;update
//+kubebuilder:rbac:groups="",resources=podtemplates,verbs=get;list;watch;update

// PodTemplateReconciler responsible for the Jenkins podTemplate sync
type PodTemplateReconciler struct {
	LabelSelector            string
	TargetConfigMapName      string
	TargetConfigMapNamespace string
	TargetConfigMapKey       string

	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile is the entrypoint of this reconciler
func (r *PodTemplateReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	r.log.Info("start to reconcile PodTemplate", "resource", req)
	ctx := context.Background()

	podTemplate := &v1.PodTemplate{}
	if err = r.Get(ctx, req.NamespacedName, podTemplate); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// make sure all the PodTemplates have finalizer
	if k8sutil.AddFinalizer(&podTemplate.ObjectMeta, podTemplateFinalizer) {
		if err = r.Update(ctx, podTemplate); err != nil {
			return
		}
	}

	// get the Jenkins CasC data that we will manipulate
	cm := &v1.ConfigMap{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: r.TargetConfigMapNamespace,
		Name:      r.TargetConfigMapName,
	}, cm); err != nil {
		// we will handle it only when the cm exists
		err = client.IgnoreNotFound(err)
		return
	}
	data := strings.TrimSpace(cm.Data[r.TargetConfigMapKey])
	if data == "" {
		r.log.V(7).Info("skip update cm due to expect key is empty", "resource", req)
		return
	}

	casc := k8s.JenkinsConfig{
		Config: []byte(data),
	}

	// manipulate the data
	if podTemplate.DeletionTimestamp.IsZero() {
		if err = casc.ReplaceOrAddPodTemplate(podTemplate); err == nil {
			cm.Data[r.TargetConfigMapKey] = casc.GetConfigAsString()

			// write back the data
			err = r.Update(ctx, cm)
		}
	} else {
		if err = casc.RemovePodTemplate(podTemplate.Name); err == nil {
			cm.Data[r.TargetConfigMapKey] = casc.GetConfigAsString()
			k8sutil.RemoveFinalizer(&podTemplate.ObjectMeta, podTemplateFinalizer)
			if err = r.Update(ctx, podTemplate); err == nil {
				// write back the data
				err = r.Update(ctx, cm)
			}
		}
	}
	return
}

// GetName returns the name of this reconcile
func (r *PodTemplateReconciler) GetName() string {
	return "pod-template"
}

// GetGroupName ret
func (r *PodTemplateReconciler) GetGroupName() string {
	return reconcilerGroupName
}

// SetupWithManager setups the reconciler
func (r *PodTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	r.LabelSelector = stringutils.SetOrDefault(r.LabelSelector, "jenkins.agent.pod")
	r.TargetConfigMapName = stringutils.SetOrDefault(r.TargetConfigMapName, "jenkins-casc-config")
	r.TargetConfigMapNamespace = stringutils.SetOrDefault(r.TargetConfigMapNamespace, "kubesphere-devops-system")
	r.TargetConfigMapKey = stringutils.SetOrDefault(r.TargetConfigMapKey, "jenkins_user.yaml")

	var withLabelPredicate = predicate.NewPredicateFuncs(predicate.NewFilterHasLabel(r.LabelSelector))
	return ctrl.NewControllerManagedBy(mgr).WithEventFilter(withLabelPredicate).For(&v1.PodTemplate{}).Complete(r)
}
