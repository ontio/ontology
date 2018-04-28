package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestConfigGeneration(t *testing.T){
	polarisConfig := newPolarisConfig()
	assert.Equal(t, polarisConfig, Parameters)
	defaultConfig := newDefaultConfig()
	assert.NotEqual(t, defaultConfig, polarisConfig)
}