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

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FluxApplication struct {
	Spec FluxApplicationSpec `json:"spec,omitempty"`
}

type FluxApplicationSpec struct {
	Source *FluxApplicationSource `json:"source,omitempty"`
	Config *FluxApplicationConfig `json:"config"`
}

type FluxApplicationSource struct {
	SourceRef CrossNamespaceObjectReference `json:"sourceRef"`
}

// CrossNamespaceObjectReference contains enough information to let you locate
// the typed referenced object at cluster level.
type CrossNamespaceObjectReference struct {
	// APIVersion of the referent.
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referent.
	// Enum=HelmRepository;GitRepository;Bucket
	Kind string `json:"kind,omitempty"`

	// Name of the referent.
	Name string `json:"name"`

	// Namespace of the referent.
	Namespace string `json:"namespace,omitempty"`
}

type FluxApplicationDestination struct {
	// The KubeConfig for reconciling the Kustomization on a remote cluster.
	// When used in combination with KustomizationSpec.ServiceAccountName,
	// forces the controller to act on behalf of that Service Account at the
	// target cluster.
	// If the --default-service-account flag is set, its value will be used as
	// a controller level fallback for when KustomizationSpec.ServiceAccountName
	// is empty.
	KubeConfig *KubeConfig `json:"kubeConfig,omitempty"`
	// TargetNamespace to target when performing operations for the HelmRelease.
	// Defaults to the namespace of the HelmRelease.
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

// KubeConfig references a Kubernetes secret that contains a kubeconfig file.
type KubeConfig struct {
	// SecretRef holds the name of a secret that contains a key with
	// the kubeconfig file as the value. If no key is set, the key will default
	// to 'value'. The secret must be in the same namespace as
	// the Kustomization.
	// It is recommended that the kubeconfig is self-contained, and the secret
	// is regularly updated if credentials such as a cloud-access-token expire.
	// Cloud specific `cmd-path` auth helpers will not function without adding
	// binaries and credentials to the Pod that is responsible for reconciling
	// the Kustomization.
	SecretRef SecretKeyReference `json:"secretRef,omitempty"`
}

// SecretKeyReference contains enough information to locate the referenced Kubernetes Secret object in the same
// namespace. Optionally a key can be specified.
// Use this type instead of core/v1 SecretKeySelector when the Key is optional and the Optional field is not
// applicable.
type SecretKeyReference struct {
	// Name of the Secret.
	Name string `json:"name"`

	// Key in the Secret, when not specified an implementation-specific default key is used.
	Key string `json:"key,omitempty"`
}

type FluxApplicationConfig struct {
	// HelmRelease for FluxCD HelmRelease
	HelmRelease *HelmReleaseSpec `json:"helmRelease,omitempty"`

	// Kustomization for FluxCD Kustomization
	Kustomization []*KustomizationSpec `json:"kustomization,omitempty"`
}

// HelmReleaseSpec defines the desired state of a Helm release.
type HelmReleaseSpec struct {
	// Chart defines the template of the v1beta2.HelmChart that should be created
	// for this HelmRelease.
	Chart *HelmChartTemplateSpec `json:"chart,omitempty"`

	// Template ref a HelmTemplate that has been saved before
	Template string `json:"template,omitempty"`

	// HelmReleaseConfig stand for multi-clusters and multi-targetNamespace config
	Deploy []*Deploy `json:"deploy"`
}

// HelmChartTemplateSpec is just a simple copy of helm.toolkit.fluxcd.io/HelmRelease's HelmChartTemplateSpec fields
// exclude sourceRef field because it's in the fluxApp.spec.source field
type HelmChartTemplateSpec struct {
	// The name or path the Helm chart is available at in the SourceRef.
	Chart string `json:"chart"`

	// Version semver expression, ignored for charts from v1beta2.GitRepository and
	// v1beta2.Bucket sources. Defaults to latest when omitted.
	// +kubebuilder:default:=*
	Version string `json:"version,omitempty"`

	// Interval at which to check the v1beta2.Source for updates. Defaults to
	// 'HelmReleaseSpec.Interval'.
	Interval *metav1.Duration `json:"interval,omitempty"`

	// Determines what enables the creation of a new artifact. Valid values are
	// ('ChartVersion', 'Revision').
	// See the documentation of the values for an explanation on their behavior.
	// Defaults to ChartVersion when omitted.
	ReconcileStrategy string `json:"reconcileStrategy,omitempty"`

	// Alternative list of values files to use as the chart values (values.yaml
	// is not included by default), expected to be a relative path in the SourceRef.
	// Values files are merged in the order of this list with the last file overriding
	// the first. Ignored when omitted.
	ValuesFiles []string `json:"valuesFiles,omitempty"`
}

type Deploy struct {
	// Destination stand for the destination of the helmrelease
	Destination FluxApplicationDestination `json:"destination"`

	// The interval at which to reconcile the Kustomization.
	Interval metav1.Duration `json:"interval"`
	// Suspend tells the controller to suspend reconciliation for this HelmRelease,
	// it does not apply to already started reconciliations. Defaults to false.
	Suspend bool `json:"suspend,omitempty"`

	// Timeout is the time to wait for any individual Kubernetes operation (like Jobs
	// for hooks) during the performance of a Helm action. Defaults to '5m0s'.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// DependsOn may contain a meta.NamespacedObjectReference slice with
	// references to HelmRelease resources that must be ready before this HelmRelease
	// can be reconciled.
	DependsOn []NamespacedObjectReference `json:"dependsOn,omitempty"`

	// The name of the Kubernetes service account to impersonate
	// when reconciling this HelmRelease.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// ReleaseName used for the Helm release. Defaults to a composition of
	// '[TargetNamespace-]Name'.
	ReleaseName string `json:"releaseName,omitempty"`

	// StorageNamespace used for the Helm storage.
	// Defaults to the namespace of the HelmRelease.
	StorageNamespace string `json:"storageNamespace,omitempty"`

	// MaxHistory is the number of revisions saved by Helm for this HelmRelease.
	// Use '0' for an unlimited number of revisions; defaults to '10'.
	MaxHistory *int `json:"maxHistory,omitempty"`

	// Install holds the configuration for Helm install actions for this HelmRelease.
	Install *Install `json:"install,omitempty"`

	// Upgrade holds the configuration for Helm upgrade actions for this HelmRelease.
	Upgrade *Upgrade `json:"upgrade,omitempty"`

	// Test holds the configuration for Helm test actions for this HelmRelease.
	Test *Test `json:"test,omitempty"`

	// Rollback holds the configuration for Helm rollback actions for this HelmRelease.
	Rollback *Rollback `json:"rollback,omitempty"`

	// Uninstall holds the configuration for Helm uninstall actions for this HelmRelease.
	Uninstall *Uninstall `json:"uninstall,omitempty"`

	// ValuesFrom holds references to resources containing Helm values for this HelmRelease,
	// and information about how they should be merged.
	ValuesFrom []ValuesReference `json:"valuesFrom,omitempty"`

	// Values holds the values for this Helm release.
	Values *apiextensionsv1.JSON `json:"values,omitempty"`

	// PostRenderers holds an array of Helm PostRenderers, which will be applied in order
	// of their definition.
	PostRenderers []PostRenderer `json:"postRenderers,omitempty"`
}

type Install struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm install action. Defaults to
	// 'HelmReleaseSpec.Timeout'.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Remediation holds the remediation configuration for when the Helm install
	// action for the HelmRelease fails. The default is to not perform any action.
	Remediation *InstallRemediation `json:"remediation,omitempty"`

	// DisableWait disables the waiting for resources to be ready after a Helm
	// install has been performed.
	DisableWait bool `json:"disableWait,omitempty"`

	// DisableWaitForJobs disables waiting for jobs to complete after a Helm
	// install has been performed.
	DisableWaitForJobs bool `json:"disableWaitForJobs,omitempty"`

	// DisableHooks prevents hooks from running during the Helm install action.
	DisableHooks bool `json:"disableHooks,omitempty"`

	// DisableOpenAPIValidation prevents the Helm install action from validating
	// rendered templates against the Kubernetes OpenAPI Schema.
	DisableOpenAPIValidation bool `json:"disableOpenAPIValidation,omitempty"`

	// Replace tells the Helm install action to re-use the 'ReleaseName', but only
	// if that name is a deleted release which remains in the history.
	Replace bool `json:"replace,omitempty"`

	// SkipCRDs tells the Helm install action to not install any CRDs. By default,
	// CRDs are installed if not already present.
	SkipCRDs bool `json:"skipCRDs,omitempty"`

	// CRDs upgrade CRDs from the Helm Chart's crds directory according
	// to the CRD upgrade policy provided here. Valid values are `Skip`,
	// `Create` or `CreateReplace`. Default is `Create` and if omitted
	// CRDs are installed but not updated.
	//
	// Skip: do neither install nor replace (update) any CRDs.
	//
	// Create: new CRDs are created, existing CRDs are neither updated nor deleted.
	//
	// CreateReplace: new CRDs are created, existing CRDs are updated (replaced)
	// but not deleted.
	//
	// By default, CRDs are applied (installed) during Helm install action.
	// With this option users can opt-in to CRD replace existing CRDs on Helm
	// install actions, which is not (yet) natively supported by Helm.
	// https://helm.sh/docs/chart_best_practices/custom_resource_definitions.
	CRDs CRDsPolicy `json:"crds,omitempty"`

	// CreateNamespace tells the Helm install action to create the
	// HelmReleaseSpec.TargetNamespace if it does not exist yet.
	// On uninstall, the namespace will not be garbage collected.
	CreateNamespace bool `json:"createNamespace,omitempty"`
}

