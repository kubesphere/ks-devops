package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	config := getDefaultConfig()
	assert.NotNil(t, config)
	assert.NotNil(t, config["pod.concurrent"])
}

func TestGetHighConfig(t *testing.T) {
	config := getHighConfig()
	assert.NotNil(t, config)
	assert.NotNil(t, config["pod.concurrent"])
}
