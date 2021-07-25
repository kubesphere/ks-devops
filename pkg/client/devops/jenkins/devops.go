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
	"kubesphere.io/devops/pkg/client/devops"
	"net/http"
)

func NewDevopsClient(options *Options) (devops.Interface, error) {
	// we have to create http client with no redirection
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	jenkins := CreateJenkins(client, options.Host, options.MaxConnections, options.Username, options.Password)

	return jenkins, nil
}
