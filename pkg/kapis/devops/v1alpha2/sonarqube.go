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

package v1alpha2

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/ks-devops/pkg/api"
	"github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/client/k8s"
	"github.com/kubesphere/ks-devops/pkg/client/sonarqube"
	"github.com/kubesphere/ks-devops/pkg/constants"
)

func addSonarqubeToWebService(webservice *restful.WebService, devopsClient devops.Interface, sonarClient sonarqube.SonarInterface,
	k8sClient k8s.Client) error {
	sonarHandler := NewPipelineSonarHandler(devopsClient, sonarClient, k8sClient)
	webservice.Route(webservice.GET("/namespaces/{devops}/pipelines/{pipeline}/sonarstatus").
		To(sonarHandler.GetPipelineSonarStatusHandler).
		Doc("Get the sonar quality information for the specified pipeline of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Returns(http.StatusOK, api.StatusOK, []sonarqube.SonarStatus{}).
		Writes([]sonarqube.SonarStatus{}))

	webservice.Route(webservice.GET("/namespaces/{devops}/pipelines/{pipeline}/branches/{branch}/sonarstatus").
		To(sonarHandler.GetMultiBranchesPipelineSonarStatusHandler).
		Doc("Get the sonar quality check information for the specified pipeline branch of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
		Metadata(restfulspec.KeyOpenAPITags, constants.DevOpsPipelineTags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Param(webservice.PathParameter("branch", "branch name, e.g. master")).
		Returns(http.StatusOK, api.StatusOK, []sonarqube.SonarStatus{}).
		Writes([]sonarqube.SonarStatus{}))

	return nil
}
