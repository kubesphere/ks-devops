package addon

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_getStrategyName(t *testing.T) {
	type args struct {
		operatorName string
		kind         string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "have upper case",
		args: args{
			operatorName: "Name",
			kind:         "Kind",
		},
		want: "kind-name",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStrategyName(tt.args.operatorName, tt.args.kind); got != tt.want {
				t.Errorf("getStrategyName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperatorCRDReconciler_operatorsHandle(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	type fields struct {
		Client client.Client
		log    logr.Logger
	}
	type args struct {
		name    string
		version string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		verify  func(t *testing.T, c client.Client)
	}{{
		name: "not support operator",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema),
			log:    logr.Discard(),
		},
		args: args{
			name:    "fake",
			version: "devops.kubesphere.io/v1alpha1",
		},
		wantErr: false,
		verify:  func(t *testing.T, c client.Client) {},
	}, {
		name: "normal case",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema),
			log:    logr.Discard(),
		},
		args: args{
			name:    "ReleaserController",
			version: "devops.kubesphere.io/v1alpha1",
		},
		wantErr: false,
		verify: func(t *testing.T, c client.Client) {
			result := &v1alpha1.AddonStrategy{}
			err := c.Get(context.TODO(), types.NamespacedName{
				Name: "simple-operator-releasercontroller",
			}, result)
			assert.Nil(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, v1alpha1.AddonInstallStrategy("simple-operator"), result.Spec.Type)
			assert.Equal(t, "ReleaserController", result.Spec.SimpleOperator.Kind)
			assert.Equal(t, "devops.kubesphere.io/v1alpha1", result.Spec.SimpleOperator.APIVersion)
		},
	}, {
		name: "update the existing",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, &v1alpha1.AddonStrategy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "simple-operator-releasercontroller",
				},
				Spec: v1alpha1.AddStrategySpec{
					Type: v1alpha1.AddonInstallStrategySimpleOperator,
					SimpleOperator: v1.ObjectReference{
						Kind:       "ReleaserController",
						APIVersion: "devops.kubesphere.io/v1",
					},
				},
			}),
			log: logr.Discard(),
		},
		args: args{
			name:    "ReleaserController",
			version: "devops.kubesphere.io/v1alpha1",
		},
		wantErr: false,
		verify: func(t *testing.T, c client.Client) {
			result := &v1alpha1.AddonStrategy{}
			err := c.Get(context.TODO(), types.NamespacedName{
				Name: "simple-operator-releasercontroller",
			}, result)
			assert.Nil(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Spec.Available)
			assert.Equal(t, v1alpha1.AddonInstallStrategy("simple-operator"), result.Spec.Type)
			assert.Equal(t, "ReleaserController", result.Spec.SimpleOperator.Kind)
			assert.Equal(t, "devops.kubesphere.io/v1alpha1", result.Spec.SimpleOperator.APIVersion)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OperatorCRDReconciler{
				Client: tt.fields.Client,
				log:    tt.fields.log,
			}
			if err := r.operatorsHandle(tt.args.name, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("operatorsHandle() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.verify(t, tt.fields.Client)
		})
	}
}

func Test_operatorSupport(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name        string
		args        args
		wantSupport bool
	}{{
		name: "supported: ReleaserController",
		args: args{
			name: "ReleaserController",
		},
		wantSupport: true,
	}, {
		name: "supported: fake",
		args: args{
			name: "fake",
		},
		wantSupport: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantSupport, operatorSupport(tt.args.name), "operatorSupport(%v)", tt.args.name)
		})
	}
}
