// Package report provides unit tests for JSON generator.
package report

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// TestJSONGenerator_Format tests format detection.
func TestJSONGenerator_Format(t *testing.T) {
	gen := NewJSONGenerator()
	if gen.Format() != report.FormatJSON {
		t.Errorf("Format() = %v, want %v", gen.Format(), report.FormatJSON)
	}
}

// TestJSONGenerator_Generate tests report generation.
func TestJSONGenerator_Generate(t *testing.T) {
	gen := NewJSONGenerator()

	now := time.Now()
	duration := 5 * time.Minute

	data := &report.GenerateContext{
		RunID:          "test-run-1",
		Config:         report.DefaultConfig(report.FormatJSON),
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
	if rpt.Format != report.FormatJSON {
		t.Errorf("Format = %v, want %v", rpt.Format, report.FormatJSON)
	}
	if len(rpt.Content) == 0 {
		t.Error("Content should not be empty")
	}

	// Parse JSON to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(rpt.Content, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify top-level keys
	if _, ok := result["meta"]; !ok {
		t.Error("JSON should contain 'meta' key")
	}
	if _, ok := result["summary"]; !ok {
		t.Error("JSON should contain 'summary' key")
	}
	if _, ok := result["metrics"]; !ok {
		t.Error("JSON should contain 'metrics' key")
	}
	if _, ok := result["time_series"]; !ok {
		t.Error("JSON should contain 'time_series' key")
	}

	// Verify meta
	meta := result["meta"].(map[string]interface{})
	if meta["run_id"] != "test-run-1" {
		t.Errorf("run_id = %v, want test-run-1", meta["run_id"])
	}
	if meta["format"] != "json" {
		t.Errorf("format = %v, want json", meta["format"])
	}

	// Verify summary
	summary := result["summary"].(map[string]interface{})
	if summary["status"] != "completed" {
		t.Errorf("status = %v, want completed", summary["status"])
	}
	if summary["tool"] != "sysbench" {
		t.Errorf("tool = %v, want sysbench", summary["tool"])
	}

	// Verify metrics
	metrics := result["metrics"].(map[string]interface{})
	if metrics["tps"].(float64) != 1234.56 {
		t.Errorf("tps = %v, want 1234.56", metrics["tps"])
	}
}

// TestJSONGenerator_GenerateFailedRun tests report generation for failed run.
func TestJSONGenerator_GenerateFailedRun(t *testing.T) {
	gen := NewJSONGenerator()

	data := &report.GenerateContext{
		RunID:          "failed-run-1",
		Config:         report.DefaultConfig(report.FormatJSON),
		State:          "failed",
		CreatedAt:      time.Now(),
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

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal(rpt.Content, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify failed status
	summary := result["summary"].(map[string]interface{})
	if summary["status"] != "failed" {
		t.Errorf("status = %v, want failed", summary["status"])
	}
	if summary["error"] != "Connection failed" {
		t.Errorf("error = %v, want 'Connection failed'", summary["error"])
	}
}

// TestJSONGenerator_Validation tests validation.
func TestJSONGenerator_Validation(t *testing.T) {
	gen := NewJSONGenerator()

	// Missing run ID
	data := &report.GenerateContext{
		Config: report.DefaultConfig(report.FormatJSON),
	}

	_, err := gen.Generate(data)
	if err == nil {
		t.Error("Generate() should fail with missing run ID")
	}
}
