// Package usecase provides history record business logic.
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

const (
	// MaxTimeSeriesSize is the maximum size of time series data in bytes (1MB).
	MaxTimeSeriesSize = 1 * 1024 * 1024
)

// HistoryUseCase provides history record business logic.
type HistoryUseCase struct {
	historyRepo repository.HistoryRepository
}

// NewHistoryUseCase creates a new history use case.
func NewHistoryUseCase(historyRepo repository.HistoryRepository) *HistoryUseCase {
	return &HistoryUseCase{
		historyRepo: historyRepo,
	}
}

// SaveRunToHistory saves a completed benchmark run to history.
func (uc *HistoryUseCase) SaveRunToHistory(ctx context.Context, run *execution.Run) error {
	if run.Result == nil {
		return nil // No result to save
	}

	// Convert execution.MetricSample to history.MetricSample
	timeSeries := make([]history.MetricSample, len(run.Result.TimeSeries))
	for i, sample := range run.Result.TimeSeries {
		timeSeries[i] = history.MetricSample{
			Timestamp:  sample.Timestamp,
			Phase:      sample.Phase,
			TPS:        sample.TPS,
			QPS:        sample.QPS,
			LatencyAvg: sample.LatencyAvg,
			LatencyP95: sample.LatencyP95,
			LatencyP99: sample.LatencyP99,
			ErrorRate:  sample.ErrorRate,
			RawLine:    sample.RawLine,
		}
	}

	// Sample time series if too large
	timeSeries = uc.sampleTimeSeries(timeSeries, MaxTimeSeriesSize)

	// Create history record from run result
	record := &history.Record{
		ID:        run.ID,
		CreatedAt: time.Now(),

		// Connection and Template Info
		ConnectionName: run.Result.ConnectionName,
		TemplateName:   run.Result.TemplateName,
		DatabaseType:   run.Result.DatabaseType,
		Threads:        run.Result.Threads,

		// Timing
		StartTime: run.Result.StartTime,
		Duration:  run.Result.Duration,

		// Core metrics
		TPSCalculated: run.Result.TPSCalculated,

		// Latency (ms)
		LatencyAvg: run.Result.LatencyAvg,
		LatencyMin: run.Result.LatencyMin,
		LatencyMax: run.Result.LatencyMax,
		LatencyP95: run.Result.LatencyP95,
		LatencyP99: run.Result.LatencyP99,
		LatencySum: run.Result.LatencySum,

		// SQL Statistics
		ReadQueries:       run.Result.ReadQueries,
		WriteQueries:      run.Result.WriteQueries,
		OtherQueries:      run.Result.OtherQueries,
		TotalQueries:      run.Result.TotalQueries,
		TotalTransactions: run.Result.TotalTransactions,

		// Errors and Reconnects
		IgnoredErrors: run.Result.IgnoredErrors,
		Reconnects:    run.Result.Reconnects,

		// General Statistics
		TotalTime:   run.Result.TotalTime,
		TotalEvents: run.Result.TotalEvents,

		// Threads Fairness
		EventsAvg:      run.Result.EventsAvg,
		EventsStddev:   run.Result.EventsStddev,
		ExecTimeAvg:    run.Result.ExecTimeAvg,
		ExecTimeStddev: run.Result.ExecTimeStddev,

		// Time Series Data
		TimeSeries: timeSeries,
	}

	err := uc.historyRepo.Save(ctx, record)
	if err != nil {
		return err
	}

	// Verify save by reading back
	saved, err := uc.historyRepo.GetByID(ctx, record.ID)
	if err != nil {
		return fmt.Errorf("saved but cannot verify: %w", err)
	}
	if saved == nil {
		return fmt.Errorf("saved but GetByID returns nil")
	}

	return nil
}

// sampleTimeSeries samples time series data if it exceeds maxSize.
// Keeps first 20% and last 80% of data points.
func (uc *HistoryUseCase) sampleTimeSeries(series []history.MetricSample, maxSize int) []history.MetricSample {
	if len(series) == 0 {
		return series
	}

	// Check size
	data, err := json.Marshal(series)
	if err != nil {
		return series // Return original if marshaling fails
	}
	if len(data) <= maxSize {
		return series // Size is acceptable
	}

	// Sample: keep first 20% and last 80%
	n := len(series)
	headSize := n / 5       // First 20%
	tailSize := (n * 4) / 5 // Last 80%
	sampled := make([]history.MetricSample, 0, headSize+tailSize)

	// Add head
	sampled = append(sampled, series[:headSize]...)

	// Add tail
	sampled = append(sampled, series[n-tailSize:]...)

	return sampled
}

// GetAllRecords retrieves all history records.
func (uc *HistoryUseCase) GetAllRecords(ctx context.Context) ([]*history.Record, error) {
	return uc.historyRepo.GetAll(ctx)
}

// GetRecordByID retrieves a history record by ID.
func (uc *HistoryUseCase) GetRecordByID(ctx context.Context, id string) (*history.Record, error) {
	return uc.historyRepo.GetByID(ctx, id)
}

// DeleteRecord deletes a history record by ID.
func (uc *HistoryUseCase) DeleteRecord(ctx context.Context, id string) error {
	return uc.historyRepo.Delete(ctx, id)
}

// ListRecords retrieves history records with options.
func (uc *HistoryUseCase) ListRecords(ctx context.Context, opts *repository.ListOptions) ([]*history.Record, error) {
	return uc.historyRepo.List(ctx, opts)
}
