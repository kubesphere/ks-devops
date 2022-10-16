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

package jclient

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	jcredential "github.com/jenkins-zh/jenkins-client/pkg/credential"
	"github.com/jenkins-zh/jenkins-client/pkg/mock/mhttp"
	"github.com/jenkins-zh/jenkins-client/pkg/util"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	devopsv1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"
	devopsutil "kubesphere.io/devops/pkg/client/devops/util"

	"github.com/jenkins-zh/jenkins-client/pkg/core"
)

func TestDeleteCredentialInProject(t *testing.T) {
	var roundTripper *mhttp.MockRoundTripper
	ctrl := gomock.NewController(t)
	roundTripper = mhttp.NewMockRoundTripper(ctrl)

	client := &JenkinsClient{
		Core: core.JenkinsCore{
			URL:          "http://localhost",
			RoundTripper: roundTripper,
		},
	}

	const folder = "fake"
	const id = "id"
	jcredential.PrepareForDeleteCredentialInFolder(roundTripper, "http://localhost", "", "", folder, id)
	val, err := client.DeleteCredentialInProject(folder, id)
	assert.Nil(t, err)
	assert.Equal(t, id, val)
}

func TestUpdateCredentialInProject(t *testing.T) {
	var roundTripper *mhttp.MockRoundTripper
	ctrl := gomock.NewController(t)
	roundTripper = mhttp.NewMockRoundTripper(ctrl)

	client := &JenkinsClient{
		Core: core.JenkinsCore{
			URL:          "http://localhost",
			RoundTripper: roundTripper,
		},
	}

	secret := &v1.Secret{}
	secret.SetName("id")
	secret.Type = devopsv1alpha3.SecretTypeBasicAuth

	data, err := devopsutil.ConvertSecretToCredential(secret.DeepCopy())
	assert.Nil(t, err)

	formData := url.Values{}
	formData.Add("json", util.TOJSON(data))

	const folder = "fake"
	const id = "id"
	jcredential.PrepareForUpdateCredentialInFolder(roundTripper, "http://localhost", "", "",
		folder, id, strings.NewReader(formData.Encode()))
	val, err := client.UpdateCredentialInProject(folder, secret.DeepCopy())
	assert.Nil(t, err)
	assert.Empty(t, val)
}

func TestCreateCredentialInProject(t *testing.T) {
	var roundTripper *mhttp.MockRoundTripper
	ctrl := gomock.NewController(t)
	roundTripper = mhttp.NewMockRoundTripper(ctrl)

	client := &JenkinsClient{
		Core: core.JenkinsCore{
			URL:          "http://localhost",
			RoundTripper: roundTripper,
		},
	}

	secret := &v1.Secret{}
	secret.SetName("id")
	secret.Type = devopsv1alpha3.SecretTypeBasicAuth
	secret.Data = map[string][]byte{
		devopsv1alpha3.BasicAuthUsernameKey: []byte("username"),
		devopsv1alpha3.BasicAuthPasswordKey: []byte("password"),
	}

	unknownSecret := secret.DeepCopy()
	unknownSecret.Type = "fake"

	data, err := devopsutil.ConvertSecretToCredential(secret.DeepCopy())
	assert.Nil(t, err)

	formData := url.Values{}
	formData.Add("json", fmt.Sprintf(`{"credentials": %s}`, util.TOJSON(data)))

	const folder = "fake"
	jcredential.PrepareForCreateCredentialInFolder(roundTripper, "http://localhost", "", "",
		folder, strings.NewReader(formData.Encode()))
	val, err := client.CreateCredentialInProject(folder, secret.DeepCopy())
	assert.Nil(t, err)
	assert.Empty(t, val)

	// create with an unknown secret
	_, err = client.CreateCredentialInProject(folder, unknownSecret.DeepCopy())
	assert.NotNil(t, err)
}
