// Package execution provides benchmark run domain model.
// Implements: REQ-STORAGE-001, REQ-STORAGE-005
package execution

import (
	"encoding/json"
	"fmt"
	"time"
)

// Run represents a single execution of a benchmark task.
// Implements: spec.md 3.6.1
type Run struct {
	// Basic information
	ID    string `json:"id"`    // UUID
	TaskID string `json:"task_id"` // Associated task ID

	// State (spec.md 3.4.2)
	State RunState `json:"state"`

	// Timestamps
	CreatedAt   time.Time      `json:"created_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Duration    *time.Duration `json:"duration,omitempty"` // Calculated duration

	// Results
	Result       *BenchmarkResult `json:"result,omitempty"`
	ErrorMessage string           `json:"error_message,omitempty"`

	// Work directory for storing logs and artifacts
	WorkDir string `json:"work_dir,omitempty"`
}

// BenchmarkResult represents the parsed result of a benchmark execution.
// Implements: spec.md 3.5.1
type BenchmarkResult struct {
	// Basic information
	RunID string `json:"run_id"`

	// Core metrics (spec.md 3.5.2)
	TPSCalculated float64 `json:"tps_calculated"` // Calculated TPS
	LatencyAvg    float64 `json:"latency_avg_ms"` // Average latency (ms)
	LatencyP95    float64 `json:"latency_p95_ms"` // 95th percentile latency (ms)
	LatencyP99    float64 `json:"latency_p99_ms"` // 99th percentile latency (ms)
	ErrorCount    int64   `json:"error_count"`    // Total errors
	ErrorRate     float64 `json:"error_rate_percent"` // Error rate (%)

	// Statistics
	Duration          time.Duration `json:"duration"`           // Run duration
	TotalTransactions int64         `json:"total_transactions"` // Total transactions
	TotalQueries      int64         `json:"total_queries,omitempty"` // Total queries

	// Time series data
	TimeSeries []MetricSample `json:"time_series,omitempty"` // Time series metrics
}

// MetricSample represents a single metric sample.
// Implements: spec.md 3.5.1
type MetricSample struct {
	Timestamp   time.Time `json:"timestamp"`    // Sample timestamp
	Phase       string    `json:"phase"`        // Phase: warmup/run/cooldown
	TPS         float64   `json:"tps"`          // Transactions per second
	QPS         float64   `json:"qps,omitempty"` // Queries per second
	LatencyAvg  float64   `json:"latency_avg_ms"` // Average latency (ms)
	LatencyP95  float64   `json:"latency_p95_ms"` // 95th percentile latency (ms)
	LatencyP99  float64   `json:"latency_p99_ms"` // 99th percentile latency (ms)
	ErrorRate   float64   `json:"error_rate_percent"` // Error rate (%)
}

// IsCompleted checks if the run is in a terminal state.
func (r *Run) IsCompleted() bool {
	return r.State.IsTerminal()
}

// SetState sets the state with validation.
// Returns an error if the transition is invalid.
func (r *Run) SetState(newState RunState) error {
	if !r.State.CanTransitionTo(newState) {
		return &InvalidStateTransitionError{
			From: r.State,
			To:   newState,
		}
	}
	r.State = newState
	return nil
}

// CalculateDuration calculates and sets the duration based on started_at and completed_at.
func (r *Run) CalculateDuration() {
	if r.StartedAt != nil && r.CompletedAt != nil {
		duration := r.CompletedAt.Sub(*r.StartedAt)
		r.Duration = &duration
	}
}

// ToJSON serializes the run to JSON.
func (r *Run) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// InvalidStateTransitionError represents an invalid state transition.
type InvalidStateTransitionError struct {
	From RunState
	To   RunState
}

func (e *InvalidStateTransitionError) Error() string {
	return fmt.Sprintf("invalid state transition: %s -> %s", e.From, e.To)
}

// =============================================================================
// Benchmark Task (simplified for now, will be expanded)
// Implements: spec.md 3.4.1
// =============================================================================

// BenchmarkTask represents a benchmark task configuration.
type BenchmarkTask struct {
	ID           string                 `json:"id"`           // UUID
	Name         string                 `json:"name"`         // Task name
	ConnectionID string                 `json:"connection_id"` // Connection ID
	TemplateID   string                 `json:"template_id"`   // Template ID
	Parameters   map[string]interface{} `json:"parameters"`   // Parameter overrides
	Options      TaskOptions            `json:"options"`      // Execution options
	Tags         []string               `json:"tags"`         // Tags
	CreatedAt    time.Time              `json:"created_at"`
}

// Validate validates the task configuration.
func (t *BenchmarkTask) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task id is required")
	}
	if t.Name == "" {
		return fmt.Errorf("task name is required")
	}
	if t.ConnectionID == "" {
		return fmt.Errorf("connection_id is required")
	}
	if t.TemplateID == "" {
		return fmt.Errorf("template_id is required")
	}
	return nil
}

// TaskOptions represents execution options for a task.
// Implements: spec.md 3.4.1
type TaskOptions struct {
	SkipPrepare    bool          `json:"skip_prepare"`     // Skip data preparation
	SkipCleanup    bool          `json:"skip_cleanup"`     // Skip data cleanup
	WarmupTime     int           `json:"warmup_time"`      // Warmup duration (seconds)
	SampleInterval time.Duration `json:"sample_interval"`  // Sample interval (default 1s)
	DryRun         bool          `json:"dry_run"`          // Show commands only, don't execute (REQ-EXEC-010)
	PrepareTimeout time.Duration `json:"prepare_timeout"`  // Prepare phase timeout (default 30m)
	RunTimeout     time.Duration `json:"run_timeout"`      // Run phase timeout (default 24h)
}
