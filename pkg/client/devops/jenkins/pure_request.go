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

package jenkins

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"k8s.io/klog"

	"kubesphere.io/devops/pkg/client/devops"
)

// TODO: deprecated, use SendJenkinsRequestWithHeaderResp() instead
func (j *Jenkins) SendPureRequest(path string, httpParameters *devops.HttpParameters) ([]byte, error) {
	resBody, _, err := j.SendPureRequestWithHeaderResp(path, httpParameters)

	return resBody, err
}

// SendPureRequestWithHeaderResp provider request header to call jenkins api.
// transfer bearer token to basic token for inner Oauth and Jenkins
func (j *Jenkins) SendPureRequestWithHeaderResp(path string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	apiURL, err := url.Parse(j.Server + path)
	if err != nil {
		klog.V(8).Info(err)
		return nil, nil, err
	}

	apiURL.RawQuery = httpParameters.Url.RawQuery
	client := &http.Client{Timeout: 30 * time.Second}

	header := httpParameters.Header.Clone()
	if header == nil {
		header = http.Header{}
	}

	if j.Requester != nil {
		auth := j.Requester.BasicAuth

		creds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", auth.Username, auth.Password)))
		header.Set("Authorization", fmt.Sprintf("Basic %s", creds))

		// we only set crumb to request header for POST method
		// See also https://github.com/jenkinsci/jenkins/blob/ca646b3742e6921d332078d0b287fd359970d7e9/core/src/main/java/hudson/security/csrf/CrumbFilter.java#L126
		if httpParameters.Method == http.MethodPost {
			if err := j.Requester.SetCrumbForConsumer(func(crumbRequestField, crumb string) {
				header.Set(crumbRequestField, crumb)
			}); err != nil {
				klog.Errorf("unable to set crumb to HTTP header, err: %v", err)
				return nil, nil, err
			}
		}
	}

	newRequest := &http.Request{
		Method:   httpParameters.Method,
		URL:      apiURL,
		Header:   header,
		Body:     httpParameters.Body,
		Form:     httpParameters.Form,
		PostForm: httpParameters.PostForm,
	}

	resp, err := client.Do(newRequest)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	resBody, _ := getRespBody(resp)
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		klog.Errorf("%+v", string(resBody))
		jkerr := new(JkError)
		jkerr.Code = resp.StatusCode
		jkerr.Message = string(resBody)
		return nil, nil, jkerr
	}

	return resBody, resp.Header, nil
}
