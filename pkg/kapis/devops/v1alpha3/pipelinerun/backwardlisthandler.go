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

package pipelinerun

import (
	"context"

	"github.com/kubesphere/ks-devops/pkg/api"
	cmstore "github.com/kubesphere/ks-devops/pkg/store/configmap"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type backwardListHandler struct {
	client client.Client
}

func (b backwardListHandler) Comparator() resourcesv1alpha3.CompareFunc {
	return api.DefaultCompareFunc
}

func (b backwardListHandler) Filter() resourcesv1alpha3.FilterFunc {
	return func(obj runtime.Object, filter query.Filter) bool {
		ok := b.backwardFilter(obj)
		if !ok {
			return false
		}
		return api.DefaultFilterFunc(obj, filter)
	}
}

func (b backwardListHandler) Transformer() resourcesv1alpha3.TransformFunc {
	return api.NoTransformFunc
}

func checkPipelineRun(object runtime.Object) (*v1alpha3.PipelineRun, bool) {
	pr, ok := object.(*v1alpha3.PipelineRun)
	if !ok || pr == nil {
		return nil, false
	}
	return pr, true
}

func (b backwardListHandler) backwardFilter(object runtime.Object) bool {
	if pr, valid := checkPipelineRun(object); valid {
		statusJSON, ok := pr.Annotations[v1alpha3.JenkinsPipelineRunStatusAnnoKey]
		if !ok {
			if pipelineRunStore, err := cmstore.NewConfigMapStore(context.Background(), types.NamespacedName{
				Namespace: pr.Namespace,
				Name:      pr.Name,
			}, b.client); err == nil {
				statusJSON = pipelineRunStore.GetStatus()
			} else {
				klog.Error(err, "failed to get status from configmap store")
			}
		}

		return statusJSON != ""
	}
	return false
}
