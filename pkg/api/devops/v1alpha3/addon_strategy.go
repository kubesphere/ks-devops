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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddStrategySpec is the specification of an AddonStrategy
type AddStrategySpec struct {
	Available      bool                 `json:"available,omitempty"`
	Type           AddonInstallStrategy `json:"type,omitempty"`
	YAML           string               `json:"yaml,omitempty"`
	Operator       v1.ObjectReference   `json:"operator,omitempty"`
	SimpleOperator v1.ObjectReference   `json:"simpleOperator,omitempty"`
	HelmRepo       string               `json:"helmRepo,omitempty"`
	Template       string               `json:"template,omitempty"`
	Parameters     map[string]string    `json:"parameters,omitempty"`
}

// AddonInstallStrategy represents the addon installation strategy
type AddonInstallStrategy string

const (
	// AddonInstallStrategySimple represents a single YAML file to install addon
	AddonInstallStrategySimple AddonInstallStrategy = "simple"
	// AddonInstallStrategyHelm represents to install via helm
	AddonInstallStrategyHelm AddonInstallStrategy = "helm"
	// AddonInstallStrategyOperator represents the full operator feature
	AddonInstallStrategyOperator AddonInstallStrategy = "operator"
	// AddonInstallStrategySimpleOperator represents a single Operator image to install addon
	AddonInstallStrategySimpleOperator AddonInstallStrategy = "simple-operator"
)

// IsValid checks if this is valid
func (a AddonInstallStrategy) IsValid() bool {
	switch a {
	case AddonInstallStrategySimple, AddonInstallStrategyHelm,
		AddonInstallStrategyOperator, AddonInstallStrategySimpleOperator:
		return true
	default:
		return false
	}
}

// +genclient:nonNamespaced
// +kubebuilder:resource:scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// AddonStrategy represents an addonStrategy
type AddonStrategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AddStrategySpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AddonStrategyList contains a list of AddonStrategy
type AddonStrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddonStrategy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AddonStrategy{}, &AddonStrategyList{})
}
