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
	assert2 "github.com/stretchr/testify/assert"
	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	"reflect"
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
			"kubesphere.io/synced":  "true",
		},
		filter:        "kubesphere.io/creator=admin,kubesphere.io/status=success,kubesphere.io/synced=true",
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
	now := metav1.Now()
	tables := []struct {
		name              string
		left              metav1.ObjectMeta
		right             metav1.ObjectMeta
		field             query.Field
		expectedCmpResult bool
	}{{
		name: "compare name with descending order",
		left: metav1.ObjectMeta{
			Name: "b",
		},
		right: metav1.ObjectMeta{
			Name: "a",
		},
		field:             query.FieldName,
		expectedCmpResult: true,
	}, {
		name: "compare same name",
		left: metav1.ObjectMeta{
			Name: "a",
		},
		right: metav1.ObjectMeta{
			Name: "a",
		},
		field:             query.FieldName,
		expectedCmpResult: false,
	}, {
		name: "compare name with ascending order",
		left: metav1.ObjectMeta{
			Name: "a",
		},
		right: metav1.ObjectMeta{
			Name: "b",
		},
		field:             query.FieldName,
		expectedCmpResult: false,
	}, {
		name: "compare same creation timestamp",
		left: metav1.ObjectMeta{
			Name:              "a",
			CreationTimestamp: now,
		},
		right: metav1.ObjectMeta{
			Name:              "b",
			CreationTimestamp: now,
		},
		field:             query.FieldCreationTimeStamp,
		expectedCmpResult: false,
	}, {
		name: "compare creation timestamp with descending order",
		left: metav1.ObjectMeta{
			CreationTimestamp: now,
		},
		right: metav1.ObjectMeta{
			CreationTimestamp: metav1.NewTime(now.Truncate(time.Second)),
		},
		field:             query.FieldCreationTimeStamp,
		expectedCmpResult: true,
	}, {
		name: "compare creation timestamp with ascending order",
		left: metav1.ObjectMeta{
			CreationTimestamp: now,
		},
		right: metav1.ObjectMeta{
			CreationTimestamp: metav1.NewTime(now.Add(time.Second)),
		},
		field:             query.FieldCreationTimeStamp,
		expectedCmpResult: false,
	}, {
		name: "compare others",
		left: metav1.ObjectMeta{
			CreationTimestamp: now,
		},
		right: metav1.ObjectMeta{
			CreationTimestamp: metav1.NewTime(now.Add(time.Second)),
		},
		field:             query.FieldUID,
		expectedCmpResult: false,
	},
	}

	for _, item := range tables {
		t.Run(item.name, func(t *testing.T) {
			result := DefaultObjectMetaCompare(&item.left, &item.right, item.field)
			assert.Equal(t, item.expectedCmpResult, result)
		})
	}
}

func TestDefaultList(t *testing.T) {
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
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-a",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
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
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-a",
					},
				},
				&v1alpha3.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-b",
					},
				},
			},
		},
	}, {
		name: "filter name",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-abcd",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-efgh",
				},
			},
		},
		filterFunc:  DefaultFilter(),
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
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-abcd",
					},
				},
			},
		},
	}, {
		name: "filter name",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-abcd",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-efgh",
				},
			},
		},
		filterFunc:  nil,
		compareFunc: DefaultCompare(),
		query: query.Query{
			SortBy:    query.FieldName,
			Ascending: false,
		},
		expectedResult: api.ListResult{
			TotalItems: 2,
			Items: []interface{}{
				&v1alpha3.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-efgh",
					},
				},
				&v1alpha3.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-abcd",
					},
				},
			},
		},
	}, {
		name: "filter and compare name",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-cdef",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-abcd",
				},
			},
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-efgh",
				},
			},
		},
		filterFunc:  DefaultFilter(),
		compareFunc: DefaultCompare(),
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
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-abcd",
					},
				},
				&v1alpha3.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-cdef",
					},
				},
			},
		},
	}, {
		name: "Filter nil items",
		items: []runtime.Object{
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-abcd",
				},
			},
			nil,
			&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pipeline-cdef",
				},
			},
			nil,
			nil,
		},
		filterFunc: DefaultFilter(),
		expectedResult: api.ListResult{
			TotalItems: 2,
			Items: []interface{}{
				&v1alpha3.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pipeline-abcd",
					},
				},
				&v1alpha3.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
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

