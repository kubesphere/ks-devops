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

package steptemplate

import (
	"bytes"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ksruntime "kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/common"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestAPIs(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type args struct {
		api     string
		method  string
		getBody func() io.Reader
	}
	tests := []struct {
		name         string
		args         args
		getInstances func() []runtime.Object
		wantCode     int
		verify       func([]byte, *testing.T)
	}{{
		name: "an empty list of the clusterStepTemplate",
		args: args{
			api:    "/clustersteptemplates",
			method: http.MethodGet,
		},
		wantCode: http.StatusOK,
	}, {
		name: "the whole list of the clusterStepTemplate",
		args: args{
			api:    "/clustersteptemplates",
			method: http.MethodGet,
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{&v1alpha3.ClusterStepTemplate{}}
		},
		wantCode: http.StatusOK,
	}, {
		name: "get a clusterStepTemplate by name",
		args: args{
			api:    "/clustersteptemplates/fake",
			method: http.MethodGet,
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{&v1alpha3.ClusterStepTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
			}}
		},
		wantCode: http.StatusOK,
	}, {
		name: "render a clusterStepTemplate by name",
		args: args{
			api:    "/clustersteptemplates/fake/render?secret=secret&secretNamespace=ns",
			method: http.MethodPost,
			getBody: func() io.Reader {
				return bytes.NewBufferString(`{}`)
			},
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{&v1alpha3.ClusterStepTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1alpha3.StepTemplateSpec{
					Parameters: []v1alpha3.ParameterInStep{{
						Name:         "number",
						DefaultValue: "2",
					}},
					Template: `echo {{.param.number}}`,
				},
			}, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "secret",
				},
				Type: v1.SecretTypeBasicAuth,
				Data: map[string][]byte{
					v1.BasicAuthUsernameKey: []byte("username"),
					v1.BasicAuthPasswordKey: []byte("password"),
				},
			}}
		},
		verify: func(bytes []byte, t *testing.T) {
			assert.Equal(t, `{
 "data": "{\n  \"arguments\": [\n    {\n      \"key\": \"script\",\n      \"value\": {\n        \"isLiteral\": true,\n        \"value\": \"echo 2\"\n      }\n    }\n  ],\n  \"name\": \"sh\"\n}"
}`, string(bytes))
		},
		wantCode: http.StatusOK,
	}, {
		name: "render a clusterStepTemplate by namewith parameter body ",
		args: args{
			api:    "/clustersteptemplates/fake/render?secret=secret&secretNamespace=ns",
			method: http.MethodPost,
			getBody: func() io.Reader {
				return bytes.NewBufferString(`{"number":"3"}`)
			},
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{&v1alpha3.ClusterStepTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1alpha3.StepTemplateSpec{
					Parameters: []v1alpha3.ParameterInStep{{
						Name:         "number",
						DefaultValue: "2",
					}},
					Template: `echo {{.param.number}}`,
				},
			}, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "secret",
				},
				Type: v1.SecretTypeBasicAuth,
				Data: map[string][]byte{
					v1.BasicAuthUsernameKey: []byte("username"),
					v1.BasicAuthPasswordKey: []byte("password"),
				},
			}}
		},
		verify: func(bytes []byte, t *testing.T) {
			assert.Equal(t, `{
 "data": "{\n  \"arguments\": [\n    {\n      \"key\": \"script\",\n      \"value\": {\n        \"isLiteral\": true,\n        \"value\": \"echo 3\"\n      }\n    }\n  ],\n  \"name\": \"sh\"\n}"
}`, string(bytes))
		},
		wantCode: http.StatusOK,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.getInstances == nil {
				tt.getInstances = func() []runtime.Object {
					return []runtime.Object{}
				}
			}

			ws := ksruntime.NewWebService(runtimeSchema.GroupVersion{Group: api.GroupName, Version: "v1alpha3"})
			RegisterRoutes(ws, &common.Options{
				GenericClient: fake.NewFakeClientWithScheme(schema, tt.getInstances()...),
			})
			container := restful.NewContainer()
			container.Add(ws)

			var requestBody io.Reader
			if tt.args.getBody != nil {
				requestBody = tt.args.getBody()
			}

			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.api, requestBody)
			httpRequest.Header.Set("Content-Type", "application/json")
			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, tt.wantCode, httpWriter.Code)

			if tt.verify != nil {
				tt.verify(httpWriter.Body.Bytes(), t)
			}
		})
	}
}
