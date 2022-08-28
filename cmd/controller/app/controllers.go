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
	"kubesphere.io/devops/controllers/addon"
	"kubesphere.io/devops/controllers/argocd"
	"kubesphere.io/devops/controllers/fluxcd"
	"kubesphere.io/devops/controllers/gitrepository"
	"kubesphere.io/devops/controllers/jenkins/devopscredential"
	"kubesphere.io/devops/controllers/jenkins/devopsproject"
	"kubesphere.io/devops/pkg/jwt/token"
	"kubesphere.io/devops/pkg/server/errors"

	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"k8s.io/klog/v2"
	"kubesphere.io/devops/cmd/controller/app/options"
	"kubesphere.io/devops/controllers/jenkins/config"
	jenkinspipeline "kubesphere.io/devops/controllers/jenkins/pipeline"
	"kubesphere.io/devops/controllers/jenkins/pipelinerun"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/informers"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func addControllers(mgr manager.Manager, client k8s.Client, informerFactory informers.InformerFactory,
	devopsClient devops.Interface, jenkinsCore core.JenkinsCore,
	s *options.DevOpsControllerManagerOptions) error {
	if devopsClient == nil {
		return errors.New("devopsClient should not be nil")
	}

	reconcilers := getAllControllers(mgr, client, informerFactory, devopsClient, s, jenkinsCore)
	reconcilers["pipeline"] = func(mgr manager.Manager) (err error) {
		tokenIssuer := token.NewTokenIssuer(s.JWTOptions.Secret, s.JWTOptions.MaximumClockSkew)
		// add PipelineRun controller
		if err = (&pipelinerun.Reconciler{
			Client:       mgr.GetClient(),
			Scheme:       mgr.GetScheme(),
			DevOpsClient: devopsClient,
			JenkinsCore:  jenkinsCore,
			TokenIssuer:  tokenIssuer,
		}).SetupWithManager(mgr); err != nil {
			klog.Errorf("unable to create pipelinerun-controller, err: %v", err)
			return
		}

		// add PipelineRun Synchronizer
		if err = (&pipelinerun.SyncReconciler{
			Client:      mgr.GetClient(),
			JenkinsCore: jenkinsCore,
		}).SetupWithManager(mgr); err != nil {
			klog.Errorf("unable to create pipelinerun-synchronizer, err: %v", err)
			return
		}

		// add Pipeline metadata controller
		err = (&jenkinspipeline.Reconciler{
			Client:      mgr.GetClient(),
			JenkinsCore: jenkinsCore,
		}).SetupWithManager(mgr)
		return
	}

	// Add all controllers into manager.
	for name, ok := range s.FeatureOptions.GetControllers() {
		ctrl := reconcilers[name]
		if ctrl == nil || !ok {
			klog.V(4).Infof("%s is not going to run due to dependent component disabled.", name)
			continue
		}

		if err := ctrl(mgr); err != nil {
			klog.Error(err, "add controller to manager failed ", name)
			return err
		}
	}
	return nil
}

