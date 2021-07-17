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

package v1alpha3

import (
	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	"testing"
	"time"
)

func TestLabelMatch(t *testing.T) {
	tests := []struct {
		labels       map[string]string
		filter       string
		expectResult bool
	}{
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace",
			expectResult: true,
		},
		{
			labels: map[string]string{
				"kubesphere.io/creator": "system",
			},
			filter:       "kubesphere.io/workspace",
			expectResult: false,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace=",
			expectResult: false,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace!=",
			expectResult: true,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace!=kubesphere-system",
			expectResult: false,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace=kubesphere-system",
			expectResult: true,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace=system",
			expectResult: false,
		},
	}
	for i, test := range tests {
		result := labelMatch(test.labels, test.filter)
		if result != test.expectResult {
			t.Errorf("case %d, got %#v, expected %#v", i, result, test.expectResult)
		}
	}
}

func TestLabelsMatch(t *testing.T) {
	table := []struct {
		name          string
		labels        map[string]string
		filter        string
		expectedMatch bool
	}{{
		name: "fully match(single)",
		labels: map[string]string{
			"kubesphere.io/creator": "admin",
		},
		filter:        "kubesphere.io/creator=admin",
		expectedMatch: true,
	}, {
		name: "fully mismatch(single)",
		labels: map[string]string{
			"kubesphere.io/creator": "admin",
		},
		filter:        "kubesphere.io/creator=tester",
		expectedMatch: false,
	}, {
		name: "fully match(multi)",
		labels: map[string]string{
			"kubesphere.io/creator": "admin",
			"kubesphere.io/status":  "success",
		},
		filter:        "kubesphere.io/creator=admin,kubesphere.io/status=success",
		expectedMatch: true,
	}, {
		name: "partial match(multi)",
		labels: map[string]string{
			"kubesphere.io/creator": "admin",
			"kubesphere.io/status":  "success",
		},
		filter:        "kubesphere.io/creator=tester,kubesphere.io/status=success",
		expectedMatch: false,
	}, {
		name: "fully mismatch(multi)",
		labels: map[string]string{
			"kubesphere.io/creator": "admin",
			"kubesphere.io/status":  "success",
		},
		filter:        "kubesphere.io/creator=tester,kubesphere.io/status=fail",
		expectedMatch: false,
	}, {
		name:          "empty labels",
		labels:        map[string]string{},
		filter:        "kubesphere.io/creator=admin",
		expectedMatch: false,
	},
	}
	for _, item := range table {
		t.Run(item.name, func(t *testing.T) {
			match := labelsMatch(item.labels, item.filter)
			assert.Equal(t, item.expectedMatch, match)
		})
	}
}

func TestDefaultObjectMetaCompare(t *testing.T) {
	now := v1.Now()
	tables := []struct {
		name              string
		left              v1.ObjectMeta
		right             v1.ObjectMeta
		field             query.Field
		expectedCmpResult bool
	}{{
		name: "compare name with descending order",
		left: v1.ObjectMeta{
			Name: "b",
		},
		right: v1.ObjectMeta{
			Name: "a",
		},
		field:             query.FieldName,
		expectedCmpResult: true,
	}, {
		name: "compare same name",
		left: v1.ObjectMeta{
			Name: "a",
		},
		right: v1.ObjectMeta{
			Name: "a",
		},
		field:             query.FieldName,
		expectedCmpResult: false,
	}, {
		name: "compare name with ascending order",
		left: v1.ObjectMeta{
			Name: "a",
		},
		right: v1.ObjectMeta{
			Name: "b",
		},
		field:             query.FieldName,
		expectedCmpResult: false,
	}, {
		name: "compare same creation timestamp",
		left: v1.ObjectMeta{
			Name:              "a",
			CreationTimestamp: now,
		},
		right: v1.ObjectMeta{
			Name:              "b",
			CreationTimestamp: now,
		},
		field:             query.FieldCreationTimeStamp,
		expectedCmpResult: false,
	}, {
		name: "compare creation timestamp with descending order",
		left: v1.ObjectMeta{
			CreationTimestamp: now,
		},
		right: v1.ObjectMeta{
			CreationTimestamp: v1.NewTime(now.Truncate(time.Second)),
		},
		field:             query.FieldCreationTimeStamp,
		expectedCmpResult: true,
	}, {
		name: "compare creation timestamp with ascending order",
		left: v1.ObjectMeta{
			CreationTimestamp: now,
		},
		right: v1.ObjectMeta{
			CreationTimestamp: v1.NewTime(now.Add(time.Second)),
		},
		field:             query.FieldCreationTimeStamp,
		expectedCmpResult: false,
	}, {
		name: "compare others",
		left: v1.ObjectMeta{
			CreationTimestamp: now,
		},
		right: v1.ObjectMeta{
			CreationTimestamp: v1.NewTime(now.Add(time.Second)),
		},
		field:             query.FieldUID,
		expectedCmpResult: false,
	},
	}

	for _, item := range tables {
		t.Run(item.name, func(t *testing.T) {
			result := DefaultObjectMetaCompare(item.left, item.right, item.field)
			assert.Equal(t, item.expectedCmpResult, result)
		})
	}
}

