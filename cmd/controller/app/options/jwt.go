package options

import "time"

// JWTOptions contain some options of JWT, such as secret and clock skew.
type JWTOptions struct {
	// MaximumClockSkew indicates token verification maximum time difference.
	MaximumClockSkew time.Duration
	// Secret is used to sign JWT token.
	Secret string
}
