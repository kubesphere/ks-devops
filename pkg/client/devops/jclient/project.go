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

// CreateDevOpsProject creates a devops project
func (j *JenkinsClient) CreateDevOpsProject(projectID string) (string, error) {
	return j.jenkins.CreateDevOpsProject(projectID)
}

// DeleteDevOpsProject deletes a devops project
func (j *JenkinsClient) DeleteDevOpsProject(projectID string) error {
	return j.jenkins.DeleteDevOpsProject(projectID)
}

// GetDevOpsProject returns the devops project
func (j *JenkinsClient) GetDevOpsProject(projectID string) (string, error) {
	return j.jenkins.GetDevOpsProject(projectID)
}
