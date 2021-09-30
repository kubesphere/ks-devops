package pipeline

import "github.com/jenkins-zh/jenkins-client/pkg/job"

// metadata holds some of pipeline fields that are only things we needed instead of whole job.Pipeline.
type metadata struct {
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

func convert(jobPipeline *job.Pipeline) *metadata {
	return &metadata{
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
