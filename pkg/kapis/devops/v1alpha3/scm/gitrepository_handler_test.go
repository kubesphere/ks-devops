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

package scm

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/emicklei/go-restful"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ksruntime "kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/constants"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestGitRepositoryAPI(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	repo := &v1alpha3.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "repo-1",
		},
	}

	type args struct {
		method  string
		uri     string
		getBody func() io.Reader
	}
	tests := []struct {
		name         string
		args         args
		getInstances func() []runtime.Object
		verify       func(code int, response []byte, t *testing.T)
	}{{
		name: "get repository list",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/ns/gitrepositories",
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{repo.DeepCopy()}
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)

			repoList := v1alpha3.GitRepositoryList{}
			err := json.Unmarshal(response, &repoList)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(repoList.Items))
		},
	}, {
		name: "get a particular repository object",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/ns/gitrepositories/repo-1",
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{repo.DeepCopy()}
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)

			repo := v1alpha3.GitRepository{}
			err := json.Unmarshal(response, &repo)
			assert.Nil(t, err)
			assert.Equal(t, "repo-1", repo.Name)
		},
	}, {
		name: "delete a particular repository object",
		args: args{
			method: http.MethodDelete,
			uri:    "/namespaces/ns/gitrepositories/repo-1",
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{repo.DeepCopy()}
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)

			repo := v1alpha3.GitRepository{}
			err := json.Unmarshal(response, &repo)
			assert.Nil(t, err)
			assert.Equal(t, "repo-1", repo.Name)
		},
	}, {
		name: "create a repository",
		args: args{
			method: http.MethodPost,
			uri:    "/namespaces/ns/gitrepositories",
			getBody: func() io.Reader {
				data, _ := json.Marshal(repo.DeepCopy())
				return bytes.NewBuffer(data)
			},
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{}
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)
		},
	}, {
		name: "update a repository",
		args: args{
			method: http.MethodPut,
			uri:    "/namespaces/ns/gitrepositories/repo-1",
			getBody: func() io.Reader {
				newRepo := repo.DeepCopy()
				newRepo.Spec.URL = "http://newrepo.com"
				data, _ := json.Marshal(newRepo)
				return bytes.NewBuffer(data)
			},
		},
		getInstances: func() []runtime.Object {
			return []runtime.Object{repo.DeepCopy()}
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)

			repo := &v1alpha3.GitRepository{}
			err := json.Unmarshal(response, repo)
			assert.Nil(t, err)
			assert.Equal(t, "http://newrepo.com", repo.Spec.URL)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			var requestBody io.Reader
			if tt.args.getBody != nil {
				requestBody = tt.args.getBody()
			}
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, requestBody)
			httpRequest = httpRequest.WithContext(context.WithValue(context.TODO(), constants.K8SToken, constants.ContextKeyK8SToken("")))
			httpRequest.Header.Set("Content-Type", "application/json")

			ws := ksruntime.NewWebService(runtimeSchema.GroupVersion{Group: api.GroupName, Version: "v1alpha3"})
			RegisterRoutersForSCM(fake.NewFakeClientWithScheme(schema, tt.getInstances()...), ws)
			container := restful.NewContainer()
			container.Add(ws)

			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, httpRequest)
			tt.verify(httpWriter.Code, httpWriter.Body.Bytes(), t)
		})
	}
}
