/*
Copyright 2020 KubeSphere Authors

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

package v1alpha3

import (
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/apiserver/query"
)

type Interface interface {
	// Get retrieves a single object by its namespace and name
	Get(namespace, name string) (runtime.Object, error)

	// List retrieves a collection of objects matches given query
	List(namespace string, query *query.Query) (*api.ListResult, error)
}

// CompareFunc return true is left great than right
type CompareFunc func(runtime.Object, runtime.Object, query.Field) bool

type FilterFunc func(runtime.Object, query.Filter) bool

type TransformFunc func(runtime.Object) runtime.Object

func DefaultList(objects []runtime.Object, q *query.Query, compareFunc CompareFunc, filterFunc FilterFunc, transformFuncs ...TransformFunc) *api.ListResult {
	// selected matched ones
	var filtered []runtime.Object
	for _, object := range objects {
		selected := true
		for field, value := range q.Filters {
			if filterFunc != nil && !filterFunc(object, query.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}

		if selected {
			for _, transform := range transformFuncs {
				if transform != nil {
					object = transform(object)
				}
			}
			filtered = append(filtered, object)
		}
	}

	// sort by sortBy field
	if compareFunc != nil {
		sort.Slice(filtered, func(i, j int) bool {
			if !q.Ascending {
				return compareFunc(filtered[i], filtered[j], q.SortBy)
			}
			return !compareFunc(filtered[i], filtered[j], q.SortBy)
		})
	}

	total := len(filtered)

	if q.Pagination == nil {
		q.Pagination = query.NoPagination
	}

	start, end := q.Pagination.GetValidPagination(total)

	return &api.ListResult{
		TotalItems: len(filtered),
		Items:      objectsToInterfaces(filtered[start:end]),
	}
}

// DefaultObjectMetaCompare return true is left great than right
func DefaultObjectMetaCompare(left, right metav1.ObjectMeta, sortBy query.Field) bool {
	switch sortBy {
	// ?sortBy=name
	case query.FieldName:
		// sort the name in ascending order
		return strings.Compare(left.Name, right.Name) > 0
	//	?sortBy=creationTimestamp
	default:
		fallthrough
	case query.FieldCreationTimeStamp:
		// compare by name if creation timestamp is equal
		if left.CreationTimestamp.Equal(&right.CreationTimestamp) {
			return strings.Compare(left.Name, right.Name) > 0
		}
		return left.CreationTimestamp.After(right.CreationTimestamp.Time)
	}
}

// DefaultObjectMetaFilter filters data with given filter
func DefaultObjectMetaFilter(item metav1.ObjectMeta, filter query.Filter) bool {
	switch filter.Field {
	case query.FieldNames:
		for _, name := range strings.Split(string(filter.Value), ",") {
			if item.Name == name {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&name=default
	case query.FieldName:
		return strings.Contains(item.Name, string(filter.Value))
		// /namespaces?page=1&limit=10&uid=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldUID:
		return string(item.UID) == string(filter.Value)
		// /deployments?page=1&limit=10&namespace=kubesphere-system
	case query.FieldNamespace:
		return item.Namespace == string(filter.Value)
		// /namespaces?page=1&limit=10&ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldOwnerReference:
		for _, ownerReference := range item.OwnerReferences {
			if string(ownerReference.UID) == string(filter.Value) {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&ownerKind=Workspace
	case query.FieldOwnerKind:
		for _, ownerReference := range item.OwnerReferences {
			if ownerReference.Kind == string(filter.Value) {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&annotation=openpitrix_runtime
	case query.FieldAnnotation:
		return labelsMatch(item.Annotations, string(filter.Value))
		// /namespaces?page=1&limit=10&label=kubesphere.io/workspace=system-workspace
	case query.FieldLabel:
		return labelsMatch(item.Labels, string(filter.Value))
	default:
		// We should allow fields that are not found
		return true
	}
}

// labelsMatch handles multi-label value pairs split by ",".
// e.g. devops.ks.io/creator=admin,devops.ks.io/status=success
func labelsMatch(labels map[string]string, filterStr string) bool {
	filters := strings.Split(filterStr, ",")
	var match = true
	for _, filter := range filters {
		match = match && labelMatch(labels, strings.TrimSpace(filter))
		if !match {
			break
		}
	}
	return match
}

func labelMatch(labels map[string]string, filter string) bool {
	fields := strings.SplitN(filter, "=", 2)
	var key, value string
	var opposite bool
	if len(fields) == 2 {
		key = fields[0]
		if strings.HasSuffix(key, "!") {
			key = strings.TrimSuffix(key, "!")
			opposite = true
		}
		value = fields[1]
	} else {
		key = fields[0]
		value = "*"
	}
	for k, v := range labels {
		if opposite {
			if (k == key) && v != value {
				return true
			}
		} else {
			if (k == key) && (value == "*" || v == value) {
				return true
			}
		}
	}
	return false
}

func objectsToInterfaces(objs []runtime.Object) []interface{} {
	res := make([]interface{}, 0)
	for _, obj := range objs {
		res = append(res, obj)
	}
	return res
}
