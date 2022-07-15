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

package pipeline

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/golang/mock/gomock"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"github.com/jenkins-zh/jenkins-client/pkg/mock/mhttp"
	"github.com/jenkins-zh/jenkins-client/pkg/util"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	organizationAPIPrefix = "/blue/rest/organizations"
	organization          = "jenkins"
)

func (c *Reconciler) getPipelineAPI(pipelineName string, folders ...string) string {
	api := fmt.Sprintf("%s/%s", organizationAPIPrefix, organization)
	folders = append(folders, pipelineName)
	for _, folder := range folders {
		api = fmt.Sprintf("%s/pipelines/%s", api, folder)
	}
	return api
}

func (c *Reconciler) getPipelineBranchAPI(option *job.GetBranchesOption) string {
	api := c.getPipelineAPI(option.PipelineName, option.Folders...)
	api = api + "/branches/"
	apiURL := &url.URL{
		Path: api,
	}
	query := apiURL.Query()
	if option.Filter != "" {
		query.Add("filter", string(option.Filter))
	}
	if option.Start > 0 {
		query.Add("start", strconv.Itoa(option.Start))
	}
	if option.Limit > 0 {
		query.Add("limit", strconv.Itoa(option.Limit))
	}
	apiURL.RawQuery = query.Encode()
	return apiURL.String()
}

