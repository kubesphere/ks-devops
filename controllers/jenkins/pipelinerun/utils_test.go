package pipelinerun

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"testing"
	"time"
)

func Test_pipelineBuildApplier_apply(t *testing.T) {
	type fields struct {
		pb *job.PipelineBuild
	}
	type args struct {
		prStatus *devopsv1alpha4.PipelineRunStatus
	}

	commonStatusAssert := func(prStatus *devopsv1alpha4.PipelineRunStatus) {
		assert.Equal(t, 1, len(prStatus.Conditions))
		assert.NotNil(t, prStatus.Conditions[0].LastProbeTime)
		assert.NotNil(t, prStatus.Conditions[0].LastTransitionTime)
		assert.NotNil(t, prStatus.UpdateTime)
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion func(prStatus *devopsv1alpha4.PipelineRunStatus)
	}{{
		name: "PipelineRun was in queue",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:    "1",
				State: Queued.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionReady, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Pending, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was running",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:    "1",
				State: Running.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionReady, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Running, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was paused",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:    "1",
				State: Paused.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionReady, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Pending, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was skipped",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:    "1",
				State: Skipped.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionTrue, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Succeeded, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was not built",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:    "1",
				State: NotBuiltState.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionReady, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Unknown, prStatus.Phase)
		},
	}, {
		name: "Unknown PipelineRun state",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:    "1",
				State: "this_is_an_invalid_state",
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionReady, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Unknown, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was finished with succeeded result",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:      "1",
				State:   Finished.String(),
				Result:  Success.String(),
				EndTime: job.Time{Time: time.Date(2021, 8, 27, 11, 16, 38, 0, time.Local)},
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionTrue, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Succeeded, prStatus.Phase)
			assert.Equal(t, time.Date(2021, 8, 27, 11, 16, 38, 0, time.Local), prStatus.CompletionTime.Time)
		},
	}, {
		name: "PipelineRun was finished but with unstable result",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:     "1",
				State:  Finished.String(),
				Result: Unstable.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionFalse, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Failed, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was finished but failed",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:     "1",
				State:  Finished.String(),
				Result: Failure.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionFalse, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Failed, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was finished but with not built result",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:     "1",
				State:  Finished.String(),
				Result: NotBuiltResult.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Unknown, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was finished but with unknown result",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:     "1",
				State:  Finished.String(),
				Result: Unknown.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Unknown, prStatus.Phase)
		},
	}, {
		name: "PipelineRun was finished but with aborted result",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:     "1",
				State:  Finished.String(),
				Result: Aborted.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			commonStatusAssert(prStatus)
			assert.Equal(t, devopsv1alpha4.ConditionFalse, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.Failed, prStatus.Phase)
		},
	}, {
		name: "PipelineRun with new condition",
		fields: fields{
			pb: &job.PipelineBuild{
				ID:     "1",
				State:  Finished.String(),
				Result: Success.String(),
			},
		},
		args: args{
			prStatus: &devopsv1alpha4.PipelineRunStatus{
				Conditions: []devopsv1alpha4.Condition{
					{
						Type:          devopsv1alpha4.ConditionReady,
						Status:        devopsv1alpha4.ConditionUnknown,
						LastProbeTime: metav1.Now(),
					},
				},
			},
		},
		assertion: func(prStatus *devopsv1alpha4.PipelineRunStatus) {
			assert.Equal(t, 2, len(prStatus.Conditions))
			assert.Equal(t, devopsv1alpha4.ConditionSucceeded, prStatus.Conditions[0].Type)
			assert.Equal(t, devopsv1alpha4.ConditionTrue, prStatus.Conditions[0].Status)
			assert.Equal(t, devopsv1alpha4.ConditionReady, prStatus.Conditions[1].Type)
			assert.Equal(t, devopsv1alpha4.ConditionUnknown, prStatus.Conditions[1].Status)
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pbApplier := &pipelineBuildApplier{
				PipelineBuild: tt.fields.pb,
			}
			pbApplier.apply(tt.args.prStatus)
			tt.assertion(tt.args.prStatus)
		})
	}
}
