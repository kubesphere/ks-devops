/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"
	apiserverApp "kubesphere.io/devops/cmd/apiserver/app"
	"kubesphere.io/devops/cmd/apiserver/app/options"
	controllerApp "kubesphere.io/devops/cmd/controller/app"
	controllerOpt "kubesphere.io/devops/cmd/controller/app/options"
	"kubesphere.io/devops/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use: "ks-devops",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			stopChain := signals.SetupSignalHandler()
			go func(stopCh <-chan struct{}) {
				if err := runControllerManager(stopCh); err != nil {
					panic(err)
				}
			}(stopChain)
			err = runAPIServer(stopChain)
			return
		},
	}
}

func runAPIServer(stopCh <-chan struct{}) error {
	s := options.NewServerRunOptions()

	// Load configuration from file
	conf, err := config.TryLoadFromDisk()
	if err == nil {
		s = &options.ServerRunOptions{
			GenericServerRunOptions: s.GenericServerRunOptions,
			Config:                  conf,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	if errs := s.Validate(); len(errs) != 0 {
		return utilerrors.NewAggregate(errs)
	}

	return apiserverApp.Run(s, stopCh)
}

func runControllerManager(stopCh <-chan struct{}) (err error) {
	var conf *config.Config
	s := controllerOpt.NewDevOpsControllerManagerOptions()
	conf, err = config.TryLoadFromDisk()
	if err == nil {
		// make sure LeaderElection is not nil
		s = &controllerOpt.DevOpsControllerManagerOptions{
			KubernetesOptions: conf.KubernetesOptions,
			JenkinsOptions:    conf.JenkinsOptions,
			S3Options:         conf.S3Options,
			LeaderElection:    s.LeaderElection,
			LeaderElect:       s.LeaderElect,
			WebhookCertDir:    s.WebhookCertDir,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	if errs := s.Validate(); len(errs) == 0 {
		err = controllerApp.Run(s, stopCh)
	}
	return
}
