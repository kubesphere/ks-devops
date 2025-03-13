/*
Copyright 2020 The KubeSphere Authors.

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

package api

import (
	"github.com/kubesphere/ks-devops/pkg/utils/k8sutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type ListResult struct {
	Items      []interface{} `json:"items"`
	TotalItems int           `json:"totalItems"`
}

// NewListResult creates a ListResult for the given items and total.
func NewListResult(items []interface{}, total int) *ListResult {
	if items == nil {
		items = make([]interface{}, 0)
	}
	return &ListResult{
		Items:      items,
		TotalItems: total,
	}
}

func FromKSListResult(result *api.ListResult) *ListResult {
	if result == nil {
		return nil
	}
	var items []interface{}
	for _, item := range result.Items {
		items = append(items, item)
	}
	return &ListResult{
		Items:      items,
		TotalItems: result.TotalItems,
	}
}

func DefaultCompareFunc(left runtime.Object, right runtime.Object, field query.Field) bool {
	a, err := k8sutil.ExtractObjectMeta(left)
	if err != nil {
		return false
	}
	b, err := k8sutil.ExtractObjectMeta(right)
	if err != nil {
		return false
	}
	return resourcesv1alpha3.DefaultObjectMetaCompare(*a, *b, field)
}

func DefaultFilterFunc(obj runtime.Object, filter query.Filter) bool {
	a, err := k8sutil.ExtractObjectMeta(obj)
	if err != nil {
		return false
	}
	return resourcesv1alpha3.DefaultObjectMetaFilter(*a, filter)
}

func NoTransformFunc(obj runtime.Object) runtime.Object {
	return obj
}

type ResourceQuota struct {
	Namespace string                     `json:"namespace" description:"namespace"`
	Data      corev1.ResourceQuotaStatus `json:"data" description:"resource quota status"`
}

type NamespacedResourceQuota struct {
	Namespace string `json:"namespace,omitempty"`

	Data struct {
		corev1.ResourceQuotaStatus

		// quota left status, do the math on the side, cause it's
		// a lot easier with go-client library
		Left corev1.ResourceList `json:"left,omitempty"`
	} `json:"data,omitempty"`
}

type Router struct {
	RouterType  string            `json:"type"`
	Annotations map[string]string `json:"annotations"`
}

type GitCredential struct {
	RemoteUrl string                  `json:"remoteUrl" description:"git server url"`
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty" description:"auth secret reference"`
}

type RegistryCredential struct {
	Username   string `json:"username" description:"username"`
	Password   string `json:"password" description:"password"`
	ServerHost string `json:"serverhost" description:"registry server host"`
}

type Workloads struct {
	Namespace string                 `json:"namespace" description:"the name of the namespace"`
	Count     map[string]int         `json:"data" description:"the number of unhealthy workloads"`
	Items     map[string]interface{} `json:"items,omitempty" description:"unhealthy workloads"`
}

type ClientType string

const (
	StatusOK  = "ok"
	GroupName = "devops.kubesphere.io"
)
