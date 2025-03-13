/*
Copyright 2019 The KubeSphere Authors.

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

package swagger

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/kubesphere/ks-devops/pkg/version"
)

const APIDocPath = "/openapi.json"

func GetSwaggerConfig(container *restful.Container) restfulspec.Config {
	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		APIPath:                       APIDocPath,
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}

	return config
}

func enrichSwaggerObject(swo *spec.Swagger) {
	// v1.Time represents metav1.Time of k8s.
	// We need to hardcode its schema because the generated schema is not correct.
	swo.Definitions["v1.Time"] = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:   []string{"string"},
			Format: "date-time",
		},
	}

	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KubeSphere DevOps",
			Description: "KubeSphere DevOps OpenAPI",
			Version:     version.Get().GitVersion,
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "KubeSphere",
					URL:   "https://kubesphere.io/",
					Email: "kubesphere@yunify.com",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache 2.0",
					URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
		},
	}
}
