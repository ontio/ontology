package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigGeneration(t *testing.T) {
	polarisConfig := newPolarisConfig()
	assert.Equal(t, polarisConfig, Parameters)
	defaultConfig := newDefaultConfig()
	assert.NotEqual(t, defaultConfig, polarisConfig)
}
