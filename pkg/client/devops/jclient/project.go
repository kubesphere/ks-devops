package jclient

func (j *JenkinsClient) CreateDevOpsProject(projectID string) (string, error) {
	return j.jenkins.CreateDevOpsProject(projectID)
}

func (j *JenkinsClient) DeleteDevOpsProject(projectID string) error {
	return j.jenkins.DeleteDevOpsProject(projectID)
}

func (j *JenkinsClient) GetDevOpsProject(projectID string) (string, error) {
	return j.jenkins.GetDevOpsProject(projectID)
}