type InstallRemediation struct {
	// Retries is the number of retries that should be attempted on failures before
	// bailing. Remediation, using an uninstall, is performed between each attempt.
	// Defaults to '0', a negative integer equals to unlimited retries.
	Retries int `json:"retries,omitempty"`

	// IgnoreTestFailures tells the controller to skip remediation when the Helm
	// tests are run after an install action but fail. Defaults to
	// 'Test.IgnoreFailures'.
	IgnoreTestFailures *bool `json:"ignoreTestFailures,omitempty"`

	// RemediateLastFailure tells the controller to remediate the last failure, when
	// no retries remain. Defaults to 'false'.
	RemediateLastFailure *bool `json:"remediateLastFailure,omitempty"`
}

type Upgrade struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm upgrade action. Defaults to
	// 'HelmReleaseSpec.Timeout'.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Remediation holds the remediation configuration for when the Helm upgrade
	// action for the HelmRelease fails. The default is to not perform any action.
	Remediation *UpgradeRemediation `json:"remediation,omitempty"`

	// DisableWait disables the waiting for resources to be ready after a Helm
	// upgrade has been performed.
	DisableWait bool `json:"disableWait,omitempty"`

	// DisableWaitForJobs disables waiting for jobs to complete after a Helm
	// upgrade has been performed.
	DisableWaitForJobs bool `json:"disableWaitForJobs,omitempty"`

	// DisableHooks prevents hooks from running during the Helm upgrade action.
	DisableHooks bool `json:"disableHooks,omitempty"`

	// DisableOpenAPIValidation prevents the Helm upgrade action from validating
	// rendered templates against the Kubernetes OpenAPI Schema.
	DisableOpenAPIValidation bool `json:"disableOpenAPIValidation,omitempty"`

	// Force forces resource updates through a replacement strategy.
	Force bool `json:"force,omitempty"`

	// PreserveValues will make Helm reuse the last release's values and merge in
	// overrides from 'Values'. Setting this flag makes the HelmRelease
	// non-declarative.
	PreserveValues bool `json:"preserveValues,omitempty"`

	// CleanupOnFail allows deletion of new resources created during the Helm
	// upgrade action when it fails.
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`

	// CRDs upgrade CRDs from the Helm Chart's crds directory according
	// to the CRD upgrade policy provided here. Valid values are `Skip`,
	// `Create` or `CreateReplace`. Default is `Skip` and if omitted
	// CRDs are neither installed nor upgraded.
	//
	// Skip: do neither install nor replace (update) any CRDs.
	//
	// Create: new CRDs are created, existing CRDs are neither updated nor deleted.
	//
	// CreateReplace: new CRDs are created, existing CRDs are updated (replaced)
	// but not deleted.
	//
	// By default, CRDs are not applied during Helm upgrade action. With this
	// option users can opt-in to CRD upgrade, which is not (yet) natively supported by Helm.
	// https://helm.sh/docs/chart_best_practices/custom_resource_definitions.
	CRDs CRDsPolicy `json:"crds,omitempty"`
}

type UpgradeRemediation struct {
	// Retries is the number of retries that should be attempted on failures before
	// bailing. Remediation, using 'Strategy', is performed between each attempt.
	// Defaults to '0', a negative integer equals to unlimited retries.
	Retries int `json:"retries,omitempty"`

	// IgnoreTestFailures tells the controller to skip remediation when the Helm
	// tests are run after an upgrade action but fail.
	// Defaults to 'Test.IgnoreFailures'.
	IgnoreTestFailures *bool `json:"ignoreTestFailures,omitempty"`

	// RemediateLastFailure tells the controller to remediate the last failure, when
	// no retries remain. Defaults to 'false' unless 'Retries' is greater than 0.
	RemediateLastFailure *bool `json:"remediateLastFailure,omitempty"`

	// Strategy to use for failure remediation. Defaults to 'rollback'.
	Strategy *RemediationStrategy `json:"strategy,omitempty"`
}

// CRDsPolicy defines the install/upgrade approach to use for CRDs when
// installing or upgrading a HelmRelease.
type CRDsPolicy string

// RemediationStrategy returns the strategy to use to remediate a failed install
// or upgrade.
type RemediationStrategy string

type Test struct {
	// Enable enables Helm test actions for this HelmRelease after an Helm install
	// or upgrade action has been performed.
	Enable bool `json:"enable,omitempty"`

	// Timeout is the time to wait for any individual Kubernetes operation during
	// the performance of a Helm test action. Defaults to 'HelmReleaseSpec.Timeout'.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// IgnoreFailures tells the controller to skip remediation when the Helm tests
	// are run but fail. Can be overwritten for tests run after install or upgrade
	// actions in 'Install.IgnoreTestFailures' and 'Upgrade.IgnoreTestFailures'.
	IgnoreFailures bool `json:"ignoreFailures,omitempty"`
}

type Rollback struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm rollback action. Defaults to
	// 'HelmReleaseSpec.Timeout'.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// DisableWait disables the waiting for resources to be ready after a Helm
	// rollback has been performed.
	DisableWait bool `json:"disableWait,omitempty"`

	// DisableWaitForJobs disables waiting for jobs to complete after a Helm
	// rollback has been performed.
	DisableWaitForJobs bool `json:"disableWaitForJobs,omitempty"`

	// DisableHooks prevents hooks from running during the Helm rollback action.
	DisableHooks bool `json:"disableHooks,omitempty"`

	// Recreate performs pod restarts for the resource if applicable.
	Recreate bool `json:"recreate,omitempty"`

	// Force forces resource updates through a replacement strategy.
	Force bool `json:"force,omitempty"`

	// CleanupOnFail allows deletion of new resources created during the Helm
	// rollback action when it fails.
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`
}

