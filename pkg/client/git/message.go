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

// VerifyResponse represents a response of SCM auth verify
type VerifyResponse struct {
	// Message is the detail of the result
	Message string `json:"message"`
	// Code represents a group of cases
	Code int `json:"code"`
}

func VerifyPass() *VerifyResponse {
	return &VerifyResponse{
		Message: "ok",
	}
}

func VerifyFailed(message string, code int) *VerifyResponse {
	return &VerifyResponse{
		Message: message,
		Code:    code,
	}
}

func VerifyResult(err error, code int) *VerifyResponse {
	if err == nil {
		return VerifyPass()
	}
	return VerifyFailed(err.Error(), code)
}
