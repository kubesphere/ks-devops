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

package fake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	devopsv1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

func TestCredential(t *testing.T) {
	secret := &v1.Secret{}
	secret.SetName("secret")
	secret.SetNamespace("project")

	secret1 := secret.DeepCopy()
	secret1.SetName("secret1")

	client := NewWithCredentials("project", secret.DeepCopy())

	// find credential by a correct name
	cre, err := client.GetCredentialInProject("project", "secret")
	assert.NotNil(t, cre)
	assert.Nil(t, err)

	// not found
	cre, err = client.GetCredentialInProject("fake", "fake")
	assert.NotNil(t, err)
	assert.Nil(t, cre)

	// create a secert that already exist
	var id string
	id, err = client.CreateCredentialInProject("project", secret.DeepCopy())
	assert.NotNil(t, err)
	assert.Empty(t, id)

	// delete an existing
	id, err = client.DeleteCredentialInProject("project", "secret")
	assert.Nil(t, err)
	assert.Equal(t, "", id)

	// delete a non-exsting
	id, err = client.DeleteCredentialInProject("fake", "fake")
	assert.NotNil(t, err)
	assert.Empty(t, id)

	// create it once it was deleted
	id, err = client.CreateCredentialInProject("project", secret.DeepCopy())
	assert.Nil(t, err)
	assert.Equal(t, secret.DeepCopy().GetName(), id)

	// update
	id, err = client.UpdateCredentialInProject("project", secret.DeepCopy())
	assert.Nil(t, err)
	assert.Equal(t, secret.DeepCopy().GetName(), id)

	// update a non-existing
	id, err = client.UpdateCredentialInProject("project", secret1.DeepCopy())
	assert.NotNil(t, err)
	assert.Empty(t, id)
}

func TestPipeline(t *testing.T) {
	pip := &devopsv1alpha3.Pipeline{}
	pip.SetName("pip")
	pip.SetNamespace("project")

	pip1 := pip.DeepCopy()
	pip1.SetName("pip1")

	client := NewWithPipelines("project", pip.DeepCopy())

	var id string
	var err error

	// create a non-exsiting project
	id, err = client.CreateDevOpsProject("project1")
	assert.Nil(t, err)
	assert.Equal(t, "project1", id)

	// create an exsiting project
	id, err = client.CreateDevOpsProject("project")
	assert.Nil(t, err)
	assert.Equal(t, "project", id)

	// delete a non-exsiting project
	assert.NotNil(t, client.DeleteDevOpsProject("project2"))

	// delete an exsiting project
	assert.Nil(t, client.DeleteDevOpsProject("project1"))

	// get an exsiting project
	id, err = client.GetDevOpsProject("project")
	assert.Nil(t, err)
	assert.Equal(t, "project", id)

	// get a non-existing project
	id, err = client.GetDevOpsProject("project4")
	assert.NotNil(t, err)
	assert.Empty(t, id)

	// create an existing pipeline
	id, err = client.CreateProjectPipeline("project", pip.DeepCopy())
	assert.NotNil(t, err)
	assert.Empty(t, id)

	// create a non-existing pipeline
	id, err = client.CreateProjectPipeline("project", pip1.DeepCopy())
	assert.Nil(t, err)
	assert.Empty(t, id)

	// delete a non-existing pipeline
	id, err = client.DeleteProjectPipeline("project", "pip2")
	assert.NotNil(t, err)
	assert.Empty(t, id)

	// delete an existing pipeline
	id, err = client.DeleteProjectPipeline("project", "pip1")
	assert.Nil(t, err)
	assert.Empty(t, id)

	// update an existing pipeline
	id, err = client.UpdateProjectPipeline("project", pip.DeepCopy())
	assert.Nil(t, err)
	assert.Empty(t, id)

	// update a non-existing pipeline
	id, err = client.UpdateProjectPipeline("project", pip1.DeepCopy())
	assert.NotNil(t, err)
	assert.Empty(t, id)

	// get an existing pipeline
	var obj interface{}
	obj, err = client.GetProjectPipelineConfig("project", "pip")
	assert.Nil(t, err)
	assert.NotNil(t, obj)

	// get a non-existing pipeline
	obj, err = client.GetProjectPipelineConfig("project", "pip3")
	assert.NotNil(t, err)
	assert.Nil(t, obj)
}

func TestNotImplement(t *testing.T) {
	client := New("fake")
	assert.NotNil(t, client)
	client = NewFakeDevops(nil)
	assert.Nil(t, client.ReloadConfiguration())
	assert.Nil(t, client.ApplyNewSource("fake"))

	var (
		o1 interface{}
		o2 interface{}
		o3 interface{}
	)

	o1, o2 = client.GetProjectPipelineBuildByType("", "", "")
	assertNils(t, o1, o2)
	o1, o2 = client.GetMultiBranchPipelineBuildByType("", "", "", "")
	assertNils(t, o1, o2)
	o1, o2 = client.CheckCron("", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.CheckScriptCompile("", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetPipeline("", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.ListPipelines(nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetPipelineRun("", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.StopPipeline("", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.ReplayPipeline("", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetArtifacts("", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.DownloadArtifact("", "", "", "", false, "")
	assertNils(t, o1, o2)
	o1, o2 = client.GetRunLog("", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2, o3 = client.GetStepLog("", "", "", "", "", nil)
	assertNils(t, o1, o2, o3)
	o1, o2 = client.RunPipeline("", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.ListPipelineRuns("", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.SubmitInputStep("", "", "", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetBranchPipeline("", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetBranchPipelineRun("", "", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.StopBranchPipeline("", "", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.ReplayBranchPipeline("", "", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.RunBranchPipeline("", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetBranchArtifacts("", "", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetBranchRunLog("", "", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2, o3 = client.GetBranchStepLog("", "", "", "", "", "", nil)
	assertNils(t, o1, o2, o3)
	o1, o2 = client.SubmitBranchInputStep("", "", "", "", "", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetPipelineBranch("", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.ScanBranch("", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetConsoleLog("", "", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetCrumb(nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetSCMServers("", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetSCMOrg("", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.CreateSCMServers("", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.Validate("", nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GetNotifyCommit(nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GithubWebhook(nil)
	assertNils(t, o1, o2)
	o1, o2 = client.GenericWebhook(nil)
	assertNils(t, o1, o2)
}

func assertNils(t *testing.T, obj ...interface{}) {
	for _, item := range obj {
		assert.Nil(t, item)
	}
}
