package autobench

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewSuiteUsesStageOneDefaults(t *testing.T) {
	suite := NewSuite("nightly-suite", []string{"conn-a", "conn-b"}, []ProfileType{ProfileTest, ProfileCPU, ProfileIO})

	if suite.Name != "nightly-suite" {
		t.Fatalf("Name = %q, want nightly-suite", suite.Name)
	}
	if suite.Status != SuiteStatusDraft {
		t.Fatalf("Status = %q, want %q", suite.Status, SuiteStatusDraft)
	}
	if suite.ExecutionPolicy.Mode != ExecutionModeSerial {
		t.Fatalf("ExecutionPolicy.Mode = %q, want %q", suite.ExecutionPolicy.Mode, ExecutionModeSerial)
	}
	if suite.ExecutionPolicy.FailurePolicy != FailurePolicyContinueByConnection {
		t.Fatalf("ExecutionPolicy.FailurePolicy = %q, want %q", suite.ExecutionPolicy.FailurePolicy, FailurePolicyContinueByConnection)
	}
	if !suite.ExecutionPolicy.CleanupEnabled {
		t.Fatal("ExecutionPolicy.CleanupEnabled = false, want true")
	}
	expectedOrder := []ProfileType{ProfileTest, ProfileCPU, ProfileIO}
	if len(suite.ExecutionPolicy.ProfileOrder) != len(expectedOrder) {
		t.Fatalf("ProfileOrder len = %d, want %d", len(suite.ExecutionPolicy.ProfileOrder), len(expectedOrder))
	}
	for i, profile := range expectedOrder {
		if suite.ExecutionPolicy.ProfileOrder[i] != profile {
			t.Fatalf("ProfileOrder[%d] = %q, want %q", i, suite.ExecutionPolicy.ProfileOrder[i], profile)
		}
	}
}

func TestSuiteItemLinkingAndReportArtifacts(t *testing.T) {
	item := SuiteItem{
		ID:           "item-1",
		SuiteID:      "suite-1",
		ConnectionID: "conn-1",
		ProfileType:  ProfileCPU,
		Status:       SuiteItemStatusPending,
		LinkedTaskID: "task-1",
	}

	if item.Status != SuiteItemStatusPending {
		t.Fatalf("Status = %q, want %q", item.Status, SuiteItemStatusPending)
	}
	if item.LinkedTaskID != "task-1" {
		t.Fatalf("LinkedTaskID = %q, want task-1", item.LinkedTaskID)
	}
}

func TestSuiteItemReportIDField(t *testing.T) {
	item := SuiteItem{
		ID:           "item-1",
		SuiteID:      "suite-1",
		ConnectionID: "conn-1",
		ProfileType:  ProfileCPU,
		Status:       SuiteItemStatusSuccess,
		ReportID:     "report-abc123",
	}

	if item.ReportID != "report-abc123" {
		t.Fatalf("ReportID = %q, want report-abc123", item.ReportID)
	}

	// Verify JSON serialization includes report_id
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded["report_id"] != "report-abc123" {
		t.Fatalf("report_id = %#v, want report-abc123", decoded["report_id"])
	}

	// Verify empty ReportID omits field (omitempty)
	emptyItem := SuiteItem{ID: "item-2"}
	emptyData, err := json.Marshal(emptyItem)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var emptyDecoded map[string]interface{}
	if err := json.Unmarshal(emptyData, &emptyDecoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, exists := emptyDecoded["report_id"]; exists {
		t.Fatal("report_id should be omitted when empty")
	}
}

func TestSuiteReportToJSONUsesTypedReportShape(t *testing.T) {
	generatedAt := time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	report := SuiteReport{
		SuiteID:     "suite-1",
		GeneratedAt: generatedAt,
		Summary: SuiteReportSummary{
			Status:           SuiteStatusPartialSuccess,
			TotalItems:       3,
			SuccessItemCount: 1,
			FailedItemCount:  1,
			SkippedItemCount: 1,
		},
		ConnectionRows: []SuiteReportConnectionRow{
			{
				ConnectionID:       "conn-1",
				DatabaseType:       "mysql",
				Status:             SuiteStatusPartialSuccess,
				ProfileTypes:       []ProfileType{ProfileTest, ProfileCPU},
				SuccessItemCount:   1,
				FailedItemCount:    1,
				SkippedItemCount:   0,
				CompletedItemCount: 2,
			},
		},
		Failures: []SuiteReportFailure{
			{
				SuiteItemID:  "item-2",
				ConnectionID: "conn-1",
				ProfileType:  ProfileCPU,
				ErrorSummary: "benchmark run ended with state failed",
				LinkedTaskID: "task-2",
			},
		},
		Recommendations: []string{"Review failed profiles before enabling HTML export."},
		ArtifactPaths: SuiteReportArtifactPaths{
			JSON: "reports/suite-1.json",
			HTML: "reports/suite-1.html",
		},
	}

	data, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded["suite_id"] != "suite-1" {
		t.Fatalf("suite_id = %#v, want suite-1", decoded["suite_id"])
	}
	if _, ok := decoded["generated_at"].(string); !ok {
		t.Fatalf("generated_at = %#v, want string timestamp", decoded["generated_at"])
	}
	artifactPaths, ok := decoded["artifact_paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("artifact_paths = %#v, want object", decoded["artifact_paths"])
	}
	if artifactPaths["json"] != "reports/suite-1.json" {
		t.Fatalf("artifact_paths.json = %#v, want reports/suite-1.json", artifactPaths["json"])
	}
	connectionRows, ok := decoded["connection_rows"].([]interface{})
	if !ok || len(connectionRows) != 1 {
		t.Fatalf("connection_rows = %#v, want len 1", decoded["connection_rows"])
	}
	failures, ok := decoded["failures"].([]interface{})
	if !ok || len(failures) != 1 {
		t.Fatalf("failures = %#v, want len 1", decoded["failures"])
	}
}