func getAllControllers(mgr manager.Manager, client k8s.Client, informerFactory informers.InformerFactory,
	devopsClient devops.Interface, s *options.DevOpsControllerManagerOptions, jenkinsCore core.JenkinsCore) map[string]func(mgr manager.Manager) error {

	argocdReconciler := &argocd.Reconciler{
		Client:        mgr.GetClient(),
		ArgoNamespace: s.ArgoCDOption.Namespace,
	}
	argocdAppReconciler := &argocd.ApplicationReconciler{
		Client: mgr.GetClient(),
	}
	argocdClusterReconciler := &argocd.MultiClusterReconciler{
		Client: mgr.GetClient(),
	}
	argocdAppStatusReconciler := &argocd.ApplicationStatusReconciler{
		Client: mgr.GetClient(),
	}
	argocdGitRepoReconciler := &argocd.GitRepositoryController{
		Client:        mgr.GetClient(),
		ArgoNamespace: s.ArgoCDOption.Namespace,
	}
	argcdImageUpdaterReconciler := &argocd.ImageUpdaterReconciler{
		Client: mgr.GetClient(),
	}
	gitRepoReconcilers := gitrepository.GetReconcilers(mgr.GetClient())

	fluxcdGitRepoReconciler := &fluxcd.GitRepositoryReconciler{
		Client: mgr.GetClient(),
	}
	tokenIssuer := token.NewTokenIssuer(s.JWTOptions.Secret, s.JWTOptions.MaximumClockSkew)
	jenkinsAgentLabelsReconciler := config.AgentLabelsReconciler{
		Client:          mgr.GetClient(),
		TargetNamespace: s.FeatureOptions.SystemNamespace,
		TokenIssuer:     tokenIssuer,
		JenkinsClient:   jenkinsCore,
	}
	jenkinsPodTemplate := config.PodTemplateReconciler{
		Client:                   mgr.GetClient(),
		TargetConfigMapNamespace: s.FeatureOptions.SystemNamespace,
	}
	fluxcdApplicationReconciler := &fluxcd.ApplicationReconciler{
		Client: mgr.GetClient(),
	}
	fluxcdMultiClusterReconciler := &fluxcd.MultiClusterReconciler{
		Client: mgr.GetClient(),
	}

	return map[string]func(mgr manager.Manager) error{
		gitRepoReconcilers.GetName(): func(mgr manager.Manager) error {
			err := (&gitrepository.PullRequestStatusReconciler{
				Client:          mgr.GetClient(),
				ExternalAddress: s.FeatureOptions.ExternalAddress,
			}).SetupWithManager(mgr)
			if err != nil {
				return err
			}
			return gitRepoReconcilers.SetupWithManager(mgr)
		},
		"addon": func(mgr manager.Manager) error {
			err := (&addon.OperatorCRDReconciler{
				Client: mgr.GetClient(),
			}).SetupWithManager(mgr)
			if err == nil {
				err = (&addon.Reconciler{
					Client: mgr.GetClient(),
				}).SetupWithManager(mgr)
			}
			return err
		},
		"jenkinsagent": func(mgr manager.Manager) error {
			return jenkinsPodTemplate.SetupWithManager(mgr)
		},
		"jenkinsconfig": func(mgr manager.Manager) error {
			return mgr.Add(config.NewController(&config.ControllerOptions{
				LimitRangeClient:    client.Kubernetes().CoreV1(),
				ResourceQuotaClient: client.Kubernetes().CoreV1(),
				ConfigMapClient:     client.Kubernetes().CoreV1(),

				ConfigMapInformer: informerFactory.KubernetesSharedInformerFactory().Core().V1().ConfigMaps(),
				NamespaceInformer: informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
				InformerFactory:   informerFactory,

				ConfigOperator:  devopsClient,
				ReloadCasCDelay: s.JenkinsOptions.ReloadCasCDelay,
			}, s.JenkinsOptions))
		},
		"jenkins": func(mgr manager.Manager) error {
			err := mgr.Add(devopscredential.NewController(client.Kubernetes(),
				devopsClient,
				informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
				informerFactory.KubernetesSharedInformerFactory().Core().V1().Secrets()))
			if err == nil {
				err = mgr.Add(devopsproject.NewController(client.Kubernetes(),
					client.KubeSphere(), devopsClient,
					informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
					informerFactory.KubeSphereSharedInformerFactory().Devops().V1alpha3().DevOpsProjects()))
			}
			if err == nil {
				err = mgr.Add(jenkinspipeline.NewController(client.Kubernetes(),
					client.KubeSphere(), devopsClient,
					informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
					informerFactory.KubeSphereSharedInformerFactory().Devops().V1alpha3().Pipelines()))
			}

			if err == nil {
				jenkinsfileReconciler := &jenkinspipeline.JenkinsfileReconciler{
					Client:      mgr.GetClient(),
					TokenIssuer: tokenIssuer,
					JenkinsCore: jenkinsCore,
				}
				err = jenkinsfileReconciler.SetupWithManager(mgr)
			}
			if err == nil {
				err = jenkinsAgentLabelsReconciler.SetupWithManager(mgr)
			}
			return err
		},
		argocdReconciler.GetGroupName(): func(mgr manager.Manager) (err error) {
			if err = argocdReconciler.SetupWithManager(mgr); err != nil {
				return
			}
			if err = argocdClusterReconciler.SetupWithManager(mgr); err != nil {
				return
			}
			if err = argocdAppStatusReconciler.SetupWithManager(mgr); err != nil {
				return
			}
			if err = argocdGitRepoReconciler.SetupWithManager(mgr); err != nil {
				return
			}
			return argocdAppReconciler.SetupWithManager(mgr)
		},
		argcdImageUpdaterReconciler.GetGroupName() + "-image-updater": func(mgr manager.Manager) error {
			return argcdImageUpdaterReconciler.SetupWithManager(mgr)
		},
		fluxcdApplicationReconciler.GetGroupName(): func(mgr manager.Manager) (err error) {
			if err = fluxcdGitRepoReconciler.SetupWithManager(mgr); err != nil {
				return
			}
			if err = fluxcdMultiClusterReconciler.SetupWithManager(mgr); err != nil {
				return
			}
			return fluxcdApplicationReconciler.SetupWithManager(mgr)
		},
	}
}
