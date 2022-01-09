package v1alpha3

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestResource(t *testing.T) {
	type args struct {
		resource string
	}
	tests := []struct {
		name string
		args args
		want schema.GroupResource
	}{{
		name: "normal case",
		args: args{
			resource: "pipeline",
		},
		want: Resource("pipeline"),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Resource(tt.args.resource), "Resource(%v)", tt.args.resource)
		})
	}
}
