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
package pipelinerun

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/apiserver/request"
	"github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/pipelineengine"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/ks-devops/pkg/apiserver/runtime"
	fakedevops "github.com/kubesphere/ks-devops/pkg/client/devops/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApis(t *testing.T) {
	wsWithGroup := runtime.NewWebService(v1alpha3.GroupVersion)
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	RegisterRoutes(wsWithGroup, fakedevops.NewFakeDevops(nil), fake.NewClientBuilder().WithScheme(schema).WithObjects(&v1alpha3.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake",
			Namespace: "fake",
		},
		Spec: v1alpha3.PipelineSpec{
			Type: v1alpha3.NoScmPipelineType,
		},
	}).Build())
	restful.DefaultContainer.Add(wsWithGroup)

	type args struct {
		method  string
		uri     string
		getBody func() io.Reader
		ctx     context.Context
		status  int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create a pipelinerun",
			args: args{
				method: http.MethodPost,
				uri:    "/namespaces/fake/pipelines/fake/pipelineruns",
				getBody: func() io.Reader {
					payload := &devops.RunPayload{
						Parameters: []devops.Parameter{{
							Name:  "aname",
							Value: "avalue",
						}},
					}
					data, _ := json.Marshal(payload)
					return bytes.NewBuffer(data)
				},
				ctx:    request.NewContext(),
				status: 401,
			},
		},
		{
			name: "create a pipelinerun with a mock user",
			args: args{
				method: http.MethodPost,
				uri:    "/namespaces/fake/pipelines/fake/pipelineruns",
				getBody: func() io.Reader {
					payload := &devops.RunPayload{
						Parameters: []devops.Parameter{{
							Name:  "aname",
							Value: "avalue",
						}},
					}
					data, _ := json.Marshal(payload)
					return bytes.NewBuffer(data)
				},
				ctx: request.WithUser(
					request.NewContext(),
					&user.DefaultInfo{
						Name:   "bob",
						UID:    "123",
						Groups: []string{"group1"},
						Extra:  map[string][]string{"foo": {"bar"}},
					},
				),
				status: 200,
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody io.Reader
			if tt.args.getBody != nil {
				requestBody = tt.args.getBody()
			}
			httpRequest, _ := http.NewRequestWithContext(tt.args.ctx, tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, requestBody)
			httpRequest.Header.Set("Content-Type", "application/json")
			httpWriter := httptest.NewRecorder()
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, tt.args.status, httpWriter.Code)
		})
	}
}

