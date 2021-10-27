package pipeline

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/apimachinery/pkg/util/validation"
	"kubesphere.io/devops/pkg/models/pipeline"
)

type branchPredicate func(pipeline.Branch) bool

type branchSlice []pipeline.Branch

func (branches branchSlice) filter(predicate branchPredicate) []pipeline.Branch {
	var resultBranches []pipeline.Branch
	for _, branch := range branches {
		if errors := validation.IsValidLabelValue(branch.Name); len(errors) != 0 {
			continue
		}
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
