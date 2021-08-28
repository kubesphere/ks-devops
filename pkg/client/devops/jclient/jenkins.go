package jclient

import (
	"io/ioutil"

	client "github.com/jenkins-zh/jenkins-client/pkg/core"
	"gopkg.in/yaml.v2"
)

type JenkinsClient struct {
	
}
type Config struct {
	Jenkins JenkinsConfig `yaml:"devops,omitempty"`
}
type JenkinsConfig struct {
	URL      string `yaml:"host"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token,omitempty"`
}

func GetJenkinsCore() (core client.JenkinsCore, e error) {
	// read configuration files
	jenkinsConfigPath := "/etc/kubesphere/kubesphere.yaml"
	yamlFile, e := ioutil.ReadFile(jenkinsConfigPath)
	if e != nil {
		return
	}
	// yaml parse
	var config Config
	e = yaml.Unmarshal(yamlFile, &config)
	if e != nil {
		return
	}
	// capsulate JenkinsCore
	core = client.JenkinsCore{
		URL:      config.Jenkins.URL,
		UserName: config.Jenkins.UserName,
		Token:    config.Jenkins.Password,
	}
	crumbIssuer, e := core.GetCrumb()
	if e != nil {
		return
	} else if crumbIssuer != nil {
		core.JenkinsCrumb = *crumbIssuer
	}
	return
}
