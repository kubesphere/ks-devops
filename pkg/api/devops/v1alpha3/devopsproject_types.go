/*
Copyright 2020 The KubeSphere Authors.

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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ResourceKindDevOpsProject      = "DevOpsProject"
	ResourceSingularDevOpsProject  = "devopsproject"
	ResourcePluralDevOpsProject    = "devopsprojects"
	DevOpsProjectPrefix            = "devopsproject.devops.kubesphere.io/"
	DevOpsProjectFinalizerName     = "devopsproject.finalizers.kubesphere.io"
	DevOpeProjectSyncStatusAnnoKey = DevOpsProjectPrefix + "syncstatus"
	DevOpeProjectSyncTimeAnnoKey   = DevOpsProjectPrefix + "synctime"
)

// DevOpsProjectSpec defines the desired state of DevOpsProject
type DevOpsProjectSpec struct {
	Argo *Argo `json:"argo,omitempty"`
}

// Argo represents the Argo CD specification
type Argo struct {
	// SourceRepos contains list of repository URLs which can be used for deployment
	SourceRepos []string `json:"sourceRepos,omitempty"`
	// Destinations contains list of destinations available for deployment
	Destinations []ApplicationDestination `json:"destinations,omitempty"`
	// Description contains optional project description
	Description string `json:"description,omitempty"`
	// Roles are user defined RBAC roles associated with this project
	Roles []ProjectRole `json:"roles,omitempty" protobuf:"bytes,4,rep,name=roles"`
	// ClusterResourceWhitelist contains list of whitelisted cluster level resources
	ClusterResourceWhitelist []metav1.GroupKind `json:"clusterResourceWhitelist,omitempty"`
	// NamespaceResourceBlacklist contains list of blacklisted namespace level resources
	NamespaceResourceBlacklist []metav1.GroupKind `json:"namespaceResourceBlacklist,omitempty"`
	// OrphanedResources specifies if controller should monitor orphaned resources of apps in this project
	OrphanedResources *OrphanedResourcesMonitorSettings `json:"orphanedResources,omitempty"`
	// SyncWindows controls when syncs can be run for apps in this project
	SyncWindows SyncWindows `json:"syncWindows,omitempty"`
	// NamespaceResourceWhitelist contains list of whitelisted namespace level resources
	NamespaceResourceWhitelist []metav1.GroupKind `json:"namespaceResourceWhitelist,omitempty"`
	// SignatureKeys contains a list of PGP key IDs that commits in Git must be signed with in order to be allowed for sync
	SignatureKeys []SignatureKey `json:"signatureKeys,omitempty"`
	// ClusterResourceBlacklist contains list of blacklisted cluster level resources
	ClusterResourceBlacklist []metav1.GroupKind `json:"clusterResourceBlacklist,omitempty"`
}

// ApplicationDestination holds information about the application's destination
type ApplicationDestination struct {
	// Server specifies the URL of the target cluster and must be set to the Kubernetes control plane API
	Server string `json:"server,omitempty"`
	// Namespace specifies the target namespace for the application's resources.
	// The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace
	Namespace string `json:"namespace,omitempty"`
	// Name is an alternate way of specifying the target cluster by its symbolic name
	Name string `json:"name,omitempty"`
}

// ProjectRole represents a role that has access to a project
type ProjectRole struct {
	// Name is a name for this role
	Name string `json:"name"`
	// Description is a description of the role
	Description string `json:"description,omitempty"`
	// Policies Stores a list of casbin formatted strings that define access policies for the role in the project
	Policies []string `json:"policies,omitempty"`
	// JWTTokens are a list of generated JWT tokens bound to this role
	JWTTokens []JWTToken `json:"jwtTokens,omitempty"`
	// Groups are a list of OIDC group claims bound to this role
	Groups []string `json:"groups,omitempty"`
}

// JWTToken holds the issuedAt and expiresAt values of a token
type JWTToken struct {
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp,omitempty"`
	ID        string `json:"id,omitempty"`
}

// OrphanedResourcesMonitorSettings holds settings of orphaned resources monitoring
type OrphanedResourcesMonitorSettings struct {
	// Warn indicates if warning condition should be created for apps which have orphaned resources
	Warn *bool `json:"warn,omitempty"`
	// Ignore contains a list of resources that are to be excluded from orphaned resources monitoring
	Ignore []OrphanedResourceKey `json:"ignore,omitempty"`
}

// OrphanedResourceKey is a reference to a resource to be ignored from
type OrphanedResourceKey struct {
	Group string `json:"group,omitempty"`
	Kind  string `json:"kind,omitempty"`
	Name  string `json:"name,omitempty"`
}

// SyncWindows is a collection of sync windows in this project
type SyncWindows []*SyncWindow

// SyncWindow contains the kind, time, duration and attributes that are used to assign the syncWindows to apps
type SyncWindow struct {
	// Kind defines if the window allows or blocks syncs
	Kind string `json:"kind,omitempty"`
	// Schedule is the time the window will begin, specified in cron format
	Schedule string `json:"schedule,omitempty"`
	// Duration is the amount of time the sync window will be open
	Duration string `json:"duration,omitempty"`
	// Applications contains a list of applications that the window will apply to
	Applications []string `json:"applications,omitempty"`
	// Namespaces contains a list of namespaces that the window will apply to
	Namespaces []string `json:"namespaces,omitempty"`
	// Clusters contains a list of clusters that the window will apply to
	Clusters []string `json:"clusters,omitempty"`
	// ManualSync enables manual syncs when they would otherwise be blocked
	ManualSync bool `json:"manualSync,omitempty"`
	//TimeZone of the sync that will be applied to the schedule
	TimeZone string `json:"timeZone,omitempty"`
}

// SignatureKey is the specification of a key required to verify commit signatures with
type SignatureKey struct {
	// The ID of the key in hexadecimal notation
	KeyID string `json:"keyID"`
}

// DevOpsProjectStatus defines the observed state of DevOpsProject
type DevOpsProjectStatus struct {
	AdminNamespace string `json:"adminNamespace,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DevOpsProject is the Schema for the devopsprojects API
// +kubebuilder:resource:categories="devops",scope="Cluster"
// +k8s:openapi-gen=true
type DevOpsProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DevOpsProjectSpec   `json:"spec,omitempty"`
	Status DevOpsProjectStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DevOpsProjectList contains a list of DevOpsProject
type DevOpsProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevOpsProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevOpsProject{}, &DevOpsProjectList{})
}
