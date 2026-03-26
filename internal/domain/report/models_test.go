// Package report provides report domain models.
package report

import (
	"testing"
	"time"
)

func TestNewReport(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		report  Report
		checkFn func(t *testing.T, r Report)
	}{
		{
			name: "basic report fields",
			report: Report{
				ID:           "test-id",
				SuiteID:      "standalone",
				SourceType:   SourceTypeBenchmark,
				ConnectionID: "conn-1",
				DatabaseType: "mysql",
				Status:       StatusCompleted,
				StartedAt:    now,
			},
			checkFn: func(t *testing.T, r Report) {
				if r.ID != "test-id" {
					t.Errorf("expected ID test-id, got %s", r.ID)
				}
				if r.SuiteID != "standalone" {
					t.Errorf("expected SuiteID standalone, got %s", r.SuiteID)
				}
				if r.SourceType != SourceTypeBenchmark {
					t.Errorf("expected SourceTypeBenchmark, got %s", r.SourceType)
				}
				if r.ConnectionID != "conn-1" {
					t.Errorf("expected ConnectionID conn-1, got %s", r.ConnectionID)
				}
				if r.DatabaseType != "mysql" {
					t.Errorf("expected DatabaseType mysql, got %s", r.DatabaseType)
				}
				if r.Status != StatusCompleted {
					t.Errorf("expected StatusCompleted, got %s", r.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFn(t, tt.report)
		})
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
