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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	fakedevops "kubesphere.io/devops/pkg/client/devops/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAPIsExist(t *testing.T) {
	httpWriter := httptest.NewRecorder()
	wsWithGroup := runtime.NewWebService(v1alpha1.GroupVersion)

	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	RegisterRoutes(wsWithGroup, fakedevops.NewFakeDevops(nil), fake.NewFakeClientWithScheme(schema))
	restful.DefaultContainer.Add(wsWithGroup)

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "pipelinerun list",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/pipelines/fake/pipelineruns",
		},
	}, {
		name: "create a pipelinerun",
		args: args{
			method: http.MethodPost,
			uri:    "/namespaces/fake/pipelines/fake/pipelineruns",
		},
	}, {
		name: "get a pipelinerun",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/pipelineruns/fake",
		},
	}, {
		name: "get node details",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/pipelineruns/fake/nodedetails",
		},
	}, {
		name: "receive pipeline event",
		args: args{
			method: http.MethodPost,
			uri:    "/webhook/pipeline-event",
		},
	}, {
		name: "download artifact",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/pipelineruns/fake/artifacts/download",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha1"+tt.args.uri, nil)
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.NotEqual(t, httpWriter.Code, 404)
		})
	}
}
