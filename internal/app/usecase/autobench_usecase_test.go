package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
)

func TestAutoBenchSuiteUseCase_CreateSuiteUsesStageOneDefaults(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	suite, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:          "stage-one-defaults",
		ConnectionIDs: []string{"conn-a"},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	if suite.ExecutionPolicy.Mode != domainautobench.ExecutionModeSerial {
		t.Fatalf("ExecutionPolicy.Mode = %q, want %q", suite.ExecutionPolicy.Mode, domainautobench.ExecutionModeSerial)
	}
	if suite.ExecutionPolicy.FailurePolicy != domainautobench.FailurePolicyContinueByConnection {
		t.Fatalf("ExecutionPolicy.FailurePolicy = %q, want %q", suite.ExecutionPolicy.FailurePolicy, domainautobench.FailurePolicyContinueByConnection)
	}
	if !suite.ExecutionPolicy.CleanupEnabled {
		t.Fatal("ExecutionPolicy.CleanupEnabled = false, want true")
	}
	expectedProfiles := []domainautobench.ProfileType{
		domainautobench.ProfileTest,
		domainautobench.ProfileCPU,
		domainautobench.ProfileIO,
	}
	assertProfileOrder(t, suite.ExecutionPolicy.ProfileOrder, expectedProfiles)
}

func TestAutoBenchSuiteUseCase_CreateSuiteGeneratesItemsInProfileOrderPerConnection(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	suite, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:          "ordered-items",
		ConnectionIDs: []string{"conn-a", "conn-b"},
		Profiles: []domainautobench.ProfileType{
			domainautobench.ProfileIO,
			domainautobench.ProfileTest,
		},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	if len(suite.Items) != 4 {
		t.Fatalf("len(Items) = %d, want 4", len(suite.Items))
	}

	expected := []struct {
		connectionID string
		profile      domainautobench.ProfileType
	}{
		{"conn-a", domainautobench.ProfileTest},
		{"conn-a", domainautobench.ProfileIO},
		{"conn-b", domainautobench.ProfileTest},
		{"conn-b", domainautobench.ProfileIO},
	}
	for i, want := range expected {
		item := suite.Items[i]
		if item.ConnectionID != want.connectionID {
			t.Fatalf("Items[%d].ConnectionID = %q, want %q", i, item.ConnectionID, want.connectionID)
		}
		if item.ProfileType != want.profile {
			t.Fatalf("Items[%d].ProfileType = %q, want %q", i, item.ProfileType, want.profile)
		}
		if item.Status != domainautobench.SuiteItemStatusPending {
			t.Fatalf("Items[%d].Status = %q, want %q", i, item.Status, domainautobench.SuiteItemStatusPending)
		}
	}
}

func TestAutoBenchSuiteUseCase_BuildExecutionPlanIsSequentialAndDoesNotExecute(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	suite, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:          "plan-only",
		ConnectionIDs: []string{"conn-a"},
		Profiles: []domainautobench.ProfileType{
			domainautobench.ProfileTest,
			domainautobench.ProfileCPU,
		},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	plan, err := uc.BuildExecutionPlan(ctx, suite.ID)
	if err != nil {
		t.Fatalf("BuildExecutionPlan() failed: %v", err)
	}

	if !plan.Sequential {
		t.Fatal("Sequential = false, want true")
	}
	if plan.Mode != domainautobench.ExecutionModeSerial {
		t.Fatalf("Mode = %q, want %q", plan.Mode, domainautobench.ExecutionModeSerial)
	}
	if plan.SuiteID != suite.ID {
		t.Fatalf("SuiteID = %q, want %q", plan.SuiteID, suite.ID)
	}
	if len(plan.Items) != len(suite.Items) {
		t.Fatalf("len(Plan.Items) = %d, want %d", len(plan.Items), len(suite.Items))
	}
	for i, item := range plan.Items {
		if item.Sequence != i+1 {
			t.Fatalf("Plan.Items[%d].Sequence = %d, want %d", i, item.Sequence, i+1)
		}
		if item.ConnectionID != suite.Items[i].ConnectionID {
			t.Fatalf("Plan.Items[%d].ConnectionID = %q, want %q", i, item.ConnectionID, suite.Items[i].ConnectionID)
		}
		if item.ProfileType != suite.Items[i].ProfileType {
			t.Fatalf("Plan.Items[%d].ProfileType = %q, want %q", i, item.ProfileType, suite.Items[i].ProfileType)
		}
	}

	status, err := uc.GetSuiteStatus(ctx, suite.ID)
	if err != nil {
		t.Fatalf("GetSuiteStatus() failed: %v", err)
	}
	if status.Status != domainautobench.SuiteStatusDraft {
		t.Fatalf("Status after planning = %q, want %q", status.Status, domainautobench.SuiteStatusDraft)
	}
}

