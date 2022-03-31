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

// AnnotationKeyWebhookUpdates is a signal that should update the webhooks
const AnnotationKeyWebhookUpdates = "devops.kubesphere.io/webhook-updates"

// GitRepoFinalizerName is the finalizer name of the git repository
const GitRepoFinalizerName = "finalizer.gitrepository.devops.kubesphere.io"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitRepository is the Schema for the webhook API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Server",type="string",JSONPath=".spec.server"
type GitRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GitRepositorySpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitRepositoryList contains a list of GitRepository
type GitRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitRepository `json:"items"`
}

// GitRepositorySpec represents the desired state of a GitRepository
type GitRepositorySpec struct {
	Provider string                    `json:"provider,omitempty"`
	URL      string                    `json:"url,omitempty"`
	Secret   *v1.SecretReference       `json:"secret,omitempty"`
	Webhooks []v1.LocalObjectReference `json:"webhooks,omitempty"`
}

func init() {
	SchemeBuilder.Register(&GitRepository{}, &GitRepositoryList{})
}
