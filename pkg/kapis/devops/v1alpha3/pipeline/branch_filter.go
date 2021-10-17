package pipeline

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"kubesphere.io/devops/pkg/models/pipeline"
)

type branchPredicate func(pipeline.Branch) bool

type branchFilter []pipeline.Branch

func (branches branchFilter) filter(predicate branchPredicate) []pipeline.Branch {
	resultBranches := []pipeline.Branch{}
	for _, branch := range branches {
		if predicate != nil && predicate(branch) {
			resultBranches = append(resultBranches, branch)
		}
	}
	return resultBranches
}

func filterBranches(branches []pipeline.Branch, filter string) []pipeline.Branch {
	var predicate branchPredicate
	switch filter {
	case string(job.PullRequestFilter):
		predicate = func(branch pipeline.Branch) bool {
			return branch.PullRequest != nil && branch.PullRequest.ID != ""
		}
	case string(job.OriginFilter):
		predicate = func(branch pipeline.Branch) bool {
			return branch.PullRequest == nil || branch.PullRequest.ID == ""
		}
	default:
		predicate = func(pb pipeline.Branch) bool {
			return true
		}
	}
	return branchFilter(branches).filter(predicate)
}
