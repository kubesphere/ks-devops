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
