package template

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"testing"
)

func Test_render(t *testing.T) {
	type args struct {
		templateObject v1alpha1.TemplateObject
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, object v1alpha1.TemplateObject)
	}{{
		name: "Should render template into annotations",
		args: args{
			templateObject: &v1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-template",
				},
				Spec: v1alpha1.TemplateSpec{
					Template: "fake-template-content",
				},
			},
		},
		verify: func(t *testing.T, object v1alpha1.TemplateObject) {
			gotRenderResult := object.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey]
			wantRenderResult := object.TemplateSpec().Template
			assert.Equal(t, wantRenderResult, gotRenderResult)
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			render(tt.args.templateObject)
			if tt.verify != nil {
				tt.verify(t, tt.args.templateObject)
			}
		})
	}
}
