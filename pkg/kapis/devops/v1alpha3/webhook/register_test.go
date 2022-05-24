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
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"io"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/devops/pkg/jwt/token"
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
			RegisterWebhooks(fakeClient, wsWithGroup, &token.FakeIssuer{}, core.JenkinsCore{})
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

func TestSCMWebhook(t *testing.T) {
	defaultPipeline := &v1alpha3.Pipeline{}
	defaultPipeline.SetName("fake")
	defaultPipeline.SetClusterName("default")
	defaultPipeline.Spec.MultiBranchPipeline = &v1alpha3.MultiBranchPipeline{
		GitSource: &v1alpha3.GitSource{
			Url: "https://gitlab.com/linuxsuren/test",
		},
	}
	defaultPipeline.SetAnnotations(map[string]string{
		v1alpha3.PipelineJenkinsBranchesAnnoKey: `[{"name":"master"}]`,
	})

	type args struct {
		method     string
		uri        string
		bodyJSON   string
		header     map[string]string
		initObject []runtime.Object
	}
	tests := []struct {
		name      string
		args      args
		assertion func(t *testing.T, c client.Client, body string)
	}{{
		name: "unknown SCM webhook",
		args: args{
			method:     http.MethodPost,
			uri:        "/webhooks/scm",
			initObject: []runtime.Object{},
		},
		assertion: func(t *testing.T, c client.Client, body string) {
			assert.Equal(t, "unknown SCM type", body)
		},
	}, {
		name: "gitlab webhook with no pipeline matched",
		args: args{
			method:     http.MethodPost,
			uri:        "/webhooks/scm",
			initObject: []runtime.Object{},
			bodyJSON:   gitlabWebhookBody,
			header: map[string]string{
				"X-Gitlab-Event": "Push Hook",
			},
		},
		assertion: func(t *testing.T, c client.Client, body string) {
			assert.Equal(t, "no pipeline matched", body)
		},
	}, {
		name: "gitlab webhook",
		args: args{
			method:     http.MethodPost,
			uri:        "/webhooks/scm",
			initObject: []runtime.Object{defaultPipeline.DeepCopy()},
			bodyJSON:   gitlabWebhookBody,
			header: map[string]string{
				"X-Gitlab-Event": "Push Hook",
			},
		},
		assertion: func(t *testing.T, c client.Client, body string) {
			assert.Equal(t, "ok", body)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilruntime.Must(v1alpha3.AddToScheme(scheme.Scheme))
			fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, tt.args.initObject...)

			container := restful.NewContainer()
			wsWithGroup := apiserverruntime.NewWebService(v1alpha3.GroupVersion)
			RegisterWebhooks(fakeClient, wsWithGroup, &token.FakeIssuer{}, core.JenkinsCore{})
			container.Add(wsWithGroup)

			var bodyReader io.Reader
			if tt.args.bodyJSON != "" {
				bodyReader = strings.NewReader(tt.args.bodyJSON)
			}

			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, bodyReader)
			httpRequest.Header.Set("Content-Type", "application/json")
			for k, v := range tt.args.header {
				httpRequest.Header.Set(k, v)
			}
			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, 200, httpWriter.Code)
			if tt.assertion != nil {
				body := httpWriter.Body
				var bodyResponse string
				if body != nil {
					bodyResponse = body.String()
				}
				tt.assertion(t, fakeClient, bodyResponse)
			}
		})
	}
}

const gitlabWebhookBody = `{
  "object_kind": "push",
  "event_name": "push",
  "before": "8f4b347e7d6b7647b51647dcd07ddafd4bded19f",
  "after": "bd4f171cec5c6f9b8b184107ce318bf9a54dce26",
  "ref": "refs/heads/master",
  "checkout_sha": "bd4f171cec5c6f9b8b184107ce318bf9a54dce26",
  "message": null,
  "user_id": 1269616,
  "user_name": "Rick",
  "user_username": "linuxsuren",
  "user_email": "",
  "user_avatar": "https://secure.gravatar.com/avatar/1f14a87866e88e19426d164e4d3e078a?s=80&d=identicon",
  "project_id": 26266137,
  "project": {
    "id": 26266137,
    "name": "test",
    "description": "",
    "web_url": "https://gitlab.com/linuxsuren/test",
    "avatar_url": null,
    "git_ssh_url": "git@gitlab.com:linuxsuren/test.git",
    "git_http_url": "https://gitlab.com/linuxsuren/test.git",
    "namespace": "Rick",
    "visibility_level": 20,
    "path_with_namespace": "linuxsuren/test",
    "default_branch": "master",
    "ci_config_path": "",
    "homepage": "https://gitlab.com/linuxsuren/test",
    "url": "git@gitlab.com:linuxsuren/test.git",
    "ssh_url": "git@gitlab.com:linuxsuren/test.git",
    "http_url": "https://gitlab.com/linuxsuren/test.git"
  },
  "commits": [
    {
      "id": "bd4f171cec5c6f9b8b184107ce318bf9a54dce26",
      "message": "Merge branch '10' into 'master'\n\nAdd new file\n\nSee merge request linuxsuren/test!1",
      "title": "Merge branch '10' into 'master'",
      "timestamp": "2021-04-29T10:41:08+00:00",
      "url": "https://gitlab.com/linuxsuren/test/-/commit/bd4f171cec5c6f9b8b184107ce318bf9a54dce26",
      "author": {
        "name": "Rick",
        "email": "linuxsuren@gmail.com"
      },
      "added": [
        "Jenkinsfile"
      ],
      "modified": [

      ],
      "removed": [

      ]
    },
    {
      "id": "1c433d04a2c306c414d8542908fc30ab8b855c65",
      "message": "Add new file",
      "title": "Add new file",
      "timestamp": "2021-04-29T10:40:41+00:00",
      "url": "https://gitlab.com/linuxsuren/test/-/commit/1c433d04a2c306c414d8542908fc30ab8b855c65",
      "author": {
        "name": "Rick",
        "email": "linuxsuren@gmail.com"
      },
      "added": [
        "Jenkinsfile"
      ],
      "modified": [

      ],
      "removed": [

      ]
    },
    {
      "id": "8f4b347e7d6b7647b51647dcd07ddafd4bded19f",
      "message": "Initial commit",
      "title": "Initial commit",
      "timestamp": "2021-04-29T10:37:37+00:00",
      "url": "https://gitlab.com/linuxsuren/test/-/commit/8f4b347e7d6b7647b51647dcd07ddafd4bded19f",
      "author": {
        "name": "Rick",
        "email": "linuxsuren@gmail.com"
      },
      "added": [
        "README.md"
      ],
      "modified": [

      ],
      "removed": [

      ]
    }
  ],
  "total_commits_count": 3,
  "push_options": {
  },
  "repository": {
    "name": "test",
    "url": "git@gitlab.com:linuxsuren/test.git",
    "description": "",
    "homepage": "https://gitlab.com/linuxsuren/test",
    "git_http_url": "https://gitlab.com/linuxsuren/test.git",
    "git_ssh_url": "git@gitlab.com:linuxsuren/test.git",
    "visibility_level": 20
  }
}`

func Test(t *testing.T) {
	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		time.Sleep(time.Second)
		return errors.New("")
	})
	fmt.Println(err)
}
