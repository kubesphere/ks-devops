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

package fluxcd

const controllerGroupName = "fluxcd"

// AppType stand for the GitOps Application Type
type AppType string

const (
	// HelmRelease stand for FluxCD HelmRelease Application Type
	HelmRelease AppType = "HelmRelease"
	// Kustomization stand for FluxCD Kustomization Application Type
	Kustomization AppType = "Kustomization"
)

// DefaultKubeConfigKey is the Default key for kubeconfig
// https://fluxcd.io/docs/components/helm/helmreleases/#remote-clusters--cluster-api
const DefaultKubeConfigKey = "value"
