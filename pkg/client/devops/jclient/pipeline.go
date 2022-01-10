package jclient

import (
	"net/http"

	"kubesphere.io/devops/pkg/client/devops"
)

// GetPipeline returns a pipeline
func (j *JenkinsClient) GetPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.Pipeline, error) {
	return j.jenkins.GetPipeline(projectName, pipelineName, httpParameters)
}

// ListPipelines lists the pipelines
func (j *JenkinsClient) ListPipelines(httpParameters *devops.HttpParameters) (*devops.PipelineList, error) {
	return j.jenkins.ListPipelines(httpParameters)
}

// GetPipelineRun returns a pipeline run
func (j *JenkinsClient) GetPipelineRun(projectName, pipelineName, runID string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error) {
	return j.jenkins.GetPipelineRun(projectName, pipelineName, runID, httpParameters)
}

// ListPipelineRuns returns the pipeline runs
func (j *JenkinsClient) ListPipelineRuns(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.PipelineRunList, error) {
	return j.jenkins.ListPipelineRuns(projectName, pipelineName, httpParameters)
}

// StopPipeline stops a pipeline run
func (j *JenkinsClient) StopPipeline(projectName, pipelineName, runID string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error) {
	return j.jenkins.StopPipeline(projectName, pipelineName, runID, httpParameters)
}

// ReplayPipeline replays a pipeline
func (j *JenkinsClient) ReplayPipeline(projectName, pipelineName, runID string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error) {
	return j.jenkins.ReplayPipeline(projectName, pipelineName, runID, httpParameters)
}

// RunPipeline triggers a pipeline
func (j *JenkinsClient) RunPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error) {
	return j.jenkins.RunPipeline(projectName, pipelineName, httpParameters)
}

// GetArtifacts returns the artifacts of a pipeline run
func (j *JenkinsClient) GetArtifacts(projectName, pipelineName, runID string, httpParameters *devops.HttpParameters) ([]devops.Artifacts, error) {
	return j.jenkins.GetArtifacts(projectName, pipelineName, runID, httpParameters)
}

// GetRunLog returns the log output of a pipeline run
func (j *JenkinsClient) GetRunLog(projectName, pipelineName, runID string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.GetRunLog(projectName, pipelineName, runID, httpParameters)
}

// GetStepLog returns the log output of a step
func (j *JenkinsClient) GetStepLog(projectName, pipelineName, runID, nodeID, stepID string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	return j.jenkins.GetStepLog(projectName, pipelineName, runID, nodeID, stepID, httpParameters)
}

// GetNodeSteps returns the node steps
func (j *JenkinsClient) GetNodeSteps(projectName, pipelineName, runID, nodeID string, httpParameters *devops.HttpParameters) ([]devops.NodeSteps, error) {
	return j.jenkins.GetNodeSteps(projectName, pipelineName, runID, nodeID, httpParameters)
}

// GetPipelineRunNodes returns the nodes of a pipeline run
func (j *JenkinsClient) GetPipelineRunNodes(projectName, pipelineName, runID string, httpParameters *devops.HttpParameters) ([]devops.PipelineRunNodes, error) {
	return j.jenkins.GetPipelineRunNodes(projectName, pipelineName, runID, httpParameters)
}

// SubmitInputStep submits a pipeline input step
func (j *JenkinsClient) SubmitInputStep(projectName, pipelineName, runID, nodeID, stepID string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.SubmitInputStep(projectName, pipelineName, runID, nodeID, stepID, httpParameters)
}

// GetBranchPipeline returns the branch pipeline
func (j *JenkinsClient) GetBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.BranchPipeline, error) {
	return j.jenkins.GetBranchPipeline(projectName, pipelineName, branchName, httpParameters)
}

// GetBranchPipelineRun returns the pipeline run
func (j *JenkinsClient) GetBranchPipelineRun(projectName, pipelineName, branchName, runID string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error) {
	return j.jenkins.GetBranchPipelineRun(projectName, pipelineName, branchName, runID, httpParameters)
}

// StopBranchPipeline stops a pipeline run
func (j *JenkinsClient) StopBranchPipeline(projectName, pipelineName, branchName, runID string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error) {
	return j.jenkins.StopBranchPipeline(projectName, pipelineName, branchName, runID, httpParameters)
}

// ReplayBranchPipeline replays a pipeline
func (j *JenkinsClient) ReplayBranchPipeline(projectName, pipelineName, branchName, runID string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error) {
	return j.jenkins.ReplayBranchPipeline(projectName, pipelineName, branchName, runID, httpParameters)
}

// RunBranchPipeline triggers a pipeline
func (j *JenkinsClient) RunBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error) {
	return j.jenkins.RunBranchPipeline(projectName, pipelineName, branchName, httpParameters)
}

// GetBranchArtifacts returns the artifacts of a pipeline run
func (j *JenkinsClient) GetBranchArtifacts(projectName, pipelineName, branchName, runID string, httpParameters *devops.HttpParameters) ([]devops.Artifacts, error) {
	return j.jenkins.GetBranchArtifacts(projectName, pipelineName, branchName, runID, httpParameters)
}

