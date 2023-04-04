/*
Copyright 2023 The KubeSphere Authors.

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

package app

import (
	"github.com/spf13/cobra"
	"kubesphere.io/devops/pkg/client/k8s"
)

var toolOpt *ToolOption

type ToolOption struct {
	Namespace     string
	ConfigMapName string

	K8sClient k8s.Client
}

func (o *ToolOption) runHelpE(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

// NewToolsCmd creates a root command for tools
func NewToolsCmd() (cmd *cobra.Command) {
	toolOpt = &ToolOption{}

	rootCmd := &cobra.Command{
		Use:   "devops-tool",
		Short: "Tools for DevOps apiserver and controller-manager",
		RunE:  toolOpt.runHelpE,
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&toolOpt.Namespace, "namespace", "n", "kubesphere-devops-system",
		"The namespace of DevOps service")
	flags.StringVarP(&toolOpt.ConfigMapName, "configmap", "c", "devops-config",
		"The configmap name of DevOps service")

	rootCmd.AddCommand(NewInitCmd())
	return rootCmd
}
