// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

const (
	GroupName                  = "gitops.kubesphere.io"
	HealthStatusLabelKey       = GroupName + "/health-status"
	SyncStatusLabelKey         = GroupName + "/sync-status"
	ArgoCDLocationLabelKey     = GroupName + "/argocd-location"
	ArgoCDAppNameLabelKey      = GroupName + "/argocd-application"
	ArgoCDAppControlByLabelKey = GroupName + "/argocd-application-control-by"
	AppNamespaceLabelKey       = GroupName + "/application-namespace"
	AppNameLabelKey            = GroupName + "/application-name"
)

const (
	// AnnoKeyImages is the key for the image list
	AnnoKeyImages = GroupName + "/images"
)

// ApplicationFinalizerName is the name of PipelineRun finalizer
const ApplicationFinalizerName = "application." + GroupName

// ArgoCDResourcesFinalizer is the name of Argo CD resource finalizer
const ArgoCDResourcesFinalizer = "resources-finalizer.argocd.argoproj.io"

// ArtifactRepoLabelKey is the annotation key that indicate whether
// the Repository is a ArtifactRepository
const ArtifactRepoLabelKey = GroupName + "/is-artifact-repository"

// Engine is the backend GitOps Solutions type
type Engine string

const (
	// ArgoCD is a declarative, GitOps continuous delivery tool for Kubernetes.
	// https://argo-cd.readthedocs.io/en/stable/
	ArgoCD Engine = "argocd"
	// FluxCD is a set of continuous and progressive delivery solutions
	// for Kubernetes that are open and extensible.
	// https://fluxcd.io/
	FluxCD Engine = "fluxcd"
)

const (
	// SaveTemplateLabelKey control whether to save a HelmTemplate
	SaveTemplateLabelKey = GroupName + "/save-helm-template"

	// HelmTemplateName represent the user interface HelmTemplate name
	HelmTemplateName = GroupName + "/helm-template-name"
)
