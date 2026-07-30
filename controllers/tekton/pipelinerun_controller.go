/*
Copyright 2026 The KubeSphere Authors.

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

// Package tekton implements the Tekton execution adapter for KubeSphere DevOps.
package tekton

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/pipelineengine"
)

const (
	defaultPollInterval      = 2 * time.Second
	managedByValue           = "ks-devops-tekton-adapter"
	labelManagedBy           = "app.kubernetes.io/managed-by"
	annotationKSEPipeline    = "devops.kubesphere.io/kse-pipeline"
	annotationKSEPipelineRun = "devops.kubesphere.io/kse-pipelinerun"
)

var tektonPipelineRunGVK = schema.GroupVersionKind{
	Group:   "tekton.dev",
	Version: "v1",
	Kind:    "PipelineRun",
}

// Reconciler creates native Tekton PipelineRuns and mirrors their status to KubeSphere PipelineRuns.
type Reconciler struct {
	client.Client
	PollInterval time.Duration
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;watch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch

// Reconcile ensures that one KubeSphere Tekton PipelineRun has one native Tekton PipelineRun.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	run := &devopsv1alpha3.PipelineRun{}
	if err := r.Get(ctx, req.NamespacedName, run); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	runSelectsTekton := pipelineengine.IsTekton(run)
	if !runSelectsTekton && (run.Spec.PipelineRef == nil || run.Spec.PipelineRef.Name == "") {
		return ctrl.Result{}, nil
	}
	pipeline, err := r.getPipeline(ctx, run)
	if err != nil {
		if !runSelectsTekton && apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !runSelectsTekton && !pipelineengine.IsTekton(pipeline) {
		return ctrl.Result{}, nil
	}
	if !run.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	desired, err := buildTektonPipelineRun(run, pipeline)
	if err != nil {
		return ctrl.Result{}, r.markConfigurationFailed(ctx, run, err)
	}

	actual := newTektonPipelineRun(run.Namespace, run.Name)
	if err = r.Get(ctx, client.ObjectKeyFromObject(actual), actual); apierrors.IsNotFound(err) {
		if err = r.Create(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.recordNativeName(ctx, run, desired.GetName()); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.updateStatus(ctx, run, pendingStatus()); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: r.pollInterval()}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	if !isManagedByRun(actual, run) {
		return ctrl.Result{}, fmt.Errorf("native Tekton PipelineRun %s/%s already exists and is not managed by KubeSphere PipelineRun %s", actual.GetNamespace(), actual.GetName(), run.Name)
	}
	if err = r.recordNativeName(ctx, run, actual.GetName()); err != nil {
		return ctrl.Result{}, err
	}

	status := statusFromTekton(actual)
	if !isTerminal(status.Phase) {
		if err = r.cancelIfRequested(ctx, run, actual); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err = r.updateStatus(ctx, run, status); err != nil {
		return ctrl.Result{}, err
	}
	if isTerminal(status.Phase) {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{RequeueAfter: r.pollInterval()}, nil
}

// SetupWithManager registers the Tekton adapter with the controller manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("tekton_pipelinerun_controller").
		For(&devopsv1alpha3.PipelineRun{}).
		Complete(r)
}

// getPipeline returns the KubeSphere Pipeline referenced by a PipelineRun.
func (r *Reconciler) getPipeline(ctx context.Context, run *devopsv1alpha3.PipelineRun) (*devopsv1alpha3.Pipeline, error) {
	if run.Spec.PipelineRef == nil || run.Spec.PipelineRef.Name == "" {
		return nil, fmt.Errorf("PipelineRun %s/%s does not define spec.pipelineRef.name", run.Namespace, run.Name)
	}
	pipeline := &devopsv1alpha3.Pipeline{}
	key := client.ObjectKey{Namespace: run.Namespace, Name: run.Spec.PipelineRef.Name}
	if err := r.Get(ctx, key, pipeline); err != nil {
		return nil, fmt.Errorf("get Pipeline %s: %w", key, err)
	}
	return pipeline, nil
}

// pollInterval returns the configured status polling interval.
func (r *Reconciler) pollInterval() time.Duration {
	if r.PollInterval > 0 {
		return r.PollInterval
	}
	return defaultPollInterval
}

// recordNativeName stores the native Tekton PipelineRun identity on the KubeSphere object.
func (r *Reconciler) recordNativeName(ctx context.Context, run *devopsv1alpha3.PipelineRun, name string) error {
	if run.Annotations[pipelineengine.AnnotationTektonPipelineRun] == name {
		return nil
	}
	base := run.DeepCopy()
	if run.Annotations == nil {
		run.Annotations = map[string]string{}
	}
	run.Annotations[pipelineengine.AnnotationTektonPipelineRun] = name
	return r.Patch(ctx, run, client.MergeFrom(base))
}

// cancelIfRequested propagates a KubeSphere stop action to Tekton.
func (r *Reconciler) cancelIfRequested(ctx context.Context, run *devopsv1alpha3.PipelineRun, native *unstructured.Unstructured) error {
	if run.Spec.Action == nil || *run.Spec.Action != devopsv1alpha3.Stop {
		return nil
	}
	status, _, _ := unstructured.NestedString(native.Object, "spec", "status")
	if status == "Cancelled" {
		return nil
	}
	if err := unstructured.SetNestedField(native.Object, "Cancelled", "spec", "status"); err != nil {
		return err
	}
	return r.Update(ctx, native)
}

// markConfigurationFailed exposes invalid adapter configuration through the normal KubeSphere status.
func (r *Reconciler) markConfigurationFailed(ctx context.Context, run *devopsv1alpha3.PipelineRun, cause error) error {
	now := metav1.Now()
	status := devopsv1alpha3.PipelineRunStatus{
		CompletionTime: &now,
		Phase:          devopsv1alpha3.Failed,
		Conditions: []devopsv1alpha3.Condition{{
			Type:               devopsv1alpha3.ConditionSucceeded,
			Status:             devopsv1alpha3.ConditionFalse,
			Reason:             "InvalidTektonConfiguration",
			Message:            cause.Error(),
			LastProbeTime:      now,
			LastTransitionTime: now,
		}},
	}
	return r.updateStatus(ctx, run, status)
}

// updateStatus updates the KubeSphere status only when the mirrored state changed.
func (r *Reconciler) updateStatus(ctx context.Context, run *devopsv1alpha3.PipelineRun, desired devopsv1alpha3.PipelineRunStatus) error {
	desired.UpdateTime = nil
	current := run.Status.DeepCopy()
	current.UpdateTime = nil
	if reflect.DeepEqual(*current, desired) {
		return nil
	}
	base := run.DeepCopy()
	now := metav1.Now()
	desired.UpdateTime = &now
	run.Status = desired
	return r.Status().Patch(ctx, run, client.MergeFrom(base))
}

// buildTektonPipelineRun translates the engine-neutral KubeSphere run into a native Tekton v1 resource.
func buildTektonPipelineRun(run *devopsv1alpha3.PipelineRun, pipeline *devopsv1alpha3.Pipeline) (*unstructured.Unstructured, error) {
	pipelineName := pipelineengine.ResolveAnnotation(run, pipeline, pipelineengine.AnnotationTektonPipeline)
	if pipelineName == "" {
		pipelineName = pipeline.Name
	}
	if pipelineName == "" {
		return nil, fmt.Errorf("native Tekton Pipeline name is empty")
	}

	params := make([]interface{}, 0, len(run.Spec.Parameters))
	for _, parameter := range run.Spec.Parameters {
		params = append(params, map[string]interface{}{
			"name":  parameter.Name,
			"value": parameter.Value,
		})
	}
	spec := map[string]interface{}{
		"pipelineRef": map[string]interface{}{"name": pipelineName},
	}
	if len(params) > 0 {
		spec["params"] = params
	}
	if serviceAccount := pipelineengine.ResolveAnnotation(run, pipeline, pipelineengine.AnnotationTektonServiceAccount); serviceAccount != "" {
		spec["taskRunTemplate"] = map[string]interface{}{"serviceAccountName": serviceAccount}
	}
	if timeout := pipelineengine.ResolveAnnotation(run, pipeline, pipelineengine.AnnotationTektonTimeout); timeout != "" {
		if _, err := time.ParseDuration(timeout); err != nil {
			return nil, fmt.Errorf("invalid Tekton timeout %q: %w", timeout, err)
		}
		spec["timeouts"] = map[string]interface{}{"pipeline": timeout}
	}
	if workspace, err := workspaceBinding(run, pipeline); err != nil {
		return nil, err
	} else if workspace != nil {
		spec["workspaces"] = []interface{}{workspace}
	}

	native := newTektonPipelineRun(run.Namespace, run.Name)
	native.SetLabels(map[string]string{
		labelManagedBy: managedByValue,
	})
	native.SetAnnotations(map[string]string{
		annotationKSEPipeline:    pipeline.Name,
		annotationKSEPipelineRun: run.Name,
	})
	controller, blockOwnerDeletion := true, true
	native.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion:         devopsv1alpha3.GroupVersion.String(),
		Kind:               "PipelineRun",
		Name:               run.Name,
		UID:                run.UID,
		Controller:         &controller,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}})
	native.Object["spec"] = spec
	return native, nil
}

// workspaceBinding builds one optional Tekton workspace binding from Pipeline annotations.
func workspaceBinding(run *devopsv1alpha3.PipelineRun, pipeline *devopsv1alpha3.Pipeline) (map[string]interface{}, error) {
	name := pipelineengine.ResolveAnnotation(run, pipeline, pipelineengine.AnnotationTektonWorkspaceName)
	claim := pipelineengine.ResolveAnnotation(run, pipeline, pipelineengine.AnnotationTektonWorkspaceClaim)
	emptyDirValue := pipelineengine.ResolveAnnotation(run, pipeline, pipelineengine.AnnotationTektonWorkspaceEmptyDir)
	if name == "" && claim == "" && emptyDirValue == "" {
		return nil, nil
	}
	if name == "" {
		return nil, fmt.Errorf("%s is required when a Tekton workspace binding is configured", pipelineengine.AnnotationTektonWorkspaceName)
	}
	binding := map[string]interface{}{"name": name}
	if claim != "" {
		binding["persistentVolumeClaim"] = map[string]interface{}{"claimName": claim}
		return binding, nil
	}
	emptyDir, err := strconv.ParseBool(emptyDirValue)
	if err != nil || !emptyDir {
		return nil, fmt.Errorf("workspace %q requires either %s or %s=true", name, pipelineengine.AnnotationTektonWorkspaceClaim, pipelineengine.AnnotationTektonWorkspaceEmptyDir)
	}
	binding["emptyDir"] = map[string]interface{}{}
	return binding, nil
}

// newTektonPipelineRun creates an unstructured native Tekton PipelineRun identity.
func newTektonPipelineRun(namespace, name string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(tektonPipelineRunGVK)
	object.SetNamespace(namespace)
	object.SetName(name)
	return object
}

// isManagedByRun verifies that a same-name native object belongs to the expected KubeSphere run.
func isManagedByRun(native *unstructured.Unstructured, run *devopsv1alpha3.PipelineRun) bool {
	labels := native.GetLabels()
	return labels[labelManagedBy] == managedByValue && metav1.IsControlledBy(native, run)
}

// pendingStatus returns the initial state exposed while Tekton schedules the run.
func pendingStatus() devopsv1alpha3.PipelineRunStatus {
	return devopsv1alpha3.PipelineRunStatus{Phase: devopsv1alpha3.Pending}
}

// statusFromTekton converts native Tekton status fields into the existing KubeSphere status contract.
func statusFromTekton(native *unstructured.Unstructured) devopsv1alpha3.PipelineRunStatus {
	status := pendingStatus()
	status.StartTime = nestedTime(native.Object, "status", "startTime")
	status.CompletionTime = nestedTime(native.Object, "status", "completionTime")
	conditions, _, _ := unstructured.NestedSlice(native.Object, "status", "conditions")
	for _, item := range conditions {
		conditionMap, ok := item.(map[string]interface{})
		if !ok || stringValue(conditionMap, "type") != "Succeeded" {
			continue
		}
		transition := parseTime(stringValue(conditionMap, "lastTransitionTime"))
		condition := devopsv1alpha3.Condition{
			Type:               devopsv1alpha3.ConditionSucceeded,
			Status:             devopsv1alpha3.ConditionStatus(stringValue(conditionMap, "status")),
			Reason:             stringValue(conditionMap, "reason"),
			Message:            stringValue(conditionMap, "message"),
			LastProbeTime:      transition,
			LastTransitionTime: transition,
		}
		status.Conditions = []devopsv1alpha3.Condition{condition}
		status.Phase = phaseFromCondition(condition)
		break
	}
	if len(status.Conditions) == 0 && status.StartTime != nil {
		status.Phase = devopsv1alpha3.Running
	}
	return status
}

// phaseFromCondition maps the Tekton Succeeded condition to a KubeSphere phase.
func phaseFromCondition(condition devopsv1alpha3.Condition) devopsv1alpha3.RunPhase {
	switch condition.Status {
	case devopsv1alpha3.ConditionTrue:
		return devopsv1alpha3.Succeeded
	case devopsv1alpha3.ConditionFalse:
		if strings.Contains(strings.ToLower(condition.Reason), "cancel") || strings.Contains(strings.ToLower(condition.Message), "cancel") {
			return devopsv1alpha3.Cancelled
		}
		return devopsv1alpha3.Failed
	case devopsv1alpha3.ConditionUnknown:
		return devopsv1alpha3.Running
	default:
		return devopsv1alpha3.Unknown
	}
}

// nestedTime reads an RFC3339 timestamp from an unstructured object.
func nestedTime(object map[string]interface{}, fields ...string) *metav1.Time {
	value, found, _ := unstructured.NestedString(object, fields...)
	if !found || value == "" {
		return nil
	}
	parsed := parseTime(value)
	if parsed.IsZero() {
		return nil
	}
	return &parsed
}

// parseTime converts an RFC3339 timestamp and returns the zero value when it is invalid.
func parseTime(value string) metav1.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return metav1.Time{}
	}
	return metav1.NewTime(parsed)
}

// stringValue returns one string field from an unstructured map.
func stringValue(object map[string]interface{}, key string) string {
	value, _ := object[key].(string)
	return value
}

// isTerminal reports whether a KubeSphere run phase no longer requires polling.
func isTerminal(phase devopsv1alpha3.RunPhase) bool {
	return phase == devopsv1alpha3.Succeeded || phase == devopsv1alpha3.Failed || phase == devopsv1alpha3.Cancelled
}
