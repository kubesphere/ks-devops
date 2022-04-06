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

	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func createPipelineRun(name string, runID string) v1alpha3.PipelineRun {
	return v1alpha3.PipelineRun{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "fake-pipelinerun-namespace",
			Annotations: map[string]string{
				v1alpha3.JenkinsPipelineRunIDAnnoKey: runID,
			},
		},
	}
}

func createMultiBranchPipelineRun(name, runID, scmRefName string) v1alpha3.PipelineRun {
	pipelineRun := createPipelineRun(name, runID)
	pipelineRun.Spec.SCM = &v1alpha3.SCM{
		RefName: scmRefName,
	}
	return pipelineRun
}

func createJobRun(name, runID string) job.PipelineRun {
	return job.PipelineRun{
		BlueItemRun: job.BlueItemRun{
			Name: name,
			ID:   runID,
		},
	}
}
func createMultiBranchJobRun(name, runID, scmRefName string) job.PipelineRun {
	jobRun := createJobRun(name, runID)
	jobRun.Pipeline = scmRefName
	return jobRun
}

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
		name string
		args args
		want *v1alpha3.PipelineRun
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
			got := createBarePipelineRun(tt.args.pipeline, tt.args.run)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createBarePipelineRun() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func Test_collectExistingPipelineRuns(t *testing.T) {
	type args struct {
		pipelineRuns  []v1alpha3.PipelineRun
		isMultiBranch bool
		jobRuns       []job.PipelineRun
	}
	tests := []struct {
		name                    string
		args                    args
		pipelineRunsToBeDeleted []v1alpha3.PipelineRun
	}{{
		name: "Should return empty PipelineRun set if no PipelineRuns and no JobRuns",
		args: args{
			pipelineRuns:  nil,
			isMultiBranch: false,
			jobRuns:       nil,
		},
		pipelineRunsToBeDeleted: nil,
	}, {
		name: "Should return empty PipelineRun set if no PipelineRuns",
		args: args{
			pipelineRuns:  nil,
			isMultiBranch: false,
			jobRuns:       []job.PipelineRun{createJobRun("fake-jobrun", "1")},
		},
		pipelineRunsToBeDeleted: nil,
	}, {
		name: "Should return empty PipelineRun set if no JobRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{
				createPipelineRun("fake-pipelinerun", "1"),
			},
			isMultiBranch: false,
			jobRuns:       nil,
		},
		pipelineRunsToBeDeleted: []v1alpha3.PipelineRun{
			createPipelineRun("fake-pipelinerun", "1"),
		},
	}, {
		name: "Should found a PipelineRun if PipelineRuns contain all of JobRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{
				createPipelineRun("fake-pipeline-1", "1"),
				createPipelineRun("fake-pipeline-2", "2"),
				createPipelineRun("fake-pipeline-3", "3"),
			},
			jobRuns: []job.PipelineRun{
				createJobRun("fake-jobrun-1", "1"),
				createJobRun("fake-jobrun-2", "2"),
			},
		},
		pipelineRunsToBeDeleted: []v1alpha3.PipelineRun{
			createPipelineRun("fake-pipeline-3", "3"),
		},
	}, {
		name: "Should found a PipelineRun if PipelineRuns contain one of JobRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{
				createPipelineRun("fake-pipeline-1", "1"),
				createPipelineRun("fake-pipeline-2", "2"),
			},
			jobRuns: []job.PipelineRun{
				createJobRun("fake-jobrun-1", "1"),
				createJobRun("fake-jobrun-3", "3"),
			},
		},
		pipelineRunsToBeDeleted: []v1alpha3.PipelineRun{
			createPipelineRun("fake-pipeline-2", "2"),
		},
	}, {
		name: "Should found a multi-branch PipelineRun if PipelineRuns contain one of JobRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{
				createMultiBranchPipelineRun("fake-pipeline-1", "1", "main"),
				createMultiBranchPipelineRun("fake-pipeline-2", "1", "dev"),
			},
			jobRuns: []job.PipelineRun{
				createMultiBranchJobRun("fake-jobrun-1", "1", "main"),
			},
			isMultiBranch: true,
		},
		pipelineRunsToBeDeleted: []v1alpha3.PipelineRun{
			createMultiBranchPipelineRun("fake-pipeline-2", "1", "dev"),
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectPipelineRunsDeletedInJenkins(newPipelineRunFinder(tt.args.pipelineRuns), tt.args.isMultiBranch, tt.args.jobRuns, tt.args.pipelineRuns); !reflect.DeepEqual(got, tt.pipelineRunsToBeDeleted) {
				t.Errorf("collectExistingPipelineRuns() = %v, want %v", got, tt.pipelineRunsToBeDeleted)
			}
		})
	}
}

