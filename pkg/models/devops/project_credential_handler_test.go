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

package devops

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/devops/fake"
	"testing"
)

func TestNewProjectCredentialOperator(t *testing.T) {
	operator := NewProjectCredentialOperator(nil)
	assert.NotNil(t, operator)
}

func Test_projectCredentialGetter_GetProjectCredentialUsage(t *testing.T) {
	type fields struct {
		devopsClient devops.Interface
	}
	type args struct {
		projectID    string
		credentialID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *devops.Credential
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "not found",
		fields: fields{
			devopsClient: &fake.Devops{},
		},
		args: args{
			projectID:    "projectID",
			credentialID: "credentialID",
		},
		want: nil,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err)
			return true
		},
	}, {
		name: "normal",
		fields: fields{
			devopsClient: &fake.Devops{
				Credentials: map[string]map[string]*v1.Secret{
					"projectID": {
						"credentialID": nil,
					},
				},
			},
		},
		args: args{
			projectID:    "projectID",
			credentialID: "credentialID",
		},
		want: &devops.Credential{
			Id: "credentialID",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &projectCredentialGetter{
				devopsClient: tt.fields.devopsClient,
			}
			got, err := o.GetProjectCredentialUsage(tt.args.projectID, tt.args.credentialID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetProjectCredentialUsage(%v, %v)", tt.args.projectID, tt.args.credentialID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetProjectCredentialUsage(%v, %v)", tt.args.projectID, tt.args.credentialID)
		})
	}
}