type Uninstall struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm uninstall action. Defaults
	// to 'HelmReleaseSpec.Timeout'.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// DisableHooks prevents hooks from running during the Helm rollback action.
	DisableHooks bool `json:"disableHooks,omitempty"`

	// KeepHistory tells Helm to remove all associated resources and mark the
	// release as deleted, but retain the release history.
	KeepHistory bool `json:"keepHistory,omitempty"`

	// DisableWait disables waiting for all the resources to be deleted after
	// a Helm uninstall is performed.
	DisableWait bool `json:"disableWait,omitempty"`
}

type ValuesReference struct {
	// Kind of the values referent, valid values are ('Secret', 'ConfigMap').
	Kind string `json:"kind"`

	// Name of the values referent. Should reside in the same namespace as the
	// referring resource.
	Name string `json:"name"`

	// ValuesKey is the data key where the values.yaml or a specific value can be
	// found at. Defaults to 'values.yaml'.
	ValuesKey string `json:"valuesKey,omitempty"`

	// TargetPath is the YAML dot notation path the value should be merged at. When
	// set, the ValuesKey is expected to be a single flat value. Defaults to 'None',
	// which results in the values getting merged at the root.
	TargetPath string `json:"targetPath,omitempty"`

	// Optional marks this ValuesReference as optional. When set, a not found error
	// for the values reference is ignored, but any ValuesKey, TargetPath or
	// transient error will still result in a reconciliation failure.
	Optional bool `json:"optional,omitempty"`
}

