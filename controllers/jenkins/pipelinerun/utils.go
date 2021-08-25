package pipelinerun

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