func Test_noTransformFunc(t *testing.T) {
	tests := []struct {
		name string
		arg  runtime.Object
		want runtime.Object
	}{{
		name: "Nil object",
		arg:  nil,
	}, {
		name: "Non-nil object",
		arg:  &v1.Pod{},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NoTransformFunc()(tt.arg); got != tt.arg {
				t.Errorf("NoTransformFunc() = %v, want %v", got, tt.arg)
			}
		})
	}
}

func Test_nilFilter(t *testing.T) {
	defaultTransformFunc := NoTransformFunc()
	fakeTransformFunc := func() TransformFunc {
		return func(object runtime.Object) interface{} {
			return nil
		}
	}
	type args struct {
		transformFuncs []TransformFunc
	}
	tests := []struct {
		name      string
		args      args
		assertion func([]TransformFunc)
	}{{
		name: "Nil transform functions",
		args: args{
			transformFuncs: nil,
		},
		assertion: func(funcs []TransformFunc) {
			assert2.Nil(t, funcs)
		},
	}, {
		name: "Contain nil transformFuncs",
		args: args{
			transformFuncs: []TransformFunc{defaultTransformFunc, nil},
		},
		assertion: func(funcs []TransformFunc) {
			assert2.Equal(t, 1, len(funcs))
			assert2.Equal(t, reflect.TypeOf(defaultTransformFunc), reflect.TypeOf(funcs[0]))
		},
	}, {
		name: "Keep the same sequence",
		args: args{
			transformFuncs: []TransformFunc{defaultTransformFunc, nil, fakeTransformFunc()},
		},
		assertion: func(funcs []TransformFunc) {
			assert2.Equal(t, 2, len(funcs))
			assert2.Equal(t, reflect.TypeOf(defaultTransformFunc), reflect.TypeOf(funcs[0]))
			assert2.Equal(t, reflect.TypeOf(fakeTransformFunc()), reflect.TypeOf(funcs[1]))
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(nilFilter(tt.args.transformFuncs))
		})
	}
}

