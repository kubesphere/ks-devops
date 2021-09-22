package pipelinerun

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/pipelinerun/v1alpha3"
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

// pipelineBuildApplier applies PipelineBuilder to PipelineRunStatus.
type pipelineBuildApplier struct {
	*job.PipelineRun
}

func (pbApplier pipelineBuildApplier) apply(prStatus *v1alpha3.PipelineRunStatus) {
	condition := v1alpha3.Condition{
		Type:               v1alpha3.ConditionReady,
		Status:             v1alpha3.ConditionUnknown,
		LastProbeTime:      v1.Now(),
		LastTransitionTime: v1.Now(),
		Reason:             pbApplier.State,
	}

	prStatus.Phase = v1alpha3.Unknown

	switch pbApplier.State {
	case Queued.String():
		condition.Status = v1alpha3.ConditionUnknown
		prStatus.Phase = v1alpha3.Pending
	case Running.String():
		condition.Status = v1alpha3.ConditionUnknown
		prStatus.Phase = v1alpha3.Running
	case Paused.String():
		condition.Status = v1alpha3.ConditionUnknown
		prStatus.Phase = v1alpha3.Pending
	case Skipped.String():
		condition.Type = v1alpha3.ConditionSucceeded
		condition.Status = v1alpha3.ConditionTrue
		prStatus.Phase = v1alpha3.Succeeded
	case NotBuiltState.String():
		condition.Status = v1alpha3.ConditionUnknown
	case Finished.String():
		pbApplier.whenPipelineRunFinished(&condition, prStatus)
	}
	prStatus.AddCondition(&condition)
	prStatus.UpdateTime = &v1.Time{Time: time.Now()}
}

func (pbApplier pipelineBuildApplier) whenPipelineRunFinished(condition *v1alpha3.Condition, prStatus *v1alpha3.PipelineRunStatus) {
	// mark as completed
	if !pbApplier.EndTime.IsZero() {
		prStatus.CompletionTime = &v1.Time{Time: pbApplier.EndTime.Time}
	} else {
		// should never happen
		prStatus.CompletionTime = &v1.Time{Time: time.Now()}
	}
	condition.Type = v1alpha3.ConditionSucceeded
	// handle result
	switch pbApplier.Result {
	case Success.String():
		condition.Status = v1alpha3.ConditionTrue
		prStatus.Phase = v1alpha3.Succeeded
	case Unstable.String():
		condition.Status = v1alpha3.ConditionFalse
		prStatus.Phase = v1alpha3.Failed
	case Failure.String():
		condition.Status = v1alpha3.ConditionFalse
		prStatus.Phase = v1alpha3.Failed
	case NotBuiltResult.String():
		condition.Status = v1alpha3.ConditionUnknown
		prStatus.Phase = v1alpha3.Unknown
	case Unknown.String():
		condition.Status = v1alpha3.ConditionUnknown
		prStatus.Phase = v1alpha3.Unknown
	case Aborted.String():
		condition.Status = v1alpha3.ConditionFalse
		prStatus.Phase = v1alpha3.Failed
	}
}

// parameterConverter is responsible to convert Parameter slice of PipelineRun into job.Parameter slice.
type parameterConverter struct {
	parameters []v1alpha3.Parameter
}

func (converter parameterConverter) convert() []job.Parameter {
	var params []job.Parameter
	for _, param := range converter.parameters {
		params = append(params, job.Parameter{
			Name:  param.Name,
			Value: param.Value,
		})
	}
	return params
}
