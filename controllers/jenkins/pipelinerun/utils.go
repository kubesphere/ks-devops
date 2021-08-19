package pipelinerun

import (
	"encoding/json"
	"errors"
	"io"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	jenkinsTimeLayout = "2006-01-02T15:04:05+0000"
	staticDevOpsUrl   = "https://devops.kubesphere.io/"
)

func parseJenkinsTime(jenkinsTime string) (time.Time, error) {
	return time.Parse(jenkinsTimeLayout, jenkinsTime)
}

// mockClientURL is only for HttpParameters.
// Generated URL has no practical significance, but it is indispensable.
func mockClientURL() *url.URL {
	stubURL, err := url.Parse(staticDevOpsUrl)
	if err != nil {
		// should never happen
		panic("invalid stub URL: " + staticDevOpsUrl)
	}
	return stubURL
}

func buildHTTPParametersForRunning(prSpec *devopsv1alpha4.PipelineRunSpec) (*devopsClient.HttpParameters, error) {
	if prSpec == nil {
		return nil, errors.New("invalid PipelineRun")
	}
	parameters := prSpec.Parameters
	if parameters == nil {
		parameters = make([]devopsv1alpha4.Parameter, 0)
	}

	// build http parameters
	var body = map[string]interface{}{
		"parameters": parameters,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &devopsClient.HttpParameters{
		Url:    mockClientURL(),
		Method: http.MethodPost,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(string(bodyJSON))),
	}, nil
}

type JenkinsRunState string

const (
	Queued        JenkinsRunState = "QUEUED"
	Running       JenkinsRunState = "RUNNING"
	Paused        JenkinsRunState = "PAUSED"
	Skipped       JenkinsRunState = "SKIPPED"
	NotBuiltState JenkinsRunState = "NOT_BUILT"
	Finished      JenkinsRunState = "FINISHED"
)

func (state JenkinsRunState) String() string {
	return string(state)
}

type JenkinsRunResult string

const (
	Success        JenkinsRunResult = "SUCCESS"
	Unstable       JenkinsRunResult = "UNSTABLE"
	Failure        JenkinsRunResult = "FAILURE"
	NotBuiltResult JenkinsRunResult = "NOT_BUILT"
	Unknown        JenkinsRunResult = "UNKNOWN"
	Aborted        JenkinsRunResult = "ABORTED"
)

func (result JenkinsRunResult) String() string {
	return string(result)
}
