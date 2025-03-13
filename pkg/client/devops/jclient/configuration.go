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

package jclient

import "github.com/jenkins-zh/jenkins-client/pkg/casc"

// ReloadConfiguration reloads the Jenkins Configuration as Code YAML file
func (j *JenkinsClient) ReloadConfiguration() (err error) {
	client := casc.Manager{}
	if j != nil {
		client.JenkinsCore = j.Core
	}
	err = client.Reload()
	return
}

// ApplyNewSource apply a new source
func (j *JenkinsClient) ApplyNewSource(s string) (err error) {
	client := casc.Manager{}
	if j != nil {
		client.JenkinsCore = j.Core
	}
	if err = client.CheckNewSource(s); err == nil {
		err = client.Replace(s)
	}
	return
}
