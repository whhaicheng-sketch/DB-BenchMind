// Package collector provides a ring buffer for metrics storage.
package collector

import (
	"math"
	"sync"
	"time"
)

// RingBuffer is a thread-safe circular buffer for storing metric points.
type RingBuffer struct {
	mu       sync.RWMutex
	points  []MetricPoint
	size    int
	head    int // Next write position
	count   int // Number of items currently in buffer
}

// NewRingBuffer creates a new ring buffer with the specified size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		points: make([]MetricPoint, size),
		size:  size,
		head:  0,
		count: 0,
	}
}

// Add adds a new metric point to the buffer.
func (rb *RingBuffer) Add(point MetricPoint) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.points[rb.head] = point
	rb.head = (rb.head + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	}
}

// AddTPM adds a TPM data point.
func (rb *RingBuffer) AddTPM(tpm float64) {
	rb.Add(MetricPoint{
		Timestamp: time.Now(),
		TPM:      tpm,
	})
}

// AddTPS adds a TPS data point.
func (rb *RingBuffer) AddTPS(tps float64) {
	rb.Add(MetricPoint{
		Timestamp: time.Now(),
		TPS:       tps,
	})
}

// AddFull adds a full metric point with all metrics.
func (rb *RingBuffer) AddFull(tpm, tps, latency float64, errors int64) {
	rb.Add(MetricPoint{
		Timestamp: time.Now(),
		TPM:       tpm,
		TPS:       tps,
		Latency:    latency,
		Errors:     errors,
	})
}

// GetAll returns all points in chronological order.
func (rb *RingBuffer) GetAll() []MetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return []MetricPoint{}
	}

	result := make([]MetricPoint, rb.count)

	// Calculate the starting position
	start := rb.head - rb.count
	if start < 0 {
		start += rb.size
	}

	// Copy points in chronological order
	for i := 0; i < rb.count; i++ {
		idx := (start + i) % rb.size
		result[i] = rb.points[idx]
	}

	return result
}

// GetLast returns the last n points in chronological order.
func (rb *RingBuffer) GetLast(n int) []MetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return []MetricPoint{}
	}

	if n > rb.count {
		n = rb.count
	}

	result := make([]MetricPoint, n)

	// Calculate the starting position for the last n elements
	start := rb.head - n
	if start < 0 {
		start += rb.size
	}

	// Copy points in chronological order
	for i := 0; i < n; i++ {
		idx := (start + i) % rb.size
		result[i] = rb.points[idx]
	}

	return result
}

// CalculateStats calculates statistics from all points.
func (rb *RingBuffer) CalculateStats() *MetricStats {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	stats := &MetricStats{}

	if rb.count == 0 {
		return stats
	}

	points := rb.GetAll()
	return rb.calculateStatsFromPoints(points)
}

// CalculateStatsFromWindow calculates statistics from the last N seconds.
func (rb *RingBuffer) CalculateStatsFromWindow(windowSeconds int) *MetricStats {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	stats := &MetricStats{}

	if rb.count == 0 {
		return stats
	}

	// Get points within the time window
	cutoff := time.Now().Add(-time.Duration(windowSeconds) * time.Second)
	var windowPoints []MetricPoint

	points := rb.GetAll()
	for _, p := range points {
		if p.Timestamp.After(cutoff) || p.Timestamp.Equal(cutoff) {
			windowPoints = append(windowPoints, p)
		}
	}

	if len(windowPoints) == 0 {
		return stats
	}

	return rb.calculateStatsFromPoints(windowPoints)
}

// calculateStatsFromPoints calculates statistics from a slice of points.
func (rb *RingBuffer) calculateStatsFromPoints(points []MetricPoint) *MetricStats {
	stats := &MetricStats{}
	n := len(points)

	if n == 0 {
		return stats
	}

	// First pass: calculate sum, min, max
	var tpmSum, tpsSum float64
	tpmMin, tpmMax := points[0].TPM, points[0].TPM
	tpsMin, tpsMax := points[0].TPS, points[0].TPS

	for _, p := range points {
		tpmSum += p.TPM
		tpsSum += p.TPS
		if p.TPM < tpmMin {
			tpmMin = p.TPM
		}
		if p.TPM > tpmMax {
			tpmMax = p.TPM
		}
		if p.TPS < tpsMin {
			tpsMin = p.TPS
		}
		if p.TPS > tpsMax {
			tpsMax = p.TPS
		}
	}

	tpmAvg := tpmSum / float64(n)
	tpsAvg := tpsSum / float64(n)

	stats.TPMAvg = tpmAvg
	stats.TPMMax = tpmMax
	stats.TPMMin = tpmMin
	stats.TPSAvg = tpsAvg
	stats.TPSMax = tpsMax
	stats.TPSMin = tpsMin

	// Second pass: calculate standard deviation
	var tpmVarianceSum, tpsVarianceSum float64
	for _, p := range points {
		tpmVarianceSum += math.Pow(p.TPM-tpmAvg, 2)
		tpsVarianceSum += math.Pow(p.TPS-tpsAvg, 2)
	}

	tpmStdDev := math.Sqrt(tpmVarianceSum / float64(n))
	tpsStdDev := math.Sqrt(tpsVarianceSum / float64(n))

	stats.TPMStdDev = tpmStdDev
	stats.TPSStdDev = tpsStdDev

	// Calculate CV (Coefficient of Variation)
	// CV = stddev / mean, only if mean > 0
	if tpmAvg > 0 {
		stats.TPMCV = tpmStdDev / tpmAvg
	}
	if tpsAvg > 0 {
		stats.TPSCV = tpsStdDev / tpsAvg
	}

	// Calculate direction changes (sawtooth detection)
	stats.TPMDirectionChanges = rb.countDirectionChanges(points, func(p MetricPoint) float64 {
		return p.TPM
	})
	stats.TPSDirectionChanges = rb.countDirectionChanges(points, func(p MetricPoint) float64 {
		return p.TPS
	})

	return stats
}

// countDirectionChanges counts how many times the value changes direction.
// This helps detect "sawtooth" patterns.
func (rb *RingBuffer) countDirectionChanges(points []MetricPoint, getValue func(MetricPoint) float64) int {
	if len(points) < 3 {
		return 0
	}

	changes := 0
	prevDiff := 0.0 // 0 = no change, positive = increasing, negative = decreasing

	for i := 1; i < len(points); i++ {
		currDiff := getValue(points[i]) - getValue(points[i-1])

		// Only count significant changes (threshold to avoid noise)
		threshold := 0.001
		if math.Abs(currDiff) < threshold {
			continue
		}

		// Check if direction changed
		if prevDiff != 0 && ((prevDiff > 0 && currDiff < 0) || (prevDiff < 0 && currDiff > 0)) {
			changes++
		}
		prevDiff = currDiff
	}

	return changes
}

// GetCurrent returns the most recent point.
func (rb *RingBuffer) GetCurrent() *MetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	// The last element is at head - 1
	idx := rb.head - 1
	if idx < 0 {
		idx += rb.size
	}

	point := rb.points[idx]
	return &point
}

// Clear resets the buffer.
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.count = 0
	// Clear the points array
	rb.points = make([]MetricPoint, rb.size)
}

// Count returns the number of items in the buffer.
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}
