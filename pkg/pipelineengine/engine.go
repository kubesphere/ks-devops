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

// Package pipelineengine contains the engine selection contract shared by the Jenkins and Tekton adapters.
package pipelineengine

import (
	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AnnotationEngine selects the execution engine for a KubeSphere Pipeline or PipelineRun.
	AnnotationEngine = "devops.kubesphere.io/pipeline-engine"
	// AnnotationTektonPipeline overrides the native Tekton Pipeline name.
	AnnotationTektonPipeline = "devops.kubesphere.io/tekton-pipeline"
	// AnnotationTektonPipelineRun records the native Tekton PipelineRun name.
	AnnotationTektonPipelineRun = "devops.kubesphere.io/tekton-pipelinerun"
	// AnnotationTektonServiceAccount selects the ServiceAccount used by Tekton TaskRuns.
	AnnotationTektonServiceAccount = "devops.kubesphere.io/tekton-service-account"
	// AnnotationTektonWorkspaceName selects a workspace declared by the native Tekton Pipeline.
	AnnotationTektonWorkspaceName = "devops.kubesphere.io/tekton-workspace-name"
	// AnnotationTektonWorkspaceClaim binds the selected workspace to a PersistentVolumeClaim.
	AnnotationTektonWorkspaceClaim = "devops.kubesphere.io/tekton-workspace-claim"
	// AnnotationTektonWorkspaceEmptyDir binds the selected workspace to an ephemeral emptyDir volume.
	AnnotationTektonWorkspaceEmptyDir = "devops.kubesphere.io/tekton-workspace-empty-dir"
	// AnnotationTektonTimeout configures the native Tekton PipelineRun pipeline timeout.
	AnnotationTektonTimeout = "devops.kubesphere.io/tekton-timeout"

	// EngineJenkins identifies the existing Jenkins execution engine.
	EngineJenkins = string(devopsv1alpha3.PipelineEngineJenkins)
	// EngineTekton identifies the Tekton execution engine.
	EngineTekton = string(devopsv1alpha3.PipelineEngineTekton)
)

var propagatedAnnotations = []string{
	AnnotationEngine,
	AnnotationTektonPipeline,
	AnnotationTektonServiceAccount,
	AnnotationTektonWorkspaceName,
	AnnotationTektonWorkspaceClaim,
	AnnotationTektonWorkspaceEmptyDir,
	AnnotationTektonTimeout,
}

// Name returns the spec-selected engine, falls back to the compatibility annotation, and finally defaults to Jenkins.
func Name(object metav1.Object) string {
	if object == nil {
		return EngineJenkins
	}
	if engine := nameFromSpec(object); engine != "" {
		return engine
	}
	if engine := object.GetAnnotations()[AnnotationEngine]; engine != "" {
		return engine
	}
	return EngineJenkins
}

// nameFromSpec returns an engine explicitly stored in a Pipeline or PipelineRun snapshot.
func nameFromSpec(object metav1.Object) string {
	switch typed := object.(type) {
	case *devopsv1alpha3.Pipeline:
		if typed.Spec.Engine != nil {
			return string(typed.Spec.Engine.Type)
		}
	case *devopsv1alpha3.PipelineRun:
		if typed.Spec.PipelineSpec != nil && typed.Spec.PipelineSpec.Engine != nil {
			return string(typed.Spec.PipelineSpec.Engine.Type)
		}
	}
	return ""
}

// IsTekton reports whether an object explicitly selects the Tekton engine.
func IsTekton(object metav1.Object) bool {
	return Name(object) == EngineTekton
}

// ResolveAnnotation returns a run-level value first and falls back to the Pipeline value.
func ResolveAnnotation(run metav1.Object, pipeline metav1.Object, key string) string {
	if run != nil {
		if value := run.GetAnnotations()[key]; value != "" {
			return value
		}
	}
	if pipeline != nil {
		return pipeline.GetAnnotations()[key]
	}
	return ""
}

// PropagatePipelineAnnotations copies the engine contract from a Pipeline to a newly created PipelineRun.
func PropagatePipelineAnnotations(pipeline metav1.Object, run metav1.Object) {
	if pipeline == nil || run == nil {
		return
	}
	specEngine := nameFromSpec(pipeline)
	annotations := run.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	for _, key := range propagatedAnnotations {
		if key == AnnotationEngine && specEngine != "" {
			continue
		}
		if value := pipeline.GetAnnotations()[key]; value != "" {
			annotations[key] = value
		}
	}
	run.SetAnnotations(annotations)
}
