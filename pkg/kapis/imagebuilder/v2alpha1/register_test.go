/*

  Copyright 2020 The KubeSphere Authors.

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

package v2alpha1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/api/devops/v2alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	fakedevops "kubesphere.io/devops/pkg/client/devops/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAPIsExist(t *testing.T) {
	httpWriter := httptest.NewRecorder()
	wsWithGroup := runtime.NewWebService(v2alpha1.GroupVersion)

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
		name: "create an imageBuild",
		args: args{
			method: http.MethodPost,
			uri:    "/namespaces/fake/ImageBuilds/fake",
		},
	}, {
		name: "get all imageBuilds",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/ImageBuilds",
		},
	}, {
		name: "get an imageBuild",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/ImageBuilds/fake",
		},
	}, {
		name: "delete an imageBuild",
		args: args{
			method: http.MethodPost,
			uri:    "/namespaces/fake/ImageBuilds/fake",
		},
	}, {
		name: "update an imageBuild",
		args: args{
			method: http.MethodPost,
			uri:    "/namespaces/fake/ImageBuilds/fake",
		},
	}, {
		name: "get all imageBuildStrategies",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/ImageBuildStrategies",
		},
	}, {
		name: "get an imageBuildStrategy",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/imageBuildStrategies/fake",
		},
	}, {
		name: "create an imageBuildRun",
		args: args{
			method: http.MethodPost,
			uri:    "/namespaces/fake/ImageBuildRuns/fake",
		},
	}, {
		name: "get all imageBuildRun",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/ImageBuildRuns",
		},
	}, {
		name: "get an imageBuildRun",
		args: args{
			method: http.MethodGet,
			uri:    "/namespace/fake/ImageBuildRuns/fake",
		},
	}, {
		name: "delete an imageBuildRun",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/ImageBuildRuns/fake",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v2alpha1"+tt.args.uri, nil)
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.NotEqual(t, httpWriter.Code, 404)
		})
	}
}
