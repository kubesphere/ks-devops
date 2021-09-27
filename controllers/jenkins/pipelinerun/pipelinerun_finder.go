package pipelinerun

import (
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

// pipelineRunIdentity holds id and SCM reference name to identity unique PipelineRun.
type pipelineRunIdentity struct {
	id      string
	refName string
}

type pipelineRunFinder map[pipelineRunIdentity]*v1alpha3.PipelineRun

func (finder pipelineRunFinder) find(run *job.PipelineRun, isMultiBranch bool) (*v1alpha3.PipelineRun, bool) {
	identity := pipelineRunIdentity{
		id: run.ID,
	}
	if isMultiBranch {
		identity.refName = run.Pipeline
	}
	pipelineRun, ok := finder[identity]
	return pipelineRun, ok
}

func newPipelineRunFinder(pipelineRuns []v1alpha3.PipelineRun) pipelineRunFinder {
	// convert PipelineRuns to map[pipelineRunIdentity]PipelineRun
	finder := pipelineRunFinder{}
	for i := range pipelineRuns {
		pipelineRun := pipelineRuns[i]
		pipelineRunIdentity := pipelineRunIdentity{
			id: pipelineRun.Annotations[v1alpha3.JenkinsPipelineRunIDKey],
		}
		if pipelineRun.Spec.SCM != nil {
			pipelineRunIdentity.refName = pipelineRun.Spec.SCM.RefName
		}
		finder[pipelineRunIdentity] = &pipelineRun
	}
	return finder
}
