// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package argocd

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// InterfaceToMap converts value to map, which type of key is string and type of value is interface{} (Actually it is
// map[string]interface{}).
// The type of value must be struct. Any other types will lead an error.
func InterfaceToMap(value interface{}) (map[string]interface{}, error) {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	uValue := map[string]interface{}{}
	if err := json.Unmarshal(valueBytes, &uValue); err != nil {
		return nil, err
	}
	return uValue, nil
}

// SetNestedField sets nested field into the object with map type. The type of field value must be struct. Any other
// types will lead to an error.
func SetNestedField(obj map[string]interface{}, value interface{}, fields ...string) error {
	// convert value to unstructured
	mapValue, err := InterfaceToMap(value)
	if err != nil {
		return err
	}
	return unstructured.SetNestedField(obj, mapValue, fields...)
}
