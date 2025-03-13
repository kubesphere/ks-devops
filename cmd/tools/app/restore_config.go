/*
Copyright 2024 The KubeSphere Authors.

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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	SysNs    = "kubesphere-devops-system"
	WorkerNs = "kubesphere-devops-worker"

	AgentCM = "ks-devops-agent"
	CascCM  = "jenkins-casc-config"

	ConfigMapBackupFmt = "%s-backup"
)

// only used to upgrade 3.5.x to 4.1.x

type restoreOption struct {
	kubeConfig string
	client     runtimeclient.WithWatch
}

func (r *restoreOption) preRunE(cmd *cobra.Command, args []string) error {
	var err error
	if r.client, err = NewRuntimeClient(r.kubeConfig); err != nil {
		return err
	}
	return nil
}

func (r *restoreOption) runE(cmd *cobra.Command, args []string) error {
	klog.Infof("restore configmaps ..")
	if err := r.restoreConfigmaps(context.Background()); err != nil {
		klog.Errorf("restore configmaps error: %+v", err)
		return err
	}
	return nil
}

func (r *restoreOption) restoreConfigmaps(ctx context.Context) (err error) {
	key := types.NamespacedName{
		Namespace: SysNs,
		Name:      fmt.Sprintf(ConfigMapBackupFmt, CascCM),
	}

	klog.Infof("restore configmap %s..", CascCM)
	cascCm := new(corev1.ConfigMap)
	oldCascCm := new(corev1.ConfigMap)
	if err = r.client.Get(ctx, key, oldCascCm); err == nil {
		if cascCm, err = getConfigmapWithWatch(ctx, r.client, SysNs, CascCM); err != nil {
			return
		}
		if err = mergeCascConfigmap(cascCm, oldCascCm); err != nil {
			return
		}
		if err = r.client.Update(ctx, cascCm); err != nil {
			return
		}
		// delete backup configmap after update configmap
		klog.Infof("delete old configmap %s ..", oldCascCm.Name)
		err = r.client.Delete(ctx, oldCascCm)
	} else {
		if !apierrors.IsNotFound(err) {
			return
		}
		klog.Infof("the old configmap %s not exist, ignore", key.Name)
	}

	klog.Infof("restore configmap %s..", AgentCM)
	// all configmaps backup in namespace kubesphere-devops-system
	agentCm := new(corev1.ConfigMap)
	oldAgentCm := new(corev1.ConfigMap)
	key.Name = fmt.Sprintf(ConfigMapBackupFmt, AgentCM)
	if err = r.client.Get(ctx, key, oldAgentCm); err == nil {
		if agentCm, err = getConfigmapWithWatch(ctx, r.client, WorkerNs, AgentCM); err != nil {
			return
		}
		agentCm.Data = oldAgentCm.Data
		if err = r.client.Update(ctx, agentCm); err != nil {
			return
		}
		// delete backup configmap after update configmap
		klog.Infof("delete old configmap %s ..", oldAgentCm.Name)
		err = r.client.Delete(ctx, oldAgentCm)
	}
	if apierrors.IsNotFound(err) {
		klog.Infof("the old configmap %s not exist, ignore", key.Name)
		err = nil
	}
	return
}

func mergeCascConfigmap(cm, oldCm *corev1.ConfigMap) (err error) {
	// get securityRealm(OIDC) from new configmap jenkins-casc-config
	var jenkinsJson, jenkinsYaml []byte
	if jenkinsJson, err = yaml.YAMLToJSON([]byte(cm.Data["jenkins.yaml"])); err != nil {
		return
	}
	securityRealm := gjson.GetBytes(jenkinsJson, "jenkins.securityRealm")
	if !securityRealm.Exists() {
		err = errors.New("the jenkins.securityRealm not exist in jenkins-casc-config")
		return
	}

	var securityRealmMap = map[string]map[string]interface{}{}
	if err = yaml.Unmarshal([]byte(securityRealm.Raw), &securityRealmMap); err != nil {
		return
	}

	// setup securityRealm(OIDC) into jenkinsYaml
	if jenkinsJson, err = yaml.YAMLToJSON([]byte(oldCm.Data["jenkins.yaml"])); err != nil {
		return
	}
	if jenkinsJson, err = sjson.SetBytes(jenkinsJson, "jenkins.securityRealm", securityRealmMap); err != nil {
		return
	}
	// delete kubespheretokenauthglobalconfiguration
	if jenkinsJson, err = sjson.DeleteBytes(jenkinsJson, "unclassified.kubespheretokenauthglobalconfiguration"); err != nil {
		return
	}
	if jenkinsYaml, err = yaml.JSONToYAML(jenkinsJson); err != nil {
		return
	}
	cm.Data["jenkins.yaml"] = string(jenkinsYaml)

	// setup securityRealm(OIDC) into jenkinsUserYaml
	if jenkinsJson, err = yaml.YAMLToJSON([]byte(oldCm.Data["jenkins_user.yaml"])); err != nil {
		return
	}
	if jenkinsJson, err = sjson.SetBytes(jenkinsJson, "jenkins.securityRealm", securityRealmMap); err != nil {
		return
	}
	// delete kubespheretokenauthglobalconfiguration
	if jenkinsJson, err = sjson.DeleteBytes(jenkinsJson, "unclassified.kubespheretokenauthglobalconfiguration"); err != nil {
		return
	}
	if jenkinsYaml, err = yaml.JSONToYAML(jenkinsJson); err != nil {
		return
	}
	cm.Data["jenkins_user.yaml"] = string(jenkinsYaml)
	return
}

// NewRestoreCmd creates a root command for restore-config
// only used to restore configmaps(jenkins-casc-config and ks-devops-agent) in post-job when upgrade devops from 3.5.x to 4.1.x
func NewRestoreCmd(kubeconfig string) (cmd *cobra.Command) {
	opt := &restoreOption{
		kubeConfig: kubeconfig,
	}

	cmd = &cobra.Command{
		Use:     "restore-configmap",
		Short:   "restore configmaps(jenkins-casc-config and ks-devops-agent) when upgrade devops",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}
	return cmd
}
