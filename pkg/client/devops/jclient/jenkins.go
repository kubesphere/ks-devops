package jclient

import (
	"fmt"
	"time"

	appCfg "github.com/jenkins-zh/jenkins-cli/app/config"
	"github.com/jenkins-zh/jenkins-cli/client"
	"kubesphere.io/devops/pkg/client/devops/jenkins"
)

type JenkinsClient struct {
}

type JenkinsOptions struct {
	Host            string        `json:",omitempty" yaml:"host" description:"Jenkins service host address"`
	Username        string        `json:",omitempty" yaml:"username" description:"Jenkins admin username"`
	Password        string        `json:",omitempty" yaml:"password" description:"Jenkins admin password"`
	MaxConnections  int           `json:"maxConnections,omitempty" yaml:"maxConnections" description:"Maximum connections allowed to connect to Jenkins"`
	Namespace       string        `json:"namespace,omitempty" yaml:"namespace"`
	WorkerNamespace string        `json:"workerNamespace,omitempty" yaml:"workerNamespace"`
	ReloadCasCDelay time.Duration `json:"reloadCasCDelay,omitempty" yaml:"reloadCasCDelay"`
}

var rootJenkinsOptions JenkinsOptions

// newJenkinsServer
func NewJenkinsRootOptions(options *jenkins.Options) error {
	if options.Host == "" {
		return fmt.Errorf("cannot get jenkins host")
	}
	rootJenkinsOptions.Host = options.Host
	rootJenkinsOptions.Username = options.Username
	rootJenkinsOptions.Password = options.Password
	return nil
}

// getCurrentJenkinsClient gets the current jenkins jenkinscore and returns the jenkins server
func getCurrentJenkinsClient(jenkinsCli *client.JenkinsCore) (jenkins *appCfg.JenkinsServer) {
	jenkinsCli.URL = rootJenkinsOptions.Host
	jenkinsCli.UserName = rootJenkinsOptions.Username
	jenkinsCli.Token = rootJenkinsOptions.Password
	return
}
