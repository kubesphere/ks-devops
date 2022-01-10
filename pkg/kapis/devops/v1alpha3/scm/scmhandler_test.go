package scm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler(nil)
	assert.NotNil(t, handler)
}
