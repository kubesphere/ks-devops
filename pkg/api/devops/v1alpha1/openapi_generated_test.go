package v1alpha1

import (
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"k8s.io/kube-openapi/pkg/common"
	"testing"
)

func TestGetOpenAPIDefinitions(t *testing.T) {
	type args struct {
		ref common.ReferenceCallback
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, result map[string]common.OpenAPIDefinition)
	}{{
		name: "normal case",
		args: args{
			ref: func(path string) spec.Ref {
				return spec.Ref{}
			},
		},
		verify: func(t *testing.T, result map[string]common.OpenAPIDefinition) {
			assert.NotNil(t, result)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOpenAPIDefinitions(tt.args.ref)
			tt.verify(t, result)
		})
	}
}
