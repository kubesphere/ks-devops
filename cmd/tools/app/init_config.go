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
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/config"
)

type initConfigOption struct {
	*ToolOption

	ksNamespace     string
	ksConfigMapName string
}

func (o *initConfigOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	o.K8sClient, err = k8s.NewKubernetesClient(k8s.NewKubernetesOptions())
	return
}

func (o *initConfigOption) runE(cmd *cobra.Command, args []string) (err error) {
	// if kubesphere-config in namespace kubesphere-system not exist, generate devops.password by jwt
	// used for deploy devops independence
	if _, _, err = o.getKubesphereConfig(o.ksNamespace, o.ksConfigMapName); err != nil {
		if !errors.IsNotFound(err) {
			klog.Error("check if kubesphere-config exist failed")
			return
		}
		klog.Infof("update configmap %s in %s by jwt", o.ConfigMapName, o.Namespace)
		err = jwt.JwtFunc("", o.Namespace, o.ConfigMapName)
		return
	}

	klog.Infof("get patch configuration from configmap %s in %s", o.ksConfigMapName, o.ksNamespace)
	var updateConfig map[string]interface{}
	if updateConfig, err = o.getPatchConfigFromKs(); err != nil {
		return
	}
	if len(updateConfig) == 0 {
		klog.Info("nothing need to update in configmap, ignore")
		return
	}

	klog.Infof("update configmap %s in %s ..", o.ConfigMapName, o.Namespace)
	err = o.updateKubeSphereConfig(updateConfig)
	return
}

func (o initConfigOption) getKubesphereConfig(namespace, configmapName string) (cm *v1.ConfigMap, ksYaml map[string]interface{}, err error) {
	if cm, err = o.K8sClient.Kubernetes().CoreV1().ConfigMaps(namespace).
		Get(context.TODO(), configmapName, metav1.GetOptions{}); err != nil {
		return
	}

	ksYaml = map[string]interface{}{}
	err = yaml.Unmarshal([]byte(cm.Data[config.DefaultConfigurationFileName]), ksYaml)
	return
}

func (o *initConfigOption) getPatchConfigFromKs() (conf map[string]interface{}, err error) {
	var ksConf, devopsConf map[string]interface{}
	if _, ksConf, err = o.getKubesphereConfig(o.ksNamespace, o.ksConfigMapName); err != nil {
		return
	}
	if _, devopsConf, err = o.getKubesphereConfig(o.Namespace, o.ConfigMapName); err != nil {
		return
	}

	conf = map[string]interface{}{}

	password := ksConf["devops"].(map[string]interface{})["password"].(string)
	devopsPassword := devopsConf["devops"].(map[string]interface{})["password"].(string)
	if password == "" {
		err = fmt.Errorf("the password in configmap %s is nil", o.ksConfigMapName)
		return
	}
	if devopsPassword != password {
		conf["devops"] = map[string]interface{}{"password": password}
	}

	if sonarqube, exist := ksConf["sonarQube"]; exist {
		conf["sonarQube"] = sonarqube
	}
	return
}

func (o *initConfigOption) updateKubeSphereConfig(updateConfig map[string]interface{}) error {
	devopsCm, devopsKsConf, err := o.getKubesphereConfig(o.Namespace, o.ConfigMapName)
	if err != nil {
		return fmt.Errorf("cannot found ConfigMap %s/%s, %v", o.Namespace, o.ConfigMapName, err)
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
		"The namespace of kubesphere namespace")
	flags.StringVar(&opt.ksConfigMapName, "ks-configmap", "kubesphere-config",
		"The name of kubesphere configmap")

	return initCmd
}
