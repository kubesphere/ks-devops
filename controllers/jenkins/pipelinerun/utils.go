/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipelinerun

import (
	"time"

	"github.com/jenkins-zh/jenkins-client/pkg/job"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

// JenkinsRunState represents current PipelineRun state.
type JenkinsRunState string

const (
	// Queued indicates the PipelineRun is queued.
	Queued JenkinsRunState = "QUEUED"
	// Running indicates the PipelineRun is running.
	Running JenkinsRunState = "RUNNING"
	// Paused indicates the PipelineRun has been paused.
	Paused JenkinsRunState = "PAUSED"
	// Skipped indicates the PipelineRun has been skipped.
	Skipped JenkinsRunState = "SKIPPED"
	// NotBuiltState indicates the PipelineRun hasn't been built yet.
	NotBuiltState JenkinsRunState = "NOT_BUILT"
	// Finished indicates the PipelineRun has finished.
	Finished JenkinsRunState = "FINISHED"
)

// String returns string value of state of PipelineRun.
func (state JenkinsRunState) String() string {
	return string(state)
}

// JenkinsRunResult represents result of PipelineRun.
type JenkinsRunResult string

const (
	// Success indicates the PipelineRun runs successfully.
	Success JenkinsRunResult = "SUCCESS"
	// Unstable indicates the PipelineRun runs with unstable result.
	Unstable JenkinsRunResult = "UNSTABLE"
	// Failure indicates the PipelineRun runs with failure.
	Failure JenkinsRunResult = "FAILURE"
	// NotBuiltResult indicates the PipelineRun hasn't been built but finishied.
	NotBuiltResult JenkinsRunResult = "NOT_BUILT"
	// Unknown indicates the PipelineRun hasn't running yet or runs with unknown result.
	Unknown JenkinsRunResult = "UNKNOWN"
	// Aborted indicates the PipelineRun runs with aborted result.
	Aborted JenkinsRunResult = "ABORTED"
)

// String returns string value of result of PipelineRun.
func (result JenkinsRunResult) String() string {
	return string(result)
}

// pipelineBuildApplier applies PipelineBuilder to PipelineRunStatus.
type pipelineBuildApplier struct {
	*job.PipelineRun
}

func (pbApplier pipelineBuildApplier) apply(prStatus *v1alpha3.PipelineRunStatus) {
	if pbApplier.PipelineRun == nil {
		return
	}
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
	prStatus.StartTime = &v1.Time{Time: pbApplier.StartTime.Time}
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
	params := make([]job.Parameter, 0, len(converter.parameters))
	for _, param := range converter.parameters {
		params = append(params, job.Parameter{
			Name:  param.Name,
			Value: param.Value,
		})
	}
	return params
}
