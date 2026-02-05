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
	Disk     DiskConfig     `yaml:"disk"`
	Services ServicesConfig `yaml:"services"`
	System   SystemConfig   `yaml:"system"`
	PTP      PTPConfig      `yaml:"ptp"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DiskConfig struct {
	Enabled     bool        `yaml:"enabled"`
	MonitorPath string      `yaml:"monitor_path"` // Deprecated: use Disks instead, kept for backward compatibility
	Disks       []DiskEntry `yaml:"disks"`
}

type DiskEntry struct {
	Name        string `yaml:"name"`        // disk_id used in resource name
	MountPath   string `yaml:"mount_path"`  // mount path to monitor
	Description string `yaml:"description"` // optional description
}

type ServicesConfig struct {
	Enabled  bool                      `yaml:"enabled"`
	Services map[string]ServiceMapping `yaml:"services"`
}

type ServiceMapping struct {
	SystemdName string `yaml:"systemd_name"`
	Description string `yaml:"description"`
}

type SystemConfig struct {
	EnableReboot   bool `yaml:"enable_reboot"`
	EnableShutdown bool `yaml:"enable_shutdown"`
}

type PTPConfig struct {
	Enabled       bool           `yaml:"enabled"`
	RemoteDevices []RemoteDevice `yaml:"remote_devices"`
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
		Disk: DiskConfig{
			Enabled:     true,
			MonitorPath: "/",
			Disks:       []DiskEntry{}, // Will be populated from MonitorPath if empty during validation/loading
		},
		Services: ServicesConfig{
			Enabled:  true,
			Services: map[string]ServiceMapping{
				"drs_sensor": {
					SystemdName: "drs_sensor.service",
					Description: "Data recording sensor management service",
				},
				"drs_recorder": {
					SystemdName: "drs_recorder.service",
					Description: "Data recording service",
				},
			},
		},
		System: SystemConfig{
			EnableReboot:   true,
			EnableShutdown: true,
		},
		PTP: PTPConfig{
			Enabled:       true,
			RemoteDevices: []RemoteDevice{},
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
		"/etc/module-manager/config.yaml",
		"/etc/module-manager/config.yml",
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

	// Validate disk path
	// Validate disk path and setup backward compatibility
	if config.Disk.Enabled {
		// If Disks is empty but MonitorPath is set, migrate it to Disks
		if len(config.Disk.Disks) == 0 && config.Disk.MonitorPath != "" {
			config.Disk.Disks = []DiskEntry{
				{
					Name:        "root",
					MountPath:   config.Disk.MonitorPath,
					Description: "Primary Disk",
				},
			}
		}

		if len(config.Disk.Disks) == 0 {
			return fmt.Errorf("no disks configured when disk monitoring is enabled")
		}

		for _, disk := range config.Disk.Disks {
			if disk.Name == "" {
				return fmt.Errorf("disk name cannot be empty")
			}
			if disk.MountPath == "" {
				return fmt.Errorf("disk mount_path cannot be empty for disk %s", disk.Name)
			}
		}
	}


	return nil
}

// GetDiskPath returns the configured disk monitor path
func (c *Config) GetDiskPath() string {
	return c.Disk.MonitorPath
}

// IsServiceAllowed checks if a service is allowed to be managed
func (c *Config) IsServiceAllowed(serviceName string) bool {
	// Check if the service is defined in the Services map
	for _, service := range c.Services.Services {
		if service.SystemdName == serviceName {
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

// GetAllEnabledServices returns all services defined in config
func (c *Config) GetAllEnabledServices() []string {
	var services []string
	for id := range c.Services.Services {
		services = append(services, id)
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