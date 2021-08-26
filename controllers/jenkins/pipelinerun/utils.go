package pipelinerun

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"time"
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

func convertPipelineBuild(pipelineBuild *job.PipelineBuild, prStatus *devopsv1alpha4.PipelineRunStatus) {
	condition := devopsv1alpha4.Condition{
		Type:               devopsv1alpha4.ConditionReady,
		LastProbeTime:      v1.Now(),
		LastTransitionTime: v1.Now(),
		Reason:             pipelineBuild.State,
	}

	var phase = devopsv1alpha4.Unknown

	switch pipelineBuild.State {
	case Queued.String():
		condition.Status = devopsv1alpha4.ConditionUnknown
		phase = devopsv1alpha4.Pending
	case Running.String():
		condition.Status = devopsv1alpha4.ConditionUnknown
		phase = devopsv1alpha4.Running
	case Paused.String():
		condition.Status = devopsv1alpha4.ConditionUnknown
		phase = devopsv1alpha4.Pending
	case Skipped.String():
		condition.Type = devopsv1alpha4.ConditionSucceeded
		condition.Status = devopsv1alpha4.ConditionTrue
		phase = devopsv1alpha4.Succeeded
	case NotBuiltState.String():
		condition.Status = devopsv1alpha4.ConditionUnknown
		phase = devopsv1alpha4.Unknown
	case Finished.String():
		// mark as completed
		if !pipelineBuild.EndTime.IsZero() {
			prStatus.CompletionTime = &v1.Time{Time: pipelineBuild.EndTime.Time}
		} else {
			// should never happen
			prStatus.CompletionTime = &v1.Time{Time: time.Now()}
		}
		// handle result
		switch pipelineBuild.Result {
		case Success.String():
			condition.Type = devopsv1alpha4.ConditionSucceeded
			condition.Status = devopsv1alpha4.ConditionTrue
			phase = devopsv1alpha4.Succeeded
		case Unstable.String():
			condition.Status = devopsv1alpha4.ConditionFalse
			phase = devopsv1alpha4.Failed
		case Failure.String():
			condition.Status = devopsv1alpha4.ConditionFalse
			phase = devopsv1alpha4.Failed
		case NotBuiltResult.String():
			condition.Status = devopsv1alpha4.ConditionUnknown
			phase = devopsv1alpha4.Unknown
		case Unknown.String():
			condition.Status = devopsv1alpha4.ConditionUnknown
			phase = devopsv1alpha4.Unknown
		case Aborted.String():
			condition.Status = devopsv1alpha4.ConditionFalse
			phase = devopsv1alpha4.Failed
		}
	default:
		condition.Status = devopsv1alpha4.ConditionUnknown
	}

	prStatus.Phase = phase
	prStatus.AddCondition(&condition)
	prStatus.UpdateTime = &v1.Time{Time: time.Now()}
}
