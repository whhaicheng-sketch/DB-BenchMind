// Package adapter provides benchmark tool adapter interfaces and implementations.
// Implements: Phase 3 - Tool Adapters
package adapter

import (
	"context"
	"io"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// AdapterType represents the type of benchmark adapter.
type AdapterType string

const (
	// AdapterTypeSysbench is for sysbench tool.
	AdapterTypeSysbench AdapterType = "sysbench"
	// AdapterTypeSwingbench is for swingbench tool.
	AdapterTypeSwingbench AdapterType = "swingbench"
	// AdapterTypeHammerDB is for hammerdb tool.
	AdapterTypeHammerDB AdapterType = "hammerdb"
	// AdapterTypeTPCC is for tpcc tool.
	AdapterTypeTPCC AdapterType = "tpcc"
)

// Config represents the configuration for running a benchmark.
// Implements: REQ-EXEC-001, REQ-EXEC-002
type Config struct {
	// Connection information
	Connection connection.Connection `json:"connection"`
	// Template to use
	Template *template.Template `json:"template"`
	// Template parameters
	Parameters map[string]interface{} `json:"parameters"`
	// Execution options
	Options execution.TaskOptions `json:"options"`
	// Working directory
	WorkDir string `json:"work_dir"`
}

// Command represents a command to be executed.
type Command struct {
	// Command line (including arguments)
	CmdLine string `json:"cmd_line"`
	// Working directory
	WorkDir string `json:"work_dir"`
	// Environment variables
	Env []string `json:"env,omitempty"`
}

// Result represents the parsed result of a benchmark execution.
// Implements: spec.md 3.5.1
type Result struct {
	// Core metrics
	TPS            float64       `json:"tps"`              // Transactions per second
	LatencyAvg     float64       `json:"latency_avg_ms"`   // Average latency (ms)
	LatencyMin     float64       `json:"latency_min_ms"`   // Minimum latency (ms)
	LatencyMax     float64       `json:"latency_max_ms"`   // Maximum latency (ms)
	LatencyP95     float64       `json:"latency_p95_ms"`   // 95th percentile latency (ms)
	LatencyP99     float64       `json:"latency_p99_ms"`   // 99th percentile latency (ms)
	TotalQueries   int64         `json:"total_queries"`     // Total queries executed
	TotalErrors    int64         `json:"total_errors"`     // Total errors
	ErrorRate      float64       `json:"error_rate"`       // Error rate (%)

	// Statistics
	Duration          time.Duration `json:"duration"`            // Actual run duration
	TotalTransactions int64         `json:"total_transactions"` // Total transactions

	// Raw output for debugging
	RawOutput string `json:"raw_output,omitempty"`
}

// Sample represents a realtime metric sample.
// Implements: REQ-EXEC-004, spec.md 3.5
type Sample struct {
	Timestamp   time.Time `json:"timestamp"`
	TPS         float64   `json:"tps"`
	LatencyAvg  float64   `json:"latency_avg_ms"`
	LatencyP95  float64   `json:"latency_p95_ms"`
	LatencyP99  float64   `json:"latency_p99_ms"`
	ErrorRate   float64   `json:"error_rate"`
	ThreadCount int       `json:"thread_count,omitempty"`
}

// ProgressUpdate represents a progress update during execution.
type ProgressUpdate struct {
	Phase      string    `json:"phase"`       // prepare, warmup, run, cleanup
	Timestamp  time.Time `json:"timestamp"`
	Percentage float64   `json:"percentage"`  // 0-100
	Message    string    `json:"message"`
}

// BenchmarkAdapter defines the interface for benchmark tool adapters.
// Each benchmark tool (sysbench, swingbench, hammerdb) implements this interface.
// Implements: Phase 3 - Tool Adapters
type BenchmarkAdapter interface {
	// Type returns the adapter type.
	Type() AdapterType

	// BuildPrepareCommand builds the command for data preparation phase.
	// Implements: REQ-EXEC-002 (prepare phase)
	BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error)

	// BuildRunCommand builds the command for the main benchmark run.
	// Implements: REQ-EXEC-002 (run phase)
	BuildRunCommand(ctx context.Context, config *Config) (*Command, error)

	// BuildCleanupCommand builds the command for cleanup phase.
	// Implements: REQ-EXEC-002 (cleanup phase)
	BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error)

	// ParseRunOutput parses the output from a benchmark run.
	// Returns the parsed result or an error.
	// Implements: REQ-EXEC-004, REQ-EXEC-008
	ParseRunOutput(ctx context.Context, stdout string, stderr string) (*Result, error)

	// StartRealtimeCollection starts realtime metric collection from the running process.
	// Returns a channel that will receive samples until the context is cancelled.
	// Implements: REQ-EXEC-004 (realtime monitoring)
	StartRealtimeCollection(ctx context.Context, stdout io.Reader, stderr io.Reader) (<-chan Sample, <-chan error)

	// ValidateConfig validates the configuration for this adapter.
	// Returns an error if the configuration is invalid.
	// Implements: REQ-EXEC-001 (pre-check)
	ValidateConfig(ctx context.Context, config *Config) error

	// SupportsDatabase checks if this adapter supports the given database type.
	SupportsDatabase(dbType connection.DatabaseType) bool
}

// AdapterRegistry manages benchmark adapters.
// Implements: Adapter lookup and registration
type AdapterRegistry struct {
	adapters map[AdapterType]BenchmarkAdapter
}

// NewAdapterRegistry creates a new adapter registry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[AdapterType]BenchmarkAdapter),
	}
}

// Register registers a benchmark adapter.
func (r *AdapterRegistry) Register(adapter BenchmarkAdapter) {
	r.adapters[adapter.Type()] = adapter
}

// Get returns an adapter by type.
// Returns nil if the adapter is not registered.
func (r *AdapterRegistry) Get(adapterType AdapterType) BenchmarkAdapter {
	return r.adapters[adapterType]
}

// GetByTool returns an adapter by tool name (from template).
// Returns nil if the adapter is not found.
func (r *AdapterRegistry) GetByTool(tool string) BenchmarkAdapter {
	// Map tool names to adapter types
	switch tool {
	case "sysbench":
		return r.adapters[AdapterTypeSysbench]
	case "swingbench":
		return r.adapters[AdapterTypeSwingbench]
	case "hammerdb":
		return r.adapters[AdapterTypeHammerDB]
	case "tpcc":
		return r.adapters[AdapterTypeTPCC]
	default:
		return nil
	}
}

// List returns all registered adapter types.
func (r *AdapterRegistry) List() []AdapterType {
	var types []AdapterType
	for typ := range r.adapters {
		types = append(types, typ)
	}
	return types
}
