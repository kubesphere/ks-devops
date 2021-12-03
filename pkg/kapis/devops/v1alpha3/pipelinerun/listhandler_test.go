package pipelinerun

import (
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

func Test_listHandler_Comparator(t *testing.T) {
	now := v1.Now()
	tomorrow := v1.Time{Time: now.Add(1 * time.Hour)}
	createPipelineRun := func(name string, creationTime v1.Time, startTime *v1.Time) *v1alpha3.PipelineRun {
		return &v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Name:              name,
				CreationTimestamp: creationTime,
			},
			Status: v1alpha3.PipelineRunStatus{
				StartTime: startTime,
			},
		}
	}
	type args struct {
		left  *v1alpha3.PipelineRun
		right *v1alpha3.PipelineRun
	}
	tests := []struct {
		name string
		args args
		// expect whether we need to exchange left and right while sorting.
		// false: left position should swap with right.
		// true: left and right should keep their position.
		shouldNotSwap bool
	}{{
		name: "Compare with start time and first and left is earlier than right",
		args: args{
			left:  createPipelineRun("b", now, &now),
			right: createPipelineRun("a", now, &tomorrow),
		},
		shouldNotSwap: false,
	}, {
		name: "Compare with start time and first and left is later than right",
		args: args{
			left:  createPipelineRun("b", now, &tomorrow),
			right: createPipelineRun("a", now, &now),
		},
		shouldNotSwap: true,
	}, {
		name: "Compare with start time and start times are equal",
		args: args{
			left:  createPipelineRun("b", now, &now),
			right: createPipelineRun("a", now, &now),
		},
		shouldNotSwap: false,
	}, {
		name: "Return to compare with creation time while start time is nil",
		args: args{
			left:  createPipelineRun("b", now, nil),
			right: createPipelineRun("a", now, &tomorrow),
		},
		shouldNotSwap: false,
	}, {
		name: "Return to compare with creation time while one of start time is nil",
		args: args{
			left:  createPipelineRun("b", now, nil),
			right: createPipelineRun("a", now, &tomorrow),
		},
		shouldNotSwap: false,
	}, {
		name: "Return to compare with name while one of start time is nil and star time is equal to creation time",
		args: args{
			left:  createPipelineRun("a", now, nil),
			right: createPipelineRun("b", now, &now),
		},
		shouldNotSwap: true,
	}, {
		name: "Return to compare with name while start times are equal",
		args: args{
			left:  createPipelineRun("a", now, &now),
			right: createPipelineRun("b", now, &now),
		},
		shouldNotSwap: true,
	}, {
		name: "Return to compare with name while start times are nil and creation times are equal",
		args: args{
			left:  createPipelineRun("b", now, nil),
			right: createPipelineRun("a", now, nil),
		},
		shouldNotSwap: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := listHandler{}
			if got := h.Comparator()(tt.args.left, tt.args.right, ""); !reflect.DeepEqual(got, tt.shouldNotSwap) {
				t.Errorf("pipelineRunListHandler.Comparator() = %v, want %v", got, tt.shouldNotSwap)
			}
		})
	}
}
