/*
Copyright 2019 The KubeSphere Authors.

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

package jwttoken

import (
	"context"
	"time"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	jwt "kubesphere.io/devops/pkg/apiserver/authentication/token"
)

// TokenAuthenticator implements kubernetes secret authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given secret. If empty or invalid secret
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type tokenAuthenticator struct {
	secret string
}

func NewTokenAuthenticator(secret string) authenticator.Token {
	return &tokenAuthenticator{
		secret: secret,
	}
}

func (t *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	issuer := jwt.NewTokenIssuer(t.secret, time.Second)

	if authenticated, _, err := issuer.Verify(token); err == nil {
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name:   authenticated.GetName(),
				Extra:  authenticated.GetExtra(),
				Groups: append(authenticated.GetGroups(), user.AllAuthenticated),
			},
		}, true, nil
	} else {
		return nil, false, err
	}
}
