package jclient

import (
	"kubesphere.io/devops/pkg/client/devops"
)

// RoleOperator

func (j *JenkinsClient) AddGlobalRole(roleName string, ids devops.GlobalPermissionIds, overwrite bool) error {
	// TODO: Refactor function
	return j.jenkins.AddGlobalRole(roleName, ids, overwrite)
}

func (j *JenkinsClient) GetGlobalRole(roleName string) (string, error) {
	// TODO: Refactor function
	return j.jenkins.GetGlobalRole(roleName)
}

func (j *JenkinsClient) AddProjectRole(roleName string, pattern string, ids devops.ProjectPermissionIds, overwrite bool) error {
	// TODO: Refactor function
	return j.jenkins.AddProjectRole(roleName, pattern, ids, overwrite)
}

func (j *JenkinsClient) DeleteProjectRoles(roleName ...string) error {
	// TODO: Refactor function
	return j.jenkins.DeleteProjectRoles(roleName...)
}

func (j *JenkinsClient) AssignProjectRole(roleName string, sid string) error {
	// TODO: Refactor function
	return j.jenkins.AssignProjectRole(roleName, sid)
}

func (j *JenkinsClient) UnAssignProjectRole(roleName string, sid string) error {
	// TODO: Refactor function
	return j.jenkins.UnAssignProjectRole(roleName, sid)
}

func (j *JenkinsClient) AssignGlobalRole(roleName string, sid string) error {
	// TODO: Refactor function
	return j.jenkins.AssignGlobalRole(roleName, sid)
}

func (j *JenkinsClient) UnAssignGlobalRole(roleName string, sid string) error {
	// TODO: Refactor function
	return j.jenkins.UnAssignGlobalRole(roleName, sid)
}

func (j *JenkinsClient) DeleteUserInProject(sid string) error {
	// TODO: Refactor function
	return j.jenkins.DeleteUserInProject(sid)
}
