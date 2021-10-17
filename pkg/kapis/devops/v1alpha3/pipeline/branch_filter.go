package pipeline

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"kubesphere.io/devops/pkg/models/pipeline"
)

type branchPredicate func(pipeline.PipelineBranch) bool

type branchFilter []pipeline.PipelineBranch

func (branches branchFilter) filter(predicate branchPredicate) []pipeline.PipelineBranch {
	resultBranches := []pipeline.PipelineBranch{}
	for _, branch := range branches {
		if predicate != nil && predicate(branch) {
			resultBranches = append(resultBranches, branch)
		}
	}
	return resultBranches
}

func filterBranches(branches []pipeline.PipelineBranch, filter string) []pipeline.PipelineBranch {
	var predicate branchPredicate
	switch filter {
	case string(job.PullRequestFilter):
		predicate = func(branch pipeline.PipelineBranch) bool {
			return branch.PullRequest != nil && branch.PullRequest.ID != ""
		}
	case string(job.OriginFilter):
		predicate = func(branch pipeline.PipelineBranch) bool {
			return branch.PullRequest == nil || branch.PullRequest.ID == ""
		}
	default:
		predicate = func(pb pipeline.PipelineBranch) bool {
			return true
		}
	}
	return branchFilter(branches).filter(predicate)
}
