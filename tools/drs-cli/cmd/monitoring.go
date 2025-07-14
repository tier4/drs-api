package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/drs-api/tools/drs-cli/internal/client"
	modulev1 "github.com/drs-api/tools/drs-cli/drs/module/v1"
)

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Monitoring commands",
	Long:  "Commands for resource monitoring via MonitoringService",
}

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Get disk usage information",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.GetDiskUsage()
		if err != nil {
			return fmt.Errorf("failed to get disk usage: %v", err)
		}

		usage := resp.DiskUsage
		fmt.Printf("Disk Usage:\n")
		fmt.Printf("  Total: %s\n", formatBytes(usage.TotalBytes))
		fmt.Printf("  Used:  %s\n", formatBytes(usage.UsedBytes))
		fmt.Printf("  Free:  %s\n", formatBytes(usage.FreeBytes))
		fmt.Printf("  Usage: %.1f%%\n", usage.UsagePercentage)

		return nil
	},
}

var ptpCmd = &cobra.Command{
	Use:   "ptp",
	Short: "Get PTP status information",
	RunE: func(cmd *cobra.Command, args []string) error {
		includeRemote, _ := cmd.Flags().GetBool("include-remote")
		
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.GetPTPStatus(includeRemote)
		if err != nil {
			return fmt.Errorf("failed to get PTP status: %v", err)
		}

		fmt.Printf("Local PTP Status:\n")
		printPTPStatus(resp.LocalStatus, "  ")

		if includeRemote && len(resp.RemoteStatuses) > 0 {
			fmt.Printf("\nRemote Device PTP Status:\n")
			for _, remote := range resp.RemoteStatuses {
				fmt.Printf("  Device: %s (%s)\n", remote.DeviceName, remote.IpAddress)
				if remote.IsReachable {
					printPTPStatus(remote.Status, "    ")
				} else {
					fmt.Printf("    Status: Unreachable\n")
					fmt.Printf("    Error: %s\n", remote.ErrorMessage)
				}
				fmt.Printf("\n")
			}
		}

		return nil
	},
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Get environment variables",
	Long:  "Get SENSING_SYSTEM_ID and MODULE_ID environment variables from the module manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.GetEnvironment()
		if err != nil {
			return fmt.Errorf("failed to get environment: %v", err)
		}

		fmt.Printf("Environment Variables:\n")
		fmt.Printf("  SENSING_SYSTEM_ID: %s\n", resp.SensingSystemId)
		fmt.Printf("  MODULE_ID:         %s\n", resp.ModuleId)

		return nil
	},
}

func printPTPStatus(status *modulev1.PTPStatus, indent string) {
	if status == nil {
		fmt.Printf("%sStatus: No data available\n", indent)
		return
	}
	
	fmt.Printf("%sClock ID: %s\n", indent, status.ClockId)
	fmt.Printf("%sMaster Offset: %d ns\n", indent, status.MasterOffsetNs)
	fmt.Printf("%sIngress Time: %d\n", indent, status.IngressTime)
	fmt.Printf("%sGM Present: %t\n", indent, status.GmPresent)
	fmt.Printf("%sGM Identity: %s\n", indent, status.GmIdentity)
	fmt.Printf("%sSynced: %t\n", indent, status.IsSynced)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	// Commands will be added to moduleCmd in module.go
	monitoringCmd.AddCommand(diskCmd)
	monitoringCmd.AddCommand(ptpCmd)
	monitoringCmd.AddCommand(envCmd)
	
	// Add include-remote flag to ptp command
	ptpCmd.Flags().BoolP("include-remote", "r", false, "include remote device PTP status")
}