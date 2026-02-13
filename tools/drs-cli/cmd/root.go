package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tier4/drs-api/tools/drs-cli/internal/config"
)

var (
	cfgFile string
	address string
	timeout int
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "drs-cli",
	Short: "CLI client for DRS services",
	Long: `A command line client for interacting with DRS gRPC services.

Supports:
- module-manager service (ServiceManagerService, SystemControlService, and MonitoringService)
- ros2-bridge service (SensingService and RecordingService)

Use --address to specify different service endpoints.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&address, "address", "localhost:50051", "server address (50051 for module-manager, 50052 for ros2-bridge)")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 30, "request timeout in seconds")
}

func initConfig() {
	cfg = config.DefaultConfig()

	// Override with command line flags
	if address != "" {
		cfg.Server.Address = address
	}
	if timeout > 0 {
		cfg.Server.Timeout = cfg.Server.Timeout
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}
}
