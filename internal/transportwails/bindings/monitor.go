// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/collector"
)

// MonitorBinding provides Wails bindings for real-time monitoring.
type MonitorBinding struct {
	ctx         context.Context
	tpmBuffer   *collector.RingBuffer
	tpsBuffer   *collector.RingBuffer
	mu          sync.RWMutex
	isRunning   bool
	currentRunID string

	// System collector
	systemCollector *collector.SystemCollector
	systemMu        sync.RWMutex
	systemRunning   bool
	systemCancel    context.CancelFunc
	systemWg        sync.WaitGroup
}

// MonitorConfigDTO represents monitoring configuration.
type MonitorConfigDTO struct {
	BufferSize    int `json:"buffer_size"`
	SampleRate    int `json:"sample_rate_ms"`
}

// MonitorStateDTO represents monitoring state.
type MonitorStateDTO struct {
	IsRunning     bool `json:"is_running"`
	RunID         string `json:"run_id"`
	TPMCount      int  `json:"tpm_count"`
	TPSCount      int  `json:"tps_count"`
	SystemRunning bool `json:"system_running"`
}

// MonitorDataDTO represents current monitoring data.
type MonitorDataDTO struct {
	CurrentTPM float64                `json:"current_tpm"`
	CurrentTPS float64                `json:"current_tps"`
	TPMPoints  []collector.MetricPoint `json:"tpm_points"`
	TPSPoints  []collector.MetricPoint `json:"tps_points"`
	Stats      *collector.MetricStats  `json:"stats"`
}

// SystemMetricsDTO represents current system metrics.
type SystemMetricsDTO struct {
	CPUPercent      float64 `json:"cpu_percent"`
	DiskReadBps     float64 `json:"disk_read_bps"`
	DiskWriteBps    float64 `json:"disk_write_bps"`
	DiskUsedPercent float64 `json:"disk_used_percent"`
	DiskUsedGB      float64 `json:"disk_used_gb"`
	DiskTotalGB     float64 `json:"disk_total_gb"`
}

// SystemHistoryDTO represents system metrics history.
type SystemHistoryDTO struct {
	CPU       []collector.SystemMetricPoint `json:"cpu"`
	DiskIO    []collector.SystemMetricPoint `json:"disk_io"`
	DiskSpace []collector.SystemMetricPoint `json:"disk_space"`
}

// NewMonitorBinding creates a new MonitorBinding.
func NewMonitorBinding() *MonitorBinding {
	return &MonitorBinding{
		tpmBuffer: collector.NewRingBuffer(60), // 60 data points for 60 seconds
		tpsBuffer: collector.NewRingBuffer(60),
		systemCollector: collector.NewSystemCollector(collector.SystemCollectorConfig{
			Interval:   time.Second,
			BufferSize: 60,
		}),
	}
}

// SetContext sets the Wails context for event emission.
func (m *MonitorBinding) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// StartMonitoring starts monitoring for a benchmark run.
func (m *MonitorBinding) StartMonitoring(runID string, config MonitorConfigDTO) map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return map[string]interface{}{
			"success": false,
			"error":   "monitoring already running",
		}
	}

	// Create new buffers with configured size
	size := config.BufferSize
	if size <= 0 {
		size = 60
	}

	m.tpmBuffer = collector.NewRingBuffer(size)
	m.tpsBuffer = collector.NewRingBuffer(size)
	m.isRunning = true
	m.currentRunID = runID

	slog.Info("Monitoring started", "run_id", runID, "buffer_size", size)

	// Emit event
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "monitor:started", map[string]string{
			"run_id": runID,
		})
	}

	return map[string]interface{}{
		"success": true,
	}
}

// StopMonitoring stops monitoring.
func (m *MonitorBinding) StopMonitoring() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isRunning = false
	m.currentRunID = ""

	// Emit event
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "monitor:stopped", map[string]bool{
		"success": true,
		})
	}

	slog.Info("Monitoring stopped")

	return map[string]interface{}{
		"success": true,
	}
}

// AddTPMPoint adds a TPM data point.
func (m *MonitorBinding) AddTPMPoint(tpm float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isRunning {
		return
	}

	m.tpmBuffer.AddTPM(tpm)

	// Emit update event
	if m.ctx != nil {
		current := m.tpmBuffer.GetCurrent()
		stats := m.tpmBuffer.CalculateStats()
		runtime.EventsEmit(m.ctx, "monitor:tpm_update", map[string]interface{}{
			"current": current,
			"stats":   stats,
		})
	}
}

