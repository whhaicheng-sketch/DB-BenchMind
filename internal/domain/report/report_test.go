// Package report provides unit tests for report domain models.
package report

import (
	"testing"
	"time"
)

// TestReportFormat_Validate tests format validation.
func TestReportFormat_Validate(t *testing.T) {
	tests := []struct {
		name    string
		format  ReportFormat
		wantErr bool
	}{
		{
			name:    "valid markdown",
			format:  FormatMarkdown,
			wantErr: false,
		},
		{
			name:    "valid html",
			format:  FormatHTML,
			wantErr: false,
		},
		{
			name:    "valid json",
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:    "valid pdf",
			format:  FormatPDF,
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  ReportFormat("unknown"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.format.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ReportFormat.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReportFormat_FileExtension tests file extension.
func TestReportFormat_FileExtension(t *testing.T) {
	tests := []struct {
		name string
		format ReportFormat
		want string
	}{
		{"markdown", FormatMarkdown, ".md"},
		{"html", FormatHTML, ".html"},
		{"json", FormatJSON, ".json"},
		{"pdf", FormatPDF, ".pdf"},
		{"unknown", ReportFormat("unknown"), ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.FileExtension(); got != tt.want {
				t.Errorf("ReportFormat.FileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestReportConfig_DefaultConfig tests default configuration.
func TestReportConfig_DefaultConfig(t *testing.T) {
	config := DefaultConfig(FormatMarkdown)

	if config.Format != FormatMarkdown {
		t.Errorf("Format = %v, want %v", config.Format, FormatMarkdown)
	}
	if !config.IncludeCharts {
		t.Error("IncludeCharts should be true by default")
	}
	if config.IncludeLogs {
		t.Error("IncludeLogs should be false by default")
	}
	if !config.IncludeTimeSeries {
		t.Error("IncludeTimeSeries should be true by default")
	}
	if !config.IncludeParameters {
		t.Error("IncludeParameters should be true by default")
	}
	if config.ChartWidth != 60 {
		t.Errorf("ChartWidth = %d, want 60", config.ChartWidth)
	}
	if config.ChartHeight != 10 {
		t.Errorf("ChartHeight = %d, want 10", config.ChartHeight)
	}
}

// TestGenerateContext_Validate tests context validation.
func TestGenerateContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ctx     *GenerateContext
		wantErr bool
	}{
		{
			name: "valid context",
			ctx: &GenerateContext{
				RunID:  "run-1",
				Config: DefaultConfig(FormatMarkdown),
			},
			wantErr: false,
		},
		{
			name: "missing run id",
			ctx: &GenerateContext{
				Config: DefaultConfig(FormatMarkdown),
			},
			wantErr: true,
		},
		{
			name: "missing config",
			ctx: &GenerateContext{
				RunID: "run-1",
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			ctx: &GenerateContext{
				RunID: "run-1",
				Config: &ReportConfig{
					Format: ReportFormat("invalid"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ctx.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("GenerateContext.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateContext_Helpers tests helper methods.
func TestGenerateContext_Helpers(t *testing.T) {
	now := time.Now()
	duration := 5 * time.Minute

	ctx := &GenerateContext{
		RunID:      "run-1",
		Config:     DefaultConfig(FormatMarkdown),
		TPS:        1000.5,
		LatencyAvg: 5.25,
		ErrorMessage: "test error",
		Duration:   &duration,
		Samples:    []MetricSample{{TPS: 100}},
		Logs:       []LogEntry{{Content: "test"}},
	}

	// Test HasMetrics
	if !ctx.HasMetrics() {
		t.Error("HasMetrics() should return true when TPS > 0")
	}

	ctx.TPS = 0
	ctx.LatencyAvg = 0
	if ctx.HasMetrics() {
		t.Error("HasMetrics() should return false when no metrics")
	}

	// Test HasSamples
	if !ctx.HasSamples() {
		t.Error("HasSamples() should return true when samples exist")
	}

	// Test IsFailed
	if !ctx.IsFailed() {
		t.Error("IsFailed() should return true when ErrorMessage is set")
	}

	ctx.ErrorMessage = ""
	if ctx.IsFailed() {
		t.Error("IsFailed() should return false when ErrorMessage is empty")
	}

	// Test GetDuration
	got := ctx.GetDuration()
	if got != "5m0s" {
		t.Errorf("GetDuration() = %v, want 5m0s", got)
	}

	// Test GetTimestamp with nil
	if GetTimestamp(nil) != "N/A" {
		t.Error("GetTimestamp(nil) should return N/A")
	}

	// Test GetTimestamp with value
	if GetTimestamp(&now) == "N/A" {
		t.Error("GetTimestamp(&now) should not return N/A")
	}
}

// TestFormatFloat tests float formatting.
func TestFormatFloat(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		precision int
		want      string
	}{
		{"zero value", 0, 2, "N/A"},
		{"two decimals", 123.456, 2, "123.46"},
		{"one decimal", 123.456, 1, "123.5"},
		{"no decimals", 123.456, 0, "123"},
		{"small value", 0.001, 4, "0.0010"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFloat(tt.value, tt.precision); got != tt.want {
				t.Errorf("FormatFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewGenerateContext tests context creation.
func TestNewGenerateContext(t *testing.T) {
	config := DefaultConfig(FormatHTML)
	ctx := NewGenerateContext("run-123", config)

	if ctx.RunID != "run-123" {
		t.Errorf("RunID = %v, want run-123", ctx.RunID)
	}
	if ctx.Config != config {
		t.Error("Config not set correctly")
	}
	if len(ctx.Samples) != 0 {
		t.Error("Samples should be initialized as empty slice")
	}
	if len(ctx.Logs) != 0 {
		t.Error("Logs should be initialized as empty slice")
	}
}
