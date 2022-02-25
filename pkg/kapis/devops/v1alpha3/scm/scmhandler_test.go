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
	"context"
	"encoding/json"
	"github.com/emicklei/go-restful"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/client/git"
	"kubesphere.io/devops/pkg/constants"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler := newHandler(nil)
	assert.NotNil(t, handler)
}

func TestSCMAPI(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name    string
		args    args
		prepare func()
		verify  func(code int, response []byte, t *testing.T)
	}{{
		name: "verify",
		args: args{
			method: http.MethodPost,
			uri:    "/scms/github/verify?secret=token&secretNamespace=default",
		},
		prepare: func() {
			var mockHeaders = map[string]string{
				"X-GitHub-Request-Id":   "DD0E:6011:12F21A8:1926790:5A2064E2",
				"X-RateLimit-Limit":     "60",
				"X-RateLimit-Remaining": "59",
				"X-RateLimit-Reset":     "1512076018",
			}

			var mockPageHeaders = map[string]string{
				"Link": `<https://api.github.com/resource?page=1>; rel="next",` +
					`<https://api.github.com/resource?page=1>; rel="prev",` +
					`<https://api.github.com/resource?page=1>; rel="first",` +
					`<https://api.github.com/resource?page=1>; rel="last"`,
			}

			gock.New("https://api.github.com").
				Get("/user/orgs").
				MatchParam("per_page", "1").
				MatchParam("page", "1").
				Reply(200).
				Type("application/json").
				SetHeaders(mockHeaders).
				SetHeaders(mockPageHeaders).
				File("testdata/orgs.json")
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)

			resp := &git.VerifyResponse{}
			err := json.Unmarshal(response, resp)
			assert.Nil(t, err)
			assert.Equal(t, "ok", resp.Message)
			assert.Equal(t, "token", resp.CredentialID)
		},
	}, {
		name: "verify against a fake scm type",
		args: args{
			method: http.MethodPost,
			uri:    "/scms/fake/verify?secret=token&secretNamespace=default",
		},
		prepare: func() {},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, http.StatusOK, code)

			resp := &git.VerifyResponse{}
			err := json.Unmarshal(response, resp)
			assert.Nil(t, err, string(response))
			assert.Equal(t, 100, resp.Code)
		},
	}, {
		name: "get the organization list",
		args: args{
			method: http.MethodGet,
			uri:    "/scms/github/organizations?secret=token&secretNamespace=default",
		},
		prepare: func() {
			var mockHeaders = map[string]string{
				"X-GitHub-Request-Id":   "DD0E:6011:12F21A8:1926790:5A2064E2",
				"X-RateLimit-Limit":     "60",
				"X-RateLimit-Remaining": "59",
				"X-RateLimit-Reset":     "1512076018",
			}

			var mockPageHeaders = map[string]string{
				"Link": `<https://api.github.com/resource?page=1>; rel="next",` +
					`<https://api.github.com/resource?page=1>; rel="prev",` +
					`<https://api.github.com/resource?page=1>; rel="first",` +
					`<https://api.github.com/resource?page=1>; rel="last"`,
			}

			gock.New("https://api.github.com").
				Get("/user/orgs").
				MatchParam("per_page", "1").
				MatchParam("page", "1").
				Reply(200).
				Type("application/json").
				SetHeaders(mockHeaders).
				SetHeaders(mockPageHeaders).
				File("testdata/orgs.json")
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)

			var orgs []organization
			err := json.Unmarshal(response, &orgs)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(orgs))
			assert.Equal(t, "https://github.com/images/error/octocat_happy.gif", orgs[0].Avatar)
			assert.Equal(t, "github", orgs[0].Name)
		},
	}, {
		name: "get the repository list",
		args: args{
			method: http.MethodGet,
			uri:    "/scms/github/organizations/octocat/repositories?secret=token&secretNamespace=default",
		},
		prepare: func() {
			var mockHeaders = map[string]string{
				"X-GitHub-Request-Id":   "DD0E:6011:12F21A8:1926790:5A2064E2",
				"X-RateLimit-Limit":     "60",
				"X-RateLimit-Remaining": "59",
				"X-RateLimit-Reset":     "1512076018",
			}

			var mockPageHeaders = map[string]string{
				"Link": `<https://api.github.com/resource?page=1>; rel="next",` +
					`<https://api.github.com/resource?page=1>; rel="prev",` +
					`<https://api.github.com/resource?page=1>; rel="first",` +
					`<https://api.github.com/resource?page=1>; rel="last"`,
			}

			gock.New("https://api.github.com").
				Get("user/repos").
				MatchParam("per_page", "1").
				MatchParam("page", "1").
				Reply(200).
				Type("application/json").
				SetHeaders(mockHeaders).
				SetHeaders(mockPageHeaders).
				File("testdata/repos.json")
		},
		verify: func(code int, response []byte, t *testing.T) {
			assert.Equal(t, 200, code)

			var repos []repository
			err := json.Unmarshal(response, &repos)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(repos))
			assert.Equal(t, "Hello-World", repos[0].Name)
			assert.Equal(t, "master", repos[0].DefaultBranch)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			tt.prepare()

			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, nil)
			httpRequest = httpRequest.WithContext(context.WithValue(context.TODO(), constants.K8SToken, constants.ContextKeyK8SToken("")))

			ws := runtime.NewWebService(runtimeSchema.GroupVersion{Group: api.GroupName, Version: "v1alpha3"})
			RegisterRoutersForSCM(fake.NewFakeClientWithScheme(schema, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "token", Namespace: "default",
				},
				Type: v1.SecretTypeOpaque,
				Data: map[string][]byte{
					v1.ServiceAccountTokenKey: []byte("token"),
				},
			}), ws)
			container := restful.NewContainer()
			container.Add(ws)

			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, httpRequest)
			tt.verify(httpWriter.Code, httpWriter.Body.Bytes(), t)
		})
	}
}
