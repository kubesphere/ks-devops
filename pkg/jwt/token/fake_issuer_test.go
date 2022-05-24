package token

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apiserver/pkg/authentication/user"
	"testing"
)

func TestFakeIssuer_IssueTo(t *testing.T) {
	f := &FakeIssuer{
		Token:        "Token",
		IssueToError: nil,
		VerifyError:  nil,
	}
	got, err := f.IssueTo(&user.DefaultInfo{}, "type", 0)
	assert.Equal(t, "Token", got)
	assert.Nil(t, err)

	_, _, err = f.Verify("token")
	assert.Nil(t, err)

	_, _, err = f.VerifyWithoutClaimsValidation("token")
	assert.Nil(t, err)
}
