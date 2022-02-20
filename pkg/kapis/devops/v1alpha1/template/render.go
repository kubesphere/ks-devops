package template

import (
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
)

func render(templateObject v1alpha1.TemplateObject) {
	// TODO Render template using parameters
	template := templateObject.TemplateSpec().Template

	// set template into annotations
	if templateObject.GetAnnotations() == nil {
		templateObject.SetAnnotations(map[string]string{})
	}
	templateObject.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey] = template
}
