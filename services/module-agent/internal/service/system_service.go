package service

import (
	"context"
	"fmt"
	"log"

	"github.com/drs-api/services/module-agent/internal/config"
	systemv1 "github.com/drs-api/services/module-agent/gen/system/v1"
	"github.com/drs-api/services/module-agent/internal/ptp"
	"github.com/drs-api/services/module-agent/internal/storage"
	"github.com/drs-api/services/module-agent/internal/system"
)

type SystemService struct {
	systemv1.UnimplementedSystemServiceServer
	systemManager  *system.Manager
	storageManager *storage.Manager
	ptpChecker     *ptp.Checker
	config         *config.Config
}

func NewSystemService(cfg *config.Config) *SystemService {
	return &SystemService{
		systemManager:  system.NewManager(),
		storageManager: storage.NewManager(),
		ptpChecker:     ptp.NewChecker(cfg.PTP.SyncThresholdNs),
		config:         cfg,
	}
}

func (s *SystemService) Reboot(ctx context.Context, req *systemv1.RebootRequest) (*systemv1.RebootResponse, error) {
	log.Printf("Reboot request received with delay: %d seconds", req.DelaySeconds)
	
	// Check if reboot API is enabled
	if !s.config.APIs.EnableReboot {
		log.Printf("Reboot API is disabled")
		return &systemv1.RebootResponse{
			Success: false,
			Message: "Reboot API is disabled in configuration",
		}, nil
	}
	
	err := s.systemManager.Reboot(int(req.DelaySeconds))
	if err != nil {
		log.Printf("Reboot failed: %v", err)
		return &systemv1.RebootResponse{
			Success: false,
			Message: fmt.Sprintf("Reboot failed: %v", err),
		}, nil
	}

	return &systemv1.RebootResponse{
		Success: true,
		Message: fmt.Sprintf("Reboot scheduled with %d seconds delay", req.DelaySeconds),
	}, nil
}

func (s *SystemService) Shutdown(ctx context.Context, req *systemv1.ShutdownRequest) (*systemv1.ShutdownResponse, error) {
	log.Printf("Shutdown request received with delay: %d seconds", req.DelaySeconds)
	
	// Check if shutdown API is enabled
	if !s.config.APIs.EnableShutdown {
		log.Printf("Shutdown API is disabled")
		return &systemv1.ShutdownResponse{
			Success: false,
			Message: "Shutdown API is disabled in configuration",
		}, nil
	}
	
	err := s.systemManager.Shutdown(int(req.DelaySeconds))
	if err != nil {
		log.Printf("Shutdown failed: %v", err)
		return &systemv1.ShutdownResponse{
			Success: false,
			Message: fmt.Sprintf("Shutdown failed: %v", err),
		}, nil
	}

	return &systemv1.ShutdownResponse{
		Success: true,
		Message: fmt.Sprintf("Shutdown scheduled with %d seconds delay", req.DelaySeconds),
	}, nil
}


func (s *SystemService) GetDiskUsage(ctx context.Context, req *systemv1.GetDiskUsageRequest) (*systemv1.GetDiskUsageResponse, error) {
	log.Printf("Disk usage request received")
	
	// Check if disk usage API is enabled
	if !s.config.APIs.EnableDiskUsage {
		log.Printf("Disk usage API is disabled")
		return &systemv1.GetDiskUsageResponse{
			Success: false,
			Message: "Disk usage API is disabled in configuration",
		}, nil
	}
	
	// Always use the configured monitor path - one disk per ECU
	targetPath := s.config.GetDiskPath()
	log.Printf("Getting disk usage for primary disk: %s", targetPath)
	
	usage, err := s.storageManager.GetDiskUsage(targetPath)
	if err != nil {
		return &systemv1.GetDiskUsageResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get disk usage: %v", err),
		}, nil
	}

	return &systemv1.GetDiskUsageResponse{
		Success: true,
		Message: "Disk usage retrieved successfully",
		DiskUsage: &systemv1.DiskUsage{
			TotalBytes:      usage.TotalBytes,
			UsedBytes:       usage.UsedBytes,
			FreeBytes:       usage.FreeBytes,
			UsagePercentage: usage.UsagePercentage,
		},
	}, nil
}

