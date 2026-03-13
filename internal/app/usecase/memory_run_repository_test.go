package usecase

import (
	"context"
	"testing"
)

func TestMemoryRunRepositoryGetPersistedLogs(t *testing.T) {
	dir := t.TempDir()
	repo := NewMemoryRunRepository(dir)
	runID := "run-1"

	err := repo.SaveLogEntry(context.Background(), runID, LogEntry{
		Timestamp: "2026-03-13T12:00:00Z",
		Stream:    "stderr",
		Content:   "fatal: benchmark failed",
	})
	if err != nil {
		t.Fatalf("SaveLogEntry() error = %v", err)
	}

	logs, err := repo.GetPersistedLogs(runID)
	if err != nil {
		t.Fatalf("GetPersistedLogs() error = %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("GetPersistedLogs() len = %d, want 1", len(logs))
	}
	if logs[0].Stream != "stderr" {
		t.Fatalf("stream = %s, want stderr", logs[0].Stream)
	}
	if logs[0].Content != "fatal: benchmark failed" {
		t.Fatalf("content = %q", logs[0].Content)
	}
}
