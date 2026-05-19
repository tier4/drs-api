package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	modulev1 "github.com/tier4/drs-api/services/module-manager/gen/drs/module/v1"
	"github.com/tier4/drs-api/services/module-manager/internal/config"
	"github.com/tier4/drs-api/services/module-manager/internal/ptp"
	"github.com/tier4/drs-api/services/module-manager/internal/storage"
)

type MonitoringService struct {
	modulev1.UnimplementedMonitoringServiceServer
	storageManager *storage.Manager
	ptpChecker     *ptp.Checker
	config         *config.Config
}

func NewMonitoringService(cfg *config.Config) *MonitoringService {
	return &MonitoringService{
		storageManager: storage.NewManager(cfg.Disk),
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

	// Always use the primary disk for legacy API
	// The config validation ensures there is at least one disk if enabled
	primaryDisk := s.config.Disk.Disks[0]
	log.Printf("Getting disk usage for primary disk: %s (legacy API)", primaryDisk.Name)

	usage, err := s.storageManager.GetDiskUsage(ctx, primaryDisk.Name)
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

func (s *MonitoringService) GetDisk(ctx context.Context, req *modulev1.GetDiskRequest) (*modulev1.GetDiskResponse, error) {
	log.Printf("Get disk request received for: %s", req.Name)

	// Check if disk usage API is enabled
	if !s.config.Disk.Enabled {
		return nil, status.Error(codes.Unimplemented, "disk usage monitoring is disabled")
	}

	// Extract disk name from resource name (support both "disks/name" and "name" formats)
	diskName := req.Name
	if len(diskName) > 6 && diskName[:6] == "disks/" {
		diskName = diskName[6:]
	}

	log.Printf("Getting disk usage for disk: %s", diskName)

	usage, err := s.storageManager.GetDiskUsage(ctx, diskName)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get disk usage: %v", err))
	}

	// Find disk entry for description
	var description string
	for _, disk := range s.storageManager.GetDisks() {
		if disk.Name == diskName {
			description = disk.Description
			break
		}
	}

	return &modulev1.GetDiskResponse{
		Disk: &modulev1.Disk{
			Name:      fmt.Sprintf("disks/%s", diskName),
			MountPath: "", // Will be added if needed
			Usage: &modulev1.DiskUsage{
				TotalBytes:      usage.TotalBytes,
				UsedBytes:       usage.UsedBytes,
				FreeBytes:       usage.FreeBytes,
				UsagePercentage: usage.UsagePercentage,
			},
			Description: description,
		},
	}, nil
}

func (s *MonitoringService) ListDisks(ctx context.Context, req *modulev1.ListDisksRequest) (*modulev1.ListDisksResponse, error) {
	log.Printf("List disks request received")

	// Check if disk usage API is enabled
	if !s.config.Disk.Enabled {
		log.Printf("Disk usage API is disabled")
		return nil, status.Error(codes.Unimplemented, "disk usage monitoring is disabled")
	}

	allUsages, err := s.storageManager.GetAllDiskUsages(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get disk usages: %v", err))
	}

	var disks []*modulev1.Disk
	for _, diskEntry := range s.storageManager.GetDisks() {
		usage, ok := allUsages[diskEntry.Name]
		if !ok {
			// Skip if we couldn't get usage for this disk
			continue
		}

		disk := &modulev1.Disk{
			Name:      fmt.Sprintf("disks/%s", diskEntry.Name),
			MountPath: diskEntry.MountPath,
			Usage: &modulev1.DiskUsage{
				TotalBytes:      usage.TotalBytes,
				UsedBytes:       usage.UsedBytes,
				FreeBytes:       usage.FreeBytes,
				UsagePercentage: usage.UsagePercentage,
			},
			Description: diskEntry.Description,
		}
		disks = append(disks, disk)
	}

	return &modulev1.ListDisksResponse{
		Disks: disks,
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
			ClockId:        localStatus.ClockID,
			MasterOffsetNs: localStatus.MasterOffset,
			IngressTime:    localStatus.IngressTime,
			GmPresent:      localStatus.GmPresent,
			GmIdentity:     localStatus.GmIdentity,
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
					ClockId:        timeStatus.ClockID,
					MasterOffsetNs: timeStatus.MasterOffset,
					IngressTime:    timeStatus.IngressTime,
					GmPresent:      timeStatus.GmPresent,
					GmIdentity:     timeStatus.GmIdentity,
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
