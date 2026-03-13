// Package usecase provides in-memory run repository for testing and development.
// TODO: Replace with SQLite implementation for production
package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// MemoryRunRepository provides an in-memory implementation of RunRepository.
// This is a temporary implementation for development. Production should use a persistent repository.
type MemoryRunRepository struct {
	runs    map[string]*execution.Run
	samples map[string][]execution.MetricSample
	logs    map[string][]LogEntry
	logDir  string
	mu      sync.RWMutex
}

// NewMemoryRunRepository creates a new in-memory run repository.
func NewMemoryRunRepository(logDir ...string) *MemoryRunRepository {
	repo := &MemoryRunRepository{
		runs:    make(map[string]*execution.Run),
		samples: make(map[string][]execution.MetricSample),
		logs:    make(map[string][]LogEntry),
	}
	if len(logDir) > 0 {
		repo.logDir = logDir[0]
		if repo.logDir != "" {
			if err := os.MkdirAll(repo.logDir, 0755); err != nil {
				slog.Warn("MemoryRunRepository: failed to create log dir", "dir", repo.logDir, "error", err)
				repo.logDir = ""
			}
		}
	}
	return repo
}

// Save saves a run to the repository.
func (r *MemoryRunRepository) Save(ctx context.Context, run *execution.Run) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.ID] = run
	slog.Debug("MemoryRunRepository: Saved run", "id", run.ID, "state", run.State)
	return nil
}

// FindByID finds a run by its ID.
func (r *MemoryRunRepository) FindByID(ctx context.Context, id string) (*execution.Run, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[id]
	if !ok {
		return nil, ErrBenchmarkNotFound
	}
	return run, nil
}

// FindAll finds runs with optional filtering and pagination.
func (r *MemoryRunRepository) FindAll(ctx context.Context, opts FindOptions) ([]*execution.Run, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var runs []*execution.Run
	for _, run := range r.runs {
		runs = append(runs, run)
	}
	return runs, nil
}

// UpdateState updates the state of a run.
func (r *MemoryRunRepository) UpdateState(ctx context.Context, id string, state execution.RunState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if run, ok := r.runs[id]; ok {
		run.State = state
		slog.Debug("MemoryRunRepository: Updated state", "id", id, "state", state)
		return nil
	}
	return ErrBenchmarkNotFound
}

// SaveMetricSample saves a metric sample for a run.
func (r *MemoryRunRepository) SaveMetricSample(ctx context.Context, runID string, sample execution.MetricSample) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.samples[runID] = append(r.samples[runID], sample)
	return nil
}

// GetMetricSamples retrieves all metric samples for a run.
func (r *MemoryRunRepository) GetMetricSamples(ctx context.Context, runID string) ([]execution.MetricSample, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	samples, ok := r.samples[runID]
	if !ok {
		return []execution.MetricSample{}, nil
	}
	return samples, nil
}

// SaveLogEntry saves a log entry for a run.
func (r *MemoryRunRepository) SaveLogEntry(ctx context.Context, runID string, entry LogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[runID] = append(r.logs[runID], entry)
	if r.logDir != "" {
		logPath := filepath.Join(r.logDir, fmt.Sprintf("%s.log", runID))
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := fmt.Fprintf(file, "%s [%s] %s\n", entry.Timestamp, entry.Stream, entry.Content); err != nil {
			return err
		}
	}
	return nil
}

// GetLogs returns all in-memory log entries for a run.
func (r *MemoryRunRepository) GetLogs(runID string) []LogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	logs := r.logs[runID]
	cloned := make([]LogEntry, len(logs))
	copy(cloned, logs)
	return cloned
}

// GetLogPath returns the persisted log file path for a run if log persistence is enabled.
func (r *MemoryRunRepository) GetLogPath(runID string) string {
	if r.logDir == "" {
		return ""
	}
	return filepath.Join(r.logDir, fmt.Sprintf("%s.log", runID))
}

// GetPersistedLogs reads the persisted log file for a run and returns parsed entries.
func (r *MemoryRunRepository) GetPersistedLogs(runID string) ([]LogEntry, error) {
	logPath := r.GetLogPath(runID)
	if logPath == "" {
		return nil, os.ErrNotExist
	}
	content, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	entries := make([]LogEntry, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		entry := LogEntry{Content: line, Stream: "stdout"}
		if firstSpace := strings.IndexByte(line, ' '); firstSpace > 0 {
			entry.Timestamp = line[:firstSpace]
			rest := strings.TrimSpace(line[firstSpace+1:])
			if strings.HasPrefix(rest, "[") {
				if closing := strings.Index(rest, "]"); closing > 1 {
					entry.Stream = rest[1:closing]
					entry.Content = strings.TrimSpace(rest[closing+1:])
				} else {
					entry.Content = rest
				}
			} else {
				entry.Content = rest
			}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Delete deletes a run by its ID.
func (r *MemoryRunRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.runs, id)
	delete(r.samples, id)
	delete(r.logs, id)
	return nil
}
