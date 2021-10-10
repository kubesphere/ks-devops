package fake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"github.com/jenkins-zh/jenkins-client/pkg/mock/mhttp"
	"kubesphere.io/devops/pkg/client/devops/jclient"
)

// NewFakeJenkinsClient creates a fake Jenkins client
func NewFakeJenkinsClient(roundTripper *mhttp.MockRoundTripper) (j *jclient.JenkinsClient) {
	j = &jclient.JenkinsClient{
		Core: core.JenkinsCore{
			URL:          "http://localhost",
			UserName:     "",
			Token:        "",
			RoundTripper: roundTripper,
		},
	}
	return
}

// PrepareForGetProjectPipeline404 only for test
func PrepareForGetProjectPipeline404(roundTripper *mhttp.MockRoundTripper, rootURL, user, password string, path, name string) {
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/job/%s/job/%s/api/json", rootURL, path, name), nil)
	response := &http.Response{
		StatusCode: 404,
		Proto:      "HTTP/1.1",
		Request:    request,
		Body: ioutil.NopCloser(bytes.NewBufferString(`
		{"name":""}
		`)),
	}
	roundTripper.EXPECT().
		RoundTrip(core.NewRequestMatcher(request)).Return(response, nil)
}

// PrepareForCreateProjectPipeline only for test
func PrepareForCreateProjectPipeline(roundTripper *mhttp.MockRoundTripper, rootURL, user, password string, jobPayload job.CreateJobPayload, path string) {
	PrepareForGetProjectPipeline404(roundTripper, rootURL, user, password, jobPayload.Name, path)

	playLoadData, _ := json.Marshal(jobPayload)
	formData := url.Values{
		"json": {string(playLoadData)},
		"name": {jobPayload.Name},
		"mode": {jobPayload.Mode},
		"from": {jobPayload.From},
	}
	payload := strings.NewReader(formData.Encode())
	path = job.ParseJobPath(path)
	requestCreate, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/view/all%s/createItem", rootURL, path), payload)
	requestCreate.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	requestCreate.Header.Add("CrumbRequestField", "Crumb")
	responseCreate := &http.Response{
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Request:    requestCreate,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}
	roundTripper.EXPECT().
		RoundTrip(core.NewRequestMatcher(requestCreate)).Return(responseCreate, nil)

	core.PrepareForGetIssuer(roundTripper, rootURL, user, password)
}
