package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/drs-api/tools/drs-cli/internal/client"
	ros2bridgev1 "github.com/drs-api/tools/drs-cli/gen/drs/ros2bridge/v1"
)

var recordingCmd = &cobra.Command{
	Use:   "recording",
	Short: "Recording management commands",
	Long:  "Commands for managing recording via RecordingService",
}

var startRecordingCmd = &cobra.Command{
	Use:   "start",
	Short: "Start recording",
	Long:  "Start recording on all ECUs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		response, err := c.StartRecording()
		if err != nil {
			return fmt.Errorf("failed to start recording: %v", err)
		}

		if response.Success {
			fmt.Printf("✓ %s\n", response.Message)
		} else {
			fmt.Printf("✗ %s\n", response.Message)
		}

		return nil
	},
}

var stopRecordingCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop recording",
	Long:  "Stop recording on all ECUs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		response, err := c.StopRecording()
		if err != nil {
			return fmt.Errorf("failed to stop recording: %v", err)
		}

		if response.Success {
			fmt.Printf("✓ %s\n", response.Message)
		} else {
			fmt.Printf("✗ %s\n", response.Message)
		}

		return nil
	},
}

var pauseRecordingCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause recording",
	Long:  "Pause recording on all ECUs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		response, err := c.PauseRecording()
		if err != nil {
			return fmt.Errorf("failed to pause recording: %v", err)
		}

		if response.Success {
			fmt.Printf("✓ %s\n", response.Message)
		} else {
			fmt.Printf("✗ %s\n", response.Message)
		}

		return nil
	},
}

var resumeRecordingCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume recording",
	Long:  "Resume recording on all ECUs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		response, err := c.ResumeRecording()
		if err != nil {
			return fmt.Errorf("failed to resume recording: %v", err)
		}

		if response.Success {
			fmt.Printf("✓ %s\n", response.Message)
		} else {
			fmt.Printf("✗ %s\n", response.Message)
		}

		return nil
	},
}

var getRecordingCmd = &cobra.Command{
	Use:   "get [hardware-id]",
	Short: "Get recording status",
	Long:  "Get recording status for a specific hardware ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		hardwareID := args[0]
		recording, err := c.GetRecording(hardwareID)
		if err != nil {
			return fmt.Errorf("failed to get recording status: %v", err)
		}

		fmt.Printf("Recording Status for %s:\n", recording.HardwareId)
		fmt.Printf("  Is Recording: %t\n", recording.IsRecording)
		fmt.Printf("  Error Level: %s\n", getErrorLevelString(recording.ErrorLevel))
		
		if recording.Header != nil && recording.Header.Stamp != nil {
			timestamp := time.Unix(recording.Header.Stamp.Seconds, int64(recording.Header.Stamp.Nanos))
			fmt.Printf("  Last Update: %s\n", timestamp.Format(time.RFC3339))
		}

		if len(recording.TopicStatuses) > 0 {
			fmt.Printf("  Topic Statuses (%d):\n", len(recording.TopicStatuses))
			for _, topic := range recording.TopicStatuses {
				fmt.Printf("    - %s: %.1f Hz (%s)\n", 
					topic.TopicName, 
					topic.RateHz, 
					getRateStatusString(topic.RateStatus))
			}
		}

		return nil
	},
}

var listRecordingsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all recordings",
	Long:  "List recording status for all ECUs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		filter, _ := cmd.Flags().GetString("filter")
		response, err := c.ListRecordings(filter)
		if err != nil {
			return fmt.Errorf("failed to list recordings: %v", err)
		}

		if len(response.Recordings) == 0 {
			fmt.Println("No recordings found")
			return nil
		}

		fmt.Printf("Found %d recording(s):\n", len(response.Recordings))
		for _, recording := range response.Recordings {
			fmt.Printf("  Hardware ID: %s\n", recording.HardwareId)
			fmt.Printf("  Is Recording: %t\n", recording.IsRecording)
			fmt.Printf("  Error Level: %s\n", getErrorLevelString(recording.ErrorLevel))
			fmt.Printf("  Topics: %d\n", len(recording.TopicStatuses))
			fmt.Println()
		}

		return nil
	},
}

var listTopicStatusesCmd = &cobra.Command{
	Use:   "list-topics [hardware-id]",
	Short: "List topic statuses",
	Long:  "List topic statuses for a specific hardware ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient(cfg.GetServerAddress(), cfg.GetTimeout())
		if err != nil {
			return err
		}
		defer c.Close()

		hardwareID := args[0]
		filter, _ := cmd.Flags().GetString("filter")
		response, err := c.ListTopicStatuses(hardwareID, filter)
		if err != nil {
			return fmt.Errorf("failed to list topic statuses: %v", err)
		}

		if len(response.TopicStatuses) == 0 {
			fmt.Printf("No topic statuses found for hardware ID: %s\n", hardwareID)
			return nil
		}

		fmt.Printf("Topic Statuses for %s (%d):\n", hardwareID, len(response.TopicStatuses))
		for _, topic := range response.TopicStatuses {
			fmt.Printf("  Topic: %s\n", topic.TopicName)
			fmt.Printf("    Type: %s\n", topic.MessageType)
			fmt.Printf("    Rate: %.1f Hz\n", topic.RateHz)
			fmt.Printf("    Status: %s\n", getRateStatusString(topic.RateStatus))
			
			if topic.LastUpdate != nil {
				timestamp := time.Unix(topic.LastUpdate.Seconds, int64(topic.LastUpdate.Nanos))
				fmt.Printf("    Last Update: %s\n", timestamp.Format(time.RFC3339))
			}
			fmt.Println()
		}

		return nil
	},
}

func getErrorLevelString(level ros2bridgev1.Recording_ErrorLevel) string {
	switch level {
	case ros2bridgev1.Recording_ERROR_LEVEL_OK:
		return "OK"
	case ros2bridgev1.Recording_ERROR_LEVEL_WARN:
		return "WARNING"
	case ros2bridgev1.Recording_ERROR_LEVEL_ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func getRateStatusString(status ros2bridgev1.TopicStatus_RateStatus) string {
	switch status {
	case ros2bridgev1.TopicStatus_RATE_STATUS_NORMAL:
		return "NORMAL"
	case ros2bridgev1.TopicStatus_RATE_STATUS_UNKNOWN:
		return "UNKNOWN"
	case ros2bridgev1.TopicStatus_RATE_STATUS_TOO_LOW:
		return "TOO_LOW"
	case ros2bridgev1.TopicStatus_RATE_STATUS_TOO_HIGH:
		return "TOO_HIGH"
	case ros2bridgev1.TopicStatus_RATE_STATUS_NO_MESSAGES:
		return "NO_MESSAGES"
	default:
		return "UNSPECIFIED"
	}
}

func init() {
	recordingCmd.AddCommand(startRecordingCmd)
	recordingCmd.AddCommand(stopRecordingCmd)
	recordingCmd.AddCommand(pauseRecordingCmd)
	recordingCmd.AddCommand(resumeRecordingCmd)
	recordingCmd.AddCommand(getRecordingCmd)
	recordingCmd.AddCommand(listRecordingsCmd)
	recordingCmd.AddCommand(listTopicStatusesCmd)

	// Add flags
	listRecordingsCmd.Flags().StringP("filter", "f", "", "Filter recordings by hardware ID (e.g., hardware_id:=main_ecu)")
	listTopicStatusesCmd.Flags().StringP("filter", "f", "", "Filter topics by name pattern (e.g., topic_name:=/sensing/*)")
}