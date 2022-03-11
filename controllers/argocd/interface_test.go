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

package argocd

import (
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/controllers/core"
	"testing"
)

func TestInterfaceImplement(t *testing.T) {
	type interInstance struct {
		NamedReconciler core.NamedReconciler
		GroupReconciler core.GroupReconciler
	}

	tests := []struct {
		name     string
		instance interInstance
	}{{
		name: "MultiClusterReconciler",
		instance: interInstance{
			NamedReconciler: &MultiClusterReconciler{},
			GroupReconciler: &MultiClusterReconciler{},
		},
	}, {
		name: "ApplicationReconciler",
		instance: interInstance{
			NamedReconciler: &ApplicationReconciler{},
			GroupReconciler: &ApplicationReconciler{},
		},
	}, {
		name: "ApplicationStatusReconciler",
		instance: interInstance{
			NamedReconciler: &ApplicationStatusReconciler{},
			GroupReconciler: &ApplicationStatusReconciler{},
		},
	}, {
		name: "Reconciler",
		instance: interInstance{
			NamedReconciler: &Reconciler{},
			GroupReconciler: &Reconciler{},
		},
	}}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.instance.NamedReconciler)
			assert.NotEmpty(t, tt.instance.NamedReconciler.GetName())
			assert.NotNil(t, tt.instance.GroupReconciler)
			assert.NotEmpty(t, tt.instance.GroupReconciler.GetGroupName())
		})
	}
}
