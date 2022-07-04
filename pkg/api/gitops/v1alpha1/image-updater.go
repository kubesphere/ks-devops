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

// ImageUpdaterSpec is the specification of the image updater
type ImageUpdaterSpec struct {
	// +kubebuilder:default:=argocd
	// +kubebuilder:validation:Enum=argocd;fluxcd
	Kind   string            `json:"kind,omitempty"`
	Images []string          `json:"images,omitempty"`
	Argo   *ArgoImageUpdater `json:"argo,omitempty"`
}

// ArgoImageUpdater is the specification of the Argo image updater
type ArgoImageUpdater struct {
	App v1.LocalObjectReference `json:"app"`
	// +kubebuilder:default:=built-in
	// +kubebuilder:validation:Enum=built-in;git
	Write          WriteMethod       `json:"write,omitempty"`
	UpdateStrategy map[string]string `json:"updateStrategy,omitempty"`
	AllowTags      map[string]string `json:"allowTags,omitempty"`
	IgnoreTags     map[string]string `json:"ignoreTags,omitempty"`
	Platforms      map[string]string `json:"platforms,omitempty"`
	Secrets        map[string]string `json:"secrets,omitempty"`
}

// WriteMethod is an alias of string that represents the write back method of Argo CD Image updater
type WriteMethod string

const (
	// WriteMethodBuiltIn is the default method that only effect in cluster
	WriteMethodBuiltIn WriteMethod = "built-in"
	// WriteMethodGit indicates the changes will be written back to git repository
	WriteMethodGit WriteMethod = "git"
)

// GetValue returns the value that could be recognized by Argo CD Image Updater
func (w WriteMethod) GetValue() string {
	switch w {
	case WriteMethodBuiltIn:
		return "argocd"
	case WriteMethodGit:
		return "git"
	}
	return ""
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Kind",type=string,JSONPath=`.spec.kind`

// ImageUpdater represents an image updating request
type ImageUpdater struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ImageUpdaterSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageUpdaterList represents a set of the applications
type ImageUpdaterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageUpdater `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageUpdater{}, &ImageUpdaterList{})
}
