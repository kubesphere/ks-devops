package jclient

import "kubesphere.io/devops/pkg/client/devops"

// GetProjectPipelineBuildByType returns a build
func (j *JenkinsClient) GetProjectPipelineBuildByType(projectID, pipelineID string, status string) (*devops.Build, error) {
	return j.jenkins.GetProjectPipelineBuildByType(projectID, pipelineID, status)
}

// GetMultiBranchPipelineBuildByType returns a build
func (j *JenkinsClient) GetMultiBranchPipelineBuildByType(projectID, pipelineID, branch string, status string) (*devops.Build, error) {
	return j.jenkins.GetMultiBranchPipelineBuildByType(projectID, pipelineID, branch, status)
}
