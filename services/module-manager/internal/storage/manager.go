package storage

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

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
func (m *Manager) GetDiskUsage(ctx context.Context, diskName string) (*DiskUsage, error) {
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

	return m.getDiskUsageForPath(ctx, targetDisk.MountPath)
}

// GetAllDiskUsages returns usage for all configured disks
func (m *Manager) GetAllDiskUsages(ctx context.Context) (map[string]*DiskUsage, error) {
	results := make(map[string]*DiskUsage)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, disk := range m.disks {
		wg.Add(1)
		go func(d config.DiskEntry) {
			defer wg.Done()
			
			// Create a child context with timeout for each disk check
			// We use a short timeout to avoid blocking the whole request if a disk is hung
			diskCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
			defer cancel()

			usage, err := m.getDiskUsageForPath(diskCtx, d.MountPath)
			if err != nil {
				// Log error but continue with other disks
				fmt.Printf("Failed to get usage for disk %s: %v\n", d.Name, err)
				return
			}

			mu.Lock()
			results[d.Name] = usage
			mu.Unlock()
		}(disk)
	}

	wg.Wait()
	return results, nil
}

// GetDisks returns configured disk entries
func (m *Manager) GetDisks() []config.DiskEntry {
	return m.disks
}

func (m *Manager) getDiskUsageForPath(ctx context.Context, path string) (*DiskUsage, error) {
	type result struct {
		usage *DiskUsage
		err   error
	}

	ch := make(chan result, 1)

	go func() {
		if path == "" {
			path = "/"
		}

		// Check if the path exists
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				ch <- result{nil, fmt.Errorf("monitor path does not exist: %s", path)}
				return
			}
			ch <- result{nil, fmt.Errorf("failed to access monitor path %s: %w", path, err)}
			return
		}

		var stat syscall.Statfs_t
		err := syscall.Statfs(path, &stat)
		if err != nil {
			ch <- result{nil, fmt.Errorf("failed to get filesystem stats for %s: %w", path, err)}
			return
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

		ch <- result{usage, nil}
	}()

	select {
	case res := <-ch:
		return res.usage, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}