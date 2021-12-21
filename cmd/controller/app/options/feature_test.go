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
