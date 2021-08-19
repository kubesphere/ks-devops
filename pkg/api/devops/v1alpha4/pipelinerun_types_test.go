package v1alpha4

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestConditionSlice_Sort(t *testing.T) {

	now := v1.Now()
	futureTime := v1.NewTime(now.Add(1 * time.Second))
	pastTime := v1.NewTime(now.Add(-1 * time.Second))

	type args struct {
		conditions []Condition
	}
	tests := []struct {
		name string
		args args
		want []Condition
	}{{
		name: "same probe times",
		args: args{
			conditions: []Condition{
				{
					Message:       "a",
					LastProbeTime: now,
				},
				{
					Message:       "b",
					LastProbeTime: now,
				},
			},
		},
		want: []Condition{
			{
				Message:       "a",
				LastProbeTime: now,
			},
			{
				Message:       "b",
				LastProbeTime: now,
			},
		},
	}, {
		name: "probe times with ASC order",
		args: args{
			conditions: []Condition{
				{
					Message:       "a",
					LastProbeTime: pastTime,
				},
				{
					Message:       "b",
					LastProbeTime: futureTime,
				},
			},
		},
		want: []Condition{
			{
				Message:       "b",
				LastProbeTime: futureTime,
			},
			{
				Message:       "a",
				LastProbeTime: pastTime,
			},
		},
	}, {
		name: "probe times with DESC order",
		args: args{
			conditions: []Condition{
				{
					Message:       "b",
					LastProbeTime: futureTime,
				},
				{
					Message:       "a",
					LastProbeTime: pastTime,
				},
			},
		},
		want: []Condition{
			{
				Message:       "b",
				LastProbeTime: futureTime,
			},
			{
				Message:       "a",
				LastProbeTime: pastTime,
			},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(conditionSlice(tt.args.conditions))
			got := tt.args.conditions
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Condition slice sort, got = %v, want = %v", got, tt.want)
			}
		})
	}
}
