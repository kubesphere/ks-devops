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
package v1alpha1

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestAPIsExist(t *testing.T) {
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
	}}
	for _, tt := range tests {
		container := restful.NewContainer()
		utilruntime.Must(v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme))
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
		AddToContainer(container, &common.Options{
			GenericClient: fakeClient,
		})
		t.Run(tt.name, func(t *testing.T) {
			uriWithGroupVersion := fmt.Sprintf("/kapis/%s/%s%s",
				v1alpha1.GroupVersion.Group, v1alpha1.GroupVersion.Version, tt.args.uri)
			uriWithVersion := fmt.Sprintf("/%s%s",
				v1alpha1.GroupVersion.Version, tt.args.uri)
			for _, uri := range []string{uriWithVersion, uriWithGroupVersion} {
				recorder := httptest.NewRecorder()
				request, _ := http.NewRequest(tt.args.method, uri, nil)
				container.ServeHTTP(recorder, request)
				if recorder.Code == 404 {
					assert.NotEqual(t, "404 page not found\n", recorder.Body.String())
				}
			}
		})
	}
}