func TestAutoBenchSuiteUseCase_GetSuiteStatusReturnsStableSnapshot(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	suite, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:          "status-snapshot",
		ConnectionIDs: []string{"conn-a", "conn-b"},
		Profiles:      []domainautobench.ProfileType{domainautobench.ProfileTest},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	status, err := uc.GetSuiteStatus(ctx, suite.ID)
	if err != nil {
		t.Fatalf("GetSuiteStatus() failed: %v", err)
	}

	if status.SuiteID != suite.ID {
		t.Fatalf("SuiteID = %q, want %q", status.SuiteID, suite.ID)
	}
	if status.Name != suite.Name {
		t.Fatalf("Name = %q, want %q", status.Name, suite.Name)
	}
	if status.TotalItems != 2 {
		t.Fatalf("TotalItems = %d, want 2", status.TotalItems)
	}
	if status.PendingItems != 2 {
		t.Fatalf("PendingItems = %d, want 2", status.PendingItems)
	}
	if len(status.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(status.Items))
	}
	if len(status.SelectedProfiles) != 1 || status.SelectedProfiles[0] != domainautobench.ProfileTest {
		t.Fatalf("SelectedProfiles = %#v, want [test]", status.SelectedProfiles)
	}
}

func TestAutoBenchSuiteUseCase_CreateSuiteRequiresAtLeastOneConnection(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	_, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:          "missing-connections",
		ConnectionIDs: []string{" ", ""},
	})
	if !errors.Is(err, ErrAutoBenchConnectionRequired) {
		t.Fatalf("CreateSuite() error = %v, want %v", err, ErrAutoBenchConnectionRequired)
	}
}

func TestAutoBenchSuiteUseCase_ListSupportedProfilesReturnsStableCopy(t *testing.T) {
	uc := NewAutoBenchSuiteUseCase()

	first := uc.ListSupportedProfiles()
	assertProfileOrder(t, first, []domainautobench.ProfileType{
		domainautobench.ProfileTest,
		domainautobench.ProfileCPU,
		domainautobench.ProfileIO,
	})

	first[0] = domainautobench.ProfileIO
	second := uc.ListSupportedProfiles()
	assertProfileOrder(t, second, []domainautobench.ProfileType{
		domainautobench.ProfileTest,
		domainautobench.ProfileCPU,
		domainautobench.ProfileIO,
	})
}

func TestAutoBenchSuiteUseCase_CreateSuiteRespectsCleanupOverrideAndReturnsDetachedSnapshots(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()
	disabled := false

	suite, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:           "cleanup-override",
		ConnectionIDs:  []string{"conn-a"},
		CleanupEnabled: &disabled,
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}
	if suite.ExecutionPolicy.CleanupEnabled {
		t.Fatal("CleanupEnabled = true, want false")
	}

	status, err := uc.GetSuiteStatus(ctx, suite.ID)
	if err != nil {
		t.Fatalf("GetSuiteStatus() failed: %v", err)
	}
	status.Items[0].Status = domainautobench.SuiteItemStatusSuccess
	status.ExecutionPolicy.ProfileOrder[0] = domainautobench.ProfileIO

	latest, err := uc.GetSuiteStatus(ctx, suite.ID)
	if err != nil {
		t.Fatalf("GetSuiteStatus() failed: %v", err)
	}
	if latest.Items[0].Status != domainautobench.SuiteItemStatusPending {
		t.Fatalf("stored item status = %q, want %q", latest.Items[0].Status, domainautobench.SuiteItemStatusPending)
	}
	if latest.ExecutionPolicy.ProfileOrder[0] != domainautobench.ProfileTest {
		t.Fatalf("stored profile order[0] = %q, want %q", latest.ExecutionPolicy.ProfileOrder[0], domainautobench.ProfileTest)
	}
}

