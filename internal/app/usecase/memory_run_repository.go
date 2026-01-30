// Package usecase provides in-memory run repository for testing and development.
// TODO: Replace with SQLite implementation for production
package usecase

import (
	"context"
	"log/slog"
	"sync"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// MemoryRunRepository provides an in-memory implementation of RunRepository.
// This is a temporary implementation for development. Production should use a persistent repository.
type MemoryRunRepository struct {
	runs   map[string]*execution.Run
	samples map[string][]execution.MetricSample
	logs   map[string][]LogEntry
	mu     sync.RWMutex
}

// NewMemoryRunRepository creates a new in-memory run repository.
func NewMemoryRunRepository() *MemoryRunRepository {
	return &MemoryRunRepository{
		runs:   make(map[string]*execution.Run),
		samples: make(map[string][]execution.MetricSample),
		logs:   make(map[string][]LogEntry),
	}
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
	return nil
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
