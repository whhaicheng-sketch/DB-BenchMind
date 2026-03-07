// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// BenchmarkBinding provides Wails bindings for benchmark execution.
type BenchmarkBinding struct {
	uc         *usecase.BenchmarkUseCase
	connUC     *usecase.ConnectionUseCase
	templateUC *usecase.TemplateUseCase
	ctx        context.Context // Wails context for event emission
}

// NewBenchmarkBinding creates a new BenchmarkBinding.
func NewBenchmarkBinding(uc *usecase.BenchmarkUseCase, connUC *usecase.ConnectionUseCase, templateUC *usecase.TemplateUseCase) *BenchmarkBinding {
	return &BenchmarkBinding{
		uc:         uc,
		connUC:     connUC,
		templateUC: templateUC,
	}
}

// SetContext sets the Wails context for event emission.
// This is called by the app during startup.
func (b *BenchmarkBinding) SetContext(ctx context.Context) {
	b.ctx = ctx
}

// emitLog emits a log event to the frontend.
func (b *BenchmarkBinding) emitLog(runID, stream, content string) {
	if b.ctx == nil {
		return
	}
	runtime.EventsEmit(b.ctx, "log:append", map[string]string{
		"run_id":  runID,
		"stream":  stream,
		"content": content,
	})
}

// emitStatus emits a status event to the frontend.
func (b *BenchmarkBinding) emitStatus(runID, state, message string) {
	if b.ctx == nil {
		return
	}
	runtime.EventsEmit(b.ctx, "benchmark:status", map[string]string{
		"run_id":  runID,
		"state":   state,
		"message": message,
	})
}

// emitMetric emits a metric event to the frontend.
func (b *BenchmarkBinding) emitMetric(runID string, metrics map[string]float64) {
	if b.ctx == nil {
		return
	}
	data := map[string]any{"run_id": runID}
	for k, v := range metrics {
		data[k] = v
	}
	runtime.EventsEmit(b.ctx, "benchmark:metric", data)
}

// emitProgress emits a progress event to the frontend.
func (b *BenchmarkBinding) emitProgress(runID string, percentage float64, phase string) {
	if b.ctx == nil {
		return
	}
	runtime.EventsEmit(b.ctx, "benchmark:progress", map[string]any{
		"run_id":     runID,
		"percentage": percentage,
		"phase":      phase,
	})
}

// =============================================================================
// DTO Types
// =============================================================================

// BenchmarkStartRequest represents a request to start a benchmark.
type BenchmarkStartRequest struct {
	ConnectionID string                 `json:"connection_id"`
	TemplateID   string                 `json:"template_id"`
	Parameters   map[string]interface{} `json:"parameters"`
	Options      BenchmarkOptionsDTO    `json:"options"`
}

// BenchmarkOptionsDTO represents execution options.
type BenchmarkOptionsDTO struct {
	SkipPrepare bool `json:"skip_prepare"`
	SkipCleanup bool `json:"skip_cleanup"`
	WarmupTime  int  `json:"warmup_time"`
	DryRun      bool `json:"dry_run"`
}

