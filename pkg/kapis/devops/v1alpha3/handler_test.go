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

package v1alpha3

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSuccessGenericArrayResponse(t *testing.T) {
	type args struct {
		data []string
	}
	tests := []struct {
		name string
		args args
		want *GenericArrayResponse
	}{{
		name: "normal",
		args: args{
			data: []string{"good", "bad"},
		},
		want: &GenericArrayResponse{
			Status: "success",
			Data:   []string{"good", "bad"},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSuccessGenericArrayResponse(tt.args.data), "NewSuccessGenericArrayResponse(%v)", tt.args.data)
		})
	}
}
