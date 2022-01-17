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

package webhook

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	apiserverruntime "kubesphere.io/devops/pkg/apiserver/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAPIsExist(t *testing.T) {

	type args struct {
		method     string
		uri        string
		bodyJSON   string
		initObject []runtime.Object
	}
	tests := []struct {
		name      string
		args      args
		assertion func(t *testing.T, c client.Client)
	}{{
		name: "receive pipeline event",
		args: args{
			method: http.MethodPost,
			uri:    "/webhooks/jenkins",
			initObject: []runtime.Object{
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name:      "example-pipeline",
						Namespace: "my-devops-project",
					},
				},
			},
			bodyJSON: `
				{
					"data":     {
						"multiBranch": false,
						"parentFullName": "my-devops-project",
						"run":         {
							"_class": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
							"actions":             [
												{
									"_class": "hudson.model.ParametersAction",
									"parameters": [                    {
										"_class": "hudson.model.BooleanParameterValue",
										"name": "skip",
										"value": false
									}]
								},
								{"_class": "org.jenkinsci.plugins.displayurlapi.actions.RunDisplayAction"},
								{"_class": "org.jenkinsci.plugins.pipeline.modeldefinition.actions.RestartDeclarativePipelineAction"},
								{},
								{"_class": "org.jenkinsci.plugins.workflow.job.views.FlowGraphAction"},
								{},
								{},
								{}
							],
							"artifacts": [],
							"building": true,
							"description": null,
							"displayName": "#1",
							"duration": 0,
							"estimatedDuration": -1,
							"executor": {"_class": "hudson.model.OneOffExecutor"},
							"fullDisplayName": "my-devops-project » example-pipeline #1",
							"id": "1",
							"keepLog": false,
							"number": 1,
							"queueId": 1,
							"result": null,
							"timestamp": 1642399916330,
							"changeSets": [],
							"culprits": [],
							"nextBuild": null,
							"previousBuild": null
						},
						"projectName": "example-pipeline"
					},
					"dataType": "io.jenkins.plugins.pipeline.event.data.WorkflowRunData",
					"id": "948bce89-5844-454d-aa6d-75acb886381a",
					"source": "job/my-devops-project/job/example-pipeline/",
					"time": "2022-01-17T14:11:56.359+0800",
					"type": "run.initialize"
				}`,
		},
		assertion: func(t *testing.T, c client.Client) {
			pipelineruns := &v1alpha3.PipelineList{}
			err := c.List(context.Background(), pipelineruns)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(pipelineruns.Items))
		},
	}}
	for _, tt := range tests {
		httpWriter := httptest.NewRecorder()
		wsWithGroup := apiserverruntime.NewWebService(v1alpha3.GroupVersion)

		scheme := runtime.NewScheme()
		err := v1alpha3.AddToScheme(scheme)
		assert.Nil(t, err)

		fakeClient := fake.NewFakeClientWithScheme(scheme, tt.args.initObject...)
		RegisterWebhooks(fakeClient, wsWithGroup)
		restful.DefaultContainer.Add(wsWithGroup)

		t.Run(tt.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tt.args.bodyJSON != "" {
				bodyReader = strings.NewReader(tt.args.bodyJSON)
			}
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, bodyReader)
			httpRequest.Header.Set("Content-Type", "application/json")
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, 200, httpWriter.Code)
			if tt.assertion != nil {
				tt.assertion(t, fakeClient)
			}
		})
	}
}
