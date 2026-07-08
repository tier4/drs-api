package rest

import (
	"net/http"
	"regexp"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	modulev1 "github.com/tier4/drs-api/services/api-gateway/gen/drs/module/v1"
	ros2bridgev1 "github.com/tier4/drs-api/services/api-gateway/gen/drs/ros2bridge/v1"
	"github.com/tier4/drs-api/services/api-gateway/internal/grpc"
	"github.com/tier4/drs-api/services/api-gateway/internal/models"
)

// cameraTopicPattern restricts GetCameraPreview to camera topics under the
// known /sensing/camera/ namespace (e.g. /sensing/camera/camera0/image_raw/compressed),
// so the endpoint can't be used to lazily subscribe the bridge to arbitrary
// ROS2 topics.
var cameraTopicPattern = regexp.MustCompile(`^/sensing/camera/[^/]+/image_raw/compressed$`)

// lidarTopicPattern restricts GetPointCloudPreview to raw LiDAR packet
// topics under /sensing/lidar/ (e.g. /sensing/lidar/front/seyond_packets).
// Deliberately vendor-agnostic and broader than this feature's current
// front/right/rear/left UI scope, since the bridge derives the decoded
// "_points" topic by suffix substitution regardless of vendor SDK.
var lidarTopicPattern = regexp.MustCompile(`^/sensing/lidar/[^/]+/[a-z]+_packets$`)

// defaultMaxPointCloudPreviewPoints is the fallback used when max_points is
// missing, zero, or negative. The bridge also enforces this as a hard
// ceiling server-side regardless of what the client requests.
const defaultMaxPointCloudPreviewPoints = 5000

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

// getBridgeOrRespond fetches the ROS2 bridge clients, writing a 503 response
// and returning false if the bridge is unavailable.
func (h *RecordingHandler) getBridgeOrRespond(c *gin.Context) (*grpc.ROS2BridgeClients, bool) {
	ros2Bridge, err := h.clientManager.GetROS2BridgeClients()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "ros2_bridge_unavailable",
			Message: "ROS2 bridge is not available",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return nil, false
	}
	return ros2Bridge, true
}

// GetRecordingStatus handles GET /recording/status - returns recording status of all modules
func (h *RecordingHandler) GetRecordingStatus(c *gin.Context) {
	ros2Bridge, ok := h.getBridgeOrRespond(c)
	if !ok {
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
		}

		// Set recording status
		if recording.IsRecording {
			status.RecordingStatus = "recording"
		} else {
			status.RecordingStatus = "stopped"
		}

		// Map error level enum to data status
		switch recording.ErrorLevel {
		case ros2bridgev1.Recording_ERROR_LEVEL_OK:
			status.DataStatus = "OK"
		case ros2bridgev1.Recording_ERROR_LEVEL_WARN:
			status.DataStatus = "WARN"
		case ros2bridgev1.Recording_ERROR_LEVEL_ERROR:
			status.DataStatus = "ERROR"
		default:
			status.DataStatus = "OK"
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

	ros2Bridge, ok := h.getBridgeOrRespond(c)
	if !ok {
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
			TopicName:   topic.TopicName,
			MessageType: topic.MessageType,
			RateHz:      topic.RateHz,
			Status:      status,
		})
	}

	c.JSON(http.StatusOK, models.TopicStatusResponse{
		Topics: topics,
	})
}

