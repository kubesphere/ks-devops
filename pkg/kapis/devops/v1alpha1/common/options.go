package common

import "sigs.k8s.io/controller-runtime/pkg/client"

// Options contain options needed by creating handlers.
type Options struct {
	GenericClient client.Client
}
