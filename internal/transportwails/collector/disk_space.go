// Package collector provides system metrics collection for monitoring.
package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskSpaceCollector collects disk space usage metrics.
type DiskSpaceCollector struct {
	path string
}

// DiskSpaceData represents disk space usage data.
type DiskSpaceData struct {
	UsedPercent float64   `json:"used_percent"`
	UsedGB      float64   `json:"used_gb"`
	TotalGB     float64   `json:"total_gb"`
	FreeGB      float64   `json:"free_gb"`
	Path        string    `json:"path"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewDiskSpaceCollector creates a new disk space collector.
// path is the mount point to monitor (default "/" on Linux, "C:" on Windows).
func NewDiskSpaceCollector(path string) *DiskSpaceCollector {
	if path == "" {
		path = "/" // Default to root
	}
	return &DiskSpaceCollector{
		path: path,
	}
}

// Collect collects disk space usage data.
func (c *DiskSpaceCollector) Collect(ctx context.Context) (*DiskSpaceData, error) {
	usage, err := disk.UsageWithContext(ctx, c.path)
	if err != nil {
		return nil, fmt.Errorf("get disk usage for %s: %w", c.path, err)
	}

	// Convert bytes to GB
	bytesToGB := func(bytes uint64) float64 {
		return float64(bytes) / (1024 * 1024 * 1024)
	}

	data := &DiskSpaceData{
		UsedPercent: usage.UsedPercent,
		UsedGB:      bytesToGB(usage.Used),
		TotalGB:     bytesToGB(usage.Total),
		FreeGB:      bytesToGB(usage.Free),
		Path:        c.path,
		Timestamp:   time.Now(),
	}

	return data, nil
}

// GetRootDiskPath returns the appropriate root disk path for the current OS.
func GetRootDiskPath() string {
	// On Linux/macOS, use "/"
	// On Windows, this would be "C:" but gopsutil handles this
	return "/"
}