// BenchmarkRunDTO represents a benchmark run for frontend.
type BenchmarkRunDTO struct {
	ID            string              `json:"id"`
	TaskID        string              `json:"task_id"`
	State         string              `json:"state"`
	CreatedAt     string              `json:"created_at"`
	StartedAt     string              `json:"started_at,omitempty"`
	CompletedAt   string              `json:"completed_at,omitempty"`
	Duration      string              `json:"duration,omitempty"`
	DurationMs    int64               `json:"duration_ms,omitempty"`
	Result        *BenchmarkResultDTO `json:"result,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
	Message       string              `json:"message,omitempty"`
}

// BenchmarkResultDTO represents benchmark result for frontend.
type BenchmarkResultDTO struct {
	RunID              string  `json:"run_id"`
	TPS                float64 `json:"tps"`
	TPM                float64 `json:"tpm"`
	MaxTPS             float64 `json:"max_tps"`
	AvgTPS             float64 `json:"avg_tps"`
	MaxTPM             float64 `json:"max_tpm"`
	AvgTPM             float64 `json:"avg_tpm"`
	LatencyAvg         float64 `json:"latency_avg_ms"`
	LatencyMin         float64 `json:"latency_min_ms"`
	LatencyMax         float64 `json:"latency_max_ms"`
	LatencyP95         float64 `json:"latency_p95_ms"`
	LatencyP99         float64 `json:"latency_p99_ms"`
	ErrorCount         int64   `json:"error_count"`
	ErrorRate          float64 `json:"error_rate_percent"`
	TotalTransactions  int64   `json:"total_transactions"`
	DurationSeconds    float64 `json:"duration_seconds"`
	ConnectionName     string  `json:"connection_name,omitempty"`
	TemplateName       string  `json:"template_name,omitempty"`
	DatabaseType       string  `json:"database_type,omitempty"`
	Threads            int     `json:"threads,omitempty"`
}

// BenchmarkStatusResult represents the result of GetBenchmarkStatus.
type BenchmarkStatusResult struct {
	Run   *BenchmarkRunDTO `json:"run"`
	Error string           `json:"error,omitempty"`
}

// BenchmarkStartResult represents the result of StartBenchmark.
type BenchmarkStartResult struct {
	RunID string `json:"run_id"`
	Error string `json:"error,omitempty"`
}

// BenchmarkListResult represents the result of ListBenchmarks.
type BenchmarkListResult struct {
	Runs  []BenchmarkRunDTO `json:"runs"`
	Error string            `json:"error,omitempty"`
}

// =============================================================================
// Binding Methods
// =============================================================================

// StartBenchmark starts a new benchmark run (prepare + run + cleanup).
func (b *BenchmarkBinding) StartBenchmark(req BenchmarkStartRequest) BenchmarkStartResult {
	ctx := context.Background()

	// Create task
	task := &execution.BenchmarkTask{
		ID:           uuid.New().String(),
		Name:         "Benchmark Run",
		ConnectionID: req.ConnectionID,
		TemplateID:   req.TemplateID,
		Parameters:   req.Parameters,
		Options: execution.TaskOptions{
			SkipPrepare: req.Options.SkipPrepare,
			SkipCleanup: req.Options.SkipCleanup,
			WarmupTime:  req.Options.WarmupTime,
			DryRun:      req.Options.DryRun,
		},
		CreatedAt: time.Now(),
	}

	run, err := b.uc.StartBenchmark(ctx, task)
	if err != nil {
		slog.Error("StartBenchmark failed", "error", err)
		return BenchmarkStartResult{
			Error: err.Error(),
		}
	}

	return BenchmarkStartResult{
		RunID: run.ID,
	}
}

// PrepareOnly runs only the prepare phase.
func (b *BenchmarkBinding) PrepareOnly(req BenchmarkStartRequest) BenchmarkStartResult {
	ctx := context.Background()

	task := &execution.BenchmarkTask{
		ID:           uuid.New().String(),
		Name:         "Prepare Only",
		ConnectionID: req.ConnectionID,
		TemplateID:   req.TemplateID,
		Parameters:   req.Parameters,
		Options: execution.TaskOptions{
			SkipPrepare: false,
			SkipCleanup: true, // Skip cleanup
			DryRun:      req.Options.DryRun,
		},
		CreatedAt: time.Now(),
	}

	// Set prepare-only mode via custom parameter
	task.Parameters["_prepare_only"] = true

	run, err := b.uc.StartBenchmark(ctx, task)
	if err != nil {
		slog.Error("PrepareOnly failed", "error", err)
		return BenchmarkStartResult{
			Error: err.Error(),
		}
	}

	return BenchmarkStartResult{
		RunID: run.ID,
	}
}

// RunBenchmark runs the benchmark (skip prepare, skip cleanup).
func (b *BenchmarkBinding) RunBenchmark(req BenchmarkStartRequest) BenchmarkStartResult {
	ctx := context.Background()

	task := &execution.BenchmarkTask{
		ID:           uuid.New().String(),
		Name:         "Benchmark Run",
		ConnectionID: req.ConnectionID,
		TemplateID:   req.TemplateID,
		Parameters:   req.Parameters,
		Options: execution.TaskOptions{
			SkipPrepare: true,  // Skip prepare
			SkipCleanup: true,  // Skip cleanup
			WarmupTime:  req.Options.WarmupTime,
			DryRun:      req.Options.DryRun,
		},
		CreatedAt: time.Now(),
	}

	run, err := b.uc.StartBenchmark(ctx, task)
	if err != nil {
		slog.Error("RunBenchmark failed", "error", err)
		return BenchmarkStartResult{
			Error: err.Error(),
		}
	}

	return BenchmarkStartResult{
		RunID: run.ID,
	}
}

// CleanupOnly runs only the cleanup phase.
func (b *BenchmarkBinding) CleanupOnly(req BenchmarkStartRequest) BenchmarkStartResult {
	ctx := context.Background()

	task := &execution.BenchmarkTask{
		ID:           uuid.New().String(),
		Name:         "Cleanup Only",
		ConnectionID: req.ConnectionID,
		TemplateID:   req.TemplateID,
		Parameters:   req.Parameters,
		Options: execution.TaskOptions{
			SkipPrepare: true,
			SkipCleanup: false,
			DryRun:      req.Options.DryRun,
		},
		CreatedAt: time.Now(),
	}

	// Set cleanup-only mode via custom parameter
	task.Parameters["_cleanup_only"] = true

	run, err := b.uc.StartBenchmark(ctx, task)
	if err != nil {
		slog.Error("CleanupOnly failed", "error", err)
		return BenchmarkStartResult{
			Error: err.Error(),
		}
	}

	return BenchmarkStartResult{
		RunID: run.ID,
	}
}

// StopBenchmark stops a running benchmark.
func (b *BenchmarkBinding) StopBenchmark(runID string, force bool) map[string]interface{} {
	ctx := context.Background()

	err := b.uc.StopBenchmark(ctx, runID, force)
	if err != nil {
		slog.Error("StopBenchmark failed", "runID", runID, "error", err)
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// GetBenchmarkStatus returns the status of a benchmark run.
func (b *BenchmarkBinding) GetBenchmarkStatus(runID string) BenchmarkStatusResult {
	ctx := context.Background()

	run, err := b.uc.GetBenchmarkStatus(ctx, runID)
	if err != nil {
		slog.Error("GetBenchmarkStatus failed", "runID", runID, "error", err)
		return BenchmarkStatusResult{
			Error: err.Error(),
		}
	}

	return BenchmarkStatusResult{
		Run: b.toDTO(run),
	}
}

// ListBenchmarks returns a list of benchmark runs.
func (b *BenchmarkBinding) ListBenchmarks(limit int) BenchmarkListResult {
	ctx := context.Background()

	if limit <= 0 {
		limit = 20
	}

	runs, err := b.uc.ListBenchmarks(ctx, usecase.FindOptions{
		Limit: limit,
	})
	if err != nil {
		slog.Error("ListBenchmarks failed", "error", err)
		return BenchmarkListResult{
			Error: err.Error(),
		}
	}

	dtos := make([]BenchmarkRunDTO, 0, len(runs))
	for _, run := range runs {
		dtos = append(dtos, *b.toDTO(run))
	}

	return BenchmarkListResult{
		Runs: dtos,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// toDTO converts a Run to BenchmarkRunDTO.
func (b *BenchmarkBinding) toDTO(run *execution.Run) *BenchmarkRunDTO {
	if run == nil {
		return nil
	}

	dto := &BenchmarkRunDTO{
		ID:           run.ID,
		TaskID:       run.TaskID,
		State:        string(run.State),
		CreatedAt:    run.CreatedAt.Format(time.RFC3339),
		ErrorMessage: run.ErrorMessage,
		Message:      run.Message,
	}

	if run.StartedAt != nil {
		dto.StartedAt = run.StartedAt.Format(time.RFC3339)
	}

	if run.CompletedAt != nil {
		dto.CompletedAt = run.CompletedAt.Format(time.RFC3339)
	}

	if run.Duration != nil {
		dto.Duration = run.Duration.String()
		dto.DurationMs = run.Duration.Milliseconds()
	}

	if run.Result != nil {
		dto.Result = b.resultToDTO(run.Result)
	}

	return dto
}

// resultToDTO converts a BenchmarkResult to BenchmarkResultDTO.
func (b *BenchmarkBinding) resultToDTO(result *execution.BenchmarkResult) *BenchmarkResultDTO {
	if result == nil {
		return nil
	}

	return &BenchmarkResultDTO{
		RunID:             result.RunID,
		TPS:               result.TPSCalculated,
		TPM:               result.TPMCalculated,
		MaxTPS:            result.MaxTPS,
		AvgTPS:            result.AvgTPS,
		MaxTPM:            result.MaxTPM,
		AvgTPM:            result.AvgTPM,
		LatencyAvg:        result.LatencyAvg,
		LatencyMin:        result.LatencyMin,
		LatencyMax:        result.LatencyMax,
		LatencyP95:        result.LatencyP95,
		LatencyP99:        result.LatencyP99,
		ErrorCount:        result.ErrorCount,
		ErrorRate:         result.ErrorRate,
		TotalTransactions: result.TotalTransactions,
		DurationSeconds:   result.Duration.Seconds(),
		ConnectionName:    result.ConnectionName,
		TemplateName:      result.TemplateName,
		DatabaseType:      result.DatabaseType,
		Threads:           result.Threads,
	}
}
