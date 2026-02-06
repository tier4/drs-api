package storage

import (
	"fmt"
	"os"
	"syscall"
	"github.com/tier4/drs-api/services/module-manager/internal/config"
)

type Manager struct {
	disks []config.DiskEntry
}

func NewManager(diskConfig config.DiskConfig) *Manager {
	return &Manager{
		disks: diskConfig.Disks,
	}
}

type DiskUsage struct {
	TotalBytes      uint64
	UsedBytes       uint64
	FreeBytes       uint64
	UsagePercentage float64
}

// GetDiskUsage returns usage for a specific disk by name
func (m *Manager) GetDiskUsage(diskName string) (*DiskUsage, error) {
	var targetDisk *config.DiskEntry
	for _, disk := range m.disks {
		if disk.Name == diskName {
			targetDisk = &disk
			break
		}
	}
	
	if targetDisk == nil {
		return nil, fmt.Errorf("disk not found: %s", diskName)
	}
	
	return m.getDiskUsageForPath(targetDisk.MountPath)
}

// GetAllDiskUsages returns usage for all configured disks
func (m *Manager) GetAllDiskUsages() (map[string]*DiskUsage, error) {
	results := make(map[string]*DiskUsage)
	
	for _, disk := range m.disks {
		usage, err := m.getDiskUsageForPath(disk.MountPath)
		if err != nil {
			// Log error but continue with other disks
			fmt.Printf("Failed to get usage for disk %s: %v\n", disk.Name, err)
			continue
		}
		results[disk.Name] = usage
	}
	
	return results, nil
}

// GetDisks returns configured disk entries
func (m *Manager) GetDisks() []config.DiskEntry {
	return m.disks
}

func (m *Manager) getDiskUsageForPath(path string) (*DiskUsage, error) {
	if path == "" {
		path = "/"
	}

	// Check if the path exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("monitor path does not exist: %s", path)
		}
		return nil, fmt.Errorf("failed to access monitor path %s: %w", path, err)
	}

	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats for %s: %w", path, err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)

	usage := &DiskUsage{
		TotalBytes:      total,
		UsedBytes:       used,
		FreeBytes:       free,
		UsagePercentage: float64(used) / float64(total) * 100,
	}

	return usage, nil
}