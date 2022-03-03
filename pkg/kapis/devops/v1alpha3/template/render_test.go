// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package template

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"testing"
)

func Test_render(t *testing.T) {
	createTemplate := func(name, template string) v1alpha3.TemplateObject {
		return &v1alpha3.Template{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha3.TemplateSpec{
				Template: template,
			},
		}
	}
	type args struct {
		template   v1alpha3.TemplateObject
		parameters []Parameter
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
		verify  func(*testing.T, v1alpha3.TemplateObject)
	}{{
		name: "Should render template without parameters",
		args: args{
			template:   createTemplate("fake-name", "fake-template"),
			parameters: nil,
		},
		verify: func(t *testing.T, template v1alpha3.TemplateObject) {
			got := template.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey]
			assert.Equal(t, "fake-template", got)
		},
		wantErr: assert.NoError,
	}, {
		name: "Should render template with parameters",
		args: args{
			template: createTemplate("fake-name", "The number should be {{ .params.number }}"),
			parameters: []Parameter{{
				Name:  "number",
				Value: "233",
			}},
		},
		verify: func(t *testing.T, template v1alpha3.TemplateObject) {
			got := template.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey]
			assert.Equal(t, "The number should be 233", got)
		},
		wantErr: assert.NoError,
	}, {
		name: "Should render incorrectly without corresponding parameter",
		args: args{
			template: createTemplate("fake-name", "The number should be: {{ .params.number }}"),
			parameters: []Parameter{{
				Name:  "name",
				Value: "fake-name",
			}},
		},
		verify: func(t *testing.T, template v1alpha3.TemplateObject) {
			got := template.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey]
			assert.Equal(t, "The number should be: <no value>", got)
		},
		wantErr: assert.NoError,
	}, {
		name: "Should return error if the template is invalid",
		args: args{
			template: createTemplate("fake-name", "{{}}"),
			parameters: []Parameter{{
				Name:  "name",
				Value: "fake-name",
			}},
		},
		verify: func(t *testing.T, template v1alpha3.TemplateObject) {
			assert.Nil(t, template)
		},
		wantErr: assert.Error,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := render(tt.args.template, tt.args.parameters)
			if !tt.wantErr(t, err, fmt.Sprintf("render(%v, %v)", tt.args.template, tt.args.parameters)) {
				return
			}
			tt.verify(t, template)
		})
	}
}
