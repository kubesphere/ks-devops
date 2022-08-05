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

package stringutils

import "testing"

func TestSetOrDefault(t *testing.T) {
	type args struct {
		val    string
		defVal string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "empty string",
		args: args{
			val:    "",
			defVal: "abc",
		},
		want: "abc",
	}, {
		name: "not empty string",
		args: args{
			val:    "def",
			defVal: "abc",
		},
		want: "def",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetOrDefault(tt.args.val, tt.args.defVal); got != tt.want {
				t.Errorf("SetOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "normal",
		args: args{s: "abcd"},
		want: "dcba",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.args.s); got != tt.want {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}
