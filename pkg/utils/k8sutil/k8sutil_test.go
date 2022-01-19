package k8sutil

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestIsControlledBy(t *testing.T) {
	type args struct {
		ownerReferences []metav1.OwnerReference
		kind            string
		name            string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "normal case, has the reference",
		args: args{
			ownerReferences: []metav1.OwnerReference{{
				Kind: "kind",
				Name: "name",
			}},
			kind: "kind",
			name: "name",
		},
		want: true,
	}, {
		name: "not have the reference",
		args: args{
			ownerReferences: []metav1.OwnerReference{{
				Kind: "kind",
				Name: "name",
			}},
			kind: "kind",
			name: "fake",
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsControlledBy(tt.args.ownerReferences, tt.args.kind, tt.args.name), "IsControlledBy(%v, %v, %v)", tt.args.ownerReferences, tt.args.kind, tt.args.name)
		})
	}
}
