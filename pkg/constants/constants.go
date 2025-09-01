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
	CreatorAnnotationKey           = "kubesphere.io/creator"
	WorkspaceLabelKey              = "kubesphere.io/workspace"
	DisplayNameAnnotationKey       = "kubesphere.io/alias-name"
	InsecureSkipTLSAnnotationKey   = "devops.kubesphere.io/insecure-skip-tls"
	TLSCertsNameAnnotationKey      = "devops.kubesphere.io/tls-certs"
	TLSCertsNameSpaceAnnotationKey = "devops.kubesphere.io/tls-certs-namespace"
	GitAuthorNameAnnotationKey     = "devops.kubesphere.io/git-author-name"
	GitAuthorEmailAnnotationKey    = "devops.kubesphere.io/git-author-email"
	DevOpsProjectLabelKey          = "kubesphere.io/devopsproject"

	TLSCertKey               = "ca.crt"
	AuthenticationTag        = "Authentication"
	DevOpsCredentialTag      = "DevOps Credential"
	DevOpsPipelineTag        = "DevOps Pipeline"
	DevOpsWebhookTag         = "DevOps Webhook"
	DevOpsJenkinsfileTag     = "DevOps Jenkinsfile"
	DevOpsScmTag             = "DevOps Scm"
	DevOpsJenkinsTag         = "DevOps Jenkins"
	DevOpsProjectTag         = "DevOps Project"
	DevOpsTemplateTag        = "DevOps Template"
	DevOpsStepTemplateTag    = "DevOps StepTemplate"
	DevOpsClusterTemplateTag = "DevOps ClusterTemplate"
	GitOpsTag                = "GitOps"

	DevOpsManagedKey      = "devops.kubesphere.io/managed"
	DevOpsSystemNamespace = "kubesphere-devops-system"
	DevOpsWorkerNamespace = "kubesphere-devops-worker"
)

var (
	AuthenticationTags        = []string{AuthenticationTag}
	DevOpsProjectTags         = []string{DevOpsProjectTag}
	DevOpsCredentialTags      = []string{DevOpsCredentialTag}
	DevOpsPipelineTags        = []string{DevOpsPipelineTag}
	DevOpsWebhookTags         = []string{DevOpsWebhookTag}
	DevOpsScmTags             = []string{DevOpsScmTag}
	DevOpsJenkinsTags         = []string{DevOpsJenkinsTag}
	DevOpsTemplateTags        = []string{DevOpsTemplateTag}
	DevOpsStepTemplateTags    = []string{DevOpsStepTemplateTag}
	DevOpsClusterTemplateTags = []string{DevOpsClusterTemplateTag}
	GitOpsTags                = []string{GitOpsTag}
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

const (
	FieldType = "type"
)
