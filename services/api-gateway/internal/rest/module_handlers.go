package rest

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tier4/drs-api/services/api-gateway/internal/grpc"
	"github.com/tier4/drs-api/services/api-gateway/internal/models"
	modulev1 "github.com/tier4/drs-api/services/api-gateway/gen/drs/module/v1"
	ros2bridgev1 "github.com/tier4/drs-api/services/api-gateway/gen/drs/ros2bridge/v1"
)

// ModuleHandler handles module-related REST API endpoints
type ModuleHandler struct {
	clientManager *grpc.ClientManager
}

// NewModuleHandler creates a new module handler
func NewModuleHandler(clientManager *grpc.ClientManager) *ModuleHandler {
	return &ModuleHandler{
		clientManager: clientManager,
	}
}

// GetAllModules handles GET /modules - returns status of all modules
func (h *ModuleHandler) GetAllModules(c *gin.Context) {
	moduleNames := h.clientManager.GetModuleNames()
	moduleStatuses := make([]models.ModuleStatus, 0, len(moduleNames))

	// Use goroutines to fetch module statuses concurrently
	var wg sync.WaitGroup
	statusChan := make(chan models.ModuleStatus, len(moduleNames))

	for _, hostname := range moduleNames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			status := h.getModuleStatus(hostname)
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
		moduleStatuses = append(moduleStatuses, status)
	}

	c.JSON(http.StatusOK, models.ModuleListResponse{
		Modules: moduleStatuses,
	})
}