// AddTPSPoint adds a TPS data point.
func (m *MonitorBinding) AddTPSPoint(tps float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isRunning {
		return
	}

	m.tpsBuffer.AddTPS(tps)

	// Emit update event
	if m.ctx != nil {
		current := m.tpsBuffer.GetCurrent()
		stats := m.tpsBuffer.CalculateStats()
		runtime.EventsEmit(m.ctx, "monitor:tps_update", map[string]interface{}{
			"current": current,
			"stats":   stats,
		})
	}
}

// AddMetricPoint adds a full metric point (TPM + TPS).
func (m *MonitorBinding) AddMetricPoint(tpm, tps, latency float64, errors int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isRunning {
		return
	}

	// Add to buffers
	m.tpmBuffer.AddFull(tpm, tps, latency, errors)
	m.tpsBuffer.AddFull(tpm, tps, latency, errors)

	// Emit combined update event with enhanced stats
	if m.ctx != nil {
		tpmCurrent := m.tpmBuffer.GetCurrent()
		tpsCurrent := m.tpsBuffer.GetCurrent()
		tpmStats := m.tpmBuffer.CalculateStatsFromWindow(30) // Last 30 seconds for CV
		tpsStats := m.tpsBuffer.CalculateStatsFromWindow(30)

		runtime.EventsEmit(m.ctx, "monitor:metrics_update", map[string]interface{}{
			"tpm_current":         tpmCurrent.TPM,
			"tps_current":         tpsCurrent.TPS,
			"latency":             tpsCurrent.Latency,
			"errors":              tpsCurrent.Errors,
			"tpm_stats":           tpmStats,
			"tps_stats":           tpsStats,
			"tpm_cv":              tpmStats.TPMCV,
			"tps_cv":              tpsStats.TPSCV,
			"tpm_direction_changes": tpmStats.TPMDirectionChanges,
			"tps_direction_changes": tpsStats.TPSDirectionChanges,
		})
	}
}

// GetMonitorData returns current monitoring data.
func (m *MonitorBinding) GetMonitorData() MonitorDataDTO {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := MonitorDataDTO{
		TPMPoints: m.tpmBuffer.GetAll(),
		TPSPoints: m.tpsBuffer.GetAll(),
		Stats:     m.tpmBuffer.CalculateStatsFromWindow(30), // Last 30 seconds
	}

	// Get current values
	if current := m.tpmBuffer.GetCurrent(); current != nil {
		data.CurrentTPM = current.TPM
	}
	if current := m.tpsBuffer.GetCurrent(); current != nil {
		data.CurrentTPS = current.TPS
	}

	return data
}

// GetFluctuationAnalysis returns fluctuation analysis for TPM and TPS.
func (m *MonitorBinding) GetFluctuationAnalysis() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tpmStats := m.tpmBuffer.CalculateStatsFromWindow(30)
	tpsStats := m.tpsBuffer.CalculateStatsFromWindow(30)

	return map[string]interface{}{
		"tpm": map[string]interface{}{
			"avg":               tpmStats.TPMAvg,
			"stddev":            tpmStats.TPMStdDev,
			"cv":                tpmStats.TPMCV,
			"direction_changes": tpmStats.TPMDirectionChanges,
			"status":            collector.GetFluctuationStatus(tpmStats.TPMCV),
		},
		"tps": map[string]interface{}{
			"avg":               tpsStats.TPSAvg,
			"stddev":            tpsStats.TPSStdDev,
			"cv":                tpsStats.TPSCV,
			"direction_changes": tpsStats.TPSDirectionChanges,
			"status":            collector.GetFluctuationStatus(tpsStats.TPSCV),
		},
	}
}

// GetTPMHistory returns TPM history points.
func (m *MonitorBinding) GetTPMHistory(count int) []collector.MetricPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if count <= 0 {
		return m.tpmBuffer.GetAll()
	}
	return m.tpmBuffer.GetLast(count)
}

// GetTPSHistory returns TPS history points.
func (m *MonitorBinding) GetTPSHistory(count int) []collector.MetricPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if count <= 0 {
		return m.tpsBuffer.GetAll()
	}
	return m.tpsBuffer.GetLast(count)
}

