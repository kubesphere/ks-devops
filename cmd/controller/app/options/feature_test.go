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
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
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
			"jenkins":       true,
			"jenkinsconfig": true,
			"gitrepository": true,
		},
	}, {
		name: "no input (be nil) from users",
		fields: fields{
			Controllers: nil,
		},
		want: map[string]bool{
			"jenkins":       true,
			"jenkinsconfig": true,
			"gitrepository": true,
		},
	}, {
		name: "merge with the input from users",
		fields: fields{
			Controllers: map[string]bool{
				"fake": true,
			},
		},
		want: map[string]bool{
			"jenkins":       true,
			"jenkinsconfig": true,
			"gitrepository": true,
			"fake":          true,
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

func TestFeatureOptions(t *testing.T) {
	opt := NewFeatureOptions()
	assert.NotNil(t, opt)
	assert.Equal(t, []error{}, opt.Validate())

	opt.Controllers = map[string]bool{
		"fake": true,
	}
	assert.Equal(t, []string{"fake"}, opt.knownControllers())

	opt.ExternalAddress = "fake-address"
	assert.Equal(t, "fake-address", opt.ExternalAddress)

	newOpt := NewFeatureOptions()
	assert.Empty(t, newOpt.ExternalAddress)
	opt.ApplyTo(newOpt)
	assert.Equal(t, "fake-address", newOpt.ExternalAddress)

	flagSet := &pflag.FlagSet{}
	opt.AddFlags(flagSet, opt)
	assert.True(t, flagSet.HasFlags())
	assert.NotNil(t, flagSet.Lookup("enabled-controllers"))
	assert.NotNil(t, flagSet.Lookup("system-namespace"))
	assert.NotNil(t, flagSet.Lookup("external-address"))
	assert.NotNil(t, flagSet.Lookup("cluster-name"))
	assert.NotNil(t, flagSet.Lookup("pipelinerun-data-store"))
}
