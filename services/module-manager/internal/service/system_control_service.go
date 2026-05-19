package service

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	modulev1 "github.com/tier4/drs-api/services/module-manager/gen/drs/module/v1"
	"github.com/tier4/drs-api/services/module-manager/internal/config"
	"github.com/tier4/drs-api/services/module-manager/internal/system"
)

type SystemControlService struct {
	modulev1.UnimplementedSystemControlServiceServer
	systemManager *system.Manager
	config        *config.Config
}

func NewSystemControlService(cfg *config.Config) *SystemControlService {
	return &SystemControlService{
		systemManager: system.NewManager(),
		config:        cfg,
	}
}

func (s *SystemControlService) Reboot(ctx context.Context, req *modulev1.RebootRequest) (*modulev1.RebootResponse, error) {
	log.Printf("Reboot request received with delay: %d seconds", req.DelaySeconds)

	// Check if reboot API is enabled
	if !s.config.System.EnableReboot {
		log.Printf("Reboot API is disabled")
		return nil, status.Error(codes.Unimplemented, "reboot API is disabled in configuration")
	}

	err := s.systemManager.Reboot(int(req.DelaySeconds))
	if err != nil {
		log.Printf("Reboot failed: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("reboot failed: %v", err))
	}

	return &modulev1.RebootResponse{
		Accepted:       true,
		Message:        fmt.Sprintf("Reboot scheduled with %d seconds delay", req.DelaySeconds),
		ScheduledDelay: req.DelaySeconds,
	}, nil
}

func (s *SystemControlService) Shutdown(ctx context.Context, req *modulev1.ShutdownRequest) (*modulev1.ShutdownResponse, error) {
	log.Printf("Shutdown request received with delay: %d seconds", req.DelaySeconds)

	// Check if shutdown API is enabled
	if !s.config.System.EnableShutdown {
		log.Printf("Shutdown API is disabled")
		return nil, status.Error(codes.Unimplemented, "shutdown API is disabled in configuration")
	}

	err := s.systemManager.Shutdown(int(req.DelaySeconds))
	if err != nil {
		log.Printf("Shutdown failed: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("shutdown failed: %v", err))
	}

	return &modulev1.ShutdownResponse{
		Accepted:       true,
		Message:        fmt.Sprintf("Shutdown scheduled with %d seconds delay", req.DelaySeconds),
		ScheduledDelay: req.DelaySeconds,
	}, nil
}
