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

package predicate

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8spredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Filter is a reconciler filter function
type Filter func(object client.Object) (ok bool)

// NewFilterHasLabel creats a filter that contains the specific label
func NewFilterHasLabel(label string) Filter {
	return func(object client.Object) (ok bool) {
		_, ok = object.GetLabels()[label]
		return
	}
}

// NewPredicateFuncs creates a filter function
func NewPredicateFuncs(filter Filter) k8spredicate.Funcs {
	return k8spredicate.NewPredicateFuncs(filter)
}
