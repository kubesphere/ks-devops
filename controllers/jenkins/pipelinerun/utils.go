package pipelinerun

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"kubesphere.io/devops/pkg/api/devops/v1alpha4"
)

type JenkinsRunState string

const (
	Queued        JenkinsRunState = "QUEUED"
	Running       JenkinsRunState = "RUNNING"
	Paused        JenkinsRunState = "PAUSED"
	Skipped       JenkinsRunState = "SKIPPED"
	NotBuiltState JenkinsRunState = "NOT_BUILT"
	Finished      JenkinsRunState = "FINISHED"
)

func (state JenkinsRunState) String() string {
	return string(state)
}

type JenkinsRunResult string

const (
	Success        JenkinsRunResult = "SUCCESS"
	Unstable       JenkinsRunResult = "UNSTABLE"
	Failure        JenkinsRunResult = "FAILURE"
	NotBuiltResult JenkinsRunResult = "NOT_BUILT"
	Unknown        JenkinsRunResult = "UNKNOWN"
	Aborted        JenkinsRunResult = "ABORTED"
)

func (result JenkinsRunResult) String() string {
	return string(result)
}

func generateJobName(projectName, pipelineName string, prSpec *v1alpha4.PipelineRunSpec) string {
	jobName := projectName + " " + pipelineName
	if prSpec != nil && prSpec.IsMultiBranchPipeline() {
		jobName = jobName + " " + prSpec.SCM.RefName
	}
	return jobName
}

func convertParam(params []v1alpha4.Parameter) []job.ParameterDefinition {
	if params == nil {
		return []job.ParameterDefinition{}
	}
	var paramDefs []job.ParameterDefinition
	for _, param := range params {
		paramDefs = append(paramDefs, job.ParameterDefinition{
			Name:  param.Name,
			Value: param.Value,
		})
	}
	return paramDefs
}
