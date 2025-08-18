package config

import (
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// Config represents the HoloCompute configuration
type Config struct {
	// Node configuration
	Node NodeConfig `yaml:"node"`
	
	// Network configuration
	Network NetworkConfig `yaml:"network"`
	
	// Storage configuration
	Storage StorageConfig `yaml:"storage"`
	
	// Security configuration
	Security SecurityConfig `yaml:"security"`
}

// NodeConfig contains node-specific configuration
type NodeConfig struct {
	// ID is the unique identifier for this node
	ID string `yaml:"id"`
	
	// Tags are arbitrary tags for this node
	Tags []string `yaml:"tags"`
	
	// DataDir is the directory for storing data
	DataDir string `yaml:"data_dir"`
}

// NetworkConfig contains network configuration
type NetworkConfig struct {
	// ListenAddr is the address to listen on
	ListenAddr string `yaml:"listen_addr"`
	
	// PublicAddr is the public address for this node
	PublicAddr string `yaml:"public_addr"`
	
	// BootstrapNodes are the addresses of bootstrap nodes
	BootstrapNodes []string `yaml:"bootstrap_nodes"`
	
	// EnablePQ enables post-quantum cryptography
	EnablePQ bool `yaml:"enable_pq"`
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	// CacheSize is the size of the page cache in MB
	CacheSize int `yaml:"cache_size"`
	
	// SpillThreshold is the threshold for spilling to disk in MB
	SpillThreshold int `yaml:"spill_threshold"`
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	// CertFile is the path to the TLS certificate file
	CertFile string `yaml:"cert_file"`
	
	// KeyFile is the path to the TLS key file
	KeyFile string `yaml:"key_file"`
	
	// TrustedKeysFile is the path to the trusted keys file
	TrustedKeysFile string `yaml:"trusted_keys_file"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	
	// Default data directory
	dataDir := filepath.Join(homeDir, ".holocompute")
	
	return &Config{
		Node: NodeConfig{
			ID:      "node-1",
			Tags:    []string{},
			DataDir: dataDir,
		},
		Network: NetworkConfig{
			ListenAddr:      "0.0.0.0:8443",
			PublicAddr:      "127.0.0.1:8443",
			BootstrapNodes:  []string{},
			EnablePQ:        true,
		},
		Storage: StorageConfig{
			CacheSize:       1024, // 1GB
			SpillThreshold:  512,  // 512MB
		},
		Security: SecurityConfig{
			CertFile:        filepath.Join(dataDir, "cert.pem"),
			KeyFile:         filepath.Join(dataDir, "key.pem"),
			TrustedKeysFile: filepath.Join(dataDir, "trusted_keys.pem"),
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(filename string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	// Parse YAML
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	
	return config, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(filename, data, 0644)
}