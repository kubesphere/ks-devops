/*
Copyright 2024 The KubeSphere Authors.

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
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/client/k8s"
)

func NewK8sClient(kubeconfig string) (k8s.Client, error) {
	k8sOption := k8s.NewKubernetesOptions()
	if kubeconfig != "" {
		k8sOption.KubeConfig = kubeconfig
	}
	return k8s.NewKubernetesClient(k8sOption)
}

func NewRuntimeClient(kubeconfig string) (runtimeclient.WithWatch, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return runtimeclient.NewWithWatch(restConfig, runtimeclient.Options{})
}

func getConfigmapWithWatch(ctx context.Context, client runtimeclient.WithWatch, namespace, name string) (*corev1.ConfigMap, error) {
	configmaps := &corev1.ConfigMapList{}
	timerCtx, cancelFunc := context.WithTimeout(ctx, 10*time.Minute)
	defer cancelFunc()

	watcher, err := client.Watch(timerCtx, configmaps, runtimeclient.InNamespace(namespace))
	if err != nil {
		klog.Errorf("failed to watch configmaps in namespace %s", namespace)
		return nil, err
	}

	var target *corev1.ConfigMap
	defer watcher.Stop()
	for {
		select {
		case <-timerCtx.Done():
			return nil, timerCtx.Err()
		case event := <-watcher.ResultChan():
			klog.V(4).Infof("watch event type: %s", event.Type)
			switch event.Type {
			case "", watch.Added:
				target, _ = event.Object.(*corev1.ConfigMap)
				if target.Name == name {
					return target, nil
				}
			}
		}
	}
}
