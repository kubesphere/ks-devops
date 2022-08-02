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

package hashutil

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestGetMD5(t *testing.T) {
	type args struct {
		reader io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{{
		name:    "normal",
		args:    args{reader: ioutil.NopCloser(bytes.NewBufferString("abc"))},
		want:    "900150983cd24fb0d6963f7d28e17f72",
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMD5(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMD5() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetMD5() got = %v, want %v", got, tt.want)
			}
		})
	}
}
