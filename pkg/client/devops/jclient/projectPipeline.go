package jclient

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	devopsv1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/devops"
)

func (j *JenkinsClient) CreateProjectPipeline(projectID string, pipeline *v1alpha3.Pipeline) (string, error) {
	core, err := GetJenkinsCore()
	jclient := job.Client{
		JenkinsCore: core,
		Parent: "",
	}
	if err != nil {
		return "", err
	}
	projectPipelineName := fmt.Sprintf("%s %s", projectID, pipeline.Spec.Pipeline.Name)
	job, _ := jclient.GetJob(projectPipelineName)
	if job != nil {
		err := fmt.Errorf("job name [%s] has been used", job.Name)
		return "", restful.NewError(http.StatusConflict, err.Error())
	}
	switch pipeline.Spec.Type {
	case devopsv1alpha3.NoScmPipelineType:
		createPayload, err := getCreatePayload(pipeline.Spec.Pipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		err = jclient.CreateJobInFolder(*createPayload, projectID)
		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		return pipeline.Name, nil
	case devopsv1alpha3.MultiBranchPipelineType:
		createPayload, err := getCreateMultiBranchPipelinePayload(pipeline.Spec.Pipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		err = jclient.CreateJobInFolder(*createPayload, projectID)
		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		return pipeline.Name, nil
	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

func (j *JenkinsClient) DeleteProjectPipeline(projectID string, pipelineID string) (string, error) {
	// TODO: delete a pipeline
	return "", nil
}
func (j *JenkinsClient) UpdateProjectPipeline(projectID string, pipeline *devopsv1alpha3.Pipeline) (string, error) {
	// TODO: Update a pipeline
	return "", nil
}

func (j *JenkinsClient) GetProjectPipelineConfig(projectID, pipelineID string) (*devopsv1alpha3.Pipeline, error) {
	// TODO: get a pipeline config
	return nil, nil
}

func getCreatePayload(pipeline *devopsv1alpha3.NoScmPipeline) (jobPayload *job.CreateJobPayload, err error) {
	// NoScmPipeline do not have copy mode to create a pipeline
	jobPayload = &job.CreateJobPayload{
		Mode: "org.jenkinsci.plugins.workflow.job.WorkflowJob",
		Name: pipeline.Name,
	}
	return
}

func getCreateMultiBranchPipelinePayload(pipeline *devopsv1alpha3.NoScmPipeline) (jobPayload *job.CreateJobPayload, err error) {
	// NoScmPipeline do not have copy mode to create a pipeline
	jobPayload = &job.CreateJobPayload{
		Mode: "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject",
		Name: pipeline.Name,
	}
	return
}
