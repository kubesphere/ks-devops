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

package devops

import (
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/apiserver/query"
	ksinformers "kubesphere.io/devops/pkg/client/informers/externalversions"
	"kubesphere.io/devops/pkg/models/resources/v1alpha3"
)

type devopsGetter struct {
	informers ksinformers.SharedInformerFactory
}

func New(ksinformer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &devopsGetter{informers: ksinformer}
}

func (n devopsGetter) Get(_, name string) (runtime.Object, error) {
	return n.informers.Devops().V1alpha3().DevOpsProjects().Lister().Get(name)
}

func (n devopsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	projects, err := n.informers.Devops().V1alpha3().DevOpsProjects().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, project := range projects {
		result = append(result, project)
	}

	return v1alpha3.DefaultList(result, query, v1alpha3.DefaultCompare(), v1alpha3.DefaultFilter()), nil
}
