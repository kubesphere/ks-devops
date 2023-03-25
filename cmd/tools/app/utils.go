/*
Copyright 2023 The KubeSphere Authors.

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
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
)

// patchKubeSphereConfig patches patch map into KubeSphereConfig map.
// Refer to https://github.com/evanphx/json-patch#create-and-apply-a-merge-patch.
func patchKubeSphereConfig(kubeSphereConfig map[string]interface{}, patch map[string]interface{}) (map[string]interface{}, error) {
	kubeSphereConfigBytes, err := json.Marshal(kubeSphereConfig)
	if err != nil {
		return nil, err
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	mergedBytes, err := jsonpatch.MergePatch(kubeSphereConfigBytes, patchBytes)
	if err != nil {
		return nil, err
	}
	mergedMap := make(map[string]interface{})
	if err := json.Unmarshal(mergedBytes, &mergedMap); err != nil {
		return nil, err
	}
	return mergedMap, nil
}