// GetBranchRunLog returns the pipeline run log
func (j *JenkinsClient) GetBranchRunLog(projectName, pipelineName, branchName, runID string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.GetBranchRunLog(projectName, pipelineName, branchName, runID, httpParameters)
}

// GetBranchStepLog returns the log output of a pipeline step
func (j *JenkinsClient) GetBranchStepLog(projectName, pipelineName, branchName, runID, nodeID, stepID string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	return j.jenkins.GetBranchStepLog(projectName, pipelineName, branchName, runID, nodeID, stepID, httpParameters)
}

// GetBranchNodeSteps returns the node steps
func (j *JenkinsClient) GetBranchNodeSteps(projectName, pipelineName, branchName, runID, nodeID string, httpParameters *devops.HttpParameters) ([]devops.NodeSteps, error) {
	return j.jenkins.GetBranchNodeSteps(projectName, pipelineName, branchName, runID, nodeID, httpParameters)
}

// GetBranchPipelineRunNodes returns the pipeline run nodes
func (j *JenkinsClient) GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runID string, httpParameters *devops.HttpParameters) ([]devops.BranchPipelineRunNodes, error) {
	return j.jenkins.GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runID, httpParameters)
}

// SubmitBranchInputStep submits a pipeline input step
func (j *JenkinsClient) SubmitBranchInputStep(projectName, pipelineName, branchName, runID, nodeID, stepID string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.SubmitBranchInputStep(projectName, pipelineName, branchName, runID, nodeID, stepID, httpParameters)
}

// GetPipelineBranch returns PipelineBranch
func (j *JenkinsClient) GetPipelineBranch(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.PipelineBranch, error) {
	return j.jenkins.GetPipelineBranch(projectName, pipelineName, httpParameters)
}

// ScanBranch scans a pipeline
func (j *JenkinsClient) ScanBranch(projectName, pipelineName string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.ScanBranch(projectName, pipelineName, httpParameters)
}

// GetConsoleLog returns the log output
func (j *JenkinsClient) GetConsoleLog(projectName, pipelineName string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.GetConsoleLog(projectName, pipelineName, httpParameters)
}

// GetCrumb returns the crumb
func (j *JenkinsClient) GetCrumb(httpParameters *devops.HttpParameters) (*devops.Crumb, error) {
	return j.jenkins.GetCrumb(httpParameters)
}

// GetSCMServers returns the scm servers
func (j *JenkinsClient) GetSCMServers(scmID string, httpParameters *devops.HttpParameters) ([]devops.SCMServer, error) {
	return j.jenkins.GetSCMServers(scmID, httpParameters)
}

// GetSCMOrg returns the scm org
func (j *JenkinsClient) GetSCMOrg(scmID string, httpParameters *devops.HttpParameters) ([]devops.SCMOrg, error) {
	return j.jenkins.GetSCMOrg(scmID, httpParameters)
}

// GetOrgRepo returns the org and repo
func (j *JenkinsClient) GetOrgRepo(scmID, organizationID string, httpParameters *devops.HttpParameters) (devops.OrgRepo, error) {
	return j.jenkins.GetOrgRepo(scmID, organizationID, httpParameters)
}

// CreateSCMServers creates a scm server
func (j *JenkinsClient) CreateSCMServers(scmID string, httpParameters *devops.HttpParameters) (*devops.SCMServer, error) {
	return j.jenkins.CreateSCMServers(scmID, httpParameters)
}

// GetNotifyCommit returns notify commit
func (j *JenkinsClient) GetNotifyCommit(httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.GetNotifyCommit(httpParameters)
}

// GithubWebhook trigger a github webhook
func (j *JenkinsClient) GithubWebhook(httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.GithubWebhook(httpParameters)
}

// GenericWebhook triggers a generic webhook
func (j *JenkinsClient) GenericWebhook(httpParameters *devops.HttpParameters) ([]byte, error) {
	return j.jenkins.GenericWebhook(httpParameters)
}

// Validate does the validation check
func (j *JenkinsClient) Validate(scmID string, httpParameters *devops.HttpParameters) (*devops.Validates, error) {
	return j.jenkins.Validate(scmID, httpParameters)
}

// CheckScriptCompile checks the script compile
func (j *JenkinsClient) CheckScriptCompile(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.CheckScript, error) {
	return j.jenkins.CheckScriptCompile(projectName, pipelineName, httpParameters)
}

// CheckCron does the cron check
func (j *JenkinsClient) CheckCron(projectName string, httpParameters *devops.HttpParameters) (*devops.CheckCronRes, error) {
	return j.jenkins.CheckCron(projectName, httpParameters)
}

// ToJenkinsfile sends a request to turn to jenkinsfile
func (j *JenkinsClient) ToJenkinsfile(httpParameters *devops.HttpParameters) (*devops.ResJenkinsfile, error) {
	return j.jenkins.ToJenkinsfile(httpParameters)
}

// ToJSON sends a request to turn JSOn
func (j *JenkinsClient) ToJSON(httpParameters *devops.HttpParameters) (map[string]interface{}, error) {
	return j.jenkins.ToJSON(httpParameters)
}
