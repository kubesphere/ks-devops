package pipeline

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"kubesphere.io/devops/pkg/client/devops/jenkins"
	"kubesphere.io/devops/pkg/models/pipeline"
)

func convertPipeline(jobPipeline *job.Pipeline) *pipeline.Metadata {
	return &pipeline.Metadata{
		WeatherScore:                   jobPipeline.WeatherScore,
		EstimatedDurationInMillis:      jobPipeline.EstimatedDurationInMillis,
		Parameters:                     convertParameterDefinitions(jobPipeline.Parameters),
		Name:                           jobPipeline.Name,
		Disabled:                       jobPipeline.Disabled,
		NumberOfPipelines:              jobPipeline.NumberOfPipelines,
		NumberOfFolders:                jobPipeline.NumberOfFolders,
		PipelineFolderNames:            jobPipeline.PipelineFolderNames,
		TotalNumberOfBranches:          jobPipeline.TotalNumberOfBranches,
		TotalNumberOfPullRequests:      jobPipeline.TotalNumberOfPullRequests,
		NumberOfFailingBranches:        jobPipeline.NumberOfFailingBranches,
		NumberOfFailingPullRequests:    jobPipeline.NumberOfFailingPullRequests,
		NumberOfSuccessfulBranches:     jobPipeline.NumberOfSuccessfulBranches,
		NumberOfSuccessfulPullRequests: jobPipeline.NumberOfSuccessfulPullRequests,
		BranchNames:                    jobPipeline.BranchNames,
		SCMSource:                      jobPipeline.SCMSource,
		ScriptPath:                     jobPipeline.ScriptPath,
	}
}

func convertParameterDefinitions(paramDefs []job.ParameterDefinition) []job.ParameterDefinition {
	newParamDefs := []job.ParameterDefinition{}
	for _, paramDef := range paramDefs {
		// copy the parameter definition
		if simpleType, ok := jenkins.ParameterTypeMap["hudson.model."+paramDef.Type]; ok {
			paramDef.Type = simpleType
		}
		newParamDefs = append(newParamDefs, paramDef)
	}
	return newParamDefs
}

func convertLatestRun(jobLatestRun *job.PipelineRunSummary) *pipeline.LatestRun {
	if jobLatestRun == nil {
		return nil
	}
	return &pipeline.LatestRun{
		ID:               jobLatestRun.ID,
		Name:             jobLatestRun.Name,
		Pipeline:         jobLatestRun.Pipeline,
		Result:           jobLatestRun.Result,
		State:            jobLatestRun.State,
		StartTime:        jobLatestRun.StartTime,
		EndTime:          jobLatestRun.EndTime,
		DurationInMillis: jobLatestRun.DurationInMillis,
		Causes:           convertCauses(jobLatestRun.Causes),
	}
}

func convertCauses(jobCauses []job.Cause) []pipeline.Cause {
	if jobCauses == nil {
		return nil
	}
	causes := []pipeline.Cause{}
	for _, jobCause := range jobCauses {
		causes = append(causes, pipeline.Cause{
			ShortDescription: jobCause.GetShortDescription(),
		})
	}
	return causes
}

func convertBranches(jobBranches []job.PipelineBranch) []pipeline.Branch {
	branches := make([]pipeline.Branch, 0, len(jobBranches))
	for _, jobBranch := range jobBranches {
		branches = append(branches, pipeline.Branch{
			Name:         jobBranch.Name,
			WeatherScore: jobBranch.WeatherScore,
			Branch:       jobBranch.Branch,
			PullRequest:  jobBranch.PullRequest,
			Parameters:   jobBranch.Parameters,
			Disabled:     jobBranch.Disabled,
			LatestRun:    convertLatestRun(jobBranch.LatestRun),
		})
	}
	return branches
}
