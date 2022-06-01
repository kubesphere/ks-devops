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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StepTemplateSpec defines the desired state of StepTemplate
type StepTemplateSpec struct {
	Secret     SecretInStep      `json:"secret,omitempty"`
	Container  string            `json:"container,omitempty"`
	Runtime    string            `json:"runtime,omitempty"`
	Template   string            `json:"template,omitempty"`
	Parameters []ParameterInStep `json:"parameters,omitempty"`
}

type SecretInStep struct {
	Type    string            `json:"type,omitempty"`
	Wrap    bool              `json:"wrap,omitempty"`
	Mapping map[string]string `json:"mapping,omitempty"`
}

type ParameterInStep struct {
	Name         string `json:"name"`
	Required     bool   `json:"required,omitempty"`
	Display      string `json:"display,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

// StepTemplateStatus defines the observed state of StepTemplate
type StepTemplateStatus struct {
	Phase string `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// StepTemplate is the Schema for the steptemplates API
type StepTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StepTemplateSpec   `json:"spec,omitempty"`
	Status StepTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StepTemplateList contains a list of StepTemplate
type StepTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StepTemplate `json:"items"`
}

// DefaultSecretKeyMapping
var DefaultSecretKeyMapping = map[string]string{
	":": "",
}

func init() {
	SchemeBuilder.Register(&StepTemplate{}, &StepTemplateList{})
}
