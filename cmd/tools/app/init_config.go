/*
Copyright 2023 The KubeSphere Authors.

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

package app

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	jwt "kubesphere.io/devops/cmd/tools/jwt/app"
	"kubesphere.io/devops/pkg/client/devops/jenkins"
	"kubesphere.io/devops/pkg/config"
)

type initConfigOption struct {
	*ToolOption

	ksNamespace string
	ksConfigmap string
}

func (o *initConfigOption) preRunE(cmd *cobra.Command, args []string) error {
	return o.initK8sClient()
}

func (o *initConfigOption) runE(cmd *cobra.Command, args []string) error {
	patchConf, err := o.getPatchConfig()
	if err != nil {
		return err
	}
	if len(patchConf) == 0 {
		klog.Info("the devops-config doesn't need to update, ignore")
		return nil
	}

	klog.Infof("update configmap %s in %s ..", o.Configmap, o.Namespace)
	return o.updateConfig(patchConf)
}

func (o initConfigOption) getKubesphereYaml(namespace, configmapName string) (cm *v1.ConfigMap, ksYaml map[string]interface{}, err error) {
	if cm, err = o.K8sClient.Kubernetes().CoreV1().ConfigMaps(namespace).
		Get(context.TODO(), configmapName, metav1.GetOptions{}); err != nil {
		return
	}

	ksYaml = map[string]interface{}{}
	err = yaml.Unmarshal([]byte(cm.Data[config.DefaultConfigurationFileName]), ksYaml)
	return
}

func (o *initConfigOption) getPatchConfig() (patchConf map[string]interface{}, err error) {
	patchConf = map[string]interface{}{}

	var devopsConf map[string]interface{}
	if _, devopsConf, err = o.getKubesphereYaml(o.Namespace, o.Configmap); err != nil {
		return
	}
	jwtSecret := devopsConf["authentication"].(map[string]interface{})["jwtSecret"].(string)
	if jwtSecret == "" {
		klog.Info("generate jwt secret ..")
		jwtSecret = jwt.GenerateJwtSecret()
		patchConf["authentication"] = map[string]interface{}{"jwtSecret": jwtSecret}
	}
	password := devopsConf["devops"].(map[string]interface{})["password"].(string)
	if password == jenkins.DefaultAdminPassword || password == "" {
		klog.Info("generate devops password by jwt ..")
		password = jwt.GeneratePassword(jwtSecret)
		patchConf["devops"] = map[string]interface{}{"password": password}
	}

	// copy sonarqube from kubesphere-config if exist
	var ksConf map[string]interface{}
	if _, ksConf, err = o.getKubesphereYaml(o.ksNamespace, o.ksConfigmap); err != nil {
		if errors.IsNotFound(err) {
			err = nil
		}
		return
	}
	if sonarqubeObj, exist := ksConf["sonarQube"]; exist {
		// if sonarqube in devops-config is same as that in kubesphere-config, ignore
		var devopsSonarqubeObj interface{}
		if devopsSonarqubeObj, exist = devopsConf["sonarQube"]; exist {
			sonarqube := sonarqubeObj.(map[string]interface{})
			devopsSonarqube := devopsSonarqubeObj.(map[string]interface{})
			if devopsSonarqube["host"] == sonarqube["host"] && devopsSonarqube["token"] == sonarqube["token"] {
				return
			}
		}

		patchConf["sonarQube"] = sonarqubeObj
	}

	return
}

func (o *initConfigOption) updateConfig(updateConfig map[string]interface{}) error {
	devopsCm, devopsKsConf, err := o.getKubesphereYaml(o.Namespace, o.Configmap)
	if err != nil {
		return fmt.Errorf("cannot found ConfigMap %s/%s, %v", o.Namespace, o.Configmap, err)
	}

	patchedKubeSphereConfig, err := patchKubeSphereConfig(devopsKsConf, updateConfig)
	if err != nil {
		return fmt.Errorf("pathc kubesphere config error: %+v", err)
	}
	kubeSphereConfigBytes, err := yaml.Marshal(patchedKubeSphereConfig)
	if err != nil {
		return fmt.Errorf("cannot marshal KubeSphere configuration, %v", err)
	}

	devopsCm.Data["kubesphere.yaml"] = string(kubeSphereConfigBytes)
	_, err = o.K8sClient.Kubernetes().CoreV1().ConfigMaps(o.Namespace).Update(context.TODO(), devopsCm, metav1.UpdateOptions{})
	return err
}

// NewInitCmd creates a root command for init-config
func NewInitCmd() (cmd *cobra.Command) {
	opt := &initConfigOption{
		ToolOption: toolOpt,
	}

	initCmd := &cobra.Command{
		Use:     "init-config",
		Short:   "Initialize configurations for DevOps apiserver and controller-manager",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := initCmd.Flags()
	flags.StringVar(&opt.ksNamespace, "ks-namespace", "kubesphere-system",
		"The namespace of kubesphere core service")
	flags.StringVar(&opt.ksConfigmap, "ks-configmap", "kubesphere-config",
		"The configmap name of kubesphere core service configuration")

	return initCmd
}
