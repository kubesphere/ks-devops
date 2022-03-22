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

package argocd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
)

func filterByLabels(applications []v1alpha1.Application, labels map[string]string) []v1alpha1.Application {
	if len(labels) == 0 {
		return applications
	}
	filtered := make([]v1alpha1.Application, 0)
	for i := range applications {
		contain := true
		for key, value := range labels {
			if value != "" && applications[i].Labels[key] != value {
				contain = false
			}
		}
		if contain {
			filtered = append(filtered, applications[i])
		}
	}
	return filtered
}

func toObjects(apps []v1alpha1.Application) []runtime.Object {
	objs := make([]runtime.Object, len(apps))
	for i := range apps {
		objs[i] = &apps[i]
	}
	return objs
}
