package jclient

import (
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/client/devops"
)

// CreateCredentialInProject creates a credential, then returns the ID
func (j *JenkinsClient) CreateCredentialInProject(projectID string, credential *v1.Secret) (string, error) {
	return j.jenkins.CreateCredentialInProject(projectID, credential)
}

// UpdateCredentialInProject updates a credential
func (j *JenkinsClient) UpdateCredentialInProject(projectID string, credential *v1.Secret) (string, error) {
	return j.jenkins.UpdateCredentialInProject(projectID, credential)
}

// GetCredentialInProject returns a credential
func (j *JenkinsClient) GetCredentialInProject(projectID, id string) (*devops.Credential, error) {
	return j.jenkins.GetCredentialInProject(projectID, id)
}

// DeleteCredentialInProject deletes a credential
func (j *JenkinsClient) DeleteCredentialInProject(projectID, id string) (string, error) {
	return j.jenkins.DeleteCredentialInProject(projectID, id)
}
