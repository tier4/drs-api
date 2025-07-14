package rest

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/drs-api/services/api-gateway/internal/grpc"
	"github.com/drs-api/services/api-gateway/internal/models"
	modulev1 "github.com/drs-api/services/api-gateway/drs/module/v1"
	ros2bridgev1 "github.com/drs-api/services/api-gateway/drs/ros2bridge/v1"
)

// ECUHandler handles ECU-related REST API endpoints
type ECUHandler struct {
	clientManager *grpc.ClientManager
}

// NewECUHandler creates a new ECU handler
func NewECUHandler(clientManager *grpc.ClientManager) *ECUHandler {
	return &ECUHandler{
		clientManager: clientManager,
	}
}

// GetAllECUs handles GET /ecus - returns status of all ECUs
func (h *ECUHandler) GetAllECUs(c *gin.Context) {
	ecuNames := h.clientManager.GetECUNames()
	ecuStatuses := make([]models.ECUStatus, 0, len(ecuNames))

	// Use goroutines to fetch ECU statuses concurrently
	var wg sync.WaitGroup
	statusChan := make(chan models.ECUStatus, len(ecuNames))

	for _, hostname := range ecuNames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			status := h.getECUStatus(hostname)
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
		ecuStatuses = append(ecuStatuses, status)
	}

	c.JSON(http.StatusOK, models.ECUListResponse{
		ECUs: ecuStatuses,
	})
}

// GetECU handles GET /ecus/{hostname} - returns status of a single ECU
func (h *ECUHandler) GetECU(c *gin.Context) {
	hostname := c.Param("hostname")
	
	status := h.getECUStatus(hostname)
	if status.Status == "ERROR" && status.StatusDetail.Services.DRSSensor == "" {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "ecu_not_found",
			Message: "ECU not found or unreachable",
			Details: map[string]interface{}{
				"hostname": hostname,
			},
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// getECUStatus fetches the status of a single ECU
func (h *ECUHandler) getECUStatus(hostname string) models.ECUStatus {
	clients, err := h.clientManager.GetECUClients(hostname)
	if err != nil {
		return models.ECUStatus{
			Hostname:    hostname,
			Status:      "ERROR",
			StatusDetail: models.ECUStatusDetail{},
			LastUpdated: time.Now(),
		}
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	status := models.ECUStatus{
		Hostname:    hostname,
		Status:      "OK",
		LastUpdated: time.Now(),
	}

	// Get environment variables
	if envResp, err := clients.Monitoring.GetEnvironment(ctx, &modulev1.GetEnvironmentRequest{}); err == nil {
		status.Environment = models.EnvironmentInfo{
			SensingSystemID: envResp.SensingSystemId,
			ModuleID:        envResp.ModuleId,
		}
	}

	// Get disk usage
	if h.clientManager.IsServiceEnabled(hostname, "disk") {
		if diskResp, err := clients.Monitoring.GetDiskUsage(ctx, &modulev1.GetDiskUsageRequest{}); err == nil {
			status.Disk = models.DiskInfo{
				UsagePercentage: diskResp.DiskUsage.UsagePercentage,
				FreeBytes:       diskResp.DiskUsage.FreeBytes,
				TotalBytes:      diskResp.DiskUsage.TotalBytes,
			}
		}
	}

	// Get PTP status
	if h.clientManager.IsServiceEnabled(hostname, "ptp") {
		if ptpResp, err := clients.Monitoring.GetPTPStatus(ctx, &modulev1.GetPTPStatusRequest{}); err == nil {
			status.StatusDetail.PTP = models.PTPInfo{
				OffsetNs: ptpResp.LocalStatus.MasterOffsetNs,
			}
		}
	}

	// Get service status
	if h.clientManager.IsServiceEnabled(hostname, "services") {
		serviceStatus := h.getServiceStatus(clients, hostname)
		status.StatusDetail.Services = serviceStatus
	}

	// Get recording status
	if clients.Recording != nil {
		recordingStatus := h.getRecordingStatus(clients, hostname)
		status.StatusDetail.Recording = recordingStatus
	}

	// Determine overall status
	status.Status = h.determineOverallStatus(status)

	return status
}

// getServiceStatus fetches service status from ECU
func (h *ECUHandler) getServiceStatus(clients *grpc.ECUClients, hostname string) models.ServiceStatus {
	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	serviceStatus := models.ServiceStatus{
		DRSSensor:   "unknown",
		DRSRecorder: "unknown",
	}

	// Get list of services
	if listResp, err := clients.ServiceManager.ListServices(ctx, &modulev1.ListServicesRequest{}); err == nil {
		for _, service := range listResp.Services {
			switch service.Name {
			case "services/drs_sensor":
				serviceStatus.DRSSensor = h.convertServiceState(service.State)
			case "services/drs_recorder":
				serviceStatus.DRSRecorder = h.convertServiceState(service.State)
			}
		}
	}

	return serviceStatus
}

// getRecordingStatus fetches recording status from ECU
func (h *ECUHandler) getRecordingStatus(clients *grpc.ECUClients, hostname string) models.RecordingInfo {
	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	recordingInfo := models.RecordingInfo{
		Status: "unknown",
		Active: false,
	}

	// Get recording status
	if listResp, err := clients.Recording.ListRecordings(ctx, &ros2bridgev1.ListRecordingsRequest{}); err == nil {
		for _, recording := range listResp.Recordings {
			if recording.IsRecording {
				recordingInfo.Status = "recording"
				recordingInfo.Active = true
				break
			} else {
				recordingInfo.Status = "stopped"
				recordingInfo.Active = false
			}
		}
	}

	return recordingInfo
}

// convertServiceState converts gRPC service state to string
func (h *ECUHandler) convertServiceState(state modulev1.Service_ServiceState) string {
	switch state {
	case modulev1.Service_SERVICE_STATE_ACTIVE:
		return "active"
	case modulev1.Service_SERVICE_STATE_INACTIVE:
		return "inactive"
	case modulev1.Service_SERVICE_STATE_FAILED:
		return "failed"
	case modulev1.Service_SERVICE_STATE_ACTIVATING:
		return "activating"
	case modulev1.Service_SERVICE_STATE_DEACTIVATING:
		return "deactivating"
	default:
		return "unknown"
	}
}

// determineOverallStatus determines the overall status based on various factors
func (h *ECUHandler) determineOverallStatus(status models.ECUStatus) string {
	// Check if any critical services are failed
	if status.StatusDetail.Services.DRSSensor == "failed" || status.StatusDetail.Services.DRSRecorder == "failed" {
		return "ERROR"
	}

	// Check PTP offset (consider synced if offset is within acceptable range)
	ptpOffsetThreshold := int64(1000000) // 1ms in nanoseconds
	if status.StatusDetail.PTP.OffsetNs > ptpOffsetThreshold || status.StatusDetail.PTP.OffsetNs < -ptpOffsetThreshold {
		return "WARN"
	}

	// Check if any services are inactive
	if status.StatusDetail.Services.DRSSensor == "inactive" || status.StatusDetail.Services.DRSRecorder == "inactive" {
		return "WARN"
	}

	// Check disk usage
	if status.Disk.UsagePercentage > 90 {
		return "ERROR"
	} else if status.Disk.UsagePercentage > 80 {
		return "WARN"
	}

	return "OK"
}