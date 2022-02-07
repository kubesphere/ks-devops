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
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/mock/mhttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

func Test_getJenkinsBuildNumber(t *testing.T) {
	type args struct {
		pipelineRun *v1alpha3.PipelineRun
	}
	tests := []struct {
		name    string
		args    args
		wantNum int
	}{{
		name: "no build number",
		args: args{
			pipelineRun: &v1alpha3.PipelineRun{},
		},
		wantNum: -1,
	}, {
		name: "invalid build number",
		args: args{
			pipelineRun: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunIDAnnoKey: "a",
					},
				},
			},
		},
		wantNum: -1,
	}, {
		name: "valid build number",
		args: args{
			pipelineRun: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunIDAnnoKey: "2",
					},
				},
			},
		},
		wantNum: 2,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNum := getJenkinsBuildNumber(tt.args.pipelineRun); gotNum != tt.wantNum {
				t.Errorf("getJenkinsBuildNumber() = %v, want %v", gotNum, tt.wantNum)
			}
		})
	}
}

var _ = Describe("Test deleteJenkinsJobHistory", func() {
	var (
		ctrl         *gomock.Controller
		roundTripper *mhttp.MockRoundTripper
		reconciler   *Reconciler
		jHandler     *jenkinsHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		roundTripper = mhttp.NewMockRoundTripper(ctrl)
		reconciler = &Reconciler{
			JenkinsCore: core.JenkinsCore{
				URL:          "http://localhost",
				UserName:     "",
				Token:        "",
				RoundTripper: roundTripper,
			},
		}
		jHandler = &jenkinsHandler{&reconciler.JenkinsCore}
	})

	It("delete an empty PipelineRun", func() {
		err := jHandler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("delete a valid PipelineRun", func() {
		namespace := "project1"
		pipelineName := "testPipeline"

		requestCrumb, _ := http.NewRequest(http.MethodGet, "http://localhost/crumbIssuer/api/json", nil)
		responseCrumb := &http.Response{
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			Request:    requestCrumb,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{"crumbRequestField":"CrumbRequestField","crumb":"Crumb"}
				`)),
		}
		roundTripper.EXPECT().
			RoundTrip(core.NewRequestMatcher(requestCrumb)).Return(responseCrumb, nil)

		request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost/job/%s/job/%s/2/doDelete", namespace, pipelineName), nil)
		request.Header.Set("CrumbRequestField", "Crumb")
		response := &http.Response{
			Request:    request,
			StatusCode: http.StatusOK,
		}
		roundTripper.EXPECT().
			RoundTrip(core.NewRequestMatcher(request)).Return(response, nil)

		err := jHandler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDAnnoKey: "2",
				},
			},
			Spec: v1alpha3.PipelineRunSpec{
				PipelineRef: &corev1.ObjectReference{
					Name: pipelineName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("to delete a not exist Jenkins build history", func() {
		namespace := "project1"
		pipelineName := "testPipeline"

		requestCrumb, _ := http.NewRequest(http.MethodGet, "http://localhost/crumbIssuer/api/json", nil)
		responseCrumb := &http.Response{
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			Request:    requestCrumb,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{"crumbRequestField":"CrumbRequestField","crumb":"Crumb"}
				`)),
		}
		roundTripper.EXPECT().
			RoundTrip(core.NewRequestMatcher(requestCrumb)).Return(responseCrumb, nil)

		request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost/job/%s/job/%s/2/doDelete", namespace, pipelineName), nil)
		request.Header.Set("CrumbRequestField", "Crumb")
		response := &http.Response{
			Request:    request,
			StatusCode: http.StatusNotFound,
		}
		roundTripper.EXPECT().
			RoundTrip(core.NewRequestMatcher(request)).Return(response, nil)

		err := jHandler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDAnnoKey: "2",
				},
			},
			Spec: v1alpha3.PipelineRunSpec{
				PipelineRef: &corev1.ObjectReference{
					Name: pipelineName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("failed to delete Jenkins build history", func() {
		namespace := "project1"
		pipelineName := "testPipeline"

		requestCrumb, _ := http.NewRequest(http.MethodGet, "http://localhost/crumbIssuer/api/json", nil)
		responseCrumb := &http.Response{
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			Request:    requestCrumb,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{"crumbRequestField":"CrumbRequestField","crumb":"Crumb"}
				`)),
		}
		roundTripper.EXPECT().
			RoundTrip(core.NewRequestMatcher(requestCrumb)).Return(responseCrumb, nil)

		request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost/job/%s/job/%s/2/doDelete", namespace, pipelineName), nil)
		request.Header.Set("CrumbRequestField", "Crumb")
		response := &http.Response{
			Request:    request,
			StatusCode: http.StatusOK,
		}
		roundTripper.EXPECT().
			RoundTrip(core.NewRequestMatcher(request)).Return(response, errors.New("failed"))

		err := jHandler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDAnnoKey: "2",
				},
			},
			Spec: v1alpha3.PipelineRunSpec{
				PipelineRef: &corev1.ObjectReference{
					Name: pipelineName,
				},
			},
		})
		Expect(err).To(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
	})
})

func Test_getJenkinsJobPath(t *testing.T) {
	type args struct {
		pipelineRun *v1alpha3.PipelineRun
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "empty PipelineRun",
		args: args{
			pipelineRun: &v1alpha3.PipelineRun{},
		},
		want: "",
	}, {
		name: "a regular PipelineRun",
		args: args{
			pipelineRun: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "ns",
					Name:      "run",
				},
				Spec: v1alpha3.PipelineRunSpec{
					PipelineRef: &corev1.ObjectReference{
						Namespace: "ns",
						Name:      "pipeline",
					},
				},
			},
		},
		want: "/job/ns/job/pipeline",
	}, {
		name: "a regular PipelineRun without ns in the ref",
		args: args{
			pipelineRun: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "ns",
					Name:      "run",
				},
				Spec: v1alpha3.PipelineRunSpec{
					PipelineRef: &corev1.ObjectReference{
						Name: "pipeline",
					},
				},
			},
		},
		want: "/job/ns/job/pipeline",
	}, {
		name: "a multi-branch PipelineRun",
		args: args{
			pipelineRun: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "ns",
					Name:      "run",
				},
				Spec: v1alpha3.PipelineRunSpec{
					SCM: &v1alpha3.SCM{
						RefName: "master",
					},
					PipelineRef: &corev1.ObjectReference{
						Name: "pipeline",
					},
				},
			},
		},
		want: "/job/ns/job/pipeline/job/master",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getJenkinsJobPath(tt.args.pipelineRun), "getJenkinsJobPath(%v)", tt.args.pipelineRun)
		})
	}
}
