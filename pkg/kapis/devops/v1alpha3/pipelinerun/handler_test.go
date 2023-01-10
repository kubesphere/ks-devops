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
package pipelinerun

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/request"
	"kubesphere.io/devops/pkg/client/devops"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	fakedevops "kubesphere.io/devops/pkg/client/devops/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApis(t *testing.T) {
	wsWithGroup := runtime.NewWebService(v1alpha3.GroupVersion)
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	RegisterRoutes(wsWithGroup, fakedevops.NewFakeDevops(nil), fake.NewFakeClientWithScheme(schema, &v1alpha3.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake",
			Namespace: "fake",
		},
		Spec: v1alpha3.PipelineSpec{
			Type: v1alpha3.NoScmPipelineType,
		},
	}))
	restful.DefaultContainer.Add(wsWithGroup)

	type args struct {
		method  string
		uri     string
		getBody func() io.Reader
		ctx     context.Context
		status  int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create a pipelinerun",
			args: args{
				method: http.MethodPost,
				uri:    "/namespaces/fake/pipelines/fake/pipelineruns",
				getBody: func() io.Reader {
					payload := &devops.RunPayload{
						Parameters: []devops.Parameter{{
							Name:  "aname",
							Value: "avalue",
						}},
					}
					data, _ := json.Marshal(payload)
					return bytes.NewBuffer(data)
				},
				ctx:    request.NewContext(),
				status: 401,
			},
		},
		{
			name: "create a pipelinerun with a mock user",
			args: args{
				method: http.MethodPost,
				uri:    "/namespaces/fake/pipelines/fake/pipelineruns",
				getBody: func() io.Reader {
					payload := &devops.RunPayload{
						Parameters: []devops.Parameter{{
							Name:  "aname",
							Value: "avalue",
						}},
					}
					data, _ := json.Marshal(payload)
					return bytes.NewBuffer(data)
				},
				ctx: request.WithUser(
					request.NewContext(),
					&user.DefaultInfo{
						Name:   "bob",
						UID:    "123",
						Groups: []string{"group1"},
						Extra:  map[string][]string{"foo": {"bar"}},
					},
				),
				status: 200,
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody io.Reader
			if tt.args.getBody != nil {
				requestBody = tt.args.getBody()
			}
			httpRequest, _ := http.NewRequestWithContext(tt.args.ctx, tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, requestBody)
			httpRequest.Header.Set("Content-Type", "application/json")
			httpWriter := httptest.NewRecorder()
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, tt.args.status, httpWriter.Code)
		})
	}
}
