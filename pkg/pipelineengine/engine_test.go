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

package pipelineengine

import (
	"testing"

	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestName verifies the backward-compatible engine default and explicit selection.
func TestName(t *testing.T) {
	assert.Equal(t, EngineJenkins, Name(nil))
	assert.Equal(t, EngineJenkins, Name(&metav1.ObjectMeta{}))
	assert.Equal(t, EngineTekton, Name(&metav1.ObjectMeta{Annotations: map[string]string{AnnotationEngine: EngineTekton}}))

	pipeline := &devopsv1alpha3.Pipeline{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationEngine: EngineJenkins}},
		Spec: devopsv1alpha3.PipelineSpec{Engine: &devopsv1alpha3.PipelineEngineSpec{
			Type: devopsv1alpha3.PipelineEngineTekton,
		}},
	}
	assert.Equal(t, EngineTekton, Name(pipeline), "spec must take precedence over the compatibility annotation")

	run := &devopsv1alpha3.PipelineRun{Spec: devopsv1alpha3.PipelineRunSpec{
		PipelineSpec: pipeline.Spec.DeepCopy(),
	}}
	assert.Equal(t, EngineTekton, Name(run), "a PipelineRun must use its PipelineSpec snapshot")
}

// TestPropagatePipelineAnnotations verifies that only the engine contract is copied to a run.
func TestPropagatePipelineAnnotations(t *testing.T) {
	pipeline := &metav1.ObjectMeta{Annotations: map[string]string{
		AnnotationEngine:               EngineTekton,
		AnnotationTektonPipeline:       "native-pipeline",
		AnnotationTektonServiceAccount: "builder",
		"unrelated":                    "must-not-propagate",
	}}
	run := &metav1.ObjectMeta{Annotations: map[string]string{"existing": "value"}}

	PropagatePipelineAnnotations(pipeline, run)

	assert.Equal(t, EngineTekton, run.Annotations[AnnotationEngine])
	assert.Equal(t, "native-pipeline", run.Annotations[AnnotationTektonPipeline])
	assert.Equal(t, "builder", run.Annotations[AnnotationTektonServiceAccount])
	assert.Equal(t, "value", run.Annotations["existing"])
	assert.NotContains(t, run.Annotations, "unrelated")

	specPipeline := &devopsv1alpha3.Pipeline{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			AnnotationEngine:         EngineJenkins,
			AnnotationTektonPipeline: "native-pipeline",
		}},
		Spec: devopsv1alpha3.PipelineSpec{Engine: &devopsv1alpha3.PipelineEngineSpec{
			Type: devopsv1alpha3.PipelineEngineTekton,
		}},
	}
	specRun := &devopsv1alpha3.PipelineRun{}
	PropagatePipelineAnnotations(specPipeline, specRun)
	assert.NotContains(t, specRun.Annotations, AnnotationEngine)
	assert.Equal(t, "native-pipeline", specRun.Annotations[AnnotationTektonPipeline])
}

// TestResolveAnnotation verifies that PipelineRun settings override Pipeline defaults.
func TestResolveAnnotation(t *testing.T) {
	pipeline := &metav1.ObjectMeta{Annotations: map[string]string{AnnotationTektonTimeout: "30m"}}
	run := &metav1.ObjectMeta{Annotations: map[string]string{AnnotationTektonTimeout: "5m"}}

	assert.Equal(t, "5m", ResolveAnnotation(run, pipeline, AnnotationTektonTimeout))
	assert.Equal(t, "30m", ResolveAnnotation(&metav1.ObjectMeta{}, pipeline, AnnotationTektonTimeout))
}