func (s *SystemService) CheckPTPSync(ctx context.Context, req *systemv1.CheckPTPSyncRequest) (*systemv1.CheckPTPSyncResponse, error) {
	log.Printf("PTP sync check request received (include_remote: %v)", req.IncludeRemoteDevices)
	
	// Check if PTP check API is enabled
	if !s.config.APIs.EnablePTPCheck {
		log.Printf("PTP check API is disabled")
		return &systemv1.CheckPTPSyncResponse{
			Success: false,
			Message: "PTP check API is disabled in configuration",
		}, nil
	}
	
	// Get local PTP status
	localStatus, err := s.ptpChecker.GetLocalTimeStatus()
	if err != nil {
		return &systemv1.CheckPTPSyncResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get local PTP status: %v", err),
		}, nil
	}
	
	response := &systemv1.CheckPTPSyncResponse{
		Success: true,
		Message: "PTP sync status retrieved successfully",
		LocalStatus: &systemv1.PTPStatus{
			ClockId:       localStatus.ClockID,
			MasterOffsetNs: localStatus.MasterOffset,
			IngressTime:   localStatus.IngressTime,
			GmPresent:     localStatus.GmPresent,
			GmIdentity:    localStatus.GmIdentity,
			IsSynced:      localStatus.IsSynced,
		},
		RemoteStatuses: []*systemv1.RemotePTPStatus{},
	}
	
	// Check remote devices if requested
	if req.IncludeRemoteDevices {
		for _, device := range s.config.PTP.RemoteDevices {
			log.Printf("Checking PTP status for %s (%s)", device.Name, device.IPAddress)
			
			remoteStatus := &systemv1.RemotePTPStatus{
				DeviceName: device.Name,
				IpAddress:  device.IPAddress,
			}
			
			timeStatus, err := s.ptpChecker.GetRemoteTimeStatus(device.IPAddress)
			if err != nil {
				remoteStatus.IsReachable = false
				remoteStatus.ErrorMessage = err.Error()
			} else {
				remoteStatus.IsReachable = true
				remoteStatus.Status = &systemv1.PTPStatus{
					ClockId:       timeStatus.ClockID,
					MasterOffsetNs: timeStatus.MasterOffset,
					IngressTime:   timeStatus.IngressTime,
					GmPresent:     timeStatus.GmPresent,
					GmIdentity:    timeStatus.GmIdentity,
					IsSynced:      timeStatus.IsSynced,
				}
			}
			
			response.RemoteStatuses = append(response.RemoteStatuses, remoteStatus)
		}
	}
	
	return response, nil
}
func (s *SystemService) GetService(ctx context.Context, req *systemv1.GetServiceRequest) (*systemv1.Service, error) {
	log.Printf("Get service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Parse resource name
	resourceType, resourceID, err := config.ParseResourceName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid resource name: %v", err)
	}
	
	if resourceType != "services" {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	
	// Get service mapping
	serviceMapping, err := s.config.GetServiceMapping(resourceID)
	if err != nil {
		return nil, fmt.Errorf("service not found: %v", err)
	}
	
	// Get systemd service info
	serviceInfo, err := s.systemManager.GetServiceInfo(serviceMapping.SystemdName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service info: %v", err)
	}
	
	// Convert status to proto enum
	var state systemv1.Service_ServiceState
	switch serviceInfo.Status {
	case "active":
		state = systemv1.Service_SERVICE_STATE_ACTIVE
	case "inactive":
		state = systemv1.Service_SERVICE_STATE_INACTIVE
	case "failed":
		state = systemv1.Service_SERVICE_STATE_FAILED
	case "activating":
		state = systemv1.Service_SERVICE_STATE_ACTIVATING
	case "deactivating":
		state = systemv1.Service_SERVICE_STATE_DEACTIVATING
	default:
		state = systemv1.Service_SERVICE_STATE_UNSPECIFIED
	}
	
	return &systemv1.Service{
		Name:           req.Name,
		State:          state,
		Enabled:        serviceInfo.Enabled,
		Description:    serviceMapping.Description,
		UptimeSeconds:  serviceInfo.UptimeSeconds,
	}, nil
}

func (s *SystemService) ListServices(ctx context.Context, req *systemv1.ListServicesRequest) (*systemv1.ListServicesResponse, error) {
	log.Printf("List services request")
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Get all enabled services from config
	enabledServices := s.config.GetAllEnabledServices()
	
	var services []*systemv1.Service
	for _, resourceID := range enabledServices {
		serviceMapping, err := s.config.GetServiceMapping(resourceID)
		if err != nil {
			log.Printf("Failed to get service mapping for %s: %v", resourceID, err)
			continue
		}
		
		serviceInfo, err := s.systemManager.GetServiceInfo(serviceMapping.SystemdName)
		if err != nil {
			log.Printf("Failed to get service info for %s: %v", serviceMapping.SystemdName, err)
			continue
		}
		
		// Convert status to proto enum
		var state systemv1.Service_ServiceState
		switch serviceInfo.Status {
		case "active":
			state = systemv1.Service_SERVICE_STATE_ACTIVE
		case "inactive":
			state = systemv1.Service_SERVICE_STATE_INACTIVE
		case "failed":
			state = systemv1.Service_SERVICE_STATE_FAILED
		case "activating":
			state = systemv1.Service_SERVICE_STATE_ACTIVATING
		case "deactivating":
			state = systemv1.Service_SERVICE_STATE_DEACTIVATING
		default:
			state = systemv1.Service_SERVICE_STATE_UNSPECIFIED
		}
		
		services = append(services, &systemv1.Service{
			Name:           fmt.Sprintf("services/%s", resourceID),
			State:          state,
			Enabled:        serviceInfo.Enabled,
			Description:    serviceMapping.Description,
			UptimeSeconds:  serviceInfo.UptimeSeconds,
		})
	}
	
	return &systemv1.ListServicesResponse{
		Services: services,
	}, nil
}

func (s *SystemService) StartService(ctx context.Context, req *systemv1.StartServiceRequest) (*systemv1.StartServiceResponse, error) {
	log.Printf("Start service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get systemd service name: %v", err)
	}
	
	// Start the service
	_, err = s.systemManager.ManageService(systemdName, "start")
	if err != nil {
		return nil, fmt.Errorf("failed to start service: %v", err)
	}
	
	// Get updated service info
	service, err := s.GetService(ctx, &systemv1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after start: %v", err)
	}
	
	return &systemv1.StartServiceResponse{
		Service: service,
	}, nil
}

