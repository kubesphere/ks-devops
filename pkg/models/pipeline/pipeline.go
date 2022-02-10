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

package pipeline

import "github.com/jenkins-zh/jenkins-client/pkg/job"

// Metadata holds some of pipeline fields that are only things we needed instead of whole job.Pipeline.
type Metadata struct {
	WeatherScore                   int                       `json:"weatherScore"`
	EstimatedDurationInMillis      int64                     `json:"estimatedDurationInMillis,omitempty"`
	Parameters                     []job.ParameterDefinition `json:"parameters,omitempty"`
	Name                           string                    `json:"name,omitempty"`
	Disabled                       bool                      `json:"disabled,omitempty"`
	NumberOfPipelines              int                       `json:"numberOfPipelines,omitempty"`
	NumberOfFolders                int                       `json:"numberOfFolders,omitempty"`
	PipelineFolderNames            []string                  `json:"pipelineFolderNames,omitempty"`
	TotalNumberOfBranches          int                       `json:"totalNumberOfBranches,omitempty"`
	NumberOfFailingBranches        int                       `json:"numberOfFailingBranches,omitempty"`
	NumberOfSuccessfulBranches     int                       `json:"numberOfSuccessfulBranches,omitempty"`
	TotalNumberOfPullRequests      int                       `json:"totalNumberOfPullRequests,omitempty"`
	NumberOfFailingPullRequests    int                       `json:"numberOfFailingPullRequests,omitempty"`
	NumberOfSuccessfulPullRequests int                       `json:"numberOfSuccessfulPullRequests,omitempty"`
	BranchNames                    []string                  `json:"branchNames,omitempty"`
	SCMSource                      *job.SCMSource            `json:"scmSource,omitempty"`
	ScriptPath                     string                    `json:"scriptPath,omitempty"`
}

// Branch contains branch metadata, like branch and pull request, and latest PipelineRun.
type Branch struct {
	// Name is branch name, like "feat%2FfeatureA"
	Name string `json:"name,omitempty"`
	// RawName is the display branch name, like "feat/featureA"
	RawName      string                    `json:"rawName,omitempty"`
	WeatherScore int                       `json:"weatherScore"`
	Disabled     bool                      `json:"disabled,omitempty"`
	LatestRun    *LatestRun                `json:"latestRun,omitempty"`
	Branch       *job.Branch               `json:"branch,omitempty"`
	PullRequest  *job.PullRequest          `json:"pullRequest,omitempty"`
	Parameters   []job.ParameterDefinition `json:"parameters,omitempty"`
}

// BranchSlice is alias of branch slice.
type BranchSlice []Branch

// SearchByName searchs branch by its name.
func (branches BranchSlice) SearchByName(name string) (bool, *Branch) {
	i := 0
	for ; i < len(branches); i++ {
		if branches[i].Name == name {
			break
		}
	}
	if i == len(branches) {
		return false, nil
	}
	return true, &branches[i]
}

// LatestRun contains metadata of latest PipelineRun.
type LatestRun struct {
	Causes           []Cause  `json:"causes,omitempty"`
	EndTime          job.Time `json:"endTime,omitempty"`
	DurationInMillis *int64   `json:"durationInMillis,omitempty"`
	StartTime        job.Time `json:"startTime,omitempty"`
	ID               string   `json:"id,omitempty"`
	Name             string   `json:"name,omitempty"`
	Pipeline         string   `json:"pipeline,omitempty"`
	Result           string   `json:"result,omitempty"`
	State            string   `json:"state,omitempty"`
}

// Cause contains short description of cause.
type Cause struct {
	ShortDescription string `json:"shortDescription,omitempty"`
}
