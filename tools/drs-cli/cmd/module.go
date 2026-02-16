package cmd

import (
	"github.com/spf13/cobra"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Module-manager service commands",
	Long:  "Commands for interacting with the module-manager service",
}

func init() {
	rootCmd.AddCommand(moduleCmd)

	// Add existing commands as subcommands of module
	moduleCmd.AddCommand(servicesCmd)
	moduleCmd.AddCommand(systemCmd)
	moduleCmd.AddCommand(monitoringCmd)
}
