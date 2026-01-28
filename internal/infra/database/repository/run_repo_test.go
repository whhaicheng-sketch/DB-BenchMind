// Package repository provides unit tests for run repository.
package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// setupRunTestDB creates an in-memory SQLite database for run testing.
func setupRunTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			state TEXT NOT NULL,
			created_at TEXT NOT NULL,
			started_at TEXT,
			completed_at TEXT,
			duration_seconds REAL,
			result_summary_json TEXT,
			result_detail_json TEXT,
			error_message TEXT,
			config_snapshot_path TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_runs_task_id ON runs(task_id);
		CREATE INDEX IF NOT EXISTS idx_runs_state ON runs(state);
		CREATE INDEX IF NOT EXISTS idx_runs_created_at ON runs(created_at DESC);

		CREATE TABLE IF NOT EXISTS metric_samples (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			phase TEXT NOT NULL,
			tps REAL,
			qps REAL,
			latency_avg REAL,
			latency_p95 REAL,
			latency_p99 REAL,
			error_rate REAL
		);

		CREATE INDEX IF NOT EXISTS idx_metric_samples_run_id ON metric_samples(run_id);

		CREATE TABLE IF NOT EXISTS run_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			stream TEXT NOT NULL,
			content TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_run_logs_run_id ON run_logs(run_id);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("create tables: %v", err)
	}

	return db
}

// TestSQLiteRunRepository_Save_FindByID tests Save and FindByID operations.
func TestSQLiteRunRepository_Save_FindByID(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	now := time.Now()
	run := &execution.Run{
		ID:        uuid.New().String(),
		TaskID:    uuid.New().String(),
		State:     execution.StatePending,
		CreatedAt: now,
		WorkDir:   "/tmp/run",
	}

	// Save
	err := repo.Save(ctx, run)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}

	if found.ID != run.ID {
		t.Errorf("ID = %s, want %s", found.ID, run.ID)
	}
	if found.State != run.State {
		t.Errorf("State = %s, want %s", found.State, run.State)
	}
	if found.TaskID != run.TaskID {
		t.Errorf("TaskID = %s, want %s", found.TaskID, run.TaskID)
	}
}

// TestSQLiteRunRepository_FindByID_NotFound tests finding non-existent run.
func TestSQLiteRunRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	_, err := repo.FindByID(ctx, "nonexistent")
	if err != ErrRunNotFound {
		t.Errorf("Expected ErrRunNotFound, got: %v", err)
	}
}

// TestSQLiteRunRepository_UpdateState tests state updates.
func TestSQLiteRunRepository_UpdateState(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	now := time.Now()
	run := &execution.Run{
		ID:        uuid.New().String(),
		TaskID:    uuid.New().String(),
		State:     execution.StatePending,
		CreatedAt: now,
	}

	repo.Save(ctx, run)

	// Valid transition
	err := repo.UpdateState(ctx, run.ID, execution.StatePreparing)
	if err != nil {
		t.Fatalf("UpdateState() with valid transition failed: %v", err)
	}

	// Verify
	updated, _ := repo.FindByID(ctx, run.ID)
	if updated.State != execution.StatePreparing {
		t.Errorf("State = %s, want %s", updated.State, execution.StatePreparing)
	}

	// Invalid transition
	err = repo.UpdateState(ctx, run.ID, execution.StateCompleted)
	if err == nil {
		t.Error("UpdateState() with invalid transition should return error")
	}
}

// TestSQLiteRunRepository_FindAll tests finding all runs.
func TestSQLiteRunRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	now := time.Now()

	// Create multiple runs
	runs := []*execution.Run{
		{
			ID:        uuid.New().String(),
			TaskID:    uuid.New().String(),
			State:     execution.StateCompleted,
			CreatedAt: now.Add(-2 * time.Hour),
			WorkDir:   "/tmp/run1",
		},
		{
			ID:        uuid.New().String(),
			TaskID:    uuid.New().String(),
			State:     execution.StateRunning,
			CreatedAt: now.Add(-1 * time.Hour),
			WorkDir:   "/tmp/run2",
		},
	}

	for _, run := range runs {
		if err := repo.Save(ctx, run); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
	}

	// FindAll
	all, err := repo.FindAll(ctx, usecase.FindOptions{})
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("FindAll() count = %d, want 2", len(all))
	}
}

