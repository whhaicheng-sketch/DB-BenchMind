// Package report provides unit tests for markdown generator.
package report

import (
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// TestMarkdownGenerator_Format tests format detection.
func TestMarkdownGenerator_Format(t *testing.T) {
	gen := NewMarkdownGenerator()
	if gen.Format() != report.FormatMarkdown {
		t.Errorf("Format() = %v, want %v", gen.Format(), report.FormatMarkdown)
	}
}

// TestMarkdownGenerator_Generate tests report generation.
func TestMarkdownGenerator_Generate(t *testing.T) {
	gen := NewMarkdownGenerator()

	now := time.Now()
	duration := 5 * time.Minute

	data := &report.GenerateContext{
		RunID:          "test-run-1",
		Config:         report.DefaultConfig(report.FormatMarkdown),
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
			{
				Timestamp:  now.Add(time.Second),
				TPS:        1100,
				LatencyAvg: 5.5,
				LatencyP95: 11.0,
				LatencyP99: 22.0,
				ErrorRate:  0.0,
			},
		},
		Logs: []report.LogEntry{
			{
				Timestamp: now.Format(time.RFC3339),
				Stream:    "stdout",
				Content:   "Starting benchmark...",
			},
		},
		RawOutput: "SQL statistics:\n    queries performed:\n        read: 1000\n",
	}

	rpt, err := gen.Generate(data)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify report
	if rpt.Format != report.FormatMarkdown {
		t.Errorf("Format = %v, want %v", rpt.Format, report.FormatMarkdown)
	}
	if len(rpt.Content) == 0 {
		t.Error("Content should not be empty")
	}
	if rpt.RunID != "test-run-1" {
		t.Errorf("RunID = %v, want test-run-1", rpt.RunID)
	}

	// Verify content contains key sections
	content := string(rpt.Content)
	if !contains(content, "# Benchmark Report") {
		t.Error("Content should contain title header")
	}
	if !contains(content, "## Summary") {
		t.Error("Content should contain summary section")
	}
	if !contains(content, "## Metrics") {
		t.Error("Content should contain metrics section")
	}
	if !contains(content, "1234.56") {
		t.Error("Content should contain TPS value")
	}
	if !contains(content, "5.25") {
		t.Error("Content should contain latency value")
	}
}

// TestMarkdownGenerator_GenerateFailedRun tests report generation for failed run.
func TestMarkdownGenerator_GenerateFailedRun(t *testing.T) {
	gen := NewMarkdownGenerator()

	now := time.Now()

	data := &report.GenerateContext{
		RunID:          "failed-run-1",
		Config:         report.DefaultConfig(report.FormatMarkdown),
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
	if !contains(content, "❌ Failed") {
		t.Error("Content should indicate failed status")
	}
	if !contains(content, "Connection failed") {
		t.Error("Content should contain error message")
	}
}

// TestMarkdownGenerator_Validation tests validation.
func TestMarkdownGenerator_Validation(t *testing.T) {
	gen := NewMarkdownGenerator()

	// Missing run ID
	data := &report.GenerateContext{
		Config: report.DefaultConfig(report.FormatMarkdown),
	}

	_, err := gen.Generate(data)
	if err == nil {
		t.Error("Generate() should fail with missing run ID")
	}

	// Invalid format
	data2 := &report.GenerateContext{
		RunID:  "test",
		Config: &report.ReportConfig{Format: report.ReportFormat("invalid")},
	}

	_, err = gen.Generate(data2)
	if err == nil {
		t.Error("Generate() should fail with invalid format")
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
