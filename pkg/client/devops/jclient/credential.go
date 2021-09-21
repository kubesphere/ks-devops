package jclient

import (
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/client/devops"
)

func (j *JenkinsClient) CreateCredentialInProject(projectID string, credential *v1.Secret) (string, error) {
	return j.jenkins.CreateCredentialInProject(projectID, credential)
}

func (j *JenkinsClient) UpdateCredentialInProject(projectID string, credential *v1.Secret) (string, error) {
	return j.jenkins.UpdateCredentialInProject(projectID, credential)
}

func (j *JenkinsClient) GetCredentialInProject(projectID, id string) (*devops.Credential, error) {
	return j.jenkins.GetCredentialInProject(projectID, id)
}

func (j *JenkinsClient) DeleteCredentialInProject(projectID, id string) (string, error) {
	return j.jenkins.DeleteCredentialInProject(projectID, id)
}
