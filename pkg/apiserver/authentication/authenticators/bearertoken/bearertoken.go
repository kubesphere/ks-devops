package bearertoken

import (
	"context"
	"k8s.io/apiserver/pkg/authentication/user"
	"time"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	jwt "kubesphere.io/devops/pkg/apiserver/authentication/token"
)

// tokenAuthenticator implements an anonymous auth
type tokenAuthenticator struct{}

func New() authenticator.Token {
	return &tokenAuthenticator{}
}

func (a *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	issuer := jwt.NewTokenIssuer("", time.Second)

	if authenticated, _, err := issuer.VerifyWithoutClaimsValidation(token); err == nil {
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name: authenticated.GetName(),
			}}, true, nil
	}
	return nil, false, nil
}
