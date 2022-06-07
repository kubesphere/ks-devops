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
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLastChanges(t *testing.T) {
	type args struct {
		jsonText string
	}
	tests := []struct {
		name           string
		args           args
		wantLastChange LastChanges
		wantErr        assert.ErrorAssertionFunc
	}{{
		name: "normal JSON data with map format",
		args: args{jsonText: `{"master":"1234"}`},
		wantLastChange: map[string]string{
			"master": "1234",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLastChange, err := GetLastChanges(tt.args.jsonText)
			if !tt.wantErr(t, err, fmt.Sprintf("GetLastChanges(%v)", tt.args.jsonText)) {
				return
			}
			assert.Equalf(t, tt.wantLastChange, gotLastChange, "GetLastChanges(%v)", tt.args.jsonText)
		})
	}
}

func TestLastChanges_Update(t *testing.T) {
	type args struct {
		ref  string
		hash string
	}
	tests := []struct {
		name string
		l    LastChanges
		args args
		want LastChanges
	}{{
		name: "update the not existing value",
		l:    map[string]string{},
		args: args{
			ref:  "master",
			hash: "2345",
		},
		want: map[string]string{
			"master": "2345",
		},
	}, {
		name: "update the existing value",
		l: map[string]string{
			"master": "1234",
		},
		args: args{
			ref:  "master",
			hash: "2345",
		},
		want: map[string]string{
			"master": "2345",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.l.Update(tt.args.ref, tt.args.hash), "Update(%v, %v)", tt.args.ref, tt.args.hash)
		})
	}
}

func TestLastChanges_LastHash(t *testing.T) {
	type args struct {
		ref string
	}
	tests := []struct {
		name     string
		l        LastChanges
		args     args
		wantHash string
	}{{
		name: "normal case",
		l: map[string]string{
			"master": "1234",
		},
		args:     args{ref: "master"},
		wantHash: "1234",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantHash, tt.l.LastHash(tt.args.ref), "LastHash(%v)", tt.args.ref)
		})
	}
}

func TestLastChanges_String(t *testing.T) {
	tests := []struct {
		name string
		l    LastChanges
		want string
	}{{
		name: "normal case",
		l: map[string]string{
			"master": "1234",
		},
		want: `{"master":"1234"}`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.l.String(), "String()")
		})
	}
}