func TestFilterFunc_And(t *testing.T) {
	type args struct {
		anotherFf FilterFunc
	}
	tests := []struct {
		name string
		ff   FilterFunc
		args args
		want bool
	}{{
		name: "false && true",
		ff:   alwaysFalseFilter,
		args: args{
			anotherFf: alwaysTrueFilter,
		},
		want: false,
	}, {
		name: "true && false",
		ff:   alwaysTrueFilter,
		args: args{
			anotherFf: alwaysFalseFilter,
		},
		want: false,
	}, {
		name: "false && false",
		ff:   alwaysFalseFilter,
		args: args{
			anotherFf: alwaysFalseFilter,
		},
		want: false,
	}, {
		name: "true && true",
		ff:   alwaysTrueFilter,
		args: args{
			anotherFf: alwaysTrueFilter,
		},
		want: true,
	}, {
		name: "nil && nil",
		ff:   nil,
		args: args{
			nil,
		},
		want: true,
	}, {
		name: "true && nil",
		ff:   alwaysTrueFilter,
		args: args{
			anotherFf: nil,
		},
		want: true,
	}, {
		name: "false && nil",
		ff:   alwaysFalseFilter,
		args: args{
			nil,
		},
		want: false,
	}, {
		name: "nil && true",
		ff:   nil,
		args: args{
			anotherFf: alwaysTrueFilter,
		},
		want: true,
	}, {
		name: "nil && false",
		ff:   nil,
		args: args{
			alwaysFalseFilter,
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ff.And(tt.args.anotherFf)(nil, query.Filter{}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("And() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterFunc_Or(t *testing.T) {
	type args struct {
		anotherFf FilterFunc
	}
	tests := []struct {
		name string
		ff   FilterFunc
		args args
		want bool
	}{{
		name: "false || true",
		ff:   alwaysFalseFilter,
		args: args{
			anotherFf: alwaysTrueFilter,
		},
		want: true,
	}, {
		name: "true || false",
		ff:   alwaysTrueFilter,
		args: args{
			anotherFf: alwaysFalseFilter,
		},
		want: true,
	}, {
		name: "false || false",
		ff:   alwaysFalseFilter,
		args: args{
			anotherFf: alwaysFalseFilter,
		},
		want: false,
	}, {
		name: "true || true",
		ff:   alwaysTrueFilter,
		args: args{
			anotherFf: alwaysTrueFilter,
		},
		want: true,
	}, {
		name: "nil || nil",
		ff:   nil,
		args: args{
			anotherFf: nil,
		},
		want: false,
	}, {
		name: "nil || true",
		ff:   nil,
		args: args{
			alwaysTrueFilter,
		},
		want: true,
	}, {
		name: "nil || false",
		ff:   nil,
		args: args{
			alwaysFalseFilter,
		},
		want: false,
	}, {
		name: "false || nil",
		ff:   alwaysFalseFilter,
		args: args{
			nil,
		},
		want: false,
	}, {
		name: "true || nil",
		ff:   alwaysTrueFilter,
		args: args{
			nil,
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ff.Or(tt.args.anotherFf)(nil, query.Filter{}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Or() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultObjectMetaFilter(t *testing.T) {
	obj := &v1.ConfigMap{}
	obj.SetUID(types.UID("uid"))
	obj.SetName("test")
	obj.SetNamespace("ns")
	obj.SetOwnerReferences([]metav1.OwnerReference{{
		UID:  "uid",
		Kind: "kind",
	}})
	obj.SetAnnotations(map[string]string{
		"anno-key": "anno-value",
	})
	obj.SetLabels(map[string]string{
		"label-key": "label-value",
	})

	type args struct {
		item   metav1.Object
		filter query.Filter
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "filter by names",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "names",
				Value: "test,fake",
			},
		},
		want: true,
	}, {
		name: "filter by names, not have name",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "names",
				Value: "good,fake",
			},
		},
		want: false,
	}, {
		name: "filter by name",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "name",
				Value: "test",
			},
		},
		want: true,
	}, {
		name: "filter by name, not match name",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "name",
				Value: "fake",
			},
		},
		want: false,
	}, {
		name: "filter by uid",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "uid",
				Value: "uid",
			},
		},
		want: true,
	}, {
		name: "filter by uid, not match with uid",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "uid",
				Value: "fake",
			},
		},
		want: false,
	}, {
		name: "filter by namespace",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "namespace",
				Value: "ns",
			},
		},
		want: true,
	}, {
		name: "filter by namespace, not match with namespace",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "namespace",
				Value: "fake",
			},
		},
		want: false,
	}, {
		name: "filter by ownerReference",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "ownerReference",
				Value: "uid",
			},
		},
		want: true,
	}, {
		name: "filter by ownerReference, not match",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "ownerReference",
				Value: "fake",
			},
		},
		want: false,
	}, {
		name: "filter by ownerReference kind",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "ownerKind",
				Value: "kind",
			},
		},
		want: true,
	}, {
		name: "filter by ownerReference kind, not match",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "ownerKind",
				Value: "fake",
			},
		},
		want: false,
	}, {
		name: "filter by annotation",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "annotation",
				Value: "anno-key=anno-value",
			},
		},
		want: true,
	}, {
		name: "filter by annotation, not match",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "annotation",
				Value: "fake",
			},
		},
		want: false,
	}, {
		name: "filter by label",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "label",
				Value: "label-key=label-value",
			},
		},
		want: true,
	}, {
		name: "filter by label, not match",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "label",
				Value: "fake",
			},
		},
		want: false,
	}, {
		name: "fake filter",
		args: args{
			item: obj.DeepCopy(),
			filter: query.Filter{
				Field: "fake",
				Value: "fake",
			},
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert2.Equalf(t, tt.want, DefaultObjectMetaFilter(tt.args.item, tt.args.filter), "DefaultObjectMetaFilter(%v, %v)", tt.args.item, tt.args.filter)
		})
	}
}

func TestToListResult(t *testing.T) {
	cm1 := &v1.ConfigMap{}
	cm1.SetName("cm1")
	cm2 := &v1.ConfigMap{}
	cm2.SetName("cm2")

	type args struct {
		objects []runtime.Object
		q       *query.Query
		handler ListHandler
	}
	tests := []struct {
		name string
		args args
		want *api.ListResult
	}{{
		name: "normal",
		args: args{
			objects: []runtime.Object{cm1, cm2},
			q:       &query.Query{},
			handler: nil,
		},
		want: &api.ListResult{
			Items:      []interface{}{cm2, cm1},
			TotalItems: 2,
		},
	}, {
		name: "sort by name",
		args: args{
			objects: []runtime.Object{cm1, cm2},
			q:       &query.Query{},
			handler: NamedHandler{},
		},
		want: &api.ListResult{
			Items:      []interface{}{cm2, cm1},
			TotalItems: 2,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert2.Equalf(t, tt.want, ToListResult(tt.args.objects, tt.args.q, tt.args.handler), "ToListResult(%v, %v, %v)", tt.args.objects, tt.args.q, tt.args.handler)
		})
	}
}
