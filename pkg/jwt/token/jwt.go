/*
Copyright 2020 The KubeSphere Authors.

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

package token

import (
	"fmt"
	"time"

	"kubesphere.io/devops/pkg/server/errors"

	"github.com/form3tech-oss/jwt-go"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
)

const (
	DefaultIssuerName = "kubesphere"
)

type Claims struct {
	Username  string              `json:"username"`
	Groups    []string            `json:"groups,omitempty"`
	Extra     map[string][]string `json:"extra,omitempty"`
	TokenType TokenType           `json:"token_type"`
	// Currently, we are not using any field in jwt.StandardClaims
	jwt.StandardClaims
}

type jwtTokenIssuer struct {
	name   string
	secret []byte
	// Maximum time difference
	maximumClockSkew time.Duration
}

func (s *jwtTokenIssuer) VerifyWithoutClaimsValidation(tokenString string) (userInfo user.Info, tokenType TokenType, err error) {
	parser := &jwt.Parser{
		SkipClaimsValidation: true,
	}

	// set up the default values
	tokenType = "clm.TokenType"
	userInfo = &user.DefaultInfo{}

	var token *jwt.Token
	// TODO consider to verify the token parse result
	if token, _ = parser.Parse(tokenString, s.keyFunc); token != nil {
		var mapClaims jwt.MapClaims
		var ok bool
		if mapClaims, ok = token.Claims.(jwt.MapClaims); !ok {
			err = errors.New("unexpect type (should be map[string]interface{}) of JWT token claims: %v", token.Claims)
		} else if username := getUserFromClaims(mapClaims); username != "" {
			userInfo = &user.DefaultInfo{
				Name: username,
			}
		} else {
			err = errors.New("no sub or username found from jwt, claims: %v", mapClaims)
		}
	}
	return
}

// getUserFromClaims returns the username which support sub or username as key
func getUserFromClaims(claims jwt.MapClaims) (user string) {
	var ok bool
	var val interface{}
	if val, ok = claims["sub"]; !ok {
		val, _ = claims["username"]
	}

	if strVal, ok := val.(string); ok {
		user = strVal
	}
	return
}

func (s *jwtTokenIssuer) Verify(tokenString string) (user.Info, TokenType, error) {
	clm := &Claims{}
	// verify token signature and expiration time
	_, err := jwt.ParseWithClaims(tokenString, clm, s.keyFunc)
	if err != nil {
		klog.V(4).Info(err)
		return nil, "", err
	}
	return &user.DefaultInfo{Name: clm.Username, Groups: clm.Groups, Extra: clm.Extra}, clm.TokenType, nil
}

func (s *jwtTokenIssuer) IssueTo(user user.Info, tokenType TokenType, expiresIn time.Duration) (string, error) {
	issueAt := time.Now().Unix() - int64(s.maximumClockSkew.Seconds())
	notBefore := issueAt
	clm := &Claims{
		Username:  user.GetName(),
		Groups:    user.GetGroups(),
		Extra:     user.GetExtra(),
		TokenType: tokenType,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issueAt,
			Issuer:    s.name,
			NotBefore: notBefore,
		},
	}

	if expiresIn > 0 {
		clm.ExpiresAt = clm.IssuedAt + int64(expiresIn.Seconds())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clm)

	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		klog.V(4).Info(err)
		return "", err
	}

	return tokenString, nil
}

func (s *jwtTokenIssuer) keyFunc(token *jwt.Token) (i interface{}, err error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
		return s.secret, nil
	} else {
		return nil, fmt.Errorf("expect token signed with HMAC but got %v", token.Header["alg"])
	}
}

func NewTokenIssuer(secret string, maximumClockSkew time.Duration) Issuer {
	return &jwtTokenIssuer{
		name:             DefaultIssuerName,
		secret:           []byte(secret),
		maximumClockSkew: maximumClockSkew,
	}
}