type PostRenderer struct {
	// Kustomization to apply as PostRenderer.
	Kustomize *Kustomize `json:"kustomize,omitempty"`
}

type Kustomize struct {
	// Strategic merge and JSON patches, defined as inline YAML objects,
	// capable of targeting objects based on kind, label and annotation selectors.
	Patches []Patch `json:"patches,omitempty"`

	// Strategic merge patches, defined as inline YAML objects.
	PatchesStrategicMerge []apiextensionsv1.JSON `json:"patchesStrategicMerge,omitempty"`

	// Images is a list of (image name, new name, new tag or digest)
	// for changing image names, tags or digests. This can also be achieved with a
	// patch, but this operator is simpler to specify.
	Images []Image `json:"images,omitempty" yaml:"images,omitempty"`
}

// KustomizationSpec defines the configuration to calculate the desired state from a Source using Kustomize.
type KustomizationSpec struct {
	// Destination stand for the destination of the kustomization
	Destination FluxApplicationDestination `json:"destination"`

	// DependsOn may contain a meta.NamespacedObjectReference slice
	// with references to Kustomization resources that must be ready before this
	// Kustomization can be reconciled.
	DependsOn []NamespacedObjectReference `json:"dependsOn,omitempty"`

	// Decrypt Kubernetes secrets before applying them on the cluster.
	Decryption *Decryption `json:"decryption,omitempty"`

	// The interval at which to reconcile the Kustomization.
	Interval metav1.Duration `json:"interval"`

	// The interval at which to retry a previously failed reconciliation.
	// When not specified, the controller uses the KustomizationSpec.Interval
	// value to retry failures.
	RetryInterval *metav1.Duration `json:"retryInterval,omitempty"`

	// Path to the directory containing the kustomization.yaml file, or the
	// set of plain YAMLs a kustomization.yaml should be generated for.
	// Defaults to 'None', which translates to the root path of the SourceRef.
	Path string `json:"path,omitempty"`

	// PostBuild describes which actions to perform on the YAML manifest
	// generated by building the kustomize overlay.
	PostBuild *PostBuild `json:"postBuild,omitempty"`

	// Prune enables garbage collection.
	Prune bool `json:"prune"`

	// A list of resources to be included in the health assessment.
	HealthChecks []NamespacedObjectKindReference `json:"healthChecks,omitempty"`

	// Strategic merge and JSON patches, defined as inline YAML objects,
	// capable of targeting objects based on kind, label and annotation selectors.
	Patches []Patch `json:"patches,omitempty"`

	// Images is a list of (image name, new name, new tag or digest)
	// for changing image names, tags or digests. This can also be achieved with a
	// patch, but this operator is simpler to specify.
	Images []Image `json:"images,omitempty"`

	// The name of the Kubernetes service account to impersonate
	// when reconciling this Kustomization.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// This flag tells the controller to suspend subsequent kustomize executions,
	// it does not apply to already started executions. Defaults to false.
	Suspend bool `json:"suspend,omitempty"`

	// Timeout for validation, apply and health checking operations.
	// Defaults to 'Interval' duration.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Force instructs the controller to recreate resources
	// when patching fails due to an immutable field change.
	Force bool `json:"force,omitempty"`

	// Wait instructs the controller to check the health of all the reconciled resources.
	// When enabled, the HealthChecks are ignored. Defaults to false.
	Wait bool `json:"wait,omitempty"`
}

