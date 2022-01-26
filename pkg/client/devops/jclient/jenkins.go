package jclient

import (
	"github.com/jenkins-zh/jenkins-client/pkg/casc"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/devops/jenkins"
)

// JenkinsClient represents a client of Jenkins
type JenkinsClient struct {
	Core    core.JenkinsCore
	jenkins devops.Interface // For refactor purpose only
}

// ApplyNewSource apply a new source
func (j *JenkinsClient) ApplyNewSource(s string) (err error) {
	client := casc.Manager{
		JenkinsCore: j.Core,
	}
	if err = client.CheckNewSource(s); err == nil {
		err = client.Replace(s)
	}
	return
}

var _ devops.Interface = &JenkinsClient{}

// NewJenkinsClient creates a Jenkins client
func NewJenkinsClient(options *jenkins.Options) (*JenkinsClient, error) {
	jenkinsCore := core.JenkinsCore{
		URL:      options.Host,
		UserName: options.Username,
		Token:    options.Password,
	}
	crumbIssuer, err := jenkinsCore.GetCrumb()
	if err != nil {
		return nil, err
	} else if crumbIssuer != nil {
		jenkinsCore.JenkinsCrumb = *crumbIssuer
	}

	devopsClient, _ := jenkins.NewDevopsClient(options) // For refactor purpose only
	return &JenkinsClient{
		Core:    jenkinsCore,
		jenkins: devopsClient, // For refactor purpose only
	}, nil
}
