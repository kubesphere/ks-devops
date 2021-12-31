/*
Copyright 2021 The KubeSphere Authors.

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

// AnnotationKeyGitRepos are the references of target git repositories
const AnnotationKeyGitRepos = "devops.kubesphere.io/git-repositories"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Webhook is the Schema for the webhook API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Server",type="string",JSONPath=".spec.server"
// +kubebuilder:printcolumn:name="SkipVerify",type="boolean",JSONPath=".spec.skipVerify"
type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WebhookSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebhookList contains a list of Webhook
type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webhook `json:"items"`
}

// WebhookSpec represents the desired state of a Webhook
type WebhookSpec struct {
	Server     string              `json:"server"`
	Secret     *v1.SecretReference `json:"secret,omitempty"`
	Events     []string            `json:"events,omitempty"`
	SkipVerify bool                `json:"skipVerify"`
}

func init() {
	SchemeBuilder.Register(&Webhook{}, &WebhookList{})
}
