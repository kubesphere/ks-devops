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
			id: pipelineRun.Annotations[v1alpha3.JenkinsPipelineRunIDAnnoKey],
		}
		if pipelineRun.Spec.SCM != nil {
			pipelineRunIdentity.refName = pipelineRun.Spec.SCM.RefName
		}
		finder[pipelineRunIdentity] = &pipelineRun
	}
	return finder
}
