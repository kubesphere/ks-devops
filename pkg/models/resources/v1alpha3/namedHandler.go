/*
Copyright 2022 KubeSphere Authors

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

// NamedHandler represents a handler that sort the list by name
type NamedHandler struct {
}

// Comparator compare the list by name with ascending
func (h NamedHandler) Comparator() CompareFunc {
	return NameCompare()
}

// Filter is the default filter
func (h NamedHandler) Filter() FilterFunc {
	return DefaultFilter()
}

// Transformer is the default transformer
func (h NamedHandler) Transformer() TransformFunc {
	return NoTransformFunc()
}
