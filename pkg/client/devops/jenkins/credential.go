/*
Copyright 2018 The KubeSphere Authors.
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

package jenkins

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	v1 "k8s.io/api/core/v1"

	"github.com/kubesphere/ks-devops/pkg/client/devops"
)

func (j *Jenkins) GetCredentialInProject(projectId, id string) (*devops.Credential, error) {
	responseStruct := &devops.Credential{}

	domain := "_"

	response, err := j.Requester.GetJSON(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/_/credential/%s", projectId, id),
		responseStruct, map[string]string{
			"depth": "2",
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	responseStruct.Domain = domain
	return responseStruct, nil
}

func (j *Jenkins) CreateCredentialInProject(projectId string, credential *v1.Secret) (string, error) {
	// use pkg/client/devops/jclient/credential.go instead
	panic(nil)
}

func (j *Jenkins) UpdateCredentialInProject(projectId string, credential *v1.Secret) (string, error) {
	// use pkg/client/devops/jclient/credential.go instead
	panic(nil)
}

// DeleteCredentialInProject deletes credential
//
// Deprecated
func (j *Jenkins) DeleteCredentialInProject(projectId, id string) (string, error) {
	// use pkg/client/devops/jclient/credential.go instead
	panic(nil)
}
