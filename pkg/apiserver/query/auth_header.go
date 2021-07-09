package query

import "strings"

// GetAuthorization returns the authorization from HTTP header
func GetAuthorization(header getter) (auth string) {
	if header == nil {
		return
	}

	if auth = strings.TrimSpace(header.Get("X-Authorization")); auth == "" {
		auth = strings.TrimSpace(header.Get("Authorization"))
	}
	return
}

type getter interface {
	Get(key string) string
}
