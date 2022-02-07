package v1alpha1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddonInstallStrategy_IsValid(t *testing.T) {
	tests := []struct {
		name string
		a    AddonInstallStrategy
		want bool
	}{{
		name: "normal case - simple",
		a:    AddonInstallStrategySimple,
		want: true,
	}, {
		name: "normal case - helm",
		a:    AddonInstallStrategyHelm,
		want: true,
	}, {
		name: "normal case - operator",
		a:    AddonInstallStrategyOperator,
		want: true,
	}, {
		name: "normal case - simple-operator",
		a:    AddonInstallStrategySimpleOperator,
		want: true,
	}, {
		name: "a fake strategy",
		a:    AddonInstallStrategy("fake"),
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.a.IsValid(), "IsValid()")
		})
	}
}
