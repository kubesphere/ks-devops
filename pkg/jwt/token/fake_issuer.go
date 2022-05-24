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

package token

import (
	"time"

	"k8s.io/apiserver/pkg/authentication/user"
)

// FakeIssuer is a fake issue for the test purpose
type FakeIssuer struct {
	Token        string
	IssueToError error
	VerifyError  error
}

func (f *FakeIssuer) IssueTo(user user.Info, tokenType TokenType, expiresIn time.Duration) (string, error) {
	return f.Token, f.IssueToError
}

// Verify verifies a token, and return a user info if it's a valid token, otherwise return error
func (f *FakeIssuer) Verify(string) (user.Info, TokenType, error) {
	return &user.DefaultInfo{}, "", f.VerifyError
}

// VerifyWithoutClaimsValidation verifies a token, but skip the claims validation
func (f *FakeIssuer) VerifyWithoutClaimsValidation(token string) (user.Info, TokenType, error) {
	return f.Verify(token)
}