// GetModule handles GET /modules/{hostname} - returns status of a single module
func (h *ModuleHandler) GetModule(c *gin.Context) {
	hostname := c.Param("hostname")

	status := h.getModuleStatus(hostname)
	if status.Status == "ERROR" && status.StatusDetail.Services.DRSSensor == "" {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "module_not_found",
			Message: "Module not found or unreachable",
			Details: map[string]interface{}{
				"hostname": hostname,
			},
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// getModuleStatus fetches the status of a single module
func (h *ModuleHandler) getModuleStatus(hostname string) models.ModuleStatus {
	clients, err := h.clientManager.GetModuleClients(hostname)
	if err != nil {
		return models.ModuleStatus{
			Hostname:    hostname,
			Status:      "ERROR",
			StatusDetail: models.ModuleStatusDetail{},
			LastUpdated: time.Now(),
		}
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	// Get enabled services for this module
	moduleConfig := h.clientManager.GetConfig().Modules[hostname]

	status := models.ModuleStatus{
		Hostname:        hostname,
		Status:          "OK",
		EnabledServices: moduleConfig.EnabledServices,
		LastUpdated:     time.Now(),
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
		// Try to list disks (new API)
		if disksResp, err := clients.Monitoring.ListDisks(ctx, &modulev1.ListDisksRequest{}); err == nil {
			status.Disks = make([]models.DiskDetail, 0, len(disksResp.Disks))
			for _, disk := range disksResp.Disks {
				// Extract disk name from resource name (e.g., "disks/internal" -> "internal")
				diskName := disk.Name
				if len(diskName) > 6 && diskName[:6] == "disks/" {
					diskName = diskName[6:]
				}

				status.Disks = append(status.Disks, models.DiskDetail{
					Name:            diskName,
					MountPath:       disk.MountPath,
					Description:     disk.Description,
					UsagePercentage: disk.Usage.UsagePercentage,
					FreeBytes:       disk.Usage.FreeBytes,
					TotalBytes:      disk.Usage.TotalBytes,
				})
			}

			// Populate legacy field with first disk or primary if available
			if len(status.Disks) > 0 {
				status.Disk = models.DiskInfo{
					UsagePercentage: status.Disks[0].UsagePercentage,
					FreeBytes:       status.Disks[0].FreeBytes,
					TotalBytes:      status.Disks[0].TotalBytes,
				}
			}
		} else {
			// Fallback to legacy API
			if diskResp, err := clients.Monitoring.GetDiskUsage(ctx, &modulev1.GetDiskUsageRequest{}); err == nil {
				status.Disk = models.DiskInfo{
					UsagePercentage: diskResp.DiskUsage.UsagePercentage,
					FreeBytes:       diskResp.DiskUsage.FreeBytes,
					TotalBytes:      diskResp.DiskUsage.TotalBytes,
				}
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

	// Get recording status from ROS2 bridge
	if hostname != "nas" { // NAS doesn't have recording capability
		if recordingStatus := h.getRecordingStatus(hostname); recordingStatus != nil {
			status.StatusDetail.Recording = *recordingStatus
		}
	}

	// Determine overall status
	status.Status = h.determineOverallStatus(status)

	return status
}

// getServiceStatus fetches service status from module
func (h *ModuleHandler) getServiceStatus(clients *grpc.ModuleClients, hostname string) models.ServiceStatus {
	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	serviceStatus := models.ServiceStatus{
		DRSSensor:       "unknown",
		DRSRecorder:     "unknown",
		DRSTransfer:     "unknown",
		DRSTransferring: "unknown",
	}

	// Set both timer and service states independently; only when ListServices succeeds so that
	// a connectivity failure keeps "unknown" (matching DRSSensor/DRSRecorder behavior).
	if listResp, err := clients.ServiceManager.ListServices(ctx, &modulev1.ListServicesRequest{}); err == nil {
		var transferServiceState string
		for _, service := range listResp.Services {
			switch service.Name {
			case "services/drs_sensor":
				serviceStatus.DRSSensor = h.convertServiceState(service.State)
			case "services/drs_recorder":
				serviceStatus.DRSRecorder = h.convertServiceState(service.State)
			case "services/drs_transfer":
				serviceStatus.DRSTransfer = h.convertServiceState(service.State)
			case "services/drs_transfer_service":
				transferServiceState = h.convertServiceState(service.State)
			}
		}
		if transferServiceState != "" {
			serviceStatus.DRSTransferring = deriveTransferringStatus(transferServiceState)
		}
	}

	return serviceStatus
}

// deriveTransferringStatus maps drs-transfer.service systemd state to a UI-facing transferring state.
// Type=oneshot reports "activating" (not "active") while running, so both are treated as transferring.
func deriveTransferringStatus(serviceState string) string {
	switch serviceState {
	case "active", "activating":
		return "transferring"
	case "failed":
		return "failed"
	default:
		return "stopped"
	}
}


// getRecordingStatus fetches recording status from ROS2 bridge
func (h *ModuleHandler) getRecordingStatus(hostname string) *models.RecordingInfo {
	ros2Bridge, err := h.clientManager.GetROS2BridgeClients()
	if err != nil {
		return nil
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	// Get module environment to find hardware ID
	clients, err := h.clientManager.GetModuleClients(hostname)
	if err != nil {
		return nil
	}

	envResp, err := clients.Monitoring.GetEnvironment(ctx, &modulev1.GetEnvironmentRequest{})
	if err != nil {
		return nil
	}

	// Get recording status from ROS2 bridge
	resp, err := ros2Bridge.Recording.ListRecordings(ctx, &ros2bridgev1.ListRecordingsRequest{})
	if err != nil {
		return nil
	}

	// Find recording for this module by hardware ID
	// Try matching by both module_id and hostname
	for _, recording := range resp.Recordings {
		if recording.HardwareId == envResp.ModuleId || recording.HardwareId == hostname {
			status := "stopped"
			if recording.IsRecording {
				status = "recording"
			}

			// Map error level to data status
			var dataStatus string
			switch recording.ErrorLevel {
			case ros2bridgev1.Recording_ERROR_LEVEL_OK:
				dataStatus = "OK"
			case ros2bridgev1.Recording_ERROR_LEVEL_WARN:
				dataStatus = "WARN"
			case ros2bridgev1.Recording_ERROR_LEVEL_ERROR:
				dataStatus = "ERROR"
			default:
				// If error level is not recognized, default to OK
				dataStatus = "OK"
			}

			return &models.RecordingInfo{
				Status:     status,
				DataStatus: dataStatus,
			}
		}
	}

	// Default to stopped if not found
	return &models.RecordingInfo{
		Status:     "stopped",
		DataStatus: "OK",
	}
}

// convertServiceState converts gRPC service state to string
func (h *ModuleHandler) convertServiceState(state modulev1.Service_ServiceState) string {
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
func (h *ModuleHandler) determineOverallStatus(status models.ModuleStatus) string {
	// Check if any critical services are failed
	if status.StatusDetail.Services.DRSSensor == "failed" ||
		status.StatusDetail.Services.DRSRecorder == "failed" ||
		status.StatusDetail.Services.DRSTransfer == "failed" ||
		status.StatusDetail.Services.DRSTransferring == "failed" {
		return "ERROR"
	}

	// Check PTP offset (consider synced if offset is within acceptable range)
	ptpOffsetThreshold := int64(1000000) // 1ms in nanoseconds
	if status.StatusDetail.PTP.OffsetNs > ptpOffsetThreshold || status.StatusDetail.PTP.OffsetNs < -ptpOffsetThreshold {
		return "WARN"
	}

	// Check if any services are inactive
	// Note: DRSTransfer "inactive" is the normal stopped state and is intentionally not checked here.
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