// GetPosition handles GET /modules/{hostname}/position - returns the current
// GPS/INS fix. The bridge serves one vehicle-wide position regardless of
// :hostname (there is a single shared GPS/INS unit, network-visible to every
// ECU's recorder via DDS) - this is intentional, not a routing bug.
func (h *RecordingHandler) GetPosition(c *gin.Context) {
	ros2Bridge, ok := h.getBridgeOrRespond(c)
	if !ok {
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	resp, err := ros2Bridge.Sensing.GetPosition(ctx, &ros2bridgev1.GetPositionRequest{})
	if err != nil {
		if code := grpcstatus.Code(err); code == codes.NotFound || code == codes.Unavailable {
			// No fix received yet, or the cached fix is stale - not an error.
			c.JSON(http.StatusOK, models.PositionResponse{HasData: false})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "position_query_failed",
			Message: err.Error(),
		})
		return
	}

	if !resp.HasData || resp.Position == nil {
		c.JSON(http.StatusOK, models.PositionResponse{HasData: false})
		return
	}

	c.JSON(http.StatusOK, models.PositionResponse{
		HasData: true,
		Position: &models.Position{
			Latitude:               resp.Position.Latitude,
			Longitude:              resp.Position.Longitude,
			Altitude:               resp.Position.Altitude,
			Status:                 resp.Position.NavSatStatus.GetStatus(),
			PositionCovariance:     resp.Position.PositionCovariance,
			PositionCovarianceType: resp.Position.PositionCovarianceType,
		},
	})
}

// GetCameraPreview handles GET /modules/{hostname}/camera/preview?topic=<name> -
// returns a resized JPEG frame for the given camera topic. The response is
// always Content-Type: image/jpeg; an X-Has-Data header signals whether a
// real frame or a placeholder is being returned.
func (h *RecordingHandler) GetCameraPreview(c *gin.Context) {
	topicName := c.Query("topic")
	if topicName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_topic",
			Message: "topic query parameter is required",
		})
		return
	}
	if !cameraTopicPattern.MatchString(topicName) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_topic",
			Message: "topic must be a camera topic under /sensing/camera/",
		})
		return
	}

	ros2Bridge, ok := h.getBridgeOrRespond(c)
	if !ok {
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	resp, err := ros2Bridge.Sensing.GetCameraPreview(ctx, &ros2bridgev1.GetCameraPreviewRequest{
		TopicName: topicName,
	})
	if err != nil {
		if grpcstatus.Code(err) == codes.InvalidArgument {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_topic",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "camera_preview_failed",
			Message: err.Error(),
		})
		return
	}

	if !resp.HasData {
		c.Header("X-Has-Data", "false")
		c.Data(http.StatusOK, "image/jpeg", []byte{})
		return
	}

	c.Header("X-Has-Data", "true")
	c.Data(http.StatusOK, resp.ContentType, resp.ImageData)
}

// GetPointCloudPreview handles GET /modules/{hostname}/lidar/preview?topic=<name>&max_points=<n> -
// returns a decimated point cloud frame (interleaved x,y,z,intensity float32,
// 16 bytes/point) for the given LiDAR "_packets" topic. The bridge derives
// the decoded "_points" topic itself. The response is always
// Content-Type: application/octet-stream; X-Has-Data and X-Decoder-Running
// headers let the UI distinguish "decoder not started" from "waiting for
// first frame."
func (h *RecordingHandler) GetPointCloudPreview(c *gin.Context) {
	topicName := c.Query("topic")
	if topicName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_topic",
			Message: "topic query parameter is required",
		})
		return
	}
	if !lidarTopicPattern.MatchString(topicName) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_topic",
			Message: "topic must be a LiDAR packets topic under /sensing/lidar/",
		})
		return
	}

	maxPoints := defaultMaxPointCloudPreviewPoints
	if raw := c.Query("max_points"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			maxPoints = parsed
		}
	}

	ros2Bridge, ok := h.getBridgeOrRespond(c)
	if !ok {
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	resp, err := ros2Bridge.Sensing.GetPointCloudPreview(ctx, &ros2bridgev1.GetPointCloudPreviewRequest{
		TopicName: topicName,
		MaxPoints: int32(maxPoints),
	})
	if err != nil {
		if grpcstatus.Code(err) == codes.InvalidArgument {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_topic",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "point_cloud_preview_failed",
			Message: err.Error(),
		})
		return
	}

	c.Header("X-Decoder-Running", strconv.FormatBool(resp.DecoderRunning))

	if !resp.HasData {
		c.Header("X-Has-Data", "false")
		c.Data(http.StatusOK, "application/octet-stream", []byte{})
		return
	}

	c.Header("X-Has-Data", "true")
	c.Data(http.StatusOK, resp.ContentType, resp.PointData)
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