func TestAutoBenchSuiteUseCase_GetSuiteStatusReturnsNotFoundForUnknownSuite(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	_, err := uc.GetSuiteStatus(ctx, "missing-suite")
	if !errors.Is(err, ErrAutoBenchSuiteNotFound) {
		t.Fatalf("GetSuiteStatus() error = %v, want %v", err, ErrAutoBenchSuiteNotFound)
	}
}

func TestAutoBenchSuiteUseCase_BuildSuiteReportJSONAggregatesCurrentSuiteSnapshot(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	suite, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:          "report-json",
		ConnectionIDs: []string{"conn-a", "conn-b"},
		Profiles: []domainautobench.ProfileType{
			domainautobench.ProfileTest,
			domainautobench.ProfileCPU,
		},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	if err := uc.mutateSuite(suite.ID, func(suite *domainautobench.Suite) error {
		suite.Status = domainautobench.SuiteStatusPartialSuccess
		suite.ReportJSONPath = "reports/report-json.json"
		suite.ReportPath = "reports/report-json.html"

		suite.Items[0].DatabaseType = "mysql"
		suite.Items[0].TemplateID = "mysql_test"
		suite.Items[0].Status = domainautobench.SuiteItemStatusSuccess
		suite.Items[0].PhaseStatus = "completed"
		suite.Items[0].LinkedTaskID = "task-1"

		suite.Items[1].DatabaseType = "mysql"
		suite.Items[1].TemplateID = "mysql_cpu"
		suite.Items[1].Status = domainautobench.SuiteItemStatusFailed
		suite.Items[1].PhaseStatus = "failed"
		suite.Items[1].LinkedTaskID = "task-2"
		suite.Items[1].ErrorSummary = "benchmark run ended with state failed"

		suite.Items[2].DatabaseType = "postgresql"
		suite.Items[2].TemplateID = "pg_test"
		suite.Items[2].Status = domainautobench.SuiteItemStatusSkipped
		suite.Items[2].ErrorSummary = "skipped after earlier connection failure"

		suite.Items[3].DatabaseType = "postgresql"
		suite.Items[3].TemplateID = "pg_cpu"
		suite.Items[3].Status = domainautobench.SuiteItemStatusSkipped
		suite.Items[3].ErrorSummary = "skipped after earlier connection failure"
		return nil
	}); err != nil {
		t.Fatalf("mutateSuite() failed: %v", err)
	}

	report, err := uc.BuildSuiteReport(ctx, suite.ID)
	if err != nil {
		t.Fatalf("BuildSuiteReport() failed: %v", err)
	}

	if report.SuiteID != suite.ID {
		t.Fatalf("SuiteID = %q, want %q", report.SuiteID, suite.ID)
	}
	if report.Summary.Status != domainautobench.SuiteStatusPartialSuccess {
		t.Fatalf("Summary.Status = %q, want %q", report.Summary.Status, domainautobench.SuiteStatusPartialSuccess)
	}
	if report.Summary.TotalItems != 4 {
		t.Fatalf("Summary.TotalItems = %d, want 4", report.Summary.TotalItems)
	}
	if report.Summary.SuccessItemCount != 1 {
		t.Fatalf("Summary.SuccessItemCount = %d, want 1", report.Summary.SuccessItemCount)
	}
	if report.Summary.FailedItemCount != 1 {
		t.Fatalf("Summary.FailedItemCount = %d, want 1", report.Summary.FailedItemCount)
	}
	if report.Summary.SkippedItemCount != 2 {
		t.Fatalf("Summary.SkippedItemCount = %d, want 2", report.Summary.SkippedItemCount)
	}
	if len(report.ConnectionRows) != 2 {
		t.Fatalf("len(ConnectionRows) = %d, want 2", len(report.ConnectionRows))
	}
	if len(report.Failures) != 3 {
		t.Fatalf("len(Failures) = %d, want 3", len(report.Failures))
	}
	if report.ArtifactPaths.JSON != "reports/report-json.json" {
		t.Fatalf("ArtifactPaths.JSON = %q, want reports/report-json.json", report.ArtifactPaths.JSON)
	}
	if report.ArtifactPaths.HTML != "reports/report-json.html" {
		t.Fatalf("ArtifactPaths.HTML = %q, want reports/report-json.html", report.ArtifactPaths.HTML)
	}

	data, err := uc.BuildSuiteReportJSON(ctx, suite.ID)
	if err != nil {
		t.Fatalf("BuildSuiteReportJSON() failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded["suite_id"] != suite.ID {
		t.Fatalf("suite_id = %#v, want %q", decoded["suite_id"], suite.ID)
	}
	summary, ok := decoded["summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("summary = %#v, want object", decoded["summary"])
	}
	if summary["status"] != string(domainautobench.SuiteStatusPartialSuccess) {
		t.Fatalf("summary.status = %#v, want %q", summary["status"], domainautobench.SuiteStatusPartialSuccess)
	}
	artifactPaths, ok := decoded["artifact_paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("artifact_paths = %#v, want object", decoded["artifact_paths"])
	}
	if artifactPaths["json"] != "reports/report-json.json" {
		t.Fatalf("artifact_paths.json = %#v, want reports/report-json.json", artifactPaths["json"])
	}
}

func TestAutoBenchSuiteUseCase_BuildSuiteReportHTMLRendersSectionsFromSuiteSnapshot(t *testing.T) {
	ctx := context.Background()
	uc := NewAutoBenchSuiteUseCase()

	suite, err := uc.CreateSuite(ctx, CreateSuiteInput{
		Name:          "html-report",
		ConnectionIDs: []string{"conn-a", "conn-b"},
		Profiles: []domainautobench.ProfileType{
			domainautobench.ProfileTest,
			domainautobench.ProfileCPU,
		},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	if err := uc.mutateSuite(suite.ID, func(suite *domainautobench.Suite) error {
		suite.Status = domainautobench.SuiteStatusPartialSuccess
		suite.ReportPath = "reports/html-report.html"
		suite.ReportJSONPath = "reports/html-report.json"

		suite.Items[0].DatabaseType = "mysql"
		suite.Items[0].Status = domainautobench.SuiteItemStatusSuccess
		suite.Items[0].TemplateID = "mysql_test"

		suite.Items[1].DatabaseType = "mysql"
		suite.Items[1].Status = domainautobench.SuiteItemStatusFailed
		suite.Items[1].TemplateID = "mysql_cpu"
		suite.Items[1].LinkedTaskID = "task-2"
		suite.Items[1].ErrorSummary = "benchmark run ended with state failed"

		suite.Items[2].DatabaseType = "postgresql"
		suite.Items[2].Status = domainautobench.SuiteItemStatusSuccess
		suite.Items[2].TemplateID = "pg_test"

		suite.Items[3].DatabaseType = "postgresql"
		suite.Items[3].Status = domainautobench.SuiteItemStatusSkipped
		suite.Items[3].TemplateID = "pg_cpu"
		suite.Items[3].ErrorSummary = "skipped after earlier connection failure"
		return nil
	}); err != nil {
		t.Fatalf("mutateSuite() failed: %v", err)
	}

	htmlData, err := uc.BuildSuiteReportHTML(ctx, suite.ID)
	if err != nil {
		t.Fatalf("BuildSuiteReportHTML() failed: %v", err)
	}

	html := string(htmlData)
	assertContains(t, html, "<!DOCTYPE html>")
	assertContains(t, html, "<title>AutoBench Suite Report")
	assertContains(t, html, "Executive Summary")
	assertContains(t, html, "Connection Summary")
	assertContains(t, html, "Failure Analysis")
	assertContains(t, html, "Recommendations")
	assertContains(t, html, suite.ID)
	assertContains(t, html, "partial_success")
	assertContains(t, html, "conn-a")
	assertContains(t, html, "conn-b")
	assertContains(t, html, "benchmark run ended with state failed")
	assertContains(t, html, "skipped after earlier connection failure")
	assertContains(t, html, "reports/html-report.json")
}

func assertProfileOrder(t *testing.T, got, want []domainautobench.ProfileType) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(ProfileOrder) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ProfileOrder[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected %q to contain %q", haystack, needle)
	}
}
