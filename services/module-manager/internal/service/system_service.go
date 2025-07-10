package service

import (
	"context"
	"fmt"
	"log"

	"github.com/drs-api/services/module-manager/internal/config"
	modulev1 "github.com/drs-api/services/module-manager/gen/drs/module/v1"
	"github.com/drs-api/services/module-manager/internal/ptp"
	"github.com/drs-api/services/module-manager/internal/storage"
	"github.com/drs-api/services/module-manager/internal/system"
)

type ModuleService struct {
	modulev1.UnimplementedModuleServiceServer
	systemManager  *system.Manager
	storageManager *storage.Manager
	ptpChecker     *ptp.Checker
	config         *config.Config
}

func NewModuleService(cfg *config.Config) *ModuleService {
	return &ModuleService{
		systemManager:  system.NewManager(),
		storageManager: storage.NewManager(),
		ptpChecker:     ptp.NewChecker(),
		config:         cfg,
	}
}

func (s *ModuleService) Reboot(ctx context.Context, req *modulev1.RebootRequest) (*modulev1.RebootResponse, error) {
	log.Printf("Reboot request received with delay: %d seconds", req.DelaySeconds)
	
	// Check if reboot API is enabled
	if !s.config.System.EnableReboot {
		log.Printf("Reboot API is disabled")
		return nil, fmt.Errorf("reboot API is disabled in configuration")
	}
	
	err := s.systemManager.Reboot(int(req.DelaySeconds))
	if err != nil {
		log.Printf("Reboot failed: %v", err)
		return nil, fmt.Errorf("reboot failed: %v", err)
	}

	return &modulev1.RebootResponse{
		Accepted: true,
		Message: fmt.Sprintf("Reboot scheduled with %d seconds delay", req.DelaySeconds),
		ScheduledDelay: req.DelaySeconds,
	}, nil
}

func (s *ModuleService) Shutdown(ctx context.Context, req *modulev1.ShutdownRequest) (*modulev1.ShutdownResponse, error) {
	log.Printf("Shutdown request received with delay: %d seconds", req.DelaySeconds)
	
	// Check if shutdown API is enabled
	if !s.config.System.EnableShutdown {
		log.Printf("Shutdown API is disabled")
		return nil, fmt.Errorf("shutdown API is disabled in configuration")
	}
	
	err := s.systemManager.Shutdown(int(req.DelaySeconds))
	if err != nil {
		log.Printf("Shutdown failed: %v", err)
		return nil, fmt.Errorf("shutdown failed: %v", err)
	}

	return &modulev1.ShutdownResponse{
		Accepted: true,
		Message: fmt.Sprintf("Shutdown scheduled with %d seconds delay", req.DelaySeconds),
		ScheduledDelay: req.DelaySeconds,
	}, nil
}


func (s *ModuleService) GetDiskUsage(ctx context.Context, req *modulev1.GetDiskUsageRequest) (*modulev1.GetDiskUsageResponse, error) {
	log.Printf("Disk usage request received")
	
	// Check if disk usage API is enabled
	if !s.config.Disk.Enabled {
		log.Printf("Disk usage API is disabled")
		return nil, fmt.Errorf("disk usage API is disabled in configuration")
	}
	
	// Always use the configured monitor path - one disk per ECU
	targetPath := s.config.GetDiskPath()
	log.Printf("Getting disk usage for primary disk: %s", targetPath)
	
	usage, err := s.storageManager.GetDiskUsage(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %v", err)
	}

	return &modulev1.GetDiskUsageResponse{
		DiskUsage: &modulev1.DiskUsage{
			TotalBytes:      usage.TotalBytes,
			UsedBytes:       usage.UsedBytes,
			FreeBytes:       usage.FreeBytes,
			UsagePercentage: usage.UsagePercentage,
		},
	}, nil
}

