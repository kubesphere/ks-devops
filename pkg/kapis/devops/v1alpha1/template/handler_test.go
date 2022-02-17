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
	"encoding/json"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_toObjects(t *testing.T) {
	createTemplate := func(name string) *v1alpha1.Template {
		return &v1alpha1.Template{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "fake-namespace",
			},
		}
	}
	type args struct {
		templates []v1alpha1.Template
	}
	tests := []struct {
		name string
		args args
		want []runtime.Object
	}{{
		name: "Should convert correctly",
		args: args{
			templates: []v1alpha1.Template{
				*createTemplate("template1"),
				*createTemplate("template2"),
			},
		},
		want: []runtime.Object{
			createTemplate("template1"),
			createTemplate("template2"),
		},
	}, {
		name: "Should return nil if templates argument is nil",
		args: args{
			templates: nil,
		},
		want: nil,
	}, {
		name: "Should return nil if templates argument is an empty slice",
		args: args{
			templates: []v1alpha1.Template{},
		},
		want: nil,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toObjects(tt.args.templates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toObjects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newHandler(t *testing.T) {
	fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
	type args struct {
		options *kapisv1alpha1.Options
	}
	tests := []struct {
		name string
		args args
		want *handler
	}{{
		name: "Should set handler correctly",
		args: args{
			options: &kapisv1alpha1.Options{
				GenericClient: fakeClient,
			},
		},
		want: &handler{
			genericClient: fakeClient,
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newHandler(tt.args.options); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handler_handleQuery(t *testing.T) {
	createTemplate := func(name string) *v1alpha1.Template {
		return &v1alpha1.Template{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "fake-devops",
			},
		}
	}
	createRequest := func(uri, devopsName string) *restful.Request {
		fakeRequest := httptest.NewRequest(http.MethodGet, uri, nil)
		request := restful.NewRequest(fakeRequest)
		request.PathParameters()[kapisv1alpha1.DevopsPathParameter.Data().Name] = devopsName
		return request
	}
	type args struct {
		initObjects []runtime.Object
		request     *restful.Request
	}
	tests := []struct {
		name         string
		args         args
		wantResponse interface{}
	}{{
		name: "Should return empty list if no templates found",
		args: args{
			request: createRequest("/v1alpha1/devops/fake-devops/templates", "fake-devops"),
		},
		wantResponse: api.ListResult{
			Items: []interface{}{},
		},
	}, {
		name: "Should return non-empty list if templates found",
		args: args{
			request: createRequest("/v1alpha1/devops/fake-devops/templates?sortBy=name&ascending=true", "fake-devops"),
			initObjects: []runtime.Object{
				createTemplate("template1"),
				createTemplate("template2"),
				createTemplate("template3"),
			},
		},
		wantResponse: api.ListResult{
			Items: []interface{}{
				*createTemplate("template1"),
				*createTemplate("template2"),
				*createTemplate("template3"),
			},
			TotalItems: 3,
		},
	}, {
		name: "Should return empty list if out of page",
		args: args{
			request: createRequest("/v1alpha1/devops/fake-devops/templates?sortBy=name&ascending=true&page=10", "fake-devops"),
			initObjects: []runtime.Object{
				createTemplate("template1"),
				createTemplate("template2"),
				createTemplate("template3"),
			},
		},
		wantResponse: api.ListResult{
			Items:      []interface{}{},
			TotalItems: 3,
		},
	},
	}
	for _, tt := range tests {
		utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
		utilruntime.Must(v1alpha3.AddToScheme(scheme.Scheme))
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, tt.args.initObjects...)

		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				genericClient: fakeClient,
			}
			request := tt.args.request
			recorder := httptest.NewRecorder()
			response := restful.NewResponse(recorder)
			response.SetRequestAccepts(restful.MIME_JSON)
			h.handleQuery(request, response)

			assert.Equal(t, 200, recorder.Code)
			wantResponseBytes, err := json.Marshal(tt.wantResponse)
			assert.Nil(t, err)
			assert.JSONEq(t, string(wantResponseBytes), recorder.Body.String())
		})
	}
}

func Test_handler_handleGet(t *testing.T) {
	fakeTemplate := &v1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-template",
			Namespace: "fake-devops",
		},
	}
	createRequest := func(uri, devopsName, templateName string) *restful.Request {
		fakeRequest := httptest.NewRequest(http.MethodGet, uri, nil)
		request := restful.NewRequest(fakeRequest)
		request.PathParameters()[kapisv1alpha1.DevopsPathParameter.Data().Name] = devopsName
		request.PathParameters()[TemplatePathParameter.Data().Name] = templateName
		return request
	}
	type args struct {
		initObjects []runtime.Object
		request     *restful.Request
	}
	tests := []struct {
		name      string
		args      args
		wantCode  int
		assertion func(*testing.T, *httptest.ResponseRecorder)
	}{{
		name: "Should return not found if template does not exist",
		args: args{
			request: createRequest("/v1alpha1/devops/fake-devops/templates/fake-template", "fake-devops", "fake-template"),
		},
		wantCode: 404,
		assertion: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			assert.Equal(t, "templates.devops.kubesphere.io \"fake-template\" not found\n", recorder.Body.String())
		},
	}, {
		name: "Should return template if template exists ",
		args: args{
			request:     createRequest("/v1alpha1/devops/fake-devops/templates/fake-template", "fake-devops", "fake-template"),
			initObjects: []runtime.Object{fakeTemplate},
		},
		wantCode: 200,
		assertion: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			gotTemplate := &v1alpha1.Template{}
			_ = json.Unmarshal(recorder.Body.Bytes(), gotTemplate)
			assert.Equal(t, fakeTemplate.ObjectMeta, gotTemplate.ObjectMeta)
			assert.Equal(t, fakeTemplate.Spec, gotTemplate.Spec)
			assert.Equal(t, fakeTemplate.Status, gotTemplate.Status)
		},
	}}
	for _, tt := range tests {
		utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
		utilruntime.Must(v1alpha3.AddToScheme(scheme.Scheme))
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, tt.args.initObjects...)

		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				genericClient: fakeClient,
			}
			recorder := httptest.NewRecorder()
			response := restful.NewResponse(recorder)
			response.SetRequestAccepts(restful.MIME_JSON)
			h.handleGet(tt.args.request, response)

			assert.Equal(t, tt.wantCode, recorder.Code)
			if tt.assertion != nil {
				tt.assertion(t, recorder)
			}
		})
	}
}

