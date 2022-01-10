package jclient

// ReloadConfiguration reloads the Jenkins configuration
func (j *JenkinsClient) ReloadConfiguration() error {
	return j.jenkins.ReloadConfiguration()
}
