package pipelinerun

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/models/pipelinerun"
)

// jenkinsHandler handles some actions with Jenkins endpoint.
type jenkinsHandler struct {
	*core.JenkinsCore
}

// getPipelineNodeDetails gets node details including pipeline steps.
func (handler *jenkinsHandler) getPipelineNodeDetails(pipelineName, namespace string, pr *v1alpha3.PipelineRun) ([]pipelinerun.NodeDetail, error) {
	runID, exists := pr.GetPipelineRunID()
	if !exists {
		return nil, fmt.Errorf("unable to get PipelineRun nodes due to not found run ID")
	}
	branch, err := getSCMRefName(&pr.Spec)
	if err != nil {
		return nil, err
	}
	c := job.BlueOceanClient{JenkinsCore: *handler.JenkinsCore, Organization: "jenkins"}
	nodes, err := c.GetNodes(job.GetNodesOption{
		Pipelines: []string{namespace, pipelineName},
		Branch:    branch,
		RunID:     runID,
	})
	if err != nil {
		return nil, err
	}

	// get steps for every node
	nodeDetails := []pipelinerun.NodeDetail{}
	for _, node := range nodes {
		jobSteps, err := handler.getSteps(node.ID, pipelineName, namespace, pr)
		if err != nil {
			return nil, err
		}
		steps := make([]pipelinerun.Step, 0, len(jobSteps))
		for i := range jobSteps {
			steps = append(steps, pipelinerun.Step{
				Step: jobSteps[i],
			})
		}
		nodeDetails = append(nodeDetails, pipelinerun.NodeDetail{
			Node:  node,
			Steps: steps,
		})
	}
	return nodeDetails, nil
}

func (handler *jenkinsHandler) getSteps(nodeID, pipelineName, namespace string, pr *v1alpha3.PipelineRun) ([]job.Step, error) {
	runID, exists := pr.GetPipelineRunID()
	if !exists {
		return nil, fmt.Errorf("unable to get PipelineRun all steps due to not found runID")
	}
	branch, err := getSCMRefName(&pr.Spec)
	if err != nil {
		return nil, err
	}
	c := job.BlueOceanClient{JenkinsCore: *handler.JenkinsCore, Organization: "jenkins"}
	return c.GetSteps(job.GetStepsOption{
		RunID:        runID,
		Branch:       branch,
		PipelineName: pipelineName,
		Folders:      []string{namespace},
		NodeID:       nodeID,
	})
}

func (handler *jenkinsHandler) getPipelineRunResult(devopsProjectName, pipelineName string, pr *v1alpha3.PipelineRun) (*job.PipelineRun, error) {
	runID, exists := pr.GetPipelineRunID()
	if !exists {
		return nil, fmt.Errorf("unable to get PipelineRun result due to not found run ID")
	}
	branch, err := getSCMRefName(&pr.Spec)
	if err != nil {
		return nil, err
	}
	c := job.BlueOceanClient{JenkinsCore: *handler.JenkinsCore, Organization: "jenkins"}
	return c.GetBuild(job.GetBuildOption{
		RunID:     runID,
		Pipelines: []string{devopsProjectName, pipelineName},
		Branch:    branch,
	})
}

func (handler *jenkinsHandler) triggerJenkinsJob(devopsProjectName, pipelineName string, prSpec *v1alpha3.PipelineRunSpec) (*job.PipelineRun, error) {
	c := job.BlueOceanClient{JenkinsCore: *handler.JenkinsCore, Organization: "jenkins"}

	branch, err := getSCMRefName(prSpec)
	if err != nil {
		return nil, err
	}

	return c.Build(job.BuildOption{
		Pipelines:  []string{devopsProjectName, pipelineName},
		Parameters: parameterConverter{parameters: prSpec.Parameters}.convert(),
		Branch:     branch,
	})
}

func (handler *jenkinsHandler) deleteJenkinsJobHistory(pipelineRun *v1alpha3.PipelineRun) (err error) {
	var buildNum int
	if buildNum = getJenkinsBuildNumber(pipelineRun); buildNum < 0 {
		return
	}

	jenkinsClient := job.Client{JenkinsCore: *handler.JenkinsCore}
	jobName := fmt.Sprintf("job/%s/job/%s", pipelineRun.Namespace, pipelineRun.Spec.PipelineRef.Name)
	if err = jenkinsClient.DeleteHistory(jobName, buildNum); err != nil {
		// TODO improve the way to check if the desired build record was deleted
		if strings.Contains(err.Error(), "not found resources") {
			err = nil
		} else {
			err = fmt.Errorf("failed to delete Jenkins job: %s, build: %d, error: %v", jobName, buildNum, err)
		}
	}
	return
}

// getJenkinsBuildNumber returns the build number of a Jenkins job build which related with a PipelineRun
// return a negative value if there is no valid build number
func getJenkinsBuildNumber(pipelineRun *v1alpha3.PipelineRun) (num int) {
	num = -1

	var (
		buildNum      string
		buildNumExist bool
	)

	if buildNum, buildNumExist = pipelineRun.GetPipelineRunID(); !buildNumExist {
		return
	}

	var err error
	if num, err = strconv.Atoi(buildNum); err != nil {
		num = -1
		klog.V(7).Infof("found an invalid build number from PipelineRun: %s/%s, err: %v",
			pipelineRun.Namespace, pipelineRun.Name, err)
	}
	return
}
