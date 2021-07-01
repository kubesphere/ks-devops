package anonymous

import (
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"net/http"
	"strings"
)

// Authenticator implements an anonymous auth
type Authenticator struct {}

func NewAuthenticator() authenticator.Request {
	return &Authenticator{}
}

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	if auth := strings.TrimSpace(req.Header.Get("Authorization")); auth == "" {
		return &authenticator.Response{
			User:      &user.DefaultInfo{
				Name:   "anonymous",
				Groups: []string{user.AllAuthenticated},
			},
		}, true, nil
	}
	return nil, false, nil
}
