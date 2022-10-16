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

package fluxcd

import (
	"bytes"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/kapis/common"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	type args struct {
		service *restful.WebService
		options *common.Options
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, service *restful.WebService)
	}{{
		name: "normal case",
		args: args{
			service: runtime.NewWebService(v1alpha1.GroupVersion),
			options: &common.Options{GenericClient: fake.NewFakeClientWithScheme(schema)},
		},
		verify: func(t *testing.T, service *restful.WebService) {
			assert.Greater(t, len(service.Routes()), 0)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterRoutes(tt.args.service, tt.args.options, &config.FluxCDOption{})
			tt.verify(t, tt.args.service)
		})
	}
}

func TestPublicAPIs(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	fluxApp := v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "app",
		},
		Spec: v1alpha1.ApplicationSpec{
			Kind: v1alpha1.FluxCD,
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
	}{
		{
			name: "get a normal list of the flux applications",
			request: request{
				method: http.MethodGet,
				uri:    "/namespaces/ns/applications",
			},
			k8sclient:    fake.NewFakeClientWithScheme(schema, fluxApp.DeepCopy()),
			responseCode: http.StatusOK,
			verify: func(t *testing.T, body []byte) {
				list := &api.ListResult{}
				err := yaml.Unmarshal(body, list)
				assert.Nil(t, err)

				assert.Equal(t, 1, len(list.Items))
				assert.Nil(t, err)
			},
		},
		{
			name: "get a normal flux application",
			request: request{
				method: http.MethodGet,
				uri:    "/namespaces/ns/applications/app",
			},
			k8sclient:    fake.NewFakeClientWithScheme(schema, fluxApp.DeepCopy()),
			responseCode: http.StatusOK,
			verify: func(t *testing.T, body []byte) {
				app := &unstructured.Unstructured{}
				err := yaml.Unmarshal(body, app)
				assert.Nil(t, err)

				name, _, err := unstructured.NestedString(app.Object, "metadata", "name")
				assert.Equal(t, "app", name)
				assert.Nil(t, err)
			},
		},
		{
			name: "delete a flux application",
			request: request{
				method: http.MethodDelete,
				uri:    "/namespaces/ns/applications/app",
			},
			k8sclient:    fake.NewFakeClientWithScheme(schema, fluxApp.DeepCopy()),
			responseCode: http.StatusOK,
			verify: func(t *testing.T, body []byte) {
				app := &unstructured.Unstructured{}
				err := yaml.Unmarshal(body, app)
				assert.Nil(t, err)

				name, _, err := unstructured.NestedString(app.Object, "metadata", "name")
				assert.Equal(t, "app", name)
				assert.Nil(t, err)
			},
		},
		{
			name: "create a flux application",
			request: request{
				method: http.MethodPost,
				uri:    "/namespaces/ns/applications",
				body: func() io.Reader {
					return bytes.NewBuffer([]byte(`{
  "apiVersion": "gitops.kubesphere.io/v1alpha1",
  "kind": "Application",
  "metadata": {
    "name": "app",
    "labels": {
        "gitops.kubesphere.io/save-helm-template": "true"
    }
  },
  "spec": {
    "kind": "fluxcd",
    "fluxApp": {
      "spec": {
        "source": {
          "sourceRef": {
            "kind": "GitRepository",
            "name": "fluxcd-github-repo",
            "namespace": "fake-ns"
          }
        },
        "config": {
          "helmRelease": {
            "chart": {
              "chart": "./helm-chart",
              "version": "0.1.0",
              "interval": "5m0s",
              "reconcileStrategy": "Revision",
              "valuesFiles": [
                "./helm-chart/values.yaml",
                "./helm-chart/aliyun-values.yaml"
              ]
            },
            "deploy": [
              {
                "destination": {
                  "targetNamespace": "helm-app"
                },
                "interval": "1m0s",
                "install": {
                  "createNamespace": true
                },
                "upgrade": {
                  "remediation": {
                    "remediateLastFailure": true
                  },
                  "force": true
                }
              },
              {
                "destination": {
                  "targetNamespace": "another-helm-app"
                },
                "interval": "1m0s",
                "install": {
                  "createNamespace": true
                },
                "upgrade": {
                  "remediation": {
                    "remediateLastFailure": true
                  },
                  "force": true
                }
              }
            ]
          }
        }
      }
    }
  }
}`))
				},
			},
			k8sclient:    fake.NewFakeClientWithScheme(schema),
			responseCode: http.StatusOK,
			verify: func(t *testing.T, body []byte) {
				app := &unstructured.Unstructured{}
				err := yaml.Unmarshal(body, app)
				assert.Nil(t, err)

				name, _, err := unstructured.NestedString(app.Object, "metadata", "name")
				assert.Equal(t, "app", name)
				assert.Nil(t, err)
			},
		},
		{
			name: "update a flux application",
			request: request{
				method: http.MethodPut,
				uri:    "/namespaces/ns/applications/app",
				body: func() io.Reader {
					return bytes.NewBuffer([]byte(`{
  "apiVersion": "gitops.kubesphere.io/v1alpha1",
  "kind": "Application",
  "metadata": {
    "name": "app",
	"namespace": "ns",
    "labels": {
        "gitops.kubesphere.io/save-helm-template": "true"
    }
  },
  "spec": {
    "kind": "fluxcd",
    "fluxApp": {
      "spec": {
        "source": {
          "sourceRef": {
            "kind": "GitRepository",
            "name": "fluxcd-github-repo",
            "namespace": "fake-ns"
          }
        },
        "config": {
          "helmRelease": {
            "chart": {
              "chart": "./helm-chart",
              "version": "0.1.0",
              "interval": "5m0s",
              "reconcileStrategy": "Revision",
              "valuesFiles": [
                "./helm-chart/values.yaml",
                "./helm-chart/aliyun-values.yaml"
              ]
            },
            "deploy": [
              {
                "destination": {
                  "targetNamespace": "helm-app"
                },
                "interval": "1m0s",
                "install": {
                  "createNamespace": true
                },
                "upgrade": {
                  "remediation": {
                    "remediateLastFailure": true
                  },
                  "force": true
                },
                "values": {
                    "replicaCount": 3
                }
              },
              {
                "destination": {
                  "targetNamespace": "another-helm-app"
                },
                "interval": "1m0s",
                "install": {
                  "createNamespace": true
                },
                "upgrade": {
                  "remediation": {
                    "remediateLastFailure": true
                  },
                  "force": true
                }
              }
            ]
          }
        }
      }
    }
  }
}`))
				},
			},
			k8sclient:    fake.NewFakeClientWithScheme(schema, fluxApp.DeepCopy()),
			responseCode: http.StatusOK,
			verify: func(t *testing.T, body []byte) {
				app := &unstructured.Unstructured{}
				err := yaml.Unmarshal(body, app)
				assert.Nil(t, err)

				deploy, _, err := unstructured.NestedSlice(app.Object, "spec", "fluxApp", "spec", "config", "helmRelease", "deploy")
				replicas, _, err := unstructured.NestedInt64(deploy[0].(map[string]interface{}), "values", "replicaCount")
				assert.Equal(t, int64(3), replicas)
				assert.Nil(t, err)
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsWithGroup := runtime.NewWebService(v1alpha1.GroupVersion)
			RegisterRoutes(wsWithGroup, &common.Options{GenericClient: tt.k8sclient}, &config.FluxCDOption{
				Enabled: true,
			})

			container := restful.NewContainer()
			container.Add(wsWithGroup)

			api := fmt.Sprintf("http://fake.com/kapis/gitops.kubesphere.io/%s%s", v1alpha1.GroupVersion.Version, tt.request.uri)
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
