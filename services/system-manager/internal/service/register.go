package service

import (
	"github.com/proto_api/services/system-manager/internal/config"
	systemv1 "github.com/proto_api/services/system-manager/gen/system/v1"
	"google.golang.org/grpc"
)

func RegisterServices(s *grpc.Server, cfg *config.Config) {
	systemService := NewSystemService(cfg)
	// Register the generated service
	systemv1.RegisterSystemServiceServer(s, systemService)
}