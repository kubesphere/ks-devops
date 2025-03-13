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
	"fmt"
	"path"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/homedir"
)

// oldCm: customize jenkins-casc-config in 3.5.x
// cm: jenkins-casc-config with securityRealm.oic in 4.1.0
func TestMergeCascConfigmap(t *testing.T) {
	client, err := NewRuntimeClient(path.Join(homedir.HomeDir(), ".kube/config"))
	if err != nil {
		t.Fatal(err)
	}

	backupKey := fmt.Sprintf(ConfigMapBackupFmt, CascCM)
	oldCm := &corev1.ConfigMap{}
	cm := &corev1.ConfigMap{}
	key := types.NamespacedName{
		Namespace: SysNs,
		Name:      backupKey,
	}
	ctx := context.Background()
	t.Logf("get backup confmap %s", backupKey)
	if err = client.Get(ctx, key, oldCm); err == nil {
		t.Logf("get confmap %s", CascCM)
		if cm, err = getConfigmapWithWatch(ctx, client, SysNs, CascCM); err != nil {
			t.Fatalf("error: %+v", err)
		}
		t.Log("merge configmaps")
		if err = mergeCascConfigmap(cm, oldCm); err != nil {
			t.Fatalf("error: %+v", err)
		}
		if err = client.Update(ctx, cm); err != nil {
			t.Fatalf("error: %+v", err)
		}
	}
	if err != nil {
		if errors.IsNotFound(err) {
			t.Logf("the configmap %s not exist, ignore", backupKey)
		} else {
			t.Fatalf("error: %+v", err)
		}
	}
}
