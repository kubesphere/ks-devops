package pipelinerun

import (
	"io"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func Test_getStubUrl(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{{
		name: "get stub url",
		want: "https://devops.kubesphere.io/",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStubUrl(); !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("getStubUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildHttpParametersForRunning(t *testing.T) {
	type args struct {
		pr *devopsv1alpha4.PipelineRun
	}
	tests := []struct {
		name    string
		args    args
		want    *devopsClient.HttpParameters
		wantErr bool
	}{{
		name: "nil PipelineRun",
		args: args{
			pr: nil,
		},
		wantErr: true,
	}, {
		name: "with nil parameters",
		args: args{
			pr: &devopsv1alpha4.PipelineRun{},
		},
		want: &devopsClient.HttpParameters{
			Url:    getStubUrl(),
			Method: http.MethodPost,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"parameters":[]}`)),
		},
	}, {
		name: "with nil parameters",
		args: args{
			pr: &devopsv1alpha4.PipelineRun{},
		},
		want: &devopsClient.HttpParameters{
			Url:    getStubUrl(),
			Method: http.MethodPost,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"parameters":[]}`)),
		},
	}, {
		name: "with one parameters",
		args: args{
			pr: &devopsv1alpha4.PipelineRun{
				Spec: devopsv1alpha4.PipelineRunSpec{
					Parameters: []devopsv1alpha4.Parameter{
						{
							Name:  "devops",
							Value: "wow",
						},
					},
				},
			},
		},
		want: &devopsClient.HttpParameters{
			Url:    getStubUrl(),
			Method: http.MethodPost,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"parameters":[{"name":"devops","value":"wow"}]}`)),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildHttpParametersForRunning(tt.args.pr)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildHttpParametersForRunning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildHttpParametersForRunning() got = %v, want %v", got, tt.want)
			}
		})
	}
}
