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

package auth

import (
	"github.com/stretchr/testify/assert"
	user2 "k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/devops/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/devops/pkg/apiserver/authentication/options"
	"kubesphere.io/devops/pkg/client/cache"
	"testing"
	"time"
)

func TestNewTokenOperator(t *testing.T) {
	operator := NewTokenOperator(cache.NewSimpleCache(), &authoptions.AuthenticationOptions{
		JwtSecret: "fake",
		OAuthOptions: &oauth.Options{
			AccessTokenMaxAge: time.Minute,
		},
	})
	assert.NotNil(t, operator)

	// provide a valid jwtSecret
	operator = NewTokenOperator(cache.NewSimpleCache(), &authoptions.AuthenticationOptions{
		JwtSecret: "nQk6f1gM9uPYZHXyyuLoqfSMZAZ5RdYQ",
		OAuthOptions: &oauth.Options{
			AccessTokenMaxAge: time.Second,
		},
	})
	assert.NotNil(t, operator)

	user, err := operator.Verify("fake")
	assert.NotNil(t, err)
	assert.Nil(t, user)

	// test with a valid token, but it's not found in cache
	user, err = operator.Verify("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwidG9rZW5fdHlwZSI6ImFjY2Vzc190b2tlbiIsImlhdCI6MTY2MDU0MjM4MSwiaXNzIjoia3ViZXNwaGVyZSIsIm5iZiI6MTY2MDU0MjM4MX0.5DJQCgTxmb4gn45fChUdanRF0sqNj4uSGDJ2BAci70w")
	assert.NotNil(t, err)
	assert.Nil(t, user)

	token, err := operator.IssueTo(&user2.DefaultInfo{
		Name: "fake",
	})
	assert.Nil(t, err)
	assert.NotNil(t, token)
	assert.NotEmpty(t, token.RefreshToken)
	assert.Equal(t, "Bearer", token.TokenType)

	// this should be a valid user
	user, err = operator.Verify(token.AccessToken)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "fake", user.GetName())

	err = operator.RevokeAllUserTokens("fake")
	assert.Nil(t, err)
}
