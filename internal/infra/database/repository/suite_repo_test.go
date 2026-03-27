package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
)

func TestSuiteRepository_Save(t *testing.T) {
	db := setupSuiteTestDB(t)
	defer db.Close()

	repo := NewSQLiteSuiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	suite := domainautobench.Suite{
		ID:   "test-suite-1",
		Name: "Test Suite",
		ExecutionPolicy: domainautobench.ExecutionPolicy{
			Mode:           domainautobench.ExecutionModeSerial,
			FailurePolicy:  domainautobench.FailurePolicyContinueByConnection,
			CleanupEnabled: true,
		},
		Status:    domainautobench.SuiteStatusDraft,
		StartedAt: &now,
		Items: []domainautobench.SuiteItem{
			{ID: "item-1", ConnectionID: "conn-1", ProfileType: domainautobench.ProfileTest, Status: domainautobench.SuiteItemStatusPending},
		},
	}

	err := repo.Save(ctx, suite)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify suite was saved
	found, err := repo.FindByID(ctx, "test-suite-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.ID != suite.ID {
		t.Errorf("FindByID() ID = %v, want %v", found.ID, suite.ID)
	}
	if found.Name != suite.Name {
		t.Errorf("FindByID() Name = %v, want %v", found.Name, suite.Name)
	}
}

func TestSuiteRepository_FindByID_NotFound(t *testing.T) {
	db := setupSuiteTestDB(t)
	defer db.Close()

	repo := NewSQLiteSuiteRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "non-existent-id")
	if err == nil {
		t.Error("FindByID() expected error for non-existent suite")
	}
}

func TestSuiteRepository_FindAll(t *testing.T) {
	db := setupSuiteTestDB(t)
	defer db.Close()

	repo := NewSQLiteSuiteRepository(db)
	ctx := context.Background()

	// Create multiple suites
	for i := 1; i <= 3; i++ {
		suite := domainautobench.Suite{
			ID:     string(rune('a' + i)),
			Name:   "Test Suite",
			Status: domainautobench.SuiteStatusDraft,
		}
		if err := repo.Save(ctx, suite); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	suites, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(suites) != 3 {
		t.Errorf("FindAll() returned %d suites, want 3", len(suites))
	}
}

func TestSuiteRepository_UpdateStatus(t *testing.T) {
	db := setupSuiteTestDB(t)
	defer db.Close()

	repo := NewSQLiteSuiteRepository(db)
	ctx := context.Background()

	// Create a suite first
	suite := domainautobench.Suite{
		ID:     "test-suite-update",
		Name:   "Test Suite",
		Status: domainautobench.SuiteStatusDraft,
	}
	if err := repo.Save(ctx, suite); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Update status
	err := repo.UpdateStatus(ctx, "test-suite-update", domainautobench.SuiteStatusRunning, "/path/to/manifest.json")
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	// Verify update
	found, err := repo.FindByID(ctx, "test-suite-update")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Status != domainautobench.SuiteStatusRunning {
		t.Errorf("UpdateStatus() status = %v, want %v", found.Status, domainautobench.SuiteStatusRunning)
	}
}

func TestSuiteRepository_Delete(t *testing.T) {
	db := setupSuiteTestDB(t)
	defer db.Close()

	repo := NewSQLiteSuiteRepository(db)
	ctx := context.Background()

	// Create a suite first
	suite := domainautobench.Suite{
		ID:     "test-suite-delete",
		Name:   "Test Suite",
		Status: domainautobench.SuiteStatusDraft,
	}
	if err := repo.Save(ctx, suite); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Delete
	err := repo.Delete(ctx, "test-suite-delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(ctx, "test-suite-delete")
	if err == nil {
		t.Error("FindByID() expected error after deletion")
	}
}

func TestSuiteRepository_Delete_NotFound(t *testing.T) {
	db := setupSuiteTestDB(t)
	defer db.Close()

	repo := NewSQLiteSuiteRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent-id")
	if err == nil {
		t.Error("Delete() expected error for non-existent suite")
	}
}

// setupSuiteTestDB creates an in-memory SQLite database for testing.
func setupSuiteTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	db, err := database.InitializeSQLite(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	return db
}
