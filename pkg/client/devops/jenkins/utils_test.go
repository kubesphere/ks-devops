package jenkins

import (
	"bytes"
	"mime/multipart"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJenkinsQuery(t *testing.T) {
	table := []testData{
		{
			param: "start=0&limit=10&branch=master",
			expected: url.Values{
				"start":  []string{"0"},
				"limit":  []string{"10"},
				"branch": []string{"master"},
			}, err: false,
		},
		{
			param: "branch=master", expected: url.Values{
			"branch": []string{"master"},
		}, err: false,
		},
		{
			param: "&branch=master", expected: url.Values{
			"branch": []string{"master"},
		}, err: false,
		},
		{
			param: "branch=master&", expected: url.Values{
			"branch": []string{"master"},
		}, err: false,
		},
		{
			param: "branch=%gg", expected: url.Values{}, err: true,
		},
		{
			param: "%gg=fake", expected: url.Values{}, err: true,
		},
	}

	for index, item := range table {
		result, err := ParseJenkinsQuery(item.param)
		if item.err {
			assert.NotNil(t, err, "index: [%d], unexpected error happen %v", index, err)
		} else {
			assert.Nil(t, err, "index: [%d], unexpected error happen %v", index, err)
		}
		assert.Equal(t, item.expected, result, "index: [%d], result do not match with the expect value", index)
	}
}

type testData struct {
	param    string
	expected interface{}
	err      bool
}

func TestUploadFunc(t *testing.T) {
	testFileName := "/tmp/upload.tmp"
	testWriter := multipart.NewWriter(&bytes.Buffer{})
	// The first call should fail because the file doesn't exist 
	err := UploadFunc(testFileName, testWriter)
	assert.NotNil(t, err)
	
	// Create tmp file
	_, err = os.Create(testFileName)
	assert.Nil(t, err, "create tmp file has error: %v", err)
	defer func ()  {
		err := os.Remove(testFileName)
		assert.Nil(t, err, "delete tmp file has error: %v", err)
	}()
	
	// The second call Bad should fail because writer is bad
	badWriter := multipart.NewWriter(&badWriter{}) 
	err = UploadFunc(testFileName, badWriter)
	assert.NotNil(t, err)

	// Final should succeed
	err = UploadFunc(testFileName, testWriter)
	assert.Nil(t, err, "UploadFunc has error: %v", err)
}

type badWriter struct {
	err error
}

func (w badWriter) Write([]byte) (int, error) {
	return 0, w.err
}