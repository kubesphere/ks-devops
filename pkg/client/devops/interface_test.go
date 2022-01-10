package devops

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestGetDevOpsStatusCode(t *testing.T) {
	type args struct {
		devopsErr error
	}
	tests := []struct {
		name string
		args args
		want int
	}{{
		name: "just be a number",
		args: args{devopsErr: fmt.Errorf("%d", http.StatusAccepted)},
		want: http.StatusAccepted,
	}, {
		name: "ErrorResponse",
		args: args{
			devopsErr: &ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
					Request: &http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Scheme: "http",
							Host:   "host",
						},
					},
				},
			},
		},
		want: http.StatusNotFound,
	}, {
		name: "other error type",
		args: args{devopsErr: fmt.Errorf("other error type")},
		want: http.StatusInternalServerError,
	}, {
		name: "a formatted error message",
		args: args{devopsErr: fmt.Errorf("unexpected status code: 404")},
		want: http.StatusNotFound,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDevOpsStatusCode(tt.args.devopsErr); got != tt.want {
				t.Errorf("GetDevOpsStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
