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

package v1alpha1

import (
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"k8s.io/kube-openapi/pkg/common"
	"testing"
)

func TestGetOpenAPIDefinitions(t *testing.T) {
	type args struct {
		ref common.ReferenceCallback
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, result map[string]common.OpenAPIDefinition)
	}{{
		name: "normal case",
		args: args{
			ref: func(path string) spec.Ref {
				return spec.Ref{}
			},
		},
		verify: func(t *testing.T, result map[string]common.OpenAPIDefinition) {
			assert.NotNil(t, result)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOpenAPIDefinitions(tt.args.ref)
			tt.verify(t, result)
		})
	}
}
