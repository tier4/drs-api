package storage

import (
	"fmt"
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

	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	usage := &DiskUsage{
		TotalBytes:      total,
		UsedBytes:       used,
		FreeBytes:       free,
		UsagePercentage: float64(used) / float64(total) * 100,
		Filesystem:      path,
	}

	return usage, nil
}