package bearertoken

import (
	"context"
	"time"

	"k8s.io/apiserver/pkg/authentication/user"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	jwt "kubesphere.io/devops/pkg/jwt/token"
)

// tokenAuthenticator implements an simple auth which only check the format of target JWT token
type tokenAuthenticator struct{}

func New() authenticator.Token {
	return &tokenAuthenticator{}
}

func (a *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (response *authenticator.Response, ok bool, err error) {
	issuer := jwt.NewTokenIssuer("", time.Second)

	var authenticated user.Info
	if authenticated, _, err = issuer.VerifyWithoutClaimsValidation(token); err == nil {
		response = &authenticator.Response{
			User: &user.DefaultInfo{
				Name: authenticated.GetName(),
			}}
		ok = true
	} else {
		ok = false
	}
	return
}
