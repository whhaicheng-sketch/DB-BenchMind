// Package history provides history record domain model.
// Implements: History feature for DB-BenchMind
package history

import (
	"encoding/json"
	"time"
)

// MetricSample represents a single metric sample (time series data).
// Duplicated from execution.MetricSample to avoid circular dependency.
type MetricSample struct {
	Timestamp  time.Time `json:"timestamp"`
	Phase      string    `json:"phase"`
	TPS        float64   `json:"tps"`
	QPS        float64   `json:"qps,omitempty"`
	LatencyAvg float64   `json:"latency_avg_ms"`
	LatencyP95 float64   `json:"latency_p95_ms"`
	LatencyP99 float64   `json:"latency_p99_ms"`
	ErrorRate  float64   `json:"error_rate_percent"`
	RawLine    string    `json:"raw_line,omitempty"`
}

// Record represents a saved benchmark run history record.
// Only successful runs are saved to history.
type Record struct {
	// Basic information
	ID        string    `json:"id"`         // Run ID (UUID)
	CreatedAt time.Time `json:"created_at"` // When the record was created

	// Connection and Template Info
	ConnectionName string `json:"connection_name"` // Connection name
	TemplateName   string `json:"template_name"`   // Template name
	DatabaseType   string `json:"database_type"`   // Database type (MySQL/PostgreSQL)
	Threads        int    `json:"threads"`         // Thread count

	// Timing
	StartTime time.Time     `json:"start_time"` // Benchmark start time
	Duration  time.Duration `json:"duration"`   // Run duration

	// Core metrics
	TPSCalculated float64 `json:"tps_calculated"` // Calculated TPS

	// Latency (ms)
	LatencyAvg float64 `json:"latency_avg_ms"` // Average latency (ms)
	LatencyMin float64 `json:"latency_min_ms"` // Minimum latency (ms)
	LatencyMax float64 `json:"latency_max_ms"` // Maximum latency (ms)
	LatencyP95 float64 `json:"latency_p95_ms"` // 95th percentile latency (ms)
	LatencyP99 float64 `json:"latency_p99_ms"` // 99th percentile latency (ms)
	LatencySum float64 `json:"latency_sum_ms"` // Sum of all latencies (ms)

	// SQL Statistics
	ReadQueries  int64 `json:"read_queries"`  // Read queries
	WriteQueries int64 `json:"write_queries"` // Write queries
	OtherQueries int64 `json:"other_queries"` // Other queries
	TotalQueries int64 `json:"total_queries"` // Total queries

	// Transactions
	TotalTransactions int64 `json:"total_transactions"` // Total transactions

	// Errors and Reconnects
	IgnoredErrors int64 `json:"ignored_errors"` // Ignored errors
	Reconnects    int64 `json:"reconnects"`     // Reconnects

	// General Statistics
	TotalTime   float64 `json:"total_time_seconds"` // Total time in seconds
	TotalEvents int64   `json:"total_events"`       // Total number of events

	// Threads Fairness
	EventsAvg      float64 `json:"events_avg"`       // Events average
	EventsStddev   float64 `json:"events_stddev"`    // Events stddev
	ExecTimeAvg    float64 `json:"exec_time_avg"`    // Execution time average
	ExecTimeStddev float64 `json:"exec_time_stddev"` // Execution time stddev

	// Time Series Data (realtime metrics during benchmark)
	TimeSeries []MetricSample `json:"time_series,omitempty"` // Time series samples
}

// GetTimeSeriesSize returns the approximate size of time series data in bytes when marshaled to JSON.
func (r *Record) GetTimeSeriesSize() int {
	if len(r.TimeSeries) == 0 {
		return 0
	}
	data, err := json.Marshal(r.TimeSeries)
	if err != nil {
		return 0
	}
	return len(data)
}
