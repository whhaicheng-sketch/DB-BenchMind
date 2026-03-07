// Package collector provides system metrics collection for monitoring.
package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskIOCollector collects disk I/O metrics.
type DiskIOCollector struct {
	mu             sync.Mutex
	lastStats      map[string]disk.IOCountersStat
	lastCollect    time.Time
	interval       time.Duration
}

// DiskIOData represents disk I/O data.
type DiskIOData struct {
	ReadBps      float64   `json:"read_bps"`
	WriteBps     float64   `json:"write_bps"`
	ReadOps      uint64    `json:"read_ops"`
	WriteOps     uint64    `json:"write_ops"`
	Timestamp    time.Time `json:"timestamp"`
}

// NewDiskIOCollector creates a new disk I/O collector.
func NewDiskIOCollector(interval time.Duration) *DiskIOCollector {
	if interval <= 0 {
		interval = time.Second
	}
	return &DiskIOCollector{
		lastStats:   make(map[string]disk.IOCountersStat),
		lastCollect: time.Now(),
		interval:    interval,
	}
}

// Collect collects disk I/O data.
// Returns bytes per second for read and write operations.
func (c *DiskIOCollector) Collect(ctx context.Context) (*DiskIOData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get current I/O counters
	currentStats, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get disk io counters: %w", err)
	}

	now := time.Now()
	elapsed := now.Sub(c.lastCollect).Seconds()

	data := &DiskIOData{
		Timestamp: now,
	}

	// If this is the first collection, just store the stats
	if len(c.lastStats) == 0 {
		c.lastStats = currentStats
		c.lastCollect = now
		return data, nil
	}

	// Calculate rates by summing all disks
	var totalReadBytes, totalWriteBytes uint64
	var totalReadOps, totalWriteOps uint64

	for name, current := range currentStats {
		if last, ok := c.lastStats[name]; ok {
			readDiff := current.ReadBytes - last.ReadBytes
			writeDiff := current.WriteBytes - last.WriteBytes

			totalReadBytes += readDiff
			totalWriteBytes += writeDiff
			totalReadOps += current.ReadCount - last.ReadCount
			totalWriteOps += current.WriteCount - last.WriteCount
		}
	}

	// Convert to bytes per second
	if elapsed > 0 {
		data.ReadBps = float64(totalReadBytes) / elapsed
		data.WriteBps = float64(totalWriteBytes) / elapsed
	}
	data.ReadOps = totalReadOps
	data.WriteOps = totalWriteOps

	// Update last stats
	c.lastStats = currentStats
	c.lastCollect = now

	return data, nil
}
