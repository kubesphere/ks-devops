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
	"bytes"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	type args struct {
		service *restful.WebService
		options *kapisv1alpha1.Options
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, service *restful.WebService)
	}{{
		name: "normal case",
		args: args{
			service: runtime.NewWebService(v1alpha1.GroupVersion),
			options: &kapisv1alpha1.Options{GenericClient: fake.NewFakeClientWithScheme(schema)},
		},
		verify: func(t *testing.T, service *restful.WebService) {
			assert.Equal(t, 6, len(service.Routes()))
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterRoutes(tt.args.service, tt.args.options)
			tt.verify(t, tt.args.service)
		})
	}
}

func TestAPIs(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	app := v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "app",
		},
	}

	nonArgoClusterSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "non-argo-cluster",
			Namespace: "ns",
		},
	}
	invalidArgoClusterSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-argo-cluster",
			Namespace: "ns",
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
			},
		},
		Data: map[string][]byte{
			"server": []byte("server"),
		},
	}
	validArgoClusterSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argo-cluster",
			Namespace: "ns",
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
			},
		},
		Data: map[string][]byte{
			"name":   []byte("name"),
			"server": []byte("server"),
		},
	}

	type request struct {
		method string
		uri    string
		body   func() io.Reader
	}
	tests := []struct {
		name         string
		request      request
		responseCode int
		k8sclient    client.Client
		verify       func(t *testing.T, body []byte)
	}{{
		name: "get an empty list of the applications",
		request: request{
			method: http.MethodGet,
			uri:    "/devops/fake/applications",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			list := &unstructured.Unstructured{}
			err := yaml.Unmarshal(body, list)
			assert.Nil(t, err)

			items, _, err := unstructured.NestedSlice(list.Object, "items")
			assert.Equal(t, 0, len(items))
			assert.NotNil(t, err)
		},
	}, {
		name: "get a normal list of the applications",
		request: request{
			method: http.MethodGet,
			uri:    "/devops/ns/applications",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, app.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			list := &unstructured.Unstructured{}
			err := yaml.Unmarshal(body, list)
			assert.Nil(t, err)

			items, _, err := unstructured.NestedSlice(list.Object, "items")
			assert.Equal(t, 1, len(items))
			assert.Nil(t, err)
		},
	}, {
		name: "get a normal application",
		request: request{
			method: http.MethodGet,
			uri:    "/devops/ns/applications/app",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, app.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			list := &unstructured.Unstructured{}
			err := yaml.Unmarshal(body, list)
			assert.Nil(t, err)

			name, _, err := unstructured.NestedString(list.Object, "metadata", "name")
			assert.Equal(t, "app", name)
			assert.Nil(t, err)
		},
	}, {
		name: "delete an application",
		request: request{
			method: http.MethodDelete,
			uri:    "/devops/ns/applications/app",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, app.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			list := &unstructured.Unstructured{}
			err := yaml.Unmarshal(body, list)
			assert.Nil(t, err)

			name, _, err := unstructured.NestedString(list.Object, "metadata", "name")
			assert.Equal(t, "app", name)
			assert.Nil(t, err)
		},
	}, {
		name: "create an application",
		request: request{
			method: http.MethodPost,
			uri:    "/devops/ns/applications",
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{
  "apiVersion": "devops.kubesphere.io/v1alpha1",
  "kind": "Application",
  "metadata": {
    "name": "fake"
  },
  "spec": {
    "argoApp": {
      "project": "default"
    }
  }
}`))
			},
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			list := &unstructured.Unstructured{}
			err := yaml.Unmarshal(body, list)
			assert.Nil(t, err)

			name, _, err := unstructured.NestedString(list.Object, "metadata", "name")
			assert.Equal(t, "fake", name)
			assert.Nil(t, err)
		},
	}, {
		name: "create an application, invalid payload",
		request: request{
			method: http.MethodPost,
			uri:    "/devops/ns/applications",
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`fake`))
			},
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema),
		responseCode: http.StatusInternalServerError,
	}, {
		name: "update an application",
		request: request{
			method: http.MethodPut,
			uri:    "/devops/ns/applications/app",
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{
  "apiVersion": "devops.kubesphere.io/v1alpha1",
  "kind": "Application",
  "metadata": {
    "name": "app",
    "namespace": "ns"
  },
  "spec": {
    "argoApp": {
      "project": "good"
    }
  }
}`))
			},
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, app.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			list := &unstructured.Unstructured{}
			err := yaml.Unmarshal(body, list)
			assert.Nil(t, err)

			project, _, err := unstructured.NestedString(list.Object, "spec", "argoApp", "project")
			assert.Equal(t, "good", project)
			assert.Nil(t, err)
		},
	}, {
		name: "get clusters, no expected data",
		request: request{
			method: http.MethodGet,
			uri:    "/clusters",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, nonArgoClusterSecret.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			assert.Equal(t, "[]", string(body))
		},
	}, {
		name: "get clusters, have invalid data",
		request: request{
			method: http.MethodGet,
			uri:    "/clusters",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, invalidArgoClusterSecret.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			assert.Equal(t, "[]", string(body))
		},
	}, {
		name: "get clusters, have the expected data",
		request: request{
			method: http.MethodGet,
			uri:    "/clusters",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, validArgoClusterSecret.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			assert.Equal(t, `[
 {
  "server": "server",
  "name": "name"
 }
]`, string(body))
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsWithGroup := runtime.NewWebService(v1alpha1.GroupVersion)
			RegisterRoutes(wsWithGroup, &kapisv1alpha1.Options{GenericClient: tt.k8sclient})

			container := restful.NewContainer()
			container.Add(wsWithGroup)

			api := fmt.Sprintf("http://fake.com/kapis/devops.kubesphere.io/%s%s", v1alpha1.GroupVersion.Version, tt.request.uri)
			var body io.Reader
			if tt.request.body != nil {
				body = tt.request.body()
			}
			req, err := http.NewRequest(tt.request.method, api, body)
			req.Header.Set("Content-Type", "application/json")
			assert.Nil(t, err)

			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, req)
			assert.Equal(t, tt.responseCode, httpWriter.Code)

			if tt.verify != nil {
				tt.verify(t, httpWriter.Body.Bytes())
			}
		})
	}
}