// GetMonitorState returns current monitoring state.
func (m *MonitorBinding) GetMonitorState() MonitorStateDTO {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.systemMu.RLock()
	systemRunning := m.systemRunning
	m.systemMu.RUnlock()

	return MonitorStateDTO{
		IsRunning:     m.isRunning,
		RunID:         m.currentRunID,
		TPMCount:      m.tpmBuffer.Count(),
		TPSCount:      m.tpsBuffer.Count(),
		SystemRunning: systemRunning,
	}
}

// ClearData clears all monitoring data.
func (m *MonitorBinding) ClearData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tpmBuffer.Clear()
	m.tpsBuffer.Clear()

	slog.Info("Monitor data cleared")
}

// ================== System Monitoring ==================

// StartSystemMonitoring starts collecting system metrics.
// This starts immediately when the application launches, not tied to benchmark.
func (m *MonitorBinding) StartSystemMonitoring() map[string]interface{} {
	m.systemMu.Lock()
	defer m.systemMu.Unlock()

	if m.systemRunning {
		return map[string]interface{}{
			"success": true,
			"message": "system monitoring already running",
		}
	}

	if err := m.systemCollector.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	m.systemRunning = true

	// Start event emission loop
	ctx, cancel := context.WithCancel(context.Background())
	m.systemCancel = cancel
	m.systemWg.Add(1)
	go m.systemEventLoop(ctx)

	slog.Info("System monitoring started")

	// Emit event
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "system:started", map[string]bool{
			"success": true,
		})
	}

	return map[string]interface{}{
		"success": true,
	}
}

// StopSystemMonitoring stops collecting system metrics.
func (m *MonitorBinding) StopSystemMonitoring() map[string]interface{} {
	m.systemMu.Lock()
	defer m.systemMu.Unlock()

	if !m.systemRunning {
		return map[string]interface{}{
			"success": true,
			"message": "system monitoring not running",
		}
	}

	m.systemCollector.Stop()
	if m.systemCancel != nil {
		m.systemCancel()
	}
	m.systemWg.Wait()

	m.systemRunning = false

	slog.Info("System monitoring stopped")

	// Emit event
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "system:stopped", map[string]bool{
			"success": true,
		})
	}

	return map[string]interface{}{
		"success": true,
	}
}

// systemEventLoop emits system metrics events periodically.
func (m *MonitorBinding) systemEventLoop(ctx context.Context) {
	defer m.systemWg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if m.ctx != nil {
				snapshot := m.systemCollector.GetSnapshot()
				runtime.EventsEmit(m.ctx, "system:metrics_update", SystemMetricsDTO{
					CPUPercent:      snapshot.CPUPercent,
					DiskReadBps:     snapshot.DiskReadBps,
					DiskWriteBps:    snapshot.DiskWriteBps,
					DiskUsedPercent: snapshot.DiskUsedPercent,
					DiskUsedGB:      snapshot.DiskUsedGB,
					DiskTotalGB:     snapshot.DiskTotalGB,
				})
			}
		}
	}
}

// GetSystemMetrics returns current system metrics.
func (m *MonitorBinding) GetSystemMetrics() SystemMetricsDTO {
	snapshot := m.systemCollector.GetSnapshot()
	return SystemMetricsDTO{
		CPUPercent:      snapshot.CPUPercent,
		DiskReadBps:     snapshot.DiskReadBps,
		DiskWriteBps:    snapshot.DiskWriteBps,
		DiskUsedPercent: snapshot.DiskUsedPercent,
		DiskUsedGB:      snapshot.DiskUsedGB,
		DiskTotalGB:     snapshot.DiskTotalGB,
	}
}

// GetSystemHistory returns system metrics history.
func (m *MonitorBinding) GetSystemHistory() SystemHistoryDTO {
	return SystemHistoryDTO{
		CPU:       m.systemCollector.GetCPUHistory(),
		DiskIO:    m.systemCollector.GetDiskIOHistory(),
		DiskSpace: m.systemCollector.GetDiskSpaceHistory(),
	}
}

// IsSystemMonitoring returns whether system monitoring is running.
func (m *MonitorBinding) IsSystemMonitoring() bool {
	m.systemMu.RLock()
	defer m.systemMu.RUnlock()
	return m.systemRunning
}
