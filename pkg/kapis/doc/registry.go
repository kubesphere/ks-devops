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

package doc

import (
	"github.com/emicklei/go-restful"
	swagger "github.com/emicklei/go-restful-swagger12"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"strings"
)

// AddSwaggerService adds the Swagger service
func AddSwaggerService(wss []*restful.WebService, c *restful.Container) {
	var wssWithSchema []*restful.WebService
	for _, service := range wss {
		if strings.HasPrefix(service.RootPath(), runtime.ApiRootPath) {
			wssWithSchema = append(wssWithSchema, service)
		}
	}

	config := swagger.Config{
		WebServices:     wssWithSchema,
		ApiPath:         "/apidocs.json",
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "bin/swagger-ui/dist",
		Info: swagger.Info{
			Title:             "KubeSphere DevOps",
			Description:       "KubeSphere DevOps OpenAPI",
			TermsOfServiceUrl: "https://kubesphere.io/",
			Contact:           "kubesphere@yunify.com",
			License:           "Apache 2.0",
			LicenseUrl:        "https://www.apache.org/licenses/LICENSE-2.0.html",
		},
	}
	swagger.RegisterSwaggerService(config, c)
}
