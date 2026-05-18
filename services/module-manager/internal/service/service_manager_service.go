package service

import (
	"context"
	"fmt"
	"log"

	"github.com/tier4/drs-api/services/module-manager/internal/config"
	modulev1 "github.com/tier4/drs-api/services/module-manager/gen/drs/module/v1"
	"github.com/tier4/drs-api/services/module-manager/internal/system"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServiceManagerService struct {
	modulev1.UnimplementedServiceManagerServiceServer
	systemManager *system.Manager
	config        *config.Config
}

func NewServiceManagerService(cfg *config.Config) *ServiceManagerService {
	return &ServiceManagerService{
		systemManager: system.NewManager(),
		config:        cfg,
	}
}

func (s *ServiceManagerService) GetService(ctx context.Context, req *modulev1.GetServiceRequest) (*modulev1.GetServiceResponse, error) {
	log.Printf("Get service request: %s", req.Name)

	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, status.Error(codes.Unimplemented, "service management API is disabled in configuration")
	}

	// Parse resource name
	resourceType, resourceID, err := config.ParseResourceName(req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid resource name: %v", err))
	}

	if resourceType != "services" {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unsupported resource type: %s", resourceType))
	}

	// Get service mapping
	serviceMapping, err := s.config.GetServiceMapping(resourceID)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("service not found: %v", err))
	}

	// Get systemd service info
	serviceInfo, err := s.systemManager.GetServiceInfo(serviceMapping.SystemdName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get service info: %v", err))
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

	return &modulev1.GetServiceResponse{
		Service: &modulev1.Service{
			Name:           req.Name,
			State:          state,
			Enabled:        serviceInfo.Enabled,
			Description:    serviceMapping.Description,
			UptimeSeconds:  serviceInfo.UptimeSeconds,
		},
	}, nil
}

func (s *ServiceManagerService) ListServices(ctx context.Context, req *modulev1.ListServicesRequest) (*modulev1.ListServicesResponse, error) {
	log.Printf("List services request")

	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, status.Error(codes.Unimplemented, "service management API is disabled in configuration")
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

func (s *ServiceManagerService) StartService(ctx context.Context, req *modulev1.StartServiceRequest) (*modulev1.StartServiceResponse, error) {
	log.Printf("Start service request: %s", req.Name)

	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, status.Error(codes.Unimplemented, "service management API is disabled in configuration")
	}

	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get systemd service name: %v", err))
	}

	// Start the service
	_, err = s.systemManager.ManageService(systemdName, "start")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to start service: %v", err))
	}

	// Get updated service info
	resp, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get service info after start: %v", err))
	}

	return &modulev1.StartServiceResponse{
		Service: resp.Service,
	}, nil
}

func (s *ServiceManagerService) StopService(ctx context.Context, req *modulev1.StopServiceRequest) (*modulev1.StopServiceResponse, error) {
	log.Printf("Stop service request: %s", req.Name)

	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, status.Error(codes.Unimplemented, "service management API is disabled in configuration")
	}

	// Get service mapping (needed for StopAlso)
	serviceMapping, err := s.config.GetServiceMappingByResourceName(req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get service mapping: %v", err))
	}

	// Stop secondary unit first if configured (ensures clean abort before disabling primary)
	if serviceMapping.StopAlso != "" {
		if !s.config.IsServiceAllowed(serviceMapping.StopAlso) {
			return nil, status.Error(codes.Internal, fmt.Sprintf("stop_also unit %q is not in the service allowlist", serviceMapping.StopAlso))
		}
		log.Printf("Stopping secondary unit first: %s", serviceMapping.StopAlso)
		if _, err2 := s.systemManager.ManageService(serviceMapping.StopAlso, "stop"); err2 != nil {
			log.Printf("Warning: failed to stop secondary unit %s: %v", serviceMapping.StopAlso, err2)
		}
	}

	// Stop the primary unit
	_, err = s.systemManager.ManageService(serviceMapping.SystemdName, "stop")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to stop service: %v", err))
	}

	// Get updated service info
	resp, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get service info after stop: %v", err))
	}

	return &modulev1.StopServiceResponse{
		Service: resp.Service,
	}, nil
}

func (s *ServiceManagerService) RestartService(ctx context.Context, req *modulev1.RestartServiceRequest) (*modulev1.RestartServiceResponse, error) {
	log.Printf("Restart service request: %s", req.Name)

	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, status.Error(codes.Unimplemented, "service management API is disabled in configuration")
	}

	// Get service mapping (needed for StopAlso on restart)
	serviceMapping, err := s.config.GetServiceMappingByResourceName(req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get service mapping: %v", err))
	}

	// Stop secondary unit first if configured (best-effort: ensures clean restart)
	if serviceMapping.StopAlso != "" {
		if !s.config.IsServiceAllowed(serviceMapping.StopAlso) {
			return nil, status.Error(codes.Internal, fmt.Sprintf("stop_also unit %q is not in the service allowlist", serviceMapping.StopAlso))
		}
		log.Printf("Stopping secondary unit before restart: %s", serviceMapping.StopAlso)
		if _, err2 := s.systemManager.ManageService(serviceMapping.StopAlso, "stop"); err2 != nil {
			log.Printf("Warning: failed to stop secondary unit %s before restart: %v", serviceMapping.StopAlso, err2)
		}
	}

	// Restart the primary unit
	_, err = s.systemManager.ManageService(serviceMapping.SystemdName, "restart")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to restart service: %v", err))
	}

	// Get updated service info
	resp, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get service info after restart: %v", err))
	}

	return &modulev1.RestartServiceResponse{
		Service: resp.Service,
	}, nil
}

func (s *ServiceManagerService) EnableService(ctx context.Context, req *modulev1.EnableServiceRequest) (*modulev1.EnableServiceResponse, error) {
	log.Printf("Enable service request: %s", req.Name)

	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, status.Error(codes.Unimplemented, "service management API is disabled in configuration")
	}

	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get systemd service name: %v", err))
	}

	// Enable the service
	_, err = s.systemManager.ManageService(systemdName, "enable")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to enable service: %v", err))
	}

	// Get updated service info
	resp, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get service info after enable: %v", err))
	}

	return &modulev1.EnableServiceResponse{
		Service: resp.Service,
	}, nil
}

func (s *ServiceManagerService) DisableService(ctx context.Context, req *modulev1.DisableServiceRequest) (*modulev1.DisableServiceResponse, error) {
	log.Printf("Disable service request: %s", req.Name)

	// Check if service management API is enabled
	if !s.config.Services.Enabled {
		log.Printf("Service management API is disabled")
		return nil, status.Error(codes.Unimplemented, "service management API is disabled in configuration")
	}

	// Get systemd service name
	systemdName, err := s.config.GetSystemdServiceName(req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get systemd service name: %v", err))
	}

	// Disable the service
	_, err = s.systemManager.ManageService(systemdName, "disable")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to disable service: %v", err))
	}

	// Get updated service info
	resp, err := s.GetService(ctx, &modulev1.GetServiceRequest{Name: req.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get service info after disable: %v", err))
	}

	return &modulev1.DisableServiceResponse{
		Service: resp.Service,
	}, nil
}
