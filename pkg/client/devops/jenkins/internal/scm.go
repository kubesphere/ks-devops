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

package internal

import (
	"github.com/beevik/etree"
	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"strconv"
)

// common functions for all kinds of scm

func parseFromCloneTrait(cloneTrait *etree.Element) *devopsv1alpha3.GitCloneOption {
	var cloneOption *devopsv1alpha3.GitCloneOption
	if cloneTrait != nil {
		if cloneExtension := cloneTrait.SelectElement("extension"); cloneExtension != nil {
			cloneOption = &devopsv1alpha3.GitCloneOption{}
			if shallow := cloneExtension.SelectElement("shallow"); shallow != nil {
				if value, err := strconv.ParseBool(shallow.Text()); err == nil {
					cloneOption.Shallow = value
				}
			}
			if timeout := cloneExtension.SelectElement("timeout"); timeout != nil {
				if value, err := strconv.ParseInt(timeout.Text(), 10, 32); err == nil {
					cloneOption.Timeout = int(value)
				}
			}
			if depth := cloneExtension.SelectElement("depth"); depth != nil {
				if value, err := strconv.ParseInt(depth.Text(), 10, 32); err == nil {
					cloneOption.Depth = int(value)
				}
			}
		}
	}
	return cloneOption
}
