package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/devops/cmd/tools/jwt/app/mock_app"
	"kubesphere.io/devops/pkg/config"
	"testing"
)

var _ = Describe("Password util test", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		updater *mock_app.MockconfigMapUpdater
		ns      string
		name    string
		jwt     string
		err     error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		updater = mock_app.NewMockconfigMapUpdater(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("cannot find configmap", func() {
		JustBeforeEach(func() {
			updater.EXPECT().
				GetConfigMap(ctx, ns, name).
				Return(&v1.ConfigMap{}, errors.New("fake"))

			opt := jwtOption{
				configMapUpdater: updater,
			}
			err = opt.updateJenkinsToken(jwt, ns, name)
		})

		It("cannot find configmap", func() {
			Expect(err).NotTo(BeNil())
		})
	})

	Context("no config in configmap", func() {
		JustBeforeEach(func() {
			updater.EXPECT().
				GetConfigMap(ctx, ns, name).
				Return(&v1.ConfigMap{}, nil)

			opt := jwtOption{
				configMapUpdater: updater,
			}
			err = opt.updateJenkinsToken(jwt, ns, name)
		})

		It("should return error", func() {
			Expect(err).NotTo(BeNil())
		})
	})

	Context("has config in configmap", func() {
		JustBeforeEach(func() {
			updater.EXPECT().
				GetConfigMap(ctx, ns, name).
				Return(&v1.ConfigMap{
					Data: map[string]string{
						config.DefaultConfigurationFileName: "name: rick",
					},
				}, nil)
			updater.EXPECT().
				UpdateConfigMap(ctx, &v1.ConfigMap{
					Data: map[string]string{
						config.DefaultConfigurationFileName: "name: rick",
					},
				}).Return(nil, nil)

			opt := jwtOption{
				configMapUpdater: updater,
			}
			err = opt.updateJenkinsToken(jwt, ns, name)
		})

		It("should success", func() {
			Expect(err).To(BeNil())
		})
	})

	Context("has correct config in configmap", func() {
		var result *v1.ConfigMap
		var resultErr error

		JustBeforeEach(func() {
			jwt = "token"
			updater.EXPECT().
				GetConfigMap(ctx, ns, name).
				Return(&v1.ConfigMap{
					Data: map[string]string{
						config.DefaultConfigurationFileName: `devops:
password: xxx`,
					},
				}, nil)
			updater.EXPECT().
				UpdateConfigMap(ctx, &v1.ConfigMap{
					Data: map[string]string{
						config.DefaultConfigurationFileName: `devops:
password: xxx`,
					},
				}).Return(nil, nil)
			updater.EXPECT().
				GetConfigMap(ctx, ns, name).
				Return(&v1.ConfigMap{
					Data: map[string]string{
						config.DefaultConfigurationFileName: fmt.Sprintf(`devops:
password: %s`, jwt),
					},
				}, nil)

			opt := jwtOption{
				configMapUpdater: updater,
			}
			err = opt.updateJenkinsToken(jwt, ns, name)
			result, resultErr = opt.configMapUpdater.GetConfigMap(ctx, ns, name)
		})

		It("should success", func() {
			Expect(err).To(BeNil())
			Expect(resultErr).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(result.Data[config.DefaultConfigurationFileName]).To(Equal(fmt.Sprintf(`devops:
password: %s`, jwt)))
		})
	})
})

func TestConfigUpdater(t *testing.T) {
	opt := jwtOption{
		client: fake.NewSimpleClientset(&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "name",
			},
			Data: map[string]string{
				config.DefaultConfigurationFileName: "name: rick",
			},
		}),
	}

	ctx := context.TODO()
	cm, err := opt.GetConfigMap(ctx, "ns", "name")
	assert.Nil(t, err)
	assert.Equal(t, cm.Namespace, "ns")
	assert.Equal(t, cm.Name, "name")

	_, err = opt.UpdateConfigMap(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
		},
	})
	assert.Nil(t, err)
}
