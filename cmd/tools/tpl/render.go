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

package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

type renderOption struct {
	pattern string
}

func createRenderCommand() (cmd *cobra.Command) {
	opt := &renderOption{}
	cmd = &cobra.Command{
		Use:     "render",
		Short:   "Render Pipeline and Step templates",
		Aliases: []string{"r"},
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.pattern, "pattern", "p", "*.yaml",
		"The template file path pattern")
	return
}

func (o *renderOption) runE(cmd *cobra.Command, args []string) (err error) {
	var files []string
	if files, err = filepath.Glob(o.pattern); err != nil {
		err = fmt.Errorf("failed to find file with pattern: %s, error: %v", o.pattern, err)
		return
	}

	for i := range files {
		item := files[i]

		var data []byte
		if data, err = ioutil.ReadFile(item); err != nil {
			err = fmt.Errorf("failed to read file: %s, error %v", item, err)
			return
		}

		stepTemplate := &v1alpha3.ClusterStepTemplate{}
		if err = yaml.Unmarshal(data, stepTemplate); err != nil {
			err = fmt.Errorf("failed to parse ClusterStepTemplate from file: %s, error %v", item, err)
			return
		}

		var output string
		if output, err = stepTemplate.Spec.Render(map[string]interface{}{}, &v1.Secret{}); err != nil {
			err = fmt.Errorf("failed to render ClusterStepTemplate from file: %s, error %v", item, err)
			return
		}
		cmd.Println(output)
	}
	return
}
