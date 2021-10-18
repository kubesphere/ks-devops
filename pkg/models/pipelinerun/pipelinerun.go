package pipelinerun

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
	Approvable bool `json:"approvable,omitempty"`
}
