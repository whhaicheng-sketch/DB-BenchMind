// Package collector provides unified system metrics collection.
package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// SystemCollector is the unified system metrics collector.
// It collects CPU, Disk I/O, and Disk Space metrics.
type SystemCollector struct {
	mu sync.RWMutex

	// Collectors
	cpuCollector       *CPUCollector
	diskIOCollector    *DiskIOCollector
	diskSpaceCollector *DiskSpaceCollector

	// State
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// Buffers for history
	cpuBuffer       *SystemRingBuffer
	diskIOBuffer    *SystemRingBuffer
	diskSpaceBuffer *SystemRingBuffer

	// Current values
	currentCPU       float64
	currentDiskRead  float64
	currentDiskWrite float64
	currentDiskSpace float64
	diskUsedGB       float64
	diskTotalGB      float64

	// Configuration
	interval   time.Duration
	bufferSize int
}

// SystemCollectorConfig holds configuration for SystemCollector.
type SystemCollectorConfig struct {
	Interval   time.Duration
	BufferSize int
	DiskPath   string
}

// NewSystemCollector creates a new unified system collector.
func NewSystemCollector(config SystemCollectorConfig) *SystemCollector {
	if config.Interval <= 0 {
		config.Interval = time.Second
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 60
	}
	if config.DiskPath == "" {
		config.DiskPath = GetRootDiskPath()
	}

	return &SystemCollector{
		cpuCollector:       NewCPUCollector(config.Interval),
		diskIOCollector:    NewDiskIOCollector(config.Interval),
		diskSpaceCollector: NewDiskSpaceCollector(config.DiskPath),
		cpuBuffer:          NewSystemRingBuffer(config.BufferSize),
		diskIOBuffer:       NewSystemRingBuffer(config.BufferSize),
		diskSpaceBuffer:    NewSystemRingBuffer(config.BufferSize),
		interval:           config.Interval,
		bufferSize:         config.BufferSize,
	}
}

// Start begins collecting system metrics.
func (sc *SystemCollector) Start() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.running {
		return nil
	}

	sc.ctx, sc.cancel = context.WithCancel(context.Background())
	sc.running = true

	sc.wg.Add(1)
	go sc.collectLoop()

	slog.Info("System collector started", "interval", sc.interval)
	return nil
}

// Stop stops collecting system metrics.
func (sc *SystemCollector) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.running {
		return
	}

	sc.running = false
	if sc.cancel != nil {
		sc.cancel()
	}
	sc.wg.Wait()

	slog.Info("System collector stopped")
}

// collectLoop is the main collection loop.
func (sc *SystemCollector) collectLoop() {
	defer sc.wg.Done()

	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()

	// Collect immediately on start
	sc.collect()

	for {
		select {
		case <-sc.ctx.Done():
			return
		case <-ticker.C:
			sc.collect()
		}
	}
}

// collect performs one collection cycle.
func (sc *SystemCollector) collect() {
	ctx := sc.ctx
	now := time.Now().Unix()

	// Collect CPU
	cpuData, err := sc.cpuCollector.Collect(ctx)
	if err != nil {
		slog.Debug("failed to collect CPU", "error", err)
	} else if cpuData != nil {
		sc.mu.Lock()
		sc.currentCPU = cpuData.UsagePercent
		sc.cpuBuffer.Add(SystemMetricPoint{
			Timestamp:  now,
			CPUPercent: cpuData.UsagePercent,
		})
		sc.mu.Unlock()
	}

	// Collect Disk I/O
	diskIOData, err := sc.diskIOCollector.Collect(ctx)
	if err != nil {
		slog.Debug("failed to collect Disk I/O", "error", err)
	} else if diskIOData != nil {
		sc.mu.Lock()
		sc.currentDiskRead = diskIOData.ReadBps
		sc.currentDiskWrite = diskIOData.WriteBps
		sc.diskIOBuffer.Add(SystemMetricPoint{
			Timestamp:   now,
			DiskReadBps:  diskIOData.ReadBps,
			DiskWriteBps: diskIOData.WriteBps,
		})
		sc.mu.Unlock()
	}

	// Collect Disk Space (less frequently, as it changes slowly)
	diskSpaceData, err := sc.diskSpaceCollector.Collect(ctx)
	if err != nil {
		slog.Debug("failed to collect Disk Space", "error", err)
	} else if diskSpaceData != nil {
		sc.mu.Lock()
		sc.currentDiskSpace = diskSpaceData.UsedPercent
		sc.diskUsedGB = diskSpaceData.UsedGB
		sc.diskTotalGB = diskSpaceData.TotalGB
		sc.diskSpaceBuffer.Add(SystemMetricPoint{
			Timestamp:       now,
			DiskUsedPercent: diskSpaceData.UsedPercent,
		})
		sc.mu.Unlock()
	}
}

// GetSnapshot returns current system metrics.
func (sc *SystemCollector) GetSnapshot() *SystemMetricSnapshot {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return &SystemMetricSnapshot{
		CPUPercent:      sc.currentCPU,
		DiskReadBps:     sc.currentDiskRead,
		DiskWriteBps:    sc.currentDiskWrite,
		DiskUsedPercent: sc.currentDiskSpace,
		DiskUsedGB:      sc.diskUsedGB,
		DiskTotalGB:     sc.diskTotalGB,
	}
}

// GetCPUHistory returns CPU history points.
func (sc *SystemCollector) GetCPUHistory() []SystemMetricPoint {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.cpuBuffer.GetAll()
}

// GetDiskIOHistory returns Disk I/O history points.
func (sc *SystemCollector) GetDiskIOHistory() []SystemMetricPoint {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.diskIOBuffer.GetAll()
}

// GetDiskSpaceHistory returns Disk Space history points.
func (sc *SystemCollector) GetDiskSpaceHistory() []SystemMetricPoint {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.diskSpaceBuffer.GetAll()
}

// IsRunning returns whether the collector is running.
func (sc *SystemCollector) IsRunning() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.running
}

// ClearBuffers clears all history buffers.
func (sc *SystemCollector) ClearBuffers() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.cpuBuffer.Clear()
	sc.diskIOBuffer.Clear()
	sc.diskSpaceBuffer.Clear()
}
