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

package net

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPort(t *testing.T) {
	type args struct {
		port int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "valid port",
		args: args{port: 22},
		want: true,
	}, {
		name: "negative port",
		args: args{port: -1},
		want: false,
	}, {
		name: "very big port number",
		args: args{port: 65536},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPort(tt.args.port); got != tt.want {
				t.Errorf("IsValidPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequestIP(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "have X-Real-Ip",
		args: args{req: &http.Request{Header: map[string][]string{
			"X-Real-Ip":       {"0.0.0.0"},
			"X-Forwarded-For": {"1.1.1.1"},
		}}},
		want: "0.0.0.0",
	}, {
		name: "have X-Real-Ip",
		args: args{req: &http.Request{
			Header: map[string][]string{
				"X-Forwarded-For": {"0.0.0.0"},
			},
			RemoteAddr: "2.2.2.2:22",
		}},
		want: "0.0.0.0",
	}, {
		name: "get ip from remote-address",
		args: args{req: &http.Request{
			RemoteAddr: "2.2.2.2:22",
		}},
		want: "2.2.2.2",
	}, {
		name: "get ip from remote-address that does not have the port",
		args: args{req: &http.Request{
			RemoteAddr: "2.2.2.2",
		}},
		want: "2.2.2.2",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRequestIP(tt.args.req); got != tt.want {
				t.Errorf("GetRequestIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantURL string
	}{{
		name:    "ip and port",
		address: "192.168.1.1:8080",
		wantURL: "https://192.168.1.1:8080",
	}, {
		name:    "start with http",
		address: "http://localhost/",
		wantURL: "http://localhost",
	}, {
		name:    "start with https",
		address: "https://localhost/",
		wantURL: "https://localhost",
	}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseURL(tt.address)
			assert.Equal(t, tt.wantURL, result, "failed in case [%d]", i)
		})
	}
}
