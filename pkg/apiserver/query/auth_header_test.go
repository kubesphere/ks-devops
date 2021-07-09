package query

import (
	"net/http"
	"testing"
)

func TestGetAuthorization(t *testing.T) {
	type args struct {
		header getter
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "param is nil",
		args: args{},
		want: "",
	}, {
		name: "without any valid auth in header",
		args: args{
			header: http.Header{
				"xxx": []string{"fake"},
			},
		},
		want: "",
	}, {
		name: "with X-Authorization in header",
		args: args{
			header: http.Header{
				"X-Authorization": []string{"fake"},
				"Authorization": []string{"rick"},
			},
		},
		want: "fake",
	}, {
		name: "with Authorization in header",
		args: args{
			header: http.Header{
				"Authorization": []string{"fake"},
			},
		},
		want: "fake",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAuthorization(tt.args.header); got != tt.want {
				t.Errorf("GetAuthorization() = %v, want %v", got, tt.want)
			}
		})
	}
}
