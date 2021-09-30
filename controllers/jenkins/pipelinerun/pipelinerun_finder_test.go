package pipelinerun

import (
	"reflect"
	"testing"

	"github.com/jenkins-zh/jenkins-client/pkg/job"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

var (
	pipeline1 = &v1alpha3.PipelineRun{
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline1",
			Annotations: map[string]string{
				v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
			},
		},
	}

	pipeline11 = &v1alpha3.PipelineRun{
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline11",
			Annotations: map[string]string{
				v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
			},
		},
	}

	pipeline2 = &v1alpha3.PipelineRun{
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline2",
			Annotations: map[string]string{
				v1alpha3.JenkinsPipelineRunIDAnnoKey: "456",
			},
		},
	}

	multiBranchPipeline1 = &v1alpha3.PipelineRun{
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline1",
			Annotations: map[string]string{
				v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
			},
		},
		Spec: v1alpha3.PipelineRunSpec{
			SCM: &v1alpha3.SCM{
				RefName: "main1",
			},
		},
	}

	multiBranchPipeline11 = &v1alpha3.PipelineRun{
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline11",
			Annotations: map[string]string{
				v1alpha3.JenkinsPipelineRunIDAnnoKey: "123",
			},
		},
		Spec: v1alpha3.PipelineRunSpec{
			SCM: &v1alpha3.SCM{
				RefName: "main1",
			},
		},
	}

	multiBranchPipeline2 = &v1alpha3.PipelineRun{
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline2",
			Annotations: map[string]string{
				v1alpha3.JenkinsPipelineRunIDAnnoKey: "456",
			},
		},
		Spec: v1alpha3.PipelineRunSpec{
			SCM: &v1alpha3.SCM{
				RefName: "main2",
			},
		},
	}
)

func Test_pipelineRunFinder_find(t *testing.T) {
	type args struct {
		run           *job.PipelineRun
		isMultiBranch bool
	}
	tests := []struct {
		name            string
		finder          pipelineRunFinder
		args            args
		wantPipelineRun *v1alpha3.PipelineRun
		wantFound       bool
	}{{
		name:   "Find general PipelineRuns",
		finder: newPipelineRunFinder([]v1alpha3.PipelineRun{*pipeline1, *pipeline2}),
		args: args{
			isMultiBranch: false,
			run: &job.PipelineRun{
				ID: "456",
			},
		},
		wantPipelineRun: pipeline2,
		wantFound:       true,
	}, {
		name:   "Find multi-branch PipelineRuns",
		finder: newPipelineRunFinder([]v1alpha3.PipelineRun{*multiBranchPipeline1, *multiBranchPipeline2}),
		args: args{
			isMultiBranch: true,
			run: &job.PipelineRun{
				ID:       "456",
				Pipeline: "main2",
			},
		},
		wantPipelineRun: multiBranchPipeline2,
		wantFound:       true,
	}, {
		name:   "No PipelineRuns found",
		finder: newPipelineRunFinder([]v1alpha3.PipelineRun{*pipeline1, *pipeline2}),
		args: args{
			isMultiBranch: false,
			run: &job.PipelineRun{
				ID: "invalid_id",
			},
		},
		wantPipelineRun: nil,
		wantFound:       false,
	}, {
		name:   "No Multi-branch PipelineRuns found",
		finder: newPipelineRunFinder([]v1alpha3.PipelineRun{*multiBranchPipeline1, *multiBranchPipeline2}),
		args: args{
			isMultiBranch: true,
			run: &job.PipelineRun{
				ID: "invalid_id",
			},
		},
		wantPipelineRun: nil,
		wantFound:       false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipelineRun, found := tt.finder.find(tt.args.run, tt.args.isMultiBranch)
			if !reflect.DeepEqual(pipelineRun, tt.wantPipelineRun) {
				t.Errorf("pipelineRunFinder.find() got = %v, want %v", pipelineRun, tt.wantPipelineRun)
			}
			if found != tt.wantFound {
				t.Errorf("pipelineRunFinder.find() got1 = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func Test_newPipelineRunFinder(t *testing.T) {
	type args struct {
		pipelineRuns []v1alpha3.PipelineRun
	}
	tests := []struct {
		name string
		args args
		want pipelineRunFinder
	}{{
		name: "General PipelineRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{*pipeline1, *pipeline2},
		},
		want: pipelineRunFinder{
			pipelineRunIdentity{id: "123"}: pipeline1,
			pipelineRunIdentity{id: "456"}: pipeline2,
		},
	}, {
		name: "Duplicated PipelineRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{*pipeline1, *pipeline11, *pipeline2},
		},
		want: pipelineRunFinder{
			pipelineRunIdentity{id: "123"}: pipeline11,
			pipelineRunIdentity{id: "456"}: pipeline2,
		},
	}, {
		name: "Multi-branch PipelineRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{*multiBranchPipeline1, *multiBranchPipeline2},
		},
		want: pipelineRunFinder{
			pipelineRunIdentity{id: "123", refName: "main1"}: multiBranchPipeline1,
			pipelineRunIdentity{id: "456", refName: "main2"}: multiBranchPipeline2,
		},
	}, {
		name: "Duplicated multi-branch PipelineRuns",
		args: args{
			pipelineRuns: []v1alpha3.PipelineRun{*multiBranchPipeline1, *multiBranchPipeline11, *multiBranchPipeline2},
		},
		want: pipelineRunFinder{
			pipelineRunIdentity{id: "123", refName: "main1"}: multiBranchPipeline11,
			pipelineRunIdentity{id: "456", refName: "main2"}: multiBranchPipeline2,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPipelineRunFinder(tt.args.pipelineRuns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newPipelineRunFinder() = %v, want %v", got, tt.want)
			}
		})
	}
}
