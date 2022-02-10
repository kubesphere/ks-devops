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

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"kubesphere.io/devops/pkg/models/pipeline"
)

type branchPredicate func(pipeline.Branch) bool

type branchSlice []pipeline.Branch

func (branches branchSlice) filter(predicate branchPredicate) []pipeline.Branch {
	var resultBranches []pipeline.Branch
	for _, branch := range branches {
		if predicate != nil && predicate(branch) {
			resultBranches = append(resultBranches, branch)
		}
	}
	return resultBranches
}

func (branches branchSlice) toGenericSlice() []interface{} {
	genericBranches := make([]interface{}, 0, len(branches))
	for i := range branches {
		genericBranches = append(genericBranches, branches[i])
	}
	return genericBranches
}

func filterBranches(branches []pipeline.Branch, filter job.Filter) []pipeline.Branch {
	var predicate branchPredicate
	switch filter {
	case job.PullRequestFilter:
		predicate = func(branch pipeline.Branch) bool {
			return branch.PullRequest != nil && branch.PullRequest.ID != ""
		}
	case job.OriginFilter:
		predicate = func(branch pipeline.Branch) bool {
			return branch.PullRequest == nil || branch.PullRequest.ID == ""
		}
	default:
		predicate = func(pb pipeline.Branch) bool {
			return true
		}
	}
	return branchSlice(branches).filter(predicate)
}
