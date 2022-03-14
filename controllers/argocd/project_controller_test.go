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

package argocd

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestCreateUnstructuredObject(t *testing.T) {
	project := &v1alpha3.DevOpsProject{}
	project.SetName("name")
	project.SetNamespace("namespace")

	// test against the nil situation
	_, err := createUnstructuredObject(project)
	assert.NotNil(t, err, "should not do anything if there is no spec.argo")

	var obj *unstructured.Unstructured

	// test against the default value situation
	p1 := project.DeepCopy()
	p1.Spec.Argo = &v1alpha3.Argo{}
	obj, err = createUnstructuredObject(p1)
	assert.Nil(t, err)
	result, _, _ := unstructured.NestedStringSlice(obj.Object, "spec", "sourceRepos")
	assert.Equal(t, []string{"*"}, result)
	destinations, _, _ := unstructured.NestedSlice(obj.Object, "spec", "destinations")
	assert.Equal(t, []interface{}{map[string]interface{}{"namespace": "*", "server": "*"}}, destinations)
	whiteList, _, _ := unstructured.NestedSlice(obj.Object, "spec", "clusterResourceWhitelist")
	assert.Equal(t, []interface{}{map[string]interface{}{"group": "*", "kind": "*"}}, whiteList)

	// test against a normal case
	p2 := project.DeepCopy()
	p2.Spec.Argo = &v1alpha3.Argo{
		SourceRepos: []string{"repo1", "repo2"},
		Destinations: []v1alpha3.ApplicationDestination{{
			Server:    "server1",
			Namespace: "ns1",
		}, {
			Server:    "server2",
			Namespace: "ns2",
		}},
		Description: "description",
		ClusterResourceWhitelist: []v1.GroupKind{{
			Group: "group1",
			Kind:  "kind1",
		}, {
			Group: "group2",
			Kind:  "kind2",
		}},
	}
	obj, err = createUnstructuredObject(p2)
	assert.Nil(t, err)
	p2Repos, _, _ := unstructured.NestedStringSlice(obj.Object, "spec", "sourceRepos")
	assert.Equal(t, []string{"repo1", "repo2"}, p2Repos)
	p2Desc, _, _ := unstructured.NestedString(obj.Object, "spec", "description")
	assert.Equal(t, "description", p2Desc)
	p2Destinations, _, _ := unstructured.NestedSlice(obj.Object, "spec", "destinations")
	assert.Equal(t, []interface{}{
		map[string]interface{}{"server": "server1", "namespace": "ns1"},
		map[string]interface{}{"server": "server2", "namespace": "ns2"}}, p2Destinations)
	p2ClusterResourceWhitelist, _, _ := unstructured.NestedSlice(obj.Object, "spec", "clusterResourceWhitelist")
	assert.Equal(t, []interface{}{
		map[string]interface{}{"group": "group1", "kind": "kind1"},
		map[string]interface{}{"group": "group2", "kind": "kind2"}}, p2ClusterResourceWhitelist)
}

func TestReconcileArgoProject(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	project := &v1alpha3.DevOpsProject{}
	project.SetName("fake")

	appProject := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"description": "description",
			},
		},
	}
	appProject.SetAPIVersion("argoproj.io/v1alpha1")
	appProject.SetKind("AppProject")
	appProject.SetName("fake")
	appProject.SetNamespace("fake")

	tests := []struct {
		name    string
		c       client.Client
		project func() *v1alpha3.DevOpsProject
		verify  func(t *testing.T, err error, c client.Client)
	}{{
		name: "no argo in the project",
		c:    fake.NewFakeClientWithScheme(schema),
		project: func() *v1alpha3.DevOpsProject {
			return project.DeepCopy()
		},
		verify: func(t *testing.T, err error, c client.Client) {
			assert.NotNil(t, err, "should not found the Argo AppProject")
		},
	}, {
		name: "have an empty Argo",
		c:    fake.NewFakeClientWithScheme(schema),
		project: func() *v1alpha3.DevOpsProject {
			p1 := project.DeepCopy()
			p1.Spec.Argo = &v1alpha3.Argo{}
			return p1
		},
		verify: func(t *testing.T, err error, c client.Client) {
			assert.Nil(t, err)

			// make sure the Argo AppProject was created
			appProject := appProject.DeepCopy()

			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake",
			}, appProject)
			assert.Nil(t, err)
		},
	}, {
		name: "update the existing Argo AppProject",
		c:    fake.NewFakeClientWithScheme(schema, appProject.DeepCopy()),
		project: func() *v1alpha3.DevOpsProject {
			p1 := project.DeepCopy()
			p1.Spec.Argo = &v1alpha3.Argo{
				Description: "new description",
			}
			return p1
		},
		verify: func(t *testing.T, err error, c client.Client) {
			assert.Nil(t, err)

			// make sure the Argo AppProject was updated
			appProject := appProject.DeepCopy()

			err = c.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake",
			}, appProject)
			assert.Nil(t, err)
			desc, _, _ := unstructured.NestedString(appProject.Object, "spec", "description")
			assert.Equal(t, "new description", desc)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client: tt.c,
			}
			err := r.reconcileArgoProject(tt.project())
			tt.verify(t, err, tt.c)
		})
	}
}