func (s *ModuleService) GetPTPStatus(ctx context.Context, req *modulev1.GetPTPStatusRequest) (*modulev1.GetPTPStatusResponse, error) {
	log.Printf("PTP sync check request received (include_remote: %v)", req.IncludeRemoteDevices)
	
	// Check if PTP check API is enabled
	if !s.config.PTP.Enabled {
		log.Printf("PTP check API is disabled")
		return nil, fmt.Errorf("PTP check API is disabled in configuration")
	}
	
	// Get local PTP status
	localStatus, err := s.ptpChecker.GetLocalTimeStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get local PTP status: %v", err)
	}
	
	response := &modulev1.GetPTPStatusResponse{
		LocalStatus: &modulev1.PTPStatus{
			ClockId:       localStatus.ClockID,
			MasterOffsetNs: localStatus.MasterOffset,
			IngressTime:   localStatus.IngressTime,
			GmPresent:     localStatus.GmPresent,
			GmIdentity:    localStatus.GmIdentity,
			IsSynced:      localStatus.IsSynced,
		},
		RemoteStatuses: []*modulev1.RemotePTPStatus{},
	}
	
	// Check remote devices if requested
	if req.IncludeRemoteDevices {
		for _, device := range s.config.PTP.RemoteDevices {
			log.Printf("Checking PTP status for %s (%s)", device.Name, device.IPAddress)
			
			remoteStatus := &modulev1.RemotePTPStatus{
				DeviceName: device.Name,
				IpAddress:  device.IPAddress,
			}
			
			timeStatus, err := s.ptpChecker.GetRemoteTimeStatus(device.IPAddress)
			if err != nil {
				remoteStatus.IsReachable = false
				remoteStatus.ErrorMessage = err.Error()
			} else {
				remoteStatus.IsReachable = true
				remoteStatus.Status = &modulev1.PTPStatus{
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
func (s *ModuleService) GetService(ctx context.Context, req *modulev1.GetServiceRequest) (*modulev1.Service, error) {
	log.Printf("Get service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.Services.Enabled {
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
	var state modulev1.Service_ServiceState
	switch serviceInfo.Status {
	case "active":
		state = modulev1.Service_SERVICE_STATE_ACTIVE
	case "inactive":
		state = modulev1.Service_SERVICE_STATE_INACTIVE
	case "failed":
		state = modulev1.Service_SERVICE_STATE_FAILED
	case "activating":
		state = modulev1.Service_SERVICE_STATE_ACTIVATING
	case "deactivating":
		state = modulev1.Service_SERVICE_STATE_DEACTIVATING
	default:
		state = modulev1.Service_SERVICE_STATE_UNSPECIFIED
	}
	
	return &modulev1.Service{
		Name:           req.Name,
		State:          state,
		Enabled:        serviceInfo.Enabled,
		Description:    serviceMapping.Description,
		UptimeSeconds:  serviceInfo.UptimeSeconds,
	}, nil
}

func (s *ModuleService) ListServices(ctx context.Context, req *modulev1.ListServicesRequest) (*modulev1.ListServicesResponse, error) {
	log.Printf("List services request")
	
	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, fmt.Errorf("service management API is disabled in configuration")
	}
	
	// Get all enabled services from config
	enabledServices := s.config.GetAllEnabledServices()
	
	var services []*modulev1.Service
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
		var state modulev1.Service_ServiceState
		switch serviceInfo.Status {
		case "active":
			state = modulev1.Service_SERVICE_STATE_ACTIVE
		case "inactive":
			state = modulev1.Service_SERVICE_STATE_INACTIVE
		case "failed":
			state = modulev1.Service_SERVICE_STATE_FAILED
		case "activating":
			state = modulev1.Service_SERVICE_STATE_ACTIVATING
		case "deactivating":
			state = modulev1.Service_SERVICE_STATE_DEACTIVATING
		default:
			state = modulev1.Service_SERVICE_STATE_UNSPECIFIED
		}
		
		services = append(services, &modulev1.Service{
			Name:           fmt.Sprintf("services/%s", resourceID),
			State:          state,
			Enabled:        serviceInfo.Enabled,
			Description:    serviceMapping.Description,
			UptimeSeconds:  serviceInfo.UptimeSeconds,
		})
	}
	
	return &modulev1.ListServicesResponse{
		Services: services,
	}, nil
}

func (s *ModuleService) StartService(ctx context.Context, req *modulev1.StartServiceRequest) (*modulev1.StartServiceResponse, error) {
	log.Printf("Start service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.Services.Enabled {
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
	service, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after start: %v", err)
	}
	
	return &modulev1.StartServiceResponse{
		Service: service,
	}, nil
}

func (s *ModuleService) StopService(ctx context.Context, req *modulev1.StopServiceRequest) (*modulev1.StopServiceResponse, error) {
	log.Printf("Stop service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.Services.Enabled {
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
	service, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after stop: %v", err)
	}
	
	return &modulev1.StopServiceResponse{
		Service: service,
	}, nil
}

func (s *ModuleService) RestartService(ctx context.Context, req *modulev1.RestartServiceRequest) (*modulev1.RestartServiceResponse, error) {
	log.Printf("Restart service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.Services.Enabled {
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
	service, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after restart: %v", err)
	}
	
	return &modulev1.RestartServiceResponse{
		Service: service,
	}, nil
}

func (s *ModuleService) EnableService(ctx context.Context, req *modulev1.EnableServiceRequest) (*modulev1.EnableServiceResponse, error) {
	log.Printf("Enable service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.Services.Enabled {
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
	service, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after enable: %v", err)
	}
	
	return &modulev1.EnableServiceResponse{
		Service: service,
	}, nil
}

func (s *ModuleService) DisableService(ctx context.Context, req *modulev1.DisableServiceRequest) (*modulev1.DisableServiceResponse, error) {
	log.Printf("Disable service request: %s", req.Name)
	
	// Check if service management API is enabled
	if !s.config.Services.Enabled {
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
	service, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get service info after disable: %v", err)
	}
	
	return &modulev1.DisableServiceResponse{
		Service: service,
	}, nil
}
