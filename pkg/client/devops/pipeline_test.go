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

package devops

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestGetSubmitters(t *testing.T) {
	input := &Input{}
	assert.Equal(t, len(input.GetSubmitters()), 0,
		"errors happen when try to get submitters without any submitters")

	input.Submitter = ""
	assert.Equal(t, len(input.GetSubmitters()), 0, "Should be empty submitters when submitter not set")

	input.Submitter = "a , b, c,d"
	submitters := input.GetSubmitters()
	assert.Equal(t, len(submitters), 4, "get incorrect number of submitters")
	assert.DeepEqual(t, submitters, []string{"a", "b", "c", "d"})
}

func TestApprovable(t *testing.T) {
	input := &Input{}

	assert.Equal(t, input.Approvable(""), false, "should allow anyone to approve it if there's no submitter given")
	assert.Equal(t, input.Approvable("fake"), false, "should allow anyone to approve it if there's no submitter given")

	input.Submitter = "fake"
	assert.Equal(t, input.Approvable(""), false, "should not approve by nobody if there's a particular submitter")
	assert.Equal(t, input.Approvable("rick"), false, "should not approve by who is not the specific one")
	assert.Equal(t, input.Approvable("fake"), true, "should be approvable")

	input.Submitter = "fake, good ,bad"
	assert.Equal(t, input.Approvable("fake"), true, "should be approvable")
	assert.Equal(t, input.Approvable("good"), true, "should be approvable")
	assert.Equal(t, input.Approvable("bad"), true, "should be approvable")
}

func TestPipelineJsonMarshall(t *testing.T) {
	const name = "fakeName"
	var err error
	var pipelineText string
	var pipelienList *PipelineList

	pipelineText = fmt.Sprintf(`[{"displayName":"%s", "weatherScore": 11}]`, name)
	pipelienList, err = UnmarshalPipeline(1, []byte(pipelineText))
	assert.NilError(t, err, "pipeline json marshal should be success")
	assert.Equal(t, pipelienList.Total, 1)
	assert.Equal(t, len(pipelienList.Items), 1)
	assert.Equal(t, pipelienList.Items[0].DisplayName, name)
	assert.Equal(t, pipelienList.Items[0].WeatherScore, 11)

	// test against the default value of weatherScore, it should be 100
	pipelineText = fmt.Sprintf(`[{"displayName":"%s"}]`, name)
	pipelienList, err = UnmarshalPipeline(1, []byte(pipelineText))
	assert.NilError(t, err, "pipeline json marshal should be success")
	assert.Equal(t, pipelienList.Total, 1)
	assert.Equal(t, len(pipelienList.Items), 1)
	assert.Equal(t, pipelienList.Items[0].DisplayName, name)
	assert.Equal(t, pipelienList.Items[0].WeatherScore, 100)

	// test against multiple items
	pipelineText = fmt.Sprintf(`[{"displayName":"%s"}, {"displayName":"%s-1"}]`, name, name)
	pipelienList, err = UnmarshalPipeline(2, []byte(pipelineText))
	assert.NilError(t, err, "pipeline json marshal should be success")
	assert.Equal(t, pipelienList.Total, 2)
	assert.Equal(t, len(pipelienList.Items), 2)
	assert.Equal(t, pipelienList.Items[0].DisplayName, name)
	assert.Equal(t, pipelienList.Items[0].WeatherScore, 100)
	assert.Equal(t, pipelienList.Items[1].DisplayName, fmt.Sprintf("%s-1", name))
	assert.Equal(t, pipelienList.Items[1].WeatherScore, 100)
}
