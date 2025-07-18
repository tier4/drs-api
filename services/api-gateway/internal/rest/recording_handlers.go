package rest

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/tier4/drs-api/services/api-gateway/internal/grpc"
	"github.com/tier4/drs-api/services/api-gateway/internal/models"
	modulev1 "github.com/tier4/drs-api/services/api-gateway/drs/module/v1"
	ros2bridgev1 "github.com/tier4/drs-api/services/api-gateway/drs/ros2bridge/v1"
)

// RecordingHandler handles recording control REST API endpoints
type RecordingHandler struct {
	clientManager *grpc.ClientManager
}

// NewRecordingHandler creates a new recording handler
func NewRecordingHandler(clientManager *grpc.ClientManager) *RecordingHandler {
	return &RecordingHandler{
		clientManager: clientManager,
	}
}

// GetRecordingStatus handles GET /recording/status - returns recording status of all modules
func (h *RecordingHandler) GetRecordingStatus(c *gin.Context) {
	// Get ROS2 bridge clients
	ros2Bridge, err := h.clientManager.GetROS2BridgeClients()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "ros2_bridge_unavailable",
			Message: "ROS2 bridge is not available",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	// Get recording status from ROS2 bridge
	resp, err := ros2Bridge.Recording.ListRecordings(ctx, &ros2bridgev1.ListRecordingsRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "recording_query_failed",
			Message: "Failed to query recording status",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	// Convert to response format
	recordingStatuses := make([]models.RecordingStatus, 0, len(resp.Recordings))
	for _, recording := range resp.Recordings {
		status := models.RecordingStatus{
			Hostname:   recording.HardwareId, // Use hardware_id as hostname
			HardwareID: recording.HardwareId,
			Active:     recording.IsRecording,
		}

		if recording.IsRecording {
			status.Status = "recording"
		} else {
			status.Status = "stopped"
		}

		recordingStatuses = append(recordingStatuses, status)
	}

	c.JSON(http.StatusOK, models.RecordingStatusResponse{
		RecordingStatus: recordingStatuses,
	})
}

// StartRecording handles POST /recording/start - starts recording on all ECUs
func (h *RecordingHandler) StartRecording(c *gin.Context) {
	h.performRecordingOperation(c, "start")
}

// StopRecording handles POST /recording/stop - stops recording on all ECUs
func (h *RecordingHandler) StopRecording(c *gin.Context) {
	h.performRecordingOperation(c, "stop")
}

// PauseRecording handles POST /recording/pause - pauses recording on all ECUs
func (h *RecordingHandler) PauseRecording(c *gin.Context) {
	h.performRecordingOperation(c, "pause")
}

// ResumeRecording handles POST /recording/resume - resumes recording on all ECUs
func (h *RecordingHandler) ResumeRecording(c *gin.Context) {
	h.performRecordingOperation(c, "resume")
}

// GetPTPStatus handles GET /ptp/status - returns PTP status of all ECUs
func (h *RecordingHandler) GetPTPStatus(c *gin.Context) {
	moduleNames := h.clientManager.GetModuleNames()
	ptpStatuses := make([]models.PTPStatus, 0)

	// Filter modules that have PTP service enabled
	ptpEnabledModules := make([]string, 0)
	for _, hostname := range moduleNames {
		if h.clientManager.IsServiceEnabled(hostname, "ptp") {
			ptpEnabledModules = append(ptpEnabledModules, hostname)
		}
	}

	var wg sync.WaitGroup
	statusChan := make(chan models.PTPStatus, len(ptpEnabledModules))

	for _, hostname := range ptpEnabledModules {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()

			clients, err := h.clientManager.GetModuleClients(hostname)
			if err != nil {
				// Skip modules that are not accessible
				return
			}

			ctx, cancel := h.clientManager.GetContext()
			defer cancel()

			// Get PTP status
			resp, err := clients.Monitoring.GetPTPStatus(ctx, &modulev1.GetPTPStatusRequest{
				IncludeRemoteDevices: true,
			})
			if err != nil {
				// Skip modules that fail to respond
				return
			}

			// Convert to response model
			status := models.PTPStatus{
				Hostname: hostname,
				LocalStatus: models.PTPLocalInfo{
					ClockID:        resp.LocalStatus.ClockId,
					MasterOffsetNs: resp.LocalStatus.MasterOffsetNs,
					GMPresent:      resp.LocalStatus.GmPresent,
				},
				RemoteStatus: make([]models.PTPRemoteInfo, 0),
			}

			for _, remote := range resp.RemoteStatuses {
				remoteInfo := models.PTPRemoteInfo{
					DeviceName:  remote.DeviceName,
					IPAddress:   remote.IpAddress,
					IsReachable: remote.IsReachable,
				}

				if remote.IsReachable && remote.Status != nil {
					remoteInfo.Status = &models.PTPLocalInfo{
						ClockID:        remote.Status.ClockId,
						MasterOffsetNs: remote.Status.MasterOffsetNs,
						GMPresent:      remote.Status.GmPresent,
					}
				} else {
					remoteInfo.ErrorMessage = remote.ErrorMessage
				}

				status.RemoteStatus = append(status.RemoteStatus, remoteInfo)
			}

			statusChan <- status
		}(hostname)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(statusChan)
	}()

	// Collect results
	for status := range statusChan {
		ptpStatuses = append(ptpStatuses, status)
	}

	c.JSON(http.StatusOK, models.PTPStatusResponse{
		PTPStatus: ptpStatuses,
	})
}

