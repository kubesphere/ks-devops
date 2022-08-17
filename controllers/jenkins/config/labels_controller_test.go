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
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/mock/mhttp"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	mgrcore "kubesphere.io/devops/controllers/core"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/jwt/token"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func Test_setLabelsToConfigMap(t *testing.T) {
	emptyConfigMap := &v1.ConfigMap{}

	oneItemConfigMap := &v1.ConfigMap{Data: map[string]string{"name": "good"}}

	twoItemsConfigMap := &v1.ConfigMap{
		Data: map[string]string{
			"name": "good,bad",
		},
	}

	type args struct {
		labels []string
		cm     *v1.ConfigMap
	}
	tests := []struct {
		name        string
		args        args
		wantChanged bool
	}{{
		name: "set labels to an empty",
		args: args{
			labels: []string{"good"},
			cm:     emptyConfigMap.DeepCopy(),
		},
		wantChanged: true,
	}, {
		name: "one item in the ConfigMap",
		args: args{
			labels: []string{"good"},
			cm:     oneItemConfigMap.DeepCopy(),
		},
		wantChanged: true,
	}, {
		name: "different ordered items",
		args: args{
			labels: []string{"bad", "good"},
			cm:     twoItemsConfigMap.DeepCopy(),
		},
		wantChanged: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantChanged, setLabelsToConfigMap(tt.args.labels, tt.args.cm), "setLabelsToConfigMap(%v, %v)", tt.args.labels, tt.args.cm)
		})
	}
}

func TestAgentLabelsReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	defaultReq := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "ns",
			Name:      "fake",
		},
	}

	cmWithoutLabels := &v1.ConfigMap{}
	cmWithoutLabels.Namespace = "ns"
	cmWithoutLabels.Name = "fake"

	cmWithLabels := cmWithoutLabels.DeepCopy()
	cmWithLabels.Data = map[string]string{
		"agent.labels": "a,b",
	}

	type fields struct {
		TargetNamespace string
		JenkinsClient   core.JenkinsCore
		TokenIssuer     token.Issuer
		targetName      string
		Client          client.Client
		recorder        record.EventRecorder
	}
	type args struct {
		req controllerruntime.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		prepare    func(t *testing.T, c *core.JenkinsCore)
		wantResult controllerruntime.Result
		wantErr    assert.ErrorAssertionFunc
	}{{
		name: "not found ConfigMap",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema),
			JenkinsClient: core.JenkinsCore{
				URL: "http://localhost",
			},
			TokenIssuer:     &token.FakeIssuer{},
			targetName:      "fake",
			TargetNamespace: "ns",
		},
		prepare: func(t *testing.T, c *core.JenkinsCore) {
			ctrl := gomock.NewController(t)
			roundTripper := mhttp.NewMockRoundTripper(ctrl)
			c.RoundTripper = roundTripper

			core.PrepareForToGetLabels(roundTripper, "http://localhost", "", "")
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err)
			return true
		},
		args: args{req: defaultReq},
	}, {
		name: "have the specific ConfigMap",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, cmWithoutLabels),
			JenkinsClient: core.JenkinsCore{
				URL: "http://localhost",
			},
			TokenIssuer:     &token.FakeIssuer{},
			targetName:      "fake",
			TargetNamespace: "ns",
		},
		prepare: func(t *testing.T, c *core.JenkinsCore) {
			ctrl := gomock.NewController(t)
			roundTripper := mhttp.NewMockRoundTripper(ctrl)
			c.RoundTripper = roundTripper

			core.PrepareForToGetLabels(roundTripper, "http://localhost", "", "")
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
		wantResult: controllerruntime.Result{RequeueAfter: time.Minute * 5},
		args:       args{req: defaultReq},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt.fields.JenkinsClient)
			}
			r := &AgentLabelsReconciler{
				TargetNamespace: tt.fields.TargetNamespace,
				JenkinsClient:   tt.fields.JenkinsClient,
				TokenIssuer:     tt.fields.TokenIssuer,
				targetName:      tt.fields.targetName,
				Client:          tt.fields.Client,
				log:             logr.Logger{},
				recorder:        tt.fields.recorder,
			}
			err = r.SetupWithManager(&mgrcore.FakeManager{
				Scheme: schema,
				Client: tt.fields.Client,
			})
			assert.Nil(t, err)
			gotResult, err := r.Reconcile(context.Background(), tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("Reconcile(%v)", tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.wantResult, gotResult, "Reconcile(%v)", tt.args.req)
		})
	}
}

func TestAgentLabelsReconciler_SetupWithManager(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	type fields struct {
		TargetNamespace string
		JenkinsClient   core.JenkinsCore
		TokenIssuer     token.Issuer
		targetName      string
		Client          client.Client
		log             logr.Logger
		recorder        record.EventRecorder
	}
	tests := []struct {
		name    string
		fields  fields
		verify  func(*testing.T, *AgentLabelsReconciler)
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "target name is empty",
		fields: fields{
			TargetNamespace: "ns",
			targetName:      "",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
		verify: func(t *testing.T, c *AgentLabelsReconciler) {
			assert.Equal(t, "jenkins-agent-config", c.targetName)
		},
	}, {
		name: "target namespace is empty",
		fields: fields{
			TargetNamespace: "",
			targetName:      "",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err)
			return true
		},
		verify: func(t *testing.T, c *AgentLabelsReconciler) {
			assert.Equal(t, "jenkins-agent-config", c.targetName)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AgentLabelsReconciler{
				TargetNamespace: tt.fields.TargetNamespace,
				JenkinsClient:   tt.fields.JenkinsClient,
				TokenIssuer:     tt.fields.TokenIssuer,
				targetName:      tt.fields.targetName,
				Client:          tt.fields.Client,
				log:             tt.fields.log,
				recorder:        tt.fields.recorder,
			}
			mgr := &mgrcore.FakeManager{
				Client: tt.fields.Client,
				Scheme: schema,
			}
			tt.wantErr(t, r.SetupWithManager(mgr), fmt.Sprintf("SetupWithManager(%v)", mgr))
			if tt.verify != nil {
				tt.verify(t, r)
			}
		})
	}
}

func Test_getSpecificConfigMapPredicate(t *testing.T) {
	type args struct {
		name            string
		namespace       string
		targetName      string
		targetNamespace string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "normal",
		args: args{
			name:            "name",
			namespace:       "ns",
			targetName:      "name",
			targetNamespace: "ns",
		},
		want: true,
	}, {
		name: "different name",
		args: args{
			name:            "name",
			namespace:       "ns",
			targetName:      "name-1",
			targetNamespace: "ns-1",
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcs := getSpecificConfigMapPredicate(tt.args.name, tt.args.namespace)
			result := funcs.Generic(event.GenericEvent{
				Object: &v1.ConfigMap{
					ObjectMeta: v12.ObjectMeta{
						Name:      tt.args.targetName,
						Namespace: tt.args.targetNamespace,
					},
				},
			})
			assert.Equal(t, tt.want, result)
		})
	}
}
