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

package v1alpha3

import (
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineRunFinalizerName is the name of PipelineRun finalizer
const PipelineRunFinalizerName = "pipelinerun.finalizers.kubesphere.io"

// PipelineRunSpec defines the desired state of PipelineRun
type PipelineRunSpec struct {
	// PipelineRef is the Pipeline to which the current PipelineRun belongs
	PipelineRef *v1.ObjectReference `json:"pipelineRef"`

	// PipelineSpec is the specification of Pipeline when the current PipelineRun is created.
	// +optional
	PipelineSpec *PipelineSpec `json:"pipelineSpec,omitempty"`

	// Parameters are some key/value pairs passed to runner.
	// +optional
	Parameters []Parameter `json:"parameters,omitempty"`

	// SCM is a SCM configuration that target PipelineRun requires.
	// +optional
	SCM *SCM `json:"scm,omitempty"`

	// Action indicates what we need to do with current PipelineRun.
	// +optional
	Action *Action `json:"action,omitempty"`
}

// PipelineRunStatus defines the observed state of PipelineRun
type PipelineRunStatus struct {
	// Start timestamp of the PipelineRun.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Completion timestamp of the PipelineRun.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Update timestamp of the PipelineRun.
	// +optional
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`

	// Current state of PipelineRun.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []Condition `json:"conditions,omitempty"`

	// Current phase of PipelineRun.
	// +optional
	Phase RunPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="The phase of a PipelineRun"

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

func (status *PipelineRunStatus) GetLatestCondition() *Condition {
	if len(status.Conditions) == 0 {
		return nil
	}
	return &status.Conditions[0]
}

func (status *PipelineRunStatus) AddCondition(newCondition *Condition) {
	// compare newCondition
	var typeExist bool
	for i, condition := range status.Conditions {
		if condition.Type == newCondition.Type {
			// replace with new condition as same condition type
			typeExist = true
			status.Conditions[i] = *newCondition
			break
		}
	}
	if !typeExist {
		newCondition.LastTransitionTime = metav1.Now()
		status.Conditions = append(status.Conditions, *newCondition)
	}
	// sort conditions
	sort.Slice(status.Conditions, func(i, j int) bool {
		// desc order by last probe time
		return !status.Conditions[i].LastProbeTime.Equal(&status.Conditions[j].LastProbeTime) &&
			!status.Conditions[i].LastProbeTime.Before(&status.Conditions[j].LastProbeTime)
	})
}

func (status *PipelineRunStatus) MarkCompleted(endTime time.Time) {
	completionTime := metav1.NewTime(endTime)
	status.CompletionTime = &completionTime
}

// HasStarted indicates if the PipelineRun has started already.
func (pr *PipelineRun) HasStarted() bool {
	_, ok := pr.GetPipelineRunID()
	return ok
}

// HasCompleted indicates if the PipelineRun has already completed.
func (pr *PipelineRun) HasCompleted() bool {
	return !pr.Status.CompletionTime.IsZero()
}

// LabelAsAnOrphan labels PipelineRun as an orphan.
func (pr *PipelineRun) LabelAsAnOrphan() {
	if pr == nil {
		return
	}
	if pr.Labels == nil {
		pr.Labels = make(map[string]string)
	}
	pr.Labels[PipelineRunOrphanKey] = "true"
}

// Buildable returns true if the PipelineRun is buildable, false otherwise.
func (pr *PipelineRun) Buildable() bool {
	return !pr.HasCompleted() && pr.Labels[PipelineRunOrphanKey] != "true"
}

// IsMultiBranchPipeline indicates if the PipelineRun belongs a multi-branch pipeline.
func (prSpec *PipelineRunSpec) IsMultiBranchPipeline() bool {
	return prSpec.PipelineSpec != nil && prSpec.PipelineSpec.Type == MultiBranchPipelineType
}

func (pr *PipelineRun) GetPipelineRunID() (pipelineRunID string, exist bool) {
	pipelineRunID, exist = pr.Annotations[JenkinsPipelineRunIDKey]
	return
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

// SCM is a SCM configuration that target PipelineRun requires.
type SCM struct {
	// RefType indicates that SCM reference type, such as branch, tag, pr, mr.
	RefType RefType `json:"refType"`

	// RefName indicates that SCM reference name, such as master, dev, release-v1.
	RefName string `json:"refName"`
}

// RunPhase is a label for the condition of a PipelineRun at the current time.
type RunPhase string

const (
	// Pending indicates that the PipelineRun is pending.
	Pending RunPhase = "Pending"

	// Running indicates that the PipelineRun is running.
	Running RunPhase = "Running"

	// Succeeded indicates that the PipelineRun has succeeded.
	Succeeded RunPhase = "Succeeded"

	// Failed indicates that the PipelineRun has failed.
	Failed RunPhase = "Failed"

	// Unknown indicates that the PipelineRun has an unknown status.
	Unknown RunPhase = "Unknown"
)

// ConditionType is type of PipelineRun condition.
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

// Condition contains details for the current condition of this PipelineRun.
// Reference from PodCondition
type Condition struct {
	// Type is the type of the condition.
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ConditionType"`

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

// Action indicates what we need to do with current PipelineRun.
type Action string

const (
	// Stop indicates we need to stop the current PipelineRun.
	Stop Action = "Stop"
	// Pause indicates we need to pause the current PipelineRun.
	Pause Action = "Pause"
	// Resume indicates we need to resume the current PipelineRun.
	Resume Action = "Resume"
)

// Valid values for event reasons (new reasons could be added in future)
const (
	// Started indicates PipelineRun has been triggered
	Started string = "Started"
	// Updated indicates PipelineRun's running data has been updated
	Updated string = "Updated"
	// TriggerFailed indicates that it failed to trigger build API
	TriggerFailed string = "TriggerFailed"
	// RetrieveFailed indicates that it failed to retrieve the latest running data
	RetrieveFailed string = "RetrieveFailed"
)

func init() {
	SchemeBuilder.Register(&PipelineRun{}, &PipelineRunList{})
}
