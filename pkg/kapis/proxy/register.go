package proxy

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/ks-devops/pkg/api/devops"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha2"
)

func AddToContainer(container *restful.Container) {
	versions := []string{v1alpha1.GroupVersion.Version, v1alpha2.GroupVersion.Version, v1alpha3.GroupVersion.Version}
	for _, version := range versions {
		proxyWS := new(restful.WebService)
		proxyWS.Path("/" + version).
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON)

		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodConnect, http.MethodHead, http.MethodOptions, http.MethodTrace} {
			proxyWS.Route(proxyWS.Method(method).Path("/{subpath:*}").To(func(req *restful.Request, resp *restful.Response) {
				// Rewrite the URL to include /kapis
				originalPath := req.Request.URL.Path
				newPath := fmt.Sprintf("/kapis/%s%s", devops.GroupName, originalPath)

				// Forward the request to the container with the rewritten path
				req.Request.URL.Path = newPath
				container.Dispatch(resp, req.Request)
			}))
		}

		container.Add(proxyWS)
	}
}
