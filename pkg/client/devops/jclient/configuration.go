package jclient

func (j *JenkinsClient) ReloadConfiguration() error {
	return j.jenkins.ReloadConfiguration()
}
