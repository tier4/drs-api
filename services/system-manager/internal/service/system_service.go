package service

import (
	"context"
	"fmt"
	"log"

	"github.com/proto_api/services/system-manager/internal/config"
	systemv1 "github.com/proto_api/services/system-manager/gen/system/v1"
	"github.com/proto_api/services/system-manager/internal/ptp"
	"github.com/proto_api/services/system-manager/internal/storage"
	"github.com/proto_api/services/system-manager/internal/system"
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

func (s *SystemService) CheckPTPSync(ctx context.Context, req *systemv1.CheckPTPSyncRequest) (*systemv1.CheckPTPSyncResponse, error) {
	log.Printf("PTP sync check request received (include_remote: %v)", req.IncludeRemoteDevices)
	
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