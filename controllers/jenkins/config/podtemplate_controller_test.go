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

package config

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	mgrcore "kubesphere.io/devops/controllers/core"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestPodTemplateReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	tests := []struct {
		name    string
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "normal",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PodTemplateReconciler{}
			mgr := &mgrcore.FakeManager{
				Scheme: schema,
			}
			tt.wantErr(t, r.SetupWithManager(mgr), fmt.Sprintf("SetupWithManager(%v)", mgr))
			assert.NotEmpty(t, r.LabelSelector)
			assert.NotEmpty(t, r.TargetConfigMapName)
			assert.NotEmpty(t, r.TargetConfigMapNamespace)
			assert.NotEmpty(t, r.TargetConfigMapKey)
		})
	}
}

func TestPodTemplateReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	req := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "ns",
			Name:      "pod-template",
		},
	}

	now := metav1.Now()
	podT := v1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
		},
	}
	deletingPodT := podT.DeepCopy()
	deletingPodT.DeletionTimestamp = &now

	cascData, err := ioutil.ReadFile("testdata/casc.yaml")
	assert.Nil(t, err)

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kubesphere-devops-system",
			Name:      "jenkins-casc-config",
		},
		Data: map[string]string{
			"jenkins_user.yaml": string(cascData),
		},
	}
	cmWithoutKey := cm.DeepCopy()
	cmWithoutKey.Data = map[string]string{}

	type fields struct {
		Client client.Client
	}
	type args struct {
		req controllerruntime.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult controllerruntime.Result
		wantErr    assert.ErrorAssertionFunc
		verify     func(*testing.T, client.Client)
	}{{
		name:   "not found podtemplate",
		fields: fields{Client: fake.NewFakeClientWithScheme(schema)},
		args:   args{req: req},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name:   "no related ConfigMap exist",
		fields: fields{Client: fake.NewFakeClientWithScheme(schema, podT.DeepCopy())},
		args:   args{req: req},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name:   "no expect key in specific ConfigMap",
		fields: fields{Client: fake.NewFakeClientWithScheme(schema, podT.DeepCopy(), cmWithoutKey.DeepCopy())},
		args:   args{req: req},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name:   "normal case",
		fields: fields{Client: fake.NewFakeClientWithScheme(schema, podT.DeepCopy(), cm.DeepCopy())},
		args:   args{req: req},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
		verify: func(t *testing.T, c client.Client) {
			podT := v1.PodTemplate{}
			err := c.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "pod-template",
			}, &podT)
			assert.Nil(t, err)
			assert.NotEmpty(t, podT.Finalizers)
		},
	}, {
		name:   "handle a deleting podTemplate",
		fields: fields{Client: fake.NewFakeClientWithScheme(schema, deletingPodT.DeepCopy(), cm.DeepCopy())},
		args:   args{req: req},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
		verify: func(t *testing.T, c client.Client) {
			podT := v1.PodTemplate{}
			err := c.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "pod-template",
			}, &podT)
			assert.Nil(t, err)
			assert.Empty(t, podT.Finalizers)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PodTemplateReconciler{
				Client: tt.fields.Client,
				log:    log.NullLogger{},
			}
			mgr := &mgrcore.FakeManager{
				Scheme: schema,
			}
			err = r.SetupWithManager(mgr)
			assert.Nil(t, err)
			gotResult, err := r.Reconcile(tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("Reconcile(%v)", tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.wantResult, gotResult, "Reconcile(%v)", tt.args.req)
			if tt.verify != nil {
				tt.verify(t, tt.fields.Client)
			}
		})
	}
}
