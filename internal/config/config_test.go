package config

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	// Get default config
	config := DefaultConfig()
	
	// Verify node config
	assert.NotEmpty(t, config.Node.ID)
	assert.NotNil(t, config.Node.Tags)
	assert.NotEmpty(t, config.Node.DataDir)
	
	// Verify network config
	assert.NotEmpty(t, config.Network.ListenAddr)
	assert.NotEmpty(t, config.Network.PublicAddr)
	assert.NotNil(t, config.Network.BootstrapNodes)
	
	// Verify storage config
	assert.Greater(t, config.Storage.CacheSize, 0)
	assert.Greater(t, config.Storage.SpillThreshold, 0)
	
	// Verify security config
	assert.NotEmpty(t, config.Security.CertFile)
	assert.NotEmpty(t, config.Security.KeyFile)
	assert.NotEmpty(t, config.Security.TrustedKeysFile)
}

func TestSaveLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "holocompute-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Create a config
	config := DefaultConfig()
	config.Node.ID = "test-node"
	config.Node.DataDir = filepath.Join(tempDir, "data")
	config.Network.ListenAddr = "127.0.0.1:9000"
	config.Network.PublicAddr = "127.0.0.1:9000"
	
	// Save config to file
	configFile := filepath.Join(tempDir, "config.yaml")
	err = config.SaveConfig(configFile)
	assert.NoError(t, err)
	
	// Load config from file
	loadedConfig, err := LoadConfig(configFile)
	assert.NoError(t, err)
	
	// Verify loaded config matches saved config
	assert.Equal(t, config.Node.ID, loadedConfig.Node.ID)
	assert.Equal(t, config.Node.DataDir, loadedConfig.Node.DataDir)
	assert.Equal(t, config.Network.ListenAddr, loadedConfig.Network.ListenAddr)
	assert.Equal(t, config.Network.PublicAddr, loadedConfig.Network.PublicAddr)
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Try to load non-existent config file
	config, err := LoadConfig("/non/existent/file.yaml")
	
	// Should return default config without error
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.Node.ID)
}