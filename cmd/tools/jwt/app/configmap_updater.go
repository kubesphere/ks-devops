package app

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/config"
)

type configMapUpdater interface {
	GetConfigMap(ctx context.Context, ns, name string) (*corev1.ConfigMap, error)
	UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
}

func (o *jwtOption) GetConfigMap(ctx context.Context, ns, name string) (*corev1.ConfigMap, error) {
	return o.client.CoreV1().ConfigMaps(ns).Get(ctx, name, v1.GetOptions{})
}

func (o *jwtOption) UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return o.client.CoreV1().ConfigMaps(configMap.Namespace).Update(ctx, configMap, v1.UpdateOptions{})
}

func (o *jwtOption) updateJenkinsToken(jwt, ns, configMapName string) (err error) {
	ctx := context.TODO()
	var configMap *corev1.ConfigMap
	if configMap, err = o.configMapUpdater.GetConfigMap(ctx, ns, configMapName); err != nil {
		err = fmt.Errorf("cannot find ConfigMap %s/%s, error: %v", ns, configMapName, err)
		return
	}

	var cfg string
	if content, ok := configMap.Data[config.DefaultConfigurationFileName]; !ok || content == "" {
		err = fmt.Errorf("no %s found", config.DefaultConfigurationFileName)
		return
	} else {
		cfg = content
	}

	cfg = updateToken(cfg, jwt, o.overrideJenkinsToken)

	configMap.Data[config.DefaultConfigurationFileName] = cfg
	_, err = o.configMapUpdater.UpdateConfigMap(ctx, configMap)
	return
}
