package v1alpha4

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"kubesphere.io/devops/pkg/apiserver/query"
	"reflect"
	"testing"
)

func Test_compatibleTransform(t *testing.T) {
	tests := []struct {
		name string
		obj  runtime.Object
		want interface{}
	}{{
		name: "With run status",
		obj: &v1alpha4.PipelineRun{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha4.JenkinsPipelineRunStatusKey: `{"id": "123"}`,
				},
			},
		},
		want: json.RawMessage(`{"id": "123"}`),
	}, {
		name: "Without annotations",
		obj: &v1alpha4.PipelineRun{
			ObjectMeta: v1.ObjectMeta{},
		},
		want: json.RawMessage("{}"),
	}, {
		name: "Nil PipelineRun",
		obj:  (*v1alpha4.PipelineRun)(nil),
		want: json.RawMessage("{}"),
	}, {
		name: "Nil object",
		obj:  nil,
		want: json.RawMessage("{}"),
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := backwardTransform()(tt.obj); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("backwardTransform() = %v, want %v", got, tt.want)
			} else if !reflect.TypeOf(got).AssignableTo(reflect.TypeOf((*json.Marshaler)(nil)).Elem()) {
				t.Errorf("backwardTransform() should return an instance of json.Marshaler, current type is %s", reflect.TypeOf(got))
			}
		})
	}
}

func Test_buildLabelSelector(t *testing.T) {
	parseSelector := func(selector string) labels.Selector {
		parsedSelector, err := labels.Parse(selector)
		if err != nil {
			t.Fatalf("unable to parse labele selector, err = %v", err)
		}
		return parsedSelector
	}

	type args struct {
		queryParam   *query.Query
		pipelineName string
		branchName   string
	}
	tests := []struct {
		name    string
		args    args
		want    labels.Selector
		wantErr bool
	}{{
		name: "No label selector was provided",
		args: args{
			queryParam:   &query.Query{},
			pipelineName: "pipelineA",
			branchName:   "branchA",
		},
		want: parseSelector(fmt.Sprintf("%s=pipelineA,%s=branchA", v1alpha4.PipelineNameLabelKey, v1alpha4.SCMRefNameLabelKey)),
	}, {
		name: "Label selector was provided",
		args: args{
			queryParam: &query.Query{
				LabelSelector: "a=b",
			},
			pipelineName: "pipelineA",
			branchName:   "branchA",
		},
		want: parseSelector(fmt.Sprintf("%s=pipelineA,%s=branchA,a=b", v1alpha4.PipelineNameLabelKey, v1alpha4.SCMRefNameLabelKey)),
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildLabelSelector(tt.args.queryParam, tt.args.pipelineName, tt.args.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildLabelSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildLabelSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertPipelineRunsToObject(t *testing.T) {
	type args struct {
		prs []v1alpha4.PipelineRun
	}
	tests := []struct {
		name string
		args args
		want []runtime.Object
	}{{
		name: "Make sure the sequence is correct",
		args: args{
			prs: []v1alpha4.PipelineRun{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-run-a",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "pipeline-run-b",
					},
				},
			},
		},
		want: []runtime.Object{
			&v1alpha4.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-run-a",
				},
			},
			&v1alpha4.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "pipeline-run-b",
				},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertPipelineRunsToObject(tt.args.prs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertPipelineRunsToObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_backwardFilter(t *testing.T) {
	type args struct {
		obj    runtime.Object
		filter query.Filter
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "Nil object",
		args: args{
			obj: nil,
		},
		want: false,
	}, {
		name: "Nil PipelineRun",
		args: args{
			obj: (*v1alpha4.PipelineRun)(nil),
		},
		want: false,
	}, {
		name: "PipelineRun has started but without Jenkins run status",
		args: args{
			obj: &v1alpha4.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha4.JenkinsPipelineRunIDKey: "123",
					},
				},
			},
		},
		want: false,
	}, {
		name: "PipelineRun hasn't started but with Jenkins run status",
		args: args{
			obj: &v1alpha4.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						v1alpha4.JenkinsPipelineRunStatusKey: `{"id": "123"}`,
					},
				},
			},
		},
		want: false,
	}, {
		name: "PipelineRun has started and with Jenkins run status",
		args: args{
			obj: &v1alpha4.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "abc",
					Annotations: map[string]string{
						v1alpha4.JenkinsPipelineRunStatusKey: `{"id": "123"}`,
						v1alpha4.JenkinsPipelineRunIDKey:     "123",
					},
				},
			},
			filter: query.Filter{
				Field: "name",
				Value: "abc",
			},
		},
		want: true,
	}, {
		name: "PipelineRun has started and with Jenkins run status but failed with default filter",
		args: args{
			obj: &v1alpha4.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name: "abc",
					Annotations: map[string]string{
						v1alpha4.JenkinsPipelineRunStatusKey: `{"id": "123"}`,
						v1alpha4.JenkinsPipelineRunIDKey:     "123",
					},
				},
			},
			filter: query.Filter{
				Field: "name",
				Value: "def",
			},
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := backwardFilter()(tt.args.obj, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("backwardFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
