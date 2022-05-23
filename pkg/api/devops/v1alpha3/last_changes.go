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

package v1alpha3

import "encoding/json"

// +kubebuilder:object:generate=false

// LastChanges represents a set of last SCM changes
type LastChanges map[string]string

func GetLastChanges(jsonText string) (lastChange LastChanges, err error) {
	lastChange = map[string]string{}
	err = json.Unmarshal([]byte(jsonText), lastChange)
	return
}

func (l LastChanges) Update(ref, hash string) LastChanges {
	l[ref] = hash
	return l
}

func (l LastChanges) LastHash(ref string) (hash string) {
	return l[ref]
}

func (l LastChanges) String() string {
	data, _ := json.Marshal(l)
	return string(data)
}
