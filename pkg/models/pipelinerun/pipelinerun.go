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
