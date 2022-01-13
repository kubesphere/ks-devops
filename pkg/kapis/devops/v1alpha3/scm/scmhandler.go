package scm

import (
	"context"
	goscm "github.com/drone/go-scm/scm"
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/client/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler holds all the API handlers of SCM
type Handler struct {
	k8sClient client.Client
}

// NewHandler creates the instance of the SCM handler
func NewHandler(k8sClient client.Client) *Handler {
	return &Handler{k8sClient: k8sClient}
}

// Verify verifies a SCM auth
func (h *Handler) Verify(request *restful.Request, response *restful.Response) {
	scm := request.PathParameter("scm")
	secretName := request.QueryParameter("secret")
	secretNamespace := request.QueryParameter("secretNamespace")

	factory := git.NewClientFactory(scm, &v1.SecretReference{
		Namespace: secretNamespace, Name: secretName,
	}, h.k8sClient)

	var c *goscm.Client
	var err error
	var code int

	if c, err = factory.GetClient(); err == nil {
		var resp *goscm.Response

		if _, resp, err = c.Organizations.List(context.TODO(), goscm.ListOptions{Size: 1, Page: 1}); err == nil {
			code = resp.Status
		}
	} else {
		code = 100
	}

	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	_ = response.WriteAsJson(git.VerifyResult(err, code))
}
