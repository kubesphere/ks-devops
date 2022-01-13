package git

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerifyPass(t *testing.T) {
	verifyPass := VerifyPass()
	assert.NotNil(t, verifyPass)
	assert.Equal(t, 0, verifyPass.Code)
	assert.Equal(t, "ok", verifyPass.Message)
}

func TestVerifyResult(t *testing.T) {
	type args struct {
		message string
		code    int
		err     error
	}
	tests := []struct {
		name string
		args args
		want *VerifyResponse
	}{{
		name: "no error",
		args: args{},
		want: &VerifyResponse{Message: "ok"},
	}, {
		name: "error case",
		args: args{
			message: "failed",
			code:    1234,
			err:     errors.New("failed"),
		},
		want: &VerifyResponse{
			Message: "failed",
			Code:    1234,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, VerifyResult(tt.args.err, tt.args.code), "VerifyResult(%v, %v, %v)", tt.args.message, tt.args.code, tt.args.err)
		})
	}
}
