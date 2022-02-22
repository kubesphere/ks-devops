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

package sliceutil

// RemoveString removes an item from a slice with a custom function
func RemoveString(slice []string, remove func(item string) bool) []string {
	for i := 0; i < len(slice); i++ {
		if remove(slice[i]) {
			slice = append(slice[:i], slice[i+1:]...)
			i--
		}
	}
	return slice
}

// SameItem returns a function to check if the item is same to
func SameItem(target string) func(item string) bool {
	return func(item string) bool {
		return target == item
	}
}

// HasString checks if there is a same string existing in a slice
func HasString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// AddToSlice adds an item to a slice without duplicated
func AddToSlice(item string, array []string) []string {
	if !HasString(array, item) {
		array = append(array, item)
	}
	return array
}
