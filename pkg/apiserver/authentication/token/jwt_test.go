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
	"github.com/form3tech-oss/jwt-go"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apiserver/pkg/authentication/user"
)

func TestTokenVerifyWithoutCacheValidate(t *testing.T) {
	issuer := NewTokenIssuer("kubesphere", 0)

	admin := &user.DefaultInfo{
		Name: "admin",
	}

	tokenString, err := issuer.IssueTo(admin, AccessToken, 0)

	if err != nil {
		t.Fatal(err)
	}

	got, _, err := issuer.Verify(tokenString)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(got, admin); diff != "" {
		t.Error("token validate failed")
	}
}

func Test_getUserFromClaims(t *testing.T) {
	type args struct {
		claims jwt.MapClaims
	}
	tests := []struct {
		name     string
		args     args
		wantUser string
	}{{
		name: "without any valid key",
		args: args{
			claims: jwt.MapClaims{},
		},
		wantUser: "",
	}, {
		name: "with key: sub",
		args: args{
			claims: jwt.MapClaims{
				"sub": "rick",
			},
		},
		wantUser: "rick",
	}, {
		name: "with key: username",
		args: args{
			claims: jwt.MapClaims{
				"username": "rick",
			},
		},
		wantUser: "rick",
	}, {
		name: "with key: username and sub, take sub first",
		args: args{
			claims: jwt.MapClaims{
				"sub": "rick",
				"username": "mark",
			},
		},
		wantUser: "rick",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUser := getUserFromClaims(tt.args.claims); gotUser != tt.wantUser {
				t.Errorf("getUserFromClaims() = %v, want %v", gotUser, tt.wantUser)
			}
		})
	}
}