// TestSQLiteRunRepository_FindAll_WithFilters tests filtering.
func TestSQLiteRunRepository_FindAll_WithFilters(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	now := time.Now()
	taskID := uuid.New().String()

	// Create runs with different states
	runs := []*execution.Run{
		{
			ID:        uuid.New().String(),
			TaskID:    taskID,
			State:     execution.StateCompleted,
			CreatedAt: now,
			WorkDir:   "/tmp/run1",
		},
		{
			ID:        uuid.New().String(),
			TaskID:    taskID,
			State:     execution.StateRunning,
			CreatedAt: now,
			WorkDir:   "/tmp/run2",
		},
	}

	for _, run := range runs {
		if err := repo.Save(ctx, run); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
	}

	// Filter by state
	state := execution.StateRunning
	filtered, err := repo.FindAll(ctx, usecase.FindOptions{
		StateFilter: &state,
	})
	if err != nil {
		t.Fatalf("FindAll() with filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("FindAll() with filter count = %d, want 1", len(filtered))
	}
	if filtered[0].State != execution.StateRunning {
		t.Errorf("Filtered state = %s, want %s", filtered[0].State, execution.StateRunning)
	}

	// Filter by task ID
	byTask, err := repo.FindAll(ctx, usecase.FindOptions{
		TaskID: taskID,
	})
	if err != nil {
		t.Fatalf("FindAll() with TaskID failed: %v", err)
	}

	if len(byTask) != 2 {
		t.Errorf("FindAll() with TaskID count = %d, want 2", len(byTask))
	}
}

// TestSQLiteRunRepository_SaveMetricSample tests saving metric samples.
func TestSQLiteRunRepository_SaveMetricSample(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	runID := uuid.New().String()

	// Save a metric sample
	sample := execution.MetricSample{
		Timestamp:   time.Now(),
		Phase:       "run",
		TPS:         1000.0,
		QPS:         5000.0,
		LatencyAvg:  5.0,
		LatencyP95:  10.0,
		LatencyP99:  20.0,
		ErrorRate:   0.1,
	}

	err := repo.SaveMetricSample(ctx, runID, sample)
	if err != nil {
		t.Fatalf("SaveMetricSample() failed: %v", err)
	}

	// Verify by querying directly
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM metric_samples WHERE run_id = ?", runID).Scan(&count)
	if err != nil {
		t.Fatalf("Query metric samples failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Metric sample count = %d, want 1", count)
	}
}

// TestSQLiteRunRepository_SaveLogEntry tests saving log entries.
func TestSQLiteRunRepository_SaveLogEntry(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	runID := uuid.New().String()

	// Save a log entry
	entry := usecase.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Stream:    "stdout",
		Content:   "Test log message",
	}

	err := repo.SaveLogEntry(ctx, runID, entry)
	if err != nil {
		t.Fatalf("SaveLogEntry() failed: %v", err)
	}

	// Verify by querying directly
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM run_logs WHERE run_id = ?", runID).Scan(&count)
	if err != nil {
		t.Fatalf("Query log entries failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Log entry count = %d, want 1", count)
	}
}

// TestSQLiteRunRepository_Delete tests deleting runs.
func TestSQLiteRunRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	run := &execution.Run{
		ID:        uuid.New().String(),
		TaskID:    uuid.New().String(),
		State:     execution.StatePending,
		CreatedAt: time.Now(),
		WorkDir:   "/tmp/run",
	}

	repo.Save(ctx, run)

	// Delete
	err := repo.Delete(ctx, run.ID)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deleted
	_, err = repo.FindByID(ctx, run.ID)
	if err != ErrRunNotFound {
		t.Errorf("Expected ErrRunNotFound after Delete(), got: %v", err)
	}
}

// TestSQLiteRunRepository_Delete_NotFound tests deleting non-existent run.
func TestSQLiteRunRepository_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupRunTestDB(t)
	defer db.Close()

	repo := NewSQLiteRunRepository(db)

	err := repo.Delete(ctx, "nonexistent")
	if err != ErrRunNotFound {
		t.Errorf("Expected ErrRunNotFound, got: %v", err)
	}
}
