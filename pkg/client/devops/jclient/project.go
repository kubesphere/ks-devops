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
