package storage

import (
	"fmt"
	"os"
	"syscall"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

type DiskUsage struct {
	TotalBytes      uint64
	UsedBytes       uint64
	FreeBytes       uint64
	UsagePercentage float64
	Filesystem      string
}

func (m *Manager) GetDiskUsage(path string) (*DiskUsage, error) {
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
		Filesystem:      path,
	}

	return usage, nil
}