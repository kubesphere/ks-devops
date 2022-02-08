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
	"reflect"
	"testing"

	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/devops/pkg/jwt/token"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// nolint
	// The fakeclient will undeprecated starting with v0.7.0
	// Reference:
	// - https://github.com/kubernetes-sigs/controller-runtime/issues/768
	// - https://github.com/kubernetes-sigs/controller-runtime/pull/1101
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
	}, {
		name: "Multi-branch Pipeline and SCM set, but the name is written in Chinese",
		args: args{
			prSpec: &v1alpha3.PipelineRunSpec{
				PipelineSpec: &v1alpha3.PipelineSpec{
					Type: v1alpha3.MultiBranchPipelineType,
				},
				SCM: &v1alpha3.SCM{
					RefName: "测试分支",
					RefType: "branch",
				},
			},
		},
		want: "测试分支",
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

var _ = Describe("TestReconciler_hasSamePipelineRun", func() {
	var genernalPipeline *v1alpha3.Pipeline
	var multiBranchPipeline *v1alpha3.Pipeline
	var generalPipelineRun *v1alpha3.PipelineRun
	var multiBranchPipelineRun *v1alpha3.PipelineRun
	var client client.Client

	BeforeEach(func() {
		genernalPipeline = &v1alpha3.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name:      "general-pipeline",
				Namespace: "default",
			},
			Spec: v1alpha3.PipelineSpec{
				Type: v1alpha3.NoScmPipelineType,
			},
		}
		multiBranchPipeline = &v1alpha3.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name:      "multi-branch-pipeline",
				Namespace: "default",
			},
			Spec: v1alpha3.PipelineSpec{
				Type: v1alpha3.MultiBranchPipelineType,
			},
		}

		generalPipelineRun = &v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Name:      "pipeline-run-2",
				Namespace: "default",
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
				},
				Labels: map[string]string{
					v1alpha3.PipelineNameLabelKey: "general-pipeline",
				},
			},
		}

		multiBranchPipelineRun = &v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Name:      "pipeline-run-1",
				Namespace: "default",
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
				},
				Labels: map[string]string{
					v1alpha3.PipelineNameLabelKey: "multi-branch-pipeline",
				},
			},
			Spec: v1alpha3.PipelineRunSpec{
				SCM: &v1alpha3.SCM{
					RefName: "main",
				},
			},
		}
	})

	Context("Multi-branch PipelineRun", func() {
		BeforeEach(func() {
			scheme := scheme.Scheme
			Expect(v1alpha3.AddToScheme(scheme)).To(Succeed())
			client = fake.NewFakeClientWithScheme(scheme, multiBranchPipelineRun)
		})

		It("multi-branch PipelineRun has existed", func() {
			reconciler := &Reconciler{
				Client: client,
			}
			jobRun := &job.PipelineRun{
				BlueItemRun: job.BlueItemRun{
					ID:       "123",
					Pipeline: "main",
				},
			}
			exists, err := reconciler.hasSamePipelineRun(jobRun, multiBranchPipeline)
			Expect(err).To(BeNil())
			Expect(exists).To(BeTrue())
		})

		It("Different run ID", func() {
			reconciler := &Reconciler{
				Client: client,
			}
			jobRun := &job.PipelineRun{
				BlueItemRun: job.BlueItemRun{
					ID:       "non-existent-id",
					Pipeline: "main",
				},
			}
			exists, err := reconciler.hasSamePipelineRun(jobRun, multiBranchPipeline)
			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
		})

		It("Different SCM reference name", func() {
			reconciler := &Reconciler{
				Client: client,
			}
			jobRun := &job.PipelineRun{
				BlueItemRun: job.BlueItemRun{
					ID:       "123",
					Pipeline: "non-existent-branch",
				},
			}
			exists, err := reconciler.hasSamePipelineRun(jobRun, multiBranchPipeline)
			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
		})
	})

	Context("General PipelineRun", func() {
		BeforeEach(func() {
			scheme := scheme.Scheme
			Expect(v1alpha3.AddToScheme(scheme)).To(Succeed())
			client = fake.NewFakeClientWithScheme(scheme, generalPipelineRun)
		})

		It("general PipelineRun has existed", func() {
			reconciler := &Reconciler{
				Client: client,
			}
			jobRun := &job.PipelineRun{
				BlueItemRun: job.BlueItemRun{
					ID:       "123",
					Pipeline: "general-pipeline",
				},
			}
			exists, err := reconciler.hasSamePipelineRun(jobRun, genernalPipeline)
			Expect(err).To(Succeed())
			Expect(exists).To(BeTrue())
		})

		It("Different run ID", func() {
			reconciler := &Reconciler{
				Client: client,
			}
			jobRun := &job.PipelineRun{
				BlueItemRun: job.BlueItemRun{
					ID:       "non-existent-id",
					Pipeline: "general-pipeline",
				},
			}
			exists, err := reconciler.hasSamePipelineRun(jobRun, genernalPipeline)
			Expect(err).To(Succeed())
			Expect(exists).To(BeFalse())
		})
	})
})

func TestReconciler_getOrCreateJenkinsCoreIfHasCreator(t *testing.T) {
	defaultJenkinsCore := &core.JenkinsCore{
		URL:      "https://devops.com",
		UserName: "admin",
		Token:    "fake-token",
	}
	tokenIssuer := token.NewTokenIssuer("test-secret", 0)
	accessToken, err := tokenIssuer.IssueTo(&user.DefaultInfo{Name: "tester"}, token.AccessToken, tokenExpireIn)
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		JenkinsCore core.JenkinsCore
		TokenIssuer token.Issuer
	}
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *core.JenkinsCore
		assertion func(*core.JenkinsCore)
		wantErr   bool
	}{{
		name: "Empty annotations",
		fields: fields{
			JenkinsCore: *defaultJenkinsCore,
			TokenIssuer: tokenIssuer,
		},
		want: defaultJenkinsCore,
	}, {
		name: "Has creator in annotations",
		fields: fields{
			JenkinsCore: *defaultJenkinsCore,
			TokenIssuer: tokenIssuer,
		},
		args: args{
			annotations: map[string]string{
				v1alpha3.PipelineRunCreatorAnnoKey: "tester",
			},
		},
		want: &core.JenkinsCore{
			URL:      defaultJenkinsCore.URL,
			UserName: "tester",
			Token:    accessToken,
		},
		assertion: func(jenkinsCore *core.JenkinsCore) {
			assert.Equal(t, "tester", jenkinsCore.UserName)
			assert.Equal(t, defaultJenkinsCore.URL, jenkinsCore.URL)
			userInfo, tokenType, err := tokenIssuer.Verify(jenkinsCore.Token)
			assert.Equal(t, &user.DefaultInfo{Name: "tester"}, userInfo)
			assert.Equal(t, token.AccessToken, tokenType)
			assert.Nil(t, err)
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				JenkinsCore: tt.fields.JenkinsCore,
				TokenIssuer: tt.fields.TokenIssuer,
			}
			got, err := r.getOrCreateJenkinsCore(tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconciler.getOrCreateJenkinsCoreIfHasCreator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.assertion != nil {
				tt.assertion(got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reconciler.getOrCreateJenkinsCoreIfHasCreator() = %v, want %v", got, tt.want)
			}
		})
	}
}
