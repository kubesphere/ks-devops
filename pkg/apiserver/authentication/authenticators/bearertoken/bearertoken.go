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
