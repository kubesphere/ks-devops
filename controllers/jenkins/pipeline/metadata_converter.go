package pipeline

import "github.com/jenkins-zh/jenkins-client/pkg/job"

// pipelineMetadata holds some of pipeline fields that are only things we needed instead of whole job.Pipeline.
type pipelineMetadata struct {
	WeatherScore                   int                       `json:"weatherScore,omitempty"`
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

type pipelineMetadataConverter struct {
}

func (*pipelineMetadataConverter) convertPipeline(jobPipeline *job.Pipeline) *pipelineMetadata {
	return &pipelineMetadata{
		WeatherScore:                   jobPipeline.WeatherScore,
		EstimatedDurationInMillis:      jobPipeline.EstimatedDurationInMillis,
		Parameters:                     jobPipeline.Parameters,
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

type pipelineBranch struct {
	Name         string                  `json:"name,omitempty"`
	WeatherScore int                     `json:"weatherScore,omitempty"`
	LatestRun    *job.PipelineRunSummary `json:"latestRun,omitempty"`
	Branch       *job.Branch             `json:"branch,omitempty"`
	PullRequest  *job.PullRequest        `json:"pullRequest,omitempty"`
}

type pipelineBranchConverter struct {
}

func (*pipelineBranchConverter) convertBranches(jobBranches []job.PipelineBranch) []pipelineBranch {
	branches := make([]pipelineBranch, 0, len(jobBranches))
	for _, jobBranch := range jobBranches {
		branches = append(branches, pipelineBranch{
			Name:         jobBranch.Name,
			WeatherScore: jobBranch.WeatherScore,
			Branch:       jobBranch.Branch,
			PullRequest:  jobBranch.PullRequest,
			LatestRun:    jobBranch.LatestRun,
		})
	}
	return branches
}
