package service

import (
	"google.golang.org/grpc"

	modulev1 "github.com/tier4/drs-api/services/module-manager/gen/drs/module/v1"
	"github.com/tier4/drs-api/services/module-manager/internal/config"
)

func RegisterServices(s *grpc.Server, cfg *config.Config) {
	// Register multiple services on a single gRPC server
	serviceManagerService := NewServiceManagerService(cfg)
	systemControlService := NewSystemControlService(cfg)
	monitoringService := NewMonitoringService(cfg)

	modulev1.RegisterServiceManagerServiceServer(s, serviceManagerService)
	modulev1.RegisterSystemControlServiceServer(s, systemControlService)
	modulev1.RegisterMonitoringServiceServer(s, monitoringService)
}
