package k8s

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"reflect"
	"testing"
)

func TestNewKubernetesClientWithConfig(t *testing.T) {
	type args struct {
		config *rest.Config
	}
	tests := []struct {
		name       string
		args       args
		wantClient Client
		wantErr    bool
	}{{
		name:       "nil arg",
		args:       args{},
		wantClient: nil,
		wantErr:    false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClient, err := NewKubernetesClientWithConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubernetesClientWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotClient, tt.wantClient) {
				t.Errorf("NewKubernetesClientWithConfig() gotClient = %v, want %v", gotClient, tt.wantClient)
			}
		})
	}
}

func TestNewKubernetesClientWithToken(t *testing.T) {
	type args struct {
		token  string
		master string
	}
	tests := []struct {
		name       string
		args       args
		wantClient Client
		wantErr    bool
	}{{
		name:       "nil arg",
		args:       args{},
		wantClient: nil,
		wantErr:    false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClient, err := NewKubernetesClientWithToken(tt.args.token, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubernetesClientWithToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotClient, tt.wantClient) {
				t.Errorf("NewKubernetesClientWithToken() gotClient = %v, want %v", gotClient, tt.wantClient)
			}
			if gotClient != nil {
				assert.Equal(t, tt.args.master, gotClient.Master())
			}
		})
	}
}

func TestNewKubernetesClient(t *testing.T) {
	type args struct {
		options *KubernetesOptions
	}
	tests := []struct {
		name       string
		args       args
		wantClient Client
		wantErr    bool
	}{{
		name:       "nil arg",
		args:       args{},
		wantClient: nil,
		wantErr:    false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClient, err := NewKubernetesClient(tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubernetesClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotClient, tt.wantClient) {
				t.Errorf("NewKubernetesClient() gotClient = %v, want %v", gotClient, tt.wantClient)
			}
		})
	}
}
