package app

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/jwt/token"
)

// NewCmd creates a root command for jwt
func NewCmd(k8sClientFactory k8sClientFactory) (cmd *cobra.Command) {
	opt := &jwtOption{
		k8sClientFactory: k8sClientFactory,
	}

	cmd = &cobra.Command{
		Use:     "jwt",
		Short:   "Output the JWT",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.secret, "secret", "s", "",
		"The secret for generating jwt")
	flags.StringVarP(&opt.namespace, "namespace", "", "kubesphere-devops-system",
		"The namespace of target ConfigMap")
	flags.StringVarP(&opt.name, "name", "", "devops-config",
		"The name of target ConfigMap")
	flags.StringVarP(&opt.output, "output", "o", "",
		"The destination of the JWT output. Print to the stdout if it's empty.")
	return
}

type jwtOption struct {
	secret string
	output string

	namespace string
	name      string

	client           kubernetes.Interface
	k8sClientFactory k8sClientFactory
	configMapUpdater configMapUpdater
}

func (o *jwtOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.output == "configmap" || o.secret == "" {
		if o.client, err = o.k8sClientFactory.Get(); err != nil {
			err = fmt.Errorf("cannot create Kubernetes client, error: %v", err)
			return
		}
		o.configMapUpdater = o
	}

	// get secret from ConfigMap if it's empty
	if o.secret == "" {
		if o.secret = o.getSecret(); o.secret == "" {
			// generate a new secret if the ConfigMap does not contain it, then update it into ConfigMap
			o.updateSecret(o.generateSecret())
		}
	}
	return
}

func (o *jwtOption) getSecret() string {
	if cm, err := o.configMapUpdater.GetConfigMap(context.TODO(), o.namespace, o.name); err == nil {
		if data, ok := cm.Data[config.DefaultConfigurationFileName]; ok {
			dataMap := make(map[string]map[string]string, 0)
			if err := yaml.Unmarshal([]byte(data), dataMap); err == nil {
				if _, ok := dataMap["authentication"]; ok {
					return dataMap["authentication"]["jwtSecret"]
				}
			}
		}
	}
	return ""
}

func (o *jwtOption) generateSecret() string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (o *jwtOption) updateSecret(secret string) {
	ctx := context.TODO()
	if cm, err := o.configMapUpdater.GetConfigMap(ctx, o.namespace, o.name); err == nil {
		if data, ok := cm.Data[config.DefaultConfigurationFileName]; ok {
			dataMap := make(map[string]map[string]string, 0)
			if err := yaml.Unmarshal([]byte(data), dataMap); err == nil {
				if _, ok := dataMap["authentication"]; ok {
					dataMap["authentication"]["jwtSecret"] = secret
				} else {
					dataMap["authentication"] = map[string]string{
						"jwtSecret": secret,
					}
				}

				cfg, _ := yaml.Marshal(dataMap)
				cm.Data[config.DefaultConfigurationFileName] = string(cfg)
				_, _ = o.configMapUpdater.UpdateConfigMap(ctx, cm)
			}
		}
	}
}

func (o *jwtOption) runE(cmd *cobra.Command, args []string) (err error) {
	jwt := generateJWT(o.secret)

	switch o.output {
	case "configmap":
		err = o.updateJenkinsToken(jwt, o.namespace, o.name)
	default:
		cmd.Print(jwt)
	}
	return
}

func updateToken(content, token string) string {
	dataMap := make(map[string]map[string]string, 0)
	if err := yaml.Unmarshal([]byte(content), dataMap); err == nil {
		if _, ok := dataMap["devops"]; ok {
			dataMap["devops"]["password"] = token

			if result, err := yaml.Marshal(dataMap); err == nil {
				return strings.TrimSpace(string(result))
			}
		}
	}
	return content
}

func generateJWT(secret string) (jwt string) {
	issuer := token.NewTokenIssuer(secret, 0)
	admin := &user.DefaultInfo{
		Name: "admin",
	}

	jwt, _ = issuer.IssueTo(admin, token.AccessToken, 0)
	return
}
