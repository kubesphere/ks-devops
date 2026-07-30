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

package tekton

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/pipelineengine"
)

// TestBuildTektonPipelineRun verifies translation of parameters, identity, credentials, workspace, and timeout.
func TestBuildTektonPipelineRun(t *testing.T) {
	pipeline, run := testObjects()
	pipeline.Annotations[pipelineengine.AnnotationTektonPipeline] = "native-build"
	pipeline.Annotations[pipelineengine.AnnotationTektonServiceAccount] = "builder"
	pipeline.Annotations[pipelineengine.AnnotationTektonWorkspaceName] = "source"
	pipeline.Annotations[pipelineengine.AnnotationTektonWorkspaceClaim] = "source-cache"
	pipeline.Annotations[pipelineengine.AnnotationTektonTimeout] = "20m"
	run.Spec.Parameters = []devopsv1alpha3.Parameter{{Name: "image", Value: "registry.example/app@sha256:123"}}

	native, err := buildTektonPipelineRun(run, pipeline)
	require.NoError(t, err)

	assert.Equal(t, tektonPipelineRunGVK, native.GroupVersionKind())
	assert.Equal(t, "native-build", mustNestedString(t, native.Object, "spec", "pipelineRef", "name"))
	assert.Equal(t, "builder", mustNestedString(t, native.Object, "spec", "taskRunTemplate", "serviceAccountName"))
	assert.Equal(t, "20m", mustNestedString(t, native.Object, "spec", "timeouts", "pipeline"))
	assert.Equal(t, "source-cache", mustNestedString(t, native.Object, "spec", "workspaces", "0", "persistentVolumeClaim", "claimName"))
	assert.Equal(t, managedByValue, native.GetLabels()[labelManagedBy])
	require.Len(t, native.GetOwnerReferences(), 1)
	assert.Equal(t, run.Name, native.GetOwnerReferences()[0].Name)
}

// TestBuildTektonPipelineRunRejectsInvalidConfig verifies fail-fast validation for user annotations.
func TestBuildTektonPipelineRunRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
	}{
		{
			name: "invalid timeout",
			annotations: map[string]string{
				pipelineengine.AnnotationTektonTimeout: "fifteen minutes",
			},
		},
		{
			name: "workspace without storage",
			annotations: map[string]string{
				pipelineengine.AnnotationTektonWorkspaceName: "source",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pipeline, run := testObjects()
			for key, value := range test.annotations {
				pipeline.Annotations[key] = value
			}
			_, err := buildTektonPipelineRun(run, pipeline)
			require.Error(t, err)
		})
	}
}

// TestReconcileCreatesNativeRun verifies idempotent native PipelineRun creation.
func TestReconcileCreatesNativeRun(t *testing.T) {
	pipeline, run := testObjects()
	delete(run.Annotations, pipelineengine.AnnotationEngine)
	reconciler := newTestReconciler(t, pipeline, run)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)})
	require.NoError(t, err)
	assert.Greater(t, result.RequeueAfter, time.Duration(0))

	native := newTektonPipelineRun(run.Namespace, run.Name)
	require.NoError(t, reconciler.Get(context.Background(), client.ObjectKeyFromObject(native), native))
	assert.Equal(t, pipeline.Name, mustNestedString(t, native.Object, "spec", "pipelineRef", "name"))

	updated := &devopsv1alpha3.PipelineRun{}
	require.NoError(t, reconciler.Get(context.Background(), client.ObjectKeyFromObject(run), updated))
	assert.Equal(t, run.Name, updated.Annotations[pipelineengine.AnnotationTektonPipelineRun])
	assert.Equal(t, devopsv1alpha3.Pending, updated.Status.Phase)
}

// TestReconcileIgnoresJenkinsRun verifies that unrelated Jenkins resources do not require Tekton configuration.
func TestReconcileIgnoresJenkinsRun(t *testing.T) {
	run := &devopsv1alpha3.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{Name: "jenkins-run", Namespace: "demo"},
	}
	reconciler := newTestReconciler(t, run)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)})
	require.NoError(t, err)
	assert.Zero(t, result.RequeueAfter)
}

// TestReconcileMirrorsSucceededStatus verifies Tekton-to-KubeSphere status translation.
func TestReconcileMirrorsSucceededStatus(t *testing.T) {
	pipeline, run := testObjects()
	native, err := buildTektonPipelineRun(run, pipeline)
	require.NoError(t, err)
	native.Object["status"] = map[string]interface{}{
		"startTime":      "2026-07-27T01:00:00Z",
		"completionTime": "2026-07-27T01:01:00Z",
		"conditions": []interface{}{map[string]interface{}{
			"type":               "Succeeded",
			"status":             "True",
			"reason":             "Succeeded",
			"message":            "Tasks Completed: 2 (Failed: 0, Cancelled 0), Skipped: 0",
			"lastTransitionTime": "2026-07-27T01:01:00Z",
		}},
	}
	reconciler := newTestReconciler(t, pipeline, run, native)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)})
	require.NoError(t, err)
	assert.Zero(t, result.RequeueAfter)

	updated := &devopsv1alpha3.PipelineRun{}
	require.NoError(t, reconciler.Get(context.Background(), client.ObjectKeyFromObject(run), updated))
	assert.Equal(t, devopsv1alpha3.Succeeded, updated.Status.Phase)
	require.NotNil(t, updated.Status.StartTime)
	require.NotNil(t, updated.Status.CompletionTime)
	require.Len(t, updated.Status.Conditions, 1)
	assert.Equal(t, "Succeeded", updated.Status.Conditions[0].Reason)
}

