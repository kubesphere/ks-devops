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

package v1alpha2

import (
	"context"
	sonargo "github.com/kubesphere/sonargo/sonar"
	"kubesphere.io/devops/pkg/client/s3/fake"
	"kubesphere.io/devops/pkg/client/sonarqube"
	"kubesphere.io/devops/pkg/informers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	fakeclientset "kubesphere.io/devops/pkg/client/clientset/versioned/fake"
	fakedevops "kubesphere.io/devops/pkg/client/devops/fake"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/constants"
)

func TestAPIsExist(t *testing.T) {
	httpWriter := httptest.NewRecorder()
	container := restful.NewContainer()
	assert.NotNil(t, container)

	// case 1, sonarqube client is nil
	_, err := AddToContainer(container, nil, fakedevops.NewFakeDevops(nil),
		nil,
		fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
			ObjectMeta: metav1.ObjectMeta{Name: "fake"},
		}), nil, "", k8s.NewFakeClientSets(k8sfake.NewSimpleClientset(), nil, nil, "", nil,
			fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{Name: "fake"},
			})), core.JenkinsCore{})
	assert.Nil(t, err)

	// case 2, sonarqube client is valid
	container = restful.NewContainer()
	assert.NotNil(t, container)

	k8sclient := k8s.NewFakeClientSets(k8sfake.NewSimpleClientset(), nil, nil, "", nil,
		fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
			ObjectMeta: metav1.ObjectMeta{Name: "fake"},
		}))
	ksclient := fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
		ObjectMeta: metav1.ObjectMeta{Name: "fake"},
	})
	informerFactory := informers.NewInformerFactories(k8sclient.Kubernetes(), ksclient,
		k8sclient.ApiExtensions())

	_, err = AddToContainer(container, informerFactory.KubeSphereSharedInformerFactory(), fakedevops.NewFakeDevops(nil),
		sonarqube.NewSonar(&sonargo.Client{}),
		ksclient, fake.NewFakeS3(), "", k8sclient, core.JenkinsCore{})
	assert.Nil(t, err)

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "credential usage",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/credentials/fake/usage",
		},
	}, {
		name: "get a pipeline",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake",
		},
	}, {
		name: "search API",
		args: args{
			method: http.MethodGet,
			uri:    "/search",
		},
	}, {
		name: "get a pipeline run",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake/runs/fake",
		},
	}, {
		name: "get a pipeline list",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake/runs",
		},
	}, {
		name: "stop a pipeline run",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/pipelines/fake/runs/fake/stop",
		},
	}, {
		name: "replay a pipeline run",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/pipelines/fake/runs/fake/replay",
		},
	}, {
		name: "start a pipeline run",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/pipelines/fake/runs",
		},
	}, {
		name: "get artifacts from a pipeline run",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake/runs/fake/artifacts",
		},
	}, {
		name: "get log output from a pipeline run",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake/runs/fake/log",
		},
	}, {
		name: "get log output from a pipeline run step",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake/runs/fake/nodes/fake/steps/fake/log",
		},
	}, {
		name: "get branches from a pipeline",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake/branches",
		},
	}, {
		name: "scan a pipeline",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/pipelines/fake/scan",
		},
	}, {
		name: "get consolelog from a pipeline",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake/consolelog",
		},
	}, {
		name: "get crumb issuer",
		args: args{
			method: http.MethodGet,
			uri:    "/crumbissuer",
		},
	}, {
		name: "get servers",
		args: args{
			method: http.MethodGet,
			uri:    "/scms/fake/servers",
		},
	}, {
		name: "get organizations",
		args: args{
			method: http.MethodGet,
			uri:    "/scms/fake/organizations",
		},
	}, {
		name: "get repositories",
		args: args{
			method: http.MethodGet,
			uri:    "/scms/fake/organizations/fake/repositories",
		},
	}, {
		name: "scm verify",
		args: args{
			method: http.MethodGet,
			uri:    "/scms/fake/verify",
		},
	}, {
		name: "webhook-git",
		args: args{
			method: http.MethodGet,
			uri:    "/webhook/git",
		},
	}, {
		name: "webhook-git",
		args: args{
			method: http.MethodPost,
			uri:    "/webhook/git",
		},
	}, {
		name: "webhook-github",
		args: args{
			method: http.MethodPost,
			uri:    "/webhook/github",
		},
	}, {
		name: "generic-trigger",
		args: args{
			method: http.MethodPost,
			uri:    "/webhook/generic-trigger",
		},
	}, {
		name: "checkScriptCompile",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/pipelines/fake/checkScriptCompile",
		},
	}, {
		name: "checkCron",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/checkCron",
		},
	}, {
		name: "tojenkinsfile",
		args: args{
			method: http.MethodPost,
			uri:    "/tojenkinsfile",
		},
	}, {
		name: "/devops/{devops}/pipelines/{pipeline}/sonarstatus",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/ns/pipelines/fake/sonarstatus",
		},
	}, {
		name: "/devops/{devops}/pipelines/{pipeline}/branches/{branch}/sonarstatus",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/ns/pipelines/fake/branches/master/sonarstatus",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha2"+tt.args.uri, nil)
			httpRequest = httpRequest.WithContext(context.WithValue(context.TODO(), constants.K8SToken, constants.ContextKeyK8SToken("")))
			container.Dispatch(httpWriter, httpRequest)
			assert.NotEqual(t, httpWriter.Code, 404)
		})
	}
}
