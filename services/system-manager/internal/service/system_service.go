package service

import (
	"context"
	"fmt"
	"log"

	"github.com/proto_api/services/system-manager/internal/config"
	systemv1 "github.com/proto_api/services/system-manager/gen/system/v1"
	"github.com/proto_api/services/system-manager/internal/storage"
	"github.com/proto_api/services/system-manager/internal/system"
)

type SystemService struct {
	systemv1.UnimplementedSystemServiceServer
	systemManager  *system.Manager
	storageManager *storage.Manager
	config         *config.Config
}

func NewSystemService(cfg *config.Config) *SystemService {
	return &SystemService{
		systemManager:  system.NewManager(),
		storageManager: storage.NewManager(),
		config:         cfg,
	}
}

func (s *SystemService) Reboot(ctx context.Context, req *systemv1.RebootRequest) (*systemv1.RebootResponse, error) {
	log.Printf("Reboot request received with delay: %d seconds", req.DelaySeconds)
	
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
	// Determine target path
	var targetPath string
	if req.Path == "" {
		// Use default path from config
		targetPath = s.config.Disk.DefaultPath
		log.Printf("Disk usage request for default path: %s", targetPath)
	} else {
		// Check if the requested path is a configured name
		targetPath = s.config.GetDiskPath(req.Path)
		if targetPath == s.config.Disk.DefaultPath && req.Path != s.config.Disk.DefaultPath {
			// It was a name lookup that fell back to default
			log.Printf("Disk usage request for named path '%s' -> %s", req.Path, targetPath)
		} else {
			// Direct path or exact match
			log.Printf("Disk usage request for path: %s", targetPath)
		}
	}
	
	usage, err := s.storageManager.GetDiskUsage(targetPath)
	if err != nil {
		return &systemv1.GetDiskUsageResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get disk usage for %s: %v", targetPath, err),
		}, nil
	}

	return &systemv1.GetDiskUsageResponse{
		Success: true,
		Message: fmt.Sprintf("Disk usage retrieved successfully for %s", targetPath),
		DiskUsage: &systemv1.DiskUsage{
			TotalBytes:      usage.TotalBytes,
			UsedBytes:       usage.UsedBytes,
			FreeBytes:       usage.FreeBytes,
			UsagePercentage: usage.UsagePercentage,
			Filesystem:      targetPath,
		},
	}, nil
}