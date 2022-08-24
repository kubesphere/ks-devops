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
	"net/http"
	"strconv"
)

type ValidateJenkinsfileResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result string                   `json:"result"`
		Errors []map[string]interface{} `json:"errors"`
	} `json:"data"`
}
type ValidatePipelineJsonResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result string                   `json:"result"`
		Errors []map[string]interface{} `json:"errors"`
	}
}

func (j *Jenkins) ValidateJenkinsfile(jenkinsfile string) (*ValidateJenkinsfileResponse, error) {
	responseStrut := &ValidateJenkinsfileResponse{}
	query := map[string]string{
		"jenkinsfile": jenkinsfile,
	}
	response, err := j.Requester.PostForm("/pipeline-model-converter/validateJenkinsfile", nil, responseStrut, query)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseStrut, nil

}

func (j *Jenkins) ValidatePipelineJson(json string) (*ValidatePipelineJsonResponse, error) {

	responseStruct := &ValidatePipelineJsonResponse{}
	query := map[string]string{
		"json": json,
	}
	response, err := j.Requester.PostForm("/pipeline-model-converter/validateJson", nil, responseStruct, query)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseStruct, nil
}