// TestReconcilePropagatesCancellation verifies that the existing KSE stop action cancels Tekton.
func TestReconcilePropagatesCancellation(t *testing.T) {
	pipeline, run := testObjects()
	action := devopsv1alpha3.Stop
	run.Spec.Action = &action
	native, err := buildTektonPipelineRun(run, pipeline)
	require.NoError(t, err)
	reconciler := newTestReconciler(t, pipeline, run, native)

	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)})
	require.NoError(t, err)

	updatedNative := newTektonPipelineRun(run.Namespace, run.Name)
	require.NoError(t, reconciler.Get(context.Background(), client.ObjectKeyFromObject(updatedNative), updatedNative))
	assert.Equal(t, "Cancelled", mustNestedString(t, updatedNative.Object, "spec", "status"))
}

// TestStatusFromTekton verifies running, failed, and cancelled condition mappings.
func TestStatusFromTekton(t *testing.T) {
	tests := []struct {
		name      string
		condition map[string]interface{}
		expected  devopsv1alpha3.RunPhase
	}{
		{name: "running", condition: map[string]interface{}{"type": "Succeeded", "status": "Unknown"}, expected: devopsv1alpha3.Running},
		{name: "failed", condition: map[string]interface{}{"type": "Succeeded", "status": "False", "reason": "Failed"}, expected: devopsv1alpha3.Failed},
		{name: "cancelled", condition: map[string]interface{}{"type": "Succeeded", "status": "False", "reason": "PipelineRunCancelled"}, expected: devopsv1alpha3.Cancelled},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			native := newTektonPipelineRun("demo", "run")
			native.Object["status"] = map[string]interface{}{"conditions": []interface{}{test.condition}}
			assert.Equal(t, test.expected, statusFromTekton(native).Phase)
		})
	}
}

// testObjects returns a minimal Tekton-backed KubeSphere Pipeline and PipelineRun.
func testObjects() (*devopsv1alpha3.Pipeline, *devopsv1alpha3.PipelineRun) {
	pipeline := &devopsv1alpha3.Pipeline{
		TypeMeta: metav1.TypeMeta{APIVersion: devopsv1alpha3.GroupVersion.String(), Kind: "Pipeline"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "hello-world",
			Namespace:   "demo",
			Annotations: map[string]string{},
		},
		Spec: devopsv1alpha3.PipelineSpec{
			Engine: &devopsv1alpha3.PipelineEngineSpec{Type: devopsv1alpha3.PipelineEngineTekton},
			Type:   devopsv1alpha3.NoScmPipelineType,
			Pipeline: &devopsv1alpha3.NoScmPipeline{
				Name: "hello-world",
			},
		},
	}
	run := &devopsv1alpha3.PipelineRun{
		TypeMeta: metav1.TypeMeta{APIVersion: devopsv1alpha3.GroupVersion.String(), Kind: "PipelineRun"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "hello-world-1",
			Namespace:   "demo",
			UID:         types.UID("run-uid"),
			Annotations: map[string]string{},
		},
		Spec: devopsv1alpha3.PipelineRunSpec{
			PipelineRef:  &corev1.ObjectReference{Name: pipeline.Name},
			PipelineSpec: pipeline.Spec.DeepCopy(),
		},
	}
	return pipeline, run
}

// newTestReconciler creates a fake-client-backed Tekton reconciler.
func newTestReconciler(t *testing.T, objects ...client.Object) *Reconciler {
	scheme := runtime.NewScheme()
	require.NoError(t, devopsv1alpha3.AddToScheme(scheme))
	scheme.AddKnownTypeWithName(tektonPipelineRunGVK, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(tektonPipelineRunGVK.GroupVersion().WithKind("PipelineRunList"), &unstructured.UnstructuredList{})
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&devopsv1alpha3.PipelineRun{}).
		WithObjects(objects...).
		Build()
	return &Reconciler{Client: fakeClient, PollInterval: time.Millisecond}
}

// mustNestedString reads a string from maps and array indices in test assertions.
func mustNestedString(t *testing.T, object interface{}, path ...string) string {
	current := object
	for _, segment := range path {
		switch value := current.(type) {
		case map[string]interface{}:
			current = value[segment]
		case []interface{}:
			index, err := strconv.Atoi(segment)
			require.NoError(t, err)
			require.Less(t, index, len(value))
			current = value[index]
		default:
			t.Fatalf("path %v cannot traverse %T", path, current)
		}
	}
	result, ok := current.(string)
	require.True(t, ok, "path %v yielded %T", path, current)
	return result
}
