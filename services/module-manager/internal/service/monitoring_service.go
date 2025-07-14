package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/drs-api/services/module-manager/internal/config"
	modulev1 "github.com/drs-api/services/module-manager/drs/module/v1"
	"github.com/drs-api/services/module-manager/internal/ptp"
	"github.com/drs-api/services/module-manager/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MonitoringService struct {
	modulev1.UnimplementedMonitoringServiceServer
	storageManager *storage.Manager
	ptpChecker     *ptp.Checker
	config         *config.Config
}

func NewMonitoringService(cfg *config.Config) *MonitoringService {
	return &MonitoringService{
		storageManager: storage.NewManager(),
		ptpChecker:     ptp.NewChecker(),
		config:         cfg,
	}
}

func (s *MonitoringService) GetDiskUsage(ctx context.Context, req *modulev1.GetDiskUsageRequest) (*modulev1.GetDiskUsageResponse, error) {
	log.Printf("Disk usage request received")
	
	// Check if disk usage API is enabled
	if !s.config.Disk.Enabled {
		log.Printf("Disk usage API is disabled")
		return nil, status.Error(codes.Unimplemented, "disk usage monitoring is disabled")
	}
	
	// Always use the configured monitor path - one disk per ECU
	targetPath := s.config.GetDiskPath()
	log.Printf("Getting disk usage for primary disk: %s", targetPath)
	
	usage, err := s.storageManager.GetDiskUsage(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get disk usage: %v", err))
	}

	return &modulev1.GetDiskUsageResponse{
		DiskUsage: &modulev1.DiskUsage{
			TotalBytes:      usage.TotalBytes,
			UsedBytes:       usage.UsedBytes,
			FreeBytes:       usage.FreeBytes,
			UsagePercentage: usage.UsagePercentage,
		},
	}, nil
}

func (s *MonitoringService) GetPTPStatus(ctx context.Context, req *modulev1.GetPTPStatusRequest) (*modulev1.GetPTPStatusResponse, error) {
	log.Printf("PTP sync check request received (include_remote: %v)", req.IncludeRemoteDevices)
	
	// Check if PTP check API is enabled
	if !s.config.PTP.Enabled {
		log.Printf("PTP check API is disabled")
		return nil, status.Error(codes.Unimplemented, "PTP monitoring is disabled")
	}
	
	// Get local PTP status
	localStatus, err := s.ptpChecker.GetLocalTimeStatus()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get local PTP status: %v", err))
	}
	
	response := &modulev1.GetPTPStatusResponse{
		LocalStatus: &modulev1.PTPStatus{
			ClockId:       localStatus.ClockID,
			MasterOffsetNs: localStatus.MasterOffset,
			IngressTime:   localStatus.IngressTime,
			GmPresent:     localStatus.GmPresent,
			GmIdentity:    localStatus.GmIdentity,
			IsSynced:      localStatus.IsSynced,
		},
		RemoteStatuses: []*modulev1.RemotePTPStatus{},
	}
	
	// Check remote devices if requested
	if req.IncludeRemoteDevices {
		for _, device := range s.config.PTP.RemoteDevices {
			log.Printf("Checking PTP status for %s (%s)", device.Name, device.IPAddress)
			
			remoteStatus := &modulev1.RemotePTPStatus{
				DeviceName: device.Name,
				IpAddress:  device.IPAddress,
			}
			
			timeStatus, err := s.ptpChecker.GetRemoteTimeStatus(device.IPAddress)
			if err != nil {
				remoteStatus.IsReachable = false
				remoteStatus.ErrorMessage = err.Error()
			} else {
				remoteStatus.IsReachable = true
				remoteStatus.Status = &modulev1.PTPStatus{
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

func (s *MonitoringService) GetEnvironment(ctx context.Context, req *modulev1.GetEnvironmentRequest) (*modulev1.GetEnvironmentResponse, error) {
	log.Printf("Get environment request received")
	
	// Get environment variables
	sensingSystemID := os.Getenv("SENSING_SYSTEM_ID")
	moduleID := os.Getenv("MODULE_ID")
	
	log.Printf("Environment variables - SENSING_SYSTEM_ID: %s, MODULE_ID: %s", sensingSystemID, moduleID)
	
	return &modulev1.GetEnvironmentResponse{
		SensingSystemId: sensingSystemID,
		ModuleId:        moduleID,
	}, nil
}