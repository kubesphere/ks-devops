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

package runtime

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// ApiRootPath defines the root path of all KubeSphere apis.
	ApiRootPath = "/kapis"

	// MimeMergePatchJson is the mime header used in merge request
	MimeMergePatchJson = "application/merge-patch+json"

	MimeJsonPatchJson = "application/json-patch+json"
)

// Container holds all webservice of apiserver
var Container = restful.NewContainer()

type ContainerBuilder []func(c *restful.Container) error

func init() {
	restful.RegisterEntityAccessor(MimeMergePatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
	restful.RegisterEntityAccessor(MimeJsonPatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
}

func NewWebService(gv schema.GroupVersion) *restful.WebService {
	webservice := restful.WebService{}
	webservice.Path(ApiRootPath + "/" + gv.String()).
		Produces(restful.MIME_JSON)

	return &webservice
}

func NewWebServiceWithoutGroup(gv schema.GroupVersion) *restful.WebService {
	webservice := restful.WebService{}
	webservice.Path(fmt.Sprintf("/%s", gv.Version)).Produces(restful.MIME_JSON)
	return &webservice
}

func (cb *ContainerBuilder) AddToContainer(c *restful.Container) error {
	for _, f := range *cb {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

func (cb *ContainerBuilder) Register(funcs ...func(*restful.Container) error) {
	for _, f := range funcs {
		*cb = append(*cb, f)
	}
}

func NewContainerBuilder(funcs ...func(*restful.Container) error) ContainerBuilder {
	var cb ContainerBuilder
	cb.Register(funcs...)

	return cb
}
