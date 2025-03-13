/*
Copyright 2024 The KubeSphere Authors.

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
)

var toolOpt *ToolOptions

type ToolOptions struct {
	kubeconfig string
}

func (o *ToolOptions) runHelpE(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

// NewToolsCmd creates a root command for tools
func NewToolsCmd() (cmd *cobra.Command) {
	opts := &ToolOptions{}

	rootCmd := &cobra.Command{
		Use:   "devops-tools",
		Short: "Tools for DevOps services",
		RunE:  toolOpt.runHelpE,
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&opts.kubeconfig, "kubeconfig", "k", "",
		"path of kubernetes kubeconfig file, default: Using the inClusterConfig")

	rootCmd.AddCommand(NewRestoreCmd(opts.kubeconfig))
	return rootCmd
}
