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
package kapis

import (
	"io"
	"kubesphere.io/devops/pkg/server/errors"
	"reflect"
	"testing"
)

func TestIgnoreEOF(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want error
	}{{
		name: "Should return nil if error is io.EOF",
		args: args{
			err: io.EOF,
		},
		want: nil,
	}, {
		name: "Should return the same error if error is not io.EOF",
		args: args{
			errors.New("Fake Error"),
		},
		want: errors.New("Fake Error"),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreEOF(tt.args.err); !reflect.DeepEqual(err, tt.want) {
				t.Errorf("IgnoreEOF() error = %v, want %v", err, tt.want)
			}
		})
	}
}
