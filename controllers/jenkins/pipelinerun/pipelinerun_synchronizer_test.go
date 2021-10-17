package pipelinerun

import (
	"reflect"
	"testing"

	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func Test_requestSyncPredicate(t *testing.T) {
	tests := []struct {
		name string
		want func(predicate.Predicate)
	}{{
		name: "Request to sync while creating",
		want: func(p predicate.Predicate) {
			createEvent := event.CreateEvent{
				Meta: &v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.PipelineRequestToSyncRunsAnnoKey: "true",
					},
				},
			}
			assert.True(t, p.Create(createEvent))
		},
	}, {
		name: "Request to nothing while creating",
		want: func(p predicate.Predicate) {
			createEvent := event.CreateEvent{
				Meta: &v1.ObjectMeta{},
			}
			assert.False(t, p.Create(createEvent))
		},
	}, {
		name: "Request to sync while updating",
		want: func(p predicate.Predicate) {
			updateEvent := event.UpdateEvent{
				MetaNew: &v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.PipelineRequestToSyncRunsAnnoKey: "true",
					},
				},
			}
			assert.True(t, p.Update(updateEvent))
		},
	}, {
		name: "Request to nothing while updating",
		want: func(p predicate.Predicate) {
			updateEvent := event.UpdateEvent{
				MetaNew: &v1.ObjectMeta{},
			}
			assert.False(t, p.Update(updateEvent))
		},
	}, {
		name: "Nothing to do while deleting",
		want: func(p predicate.Predicate) {
			deleteEvent := event.DeleteEvent{
				Meta: &v1.ObjectMeta{
					Annotations: map[string]string{},
				},
			}
			assert.False(t, p.Delete(deleteEvent))
		},
	}, {
		name: "Nothing to do while genericing",
		want: func(p predicate.Predicate) {
			genericEvent := event.GenericEvent{
				Meta: &v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.PipelineRequestToSyncRunsAnnoKey: "true",
					},
				},
			}
			assert.False(t, p.Generic(genericEvent))
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want(requestSyncPredicate())
		})
	}
}

func Test_createBarePipelineRun(t *testing.T) {
	multiBranchPipeline := &v1alpha3.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      "fake-pipeline",
			Namespace: "fake-namespace",
		},
		Spec: v1alpha3.PipelineSpec{
			Type: v1alpha3.MultiBranchPipelineType,
		},
	}

	generalPipeline := &v1alpha3.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      "fake-pipeline",
			Namespace: "fake-namespace",
		},
		Spec: v1alpha3.PipelineSpec{
			Type: v1alpha3.NoScmPipelineType,
		},
	}

	type args struct {
		pipeline *v1alpha3.Pipeline
		run      *job.PipelineRun
	}
	tests := []struct {
		name    string
		args    args
		want    *v1alpha3.PipelineRun
		wantErr bool
	}{{
		name: "Multi-branch pipeline",
		args: args{
			pipeline: multiBranchPipeline,
			run: &job.PipelineRun{
				BlueItemRun: job.BlueItemRun{
					Pipeline: "main",
					ID:       "123",
				},
				Branch: &job.Branch{
					URL: "main",
				},
			},
		},
		want: &v1alpha3.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Namespace:    "fake-namespace",
				GenerateName: "fake-pipeline-",
				OwnerReferences: []v1.OwnerReference{
					*v1.NewControllerRef(multiBranchPipeline, multiBranchPipeline.GroupVersionKind()),
				},
				Annotations: map[string]string{
					v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
				},
				Labels: map[string]string{
					v1alpha3.PipelineNameLabelKey: "fake-pipeline",
					v1alpha3.SCMRefNameLabelKey:   "main",
				},
			},
			Spec: v1alpha3.PipelineRunSpec{
				SCM: &v1alpha3.SCM{
					RefName: "main",
					RefType: "",
				},
				PipelineRef: &corev1.ObjectReference{
					Namespace: "fake-namespace",
					Name:      "fake-pipeline",
				},
				PipelineSpec: &multiBranchPipeline.Spec,
			},
		},
	},
		{
			name: "No SCM pipeline",
			args: args{
				pipeline: generalPipeline,
				run: &job.PipelineRun{
					BlueItemRun: job.BlueItemRun{
						ID: "123",
					},
				},
			},
			want: &v1alpha3.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name:         "",
					Namespace:    "fake-namespace",
					GenerateName: "fake-pipeline-",
					OwnerReferences: []v1.OwnerReference{
						*v1.NewControllerRef(generalPipeline, generalPipeline.GroupVersionKind()),
					},
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
					},
					Labels: map[string]string{
						v1alpha3.PipelineNameLabelKey: "fake-pipeline",
					},
				},
				Spec: v1alpha3.PipelineRunSpec{
					PipelineRef: &corev1.ObjectReference{
						Namespace: "fake-namespace",
						Name:      "fake-pipeline",
					},
					PipelineSpec: &generalPipeline.Spec}},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createBarePipelineRun(tt.args.pipeline, tt.args.run)
			if (err != nil) != tt.wantErr {
				t.Errorf("createBarePipelineRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createBarePipelineRun() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