// PostBuild describes which actions to perform on the YAML manifest
// generated by building the kustomize overlay.
type PostBuild struct {
	// Substitute holds a map of key/value pairs.
	// The variables defined in your YAML manifests
	// that match any of the keys defined in the map
	// will be substituted with the set value.
	// Includes support for bash string replacement functions
	// e.g. ${var:=default}, ${var:position} and ${var/substring/replacement}.
	Substitute map[string]string `json:"substitute,omitempty"`

	// SubstituteFrom holds references to ConfigMaps and Secrets containing
	// the variables and their values to be substituted in the YAML manifests.
	// The ConfigMap and the Secret data keys represent the var names and they
	// must match the vars declared in the manifests for the substitution to happen.
	SubstituteFrom []SubstituteReference `json:"substituteFrom,omitempty"`
}

// SubstituteReference contains a reference to a resource containing
// the variables name and value.
type SubstituteReference struct {
	// Kind of the values referent, valid values are ('Secret', 'ConfigMap').
	Kind string `json:"kind"`

	// Name of the values referent. Should reside in the same namespace as the
	// referring resource.
	Name string `json:"name"`

	// Optional indicates whether the referenced resource must exist, or whether to
	// tolerate its absence. If true and the referenced resource is absent, proceed
	// as if the resource was present but empty, without any variables defined.
	Optional bool `json:"optional,omitempty"`
}

// Decryption defines how decryption is handled for Kubernetes manifests.
type Decryption struct {
	// Provider is the name of the decryption engine.
	Provider string `json:"provider"`

	// The secret name containing the private OpenPGP keys used for decryption.
	SecretRef *LocalObjectReference `json:"secretRef,omitempty"`
}

// ResourceInventory contains a list of Kubernetes resource object references that have been applied by a Kustomization.
type ResourceInventory struct {
	// Entries of Kubernetes resource object references.
	Entries []ResourceRef `json:"entries"`
}

// ResourceRef contains the information necessary to locate a resource within a cluster.
type ResourceRef struct {
	// ID is the string representation of the Kubernetes resource object's metadata,
	// in the format '<namespace>_<name>_<group>_<kind>'.
	ID string `json:"id"`

	// Version is the API version of the Kubernetes resource object's kind.
	Version string `json:"v"`
}

// LocalObjectReference contains enough information to locate the referenced Kubernetes resource object.
type LocalObjectReference struct {
	// Name of the referent.
	// +required
	Name string `json:"name"`
}

