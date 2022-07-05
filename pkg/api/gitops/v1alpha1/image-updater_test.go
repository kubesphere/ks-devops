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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteMethod_GetValue(t *testing.T) {
	tests := []struct {
		name string
		w    WriteMethod
		want string
	}{{
		name: "built-in",
		w:    WriteMethodBuiltIn,
		want: "argocd",
	}, {
		name: "git",
		w:    WriteMethodGit,
		want: "git",
	}, {
		name: "invalid",
		w:    WriteMethod("invalid"),
		want: "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.w.GetValue(), "GetValue()")
		})
	}
}
