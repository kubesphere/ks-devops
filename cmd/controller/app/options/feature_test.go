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

package options

import (
	"reflect"
	"testing"
)

func TestFeatureOptions_GetControllers(t *testing.T) {
	type fields struct {
		Controllers map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]bool
	}{{
		name: "no input (be empty) from users",
		fields: fields{
			Controllers: map[string]bool{},
		},
		want: map[string]bool{
			"s2ibinary":        true,
			"s2irun":           true,
			"pipeline":         true,
			"devopsprojects":   true,
			"devopscredential": true,
			"jenkinsconfig":    true,
		},
	}, {
		name: "no input (be nil) from users",
		fields: fields{
			Controllers: nil,
		},
		want: map[string]bool{
			"s2ibinary":        true,
			"s2irun":           true,
			"pipeline":         true,
			"devopsprojects":   true,
			"devopscredential": true,
			"jenkinsconfig":    true,
		},
	}, {
		name: "merge with the input from users",
		fields: fields{
			Controllers: map[string]bool{
				"s2irun": false,
				"fake":   true,
			},
		},
		want: map[string]bool{
			"s2ibinary":        true,
			"s2irun":           false,
			"pipeline":         true,
			"devopsprojects":   true,
			"devopscredential": true,
			"jenkinsconfig":    true,
			"fake":             true,
		},
	}, {
		name: "only enable the specific controllers",
		fields: fields{
			Controllers: map[string]bool{
				"all":  false,
				"fake": true,
			},
		},
		want: map[string]bool{
			"fake": true,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FeatureOptions{
				Controllers: tt.fields.Controllers,
			}
			if got := o.GetControllers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetControllers() = %v, want %v", got, tt.want)
			}
		})
	}
}
