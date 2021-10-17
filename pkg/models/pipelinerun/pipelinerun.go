package pipelinerun

import "github.com/jenkins-zh/jenkins-client/pkg/job"

// NodeDetail contains metadata of node and an array of steps.
type NodeDetail struct {
	job.Node
	Steps      []job.Step `json:"steps"`
	Approvable bool       `json:"approvable"`
}
