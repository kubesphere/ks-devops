package v1alpha4

import (
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/api/devops/v1alpha4"
	"reflect"
	"testing"
)

func Test_getScm(t *testing.T) {
	type args struct {
		ps     *v1alpha3.PipelineSpec
		branch string
	}
	tests := []struct {
		name    string
		args    args
		want    *v1alpha4.SCM
		wantErr bool
	}{{
		name: "",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getScm(tt.args.ps, tt.args.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("getScm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getScm() got = %v, want %v", got, tt.want)
			}
		})
	}
}
