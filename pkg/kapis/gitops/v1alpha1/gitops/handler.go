// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitops

import (
	"context"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/api/gitops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
	"github.com/kubesphere/ks-devops/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

var (
	// pathParameterApplication is a path parameter definition for application.
	pathParameterApplication = restful.PathParameter("application", "The application name")
	syncStatusQueryParam     = restful.QueryParameter("syncStatus", `Filter by sync status. Available values: "Unknown", "Synced" and "OutOfSync"`)
	healthStatusQueryParam   = restful.QueryParameter("healthStatus", `Filter by health status. Available values: "Unknown", "Progressing", "Healthy", "Suspended", "Degraded" and "Missing"`)
	cascadeQueryParam        = restful.QueryParameter("cascade",
		"Delete both the app and its resources, rather than only the application if cascade is true").
		DefaultValue("false").DataType("boolean")
)

func (h *Handler) ApplicationList(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	healthStatus := common.GetQueryParameter(req, healthStatusQueryParam)
	syncStatus := common.GetQueryParameter(req, syncStatusQueryParam)

	applicationList := &v1alpha1.ApplicationList{}
	matchingLabels := client.MatchingLabels{}
	if syncStatus != "" {
		matchingLabels[v1alpha1.SyncStatusLabelKey] = syncStatus
	}
	if healthStatus != "" {
		matchingLabels[v1alpha1.HealthStatusLabelKey] = healthStatus
	}
	if err := h.List(context.Background(), applicationList, client.InNamespace(namespace), matchingLabels); err != nil {
		common.Response(req, res, applicationList, err)
		return
	}

	queryParam := query.ParseQueryParameter(req)
	compareFunc := func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*v1alpha1.Application).ObjectMeta, right.(*v1alpha1.Application).ObjectMeta, field)
	}
	filterFunc := func(obj runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(obj.(*v1alpha1.Application).ObjectMeta, filter)
	}
	list := v1alpha3.DefaultList(ToObjects(applicationList.Items), queryParam, compareFunc, filterFunc)

	common.Response(req, res, list, nil)
}

func (h *Handler) GetApplication(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	name := common.GetPathParameter(req, pathParameterApplication)

	application := &v1alpha1.Application{}
	err := h.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, application)
	common.Response(req, res, application, err)
}

func (h *Handler) DelApplication(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	name := common.GetPathParameter(req, pathParameterApplication)
	cascade := common.GetQueryParameter(req, cascadeQueryParam)

	ctx := context.Background()
	application := &v1alpha1.Application{}
	objectKey := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := h.Get(ctx, objectKey, application)
	if err == nil {
		if argo := application.Spec.ArgoApp; argo != nil {
			if cascade == "true" {
				if k8sutil.AddFinalizer(&application.ObjectMeta, v1alpha1.ArgoCDResourcesFinalizer) {
					if err = h.Update(ctx, application); err != nil {
						common.Response(req, res, application, err)
						return
					}

					if err = h.Get(ctx, objectKey, application); err != nil {
						common.Response(req, res, application, err)
						return
					}
				}
			}
		} else {
			klog.Errorf("Application %s in namespace %s is not type argocd", name, namespace)
		}
		err = h.Delete(ctx, application)
	}
	common.Response(req, res, application, err)
}

func (h *Handler) UpdateApplication(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	name := common.GetPathParameter(req, pathParameterApplication)

	var err error
	application := &v1alpha1.Application{}
	if err = req.ReadEntity(application); err == nil {
		latestApp := &v1alpha1.Application{}
		err = h.Get(context.Background(), types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}, latestApp)
		if err == nil {
			application.ResourceVersion = latestApp.ResourceVersion
			err = h.Update(context.Background(), application)
		}
	}
	common.Response(req, res, application, err)
}

type Handler struct {
	client.Client
}

func NewHandler(options *common.Options) *Handler {
	return &Handler{
		Client: options.GenericClient,
	}
}
