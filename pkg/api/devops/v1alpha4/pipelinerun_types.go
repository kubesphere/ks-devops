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

package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/apis"
)

// PipelineRunSpec defines the desired state of PipelineRun
type PipelineRunSpec struct {
	// Parameters are some key/value pairs passed to runner.
	// +optional
	Parameters []*Parameter `json:"parameters,omitempty"`

	// SCM is a SCM configuration that target pipeline run requires.
	SCM *SCM `json:"scm,omitempty"`
}

// PipelineRunStatus defines the observed state of PipelineRun
type PipelineRunStatus struct {
	// Start timestamp of the pipeline run.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Completion timestamp of the pipeline run.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Update timestamp of the pipeline run.
	// +optional
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`

	// Current state of pipeline run.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []Condition `json:"conditions,omitempty"`

	// Current phase of pipeline run.
	// +optional
	Phase RunPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

// Parameter is an option that can be passed with the endpoint to influence the Pipeline Run
type Parameter struct {
	// Name indicates that name of the parameter.
	Name string `json:"name" description:"parameter name"`

	// Value indicates that value of the parameter.
	Value string `json:"value" description:"parameter value"`
}

// RefType indicates that SCM reference type, such as branch, tag, pr, mr.
type RefType string

const (
	// Branch represents an independent line of development.
	Branch RefType = "branch"

	// Tag points to specific points in SCM history.
	Tag RefType = "tag"

	// PullRequest is a feature that makes it easier for developers to collaborate using GitHub, BitBucket or others.
	PullRequest RefType = "pr"

	// MergeRequest is a request from someone to merge in code from one branch to another
	MergeRequest RefType = "mr"
)

// SCM is a SCM configuration that target pipeline run requires.
type SCM struct {
	// RefType indicates that SCM reference type, such as branch, tag, pr, mr.
	RefType RefType `json:"refType"`

	// RefName indicates that SCM reference name, such as master, dev, release-v1.
	RefName string `json:"refName"`
}

// RunPhase is a label for the condition of a pipeline run at the current time.
type RunPhase string

const (
	// Pending indicates that the pipeline run is pending.
	Pending RunPhase = "Pending"

	// Running indicates that the pipeline run is running.
	Running RunPhase = "Running"

	// Succeeded indicates that the pipeline run has succeeded.
	Succeeded RunPhase = "Succeeded"

	// Failed indicates that the pipeline run has failed.
	Failed RunPhase = "Failed"

	// Unknown indicates that the pipeline run has an unknown status.
	Unknown RunPhase = "Unknown"
)

// ConditionType is type of pipeline run condition.
type ConditionType string

const (
	// ConditionReady indicates that the pipeline is ready.
	// For long-running pipeline
	ConditionReady ConditionType = "Ready"

	// ConditionSucceeded indicates that the pipeline has finished.
	// For pipeline which runs to completion
	ConditionSucceeded ConditionType = "Succeeded"
)

// ConditionStatus is the status of the current condition.
type ConditionStatus string

const (
	// ConditionTrue means a resource is in the condition.
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse means a resource is not in the condition.
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown mean kubernetes can't decide if a resource is in the condition or not.
	ConditionUnknown ConditionStatus = "Unknown"
)

// Condition contains details for the current condition of this pipeline run.
// Reference from PodCondition
type Condition struct {
	// Type is the type of the condition.
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=PodConditionType"`

	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`

	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`

	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`

	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`

	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

func init() {
	SchemeBuilder.Register(&PipelineRun{}, &PipelineRunList{})
	apis.AddToSchemes = append(apis.AddToSchemes, AddToScheme)
}