func Test_createBarePipelineRunsIfNotPresent(t *testing.T) {
	jobRun1 := createJobRun("fake-jobrun-1", "1")
	jobRun2 := createJobRun("fake-jobrun-2", "2")
	mainJobRun1 := createMultiBranchJobRun("fake-jobrun-main-1", "1", "main")
	devJobRun1 := createMultiBranchJobRun("fake-jobrun-dev-1", "1", "dev")

	pipeline := &v1alpha3.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      "fake-pipeline",
			Namespace: "fake-pipeline-namespace",
		},
	}
	multiBranchPipeline := &v1alpha3.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      "fake-pipeline",
			Namespace: "fake-pipeline-namespace",
		},
		Spec: v1alpha3.PipelineSpec{
			Type: v1alpha3.MultiBranchPipelineType,
		},
	}
	type args struct {
		pipelineRuns []v1alpha3.PipelineRun
		pipeline     *v1alpha3.Pipeline
		jobRuns      []job.PipelineRun
	}
	tests := []struct {
		name                 string
		args                 args
		expectedPipelineRuns []v1alpha3.PipelineRun
	}{{
		name: "Should create nothing when JobRuns are empty",
		args: args{
			jobRuns: nil,
		},
		expectedPipelineRuns: nil,
	}, {
		name: "Should create bare PipelineRuns for all JobRuns when PipelineRuns are empty but JobRuns are not empty",
		args: args{
			jobRuns: []job.PipelineRun{
				jobRun1,
				jobRun2,
			},
			pipeline: pipeline,
		},
		expectedPipelineRuns: []v1alpha3.PipelineRun{
			*createBarePipelineRun(pipeline, &jobRun1),
			*createBarePipelineRun(pipeline, &jobRun2),
		},
	}, {
		name: "Should create bare PipelineRun for one of JobRuns when PipelineRuns contain some of JobRuns",
		args: args{
			pipeline: pipeline,
			jobRuns: []job.PipelineRun{
				jobRun1,
				jobRun2,
			},
			pipelineRuns: []v1alpha3.PipelineRun{
				createPipelineRun("fake-pipeline-1", "1"),
			},
		},
		expectedPipelineRuns: []v1alpha3.PipelineRun{
			*createBarePipelineRun(pipeline, &jobRun2),
		},
	}, {
		name: "Should create bare PipelineRun for one of JobRuns when PipelineRuns contains some of JobRuns and Pipeline is multi-branch type",
		args: args{
			pipeline: multiBranchPipeline,
			jobRuns: []job.PipelineRun{
				mainJobRun1,
				devJobRun1,
			},
			pipelineRuns: []v1alpha3.PipelineRun{
				createMultiBranchPipelineRun("fake-pipelinerun-main-1", "1", "main"),
				createMultiBranchPipelineRun("fake-pipelinerun-test-1", "1", "test"),
			},
		},
		expectedPipelineRuns: []v1alpha3.PipelineRun{
			*createBarePipelineRun(multiBranchPipeline, &devJobRun1),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createBarePipelineRunsIfNotPresent(newPipelineRunFinder(tt.args.pipelineRuns), tt.args.pipeline, tt.args.jobRuns); !reflect.DeepEqual(got, tt.expectedPipelineRuns) {
				t.Errorf("createBarePipelineRunsIfNotPresent() = %v, want %v", got, tt.expectedPipelineRuns)
			}
		})
	}
}

func Test_hasPendingPipelineRuns(t *testing.T) {
	type args struct {
		items []v1alpha3.PipelineRun
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "have pending PipelineRuns",
		args: args{
			items: []v1alpha3.PipelineRun{{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"a": "1",
					},
				},
			}},
		},
		want: true,
	}, {
		name: "not have pending PipelineRuns",
		args: args{
			items: []v1alpha3.PipelineRun{{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha3.JenkinsPipelineRunIDAnnoKey: "1",
					},
				},
			}},
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, hasPendingPipelineRuns(tt.args.items), "hasPendingPipelineRuns(%v)", tt.args.items)
		})
	}
}
