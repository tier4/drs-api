package cmd

import (
	"github.com/spf13/cobra"
)

var ros2bridgeCmd = &cobra.Command{
	Use:   "ros2bridge",
	Short: "ROS2-bridge service commands",
	Long: `Commands for interacting with the ros2-bridge service.
The ros2-bridge service provides access to ROS2 functionality via gRPC.`,
}

func init() {
	rootCmd.AddCommand(ros2bridgeCmd)
	ros2bridgeCmd.AddCommand(sensingCmd)
	ros2bridgeCmd.AddCommand(recordingCmd)
}
