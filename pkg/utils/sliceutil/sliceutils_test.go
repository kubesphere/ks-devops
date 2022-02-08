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

package sliceutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddToSlice(t *testing.T) {
	type args struct {
		item  string
		array []string
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, array []string)
	}{{
		name: "no existing",
		args: args{
			item:  "a",
			array: []string{"b"},
		},
		verify: func(t *testing.T, array []string) {
			assert.ElementsMatch(t, []string{"a", "b"}, array)
		},
	}, {
		name: "existing",
		args: args{
			item:  "a",
			array: []string{"b", "a"},
		},
		verify: func(t *testing.T, array []string) {
			assert.ElementsMatch(t, []string{"a", "b"}, array)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.array = AddToSlice(tt.args.item, tt.args.array)
			tt.verify(t, tt.args.array)
		})
	}
}

func TestRemoveString(t *testing.T) {
	type args struct {
		slice  []string
		remove func(item string) bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{{
		name: "remove a not exist item",
		args: args{
			slice: []string{"a", "b"},
			remove: func(item string) bool {
				return item == "c"
			},
		},
		want: []string{"a", "b"},
	}, {
		name: "remove the exit item",
		args: args{
			slice: []string{"a", "b"},
			remove: func(item string) bool {
				return item == "b"
			},
		},
		want: []string{"a"},
	}, {
		name: "remove the exit item with the function",
		args: args{
			slice:  []string{"a", "b"},
			remove: SameItem("b"),
		},
		want: []string{"a"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RemoveString(tt.args.slice, tt.args.remove), "RemoveString(%v, %v)", tt.args.slice, tt.args.remove)
		})
	}
}

func TestHasString(t *testing.T) {
	type args struct {
		slice []string
		str   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "has item",
		args: args{
			slice: []string{"a", "b"},
			str:   "a",
		},
		want: true,
	}, {
		name: "not have item",
		args: args{
			slice: []string{"a", "b"},
			str:   "c",
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, HasString(tt.args.slice, tt.args.str), "HasString(%v, %v)", tt.args.slice, tt.args.str)
		})
	}
}
