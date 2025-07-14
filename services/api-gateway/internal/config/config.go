package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the API Gateway configuration
type Config struct {
	Server  ServerConfig      `yaml:"server"`
	Modules map[string]Module `yaml:"modules"`
	GRPC    GRPCConfig        `yaml:"grpc"`
	CORS    CORSConfig        `yaml:"cors"`
}

// ServerConfig represents the HTTP server configuration
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// Module represents a single module (ECU with sensors) configuration
type Module struct {
	Address            string   `yaml:"address"`
	EnabledServices    []string `yaml:"enabled_services"`
	HasROS2Bridge      bool     `yaml:"has_ros2_bridge"`
	ROS2BridgeAddress  string   `yaml:"ros2_bridge_address,omitempty"`
}

// GRPCConfig represents gRPC client configuration
type GRPCConfig struct {
	Timeout  time.Duration `yaml:"timeout"`
	MaxRetry int           `yaml:"max_retry"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// Load loads the configuration from a YAML file
func Load(configPath string) (*Config, error) {
	// Default config path
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Try multiple locations for config file
	searchPaths := []string{
		configPath,
		"./config.yaml",
		"./configs/config.yaml",
		"/etc/api-gateway/config.yaml",
	}

	var configFile string
	var err error

	for _, path := range searchPaths {
		if _, err = os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		return nil, fmt.Errorf("config file not found in any of the search paths: %v", searchPaths)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", configFile, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %v", configFile, err)
	}

	// Set defaults
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30 * time.Second
	}
	if config.GRPC.Timeout == 0 {
		config.GRPC.Timeout = 10 * time.Second
	}
	if config.GRPC.MaxRetry == 0 {
		config.GRPC.MaxRetry = 3
	}

	return &config, nil
}

// GetServerAddress returns the server address in host:port format
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsServiceEnabled checks if a service is enabled for a given module
func (c *Config) IsServiceEnabled(hostname, service string) bool {
	module, exists := c.Modules[hostname]
	if !exists {
		return false
	}

	for _, enabledService := range module.EnabledServices {
		if enabledService == service {
			return true
		}
	}
	return false
}

// GetModuleAddress returns the address for a given module
func (c *Config) GetModuleAddress(hostname string) (string, bool) {
	module, exists := c.Modules[hostname]
	if !exists {
		return "", false
	}
	return module.Address, true
}

// GetROS2BridgeAddress returns the ROS2 bridge address for a given module
func (c *Config) GetROS2BridgeAddress(hostname string) (string, bool) {
	module, exists := c.Modules[hostname]
	if !exists || !module.HasROS2Bridge {
		return "", false
	}
	return module.ROS2BridgeAddress, true
}

// GetModuleNames returns a list of all module hostnames
func (c *Config) GetModuleNames() []string {
	names := make([]string, 0, len(c.Modules))
	for name := range c.Modules {
		names = append(names, name)
	}
	return names
}