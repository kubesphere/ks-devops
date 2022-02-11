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

package argocd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"html/template"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=devopsprojects,verbs=get;list;update
//+kubebuilder:rbac:groups=argoproj.io,resources=appprojects,verbs=get;list;create;update

// Reconciler is the reconciler of the DevOpsProject with Argo AppProject
type Reconciler struct {
	client.Client
	log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile makes sure the ArgoAppProject can be maintained which comes from the DevOpsProject
func (r *Reconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	r.log.Info(fmt.Sprintf("start to reconcile project: %s", req.String()))

	project := &v1alpha3.DevOpsProject{}
	if err = r.Client.Get(ctx, req.NamespacedName, project); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if argo := project.Spec.Argo; argo != nil {
		// we only handle the project that have the Argo settings
		err = r.reconcileArgoProject(project)
	}
	return
}

func (r *Reconciler) reconcileArgoProject(project *v1alpha3.DevOpsProject) (err error) {
	ctx := context.Background()

	argoAppProject := &unstructured.Unstructured{}
	argoAppProject.SetKind("AppProject")
	argoAppProject.SetAPIVersion("argoproj.io/v1alpha1")

	if err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: project.Name,
		Name:      project.Name,
	}, argoAppProject); err != nil {
		if !apierrors.IsNotFound(err) {
			return
		}

		if argoAppProject, err = createUnstructuredObject(project); err != nil {
			return
		}

		err = r.Client.Create(ctx, argoAppProject)
	} else {
		var newProject *unstructured.Unstructured
		if newProject, err = createUnstructuredObject(project); err == nil {
			argoAppProject.Object["spec"] = newProject.Object["spec"]
			setOwnerRef(argoAppProject, project)
			err = r.Client.Update(ctx, argoAppProject)
		}
	}
	return
}

func createUnstructuredObject(project *v1alpha3.DevOpsProject) (result *unstructured.Unstructured, err error) {
	project = project.DeepCopy()
	if project.Spec.Argo == nil {
		err = fmt.Errorf("no argo found from the spec")
		return
	}

	var tpl *template.Template
	if tpl, err = template.New("argo").Parse(argoProjectTemplate); err != nil {
		return
	}

	// set some default values
	argo := project.Spec.Argo
	if len(argo.SourceRepos) == 0 {
		argo.SourceRepos = []string{"*"}
	}
	if len(argo.Destinations) == 0 {
		argo.Destinations = []v1alpha3.ApplicationDestination{{
			Server:    "*",
			Namespace: "*",
		}}
	}
	if len(argo.ClusterResourceWhitelist) == 0 {
		argo.ClusterResourceWhitelist = []metav1.GroupKind{{
			Group: "*",
			Kind:  "*",
		}}
	}

	buffer := new(bytes.Buffer)
	if err = tpl.Execute(buffer, argo); err == nil {
		if result, err = GetObjectFromYaml(buffer.String()); err == nil {
			result.SetName(project.GetName())
			result.SetNamespace(project.GetName())
			setOwnerRef(result, project)
		}
	}
	return
}

func setOwnerRef(object metav1.Object, project *v1alpha3.DevOpsProject) {
	k8sutil.SetOwnerReference(object, metav1.OwnerReference{
		Kind:       project.Kind,
		Name:       project.Name,
		APIVersion: project.APIVersion,
		UID:        project.UID,
	})
}

const argoProjectTemplate = `apiVersion: argoproj.io/v1alpha1
kind: AppProject
spec:
  description: "{{.Description}}"
  sourceRepos:
{{- range .SourceRepos}}
  - "{{.}}"
{{- end }}
  destinations:
{{range $val := .Destinations}}
  - namespace: '{{$val.Namespace}}'
    server: '{{$val.Server}}'
{{end}}
  clusterResourceWhitelist:
{{range $val := .ClusterResourceWhitelist}}
  - group: '{{$val.Group}}'
    kind: '{{$val.Kind}}'
{{end}}
`

// GetObjectFromYaml returns the Unstructured object from a YAML
func GetObjectFromYaml(yamlText string) (obj *unstructured.Unstructured, err error) {
	obj = &unstructured.Unstructured{}
	err = yaml.Unmarshal([]byte(yamlText), obj)
	return
}

// GetName returns the name of this reconciler
func (r *Reconciler) GetName() string {
	return "ArgoCDProjectReconciler"
}

// GetGroupName returns the group name
func (r *Reconciler) GetGroupName() string {
	return "argocd"
}

// SetupWithManager setups the reconciler with a manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha3.DevOpsProject{}).
		Complete(r)
}
