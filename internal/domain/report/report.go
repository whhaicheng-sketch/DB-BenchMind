// Package report provides report generation domain models.
// Implements: Phase 5 - Report Generation and Export
package report

import (
	"fmt"
	"time"
)

// ReportFormat represents the output format for a report.
type ReportFormat string

const (
	// FormatMarkdown generates Markdown format reports.
	FormatMarkdown ReportFormat = "markdown"
	// FormatHTML generates HTML format reports.
	FormatHTML ReportFormat = "html"
	// FormatJSON generates JSON format reports.
	FormatJSON ReportFormat = "json"
	// FormatPDF generates PDF format reports.
	FormatPDF ReportFormat = "pdf"
)

// String returns the string representation of the format.
func (f ReportFormat) String() string {
	return string(f)
}

// Validate checks if the format is valid.
func (f ReportFormat) Validate() error {
	switch f {
	case FormatMarkdown, FormatHTML, FormatJSON, FormatPDF:
		return nil
	default:
		return fmt.Errorf("invalid report format: %s", f)
	}
}

// FileExtension returns the file extension for this format.
func (f ReportFormat) FileExtension() string {
	switch f {
	case FormatMarkdown:
		return ".md"
	case FormatHTML:
		return ".html"
	case FormatJSON:
		return ".json"
	case FormatPDF:
		return ".pdf"
	default:
		return ".txt"
	}
}

// ReportConfig represents configuration for report generation.
type ReportConfig struct {
	// Format is the output format.
	Format ReportFormat

	// IncludeCharts enables chart generation.
	IncludeCharts bool

	// IncludeLogs includes log entries in the report.
	IncludeLogs bool

	// IncludeTimeSeries includes time series data in the report.
	IncludeTimeSeries bool

	// IncludeParameters includes task parameters.
	IncludeParameters bool

	// ChartWidth is the width for text-based charts (default: 60).
	ChartWidth int

	// ChartHeight is the height for text-based charts (default: 10).
	ChartHeight int

	// Title is the custom report title (optional).
	Title string

	// OutputPath is the file path for the report (optional).
	// If empty, a default path will be generated.
	OutputPath string
}

// DefaultConfig returns a default report configuration.
func DefaultConfig(format ReportFormat) *ReportConfig {
	return &ReportConfig{
		Format:            format,
		IncludeCharts:     true,
		IncludeLogs:       false,
		IncludeTimeSeries: true,
		IncludeParameters: true,
		ChartWidth:        60,
		ChartHeight:       10,
		Title:             "",
		OutputPath:        "",
	}
}

// Report represents a generated report.
type Report struct {
	// Format is the report format.
	Format ReportFormat

	// Content is the report content.
	Content []byte

	// GeneratedAt is when the report was generated.
	GeneratedAt time.Time

	// RunID is the associated run ID.
	RunID string

	// FilePath is the file path if saved to disk.
	FilePath string
}

// Generator is the interface for report generators.
type Generator interface {
	// Generate generates a report from the provided data.
	Generate(ctx *GenerateContext) (*Report, error)

	// Format returns the format this generator produces.
	Format() ReportFormat
}

// GenerateContext contains data for report generation.
type GenerateContext struct {
	// RunID is the benchmark run ID.
	RunID string

	// TaskID is the associated task ID.
	TaskID string

	// State is the run state.
	State string

	// CreatedAt is when the run was created.
	CreatedAt time.Time

	// StartedAt is when the run was started.
	StartedAt *time.Time

	// CompletedAt is when the run was completed.
	CompletedAt *time.Time

	// Duration is the run duration.
	Duration *time.Duration

	// ErrorMessage is the error message if the run failed.
	ErrorMessage string

	// ConnectionName is the database connection name.
	ConnectionName string

	// ConnectionType is the database type (mysql, oracle, sqlserver, postgresql).
	ConnectionType string

	// TemplateName is the benchmark template name.
	TemplateName string

	// Tool is the benchmark tool (sysbench, swingbench, hammerdb).
	Tool string

	// Parameters are the task parameters.
	Parameters map[string]interface{}

	// TPS is the transactions per second.
	TPS float64

	// LatencyAvg is the average latency in milliseconds.
	LatencyAvg float64

	// LatencyP95 is the 95th percentile latency in milliseconds.
	LatencyP95 float64

	// LatencyP99 is the 99th percentile latency in milliseconds.
	LatencyP99 float64

	// TotalTransactions is the total number of transactions.
	TotalTransactions int64

	// TotalQueries is the total number of queries.
	TotalQueries int64

	// ErrorCount is the total number of errors.
	ErrorCount int64

	// ErrorRate is the error rate percentage.
	ErrorRate float64

	// Samples is the time series metric samples.
	Samples []MetricSample

	// Logs are the log entries.
	Logs []LogEntry

	// Config is the report configuration.
	Config *ReportConfig

	// RawOutput is the raw command output.
	RawOutput string
}

// MetricSample represents a time series metric sample.
type MetricSample struct {
	Timestamp   time.Time
	TPS         float64
	LatencyAvg  float64
	LatencyP95  float64
	LatencyP99  float64
	ErrorRate   float64
}

// LogEntry represents a log entry.
type LogEntry struct {
	Timestamp string
	Stream    string // "stdout" or "stderr"
	Content   string
}

// NewGenerateContext creates a new generate context with minimal required fields.
func NewGenerateContext(runID string, config *ReportConfig) *GenerateContext {
	return &GenerateContext{
		RunID:  runID,
		Config: config,
		Samples: []MetricSample{},
		Logs:    []LogEntry{},
	}
}

// Validate validates the generate context.
func (ctx *GenerateContext) Validate() error {
	if ctx.RunID == "" {
		return fmt.Errorf("run_id is required")
	}
	if ctx.Config == nil {
		return fmt.Errorf("config is required")
	}
	if err := ctx.Config.Format.Validate(); err != nil {
		return err
	}
	return nil
}

// HasMetrics checks if the context has metric data.
func (ctx *GenerateContext) HasMetrics() bool {
	return ctx.TPS > 0 || ctx.LatencyAvg > 0
}

// HasSamples checks if the context has time series samples.
func (ctx *GenerateContext) HasSamples() bool {
	return len(ctx.Samples) > 0
}

// IsFailed checks if the run failed.
func (ctx *GenerateContext) IsFailed() bool {
	return ctx.ErrorMessage != ""
}

// GetDuration returns the formatted duration string.
func (ctx *GenerateContext) GetDuration() string {
	if ctx.Duration != nil {
		return ctx.Duration.String()
	}
	if ctx.StartedAt != nil && ctx.CompletedAt != nil {
		d := ctx.CompletedAt.Sub(*ctx.StartedAt)
		return d.String()
	}
	return "N/A"
}

// GetTimestamp returns the formatted timestamp for a time pointer.
func GetTimestamp(t *time.Time) string {
	if t == nil {
		return "N/A"
	}
	return t.Format(time.RFC3339)
}

// FormatFloat formats a float value with specified precision.
func FormatFloat(value float64, precision int) string {
	if value == 0 {
		return "N/A"
	}
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, value)
}
