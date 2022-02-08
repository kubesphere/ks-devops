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
