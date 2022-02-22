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
package template

import (
	"k8s.io/client-go/kubernetes/scheme"
	kapisv1alpha1 "kubesphere.io/devops/pkg/kapis/devops/v1alpha1/common"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_newHandler(t *testing.T) {
	fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
	type args struct {
		options *kapisv1alpha1.Options
	}
	tests := []struct {
		name string
		args args
		want *handler
	}{{
		name: "Should set handler correctly",
		args: args{
			options: &kapisv1alpha1.Options{
				GenericClient: fakeClient,
			},
		},
		want: &handler{
			genericClient: fakeClient,
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newHandler(tt.args.options); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
