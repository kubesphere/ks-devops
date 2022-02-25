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

package scm

type organization struct {
	Name string `json:"name"`
	// Avatar is the image of an organization which comes from a git provider
	// try to find a better way to have it. Now, keeping it just because we need to keep compatible with the Jenkins response
	// example: https://avatars.githubusercontent.com/jenkinsci
	Avatar string `json:"avatar"`
}

type repository struct {
	DefaultBranch string `json:"defaultBranch"`
	Name          string `json:"name"`
}
