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

package config

import (
	"github.com/spf13/pflag"
	v1alpha1 "kubesphere.io/devops/pkg/api/gitops/v1alpha1"
)

// ArgoCDOption as the ArgoCD integration configuration
type ArgoCDOption struct {
	Enabled   bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" description:"enabled FluxCD"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty" description:"Which namespace the ArgoCD located"`
}

// AddFlags adds the flags which related to argocd
func (o *ArgoCDOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "argocd-enabled", false, "Enable ArgoCD APIs")
	// see also https://argo-cd.readthedocs.io/en/stable/getting_started/
	fs.StringVarP(&o.Namespace, "argocd-namespace", o.Namespace, "argocd", "Which namespace the ArgoCD located")
}

// FluxCDOption as the FluxCD integration configuration
type FluxCDOption struct {
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty" description:"enabled FluxCD"`
}

// AddFlags adds the flags which related to fluxcd
func (o *FluxCDOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "fluxcd-enabled", false, "Enable FluxCD APIs")
}

// GetGitOpsEngine return gitops engine type
func GetGitOpsEngine(argoOption *ArgoCDOption, fluxOption *FluxCDOption) v1alpha1.Engine {
	if argoOption.Enabled && !fluxOption.Enabled {
		return v1alpha1.ArgoCD
	}
	if fluxOption.Enabled && !argoOption.Enabled {
		return v1alpha1.FluxCD
	}
	return ""
}
