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

package v1alpha1

import (
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/kapis/common"
	"kubesphere.io/devops/pkg/kapis/gitops/v1alpha1/gitops"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
)

func TestArgoAPIsExist(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	container := restful.NewContainer()
	opt := &common.Options{
		GenericClient: gitops.Handler{
			Client: fake.NewFakeClientWithScheme(schema),
		},
	}
	argoOption := &config.ArgoCDOption{Enabled: true, Namespace: "argocd"}
	fluxOption := &config.FluxCDOption{Enabled: false}
	AddToContainer(container, opt, argoOption, fluxOption)
	type args struct {
		method string
		uri    string
	}

	tests := []struct {
		name string
		args
		body       string
		expectCode int
	}{
		{
			name: "not found an application",
			args: args{
				method: http.MethodGet,
				uri:    "/namespaces/fake-ns/applications",
			},
			expectCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method, "http://fake.com/kapis/gitops.kubesphere.io/v1alpha3"+tt.args.uri, strings.NewReader(tt.body))
			httpRequest.Header.Set("Content-Type", "application/json")

			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, tt.expectCode, httpWriter.Code)
		})
	}
}

func TestFluxAPIsExist(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	container := restful.NewContainer()
	opt := &common.Options{
		GenericClient: gitops.Handler{
			Client: fake.NewFakeClientWithScheme(schema),
		},
	}
	argoOption := &config.ArgoCDOption{Enabled: false, Namespace: "argocd"}
	fluxOption := &config.FluxCDOption{Enabled: true}
	AddToContainer(container, opt, argoOption, fluxOption)
	type args struct {
		method string
		uri    string
	}

	tests := []struct {
		name string
		args
		body       string
		expectCode int
	}{
		{
			name: "not found an application",
			args: args{
				method: http.MethodGet,
				uri:    "/namespaces/fake-ns/applications",
			},
			expectCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method, "http://fake.com/kapis/gitops.kubesphere.io/v1alpha3"+tt.args.uri, strings.NewReader(tt.body))
			httpRequest.Header.Set("Content-Type", "application/json")

			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, tt.expectCode, httpWriter.Code)
		})
	}
}

func Test_NonAPIsExist(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	container := restful.NewContainer()
	opt := &common.Options{
		GenericClient: gitops.Handler{
			Client: fake.NewFakeClientWithScheme(schema),
		},
	}
	argoOption := &config.ArgoCDOption{Enabled: false, Namespace: "argocd"}
	fluxOption := &config.FluxCDOption{Enabled: false}
	wss := AddToContainer(container, opt, argoOption, fluxOption)
	assert.Nil(t, wss)
}
