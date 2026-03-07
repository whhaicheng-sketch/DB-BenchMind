// Package collector provides metrics collection and storage for real-time monitoring.
package collector

import (
	"sync"
	"time"
)

// MetricPoint represents a single metric data point.
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	TPS       float64   `json:"tps"`
	TPM       float64   `json:"tpm"`
	Latency   float64   `json:"latency_ms"`
	Errors    int64     `json:"errors"`
}

// MetricStats represents aggregated statistics.
type MetricStats struct {
	TPSAvg     float64 `json:"tps_avg"`
	TPSMax     float64 `json:"tps_max"`
	TPSMin     float64 `json:"tps_min"`
	TPMAvg     float64 `json:"tpm_avg"`
	TPMMax     float64 `json:"tpm_max"`
	TPMMin     float64 `json:"tpm_min"`
	LatencyAvg float64 `json:"latency_avg"`
	LatencyMax float64 `json:"latency_max"`
	LatencyMin float64 `json:"latency_min"`

	// Fluctuation analysis fields (T25)
	TPMStdDev    float64 `json:"tpm_stddev"`
	TPMCV        float64 `json:"tpm_cv"`
	TPSStdDev    float64 `json:"tps_stddev"`
	TPSCV        float64 `json:"tps_cv"`
	TPMDirectionChanges int `json:"tpm_direction_changes"`
	TPSDirectionChanges int `json:"tps_direction_changes"`
}

// FluctuationStatus represents the stability status based on CV.
type FluctuationStatus string

const (
	StatusStable    FluctuationStatus = "stable"     // CV < 0.05
	StatusFluctuating FluctuationStatus = "fluctuating" // 0.05 <= CV < 0.10
	StatusSawtooth  FluctuationStatus = "sawtooth"   // CV >= 0.10
)

// GetFluctuationStatus returns the status based on CV value.
func GetFluctuationStatus(cv float64) FluctuationStatus {
	if cv < 0.05 {
		return StatusStable
	} else if cv < 0.10 {
		return StatusFluctuating
	}
	return StatusSawtooth
}

// MetricSnapshot represents current metric values for frontend.
type MetricSnapshot struct {
	CurrentTPS float64      `json:"current_tps"`
	CurrentTPM float64      `json:"current_tpm"`
	Stats      *MetricStats `json:"stats,omitempty"`
}

// SystemMetricPoint represents a system metric data point.
type SystemMetricPoint struct {
	Timestamp       int64   `json:"timestamp"`
	CPUPercent      float64 `json:"cpu_percent"`
	DiskReadBps     float64 `json:"disk_read_bps"`
	DiskWriteBps    float64 `json:"disk_write_bps"`
	DiskUsedPercent float64 `json:"disk_used_percent"`
}

// SystemMetricSnapshot represents current system metric values.
type SystemMetricSnapshot struct {
	CPUPercent      float64 `json:"cpu_percent"`
	DiskReadBps     float64 `json:"disk_read_bps"`
	DiskWriteBps    float64 `json:"disk_write_bps"`
	DiskUsedPercent float64 `json:"disk_used_percent"`
	DiskUsedGB      float64 `json:"disk_used_gb"`
	DiskTotalGB     float64 `json:"disk_total_gb"`
}

// SystemRingBuffer is a thread-safe circular buffer for system metrics.
type SystemRingBuffer struct {
	mu      sync.RWMutex
	points  []SystemMetricPoint
	size    int
	head    int
	count   int
}

// NewSystemRingBuffer creates a new system ring buffer with the specified size.
func NewSystemRingBuffer(size int) *SystemRingBuffer {
	return &SystemRingBuffer{
		points: make([]SystemMetricPoint, size),
		size:   size,
		head:   0,
		count:  0,
	}
}

// Add adds a new system metric point to the buffer.
func (rb *SystemRingBuffer) Add(point SystemMetricPoint) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.points[rb.head] = point
	rb.head = (rb.head + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	}
}

// GetAll returns all points in chronological order.
func (rb *SystemRingBuffer) GetAll() []SystemMetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return []SystemMetricPoint{}
	}

	result := make([]SystemMetricPoint, rb.count)
	start := rb.head - rb.count
	if start < 0 {
		start += rb.size
	}

	for i := 0; i < rb.count; i++ {
		idx := (start + i) % rb.size
		result[i] = rb.points[idx]
	}

	return result
}

// GetLast returns the last n points.
func (rb *SystemRingBuffer) GetLast(n int) []SystemMetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return []SystemMetricPoint{}
	}

	if n > rb.count {
		n = rb.count
	}

	result := make([]SystemMetricPoint, n)
	start := rb.head - n
	if start < 0 {
		start += rb.size
	}

	for i := 0; i < n; i++ {
		idx := (start + i) % rb.size
		result[i] = rb.points[idx]
	}

	return result
}

// GetCurrent returns the most recent point.
func (rb *SystemRingBuffer) GetCurrent() *SystemMetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	idx := rb.head - 1
	if idx < 0 {
		idx += rb.size
	}

	point := rb.points[idx]
	return &point
}

// Clear resets the buffer.
func (rb *SystemRingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.count = 0
	rb.points = make([]SystemMetricPoint, rb.size)
}

// Count returns the number of items in the buffer.
func (rb *SystemRingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}
