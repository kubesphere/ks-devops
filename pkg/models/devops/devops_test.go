/*
Copyright 2020 The KubeSphere Authors.

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

package devops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	fakek8sclientset "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	devopsv1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/clientset/versioned"
	fakeclientset "kubesphere.io/devops/pkg/client/clientset/versioned/fake"
	"kubesphere.io/devops/pkg/constants"

	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/devops/fake"
)

const baseUrl = "http://127.0.0.1/kapis/devops.kubesphere.io/v1alpha2/"

func TestGetNodesDetail(t *testing.T) {
	fakeData := make(map[string]interface{})
	PipelineRunNodes := []devops.PipelineRunNodes{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "2",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "3",
			Result:      "SUCCESS",
		},
	}

	NodeSteps := []devops.NodeSteps{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
	}

	fakeData["project1-pipeline1-run1"] = PipelineRunNodes
	fakeData["project1-pipeline1-run1-1"] = NodeSteps
	fakeData["project1-pipeline1-run1-2"] = NodeSteps
	fakeData["project1-pipeline1-run1-3"] = NodeSteps

	devopsClient := fake.NewFakeDevops(fakeData)

	devopsOperator := NewDevopsOperator(devopsClient, nil, nil)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project1/pipelines/pipeline1/runs/run1/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetNodesDetail("project1", "pipeline1", "run1", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails {
		if v.Steps[0].ID == "" {
			t.Fatalf("Can not get any step.")
		}
	}
}

func TestGetBranchNodesDetail(t *testing.T) {
	fakeData := make(map[string]interface{})

	BranchPipelineRunNodes := []devops.BranchPipelineRunNodes{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "2",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "3",
			Result:      "SUCCESS",
		},
	}

	BranchNodeSteps := []devops.NodeSteps{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
	}

	fakeData["project1-pipeline1-branch1-run1"] = BranchPipelineRunNodes
	fakeData["project1-pipeline1-branch1-run1-1"] = BranchNodeSteps
	fakeData["project1-pipeline1-branch1-run1-2"] = BranchNodeSteps
	fakeData["project1-pipeline1-branch1-run1-3"] = BranchNodeSteps

	devopsClient := fake.NewFakeDevops(fakeData)

	devopsOperator := NewDevopsOperator(devopsClient, nil, nil)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project1/pipelines/pipeline1/branchs/branch1/runs/run1/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetBranchNodesDetail("project1", "pipeline1", "branch1", "run1", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails {
		if v.Steps[0].ID == "" {
			t.Fatalf("Can not get any step.")
		}
	}
}

func Test_devopsOperator_CreateDevOpsProject(t *testing.T) {
	type fields struct {
		ksclient versioned.Interface
	}
	type args struct {
		workspace string
		project   *v1alpha3.DevOpsProject
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		verify  func(client versioned.Interface, args args, t *testing.T) bool
		wantErr bool
	}{{
		name: "lack of the generate name",
		fields: fields{
			ksclient: fakeclientset.NewSimpleClientset(),
		},
		args: args{
			workspace: "ws",
			project:   &v1alpha3.DevOpsProject{},
		},
		wantErr: true,
	}, {
		name: "duplicated in the same workspace",
		fields: fields{
			ksclient: fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "fake",
					Labels: map[string]string{
						constants.WorkspaceLabelKey: "ws",
					},
				},
			}),
		},
		args: args{
			workspace: "ws",
			project: &v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "fake",
				},
			},
		},
		wantErr: true,
	}, {
		name: "allow the same name in the different workspaces",
		fields: fields{
			ksclient: fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "fake",
					Name:         "generated-name",
					Labels: map[string]string{
						constants.WorkspaceLabelKey: "ws1",
					},
				},
			}),
		},
		args: args{
			workspace: "ws",
			project: &v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "fake",
				},
			},
		},
		wantErr: false,
	}, {
		name: "normal case",
		fields: fields{
			ksclient: fakeclientset.NewSimpleClientset(),
		},
		args: args{
			workspace: "ws",
			project: &v1alpha3.DevOpsProject{
				ObjectMeta: v1.ObjectMeta{
					GenerateName: "devops",
				},
			},
		},
		wantErr: false,
		verify: func(client versioned.Interface, args args, t *testing.T) bool {
			list, err := client.DevopsV1alpha3().DevOpsProjects().List(context.TODO(), metav1.ListOptions{})
			assert.Nil(t, err)
			assert.Equal(t, len(list.Items) > 0, true)

			for i := range list.Items {
				item := list.Items[i]

				if item.GenerateName == args.project.GenerateName {
					assert.NotNil(t, item.Annotations)
					assert.NotEmpty(t, item.Annotations[v1alpha3.DevOpeProjectSyncStatusAnnoKey])
					assert.NotEmpty(t, item.Annotations[v1alpha3.DevOpeProjectSyncTimeAnnoKey])

					assert.NotNil(t, item.Labels)
					assert.Equal(t, args.workspace, item.Labels[constants.WorkspaceLabelKey])
					return true
				}
			}
			return false
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := devopsOperator{
				ksclient: tt.fields.ksclient,
			}
			got, err := d.CreateDevOpsProject(tt.args.workspace, tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDevOpsProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.verify != nil && !tt.verify(d.ksclient, tt.args, t) {
				t.Errorf("CreateDevOpsProject() got = %v", got)
			}
		})
	}
}

func Test_devopsOperator_GetDevOpsProject(t *testing.T) {
	type fields struct {
		ksclient versioned.Interface
	}
	type args struct {
		workspace   string
		projectName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(result *v1alpha3.DevOpsProject, resultErr error, t *testing.T)
	}{{
		name: "normal case",
		fields: fields{
			ksclient: fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "fake",
					Name:         "generated-name",
					Labels: map[string]string{
						constants.WorkspaceLabelKey: "ws",
					},
				},
			}),
		},
		args: args{
			workspace:   "ws",
			projectName: "fake",
		},
		verify: func(result *v1alpha3.DevOpsProject, resultErr error, t *testing.T) {
			assert.Nil(t, resultErr)
			assert.NotNil(t, result)
			assert.Equal(t, "fake", result.GenerateName)
			assert.Equal(t, "generated-name", result.Name)
		},
	}, {
		name: "cannot find by the same generateName in the different workspaces",
		fields: fields{
			ksclient: fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "fake",
					Name:         "generated-name",
					Labels: map[string]string{
						constants.WorkspaceLabelKey: "ws1",
					},
				},
			}),
		},
		args: args{
			workspace:   "ws",
			projectName: "fake",
		},
		verify: func(result *v1alpha3.DevOpsProject, resultErr error, t *testing.T) {
			assert.NotNil(t, resultErr)
			assert.Nil(t, result)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := devopsOperator{
				ksclient: tt.fields.ksclient,
			}
			got, err := d.GetDevOpsProjectByGenerateName(tt.args.workspace, tt.args.projectName)
			tt.verify(got, err, t)
		})
	}
}

func Test_devopsOperator_BuildPipelineParameters(t *testing.T) {
	type fields struct {
		ksclient  versioned.Interface
		k8sclient kubernetes.Interface
	}
	type args struct {
		projectName  string
		pipelineName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(params []v1alpha3.ParameterDefinition, resultErr error, t *testing.T)
	}{{
		name: "normal case",
		fields: fields{
			ksclient: fakeclientset.NewSimpleClientset(&v1alpha3.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "fake-pipeline",
					Name:         "fake-pipeline",
					Namespace:    "fake-ns",
				},
				Spec: v1alpha3.PipelineSpec{
					Type: "",
					Pipeline: &v1alpha3.NoScmPipeline{
						ParametersFrom: []v1alpha3.ParameterReference{
							{
								TypedLocalObjectReference: corev1.TypedLocalObjectReference{
									APIGroup: new(string),
									Kind:     "ConfigMap",
									Name:     "default/dyn-fake-cm",
								},
								ValuesKey: "params",
								Mode:      devopsv1alpha3.PARAM_REF_MODE_CONFIG,
							},
						},
					},
					MultiBranchPipeline: &v1alpha3.MultiBranchPipeline{},
				},
			},
				&v1alpha3.DevOpsProject{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "fake-project",
						Name:         "fake-project",
					},
					Status: v1alpha3.DevOpsProjectStatus{
						AdminNamespace: "fake-ns",
					},
				},
			),
			k8sclient: fakek8sclientset.NewSimpleClientset(&corev1.ConfigMap{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name:         "dyn-fake-cm",
					GenerateName: "dyn-fake-cm",
					Namespace:    "default",
				},
				Data: map[string]string{
					"params": `- default_value: |-
    k1-a
    k1-b
    k1-c
  name: "dyn-param-k1"
  type: choice
  is_quoted: true
  description: "dyn-param-k1"
- default_value: |-
    k2-d
    k2-e
    k2-f
  name: "dyn-param-k2"
  type: choice
  is_quoted: false
  description: "dyn-param-k2"`,
				},
			}),
		},

		args: args{
			projectName:  "fake-project",
			pipelineName: "fake-pipeline",
		},
		verify: func(params []v1alpha3.ParameterDefinition, resultErr error, t *testing.T) {
			assert.Nil(t, resultErr)
			assert.Len(t, params, 2)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := devopsOperator{
				ksclient:  tt.fields.ksclient,
				k8sclient: tt.fields.k8sclient,
			}
			param, err := d.BuildPipelineParameters(tt.args.projectName, tt.args.pipelineName, nil)
			tt.verify(param, err, t)
		})
	}
}

func Test_devopsOperator_mergeParameters(t *testing.T) {
	consistParams := []v1alpha3.ParameterDefinition{
		{
			Name:         "k1",
			DefaultValue: "v1",
			IsQuoted:     false,
			Type:         "choice",
			Description:  "test",
		},
	}
	externalParams := []v1alpha3.ParameterDefinition{
		{
			Name:         "k1",
			DefaultValue: "v2",
			IsQuoted:     false,
			Type:         "choice",
			Description:  "test",
		},
	}
	allParams := consistParams
	assert.Len(t, allParams, 1)
	allParams = append(allParams, externalParams...)
	assert.Len(t, allParams, 2)
	mergeParams := mergeParameters(allParams)
	assert.Len(t, mergeParams, 1)
	assert.Equal(t, mergeParams[0].Name, "k1")
	assert.Equal(t, mergeParams[0].DefaultValue, "v2")
}

func Test_devopsOperator_checkParametersCMName(t *testing.T) {
	invaildNameWithOutNs := "cm-name"
	invaildNameWithMoreThanTwoSlash := "default/ns/cm-name"
	correctName := "default/cm-name"

	ns, name, err := checkParametersCMName(invaildNameWithOutNs)
	assert.Equal(t, ns, "")
	assert.Equal(t, name, "")
	assert.Equal(t, err, fmt.Errorf("invalid name [%s]", invaildNameWithOutNs))

	ns1, name1, err1 := checkParametersCMName(invaildNameWithMoreThanTwoSlash)
	assert.Equal(t, ns1, "")
	assert.Equal(t, name1, "")
	assert.Equal(t, err1, fmt.Errorf("invalid name [%s]", invaildNameWithMoreThanTwoSlash))

	ns2, name2, err2 := checkParametersCMName(correctName)
	assert.Equal(t, ns2, "default")
	assert.Equal(t, name2, "cm-name")
	assert.Nil(t, err2)
}

func Test_devopsOperator_buildParamFromConfigMapData(t *testing.T) {
	type fields struct {
		k8sclient kubernetes.Interface
	}
	type args struct {
		cmName    string
		valuesKey string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(params []v1alpha3.ParameterDefinition, resultErr error, t *testing.T)
	}{
		{
			name: "normal configmap",
			fields: fields{
				k8sclient: fakek8sclientset.NewSimpleClientset(&corev1.ConfigMap{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name:         "dyn-fake-cm",
						GenerateName: "dyn-fake-cm",
						Namespace:    "default",
					},
					Data: map[string]string{
						"params": `- default_value: |-
    k1-a
    k1-b
    k1-c
  name: "dyn-param-k1"
  type: choice
  is_quoted: true
  description: "dyn-param-k1"
- default_value: |-
    k2-d
    k2-e
    k2-f
  name: "dyn-param-k2"
  type: choice
  is_quoted: false
  description: "dyn-param-k2"`,
					},
				}),
			},
			args: args{
				cmName:    "default/dyn-fake-cm",
				valuesKey: "params",
			},
			verify: func(params []devopsv1alpha3.ParameterDefinition, resultErr error, t *testing.T) {
				assert.Len(t, params, 2)
				assert.Nil(t, resultErr)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := buildParamFromConfigMapData(tt.fields.k8sclient, tt.args.cmName, tt.args.valuesKey)
			tt.verify(param, err, t)
		})
	}
}

func Test_devopsOperator_buildParamFromRESTfull(t *testing.T) {
	type fields struct {
		k8sclient kubernetes.Interface
	}
	type args struct {
		cmName string
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// mock here
		switch strings.TrimSpace(r.URL.Path) {
		case "/":
			params := []devopsv1alpha3.ParameterDefinition{
				{
					Name:         "key-in-rest-response",
					DefaultValue: "val-in-rest-response",
					IsQuoted:     false,
					Type:         "choice",
					Description:  "rest",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(params)
		}
	}))
	defer server.Close()
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(params []v1alpha3.ParameterDefinition, resultErr error, t *testing.T)
	}{
		{
			name: "normal configmap",
			fields: fields{
				k8sclient: fakek8sclientset.NewSimpleClientset(&corev1.ConfigMap{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name:         "dyn-fake-cm-rest",
						GenerateName: "dyn-fake-cm",
						Namespace:    "default",
					},
					Data: map[string]string{
						"url": "http://" + server.Listener.Addr().String(),
					},
				}),
			},
			args: args{
				cmName: "default/dyn-fake-cm-rest",
			},
			verify: func(params []devopsv1alpha3.ParameterDefinition, resultErr error, t *testing.T) {
				assert.Len(t, params, 1)
				assert.Nil(t, resultErr)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := buildParamFromRESTfull(tt.fields.k8sclient, tt.args.cmName, nil)
			tt.verify(param, err, t)
		})
	}
}
