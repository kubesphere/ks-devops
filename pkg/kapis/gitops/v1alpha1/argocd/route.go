package argocd

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/kapis/common"
	"net/http"
)

var (
	// pathParameterApplication is a path parameter definition for application.
	pathParameterApplication = restful.PathParameter("application", "The application name")
)

// ApplicationsSummary is the model of application summary response.
type ApplicationsSummary struct {
	Total        int            `json:"total"`
	HealthStatus map[string]int `json:"healthStatus"`
	SyncStatus   map[string]int `json:"syncStatus"`
}

// ApplicationSyncRequest is a request to apply an operation to change state.
type ApplicationSyncRequest struct {
	Revision      string                           `json:"revision"`
	DryRun        bool                             `json:"dryRun"`
	Prune         bool                             `json:"prune"`
	Strategy      *v1alpha1.SyncStrategy           `json:"strategy,omitempty"`
	Resources     []v1alpha1.SyncOperationResource `json:"resources"`
	Manifests     []string                         `json:"manifests,omitempty"`
	Infos         []*v1alpha1.Info                 `json:"infos,omitempty"`
	RetryStrategy *v1alpha1.RetryStrategy          `json:"retryStrategy,omitempty"`
	SyncOptions   *v1alpha1.SyncOptions            `json:"syncOptions,omitempty"`
}

// RegisterRoutes is for registering Argo CD Application routes into WebService.
func RegisterRoutes(service *restful.WebService, options *common.Options, argoOption *config.ArgoCDOption) {
	handler := newHandler(options, argoOption)

	service.Route(service.GET("/namespaces/{namespace}/application-summary").
		To(handler.applicationSummary).
		Param(common.NamespacePathParameter).
		Doc("Fetch applications summary").
		Returns(http.StatusOK, api.StatusOK, ApplicationsSummary{}))

	service.Route(service.POST("/namespaces/{namespace}/applications/{application}/sync").
		To(handler.handleSyncApplication).
		Param(common.NamespacePathParameter).
		Param(pathParameterApplication).
		Reads(ApplicationSyncRequest{}).
		Doc("Sync a particular application manually").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Application{}))

	service.Route(service.GET("/clusters").
		To(handler.getClusters).
		Doc("Get the clusters list").
		Returns(http.StatusOK, api.StatusOK, []v1alpha1.ApplicationDestination{}))
}
