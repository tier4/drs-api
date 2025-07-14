package service

import (
	"github.com/tier4/drs-api/services/module-manager/internal/config"
	modulev1 "github.com/tier4/drs-api/services/module-manager/drs/module/v1"
	"google.golang.org/grpc"
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