package pipelinerun

import (
	"bytes"
	"errors"
	"fmt"
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

func Test_getBranch(t *testing.T) {
	type args struct {
		prSpec *v1alpha3.PipelineRunSpec
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{{
		name: "No SCM Pipeline",
		args: args{
			prSpec: &v1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.NoScmPipelineType,
				},
			},
		},
		want: "",
	}, {
		name: "No SCM Pipeline but SCM set",
		args: args{
			prSpec: &v1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.NoScmPipelineType,
				},
				SCM: &v1alpha3.SCM{
					RefName: "main",
					RefType: "branch",
				},
			},
		},
		want: "",
	}, {
		name: "Multi-branch Pipeline but not SCM set",
		args: args{
			prSpec: &v1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
			},
		},
		wantErr: true,
	}, {
		name: "Multi-branch Pipeline and SCM set",
		args: args{
			prSpec: &v1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
				SCM: &v1alpha3.SCM{
					RefName: "main",
					RefType: "branch",
				},
			},
		},
		want: "main",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSCMRefName(tt.args.prSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSCMRefName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSCMRefName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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
						v1alpha3.JenkinsPipelineRunIDKey: "a",
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
						v1alpha3.JenkinsPipelineRunIDKey: "2",
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
	})

	It("delete an empty PipelineRun", func() {
		err := reconciler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{})
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

		err := reconciler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDKey: "2",
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

		err := reconciler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDKey: "2",
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

		err := reconciler.deleteJenkinsJobHistory(&v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDKey: "2",
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
