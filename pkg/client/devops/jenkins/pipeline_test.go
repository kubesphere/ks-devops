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

package jenkins

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"kubesphere.io/devops/pkg/client/devops"
)

func TestResetPaging(t *testing.T) {
	table := []struct {
		path     string
		rawQuery string
		start    int
		limit    int
		hasErr   bool
		message  string
	}{{
		start:   0,
		limit:   10,
		hasErr:  false,
		message: "without query, should no errors",
	}, {
		path:     "/fake/path",
		rawQuery: "?start=1&limit1",
		start:    0,
		limit:    10,
		hasErr:   false,
		message:  "without a query",
	}, {
		path:     "/fake/path",
		rawQuery: "?start=1&limit1",
		start:    3,
		limit:    13,
		hasErr:   false,
		message:  "without a query",
	}}

	for index, item := range table {
		pip := &Pipeline{
			Path: item.path,
			HttpParameters: &devops.HttpParameters{
				Url: &url.URL{
					Path:     item.path,
					RawQuery: item.rawQuery,
				},
			},
		}

		resultPath, err := pip.resetPaging(item.start, item.limit)
		if item.hasErr {
			assert.NotNil(t, err, printTestMessage(index, item.message))
		} else {
			assert.Nil(t, err, printTestMessage(index, item.message))

			assert.Equal(t, item.path, resultPath.Path, printTestMessage(index, item.message))
			assert.Equal(t, strconv.Itoa(item.start), pip.HttpParameters.Url.Query().Get("start"),
				printTestMessage(index, item.message))
			assert.Equal(t, strconv.Itoa(item.limit), pip.HttpParameters.Url.Query().Get("limit"),
				printTestMessage(index, item.message))
		}
	}
}

func TestParsePaging(t *testing.T) {
	table := []struct {
		targetUrl string
		start     int
		limit     int
		message   string
	}{{
		targetUrl: "http://localhost?start=0&limit=0",
		start:     0,
		limit:     0,
		message:   "should be success",
	}, {
		targetUrl: "http://localhost?start=1&limit=10",
		start:     1,
		limit:     10,
		message:   "should be success",
	}, {
		targetUrl: "http://localhost?start=5&limit=55",
		start:     5,
		limit:     55,
		message:   "should be success",
	}}

	for index, item := range table {
		pipUrl, _ := url.Parse(item.targetUrl)
		pip := &Pipeline{
			HttpParameters: &devops.HttpParameters{
				Url: pipUrl,
			},
		}
		resultStart, resultLimit := pip.parsePaging()

		assert.Equal(t, item.start, resultStart, printTestMessage(index, item.message))
		assert.Equal(t, item.limit, resultLimit, printTestMessage(index, item.message))
	}
}

func printTestMessage(index int, message string) string {
	return fmt.Sprintf("index: %d, message: %s", index, message)
}
