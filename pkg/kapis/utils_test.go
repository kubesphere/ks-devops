package kapis

import (
	"io"
	"kubesphere.io/devops/pkg/server/errors"
	"reflect"
	"testing"
)

func TestIgnoreEOF(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want error
	}{{
		name: "Should return nil if error is io.EOF",
		args: args{
			err: io.EOF,
		},
		want: nil,
	}, {
		name: "Should return the same error if error is not io.EOF",
		args: args{
			errors.New("Fake Error"),
		},
		want: errors.New("Fake Error"),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreEOF(tt.args.err); !reflect.DeepEqual(err, tt.want) {
				t.Errorf("IgnoreEOF() error = %v, want %v", err, tt.want)
			}
		})
	}
}
