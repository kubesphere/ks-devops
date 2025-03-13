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

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"

	v1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HttpTimeoutErrStr indicates that connection in http request is timeout(the str in error).
const HttpTimeoutErrStr = " (Client.Timeout exceeded while awaiting headers)"

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;update;patch;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// JenkinsfileReconciler will convert between JSON and Jenkinsfile (as groovy) formats
type JenkinsfileReconciler struct {
	log      logr.Logger
	recorder record.EventRecorder

	client.Client
	JenkinsClient core.Client
}

// Reconcile is the main entrypoint of this controller
func (r *JenkinsfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	pip := &v1alpha3.Pipeline{}
	if err = r.Get(ctx, req.NamespacedName, pip); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if pip.Spec.Type != v1alpha3.NoScmPipelineType || pip.Spec.Pipeline == nil {
		return
	}

	editMode := pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey]
	switch editMode {
	case v1alpha3.PipelineJenkinsfileEditModeRaw:
		result, err = r.reconcileJenkinsfileEditMode(pip, req.NamespacedName)
	case v1alpha3.PipelineJenkinsfileEditModeJSON:
		result, err = r.reconcileJSONEditMode(pip, req.NamespacedName)
	case "":
		// Reconcile pipeline version <= v3.3.2
		if _, ok := pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey]; !ok {
			if pip.Spec.Pipeline != nil && pip.Spec.Pipeline.Jenkinsfile != "" {
				result, err = r.reconcileJenkinsfileEditMode(pip, req.NamespacedName)
			}
		}
	default:
		r.log.Info(fmt.Sprintf("invalid edit mode: %s", editMode))
		return
	}
	return
}

func (r *JenkinsfileReconciler) reconcileJenkinsfileEditMode(pip *v1alpha3.Pipeline, pipelineKey client.ObjectKey) (
	result ctrl.Result, err error) {
	jenkinsfile := pip.Spec.Pipeline.Jenkinsfile
	toJsonJenkinsfile := ""
	if pip.Annotations == nil {
		pip.Annotations = map[string]string{}
	}

	// Users are able to clean jenkinsfile
	if jenkinsfile != "" {
		var toJSONResult core.GenericResult
		jenkinsfile = strings.ReplaceAll(jenkinsfile, "\\", "\\\\") // escape backslash
		if toJSONResult, err = r.JenkinsClient.ToJSON(jenkinsfile); err != nil || toJSONResult.GetStatus() != "success" {
			r.log.Error(err, "failed to convert jenkinsfile to json format")
			if err != nil {
				// ConnectRefused || Timeout when jenkins is starting(not ready), retry
				if errors.Is(err, syscall.ECONNREFUSED) || strings.Contains(err.Error(), HttpTimeoutErrStr) {
					r.log.Info("connect to jenkins failed, retry..")
					return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
				}
			}

			pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey] = ""
			pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = ""
			pip.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] = v1alpha3.PipelineJenkinsfileValidateFailure
			err = r.updateAnnotations(pip.Annotations, pipelineKey)
			return
		}
		toJsonJenkinsfile = toJSONResult.GetResult()
	}

	pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey] = toJsonJenkinsfile
	pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = ""
	pip.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] = v1alpha3.PipelineJenkinsfileValidateSuccess
	err = r.updateAnnotations(pip.Annotations, pipelineKey)
	return
}

func (r *JenkinsfileReconciler) updateAnnotations(annotations map[string]string, pipelineKey client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pipeline := &v1alpha3.Pipeline{}
		if err := r.Get(context.Background(), pipelineKey, pipeline); err != nil {
			return client.IgnoreNotFound(err)
		}
		if reflect.DeepEqual(pipeline.Annotations, annotations) {
			return nil
		}

		// update annotations
		pipeline.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey] = annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey]
		pipeline.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey]
		pipeline.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] = annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey]
		return r.Update(context.Background(), pipeline)
	})
}

func (r *JenkinsfileReconciler) reconcileJSONEditMode(pip *v1alpha3.Pipeline, pipelineKey client.ObjectKey) (
	result ctrl.Result, err error) {
	var jsonData string
	if jsonData = pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey]; jsonData != "" {
		var toResult core.GenericResult
		if toResult, err = r.JenkinsClient.ToJenkinsfile(jsonData); err != nil || toResult.GetStatus() != "success" {
			r.log.Error(err, "failed to convert json format to Jenkinsfile")
			pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = ""
			pip.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] = v1alpha3.PipelineJenkinsfileValidateFailure
			err = r.updateAnnotations(pip.Annotations, pipelineKey)
			return
		}
		jenkinsfile := toResult.GetResult()
		jenkinsfile = strings.ReplaceAll(jenkinsfile, "\\\\", "\\") // unescape backslash
		jenkinsfile = strings.ReplaceAll(jenkinsfile, `\'`, `'`)    // unescape single quote

		pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = ""
		pip.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] = v1alpha3.PipelineJenkinsfileValidateSuccess
		err = r.updateAnnotationsAndJenkinsfile(pip.Annotations, jenkinsfile, pipelineKey)
	}
	return
}

func (r *JenkinsfileReconciler) updateAnnotationsAndJenkinsfile(annotations map[string]string, jenkinsfile string, pipelineKey client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pipeline := &v1alpha3.Pipeline{}
		if err := r.Get(context.Background(), pipelineKey, pipeline); err != nil {
			return client.IgnoreNotFound(err)
		}
		if pipeline.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] == annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] &&
			pipeline.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] == annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] &&
			pipeline.Spec.Pipeline.Jenkinsfile == jenkinsfile {
			return nil
		}

		// update annotations
		pipeline.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey]
		pipeline.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey] = annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey]
		pipeline.Spec.Pipeline.Jenkinsfile = jenkinsfile
		return r.Update(context.Background(), pipeline)
	})
}

// GetName returns the name of this controller
func (r *JenkinsfileReconciler) GetName() string {
	return "JenkinsfileController"
}

// GetGroupName returns the group name of this controller
func (r *JenkinsfileReconciler) GetGroupName() string {
	return ControllerGroupName
}

// jenkinsfilePredicate returns a predicate only care about pipeline update event..
var jenkinsfilePredicate = predicate.Funcs{
	UpdateFunc: func(ue event.UpdateEvent) bool {
		oldPipeline, okOld := ue.ObjectOld.(*v1alpha3.Pipeline)
		newPipeline, okNew := ue.ObjectNew.(*v1alpha3.Pipeline)
		if okOld && okNew {
			if oldPipeline.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] != newPipeline.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] {
				return true
			}
		}
		return false
	},
	GenericFunc: func(ge event.GenericEvent) bool {
		return false
	},
}

// SetupWithManager setups the log and recorder
func (r *JenkinsfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		Named("jenkins_pipeline_jenkinsfile_controller").
		WithEventFilter(jenkinsfilePredicate).
		For(&v1alpha3.Pipeline{}).
		Complete(r)
}