// GetTopicStatus handles GET /ecus/{hostname}/topics/status - returns topic status for a single ECU
func (h *RecordingHandler) GetTopicStatus(c *gin.Context) {
	hostname := c.Param("hostname")
	
	// Get ROS2 bridge clients
	ros2Bridge, err := h.clientManager.GetROS2BridgeClients()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "ros2_bridge_unavailable",
			Message: "ROS2 bridge is not available",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	// Get topic statuses
	resp, err := ros2Bridge.Recording.ListTopicStatuses(ctx, &ros2bridgev1.ListTopicStatusesRequest{
		HardwareId: hostname, // Use hostname as hardware ID
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "topic_status_failed",
			Message: err.Error(),
		})
		return
	}

	// Convert to response model
	topics := make([]models.TopicStatus, 0, len(resp.TopicStatuses))
	for _, topic := range resp.TopicStatuses {
		status := "OK"
		if topic.RateStatus == ros2bridgev1.TopicStatus_RATE_STATUS_TOO_LOW {
			status = "WARN"
		} else if topic.RateStatus == ros2bridgev1.TopicStatus_RATE_STATUS_NO_MESSAGES {
			status = "ERROR"
		}

		topics = append(topics, models.TopicStatus{
			TopicName:      topic.TopicName,
			RateHz:         topic.RateHz,
			Status:         status,
		})
	}

	c.JSON(http.StatusOK, models.TopicStatusResponse{
		Topics: topics,
	})
}

// performRecordingOperation performs a recording operation via ROS2 bridge
func (h *RecordingHandler) performRecordingOperation(c *gin.Context, operation string) {
	// Get ROS2 bridge clients
	ros2Bridge, err := h.clientManager.GetROS2BridgeClients()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.RecordingOperationResponse{
			Success: false,
			Message: "ROS2 bridge is not available: " + err.Error(),
		})
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	var success bool
	var message string

	switch operation {
	case "start":
		resp, err := ros2Bridge.Recording.StartRecording(ctx, &ros2bridgev1.StartRecordingRequest{})
		if err != nil {
			success = false
			message = "Failed to start recording: " + err.Error()
		} else {
			success = resp.Success
			message = resp.Message
		}
	case "stop":
		resp, err := ros2Bridge.Recording.StopRecording(ctx, &ros2bridgev1.StopRecordingRequest{})
		if err != nil {
			success = false
			message = "Failed to stop recording: " + err.Error()
		} else {
			success = resp.Success
			message = resp.Message
		}
	case "pause":
		resp, err := ros2Bridge.Recording.PauseRecording(ctx, &ros2bridgev1.PauseRecordingRequest{})
		if err != nil {
			success = false
			message = "Failed to pause recording: " + err.Error()
		} else {
			success = resp.Success
			message = resp.Message
		}
	case "resume":
		resp, err := ros2Bridge.Recording.ResumeRecording(ctx, &ros2bridgev1.ResumeRecordingRequest{})
		if err != nil {
			success = false
			message = "Failed to resume recording: " + err.Error()
		} else {
			success = resp.Success
			message = resp.Message
		}
	default:
		success = false
		message = "Unknown operation: " + operation
	}

	c.JSON(http.StatusOK, models.RecordingOperationResponse{
		Success: success,
		Message: message,
	})
}