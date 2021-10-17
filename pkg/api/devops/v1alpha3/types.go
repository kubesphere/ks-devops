package v1alpha3

import "github.com/jenkins-zh/jenkins-client/pkg/job"

// NodeDetail contains metadata of node and an array of steps.
type NodeDetail struct {
	job.Node
	Steps []Step `json:"steps,omitempty"`
}

// Step conatains metadata of step with approvable.
type Step struct {
	job.Step
	// Approvable is a transient field for different users and should not be persisted.
	Approvable bool `json:"approvable"`
}

// NodeDetails is alias of NodeDetail slice.
type NodeDetails []NodeDetail

// InitApprovable inits approvable field of every step of every NodeDetail.
func (nodeDetails NodeDetails) InitApprovable() {
	for i := range nodeDetails {
		Steps(nodeDetails[i].Steps).InitApprovable()
	}
}

// Steps is alias of Step slice.
type Steps []Step

// InitApprovable inits approvable field of every Step.
func (steps Steps) InitApprovable() {
	for i := range steps {
		steps[i].Approvable = true
	}
}
