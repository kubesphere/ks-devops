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

package doc

import (
	"bytes"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIsExist(t *testing.T) {
	httpWriter := httptest.NewRecorder()

	AddSwaggerService(nil, restful.DefaultContainer)

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "check /apidocs.json",
		args: args{
			method: http.MethodGet,
			uri:    "/apidocs.json",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com"+tt.args.uri, bytes.NewBuffer([]byte("{}")))
			httpRequest.Header.Set("Content-Type", "application/json")
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, 200, httpWriter.Code)
		})
	}
}
