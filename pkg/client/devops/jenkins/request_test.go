package jenkins

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newFakeRequester() *Requester {
	// init log
	new(Jenkins).initLoggers()

	return &Requester{
		Base:   "localhost",
		Client: http.DefaultClient,
	}
}

func newFakeAPIRequest() *APIRequest {
	return NewAPIRequest("POST", "/test", nil)
}

func TestRequesterDo(t *testing.T) {
	// test upload fail logic
	requester := newFakeRequester()
	fileNames := []string{"a.tmp", "b.tmp"}
	_, err := requester.Do(newFakeAPIRequest(), nil, fileNames)
	assert.NotNil(t, err)
}

func TestRequesterDoGet(t *testing.T) {
	// test upload fail logic
	requester := newFakeRequester()
	fileNames := []string{"a.tmp", "b.tmp"}
	_, err := requester.DoGet(newFakeAPIRequest(), nil, fileNames)
	assert.NotNil(t, err)
}
