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

package template

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/common"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "List templates",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake-devops/templates",
		},
	}, {
		name: "Get a template",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake-devops/templates/fake-template",
		},
	}, {
		name: "Render a template",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake-devops/templates/fake-template/render",
		},
	}, {
		name: "List cluster templates",
		args: args{
			method: http.MethodGet,
			uri:    "/clustertemplates",
		},
	}, {
		name: "Render a cluster template",
		args: args{
			method: http.MethodPost,
			uri:    "/clustertemplates/fake-template/render",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := restful.NewContainer()
			utilruntime.Must(v1alpha3.SchemeBuilder.AddToScheme(scheme.Scheme))
			fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
			service := runtime.NewWebService(v1alpha3.GroupVersion)
			RegisterRoutes(service, &common.Options{
				GenericClient: fakeClient,
			})
			container.Add(service)

			uri := fmt.Sprintf("/kapis/%s/%s%s",
				v1alpha3.GroupVersion.Group, v1alpha3.GroupVersion.Version, tt.args.uri)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.args.method, uri, nil)
			request.Header.Add(restful.HEADER_ContentType, restful.MIME_JSON)

			container.ServeHTTP(recorder, request)
			if recorder.Code == 404 {
				assert.NotContains(t, recorder.Body.String(), "Page Not Found")
			} else {
				assert.Equal(t, 200, recorder.Code)
			}
		})
	}
}
