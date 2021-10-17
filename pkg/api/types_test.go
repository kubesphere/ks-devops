/*
Copyright 2020 The KubeSphere Authors.

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

package api

import (
	"reflect"
	"testing"
)

func TestNewListResult(t *testing.T) {
	type testStruct struct {
		key   string
		value string
	}
	type args struct {
		items interface{}
		total int
	}
	tests := []struct {
		name string
		args args
		want *ListResult
	}{{
		name: "Nil items",
		args: args{
			items: nil,
			total: 10,
		},
		want: &ListResult{
			Items:      []interface{}{},
			TotalItems: 10,
		},
	}, {
		name: "Integer itmes",
		args: args{
			items: []int{1, 2, 3},
			total: 100,
		},
		want: &ListResult{
			Items:      []interface{}{1, 2, 3},
			TotalItems: 100,
		},
	}, {
		name: "String itmes",
		args: args{
			items: []string{"a", "b"},
			total: 1000,
		},
		want: &ListResult{
			Items:      []interface{}{"a", "b"},
			TotalItems: 1000,
		},
	}, {
		name: "Boolean items",
		args: args{
			items: []bool{true, false},
			total: 100,
		},
		want: &ListResult{
			Items:      []interface{}{true, false},
			TotalItems: 100,
		},
	}, {
		name: "Struct items",
		args: args{
			items: []testStruct{{
				key:   "hello",
				value: "world",
			}},
			total: 100,
		},
		want: &ListResult{
			Items: []interface{}{
				testStruct{
					key:   "hello",
					value: "world",
				},
			},
			TotalItems: 100,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewListResult(tt.args.items, tt.args.total); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewListResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
