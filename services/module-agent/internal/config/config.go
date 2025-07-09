package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	APIs     APIsConfig     `yaml:"apis"`
	Disk     DiskConfig     `yaml:"disk"`
	Services ServicesConfig `yaml:"services"`
	System   SystemConfig   `yaml:"system"`
	PTP      PTPConfig      `yaml:"ptp"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type APIsConfig struct {
	EnableReboot           bool `yaml:"enable_reboot"`
	EnableShutdown         bool `yaml:"enable_shutdown"`
	EnableServiceManagement bool `yaml:"enable_service_management"`
	EnableDiskUsage        bool `yaml:"enable_disk_usage"`
	EnablePTPCheck         bool `yaml:"enable_ptp_check"`
}

type DiskConfig struct {
	MonitoredPaths []DiskPath `yaml:"monitored_paths"`
	DefaultPath    string     `yaml:"default_path"`
}

type DiskPath struct {
	Path        string `yaml:"path"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type ServicesConfig struct {
	EnableSystemdManage bool                      `yaml:"enable_systemd_manage"`
	AllowedServices     []string                  `yaml:"allowed_services"`
	Services            map[string]ServiceMapping `yaml:"services"`
}

type ServiceMapping struct {
	SystemdName string `yaml:"systemd_name"`
	DisplayName string `yaml:"display_name"`
	Description string `yaml:"description"`
	Enabled     bool   `yaml:"enabled"`
}

type SystemConfig struct {
	MaxDelaySeconds int  `yaml:"max_delay_seconds"`
	AllowReboot     bool `yaml:"allow_reboot"`
	AllowShutdown   bool `yaml:"allow_shutdown"`
}

type PTPConfig struct {
	RemoteDevices   []RemoteDevice `yaml:"remote_devices"`
	SyncThresholdNs int64          `yaml:"sync_threshold_ns"`
}

type RemoteDevice struct {
	Name      string `yaml:"name"`
	IPAddress string `yaml:"ip_address"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Default configuration
	config := &Config{
		Server: ServerConfig{
			Port: 50051,
		},
		APIs: APIsConfig{
			EnableReboot:           true,
			EnableShutdown:         true,
			EnableServiceManagement: true,
			EnableDiskUsage:        true,
			EnablePTPCheck:         true,
		},
		Disk: DiskConfig{
			DefaultPath: "/",
			MonitoredPaths: []DiskPath{
				{Path: "/", Name: "root", Description: "Root filesystem"},
			},
		},
		Services: ServicesConfig{
			EnableSystemdManage: true,
			AllowedServices:     []string{"drs_sensor.service", "drs_recorder.service"},
			Services: map[string]ServiceMapping{
				"drs_sensor": {
					SystemdName: "drs_sensor.service",
					DisplayName: "DRS Sensor Service",
					Description: "Data recording sensor management service",
					Enabled:     true,
				},
				"drs_recorder": {
					SystemdName: "drs_recorder.service",
					DisplayName: "DRS Recorder Service",
					Description: "Data recording service",
					Enabled:     true,
				},
			},
		},
		System: SystemConfig{
			MaxDelaySeconds: 300,
			AllowReboot:     true,
			AllowShutdown:   true,
		},
		PTP: PTPConfig{
			RemoteDevices:   []RemoteDevice{},
			SyncThresholdNs: 1000000, // 1ms default
		},
	}

	// If no config file specified, return default
	if configPath == "" {
		return config, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil // Return default if file doesn't exist
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadConfigFromDefault tries to load config from default locations
func LoadConfigFromDefault() (*Config, error) {
	// Try different locations
	possiblePaths := []string{
		"config.yaml",
		"config.yml",
		"/etc/module-agent/config.yaml",
		"/etc/module-agent/config.yml",
	}

	// Try to find executable directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		possiblePaths = append([]string{
			filepath.Join(execDir, "config.yaml"),
			filepath.Join(execDir, "config.yml"),
		}, possiblePaths...)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return LoadConfig(path)
		}
	}

	// No config file found, return default
	return LoadConfig("")
}

func validateConfig(config *Config) error {
	// Validate server port
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate disk paths
	if config.Disk.DefaultPath == "" {
		return fmt.Errorf("default_path cannot be empty")
	}

	for _, path := range config.Disk.MonitoredPaths {
		if path.Path == "" {
			return fmt.Errorf("disk path cannot be empty")
		}
		if path.Name == "" {
			return fmt.Errorf("disk path name cannot be empty")
		}
	}

	// Validate system settings
	if config.System.MaxDelaySeconds < 0 {
		return fmt.Errorf("max_delay_seconds cannot be negative")
	}

	return nil
}

// GetDiskPath returns disk path by name, or default if not found
func (c *Config) GetDiskPath(name string) string {
	if name == "" {
		return c.Disk.DefaultPath
	}

	for _, path := range c.Disk.MonitoredPaths {
		if path.Name == name {
			return path.Path
		}
	}

	return c.Disk.DefaultPath
}

// GetDiskPaths returns all configured disk paths
func (c *Config) GetDiskPaths() []DiskPath {
	return c.Disk.MonitoredPaths
}

// IsServiceAllowed checks if a service is allowed to be managed
func (c *Config) IsServiceAllowed(serviceName string) bool {
	for _, allowed := range c.Services.AllowedServices {
		if allowed == serviceName {
			return true
		}
	}
	return false
}

// ParseResourceName parses a resource name like "services/drs_sensor"
func ParseResourceName(resourceName string) (resourceType, resourceID string, err error) {
	parts := strings.Split(resourceName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource name format: %s (expected: type/id)", resourceName)
	}
	return parts[0], parts[1], nil
}

// GetSystemdServiceName converts a resource name to systemd service name
func (c *Config) GetSystemdServiceName(resourceName string) (string, error) {
	resourceType, resourceID, err := ParseResourceName(resourceName)
	if err != nil {
		return "", err
	}
	
	if resourceType != "services" {
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	
	service, exists := c.Services.Services[resourceID]
	if !exists {
		return "", fmt.Errorf("service not found: %s", resourceID)
	}
	
	if !service.Enabled {
		return "", fmt.Errorf("service is disabled: %s", resourceID)
	}
	
	return service.SystemdName, nil
}

// GetServiceMapping returns service mapping for a resource ID
func (c *Config) GetServiceMapping(resourceID string) (ServiceMapping, error) {
	service, exists := c.Services.Services[resourceID]
	if !exists {
		return ServiceMapping{}, fmt.Errorf("service not found: %s", resourceID)
	}
	return service, nil
}

// GetAllEnabledServices returns all enabled services
func (c *Config) GetAllEnabledServices() []string {
	var services []string
	for id, service := range c.Services.Services {
		if service.Enabled {
			services = append(services, id)
		}
	}
	return services
}

// IsResourceAllowed checks if a resource is allowed
func (c *Config) IsResourceAllowed(resourceName string) bool {
	systemdName, err := c.GetSystemdServiceName(resourceName)
	if err != nil {
		return false
	}
	return c.IsServiceAllowed(systemdName)
}