func (s *SystemService) StopService(ctx context.Context, req *systemv1.StopServiceRequest) (*systemv1.StopServiceResponse, error) {
	log.Printf("Stop service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get systemd service name: %v", err)
	}
	
	// Stop the service
	_, err = s.systemManager.ManageService(systemdName, "stop")
	if err != nil {
		return nil, fmt.Errorf("failed to stop service: %v", err)
	}
	
	// Get updated service info
	service, err := s.GetService(ctx, &systemv1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after stop: %v", err)
	}
	
	return &systemv1.StopServiceResponse{
		Service: service,
	}, nil
}

func (s *SystemService) RestartService(ctx context.Context, req *systemv1.RestartServiceRequest) (*systemv1.RestartServiceResponse, error) {
	log.Printf("Restart service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get systemd service name: %v", err)
	}
	
	// Restart the service
	_, err = s.systemManager.ManageService(systemdName, "restart")
	if err != nil {
		return nil, fmt.Errorf("failed to restart service: %v", err)
	}
	
	// Get updated service info
	service, err := s.GetService(ctx, &systemv1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after restart: %v", err)
	}
	
	return &systemv1.RestartServiceResponse{
		Service: service,
	}, nil
}

func (s *SystemService) EnableService(ctx context.Context, req *systemv1.EnableServiceRequest) (*systemv1.EnableServiceResponse, error) {
	log.Printf("Enable service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get systemd service name: %v", err)
	}
	
	// Enable the service
	_, err = s.systemManager.ManageService(systemdName, "enable")
	if err != nil {
		return nil, fmt.Errorf("failed to enable service: %v", err)
	}
	
	// Get updated service info
	service, err := s.GetService(ctx, &systemv1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after enable: %v", err)
	}
	
	return &systemv1.EnableServiceResponse{
		Service: service,
	}, nil
}

func (s *SystemService) DisableService(ctx context.Context, req *systemv1.DisableServiceRequest) (*systemv1.DisableServiceResponse, error) {
	log.Printf("Disable service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get systemd service name: %v", err)
	}
	
	// Disable the service
	_, err = s.systemManager.ManageService(systemdName, "disable")
	if err != nil {
		return nil, fmt.Errorf("failed to disable service: %v", err)
	}
	
	// Get updated service info
	service, err := s.GetService(ctx, &systemv1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after disable: %v", err)
	}
	
	return &systemv1.DisableServiceResponse{
		Service: service,
	}, nil
}
