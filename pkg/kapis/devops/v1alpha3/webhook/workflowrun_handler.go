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

package webhook

import (
	"context"
	"fmt"
	"kubesphere.io/devops/pkg/event/workflowrun"
	"strings"

	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/pipelinerun"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type pipelineRunIdentifier struct {
	namespaceName string
	pipelineName  string
	scmRefName    string
	buildNumber   string
}

// String populates the identifier of the PipelineRun.
func (identifier *pipelineRunIdentifier) String() string {
	if identifier == nil {
		return ""
	}
	return v1alpha3.BuildPipelineRunIdentifier(identifier.pipelineName, identifier.scmRefName, identifier.buildNumber)
}

func convertParameters(workflowRunParameters []workflowrun.Parameter) []v1alpha3.Parameter {
	var parameters []v1alpha3.Parameter
	for i := range workflowRunParameters {
		workflowRunParameter := &workflowRunParameters[i]
		if workflowRunParameter.Name == "" {
			continue
		}
		parameters = append(parameters, v1alpha3.Parameter{
			Name:  workflowRunParameter.Name,
			Value: fmt.Sprint(workflowRunParameter.Value),
		})
	}
	return parameters
}

func extractPipelineRunIdentifier(workflowRunData *workflowrun.Data) *pipelineRunIdentifier {
	if workflowRunData == nil || workflowRunData.ParentFullName == "" {
		return nil
	}
	identifier := &pipelineRunIdentifier{
		buildNumber: workflowRunData.ID,
	}

	fullName := workflowRunData.ParentFullName
	names := strings.Split(fullName, "/")
	if workflowRunData.IsMultiBranch {
		if len(names) != 2 {
			// return if this is not a standard multi-branch Pipeline in ks-devops
			return nil
		}
		identifier.namespaceName = names[0]
		identifier.pipelineName = names[1]
		identifier.scmRefName = workflowRunData.ProjectName
	} else {
		if len(names) != 1 {
			// return if this is not a standard Pipeline in ks-devops
			return nil
		}
		identifier.namespaceName = workflowRunData.ParentFullName
		identifier.pipelineName = workflowRunData.ProjectName
	}
	return identifier
}

func (handler *Handler) handleWorkflowRunInitialize(workflowRunData *workflowrun.Data) error {
	identifier := extractPipelineRunIdentifier(workflowRunData)
	if identifier == nil {
		// we should skip this event if the Pipeline is not a standard Pipeline in ks-devops.
		return nil
	}

	// TODO Execute process below asynchronously

	pipelineRunList := &v1alpha3.PipelineRunList{}
	if err := handler.List(context.Background(), pipelineRunList,
		client.InNamespace(identifier.namespaceName),
		client.MatchingFields{v1alpha3.PipelineRunIdentifierIndexerName: identifier.String()}); err != nil {
		return err
	}

	if len(pipelineRunList.Items) == 0 {
		parameters, err := workflowRunData.Actions.GetParameters()
		if err != nil {
			return err
		}

		pipelineRun, err := handler.createPipelineRun(identifier, convertParameters(parameters))
		if err != nil {
			return err
		}
		klog.Infof("Created a PipelineRun: %s/%s", pipelineRun.Namespace, pipelineRun.Name)
		return nil
	}
	return nil
}

func (handler *Handler) createPipelineRun(identifier *pipelineRunIdentifier, parameters []v1alpha3.Parameter) (*v1alpha3.PipelineRun, error) {
	pipeline := &v1alpha3.Pipeline{}
	if err := handler.Get(context.Background(), client.ObjectKey{Namespace: identifier.namespaceName, Name: identifier.pipelineName}, pipeline); err != nil {
		return nil, err
	}
	scm, err := pipelinerun.CreateScm(&pipeline.Spec, identifier.scmRefName)
	if err != nil {
		return nil, err
	}

	pipelineRun := pipelinerun.CreateBarePipelineRun(pipeline, parameters, scm)

	// Set the RunID manually
	pipelineRun.GetAnnotations()[v1alpha3.JenkinsPipelineRunIDAnnoKey] = identifier.buildNumber
	if err := handler.Create(context.Background(), pipelineRun); err != nil {
		return nil, err
	}
	return pipelineRun, nil
}
