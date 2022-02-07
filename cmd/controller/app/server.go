/*
Copyright 2019 The KubeSphere Authors.

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
	"fmt"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"kubesphere.io/devops/cmd/controller/app/options"
	"kubesphere.io/devops/pkg/apis"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/devops/jclient"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/client/s3"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/indexers"
	"kubesphere.io/devops/pkg/informers"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewControllerManagerCommand() *cobra.Command {
	// Here will create a default devops controller manager options
	s := options.NewDevOpsControllerManagerOptions()
	// Load configuration from disk via viper, /etc/kubesphere/kubesphere.[yaml,json,xxx]
	conf, err := config.TryLoadFromDisk()
	if err == nil {
		// make sure LeaderElection is not nil
		// override devops controller manager options
		s = &options.DevOpsControllerManagerOptions{
			KubernetesOptions: conf.KubernetesOptions,
			JenkinsOptions:    conf.JenkinsOptions,
			S3Options:         conf.S3Options,
			JWTOptions: &options.JWTOptions{
				Secret:           conf.AuthenticationOptions.JwtSecret,
				MaximumClockSkew: conf.AuthenticationOptions.MaximumClockSkew,
			},
			FeatureOptions: s.FeatureOptions,
			LeaderElection: s.LeaderElection,
			LeaderElect:    s.LeaderElect,
			WebhookCertDir: s.WebhookCertDir,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	// Initialize command to run our controllers later
	cmd := &cobra.Command{
		Use:   "controller-manager",
		Short: `KubeSphere DevOps controller manager`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			err = Run(s, signals.SetupSignalHandler())
			return
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	// Add pre-defined flags into command
	namedFlagSets := s.Flags()

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, 0)
	})

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of KubeSphere DevOps controller",
		Run: func(cmd *cobra.Command, args []string) {
			// cmd.Println(version.Get())
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

func Run(s *options.DevOpsControllerManagerOptions, stopCh <-chan struct{}) error {
	// Init k8s client
	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	// Init DevOps client while Jenkins options and Jenkins host
	var devopsClient devops.Interface
	if s.JenkinsOptions != nil && len(s.JenkinsOptions.Host) != 0 {
		// Make sure that Jenkins host is not empty
		devopsClient, err = jclient.NewJenkinsClient(s.JenkinsOptions)
		if err != nil {
			return fmt.Errorf("failed to connect jenkins, please check jenkins status, error: %v", err)
		}
	}

	// Init Jenkins client
	jenkinsCore := core.JenkinsCore{
		URL:      s.JenkinsOptions.Host,
		UserName: s.JenkinsOptions.Username,
		Token:    s.JenkinsOptions.Password,
	}

	// Init informers
	informerFactory := informers.NewInformerFactories(
		kubernetesClient.Kubernetes(),
		kubernetesClient.KubeSphere(),
		kubernetesClient.ApiExtensions())

	mgrOptions := manager.Options{
		CertDir: s.WebhookCertDir,
		Port:    8443,
	}

	if s.LeaderElect {
		mgrOptions = manager.Options{
			CertDir:                 s.WebhookCertDir,
			Port:                    8443,
			LeaderElection:          s.LeaderElect,
			LeaderElectionNamespace: "kubesphere-devops-system",
			LeaderElectionID:        "ks-devops-controller-manager-leader-election",
			LeaseDuration:           &s.LeaderElection.LeaseDuration,
			RetryPeriod:             &s.LeaderElection.RetryPeriod,
			RenewDeadline:           &s.LeaderElection.RenewDeadline,
		}
	}

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klogr.New())
	// Use 8443 instead of 443 cause we need root permission to bind port 443
	// Init controller manager
	mgr, err := manager.New(kubernetesClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}
	apis.AddToScheme(mgr.GetScheme())
	_ = apiextensions.AddToScheme(mgr.GetScheme())

	// Init s3 client
	var s3Client s3.Interface
	if s.S3Options != nil && len(s.S3Options.Endpoint) != 0 {
		s3Client, err = s3.NewS3Client(s.S3Options)
		if err != nil {
			return fmt.Errorf("failed to connect to s3, please check s3 service status, error: %v", err)
		}
	}

	// register common meta types into schemas.
	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	if err = addControllers(mgr,
		kubernetesClient,
		informerFactory,
		devopsClient,
		jenkinsCore,
		s3Client,
		s,
		stopCh); err != nil {
		return fmt.Errorf("unable to register controllers to the manager: %v", err)
	}

	if err := indexers.CreatePipelineRunSCMRefNameIndexer(mgr.GetCache()); err != nil {
		return err
	}

	// Start cache data after all informer is registered
	klog.V(0).Info("Starting cache resource from apiserver...")
	informerFactory.Start(stopCh)

	klog.V(0).Info("Starting the controllers.")
	if err = mgr.Start(stopCh); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}

	return nil
}
