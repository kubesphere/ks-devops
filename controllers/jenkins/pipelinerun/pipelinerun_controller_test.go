package pipelinerun

import (
	"testing"

	"github.com/jenkins-zh/jenkins-client/pkg/job"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/clientset/versioned/scheme"
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
					v1alpha3.SCMRefNameLabelKey:   "main",
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
