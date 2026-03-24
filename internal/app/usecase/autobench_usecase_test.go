package usecase

import (
	"context"
	"errors"
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