var _ = Describe("Pipeline metadata", func() {
	var (
		ctrl         *gomock.Controller
		c            Reconciler
		roundTripper *mhttp.MockRoundTripper
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = Reconciler{}
		roundTripper = mhttp.NewMockRoundTripper(ctrl)
		c.JenkinsCore.RoundTripper = roundTripper
		c.JenkinsCore.URL = "http://localhost"
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Pipeline Metadata", func() {
		It("Metadata with default namespace", func() {
			pipelineName := "pipelineA"
			api := c.getPipelineAPI(pipelineName, metav1.NamespaceDefault)
			requestURL, _ := util.URLJoinAsString(c.JenkinsCore.URL, api)
			request, _ := http.NewRequest(http.MethodGet, requestURL, nil)
			response := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name":"pipelineA"}`)),
			}
			roundTripper.EXPECT().
				RoundTrip(core.NewRequestMatcher(request)).
				Return(response, nil)

			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pipelineName,
					Namespace:   metav1.NamespaceDefault,
					Annotations: map[string]string{},
				},
			}
			err := c.obtainAndUpdatePipelineMetadata(pipeline)

			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(pipeline.Annotations).NotTo(gomega.BeNil())
			gomega.Expect(pipeline.Annotations[v1alpha3.PipelineJenkinsMetadataAnnoKey]).NotTo(gomega.BeNil())
		})
		It("Metadata with custom namespace", func() {
			pipelineName := "pipelineA"
			customNamespace := "namespaceA"
			api := c.getPipelineAPI(pipelineName, customNamespace)
			requestURL, _ := util.URLJoinAsString(c.JenkinsCore.URL, api)
			request, _ := http.NewRequest(http.MethodGet, requestURL, nil)
			response := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name":"pipelineA"}`)),
			}
			roundTripper.EXPECT().
				RoundTrip(core.NewRequestMatcher(request)).
				Return(response, nil)

			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pipelineName,
					Namespace:   customNamespace,
					Annotations: map[string]string{},
				},
			}
			err := c.obtainAndUpdatePipelineMetadata(pipeline)

			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(pipeline.Annotations).NotTo(gomega.BeNil())
			gomega.Expect(pipeline.Annotations[v1alpha3.PipelineJenkinsMetadataAnnoKey]).NotTo(gomega.BeNil())
		})
	})

	Context("Pipeline Branches", func() {
		given := func(api string, expectedStatus int, responseBody string) {
			request, _ := http.NewRequest(http.MethodGet, api, nil)
			reponse := &http.Response{
				StatusCode: expectedStatus,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			}
			roundTripper.EXPECT().RoundTrip(core.NewRequestMatcher(request)).Return(reponse, nil)
		}
		It("Multi Branch Pipeline default namespace", func() {
			pipelineName := "pipelineA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pipelineName,
					Namespace:   metav1.NamespaceDefault,
					Annotations: map[string]string{},
				},
				Spec: v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
			}
			option := &job.GetBranchesOption{
				Folders:      []string{pipeline.Namespace},
				PipelineName: pipeline.Name,
				Limit:        100000,
			}
			branchAPI := c.getPipelineBranchAPI(option)
			requestURL, _ := util.URLJoinAsString(c.JenkinsCore.URL, branchAPI)
			given(requestURL, http.StatusOK, `[]`)

			err := c.obtainAndUpdatePipelineBranches(pipeline)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(pipeline.Annotations).NotTo(gomega.BeNil())
			gomega.Expect(pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey]).NotTo(gomega.BeNil())
		})
		It("Multi Branch Pipeline custom namespace", func() {
			pipelineName := "pipelineA"
			namespace := "namespaceA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pipelineName,
					Namespace:   namespace,
					Annotations: map[string]string{},
				},
				Spec: v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
			}

			option := &job.GetBranchesOption{
				Folders:      []string{pipeline.Namespace},
				PipelineName: pipeline.Name,
				Limit:        100000,
			}
			branchAPI := c.getPipelineBranchAPI(option)
			requestURL, _ := util.URLJoinAsString(c.JenkinsCore.URL, branchAPI)
			given(requestURL, http.StatusOK, `[]`)

			err := c.obtainAndUpdatePipelineBranches(pipeline)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(pipeline.Annotations).NotTo(gomega.BeNil())
			gomega.Expect(pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey]).NotTo(gomega.BeNil())
		})
		It("Multi Branch Pipeline with query", func() {
			pipelineName := "pipelineA"
			namespace := "namespaceA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pipelineName,
					Namespace:   namespace,
					Annotations: map[string]string{},
				},
				Spec: v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
			}

			option := &job.GetBranchesOption{
				Folders:      []string{pipeline.Namespace},
				PipelineName: pipeline.Name,
				Filter:       job.OriginFilter,
				Start:        123,
				Limit:        456,
			}
			branchAPI := c.getPipelineBranchAPI(option)
			requestURL, _ := util.URLJoinAsString(c.JenkinsCore.URL, branchAPI)
			given(requestURL, http.StatusOK, `[]`)

			err := c.obtainAndUpdatePipelineBranches(pipeline)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(pipeline.Annotations).NotTo(gomega.BeNil())
			gomega.Expect(pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey]).NotTo(gomega.BeNil())
		})
		It("Multi Branch Pipeline Response a branch", func() {
			pipelineName := "pipelineA"
			namespace := "namespaceA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pipelineName,
					Namespace:   namespace,
					Annotations: map[string]string{},
				},
				Spec: v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
			}

			option := &job.GetBranchesOption{
				Folders:      []string{pipeline.Namespace},
				PipelineName: pipeline.Name,
				Limit:        100000,
			}
			branchAPI := c.getPipelineBranchAPI(option)
			requestURL, _ := util.URLJoinAsString(c.JenkinsCore.URL, branchAPI)
			given(requestURL, http.StatusOK, `
[{
   "disabled":false,
   "displayName":"v0.0.1",
   "estimatedDurationInMillis":-1,
   "fullDisplayName":"my-devops-projectsg945/github-pipeline/v0.0.1",
   "fullName":"my-devops-projectsg945/github-pipeline/v0.0.1",
   "latestRun":null,
   "name":"v0.0.1",
   "organization":"jenkins",
   "weatherScore":100,
   "branch":{
      "isPrimary":false,
      "issues":[],
      "url":"https://github.com/JohnNiang/devops-java-thin-sample/tree/v0.0.1"
   }
}]
`)

			err := c.obtainAndUpdatePipelineBranches(pipeline)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(pipeline.Annotations).NotTo(gomega.BeNil())
			gomega.Expect(pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey]).NotTo(gomega.BeNil())
		})

		It("single branch", func() {
			pipelineName := "pipelineA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: pipelineName,
				},
			}
			err := c.obtainAndUpdatePipelineBranches(pipeline)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(len(pipeline.Annotations)).To(gomega.Equal(0))
		})

		It("pipeline Metadata Predicate should call create", func() {
			pipelineName := "pipelineA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: pipelineName,
				},
			}
			instance := pipelineMetadataPredicate
			evt := event.CreateEvent{
				Object: pipeline,
			}

			gomega.Expect(instance.Create(evt)).To(gomega.BeTrue())
		})
		It("pipeline Metadata Predicate should not call delete", func() {
			pipelineName := "pipelineA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: pipelineName,
				},
			}
			instance := pipelineMetadataPredicate
			evt := event.DeleteEvent{
				Object: pipeline,
			}

			gomega.Expect(instance.Delete(evt)).To(gomega.BeFalse())
		})
		It("pipeline Metadata Predicate should not call update", func() {
			pipelineName := "pipelineA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: pipelineName,
				},
			}
			instance := pipelineMetadataPredicate
			evt := event.UpdateEvent{
				ObjectOld: pipeline,
				ObjectNew: pipeline,
			}

			gomega.Expect(instance.Update(evt)).To(gomega.BeFalse())
		})
		It("pipeline Metadata Predicate should not call Generic", func() {
			pipelineName := "pipelineA"
			pipeline := &v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: pipelineName,
				},
			}
			instance := pipelineMetadataPredicate
			evt := event.GenericEvent{
				Object: pipeline,
			}

			gomega.Expect(instance.Generic(evt)).To(gomega.BeFalse())
		})
	})
})
