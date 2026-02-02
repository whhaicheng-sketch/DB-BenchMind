// Package report provides unit tests for HTML generator.
package report

import (
	"strings"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// TestHTMLGenerator_Format tests format detection.
func TestHTMLGenerator_Format(t *testing.T) {
	gen := NewHTMLGenerator()
	if gen.Format() != report.FormatHTML {
		t.Errorf("Format() = %v, want %v", gen.Format(), report.FormatHTML)
	}
}

// TestHTMLGenerator_Generate tests report generation.
func TestHTMLGenerator_Generate(t *testing.T) {
	gen := NewHTMLGenerator()

	now := time.Now()
	duration := 5 * time.Minute

	data := &report.GenerateContext{
		RunID:          "test-run-1",
		Config:         report.DefaultConfig(report.FormatHTML),
		TaskID:         "task-1",
		State:          "completed",
		CreatedAt:      now,
		StartedAt:      &now,
		CompletedAt:    &now,
		Duration:       &duration,
		Tool:           "sysbench",
		TemplateName:   "oltp-read-write",
		ConnectionName: "Test MySQL",
		ConnectionType: "mysql",
		Parameters: map[string]interface{}{
			"threads": 8,
			"time":    60,
		},
		TPS:               1234.56,
		LatencyAvg:        5.25,
		LatencyP95:        12.34,
		LatencyP99:        23.45,
		TotalTransactions: 10000,
		TotalQueries:      50000,
		ErrorCount:        0,
		ErrorRate:         0.0,
		Samples: []report.MetricSample{
			{
				Timestamp:  now,
				TPS:        1000,
				LatencyAvg: 5.0,
				LatencyP95: 10.0,
				LatencyP99: 20.0,
				ErrorRate:  0.0,
			},
		},
		RawOutput: "SQL statistics:",
	}

	rpt, err := gen.Generate(data)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify report
	if rpt.Format != report.FormatHTML {
		t.Errorf("Format = %v, want %v", rpt.Format, report.FormatHTML)
	}
	if len(rpt.Content) == 0 {
		t.Error("Content should not be empty")
	}

	// Verify content contains key HTML elements
	content := string(rpt.Content)
	if !strings.Contains(content, "<!DOCTYPE html>") {
		t.Error("Content should contain DOCTYPE declaration")
	}
	if !strings.Contains(content, "<html") {
		t.Error("Content should contain html tag")
	}
	if !strings.Contains(content, "1234.56") {
		t.Error("Content should contain TPS value")
	}
	if !strings.Contains(content, "metric-card") {
		t.Error("Content should contain metric cards")
	}
}

// TestHTMLGenerator_GenerateFailedRun tests report generation for failed run.
func TestHTMLGenerator_GenerateFailedRun(t *testing.T) {
	gen := NewHTMLGenerator()

	now := time.Now()

	data := &report.GenerateContext{
		RunID:          "failed-run-1",
		Config:         report.DefaultConfig(report.FormatHTML),
		State:          "failed",
		CreatedAt:      now,
		ErrorMessage:   "Connection failed",
		Tool:           "sysbench",
		TemplateName:   "oltp-read-write",
		ConnectionName: "Test MySQL",
		ConnectionType: "mysql",
	}

	rpt, err := gen.Generate(data)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	content := string(rpt.Content)
	if !strings.Contains(content, "status-failed") {
		t.Error("Content should contain failed status class")
	}
	if !strings.Contains(content, "Connection failed") {
		t.Error("Content should contain error message")
	}
}

// TestHTMLGenerator_Validation tests validation.
func TestHTMLGenerator_Validation(t *testing.T) {
	gen := NewHTMLGenerator()

	// Missing run ID
	data := &report.GenerateContext{
		Config: report.DefaultConfig(report.FormatHTML),
	}

	_, err := gen.Generate(data)
	if err == nil {
		t.Error("Generate() should fail with missing run ID")
	}
}
