package service

import (
	"context"
	"fmt"
	"log"

	"github.com/proto_api/services/module-agent/internal/config"
	systemv1 "github.com/proto_api/services/module-agent/gen/system/v1"
	"github.com/proto_api/services/module-agent/internal/ptp"
	"github.com/proto_api/services/module-agent/internal/storage"
	"github.com/proto_api/services/module-agent/internal/system"
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
	
	// Check if reboot API is enabled
	if !s.config.APIs.EnableReboot {
		log.Printf("Reboot API is disabled")
		return &systemv1.RebootResponse{
			Success: false,
			Message: "Reboot API is disabled in configuration",
		}, nil
	}
	
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
	
	// Check if shutdown API is enabled
	if !s.config.APIs.EnableShutdown {
		log.Printf("Shutdown API is disabled")
		return &systemv1.ShutdownResponse{
			Success: false,
			Message: "Shutdown API is disabled in configuration",
		}, nil
	}
	
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
	// Check if disk usage API is enabled
	if !s.config.APIs.EnableDiskUsage {
		log.Printf("Disk usage API is disabled")
		return &systemv1.GetDiskUsageResponse{
			Success: false,
			Message: "Disk usage API is disabled in configuration",
		}, nil
	}
	
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
	
	// Check if PTP check API is enabled
	if !s.config.APIs.EnablePTPCheck {
		log.Printf("PTP check API is disabled")
		return &systemv1.CheckPTPSyncResponse{
			Success: false,
			Message: "PTP check API is disabled in configuration",
		}, nil
	}
	
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
func (s *SystemService) ManageDrsService(ctx context.Context, req *systemv1.ManageDrsServiceRequest) (*systemv1.ManageDrsServiceResponse, error) {
	log.Printf("DRS service management request: %v", req.Action)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return &systemv1.ManageDrsServiceResponse{
			Success: false,
			Message: "Service management API is disabled in configuration",
		}, nil
	}
	
	// Check if systemd management is enabled
	if !s.config.Services.EnableSystemdManage {
		return &systemv1.ManageDrsServiceResponse{
			Success: false,
			Message: "systemd service management is not enabled",
		}, nil
	}

	var err error
	var statusStr string
	
	switch req.Action {
	case systemv1.ManageDrsServiceRequest_SERVICE_ACTION_STOP:
		err = s.systemManager.StopDrsService()
		statusStr = "stopped"
	case systemv1.ManageDrsServiceRequest_SERVICE_ACTION_RESTART:
		err = s.systemManager.RestartDrsService()
		statusStr = "restarted"
	case systemv1.ManageDrsServiceRequest_SERVICE_ACTION_STATUS:
		statusStr, err = s.systemManager.GetDrsServiceStatus()
	default:
		return &systemv1.ManageDrsServiceResponse{
			Success: false,
			Message: "invalid action specified",
		}, nil
	}

	if err != nil {
		return &systemv1.ManageDrsServiceResponse{
			Success: false,
			Message: fmt.Sprintf("DRS service operation failed: %v", err),
		}, nil
	}

	return &systemv1.ManageDrsServiceResponse{
		Success:       true,
		Message:       fmt.Sprintf("DRS service operation completed successfully"),
		ServiceStatus: statusStr,
	}, nil
}

func (s *SystemService) ManageRecorderService(ctx context.Context, req *systemv1.ManageRecorderServiceRequest) (*systemv1.ManageRecorderServiceResponse, error) {
	log.Printf("Recorder service management request: %v", req.Action)
	
	// Check if service management API is enabled
	if !s.config.APIs.EnableServiceManagement {
		log.Printf("Service management API is disabled")
		return &systemv1.ManageRecorderServiceResponse{
			Success: false,
			Message: "Service management API is disabled in configuration",
		}, nil
	}
	
	// Check if systemd management is enabled
	if !s.config.Services.EnableSystemdManage {
		return &systemv1.ManageRecorderServiceResponse{
			Success: false,
			Message: "systemd service management is not enabled",
		}, nil
	}

	var err error
	var statusStr string
	
	switch req.Action {
	case systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_STOP:
		err = s.systemManager.StopRecorderService()
		statusStr = "stopped"
	case systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_RESTART:
		err = s.systemManager.RestartRecorderService()
		statusStr = "restarted"
	case systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_STATUS:
		statusStr, err = s.systemManager.GetRecorderServiceStatus()
	default:
		return &systemv1.ManageRecorderServiceResponse{
			Success: false,
			Message: "invalid action specified",
		}, nil
	}

	if err != nil {
		return &systemv1.ManageRecorderServiceResponse{
			Success: false,
			Message: fmt.Sprintf("Recorder service operation failed: %v", err),
		}, nil
	}

	return &systemv1.ManageRecorderServiceResponse{
		Success:       true,
		Message:       fmt.Sprintf("Recorder service operation completed successfully"),
		ServiceStatus: statusStr,
	}, nil
}
