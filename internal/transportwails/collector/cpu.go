// Package collector provides system metrics collection for monitoring.
package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

// CPUCollector collects CPU usage metrics.
type CPUCollector struct {
	interval time.Duration
}

// CPUData represents CPU usage data.
type CPUData struct {
	UsagePercent  float64   `json:"usage_percent"`
	UserPercent   float64   `json:"user_percent"`
	SystemPercent float64   `json:"system_percent"`
	IdlePercent   float64   `json:"idle_percent"`
	Cores         int       `json:"cores"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewCPUCollector creates a new CPU collector.
func NewCPUCollector(interval time.Duration) *CPUCollector {
	if interval <= 0 {
		interval = time.Second
	}
	return &CPUCollector{
		interval: interval,
	}
}

// Collect collects CPU usage data once.
func (c *CPUCollector) Collect(ctx context.Context) (*CPUData, error) {
	// Get overall CPU usage (interval-based calculation)
	percent, err := cpu.PercentWithContext(ctx, c.interval, false)
	if err != nil {
		return nil, fmt.Errorf("get cpu percent: %w", err)
	}

	// Get CPU times for detailed breakdown
	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("get cpu times: %w", err)
	}

	// Get CPU core count
	cores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		cores = 1 // Fallback to 1 core
	}

	data := &CPUData{
		Timestamp: time.Now(),
		Cores:     cores,
	}

	// Set overall usage
	if len(percent) > 0 {
		data.UsagePercent = percent[0]
	}

	// Calculate percentages from times
	if len(times) > 0 {
		t := times[0]
		total := t.User + t.System + t.Idle + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal + t.Guest + t.GuestNice
		if total > 0 {
			data.UserPercent = (t.User / total) * 100
			data.SystemPercent = (t.System / total) * 100
			data.IdlePercent = (t.Idle / total) * 100
		}
	}

	return data, nil
}
