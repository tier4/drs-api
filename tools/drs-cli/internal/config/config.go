package config

import (
	"fmt"
	"time"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address: "localhost:50051",
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Config) GetServerAddress() string {
	if c.Server.Address == "" {
		return "localhost:50051"
	}
	return c.Server.Address
}

func (c *Config) GetTimeout() time.Duration {
	if c.Server.Timeout == 0 {
		return 30 * time.Second
	}
	return c.Server.Timeout
}

func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	return nil
}
