package v1alpha2

import (
	"bytes"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/mock/gomock"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/mock/mhttp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestJenkinsProxy(t *testing.T) {
	testJenkinsProxy("fake jenkins response", true, t)
	testJenkinsProxy("failed to set auth header for Jenkins API request", false, t)
}

func testJenkinsProxy(responseStr string, correctCrumb bool, t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	roundTripper := mhttp.NewMockRoundTripper(ctrl)
	// common crumb request
	if correctCrumb {
		request, _ := http.NewRequest(http.MethodPost, "http://fake.com/gitlab/serverList", nil)
		request.Header = map[string][]string{
			"User-Agent": {""},
		}
		request.ProtoMajor = 1
		request.ProtoMinor = 0
		request.Proto = "HTTP/1.1"
		core.PrepareForGetIssuer(roundTripper, "http://fake.com", "", "")
		request.Header.Add("CrumbRequestField", "Crumb")
		response := &http.Response{
			StatusCode: 200,
			Request:    request,
			Body:       ioutil.NopCloser(bytes.NewBufferString(responseStr)),
		}
		roundTripper.EXPECT().
			RoundTrip(core.NewVerboseRequestMatcher(request).WithBody().WithQuery()).Return(response, nil)
	} else {
		core.PrepareForGetIssuerWith500(roundTripper, "http://fake.com", "", "")
	}

	client := core.JenkinsCore{
		URL:          "http://fake.com",
		RoundTripper: roundTripper,
	}
	proxy := newJenkinsProxy(client, "fake.com", "http", roundTripper)

	restfulRequest := restful.NewRequest(&http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Host: "target.com",
			Path: fmt.Sprintf("/kapis/%s/%s/devops/%s/jenkins/gitlab/serverList",
				GroupVersion.Group, GroupVersion.Version, "fake.devops"),
		},
		Header: map[string][]string{},
	})
	restfulRequest.PathParameters()["devops"] = "fake.devops"
	httpResponse := &fakeHTTPResponse{}
	restfulResponse := restful.NewResponse(httpResponse)
	proxy.proxyWithDevOps(restfulRequest, restfulResponse)
	assert.Equal(t, responseStr, httpResponse.data.String())
}

func TestNewJenkinsProxy(t *testing.T) {
	assert.NotNil(t, newJenkinsProxy(core.JenkinsCore{}, "", "", nil))
}

type fakeHTTPResponse struct {
	data       bytes.Buffer
	statusCode int
}

func (r *fakeHTTPResponse) Header() http.Header {
	return map[string][]string{}
}

func (r *fakeHTTPResponse) Write(b []byte) (int, error) {
	return r.data.Write(b)
}

func (r *fakeHTTPResponse) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *fakeHTTPResponse) CloseNotify() <-chan bool {
	return nil
}
