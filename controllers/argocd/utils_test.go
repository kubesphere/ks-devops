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
	"fmt"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestSetNestedField(t *testing.T) {
	type args struct {
		obj    map[string]interface{}
		value  interface{}
		fields []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
		verify  func(obj map[string]interface{})
	}{{
		name: "Should set field correctly when struct is simple",
		args: args{
			obj: map[string]interface{}{
				"name": "fake-obj-name",
			},
			value: struct {
				Name string `json:"name"`
			}{
				Name: "fake-name",
			},
			fields: []string{"spec"},
		},
		wantErr: assert.NoError,
		verify: func(obj map[string]interface{}) {
			assert.Equal(t, map[string]interface{}{
				"name": "fake-obj-name",
				"spec": map[string]interface{}{
					"name": "fake-name",
				},
			}, obj)
		},
	}, {
		name: "Should set field correctly when struct is complex",
		args: args{
			obj: map[string]interface{}{},
			value: struct {
				Metadata metav1.ObjectMeta `json:"metadata"`
			}{
				Metadata: metav1.ObjectMeta{
					Name:       "fake-name",
					Namespace:  "fake-namespace",
					Generation: 123,
				},
			},
			fields: []string{"spec"},
		},
		wantErr: assert.NoError,
		verify: func(obj map[string]interface{}) {
			assert.Equal(t, map[string]interface{}{
				"spec": map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":              "fake-name",
						"namespace":         "fake-namespace",
						"generation":        float64(123),
						"creationTimestamp": (interface{})(nil),
					},
				},
			}, obj)
		},
	}, {
		name: "Should occur error when value type is int",
		args: args{
			obj:    map[string]interface{}{},
			value:  123,
			fields: []string{"spec"},
		},
		wantErr: assert.Error,
	}, {
		name: "Should occur error when value type is func",
		args: args{
			obj:    map[string]interface{}{},
			value:  func() {},
			fields: []string{"spec"},
		},
		wantErr: assert.Error,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, SetNestedField(tt.args.obj, tt.args.value, tt.args.fields...), fmt.Sprintf("SetNestedField(%v, %v, %v)", tt.args.obj, tt.args.value, tt.args.fields))
			if tt.verify != nil {
				tt.verify(tt.args.obj)
			}
		})
	}
}
