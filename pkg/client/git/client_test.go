package git

import (
	"fmt"
	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
)

func TestGetClient(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	basicSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basicSecret",
			Namespace: "ns",
		},
		Type: v1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			v1.BasicAuthPasswordKey: []byte("token"),
		},
	}
	opaqueSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "opaqueSecret",
			Namespace: "ns",
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			v1.ServiceAccountTokenKey: []byte("token"),
		},
	}
	type fields struct {
		provider  string
		secretRef *v1.SecretReference
		k8sClient client.Client
	}
	type args struct {
		repo *v1alpha1.GitRepository
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantClient *scm.Client
		wantErr    assert.ErrorAssertionFunc
	}{{
		name: "not support git provider",
		args: args{
			repo: &v1alpha1.GitRepository{
				Spec: v1alpha1.GitRepositorySpec{
					Provider: "not-support",
				},
			},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err, i)
			assert.Equal(t, strings.HasPrefix(err.Error(), "not support git provider: "), true, i)
			return true
		},
	}, {
		name: "no secret found",
		fields: fields{
			k8sClient: fake.NewFakeClientWithScheme(schema),
			provider:  "github",
			secretRef: &v1.SecretReference{Namespace: "fake", Name: "fake"},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err, i)
			return true
		},
		wantClient: github.NewDefault(),
	}, {
		name: "github provider",
		fields: fields{
			k8sClient: fake.NewFakeClientWithScheme(schema, basicSecret.DeepCopy()),
			provider:  "github",
			secretRef: &v1.SecretReference{Namespace: "ns", Name: "basicSecret"},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err, i)
			return false
		},
	}, {
		name: "gitlab provider",
		fields: fields{
			k8sClient: fake.NewFakeClientWithScheme(schema, opaqueSecret.DeepCopy()),
			provider:  "gitlab",
			secretRef: &v1.SecretReference{Namespace: "ns", Name: "opaqueSecret"},
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err, i)
			return false
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewClientFactory(tt.fields.provider, tt.fields.secretRef, tt.fields.k8sClient)
			gotClient, err := r.GetClient()
			if !tt.wantErr(t, err, fmt.Sprintf("GetClient() %s", tt.name)) {
				return
			}
			assert.Equalf(t, tt.wantClient, gotClient, fmt.Sprintf("GetClient() %s", tt.name))
		})
	}
}
