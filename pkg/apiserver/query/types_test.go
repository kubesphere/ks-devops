/*
Copyright 2020 The KubeSphere Authors.

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

package query

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/google/go-cmp/cmp"
)

func TestParseQueryParameter(t *testing.T) {
	tests := []struct {
		description string
		queryString string
		expected    *Query
	}{{
		"test normal case",
		"label=app.kubernetes.io/name=book&name=foo&status=Running&page=1&limit=10&ascending=true",
		&Query{
			Pagination: newPagination(10, 0),
			SortBy:     FieldCreationTimeStamp,
			Ascending:  true,
			Filters: map[Field]Value{
				FieldLabel:  Value("app.kubernetes.io/name=book"),
				FieldName:   Value("foo"),
				FieldStatus: Value("Running"),
			},
		},
	}, {
		"test bad case",
		"xxxx=xxxx&dsfsw=xxxx&page=abc&limit=add&ascending=ssss",
		&Query{
			Pagination: NoPagination,
			SortBy:     FieldCreationTimeStamp,
			Ascending:  false,
			Filters: map[Field]Value{
				Field("xxxx"):  Value("xxxx"),
				Field("dsfsw"): Value("xxxx"),
			},
		},
	},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost?%s", test.queryString), nil)
		if err != nil {
			t.Fatal(err)
		}

		request := restful.NewRequest(req)

		t.Run(test.description, func(t *testing.T) {
			got := ParseQueryParameter(request)
			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
				return
			}
		})
	}
}

func TestPagination_GetValidPagination(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		offset     int
		total      int
		startIndex int
		endIndex   int
	}{
		{
			name:       "Valid pagination 1",
			limit:      1,
			offset:     0,
			total:      1,
			startIndex: 0,
			endIndex:   1,
		},
		{
			name:       "Valid pagination 3",
			limit:      10,
			offset:     1,
			total:      20,
			startIndex: 1,
			endIndex:   11,
		},
		{
			name:       "Invalid pagination 1",
			limit:      1,
			offset:     1,
			total:      1,
			startIndex: 1,
			endIndex:   1,
		},
		{
			name:       "Invalid pagination 2",
			limit:      10,
			offset:     10,
			total:      10,
			startIndex: 10,
			endIndex:   10,
		},
		{
			name:       "Unlimited: Offset = 0",
			limit:      -1,
			offset:     0,
			total:      1000,
			startIndex: 0,
			endIndex:   10,
		},
		{
			name:       "Unlimited: Offset > 0",
			limit:      -1,
			offset:     10,
			total:      5,
			startIndex: 0,
			endIndex:   0,
		},
		{
			name:       "Unlimited: Offset > total",
			limit:      -1,
			offset:     10,
			total:      5,
			startIndex: 0,
			endIndex:   0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pagination := newPagination(test.limit, test.offset)
			startIndex, endIndex := pagination.GetValidPagination(test.total)
			assert.Equal(t, test.startIndex, startIndex)
			assert.Equal(t, test.endIndex, endIndex)
		})
	}

}
