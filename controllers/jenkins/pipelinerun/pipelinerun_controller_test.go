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
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	ctrlCore "github.com/kubesphere/ks-devops/controllers/core"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/client/clientset/versioned/scheme"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
			ObjectMeta: metav1.ObjectMeta{
				Name:      "general-pipeline",
				Namespace: "default",
			},
			Spec: v1alpha3.PipelineSpec{
				Type: v1alpha3.NoScmPipelineType,
			},
		}
		multiBranchPipeline = &v1alpha3.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "multi-branch-pipeline",
				Namespace: "default",
			},
			Spec: v1alpha3.PipelineSpec{
				Type: v1alpha3.MultiBranchPipelineType,
			},
		}

		generalPipelineRun = &v1alpha3.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
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
			ObjectMeta: metav1.ObjectMeta{
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
			client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(multiBranchPipelineRun).Build()
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
			client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(generalPipelineRun).Build()
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

func TestPipelineRunReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)
	err = v1alpha1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type fields struct {
		Client   client.Client
		recorder record.EventRecorder
	}
	type args struct {
		mgr controllerruntime.Manager
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "normal",
		args: args{
			mgr: &ctrlCore.FakeManager{
				Client: fake.NewClientBuilder().WithScheme(schema).Build(),
				Scheme: schema,
			},
		},
		wantErr: ctrlCore.NoErrors,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client:   tt.fields.Client,
				log:      logr.Logger{},
				recorder: tt.fields.recorder,
			}
			tt.wantErr(t, r.SetupWithManager(tt.args.mgr), fmt.Sprintf("SetupWithManager(%v)", tt.args.mgr))
		})
	}
}

func TestPipelineRunReconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)
	err = v1alpha1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	pipelineRun := v1alpha3.PipelineRun{}
	pipelineRun.SetName("name")
	pipelineRun.SetNamespace("ns")

	now := metav1.Now()
	completedPipelineRun := pipelineRun.DeepCopy()
	completedPipelineRun.Status.CompletionTime = &now

	normalPipeline := pipelineRun.DeepCopy()
	normalPipeline.Spec.PipelineRef = &v1.ObjectReference{
		Namespace: "ns",
		Name:      "pipeline",
	}

	defaultReq := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "ns", Name: "name",
		},
	}

	tests := []struct {
		name      string
		k8sclient client.Client
		request   ctrl.Request
		wantErr   bool
	}{{
		name:      "not found",
		k8sclient: fake.NewClientBuilder().WithScheme(schema).Build(),
		request:   defaultReq,
		wantErr:   false,
	}, {
		name:      "orphan",
		k8sclient: fake.NewClientBuilder().WithScheme(schema).WithObjects(pipelineRun.DeepCopy()).Build(),
		request:   defaultReq,
		wantErr:   false,
	}, {
		name:      "completed PipelineRun",
		k8sclient: fake.NewClientBuilder().WithScheme(schema).WithObjects(completedPipelineRun.DeepCopy()).Build(),
		request:   defaultReq,
		wantErr:   false,
	}, {
		name:      "no corresponding Pipeline",
		k8sclient: fake.NewClientBuilder().WithScheme(schema).WithObjects(normalPipeline.DeepCopy()).Build(),
		request:   defaultReq,
		wantErr:   false,
	}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client: tt.k8sclient,
				log:    logr.New(log.NullLogSink{}),
			}
			_, err = r.Reconcile(context.Background(), tt.request)
			if tt.wantErr {
				assert.NotNil(t, err, "failed in [%d]", i)
			} else {
				assert.Nil(t, err, "failed in [%d]", i)
			}
		})
	}
}

func TestStorePipelineRunData(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)
	err = v1alpha1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	pipelineRun := v1alpha3.PipelineRun{}
	pipelineRun.SetName("name")
	pipelineRun.SetNamespace("ns")

	r := &Reconciler{
		Client:               fake.NewClientBuilder().WithScheme(schema).WithObjects(pipelineRun.DeepCopy()).Build(),
		log:                  logr.New(log.NullLogSink{}),
		PipelineRunDataStore: "fake",
	}
	assert.NotNil(t, r.storePipelineRunData("", "", pipelineRun.DeepCopy()))

	r = &Reconciler{
		Client: fake.NewClientBuilder().WithScheme(schema).WithObjects(pipelineRun.DeepCopy()).Build(),
		log:    logr.New(log.NullLogSink{}),
		req: ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "name", Namespace: "ns"},
		},
		PipelineRunDataStore: "configmap",
	}
	assert.Nil(t, r.storePipelineRunData("", "", pipelineRun.DeepCopy()))

	r = &Reconciler{
		Client:               fake.NewClientBuilder().WithScheme(schema).WithObjects(pipelineRun.DeepCopy()).Build(),
		log:                  logr.New(log.NullLogSink{}),
		PipelineRunDataStore: "",
	}
	assert.Nil(t, r.storePipelineRunData("", "", pipelineRun.DeepCopy()))
}
