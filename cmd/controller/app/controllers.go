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
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"k8s.io/klog"
	"kubesphere.io/devops/cmd/controller/app/options"
	"kubesphere.io/devops/controllers/devopscredential"
	"kubesphere.io/devops/controllers/devopsproject"
	"kubesphere.io/devops/controllers/jenkins/config"
	"kubesphere.io/devops/controllers/jenkins/pipelinerun"
	"kubesphere.io/devops/controllers/pipeline"
	"kubesphere.io/devops/controllers/s2ibinary"
	"kubesphere.io/devops/controllers/s2irun"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/client/s3"
	"kubesphere.io/devops/pkg/informers"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func addControllers(mgr manager.Manager, client k8s.Client, informerFactory informers.InformerFactory, devopsClient devops.Interface, jenkinsCore core.JenkinsCore, s3Client s3.Interface, s *options.DevOpsControllerManagerOptions, stopCh <-chan struct{}) error {

	kubesphereInformer := informerFactory.KubeSphereSharedInformerFactory()

	var (
		s2iBinaryController,
		s2iRunController,
		devopsProjectController,
		devopsPipelineController,
		devopsCredentialController,
		jenkinsConfigController manager.Runnable
	)

	if devopsClient != nil {
		s2iBinaryController = s2ibinary.NewController(client.Kubernetes(),
			client.KubeSphere(),
			kubesphereInformer.Devops().V1alpha1().S2iBinaries(),
			s3Client,
		)

		s2iRunController = s2irun.NewS2iRunController(client.Kubernetes(),
			client.KubeSphere(),
			kubesphereInformer.Devops().V1alpha1().S2iBinaries(),
			kubesphereInformer.Devops().V1alpha1().S2iRuns())

		devopsProjectController = devopsproject.NewController(client.Kubernetes(),
			client.KubeSphere(), devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubeSphereSharedInformerFactory().Devops().V1alpha3().DevOpsProjects())

		devopsPipelineController = pipeline.NewController(client.Kubernetes(),
			client.KubeSphere(), devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubeSphereSharedInformerFactory().Devops().V1alpha3().Pipelines())

		devopsCredentialController = devopscredential.NewController(client.Kubernetes(),
			devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Secrets())

		jenkinsConfigController = config.NewController(&config.ControllerOptions{
			LimitRangeClient:    client.Kubernetes().CoreV1(),
			ResourceQuotaClient: client.Kubernetes().CoreV1(),
			ConfigMapClient:     client.Kubernetes().CoreV1(),

			ConfigMapInformer: informerFactory.KubernetesSharedInformerFactory().Core().V1().ConfigMaps(),
			NamespaceInformer: informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			InformerFactory:   informerFactory,

			ConfigOperator:  devopsClient,
			ReloadCasCDelay: s.JenkinsOptions.ReloadCasCDelay,
		}, s.JenkinsOptions)

		// add PipelineRun controller
		if err := (&pipelinerun.Reconciler{
			Client:       mgr.GetClient(),
			Scheme:       mgr.GetScheme(),
			DevOpsClient: devopsClient,
			JenkinsCore:  jenkinsCore,
		}).SetupWithManager(mgr); err != nil {
			klog.Errorf("unable to create pipelinerun-controller, err: %v", err)
			return err
		}

		// add PipelineRun Synchronizer
		if err := (&pipelinerun.SyncReconciler{
			Client:      mgr.GetClient(),
			JenkinsCore: jenkinsCore,
		}).SetupWithManager(mgr); err != nil {
			klog.Errorf("unable to create pipelinerun-synchronizer, err: %v", err)
			return err
		}
	}

	controllers := map[string]manager.Runnable{
		"s2ibinary-controller": s2iBinaryController,
		"s2irun-controller":    s2iRunController,
	}

	if devopsClient != nil {
		controllers["pipeline-controller"] = devopsPipelineController
		controllers["devopsprojects-controller"] = devopsProjectController
		controllers["devopscredential-controller"] = devopsCredentialController
		controllers["jenkinsconfig-controller"] = jenkinsConfigController
	}

	// Add all controllers into manager.
	for name, ctrl := range controllers {
		if ctrl == nil {
			klog.V(4).Infof("%s is not going to run due to dependent component disabled.", name)
			continue
		}

		if err := mgr.Add(ctrl); err != nil {
			klog.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}
	return nil
}
