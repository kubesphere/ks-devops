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
	"encoding/json"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/kapis/common"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

func Test_getPathParameter(t *testing.T) {
	type args struct {
		req   func() *restful.Request
		param *restful.Parameter
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "normal case",
		args: args{
			req: func() *restful.Request {
				request := restful.NewRequest(&http.Request{})
				request.PathParameters()["name"] = "good"
				return request
			},
			param: restful.PathParameter("name", "desc"),
		},
		want: "good",
	}, {
		name: "wrong param name",
		args: args{
			req: func() *restful.Request {
				request := restful.NewRequest(&http.Request{})
				request.PathParameters()["name"] = "good"
				return request
			},
			param: restful.PathParameter("fake", "desc"),
		},
		want: "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, common.GetPathParameter(tt.args.req(), tt.args.param), "getPathParameter(%v, %v)", tt.args.req, tt.args.param)
		})
	}
}

func Test_handler_applicationList(t *testing.T) {
	createRequest := func(uri, namespace string) *restful.Request {
		fakeRequest := httptest.NewRequest(http.MethodGet, uri, nil)
		request := restful.NewRequest(fakeRequest)
		request.PathParameters()[common.NamespacePathParameter.Data().Name] = namespace
		return request
	}
	createApp := func(name string, namespace string, labels map[string]string) *v1alpha1.Application {
		return &v1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	}
	createAppWithCreationTime := func(name string, namespace string, creationTimestamp time.Time) *v1alpha1.Application {
		return &v1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				CreationTimestamp: metav1.NewTime(creationTimestamp),
			},
		}
	}
	current := time.Now()
	yesterday := current.Add(-24 * time.Hour)
	tomorrow := current.Add(24 * time.Hour)
	type args struct {
		req  *restful.Request
		apps []v1alpha1.Application
	}
	tests := []struct {
		name         string
		args         args
		wantResponse api.ListResult
	}{{
		name: "Should return empty list when namespaces mismatch",
		args: args{
			req: createRequest("/applications", "default"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", nil),
			},
		},
		wantResponse: api.ListResult{
			Items:      []interface{}{},
			TotalItems: 0,
		},
	}, {
		name: "Should return correct list when namespaces match",
		args: args{
			req: createRequest("/applications", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", nil),
			},
		},
		wantResponse: api.ListResult{
			Items: []interface{}{
				createApp("fake-app-1", "fake-namespace", nil),
			},
			TotalItems: 1,
		},
	}, {
		name: "Should return empty list when healthStatus is set but no status matches",
		args: args{
			req: createRequest("/applications?healthStatus=Unknown", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", map[string]string{
					v1alpha1.HealthStatusLabelKey: "Healthy",
				}),
			},
		},
		wantResponse: api.ListResult{
			Items:      []interface{}{},
			TotalItems: 0,
		},
	}, {
		name: "Should return partial list when healthStatus is set but no status matches",
		args: args{
			req: createRequest("/applications?healthStatus=Unknown", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", map[string]string{
					v1alpha1.HealthStatusLabelKey: "Healthy",
				}),
				*createApp("fake-app-2", "fake-namespace", map[string]string{
					v1alpha1.HealthStatusLabelKey: "Unknown",
				}),
			},
		},
		wantResponse: api.ListResult{
			Items: []interface{}{
				*createApp("fake-app-2", "fake-namespace", map[string]string{
					v1alpha1.HealthStatusLabelKey: "Unknown",
				}),
			},
			TotalItems: 1,
		},
	}, {
		name: "Should return empty list when syncStatus is set but no status matches",
		args: args{
			req: createRequest("/applications?syncStatus=Unknown", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", map[string]string{
					v1alpha1.SyncStatusLabelKey: "Synced",
				}),
			},
		},
		wantResponse: api.ListResult{
			Items:      []interface{}{},
			TotalItems: 0,
		},
	}, {
		name: "Should return partial list when syncStatus is set but no status matches",
		args: args{
			req: createRequest("/applications?syncStatus=Unknown", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", map[string]string{
					v1alpha1.SyncStatusLabelKey: "Synced",
				}),
				*createApp("fake-app-2", "fake-namespace", map[string]string{
					v1alpha1.SyncStatusLabelKey: "Unknown",
				}),
			},
		},
		wantResponse: api.ListResult{
			Items: []interface{}{
				*createApp("fake-app-2", "fake-namespace", map[string]string{
					v1alpha1.SyncStatusLabelKey: "Unknown",
				}),
			},
			TotalItems: 1,
		},
	}, {
		name: "Should return empty list when healthStatus and syncStatus are set but no status matches",
		args: args{
			req: createRequest("/applications?healthStatus=Unknown&syncStatus=Unknown", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", map[string]string{
					v1alpha1.HealthStatusLabelKey: "Healthy",
					v1alpha1.SyncStatusLabelKey:   "Synced",
				}),
			},
		},
		wantResponse: api.ListResult{
			Items:      []interface{}{},
			TotalItems: 0,
		},
	}, {
		name: "Should return applications when both healthStatus and syncStatus match at the same time",
		args: args{
			req: createRequest("/applications?healthStatus=Unknown&syncStatus=Unknown", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", map[string]string{
					v1alpha1.HealthStatusLabelKey: "Unknown",
					v1alpha1.SyncStatusLabelKey:   "Synced",
				}),
				*createApp("fake-app-2", "fake-namespace", map[string]string{
					v1alpha1.HealthStatusLabelKey: "Unknown",
					v1alpha1.SyncStatusLabelKey:   "Unknown",
				}),
			},
		},
		wantResponse: api.ListResult{Items: []interface{}{
			*createApp("fake-app-2", "fake-namespace", map[string]string{
				v1alpha1.HealthStatusLabelKey: "Unknown",
				v1alpha1.SyncStatusLabelKey:   "Unknown",
			}),
		}, TotalItems: 1},
	}, {
		name: "Should filter with name",
		args: args{
			req: createRequest("/applications?name=app", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-1", "fake-namespace", nil),
				*createApp("fake-bpp-1", "fake-namespace", nil),
			},
		},
		wantResponse: api.ListResult{
			Items: []interface{}{
				*createApp("fake-app-1", "fake-namespace", nil),
			},
			TotalItems: 1,
		},
	}, {
		name: "Should sort by name in ascending order",
		args: args{
			req: createRequest("/applications?orderBy=name&ascending=true", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createApp("fake-app-2", "fake-namespace", nil),
				*createApp("fake-app-1", "fake-namespace", nil),
				*createApp("fake-app-3", "fake-namespace", nil),
			},
		},
		wantResponse: api.ListResult{
			Items: []interface{}{
				*createApp("fake-app-1", "fake-namespace", nil),
				*createApp("fake-app-2", "fake-namespace", nil),
				*createApp("fake-app-3", "fake-namespace", nil),
			},
			TotalItems: 3,
		},
	}, {
		name: "Should sort by creationTimestamp in ascending order",
		args: args{
			req: createRequest("/applications?orderBy=creationTimestamp&ascending=true", "fake-namespace"),
			apps: []v1alpha1.Application{
				*createAppWithCreationTime("fake-app-2", "fake-namespace", current),
				*createAppWithCreationTime("fake-app-1", "fake-namespace", yesterday),
				*createAppWithCreationTime("fake-app-3", "fake-namespace", tomorrow),
			},
		},
		wantResponse: api.ListResult{
			Items: []interface{}{
				*createAppWithCreationTime("fake-app-1", "fake-namespace", yesterday),
				*createAppWithCreationTime("fake-app-2", "fake-namespace", current),
				*createAppWithCreationTime("fake-app-3", "fake-namespace", tomorrow),
			},
			TotalItems: 3,
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
			fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, toObjects(tt.args.apps)...)
			h := &handler{
				Client: fakeClient,
			}
			req := tt.args.req
			recorder := httptest.NewRecorder()
			resp := restful.NewResponse(recorder)
			resp.SetRequestAccepts(restful.MIME_JSON)
			h.applicationList(req, resp)
			assert.Equal(t, 200, recorder.Code)
			wantResponseBytes, err := json.Marshal(tt.wantResponse)
			assert.Nil(t, err)
			assert.JSONEq(t, string(wantResponseBytes), recorder.Body.String())
		})
	}
}
