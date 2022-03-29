// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package common

import (
	"github.com/emicklei/go-restful"
	"net/http"
	"testing"
)

func TestGetPageParameters(t *testing.T) {
	type args struct {
		req *restful.Request
	}
	tests := []struct {
		name           string
		args           args
		wantPageNumber int
		wantPageSize   int
	}{{
		name: "normal case, page number is 1, size is 20",
		args: args{
			req: getRestfulRequestFromURL("/api?pageNumber=1&pageSize=20"),
		},
		wantPageNumber: 1,
		wantPageSize:   20,
	}, {
		name: "with invalid page number and size",
		args: args{
			req: getRestfulRequestFromURL("/api?pageNumber=a&pageSize=b"),
		},
		wantPageNumber: 1,
		wantPageSize:   10,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPageNumber, gotPageSize := GetPageParameters(tt.args.req)
			if gotPageNumber != tt.wantPageNumber {
				t.Errorf("GetPageParameters() gotPageNumber = %v, want %v", gotPageNumber, tt.wantPageNumber)
			}
			if gotPageSize != tt.wantPageSize {
				t.Errorf("GetPageParameters() gotPageSize = %v, want %v", gotPageSize, tt.wantPageSize)
			}
		})
	}
}

func getRestfulRequestFromURL(requestURL string) *restful.Request {
	defaultRequest, _ := http.NewRequest(http.MethodGet, requestURL, nil)
	return &restful.Request{
		Request: defaultRequest,
	}
}
