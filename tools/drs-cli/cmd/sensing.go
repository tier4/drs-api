package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tier4/drs-api/tools/drs-cli/internal/client"
)

var sensingCmd = &cobra.Command{
	Use:   "sensing",
	Short: "Sensing data commands",
	Long:  "Commands for accessing sensing data via SensingService",
}

var getPositionCmd = &cobra.Command{
	Use:   "position",
	Short: "Get current position",
	Long:  "Get the current position from NavSatFix data",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		position, err := c.GetPosition()
		if err != nil {
			return fmt.Errorf("failed to get position: %v", err)
		}

		// Format position output
		fmt.Printf("Position:\n")
		fmt.Printf("  Latitude:  %.6f°\n", position.Latitude)
		fmt.Printf("  Longitude: %.6f°\n", position.Longitude)
		fmt.Printf("  Altitude:  %.3f m\n", position.Altitude)
		fmt.Printf("  Frame ID:  %s\n", position.Header.FrameId)

		// Convert timestamp
		if position.Header.Stamp != nil {
			timestamp := time.Unix(position.Header.Stamp.Seconds, int64(position.Header.Stamp.Nanos))
			fmt.Printf("  Timestamp: %s\n", timestamp.Format(time.RFC3339))
		}

		// Navigation status
		fmt.Printf("  Nav Status: %d\n", position.NavSatStatus.Status)
		fmt.Printf("  Nav Service: %d\n", position.NavSatStatus.Service)
		fmt.Printf("  Covariance Type: %d\n", position.PositionCovarianceType)

		return nil
	},
}

var listNodesCmd = &cobra.Command{
	Use:   "list-nodes",
	Short: "List ROS2 nodes",
	Long:  "List ROS2 nodes with optional filtering",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		filter, _ := cmd.Flags().GetString("filter")
		response, err := c.ListNodes(filter)
		if err != nil {
			return fmt.Errorf("failed to list nodes: %v", err)
		}

		if len(response.Nodes) == 0 {
			fmt.Println("No nodes found")
			return nil
		}

		fmt.Printf("Found %d node(s):\n", len(response.Nodes))
		for _, node := range response.Nodes {
			fmt.Printf("  Name: %s\n", node.Name)
			fmt.Printf("  Namespace: %s\n", node.Namespace)
			fmt.Println()
		}

		return nil
	},
}

func init() {
	sensingCmd.AddCommand(getPositionCmd)
	sensingCmd.AddCommand(listNodesCmd)

	// Add flags
	listNodesCmd.Flags().StringP("filter", "f", "", "Filter nodes by namespace (e.g., namespace:=/sensing)")
}
