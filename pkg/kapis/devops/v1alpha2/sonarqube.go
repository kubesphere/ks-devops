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
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/client/sonarqube"
	"kubesphere.io/devops/pkg/constants"
	"net/http"
)

func addSonarqubeToWebService(webservice *restful.WebService, devopsClient devops.Interface, sonarClient sonarqube.SonarInterface,
	k8sClient k8s.Client) error {
	sonarEnable := devopsClient != nil && sonarClient != nil
	if sonarEnable {
		sonarHandler := NewPipelineSonarHandler(devopsClient, sonarClient, k8sClient)
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/sonarstatus").
			To(sonarHandler.GetPipelineSonarStatusHandler).
			Doc("Get the sonar quality information for the specified pipeline of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
			Returns(http.StatusOK, api.StatusOK, []sonarqube.SonarStatus{}).
			Writes([]sonarqube.SonarStatus{}))

		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/sonarstatus").
			To(sonarHandler.GetMultiBranchesPipelineSonarStatusHandler).
			Doc("Get the sonar quality check information for the specified pipeline branch of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
			Param(webservice.PathParameter("branch", "branch name, e.g. master")).
			Returns(http.StatusOK, api.StatusOK, []sonarqube.SonarStatus{}).
			Writes([]sonarqube.SonarStatus{}))
	} else {
		klog.Infof("Sonarqube integration is disabled")
	}
	return nil
}
