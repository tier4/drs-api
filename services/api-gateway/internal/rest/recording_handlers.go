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
	moduleNames := h.clientManager.GetModuleNames()
	recordingStatuses := make([]models.RecordingStatus, 0)

	var wg sync.WaitGroup
	statusChan := make(chan models.RecordingStatus, len(moduleNames))

	for _, hostname := range moduleNames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			
			clients, err := h.clientManager.GetModuleClients(hostname)
			if err != nil || clients.Recording == nil {
				statusChan <- models.RecordingStatus{
					Hostname: hostname,
					Status:   "unknown",
					Active:   false,
				}
				return
			}

			ctx, cancel := h.clientManager.GetContext()
			defer cancel()

			// Get recording status
			resp, err := clients.Recording.ListRecordings(ctx, &ros2bridgev1.ListRecordingsRequest{})
			if err != nil {
				statusChan <- models.RecordingStatus{
					Hostname: hostname,
					Status:   "error",
					Active:   false,
				}
				return
			}

			// Find active recording
			status := models.RecordingStatus{
				Hostname: hostname,
				Status:   "stopped",
				Active:   false,
			}

			for _, recording := range resp.Recordings {
				if recording.IsRecording {
					status.Status = "recording"
					status.Active = true
					status.HardwareID = recording.HardwareId
					break
				} else {
					status.Status = "stopped"
					status.Active = false
					status.HardwareID = recording.HardwareId
				}
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

	var wg sync.WaitGroup
	statusChan := make(chan models.PTPStatus, len(moduleNames))

	for _, hostname := range moduleNames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			
			if !h.clientManager.IsServiceEnabled(hostname, "ptp") {
				statusChan <- models.PTPStatus{
					Hostname: hostname,
					LocalStatus: models.PTPLocalInfo{},
				}
				return
			}

			clients, err := h.clientManager.GetModuleClients(hostname)
			if err != nil {
				statusChan <- models.PTPStatus{
					Hostname: hostname,
					LocalStatus: models.PTPLocalInfo{},
				}
				return
			}

			ctx, cancel := h.clientManager.GetContext()
			defer cancel()

			// Get PTP status
			resp, err := clients.Monitoring.GetPTPStatus(ctx, &modulev1.GetPTPStatusRequest{
				IncludeRemoteDevices: true,
			})
			if err != nil {
				statusChan <- models.PTPStatus{
					Hostname: hostname,
					LocalStatus: models.PTPLocalInfo{},
				}
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
	
	clients, err := h.clientManager.GetModuleClients(hostname)
	if err != nil || clients.Recording == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "module_not_found",
			Message: "Module not found or ROS2 bridge not available",
			Details: map[string]interface{}{
				"hostname": hostname,
			},
		})
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	// Get topic statuses
	resp, err := clients.Recording.ListTopicStatuses(ctx, &ros2bridgev1.ListTopicStatusesRequest{
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
			ExpectedRateHz: 30.0, // Default expected rate
			Status:         status,
		})
	}

	c.JSON(http.StatusOK, models.TopicStatusResponse{
		Topics: topics,
	})
}

// performRecordingOperation performs a recording operation on all modules with ROS2 bridge
func (h *RecordingHandler) performRecordingOperation(c *gin.Context, operation string) {
	moduleNames := h.clientManager.GetModuleNames()
	
	var wg sync.WaitGroup
	results := make(chan models.RecordingOperationResponse, len(moduleNames))

	for _, hostname := range moduleNames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			
			clients, err := h.clientManager.GetModuleClients(hostname)
			if err != nil || clients.Recording == nil {
				results <- models.RecordingOperationResponse{
					Success: false,
					Message: "ECU not found or ROS2 bridge not available",
				}
				return
			}

			ctx, cancel := h.clientManager.GetContext()
			defer cancel()

			switch operation {
			case "start":
				resp, err := clients.Recording.StartRecording(ctx, &ros2bridgev1.StartRecordingRequest{})
				if err != nil {
					results <- models.RecordingOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.RecordingOperationResponse{
						Success: resp.Success,
						Message: resp.Message,
						Status:  "recording",
					}
				}
			case "stop":
				resp, err := clients.Recording.StopRecording(ctx, &ros2bridgev1.StopRecordingRequest{})
				if err != nil {
					results <- models.RecordingOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.RecordingOperationResponse{
						Success: resp.Success,
						Message: resp.Message,
						Status:  "stopped",
					}
				}
			case "pause":
				resp, err := clients.Recording.PauseRecording(ctx, &ros2bridgev1.PauseRecordingRequest{})
				if err != nil {
					results <- models.RecordingOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.RecordingOperationResponse{
						Success: resp.Success,
						Message: resp.Message,
						Status:  "paused",
					}
				}
			case "resume":
				resp, err := clients.Recording.ResumeRecording(ctx, &ros2bridgev1.ResumeRecordingRequest{})
				if err != nil {
					results <- models.RecordingOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.RecordingOperationResponse{
						Success: resp.Success,
						Message: resp.Message,
						Status:  "recording",
					}
				}
			}
		}(hostname)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Check results
	allSuccess := true
	messages := []string{}
	for result := range results {
		if !result.Success {
			allSuccess = false
		}
		messages = append(messages, result.Message)
	}

	if allSuccess {
		c.JSON(http.StatusOK, models.RecordingOperationResponse{
			Success: true,
			Message: "Recording operation completed successfully",
			Status:  operation,
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "recording_operation_failed",
			Message: "Some modules failed to perform the recording operation",
			Details: map[string]interface{}{
				"messages": messages,
			},
		})
	}
}