func Test_handler_handleRender(t *testing.T) {
	fakeTemplate := &v1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-template",
			Namespace: "fake-devops",
		},
		Spec: v1alpha1.TemplateSpec{
			Template: "fake template content",
		},
	}
	createRequest := func(uri, devopsName, templateName string) *restful.Request {
		fakeRequest := httptest.NewRequest(http.MethodGet, uri, nil)
		request := restful.NewRequest(fakeRequest)
		request.PathParameters()[kapisv1alpha1.DevopsPathParameter.Data().Name] = devopsName
		request.PathParameters()[TemplatePathParameter.Data().Name] = templateName
		return request
	}
	type args struct {
		initObjects []runtime.Object
		request     *restful.Request
	}
	tests := []struct {
		name      string
		args      args
		wantCode  int
		assertion func(*testing.T, *httptest.ResponseRecorder)
	}{{
		name: "Should return not found if template not found",
		args: args{
			request: createRequest("/v1alpha1/devops/fake-devops/templates/fake-template/render", "fake-devops", "fake-template"),
		},
		wantCode: 404,
		assertion: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			assert.Equal(t, "templates.devops.kubesphere.io \"fake-template\" not found\n", recorder.Body.String())
		},
	}, {
		name: "Should set render result into annotations properly if no parameters needed",
		args: args{
			request: createRequest("/v1alpha1/devops/fake-devops/templates/fake-template/render", "fake-devops", "fake-template"),
			initObjects: []runtime.Object{
				fakeTemplate,
			},
		},
		wantCode: 200,
		assertion: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			gotTemplate := &v1alpha1.Template{}
			_ = json.Unmarshal(recorder.Body.Bytes(), gotTemplate)
			renderResult := gotTemplate.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey]
			assert.Equal(t, fakeTemplate.Spec.Template, renderResult)
		},
	}}
	for _, tt := range tests {
		utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
		utilruntime.Must(v1alpha3.AddToScheme(scheme.Scheme))
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, tt.args.initObjects...)
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				genericClient: fakeClient,
			}

			recorder := httptest.NewRecorder()
			response := restful.NewResponse(recorder)
			response.SetRequestAccepts(restful.MIME_JSON)
			h.handleRender(tt.args.request, response)

			assert.Equal(t, tt.wantCode, recorder.Code)
			if tt.assertion != nil {
				tt.assertion(t, recorder)
			}
		})
	}
}
