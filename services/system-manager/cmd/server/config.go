package main

import (
	"os"
	"strings"
)

type Config struct {
	Port                int
	Mode                string // "full" or "lite"
	EnableSystemdManage bool
}

func LoadConfig() *Config {
	cfg := &Config{
		Port: 50051,
		Mode: "full",
		EnableSystemdManage: true,
	}

	// Override from environment
	if mode := os.Getenv("SERVICE_MODE"); mode != "" {
		cfg.Mode = strings.ToLower(mode)
	}

	if port := os.Getenv("GRPC_PORT"); port != "" {
		// Parse port...
	}

	// Disable features based on mode
	if cfg.Mode == "lite" {
		cfg.EnableSystemdManage = false
	}

	return cfg
}