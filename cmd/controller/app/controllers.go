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
	"k8s.io/klog"
	"kubesphere.io/devops/controllers/devopscredential"
	"kubesphere.io/devops/controllers/devopsproject"
	"kubesphere.io/devops/controllers/pipeline"
	"kubesphere.io/devops/controllers/s2ibinary"
	"kubesphere.io/devops/controllers/s2irun"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/s3"
	"kubesphere.io/devops/pkg/informers"
	"kubesphere.io/devops/pkg/k8s"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func addControllers(
	mgr manager.Manager,
	client k8s.Client,
	informerFactory informers.InformerFactory,
	devopsClient devops.Interface,
	s3Client s3.Interface,
	options *k8s.KubernetesOptions,
	stopCh <-chan struct{}) error {

	kubesphereInformer := informerFactory.KubeSphereSharedInformerFactory()
	//kubernetesInformer := informerFactory.KubernetesSharedInformerFactory()

	var vsController, drController manager.Runnable
	var s2iBinaryController, s2iRunController, devopsProjectController, devopsPipelineController, devopsCredentialController manager.Runnable
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
			client.KubeSphere(),
			devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubeSphereSharedInformerFactory().Devops().V1alpha3().Pipelines())

		devopsCredentialController = devopscredential.NewController(client.Kubernetes(),
			devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Secrets())

	}

	controllers := map[string]manager.Runnable{
		"virtualservice-controller":  vsController,
		"destinationrule-controller": drController,
		"s2ibinary-controller":       s2iBinaryController,
		"s2irun-controller":          s2iRunController,
	}

	if devopsClient != nil {
		controllers["pipeline-controller"] = devopsPipelineController
		controllers["devopsprojects-controller"] = devopsProjectController
		controllers["devopscredential-controller"] = devopsCredentialController
	}

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
