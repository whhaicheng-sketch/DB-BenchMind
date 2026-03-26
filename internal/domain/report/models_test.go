// Package report provides report domain models.
package report

import (
	"testing"
	"time"
)

func TestNewReport(t *testing.T) {
	now := time.Now()
	rpt := Report{
		ID:           "test-id",
		SuiteID:      "standalone",
		SourceType:   SourceTypeBenchmark,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       StatusCompleted,
		StartedAt:    now,
	}

	if rpt.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", rpt.ID)
	}
	if rpt.SuiteID != "standalone" {
		t.Errorf("expected SuiteID standalone, got %s", rpt.SuiteID)
	}
	if rpt.SourceType != SourceTypeBenchmark {
		t.Errorf("expected SourceTypeBenchmark, got %s", rpt.SourceType)
	}
}

func TestReportIsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   ReportStatus
		expected bool
	}{
		{"completed", StatusCompleted, true},
		{"failed", StatusFailed, true},
		{"cancelled", StatusCancelled, true},
		{"running", StatusRunning, false},
		{"pending", StatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpt := Report{Status: tt.status}
			if got := rpt.IsCompleted(); got != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSuiteIsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   SuiteStatus
		expected bool
	}{
		{"success", SuiteStatusSuccess, true},
		{"failed", SuiteStatusFailed, true},
		{"partial_success", SuiteStatusPartialSuccess, true},
		{"cancelled", SuiteStatusCancelled, true},
		{"running", SuiteStatusRunning, false},
		{"draft", SuiteStatusDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := Suite{Status: tt.status}
			if got := suite.IsCompleted(); got != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}
