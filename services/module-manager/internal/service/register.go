package service

import (
	"github.com/drs-api/services/module-manager/internal/config"
	modulev1 "github.com/drs-api/services/module-manager/gen/drs/module/v1"
	"google.golang.org/grpc"
)

func RegisterServices(s *grpc.Server, cfg *config.Config) {
	moduleService := NewModuleService(cfg)
	// Register the generated service
	modulev1.RegisterModuleServiceServer(s, moduleService)
}