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
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"testing"
)

func Test_toObjects(t *testing.T) {
	createApplication := func(name, namespaces string) *v1alpha1.Application {
		return &v1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespaces,
			},
		}
	}
	type args struct {
		apps []v1alpha1.Application
	}
	tests := []struct {
		name string
		args args
		want []runtime.Object
	}{{
		name: "Should return empty objects when applications is nil",
		args: args{
			apps: nil,
		},
		want: []runtime.Object{},
	}, {
		name: "Should return empty objects when applications is empty slice",
		args: args{
			apps: []v1alpha1.Application{},
		},
		want: []runtime.Object{},
	}, {
		name: "Should return same objects when applications is non-empty slice",
		args: args{
			apps: []v1alpha1.Application{
				*createApplication("fake-name1", "fake-namespace1"),
				*createApplication("fake-name2", "fake-namespace2"),
			},
		},
		want: []runtime.Object{
			createApplication("fake-name1", "fake-namespace1"),
			createApplication("fake-name2", "fake-namespace2"),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, toObjects(tt.args.apps), "toObjects(%v)", tt.args.apps)
		})
	}
}
