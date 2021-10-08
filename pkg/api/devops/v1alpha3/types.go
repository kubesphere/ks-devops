package v1alpha3

import "github.com/jenkins-zh/jenkins-client/pkg/job"

// NodeDetail contains metadata of node and an array of steps.
type NodeDetail struct {
	job.Node
	Steps      []job.Step `json:"steps,omitempty"`
	Approvable bool       `json:"approvable,omitempty"`
}
