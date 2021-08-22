package pipelinerun

import (
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"net/http"
	"reflect"
	"strings"
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

func Test_buildHttpParametersForRunning(t *testing.T) {
	type args struct {
		prSpec *devopsv1alpha4.PipelineRunSpec
	}
	tests := []struct {
		name    string
		args    args
		want    *devopsClient.HttpParameters
		wantErr bool
	}{{
		name: "nil PipelineRun",
		args: args{
			prSpec: nil,
		},
		wantErr: true,
	}, {
		name: "with nil parameters",
		args: args{
			prSpec: &devopsv1alpha4.PipelineRunSpec{},
		},
		want: &devopsClient.HttpParameters{
			Url:    mockClientURL(),
			Method: http.MethodPost,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: NopCloser(strings.NewReader(`{"parameters":[]}`)),
		},
	}, {
		name: "with one parameters",
		args: args{
			prSpec: &devopsv1alpha4.PipelineRunSpec{
				Parameters: []devopsv1alpha4.Parameter{
					{
						Name:  "devops",
						Value: "wow",
					},
				},
			},
		},
		want: &devopsClient.HttpParameters{
			Url:    mockClientURL(),
			Method: http.MethodPost,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: NopCloser(strings.NewReader(`{"parameters":[{"name":"devops","value":"wow"}]}`)),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildHTTPParametersForRunning(tt.args.prSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildHTTPParametersForRunning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildHTTPParametersForRunning() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStaticURL(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{{
		name: "get static URL",
		want: "https://devops.kubesphere.io/",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mockClientURL(); !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("mockClientURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
