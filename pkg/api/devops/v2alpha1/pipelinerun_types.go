/*
Copyright 2020 The KubeSphere Authors.

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

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PipelineRunFinalizerName = "tknpipelinerun.finalizers.kubesphere.io"

// PipelineRunSpec defines the desired state of PipelineRun
type PipelineRunSpec struct {
	Name        string `json:"name,omitempty"`
	PipelineRef string `json:"ref,omitempty"`
}

// PipelineRunStatus defines the observed state of PipelineRun
type PipelineRunStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// PipelineRun is the Schema for the pipelineruns API
type PipelineRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineRunSpec   `json:"spec,omitempty"`
	Status PipelineRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineRunList contains a list of PipelineRun
type PipelineRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineRun{}, &PipelineRunList{})
}
