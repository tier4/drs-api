package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tier4/drs-api/tools/drs-cli/internal/client"
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Service management commands",
	Long:  "Commands for managing systemd services via ServiceManagerService",
}

var listServicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured services",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.ListServices()
		if err != nil {
			return fmt.Errorf("failed to list services: %v", err)
		}

		fmt.Printf("Services:\n")
		for _, service := range resp.Services {
			fmt.Printf("  Name: %s\n", service.Name)
			fmt.Printf("  State: %s\n", service.State.String())
			fmt.Printf("  Enabled: %t\n", service.Enabled)
			fmt.Printf("  Description: %s\n", service.Description)
			if service.UptimeSeconds > 0 {
				uptime := time.Duration(service.UptimeSeconds) * time.Second
				fmt.Printf("  Uptime: %s\n", uptime.String())
			}
			fmt.Printf("\n")
		}

		return nil
	},
}

var getServiceCmd = &cobra.Command{
	Use:   "get [service-name]",
	Short: "Get information about a specific service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		service, err := c.GetService(serviceName)
		if err != nil {
			return fmt.Errorf("failed to get service %s: %v", serviceName, err)
		}

		fmt.Printf("Service: %s\n", service.Name)
		fmt.Printf("State: %s\n", service.State.String())
		fmt.Printf("Enabled: %t\n", service.Enabled)
		fmt.Printf("Description: %s\n", service.Description)
		if service.UptimeSeconds > 0 {
			uptime := time.Duration(service.UptimeSeconds) * time.Second
			fmt.Printf("Uptime: %s\n", uptime.String())
		}

		return nil
	},
}

var startServiceCmd = &cobra.Command{
	Use:   "start [service-name]",
	Short: "Start a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.StartService(serviceName)
		if err != nil {
			return fmt.Errorf("failed to start service %s: %v", serviceName, err)
		}

		fmt.Printf("Service %s started successfully\n", serviceName)
		fmt.Printf("Current state: %s\n", resp.Service.State.String())

		return nil
	},
}

var stopServiceCmd = &cobra.Command{
	Use:   "stop [service-name]",
	Short: "Stop a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.StopService(serviceName)
		if err != nil {
			return fmt.Errorf("failed to stop service %s: %v", serviceName, err)
		}

		fmt.Printf("Service %s stopped successfully\n", serviceName)
		fmt.Printf("Current state: %s\n", resp.Service.State.String())

		return nil
	},
}

var restartServiceCmd = &cobra.Command{
	Use:   "restart [service-name]",
	Short: "Restart a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.RestartService(serviceName)
		if err != nil {
			return fmt.Errorf("failed to restart service %s: %v", serviceName, err)
		}

		fmt.Printf("Service %s restarted successfully\n", serviceName)
		fmt.Printf("Current state: %s\n", resp.Service.State.String())

		return nil
	},
}

var enableServiceCmd = &cobra.Command{
	Use:   "enable [service-name]",
	Short: "Enable a service for auto-start",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.EnableService(serviceName)
		if err != nil {
			return fmt.Errorf("failed to enable service %s: %v", serviceName, err)
		}

		fmt.Printf("Service %s enabled successfully\n", serviceName)
		fmt.Printf("Enabled: %t\n", resp.Service.Enabled)

		return nil
	},
}

var disableServiceCmd = &cobra.Command{
	Use:   "disable [service-name]",
	Short: "Disable a service from auto-start",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.DisableService(serviceName)
		if err != nil {
			return fmt.Errorf("failed to disable service %s: %v", serviceName, err)
		}

		fmt.Printf("Service %s disabled successfully\n", serviceName)
		fmt.Printf("Enabled: %t\n", resp.Service.Enabled)

		return nil
	},
}

func init() {
	// Commands will be added to moduleCmd in module.go
	servicesCmd.AddCommand(listServicesCmd)
	servicesCmd.AddCommand(getServiceCmd)
	servicesCmd.AddCommand(startServiceCmd)
	servicesCmd.AddCommand(stopServiceCmd)
	servicesCmd.AddCommand(restartServiceCmd)
	servicesCmd.AddCommand(enableServiceCmd)
	servicesCmd.AddCommand(disableServiceCmd)
}
