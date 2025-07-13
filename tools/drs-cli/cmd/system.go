package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/drs-api/tools/drs-cli/internal/client"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System control commands",
	Long:  "Commands for system-level operations via SystemControlService",
}

var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		delay, _ := cmd.Flags().GetInt32("delay")
		
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.Reboot(delay)
		if err != nil {
			return fmt.Errorf("failed to reboot system: %v", err)
		}

		if resp.Accepted {
			if delay > 0 {
				fmt.Printf("Reboot scheduled in %d seconds\n", resp.ScheduledDelay)
			} else {
				fmt.Printf("Rebooting now\n")
			}
			fmt.Printf("Message: %s\n", resp.Message)
		} else {
			fmt.Printf("Reboot request was not accepted\n")
			fmt.Printf("Message: %s\n", resp.Message)
		}

		return nil
	},
}

var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		delay, _ := cmd.Flags().GetInt32("delay")
		
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		resp, err := c.Shutdown(delay)
		if err != nil {
			return fmt.Errorf("failed to shutdown system: %v", err)
		}

		if resp.Accepted {
			if delay > 0 {
				fmt.Printf("Shutdown scheduled in %d seconds\n", resp.ScheduledDelay)
			} else {
				fmt.Printf("Shutting down now\n")
			}
			fmt.Printf("Message: %s\n", resp.Message)
		} else {
			fmt.Printf("Shutdown request was not accepted\n")
			fmt.Printf("Message: %s\n", resp.Message)
		}

		return nil
	},
}

func init() {
	// Commands will be added to moduleCmd in module.go
	systemCmd.AddCommand(rebootCmd)
	systemCmd.AddCommand(shutdownCmd)
	
	// Add delay flag to both reboot and shutdown commands
	rebootCmd.Flags().Int32P("delay", "d", 0, "delay in seconds before reboot")
	shutdownCmd.Flags().Int32P("delay", "d", 0, "delay in seconds before shutdown")
}