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

package argocd

import (
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/kapis/common"
	"net/http"
	"testing"
)

func Test_getPathParameter(t *testing.T) {
	type args struct {
		req   func() *restful.Request
		param *restful.Parameter
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "normal case",
		args: args{
			req: func() *restful.Request {
				request := restful.NewRequest(&http.Request{})
				request.PathParameters()["name"] = "good"
				return request
			},
			param: restful.PathParameter("name", "desc"),
		},
		want: "good",
	}, {
		name: "wrong param name",
		args: args{
			req: func() *restful.Request {
				request := restful.NewRequest(&http.Request{})
				request.PathParameters()["name"] = "good"
				return request
			},
			param: restful.PathParameter("fake", "desc"),
		},
		want: "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, common.GetPathParameter(tt.args.req(), tt.args.param), "getPathParameter(%v, %v)", tt.args.req, tt.args.param)
		})
	}
}
