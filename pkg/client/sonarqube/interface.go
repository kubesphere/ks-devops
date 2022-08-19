/*
Copyright 2020 KubeSphere Authors

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

package sonarqube

import (
	sonargo "github.com/kubesphere/sonargo/sonar"
	"k8s.io/klog/v2"

	"kubesphere.io/devops/pkg/client/devops"
)

// SonarInterface represents a SonarQube interface
type SonarInterface interface {
	GetSonarResultsByTaskIds(taskIDS ...string) ([]*SonarStatus, error)
}

// SonarQube represents SonarQube instance
type SonarQube struct {
	client *sonargo.Client
}

// NewSonar creates a sonar instance
func NewSonar(client *sonargo.Client) *SonarQube {
	return &SonarQube{client: client}
}

const (
	// SonarAnalysisActionClass is the class of Sonar in Jenkins
	SonarAnalysisActionClass = "hudson.plugins.sonar.action.SonarAnalysisAction"
	// SonarMetricKeys are the metric keys
	SonarMetricKeys = "alert_status,quality_gate_details,bugs,new_bugs,reliability_rating,new_reliability_rating,vulnerabilities,new_vulnerabilities,security_rating,new_security_rating,code_smells,new_code_smells,sqale_rating,new_maintainability_rating,sqale_index,new_technical_debt,coverage,new_coverage,new_lines_to_cover,tests,duplicated_lines_density,new_duplicated_lines_density,duplicated_blocks,ncloc,ncloc_language_distribution,projects,new_lines"
	// SonarAdditionalFields is the key of the additional fields
	SonarAdditionalFields = "metrics,periods"
)

// SonarStatus represents the status of a sonar request
type SonarStatus struct {
	Measures      *sonargo.MeasuresComponentObject `json:"measures,omitempty"`
	Issues        *sonargo.IssuesSearchObject      `json:"issues,omitempty"`
	GeneralAction *devops.GeneralAction            `json:"generalAction,omitempty"`
	Task          *sonargo.CeTaskObject            `json:"task,omitempty"`
}

// GetSonarResultsByTaskIds gets the sonar result
func (s *SonarQube) GetSonarResultsByTaskIds(taskIDs ...string) ([]*SonarStatus, error) {
	sonarStatuses := make([]*SonarStatus, 0)
	for _, taskID := range taskIDs {
		sonarStatus := &SonarStatus{}
		taskOptions := &sonargo.CeTaskOption{
			Id: taskID,
		}
		ceTask, _, err := s.client.Ce.Task(taskOptions)
		if err != nil {
			klog.Errorf("get sonar task error [%+v]", err)
			continue
		}
		sonarStatus.Task = ceTask
		measuresComponentOption := &sonargo.MeasuresComponentOption{
			Component:        ceTask.Task.ComponentKey,
			AdditionalFields: SonarAdditionalFields,
			MetricKeys:       SonarMetricKeys,
		}
		measures, _, err := s.client.Measures.Component(measuresComponentOption)
		if err != nil {
			klog.Errorf("get sonar task error [%+v]", err)
			continue
		}
		sonarStatus.Measures = measures

		issuesSearchOption := &sonargo.IssuesSearchOption{
			AdditionalFields: "_all",
			ComponentKeys:    ceTask.Task.ComponentKey,
			Resolved:         "false",
			Ps:               "10",
			S:                "FILE_LINE",
			Facets:           "severities,types",
		}
		var issuesSearch *sonargo.IssuesSearchObject
		if issuesSearch, _, err = s.client.Issues.Search(issuesSearchOption); err != nil {
			continue
		}
		sonarStatus.Issues = issuesSearch
		sonarStatuses = append(sonarStatuses, sonarStatus)
	}
	return sonarStatuses, nil
}
