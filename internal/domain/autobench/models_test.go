package autobench

import "testing"

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

	report := SuiteReport{
		SuiteID:        "suite-1",
		GeneratedFiles: []string{"reports/suite-1.html", "reports/suite-1.json"},
	}

	if len(report.GeneratedFiles) != 2 {
		t.Fatalf("GeneratedFiles len = %d, want 2", len(report.GeneratedFiles))
	}
	if report.GeneratedFiles[0] != "reports/suite-1.html" {
		t.Fatalf("GeneratedFiles[0] = %q, want reports/suite-1.html", report.GeneratedFiles[0])
	}
}
