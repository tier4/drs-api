package service

import (
	"github.com/drs-api/services/module-manager/internal/config"
	systemv1 "github.com/drs-api/services/module-manager/gen/system/v1"
	"google.golang.org/grpc"
)

func RegisterServices(s *grpc.Server, cfg *config.Config) {
	systemService := NewSystemService(cfg)
	// Register the generated service
	systemv1.RegisterSystemServiceServer(s, systemService)
}