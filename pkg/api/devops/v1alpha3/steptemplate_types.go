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

// StepTemplateSpec defines the desired state of ClusterStepTemplate
type StepTemplateSpec struct {
	Secret     SecretInStep      `json:"secret,omitempty"`
	Container  string            `json:"container,omitempty"`
	Runtime    string            `json:"runtime,omitempty"`
	Template   string            `json:"template,omitempty"`
	Parameters []ParameterInStep `json:"parameters,omitempty"`
}

// SecretInStep is the secret which used in a step
type SecretInStep struct {
	Type    string            `json:"type,omitempty"`
	Wrap    bool              `json:"wrap,omitempty"`
	Mapping map[string]string `json:"mapping,omitempty"`
}

// ParameterInStep is the parameter which used in a step
type ParameterInStep struct {
	Name         string        `json:"name"`
	Type         ParameterType `json:"type,omitempty"`
	Required     bool          `json:"required,omitempty"`
	Display      string        `json:"display,omitempty"`
	DefaultValue string        `json:"defaultValue,omitempty" yaml:"defaultValue"`
}

// ParameterType represents the type of parameter
type ParameterType string

const (
	// ParameterTypeString represents a parameter in string format, expect this is a single line
	ParameterTypeString ParameterType = "string"
	// ParameterTypeText represents a parameter in string format, expect this is multi-line
	ParameterTypeText ParameterType = "text"
	// ParameterTypeCode represents a parameter in string format that contains some code lines
	ParameterTypeCode ParameterType = "code"
	// ParameterTypeBool represents a parameter in boolean format
	ParameterTypeBool ParameterType = "bool"
	// ParameterTypeEnum represents a parameter in enum format
	ParameterTypeEnum ParameterType = "enum"
)

// StepTemplateStatus defines the observed state of ClusterStepTemplate
type StepTemplateStatus struct {
	Phase StepTemplatePhase `json:"phase"`
}

// StepTemplatePhase represents the phase of the Step template
type StepTemplatePhase string

var (
	// StepTemplatePhaseReady indicates the step template is ready to use
	StepTemplatePhaseReady StepTemplatePhase = "ready"
	// StepTemplatePhaseInit indicates the step template is not ready to use
	StepTemplatePhaseInit StepTemplatePhase = "init"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ClusterStepTemplate is the Schema for the steptemplates API
type ClusterStepTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StepTemplateSpec   `json:"spec,omitempty"`
	Status StepTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterStepTemplateList contains a list of ClusterStepTemplate
type ClusterStepTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterStepTemplate `json:"items"`
}

// DefaultSecretKeyMapping mainly used as the Jenkinsfile environment variables
var DefaultSecretKeyMapping = map[string]string{
	"passwordVariable":   "PASSWORDVARIABLE",
	"usernameVariable":   "USERNAMEVARIABLE",
	"variable":           "VARIABLE",
	"sshUserPrivateKey":  "SSHUSERPRIVATEKEY",
	"keyFileVariable":    "KEYFILEVARIABLE",
	"passphraseVariable": "PASSPHRASEVARIABLE",
}

func init() {
	SchemeBuilder.Register(&ClusterStepTemplate{}, &ClusterStepTemplateList{})
}
