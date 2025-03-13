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

package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/mock/mhttp"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/jwt/token"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestJenkinsfileReconciler_Reconcile(t *testing.T) {
	schema, err := v1alpha3.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	defaultReq := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "ns",
			Name:      "name",
		},
	}

	pip := &v1alpha3.Pipeline{}
	pip.SetNamespace("ns")
	pip.SetName("name")
	pip.Annotations = map[string]string{
		v1alpha3.PipelineJenkinsfileEditModeAnnoKey: "raw",
	}
	pip.Spec.Type = v1alpha3.NoScmPipelineType
	pip.Spec.Pipeline = &v1alpha3.NoScmPipeline{
		Jenkinsfile: `jenkinsfile`,
	}

	jsonEditModePip := pip.DeepCopy()
	jsonEditModePip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = "json"
	jsonEditModePip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey] = `json`

	invalidEditMode := pip.DeepCopy()
	invalidEditMode.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = "invalid"

	emptyEditMode := pip.DeepCopy()
	emptyEditMode.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = ""

	irregularPip := pip.DeepCopy()
	irregularPip.Spec.Type = ""

	type fields struct {
		Client        client.Client
		log           logr.Logger
		recorder      record.EventRecorder
		JenkinsClient core.Client
		TokenIssuer   token.Issuer
	}
	type args struct {
		req controllerruntime.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		prepare    func(t *testing.T, c *core.Client)
		verify     func(t *testing.T, Client client.Client)
		wantResult controllerruntime.Result
		wantErr    assert.ErrorAssertionFunc
	}{{
		name: "not found",
		fields: fields{
			Client: fake.NewClientBuilder().WithScheme(schema).Build(),
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
		wantResult: controllerruntime.Result{},
	}, {
		name: "invalid edit mode",
		fields: fields{
			Client:        fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(invalidEditMode).Build(),
			JenkinsClient: core.Client{},
			log:           logr.Logger{},
			TokenIssuer:   &token.FakeIssuer{},
		},
		args: args{
			req: defaultReq,
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name: "empty edit mode",
		fields: fields{
			Client:        fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(emptyEditMode).Build(),
			JenkinsClient: core.Client{},
			log:           logr.Logger{},
			TokenIssuer:   &token.FakeIssuer{},
		},
		args: args{
			req: defaultReq,
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name: "irregular pipeline, and jenkinsfile edit mode",
		fields: fields{
			Client:        fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(irregularPip).Build(),
			JenkinsClient: core.Client{},
			log:           logr.Logger{},
		},
		args: args{
			req: defaultReq,
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name: "a regular pipeline with jenkinsfile edit mode",
		fields: fields{
			Client:        fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(pip).Build(),
			JenkinsClient: core.Client{},
			log:           logr.Logger{},
			TokenIssuer:   &token.FakeIssuer{},
		},
		args: args{
			req: defaultReq,
		},
		prepare: func(t *testing.T, c *core.Client) {
			ctrl := gomock.NewController(t)
			roundTripper := mhttp.NewMockRoundTripper(ctrl)
			c.RoundTripper = roundTripper

			core.PrepareForToJSON(roundTripper, "http://localhost", "", "")
		},
		verify: func(t *testing.T, Client client.Client) {
			pip := &v1alpha3.Pipeline{}
			err := Client.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "name",
			}, pip)
			assert.Nil(t, err)
			assert.Equal(t, `{"a":"b"}`, pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey])
			assert.Equal(t, "", pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey])
			assert.Equal(t, v1alpha3.PipelineJenkinsfileValidateSuccess, pip.Annotations[v1alpha3.PipelineJenkinsfileValidateAnnoKey])
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}, {
		name: "a regular pipeline with JSON edit mode",
		fields: fields{
			Client:        fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(jsonEditModePip).Build(),
			JenkinsClient: core.Client{},
			log:           logr.Logger{},
			TokenIssuer:   &token.FakeIssuer{},
		},
		args: args{
			req: defaultReq,
		},
		prepare: func(t *testing.T, c *core.Client) {
			ctrl := gomock.NewController(t)
			roundTripper := mhttp.NewMockRoundTripper(ctrl)
			c.RoundTripper = roundTripper

			core.PrepareForToJenkinsfile(roundTripper, "http://localhost", "", "")
		},
		verify: func(t *testing.T, Client client.Client) {
			pip := &v1alpha3.Pipeline{}
			err := Client.Get(context.Background(), types.NamespacedName{
				Namespace: "ns",
				Name:      "name",
			}, pip)
			assert.Nil(t, err)
			assert.Equal(t, "json", pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey])
			assert.Equal(t, "", pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey])
			assert.Equal(t, "jenkinsfile", pip.Spec.Pipeline.Jenkinsfile)
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt.fields.JenkinsClient)
			}
			r := &JenkinsfileReconciler{
				Client:        tt.fields.Client,
				log:           logr.New(log.NullLogSink{}),
				recorder:      tt.fields.recorder,
				JenkinsClient: tt.fields.JenkinsClient,
			}
			gotResult, err := r.Reconcile(context.Background(), tt.args.req)
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
