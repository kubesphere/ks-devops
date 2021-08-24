package pipelinerun

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseJenkinsTime(t *testing.T) {
	type args struct {
		jenkinsTime string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{{
		name: "valid UTC time",
		args: args{
			jenkinsTime: "2021-08-18T16:36:47.236+0000",
		},
		want: time.Date(2021, 8, 18, 16, 36, 47, 236000000, time.UTC),
	}, {
		name: "invalid time",
		args: args{
			jenkinsTime: "2021-08-18 16:36:47.236+0000",
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseJenkinsTime(tt.args.jenkinsTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJenkinsTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseJenkinsTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}