// NamespacedObjectReference contains enough information to locate the referenced Kubernetes resource object in any
// namespace.
type NamespacedObjectReference struct {
	// Name of the referent.
	// +required
	Name string `json:"name"`

	// Namespace of the referent, when not specified it acts as LocalObjectReference.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// NamespacedObjectKindReference contains enough information to locate the typed referenced Kubernetes resource object
// in any namespace.
type NamespacedObjectKindReference struct {
	// API version of the referent, if not specified the Kubernetes preferred version will be used.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referent.
	// +required
	Kind string `json:"kind"`

	// Name of the referent.
	// +required
	Name string `json:"name"`

	// Namespace of the referent, when not specified it acts as LocalObjectReference.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// Image contains an image name, a new name, a new tag or digest, which will replace the original name and tag.
type Image struct {
	// Name is a tag-less image name.
	// +required
	Name string `json:"name"`

	// NewName is the value used to replace the original name.
	// +optional
	NewName string `json:"newName,omitempty"`

	// NewTag is the value used to replace the original tag.
	// +optional
	NewTag string `json:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag.
	// If digest is present NewTag value is ignored.
	// +optional
	Digest string `json:"digest,omitempty"`
}

// Patch contains an inline StrategicMerge or JSON6902 patch, and the target the patch should
// be applied to.
type Patch struct {
	// Patch contains an inline StrategicMerge patch or an inline JSON6902 patch with
	// an array of operation objects.
	// +required
	Patch string `json:"patch,omitempty"`

	// Target points to the resources that the patch document should be applied to.
	// +optional
	Target Selector `json:"target,omitempty"`
}

// Selector specifies a set of resources. Any resource that matches intersection of all conditions is included in this
// set.
type Selector struct {
	// Group is the API group to select resources from.
	// Together with Version and Kind it is capable of unambiguously identifying and/or selecting resources.
	// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/api-group.md
	// +optional
	Group string `json:"group,omitempty"`

	// Version of the API Group to select resources from.
	// Together with Group and Kind it is capable of unambiguously identifying and/or selecting resources.
	// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/api-group.md
	// +optional
	Version string `json:"version,omitempty"`

	// Kind of the API Group to select resources from.
	// Together with Group and Version it is capable of unambiguously
	// identifying and/or selecting resources.
	// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/api-group.md
	// +optional
	Kind string `json:"kind,omitempty"`

	// Namespace to select resources from.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name to match resources with.
	// +optional
	Name string `json:"name,omitempty"`

	// AnnotationSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource annotations.
	// +optional
	AnnotationSelector string `json:"annotationSelector,omitempty"`

	// LabelSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource labels.
	// +optional
	LabelSelector string `json:"labelSelector,omitempty"`
}

// ApplicationSpec is the specification of the Application
type ApplicationSpec struct {
	Kind    Engine           `json:"kind"`
	ArgoApp *ArgoApplication `json:"argoApp,omitempty"`
	FluxApp *FluxApplication `json:"fluxApp,omitempty"`
}

// ArgoApplication is a definition of Argo Application resource.
// Those fields simply are copied from the argo-cd project
type ArgoApplication struct {
	Spec      ArgoApplicationSpec `json:"spec,omitempty"`
	Operation *Operation          `json:"operation,omitempty"`
}

// OperationInitiator contains information about the initiator of an operation
type OperationInitiator struct {
	// Username contains the name of a user who started operation
	Username string `json:"username,omitempty"`
	// Automated is set to true if operation was initiated automatically by the application controller.
	Automated bool `json:"automated,omitempty"`
}

// SyncOperationResource contains resources to sync.
type SyncOperationResource struct {
	Group     string `json:"group,omitempty"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// SyncOperation contains details about a sync operation.
type SyncOperation struct {
	// Revision is the revision (Git) or chart version (Helm) which to sync the application to
	// If omitted, will use the revision specified in app spec.
	Revision string `json:"revision,omitempty"`
	// Prune specifies to delete resources from the cluster that are no longer tracked in git
	Prune bool `json:"prune,omitempty"`
	// DryRun specifies to perform a `kubectl apply --dry-run` without actually performing the sync
	DryRun bool `json:"dryRun,omitempty"`
	// SyncStrategy describes how to perform the sync
	SyncStrategy *SyncStrategy `json:"syncStrategy,omitempty"`
	// Resources describes which resources shall be part of the sync
	Resources []SyncOperationResource `json:"resources,omitempty"`
	// Source overrides the source definition set in the application.
	// This is typically set in a Rollback operation and is nil during a Sync operation
	Source *ApplicationSource `json:"source,omitempty"`
	// Manifests is an optional field that overrides sync source with a local directory for development
	Manifests []string `json:"manifests,omitempty"`
	// SyncOptions provide per-sync sync-options, e.g. Validate=false
	SyncOptions SyncOptions `json:"syncOptions,omitempty"`
}

// Operation contains information about a requested or running operation
type Operation struct {
	// Sync contains parameters for the operation
	Sync *SyncOperation `json:"sync,omitempty"`
	// InitiatedBy contains information about who initiated the operations
	InitiatedBy OperationInitiator `json:"initiatedBy,omitempty"`
	// Info is a list of informational items for this operation
	Info []*Info `json:"info,omitempty"`
	// Retry controls the strategy to apply if a sync fails
	Retry RetryStrategy `json:"retry,omitempty"`
}

// ArgoApplicationSpec represents an ArgoCD Application
type ArgoApplicationSpec struct {
	// Source is a reference to the location of the application's manifests or chart
	Source ApplicationSource `json:"source"`
	// Destination is a reference to the target Kubernetes server and namespace
	Destination ApplicationDestination `json:"destination"`
	// Project is a reference to the project this application belongs to.
	// The empty string means that application belongs to the 'default' project.
	Project string `json:"project"`
	// SyncPolicy controls when and how a sync will be performed
	SyncPolicy *SyncPolicy `json:"syncPolicy,omitempty"`
	// IgnoreDifferences is a list of resources and their fields which should be ignored during comparison
	IgnoreDifferences []ResourceIgnoreDifferences `json:"ignoreDifferences,omitempty"`
	// Info contains a list of information (URLs, email addresses, and plain text) that relates to the application
	Info []Info `json:"info,omitempty"`
	// RevisionHistoryLimit limits the number of items kept in the application's revision history, which is used for informational purposes as well as for rollbacks to previous versions.
	// This should only be changed in exceptional circumstances.
	// Setting to zero will store no history. This will reduce storage used.
	// Increasing will increase the space used to store the history, so we do not recommend increasing it.
	// Default is 10.
	RevisionHistoryLimit *int64 `json:"revisionHistoryLimit,omitempty"`
}

// Info represents a name and value
type Info struct {
	Name  string `json:"name" protobuf:"bytes,1,name=name"`
	Value string `json:"value" protobuf:"bytes,2,name=value"`
}

// ResourceIgnoreDifferences contains resource filter and list of json paths which should be ignored during comparison with live state.
type ResourceIgnoreDifferences struct {
	Group             string   `json:"group,omitempty"`
	Kind              string   `json:"kind"`
	Name              string   `json:"name,omitempty"`
	Namespace         string   `json:"namespace,omitempty"`
	JSONPointers      []string `json:"jsonPointers,omitempty"`
	JQPathExpressions []string `json:"jqPathExpressions,omitempty"`
	// ManagedFieldsManagers is a list of trusted managers. Fields mutated by those managers will take precedence over the
	// desired state defined in the SCM and won't be displayed in diffs
	ManagedFieldsManagers []string `json:"managedFieldsManagers,omitempty"`
}

// SyncPolicy controls when a sync will be performed in response to updates in git
type SyncPolicy struct {
	// Automated will keep an application synced to the target revision
	Automated *SyncPolicyAutomated `json:"automated,omitempty"`
	// Options allow you to specify whole app sync-options
	SyncOptions SyncOptions `json:"syncOptions,omitempty"`
	// Retry controls failed sync retry behavior
	Retry *RetryStrategy `json:"retry,omitempty"`
}

// RetryStrategy contains information about the strategy to apply when a sync failed
type RetryStrategy struct {
	// Limit is the maximum number of attempts for retrying a failed sync. If set to 0, no retries will be performed.
	Limit int64 `json:"limit,omitempty"`
	// Backoff controls how to backoff on subsequent retries of failed syncs
	Backoff *Backoff `json:"backoff,omitempty"`
}

// SyncOptions is type alias of []string
type SyncOptions []string

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

// ApplicationSource contains all required information about the source of an application
type ApplicationSource struct {
	// RepoURL is the URL to the repository (Git or Helm) that contains the application manifests
	RepoURL string `json:"repoURL"`
	// Path is a directory path within the Git repository, and is only valid for applications sourced from Git.
	Path string `json:"path,omitempty"`
	// TargetRevision defines the revision of the source to sync the application to.
	// In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD.
	// In case of Helm, this is a semver tag for the Chart's version.
	TargetRevision string `json:"targetRevision,omitempty"`
	// Helm holds helm specific options
	Helm *ApplicationSourceHelm `json:"helm,omitempty"`
	// Kustomize holds kustomize specific options
	Kustomize *ApplicationSourceKustomize `json:"kustomize,omitempty"`
	// Ksonnet holds ksonnet specific options
	Ksonnet *ApplicationSourceKsonnet `json:"ksonnet,omitempty"`
	// Directory holds path/directory specific options
	Directory *ApplicationSourceDirectory `json:"directory,omitempty"`
	// ConfigManagementPlugin holds config management plugin specific options
	Plugin *ApplicationSourcePlugin `json:"plugin,omitempty"`
	// Chart is a Helm chart name, and must be specified for applications sourced from a Helm repo.
	Chart string `json:"chart,omitempty"`
}

// ApplicationSourcePlugin holds options specific to config management plugins
type ApplicationSourcePlugin struct {
	Name string `json:"name,omitempty"`
	Env  `json:"env,omitempty"`
}

// Env is a list of environment variable entries
type Env []*EnvEntry

// EnvEntry represents an entry in the application's environment
type EnvEntry struct {
	// Name is the name of the variable, usually expressed in uppercase
	Name string `json:"name"`
	// Value is the value of the variable
	Value string `json:"value"`
}

// ApplicationSourceDirectory holds options for applications of type plain YAML or Jsonnet
type ApplicationSourceDirectory struct {
	// Recurse specifies whether to scan a directory recursively for manifests
	Recurse bool `json:"recurse,omitempty"`
	// Jsonnet holds options specific to Jsonnet
	Jsonnet ApplicationSourceJsonnet `json:"jsonnet,omitempty"`
	// Exclude contains a glob pattern to match paths against that should be explicitly excluded from being used during manifest generation
	Exclude string `json:"exclude,omitempty"`
	// Include contains a glob pattern to match paths against that should be explicitly included during manifest generation
	Include string `json:"include,omitempty"`
}

// ApplicationSourceJsonnet holds options specific to applications of type Jsonnet
type ApplicationSourceJsonnet struct {
	// ExtVars is a list of Jsonnet External Variables
	ExtVars []JsonnetVar `json:"extVars,omitempty"`
	// TLAS is a list of Jsonnet Top-level Arguments
	TLAs []JsonnetVar `json:"tlas,omitempty"`
	// Additional library search dirs
	Libs []string `json:"libs,omitempty"`
}

// JsonnetVar represents a variable to be passed to jsonnet during manifest generation
type JsonnetVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Code  bool   `json:"code,omitempty"`
}

// ApplicationSourceKsonnet holds ksonnet specific options
type ApplicationSourceKsonnet struct {
	// Environment is a ksonnet application environment name
	Environment string `json:"environment,omitempty"`
	// Parameters are a list of ksonnet component parameter override values
	Parameters []KsonnetParameter `json:"parameters,omitempty"`
}

// KsonnetParameter is a ksonnet component parameter
type KsonnetParameter struct {
	Component string `json:"component,omitempty"`
	Name      string `json:"name"`
	Value     string `json:"value"`
}

// ApplicationSourceKustomize holds options specific to an Application source specific to Kustomize
type ApplicationSourceKustomize struct {
	// NamePrefix is a prefix appended to resources for Kustomize apps
	NamePrefix string `json:"namePrefix,omitempty"`
	// NameSuffix is a suffix appended to resources for Kustomize apps
	NameSuffix string `json:"nameSuffix,omitempty"`
	// Images is a list of Kustomize image override specifications
	Images KustomizeImages `json:"images,omitempty"`
	// CommonLabels is a list of additional labels to add to rendered manifests
	CommonLabels map[string]string `json:"commonLabels,omitempty"`
	// Version controls which version of Kustomize to use for rendering manifests
	Version string `json:"version,omitempty"`
	// CommonAnnotations is a list of additional annotations to add to rendered manifests
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty"`
	// ForceCommonLabels specifies whether to force applying common labels to resources for Kustomize apps
	ForceCommonLabels bool `json:"forceCommonLabels,omitempty"`
	// ForceCommonAnnotations specifies whether to force applying common annotations to resources for Kustomize apps
	ForceCommonAnnotations bool `json:"forceCommonAnnotations,omitempty"`
}

// KustomizeImages is a list of Kustomize images
type KustomizeImages []KustomizeImage

// KustomizeImage represents a Kustomize image definition in the format [old_image_name=]<image_name>:<image_tag>
type KustomizeImage string

// ApplicationSourceHelm holds helm specific options
type ApplicationSourceHelm struct {
	// ValuesFiles is a list of Helm value files to use when generating a template
	ValueFiles []string `json:"valueFiles,omitempty"`
	// Parameters is a list of Helm parameters which are passed to the helm template command upon manifest generation
	Parameters []HelmParameter `json:"parameters,omitempty"`
	// ReleaseName is the Helm release name to use. If omitted it will use the application name
	ReleaseName string `json:"releaseName,omitempty"`
	// Values specifies Helm values to be passed to helm template, typically defined as a block
	Values string `json:"values,omitempty"`
	// FileParameters are file parameters to the helm template
	FileParameters []HelmFileParameter `json:"fileParameters,omitempty"`
	// Version is the Helm version to use for templating (either "2" or "3")
	Version string `json:"version,omitempty"`
	// PassCredentials pass credentials to all domains (Helm's --pass-credentials)
	PassCredentials bool `json:"passCredentials,omitempty"`
	// IgnoreMissingValueFiles prevents helm template from failing when valueFiles do not exist locally by not appending them to helm template --values
	IgnoreMissingValueFiles bool `json:"ignoreMissingValueFiles,omitempty"`
	// SkipCrds skips custom resource definition installation step (Helm's --skip-crds)
	SkipCrds bool `json:"skipCrds,omitempty"`
}

// HelmParameter is a parameter that's passed to helm template during manifest generation
type HelmParameter struct {
	// Name is the name of the Helm parameter
	Name string `json:"name,omitempty"`
	// Value is the value for the Helm parameter
	Value string `json:"value,omitempty"`
	// ForceString determines whether to tell Helm to interpret booleans and numbers as strings
	ForceString bool `json:"forceString,omitempty"`
}

// HelmFileParameter is a file parameter that's passed to helm template during manifest generation
type HelmFileParameter struct {
	// Name is the name of the Helm parameter
	Name string `json:"name,omitempty"`
	// Path is the path to the file containing the values for the Helm parameter
	Path string `json:"path,omitempty"`
}

// Backoff is the backoff strategy to use on subsequent retries for failing syncs
type Backoff struct {
	// Duration is the amount to back off. Default unit is seconds, but could also be a duration (e.g. "2m", "1h")
	Duration string `json:"duration,omitempty"`
	// Factor is a factor to multiply the base duration after each failed retry
	Factor *int64 `json:"factor,omitempty"`
	// MaxDuration is the maximum amount of time allowed for the backoff strategy
	MaxDuration string `json:"maxDuration,omitempty"`
}

// SyncPolicyAutomated controls the behavior of an automated sync
type SyncPolicyAutomated struct {
	// Prune specifies whether to delete resources from the cluster that are not found in the sources anymore as part of automated sync (default: false)
	Prune bool `json:"prune,omitempty"`
	// SelfHeal specifes whether to revert resources back to their desired state upon modification in the cluster (default: false)
	SelfHeal bool `json:"selfHeal,omitempty"`
	// AllowEmpty allows apps have zero live resources (default: false)
	AllowEmpty bool `json:"allowEmpty,omitempty"`
}

// SyncStrategy controls the manner in which a sync is performed
type SyncStrategy struct {
	// Apply will perform a `kubectl apply` to perform the sync.
	Apply *SyncStrategyApply `json:"apply,omitempty"`
	// Hook will submit any referenced resources to perform the sync. This is the default strategy
	Hook *SyncStrategyHook `json:"hook,omitempty"`
}

// SyncStrategyApply uses `kubectl apply` to perform the apply
type SyncStrategyApply struct {
	// Force indicates whether or not to supply the --force flag to `kubectl apply`.
	// The --force flag deletes and re-create the resource, when PATCH encounters conflict and has
	// retried for 5 times.
	Force bool `json:"force,omitempty"`
}

// SyncStrategyHook will perform a sync using hooks annotations.
// If no hook annotation is specified falls back to `kubectl apply`.
type SyncStrategyHook struct {
	// Embed SyncStrategyApply type to inherit any `apply` options
	SyncStrategyApply `json:",inline"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true

// Application represents an application the DevOps system
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// ApplicationStatus represents the status of the Application
type ApplicationStatus struct {
	Kind    Engine `json:"kind"`
	ArgoApp string `json:"argoApp,omitempty"`
	FluxApp string `json:"fluxApp,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationList represents a set of the applications
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