// TestCreatePipelineRunPropagatesEngineAnnotations verifies engine snapshots and compatibility annotations.
func TestCreatePipelineRunPropagatesEngineAnnotations(t *testing.T) {
	tests := []struct {
		name                string
		pipelineEngine      *v1alpha3.PipelineEngineSpec
		pipelineAnnotations map[string]string
		wantAnnotations     map[string]string
		wantEngine          v1alpha3.PipelineEngineType
	}{
		{
			name:           "Tekton Pipeline spec",
			pipelineEngine: &v1alpha3.PipelineEngineSpec{Type: v1alpha3.PipelineEngineTekton},
			pipelineAnnotations: map[string]string{
				pipelineengine.AnnotationEngine:         pipelineengine.EngineJenkins,
				pipelineengine.AnnotationTektonPipeline: "native-pipeline",
				"example.com/unrelated":                 "must-not-propagate",
			},
			wantAnnotations: map[string]string{
				pipelineengine.AnnotationTektonPipeline: "native-pipeline",
			},
			wantEngine: v1alpha3.PipelineEngineTekton,
		},
		{
			name: "Legacy Tekton annotation",
			pipelineAnnotations: map[string]string{
				pipelineengine.AnnotationEngine:         pipelineengine.EngineTekton,
				pipelineengine.AnnotationTektonPipeline: "native-pipeline",
				"example.com/unrelated":                 "must-not-propagate",
			},
			wantAnnotations: map[string]string{
				pipelineengine.AnnotationEngine:         pipelineengine.EngineTekton,
				pipelineengine.AnnotationTektonPipeline: "native-pipeline",
			},
		},
		{
			name:                "Default Jenkins pipeline",
			pipelineAnnotations: nil,
			wantAnnotations:     map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, err := v1alpha3.SchemeBuilder.Register().Build()
			require.NoError(t, err)

			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "pipeline",
					Namespace:   "namespace",
					Annotations: tt.pipelineAnnotations,
				},
				Spec: v1alpha3.PipelineSpec{
					Engine: tt.pipelineEngine,
					Type:   v1alpha3.NoScmPipelineType,
				},
			}
			k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()

			container := restful.NewContainer()
			webService := runtime.NewWebService(v1alpha3.GroupVersion)
			RegisterRoutes(webService, fakedevops.NewFakeDevops(nil), k8sClient)
			container.Add(webService)

			payload, err := json.Marshal(&devops.RunPayload{
				Parameters: []devops.Parameter{{Name: "message", Value: "hello"}},
			})
			require.NoError(t, err)

			ctx := request.WithUser(request.NewContext(), &user.DefaultInfo{Name: "bob"})
			httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3/namespaces/namespace/pipelines/pipeline/pipelineruns",
				bytes.NewReader(payload))
			require.NoError(t, err)
			httpRequest.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			container.ServeHTTP(recorder, httpRequest)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

			runs := &v1alpha3.PipelineRunList{}
			require.NoError(t, k8sClient.List(context.Background(), runs))
			require.Len(t, runs.Items, 1)

			created := runs.Items[0]
			for key, value := range tt.wantAnnotations {
				assert.Equal(t, value, created.Annotations[key])
			}
			if len(tt.wantAnnotations) == 0 {
				assert.NotContains(t, created.Annotations, pipelineengine.AnnotationEngine)
			}
			if tt.wantEngine != "" {
				assert.NotContains(t, created.Annotations, pipelineengine.AnnotationEngine)
				require.NotNil(t, created.Spec.PipelineSpec)
				require.NotNil(t, created.Spec.PipelineSpec.Engine)
				assert.Equal(t, tt.wantEngine, created.Spec.PipelineSpec.Engine.Type)
			}
			assert.NotContains(t, created.Annotations, "example.com/unrelated")
			assert.Equal(t, "bob", created.Annotations[v1alpha3.PipelineRunCreatorAnnoKey])
			require.Len(t, created.Spec.Parameters, 1)
			assert.Equal(t, "message", created.Spec.Parameters[0].Name)
			assert.Equal(t, "hello", created.Spec.Parameters[0].Value)
		})
	}
}

func TestGetNodeDetails(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	pipelineRun := &v1alpha3.PipelineRun{}
	pipelineRun.SetName("pr1")
	pipelineRun.SetNamespace("ns")

	cm := &v1.ConfigMap{Data: map[string]string{}}
	cm.SetName(pipelineRun.GetName())
	cm.SetNamespace(pipelineRun.GetNamespace())
	cm.Data["stage"] = `[{"id":"id","steps":[{"approvable":true}]}]`

	handler := &apiHandler{
		apiHandlerOption: apiHandlerOption{
			client: fake.NewClientBuilder().WithScheme(schema).
				WithObjects(pipelineRun.DeepCopy()).
				WithObjects(cm.DeepCopy()).Build(),
		},
	}

	recorder := httptest.NewRecorder()
	req := restful.NewRequest(&http.Request{
		Header: map[string][]string{
			"Accept": {"*/*"},
		},
	})
	restful.DefaultResponseContentType(restful.MIME_JSON)
	req.PathParameters()["namespace"] = "ns"
	req.PathParameters()["pipelinerun"] = "pr1"
	resp := restful.NewResponse(recorder)
	handler.getNodeDetails(req, resp)

	body, err := io.ReadAll(recorder.Body)
	assert.Nil(t, err)
	assert.Equal(t, `[
 {
  "id": "id",
  "startTime": null,
  "steps": [
   {
    "startTime": null,
    "approvable": true
   }
  ]
 }
]`, string(body))
}