func TestDefaultList(t *testing.T) {
	objectMetaFilterFunc := func(object runtime.Object, filter query.Filter) bool {
		if pipeline, ok := object.(*v1alpha3.Pipeline); !ok {
			return false
		} else {
			return DefaultObjectMetaFilter(pipeline.ObjectMeta, filter)
		}
	}
	objectMetaCompareFunc := func(leftObj runtime.Object, rightObj runtime.Object, field query.Field) bool {
		leftPipeline, ok := leftObj.(*v1alpha3.Pipeline)
		if !ok {
			return false
		}
		rightPipeline, ok := rightObj.(*v1alpha3.Pipeline)
		if !ok {
			return false
		}
		return DefaultObjectMetaCompare(leftPipeline.ObjectMeta, rightPipeline.ObjectMeta, field)
	}

	table := []struct {
		name           string
		items          []runtime.Object
		query          query.Query
		filterFunc     FilterFunc
		compareFunc    CompareFunc
		transform      TransformFunc
		expectedResult api.ListResult
	}{{
		name:        "nil items, compareFunc, filterFunc and transform",
		items:       nil,
		filterFunc:  nil,
		compareFunc: nil,
		expectedResult: api.ListResult{
			TotalItems: 0,
			Items:      []interface{}{},
		},
	}, {
		name: "nil compareFunc, filterFunc and transform",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-a",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-b",
				},
			},
		},
		filterFunc:  nil,
		compareFunc: nil,
		expectedResult: api.ListResult{
			TotalItems: 2,
			Items: []interface{}{
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-a",
					},
				},
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-b",
					},
				},
			},
		},
	}, {
		name: "filter name",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-abcd",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-efgh",
				},
			},
		},
		filterFunc:  objectMetaFilterFunc,
		compareFunc: nil,
		query: query.Query{
			Filters: map[query.Field]query.Value{
				"name": "bc",
			},
		},
		expectedResult: api.ListResult{
			TotalItems: 1,
			Items: []interface{}{
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-abcd",
					},
				},
			},
		},
	}, {
		name: "filter name",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-abcd",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-efgh",
				},
			},
		},
		filterFunc:  nil,
		compareFunc: objectMetaCompareFunc,
		query: query.Query{
			SortBy:    query.FieldName,
			Ascending: false,
		},
		expectedResult: api.ListResult{
			TotalItems: 2,
			Items: []interface{}{
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-efgh",
					},
				},
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-abcd",
					},
				},
			},
		},
	}, {
		name: "filter and compare name",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-cdef",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-abcd",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-efgh",
				},
			},
		},
		filterFunc:  objectMetaFilterFunc,
		compareFunc: objectMetaCompareFunc,
		query: query.Query{
			Filters: map[query.Field]query.Value{
				query.FieldName: "cd",
			},
			SortBy:    query.FieldName,
			Ascending: true,
		},
		expectedResult: api.ListResult{
			TotalItems: 2,
			Items: []interface{}{
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-abcd",
					},
				},
				&v1alpha3.Pipeline{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-cdef",
					},
				},
			},
		},
	}}
	for _, item := range table {
		t.Run(item.name, func(t *testing.T) {
			result := DefaultList(item.items, &item.query, item.compareFunc, item.filterFunc, item.transform)
			assert.DeepEqual(t, item.expectedResult, *result)
		})
	}
}
