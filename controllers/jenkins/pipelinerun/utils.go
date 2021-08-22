package pipelinerun

import (
	"encoding/json"
	"errors"
	devopsv1alpha4 "kubesphere.io/devops/pkg/api/devops/v1alpha4"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	jenkinsTimeLayout = "2006-01-02T15:04:05+0000"
	staticDevOpsURL   = "https://devops.kubesphere.io/"
)

func parseJenkinsTime(jenkinsTime string) (time.Time, error) {
	return time.Parse(jenkinsTimeLayout, jenkinsTime)
}

// mockClientURL is only for HttpParameters.
// Generated URL has no practical significance, but it is indispensable.
func mockClientURL() *url.URL {
	stubURL, err := url.Parse(staticDevOpsURL)
	if err != nil {
		// should never happen
		panic("invalid stub URL: " + staticDevOpsURL)
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
		Body: NopCloser(strings.NewReader(string(bodyJSON))),
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

// TODO Remove those Closer, Reader, ReaderCloser. Because those types and functions from io.go with version 1.16, but
// we are using go v1.13.
// See also: https://github.com/golang/go/issues/40025

// NopCloser returns a ReadCloser with a no-op Close method wrapping
// the provided Reader r.
func NopCloser(r Reader) ReadCloser {
	return nopCloser{r}
}

type nopCloser struct {
	Reader
}

func (nopCloser) Close() error { return nil }

type Reader interface {
	Read(p []byte) (n int, err error)
}

// ReadCloser is the interface that groups the basic Read and Close methods.
type ReadCloser interface {
	Reader
	Closer
}

// Closer is the interface that wraps the basic Close method.
//
// The behavior of Close after the first call is undefined.
// Specific implementations may document their own behavior.
type Closer interface {
	Close() error
}
