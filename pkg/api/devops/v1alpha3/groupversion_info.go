/*
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

// Package v1alpha3 contains API Schema definitions for the devops.kubesphere.io v1alpha3 API group
// +kubebuilder:object:generate=true
// +groupName=devops.kubesphere.io
package v1alpha3

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	// GroupName is the group name use in this package
	GroupName = "devops.kubesphere.io"
	// JenkinsPipelineRunIDAnnoKey is annotation key of Jenkins PipelineRun ID.
	JenkinsPipelineRunIDAnnoKey = GroupName + "/jenkins-pipelinerun-id"
	// JenkinsPipelineRunStatusAnnoKey is annotation key of status of Jenkins PipelineRun.
	JenkinsPipelineRunStatusAnnoKey = GroupName + "/jenkins-pipelinerun-status"
	// JenkinsPipelineRunStagesStatusAnnoKey is annotation key of Jenkins stages' status of Jenkins PipelineRun.
	JenkinsPipelineRunStagesStatusAnnoKey = GroupName + "/jenkins-pipelinerun-stages-status"
	// PipelineRunOrphanLabelKey is label key of orphan Jenkins PipelineRun which type of value is bool.
	PipelineRunOrphanLabelKey = GroupName + "/jenkins-pipelinerun-orphan"
	// PipelineNameLabelKey is label key of Pipeline name.
	PipelineNameLabelKey = GroupName + "/pipeline"
	// SCMRefNameLabelKey is label key of SCM reference name.
	SCMRefNameLabelKey = GroupName + "/scm-ref-name"
	// PipelineRunCreatorAnnoKey is annotation key of PipelineRun's creator
	PipelineRunCreatorAnnoKey = GroupName + "/creator"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha3"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}
