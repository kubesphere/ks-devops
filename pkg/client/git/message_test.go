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
