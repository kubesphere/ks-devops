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

package v1alpha3

// ListHandler is the interface to create comparator, filter and transform
type ListHandler interface {
	Comparator() CompareFunc
	Filter() FilterFunc
	Transformer() TransformFunc
}

// defaultListHandler implements default comparator, filter and transformer.
type defaultListHandler struct {
}

func (d defaultListHandler) Comparator() CompareFunc {
	return DefaultCompare()
}

func (d defaultListHandler) Filter() FilterFunc {
	return DefaultFilter()
}

func (d defaultListHandler) Transformer() TransformFunc {
	return NoTransformFunc()
}
