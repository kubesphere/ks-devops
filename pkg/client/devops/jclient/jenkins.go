package jclient

import (
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/devops/jenkins"
)

type JenkinsClient struct {
	Core    core.JenkinsCore
	jenkins devops.Interface // For refactor purpose only
}
var _ devops.Interface = &JenkinsClient{}
func NewJenkinsClient(options *jenkins.Options) (jenkinsClient *JenkinsClient, err error) {
	core := core.JenkinsCore{
		URL:      options.Host,
		UserName: options.Username,
		Token:    options.Password,
	}
	crumbIssuer, e := core.GetCrumb()
	if e != nil {
		return
	} else if crumbIssuer != nil {
		core.JenkinsCrumb = *crumbIssuer
	}

	jenkins, _ := jenkins.NewDevopsClient(options) // For refactor purpose only
	jenkinsClient = &JenkinsClient{
		Core:    core,
		jenkins: jenkins, // For refactor purpose only
	}
	return
}
