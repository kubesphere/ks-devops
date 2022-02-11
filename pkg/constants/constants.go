/*
Copyright 2019 The KubeSphere Authors.

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

package constants

const (
	AuthenticationTag    = "Authentication"
	CreatorAnnotationKey = "kubesphere.io/creator"
	WorkspaceLabelKey    = "kubesphere.io/workspace"

	DevOpsProjectLabelKey = "kubesphere.io/devopsproject"
	DevOpsCredentialTag   = "DevOps Credential"
	DevOpsPipelineTag     = "DevOps Pipeline"
	DevOpsWebhookTag      = "DevOps Webhook"
	DevOpsJenkinsfileTag  = "DevOps Jenkinsfile"
	DevOpsScmTag          = "DevOps Scm"
	DevOpsJenkinsTag      = "Jenkins"
	DevOpsProjectTag      = "DevOps Project"
	DevOpsTemplateTag     = "DevOps Template"
)

// K8SToken is the context key of k8s token
var K8SToken = ContextKeyK8SToken("k8s.token")

// ContextKeyK8SToken represents a type alias for the context key
type ContextKeyK8SToken string

const (
	StatusActive     = "active"
	StatusDeleted    = "deleted"
	StatusDeleting   = "deleting"
	StatusFailed     = "failed"
	StatusPending    = "pending"
	StatusWorking    = "working"
	StatusSuccessful = "successful"
)
