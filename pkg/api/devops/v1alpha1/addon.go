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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddonSpec is the specification of an Addon
type AddonSpec struct {
	ExternalAddress string                  `json:"externalAddress,omitempty"`
	Version         string                  `json:"version,omitempty"`
	Strategy        v1.LocalObjectReference `json:"strategy,omitempty"`
	Parameters      map[string]string       `json:"parameters,omitempty"`
}

// AddonStatus represents the status of an addon
type AddonStatus struct {
	Phase string `json:"phase,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The version of target addon"
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// Addon represents a plugin (addon) of ks-devops
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonSpec   `json:"spec,omitempty"`
	Status AddonStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AddonList contains a list of AddonStrategy
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addon `json:"items"`
}

// AddonFinalizerName is the name of Addone finalizer
const AddonFinalizerName = "addon.finalizers.kubesphere.io"

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
