package v1alpha3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeDetails_InitApprovable(t *testing.T) {
	tests := []struct {
		name        string
		nodeDetails NodeDetails
		assertion   func(NodeDetails)
	}{{
		name: "InitApprovable",
		nodeDetails: []NodeDetail{{
			Steps: []Step{
				{
					Approvable: false,
				},
			},
		}},
		assertion: func(nodeDetails NodeDetails) {
			assert.True(t, nodeDetails[0].Steps[0].Approvable)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.nodeDetails.InitApprovable()
		})
	}
}
