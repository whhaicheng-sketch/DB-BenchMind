// Package report provides JSON report generator implementation.
// Implements: Phase 5 - Report Generation (JSON)
package report

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// JSONGenerator generates JSON format reports.
type JSONGenerator struct{}

// NewJSONGenerator creates a new JSON generator.
func NewJSONGenerator() *JSONGenerator {
	return &JSONGenerator{}
}

// Generate generates a JSON report.
func (g *JSONGenerator) Generate(data *report.GenerateContext) (*report.Report, error) {
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Build JSON structure
	output := g.buildJSON(data)

	// Marshal to JSON
	content, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	return &report.Report{
		Format:      report.FormatJSON,
		Content:     content,
		GeneratedAt: time.Now(),
		RunID:       data.RunID,
	}, nil
}

// Format returns the format this generator produces.
func (g *JSONGenerator) Format() report.ReportFormat {
	return report.FormatJSON
}

// jsonReport represents the JSON report structure.
type jsonReport struct {
	Meta        jsonMeta               `json:"meta"`
	Summary     jsonSummary            `json:"summary"`
	Environment jsonEnvironment        `json:"environment,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Metrics     jsonMetrics            `json:"metrics,omitempty"`
	TimeSeries  []jsonSample           `json:"time_series,omitempty"`
	Logs        []jsonLogEntry         `json:"logs,omitempty"`
	RawOutput   string                 `json:"raw_output,omitempty"`
}

// jsonMeta represents report metadata.
type jsonMeta struct {
	RunID       string `json:"run_id"`
	Format      string `json:"format"`
	GeneratedAt string `json:"generated_at"`
	Version     string `json:"version"`
}

// jsonSummary represents the summary section.
type jsonSummary struct {
	Status      string `json:"status"`
	Tool        string `json:"tool"`
	Template    string `json:"template"`
	Connection  string `json:"connection"`
	DBType      string `json:"db_type"`
	Duration    string `json:"duration"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	Error       string `json:"error,omitempty"`
}

// jsonEnvironment represents environment information.
type jsonEnvironment struct {
	RunID   string `json:"run_id"`
	TaskID  string `json:"task_id"`
	State   string `json:"state"`
	Created string `json:"created_at"`
}

// jsonMetrics represents metrics.
type jsonMetrics struct {
	TPS               float64 `json:"tps"`
	LatencyAvg        float64 `json:"latency_avg_ms"`
	LatencyP95        float64 `json:"latency_p95_ms,omitempty"`
	LatencyP99        float64 `json:"latency_p99_ms,omitempty"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalQueries      int64   `json:"total_queries,omitempty"`
	ErrorCount        int64   `json:"error_count"`
	ErrorRate         float64 `json:"error_rate_percent"`
}

// jsonSample represents a time series sample.
type jsonSample struct {
	Timestamp  string  `json:"timestamp"`
	TPS        float64 `json:"tps"`
	LatencyAvg float64 `json:"latency_avg_ms"`
	LatencyP95 float64 `json:"latency_p95_ms,omitempty"`
	LatencyP99 float64 `json:"latency_p99_ms,omitempty"`
	ErrorRate  float64 `json:"error_rate_percent"`
}

// jsonLogEntry represents a log entry.
type jsonLogEntry struct {
	Timestamp string `json:"timestamp"`
	Stream    string `json:"stream"`
	Content   string `json:"content"`
}

// buildJSON builds the JSON report structure.
func (g *JSONGenerator) buildJSON(data *report.GenerateContext) *jsonReport {
	// Build summary
	summary := jsonSummary{
		Status:     "completed",
		Tool:       data.Tool,
		Template:   data.TemplateName,
		Connection: data.ConnectionName,
		DBType:     data.ConnectionType,
		Duration:   data.GetDuration(),
	}

	if data.IsFailed() {
		summary.Status = "failed"
		summary.Error = data.ErrorMessage
	}

	if data.StartedAt != nil {
		summary.StartedAt = report.GetTimestamp(data.StartedAt)
	}
	if data.CompletedAt != nil {
		summary.CompletedAt = report.GetTimestamp(data.CompletedAt)
	}

	// Build environment
	env := jsonEnvironment{
		RunID:   data.RunID,
		TaskID:  data.TaskID,
		State:   data.State,
		Created: data.CreatedAt.Format(time.RFC3339),
	}

	// Build metrics
	metrics := jsonMetrics{
		TPS:               data.TPS,
		LatencyAvg:        data.LatencyAvg,
		TotalTransactions: data.TotalTransactions,
		ErrorCount:        data.ErrorCount,
		ErrorRate:         data.ErrorRate,
	}

	if data.LatencyP95 > 0 {
		metrics.LatencyP95 = data.LatencyP95
	}
	if data.LatencyP99 > 0 {
		metrics.LatencyP99 = data.LatencyP99
	}
	if data.TotalQueries > 0 {
		metrics.TotalQueries = data.TotalQueries
	}

	// Build time series
	timeSeries := make([]jsonSample, len(data.Samples))
	for i, s := range data.Samples {
		timeSeries[i] = jsonSample{
			Timestamp:  s.Timestamp.Format(time.RFC3339),
			TPS:        s.TPS,
			LatencyAvg: s.LatencyAvg,
			LatencyP95: s.LatencyP95,
			LatencyP99: s.LatencyP99,
			ErrorRate:  s.ErrorRate,
		}
	}

	// Build logs
	logs := make([]jsonLogEntry, len(data.Logs))
	for i, l := range data.Logs {
		logs[i] = jsonLogEntry{
			Timestamp: l.Timestamp,
			Stream:    l.Stream,
			Content:   l.Content,
		}
	}

	// Build report
	r := &jsonReport{
		Meta: jsonMeta{
			RunID:       data.RunID,
			Format:      report.FormatJSON.String(),
			GeneratedAt: time.Now().Format(time.RFC3339),
			Version:     "1.0",
		},
		Summary:     summary,
		Environment: env,
		Parameters:  data.Parameters,
		Metrics:     metrics,
		TimeSeries:  timeSeries,
		Logs:        logs,
		RawOutput:   data.RawOutput,
	}

	return r
}
