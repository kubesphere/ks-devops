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
	"os"

	"github.com/spf13/pflag"

	"github.com/kubesphere/ks-devops/pkg/api/gitops/v1alpha1"
)

type GitOpsOptions struct {
	// RootDir is the root directory to save git repositories data
	RootDir string `json:"rootDir,omitempty" yaml:"rootDir,omitempty"`

	// NewFilePerm is the permissions for new files or folders in git repository, default is 0755.
	NewFilePerm os.FileMode `json:"newFilePerm,omitempty" yaml:"newFilePerm,omitempty"`
}

func NewGitOpsOptions() *GitOpsOptions {
	return &GitOpsOptions{
		RootDir:     "/gitops",
		NewFilePerm: 0755,
	}
}

// ArgoCDOption as the ArgoCD integration configuration
type ArgoCDOption struct {
	Enabled   bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" description:"enabled ArgoCD"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty" description:"Which namespace the ArgoCD located"`
}

// AddFlags adds the flags which related to argocd
func (o *ArgoCDOption) AddFlags(fs *pflag.FlagSet, parentOptions *ArgoCDOption) {
	fs.BoolVar(&o.Enabled, "argocd-enabled", parentOptions.Enabled, "Enable ArgoCD APIs")
	// see also https://argo-cd.readthedocs.io/en/stable/getting_started/
	fs.StringVar(&o.Namespace, "argocd-namespace", parentOptions.Namespace, "Which namespace the ArgoCD located")
}

// FluxCDOption as the FluxCD integration configuration
type FluxCDOption struct {
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty" description:"enabled FluxCD"`
}

// AddFlags adds the flags which related to fluxcd
func (o *FluxCDOption) AddFlags(fs *pflag.FlagSet, parentOptions *FluxCDOption) {
	fs.BoolVar(&o.Enabled, "fluxcd-enabled", parentOptions.Enabled, "Enable FluxCD APIs")
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
