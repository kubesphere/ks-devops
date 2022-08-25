package argocd

import (
	"context"
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/user"
	utilretry "k8s.io/client-go/util/retry"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	apiserverrequest "kubesphere.io/devops/pkg/apiserver/request"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/kapis/common"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var argoAppNotConfiguredError = restful.NewError(http.StatusBadRequest,
	"application is not initialized, please confirm you have already configured it")
var operationAlreadyInProgressError = restful.NewError(http.StatusBadRequest,
	"another operation is already in progress")
var invalidRequestBodyError = restful.NewError(http.StatusBadRequest,
	"invalid application sync request")
var unauthenticatedError = restful.NewError(http.StatusUnauthorized,
	"unauthenticated request")

func (h *handler) applicationSummary(request *restful.Request, response *restful.Response) {
	namespace := common.GetPathParameter(request, common.NamespacePathParameter)

	summary, err := h.populateApplicationSummary(namespace)
	common.Response(request, response, summary, err)
}

func (h *handler) populateApplicationSummary(namespace string) (*ApplicationsSummary, error) {
	applicationList := &v1alpha1.ApplicationList{}
	if err := h.List(context.Background(), applicationList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	summary := &ApplicationsSummary{
		HealthStatus: map[string]int{},
		SyncStatus:   map[string]int{},
	}
	summary.Total = len(applicationList.Items)

	for i := range applicationList.Items {
		app := &applicationList.Items[i]
		// accumulate health status
		healthStatus := app.GetLabels()[v1alpha1.HealthStatusLabelKey]
		if healthStatus != "" {
			summary.HealthStatus[healthStatus]++
		}
		// accumulate sync status
		syncStatus := app.GetLabels()[v1alpha1.SyncStatusLabelKey]
		if syncStatus != "" {
			summary.SyncStatus[syncStatus]++
		}
	}
	return summary, nil
}

func (h *handler) handleSyncApplication(req *restful.Request, res *restful.Response) {
	namespace := common.GetPathParameter(req, common.NamespacePathParameter)
	name := common.GetPathParameter(req, pathParameterApplication)

	syncRequest := &ApplicationSyncRequest{}
	err := req.ReadEntity(syncRequest)
	if err != nil {
		common.Response(req, res, nil, invalidRequestBodyError)
		return
	}

	currentUser, ok := apiserverrequest.UserFrom(req.Request.Context())
	if !ok || currentUser == nil {
		common.Response(req, res, nil, unauthenticatedError)
		return
	}

	app, err := h.syncApplication(namespace, name, syncRequest, currentUser)
	common.Response(req, res, app, err)
}

func (h *handler) syncApplication(namespace, name string, syncRequest *ApplicationSyncRequest, currentUser user.Info) (*v1alpha1.Application, error) {
	app := &v1alpha1.Application{}
	ctx := context.Background()
	if err := h.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, app); err != nil {
		return nil, err
	}
	if app.Spec.ArgoApp == nil {
		return nil, argoAppNotConfiguredError
	}

	argoApp := app.Spec.ArgoApp

	// concrete operation
	var syncOptions v1alpha1.SyncOptions
	var retry *v1alpha1.RetryStrategy
	if argoApp.Spec.SyncPolicy != nil {
		syncOptions = argoApp.Spec.SyncPolicy.SyncOptions
		retry = argoApp.Spec.SyncPolicy.Retry
	}
	if syncRequest.RetryStrategy != nil {
		retry = syncRequest.RetryStrategy
	}
	if syncRequest.SyncOptions != nil {
		syncOptions = *syncRequest.SyncOptions
	}

	operation := &v1alpha1.Operation{
		Sync: &v1alpha1.SyncOperation{
			Revision:     syncRequest.Revision,
			Prune:        syncRequest.Prune,
			DryRun:       syncRequest.DryRun,
			SyncOptions:  syncOptions,
			SyncStrategy: syncRequest.Strategy,
			Resources:    syncRequest.Resources,
			Manifests:    syncRequest.Manifests,
		},
		InitiatedBy: v1alpha1.OperationInitiator{Username: currentUser.GetName()},
		Info:        syncRequest.Infos,
	}
	if retry != nil {
		operation.Retry = *retry
	}

	return h.updateOperation(namespace, name, operation)
}

func (h *handler) updateOperation(namespace, name string, operation *v1alpha1.Operation) (*v1alpha1.Application, error) {
	var app *v1alpha1.Application
	return app, utilretry.RetryOnConflict(utilretry.DefaultRetry, func() error {
		app = &v1alpha1.Application{}
		err := h.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, app)
		if err != nil {
			return err
		}
		if app.Spec.ArgoApp == nil {
			return argoAppNotConfiguredError
		}
		if app.Spec.ArgoApp.Operation != nil {
			return operationAlreadyInProgressError
		}

		app.Spec.ArgoApp.Operation = operation
		if err := h.Update(context.Background(), app); err != nil {
			return err
		}
		// updated successfully
		return nil
	})
}

func (h *handler) getClusters(req *restful.Request, res *restful.Response) {
	ctx := context.Background()

	secrets := &v1.SecretList{}
	err := h.List(ctx, secrets, client.MatchingLabels{
		"argocd.argoproj.io/secret-type": "cluster",
	})

	// always have on cluster at least
	argoClusters := []v1alpha1.ApplicationDestination{{
		Server: "https://kubernetes.default.svc",
		Name:   "in-cluster",
	}}
	if err == nil && secrets != nil && len(secrets.Items) > 0 {
		for i := range secrets.Items {
			secret := secrets.Items[i]

			name := string(secret.Data["name"])
			server := string(secret.Data["server"])

			if name != "" && server != "" {
				argoClusters = append(argoClusters, v1alpha1.ApplicationDestination{
					Server: server,
					Name:   name,
				})
			}
		}
	}
	common.Response(req, res, argoClusters, err)
}

type handler struct {
	client.Client
	ArgoCDNamespace string
}

func newHandler(options *common.Options, argoOption *config.ArgoCDOption) *handler {
	return &handler{
		Client:          options.GenericClient,
		ArgoCDNamespace: argoOption.Namespace,
	}
}
