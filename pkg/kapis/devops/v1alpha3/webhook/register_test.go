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
	"errors"
	"fmt"
	"io"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	apiserverruntime "kubesphere.io/devops/pkg/apiserver/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestJenkinsWebhook(t *testing.T) {

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
		name: "Should create PipelineRun when workflow run initialized",
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
					"_class": "io.jenkins.plugins.generic.event.data.WorkflowRunData",
					"actions":         [
					  {
						"_class": "hudson.model.ParametersAction",
						"parameters": [                {
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
					"timestamp": 1644126495293,
					"changeSets": [],
				"culprits": [],
				"nextBuild": null,
				"previousBuild": null,
				"_multiBranch": false,
				"_parentFullName": "my-devops-project",
				"_projectName": "example-pipeline"
			  },
			  "dataType": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
			  "id": "50c33b0e-d7f1-4a34-b57e-bb82cd453894",
			  "source": "job/my-devops-project/job/example-pipeline/",
			  "time": "2022-02-06T13:48:15.307+0800",
			  "type": "run.initialize"
			}`,
		},
		assertion: func(t *testing.T, c client.Client) {
			_ = retry.OnError(retry.DefaultRetry, func(err error) bool {
				return true
			}, func() error {
				pipelineRuns := &v1alpha3.PipelineRunList{}
				_ = c.List(context.Background(), pipelineRuns)
				if len(pipelineRuns.Items) > 0 {
					return nil
				}
				time.Sleep(time.Second * 2)
				return errors.New("not found")
			})
			pipelineRuns := &v1alpha3.PipelineRunList{}
			_ = c.List(context.Background(), pipelineRuns)
			assert.Equal(t, 1, len(pipelineRuns.Items))
		},
	}, {
		name: "Should not create any PipelineRuns when type doesn't start with run",
		args: args{
			method:     http.MethodPost,
			uri:        "/webhooks/jenkins",
			initObject: []runtime.Object{},
			bodyJSON: `
				{
				  "dataType": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
				  "id": "50c33b0e-d7f1-4a34-b57e-bb82cd453894",
				  "source": "job/my-devops-project/job/example-pipeline/",
				  "time": "2022-02-06T13:48:15.307+0800",
				  "type": "job.created"
				}`,
		},
		assertion: func(t *testing.T, c client.Client) {
			pipelineruns := &v1alpha3.PipelineRunList{}
			assert.Nil(t, c.List(context.Background(), pipelineruns))
			assert.Equal(t, 0, len(pipelineruns.Items))
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilruntime.Must(v1alpha3.AddToScheme(scheme.Scheme))
			fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, tt.args.initObject...)

			container := restful.NewContainer()
			wsWithGroup := apiserverruntime.NewWebService(v1alpha3.GroupVersion)
			RegisterWebhooks(fakeClient, wsWithGroup)
			container.Add(wsWithGroup)

			var bodyReader io.Reader
			if tt.args.bodyJSON != "" {
				bodyReader = strings.NewReader(tt.args.bodyJSON)
			}

			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, bodyReader)
			httpRequest.Header.Set("Content-Type", "application/json")
			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, 200, httpWriter.Code)
			if tt.assertion != nil {
				tt.assertion(t, fakeClient)
			}
		})
	}
}

func Test(t *testing.T) {
	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		time.Sleep(time.Second)
		return errors.New("")
	})
	fmt.Println(err)
}
