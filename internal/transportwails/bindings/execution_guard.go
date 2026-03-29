// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"fmt"
	"sync"
)

// ExecutionGuard enforces serial execution across all benchmark and
// AutoBench suite operations. Only one task (single benchmark run or
// AutoBench suite) may be active at a time.
//
// Concurrency safety:
//   - All exported methods are goroutine-safe via mu.
//   - The holder of an acquired lock must call Release() exactly once
//     when the task finishes (success, failure, or cancellation).
type ExecutionGuard struct {
	mu     sync.Mutex
	active *activeTask
}

// activeTask records what is currently running.
type activeTask struct {
	kind    string // "benchmark" or "autobench"
	runID   string // benchmark run ID or suite ID
	detail  string // human-readable description
}

// NewExecutionGuard creates a guard in the idle state.
func NewExecutionGuard() *ExecutionGuard {
	return &ExecutionGuard{}
}

// TryAcquire attempts to claim the execution slot.
// Returns an error if another task is already running.
// On success the caller owns the slot and must call Release when done.
func (g *ExecutionGuard) TryAcquire(kind, id, detail string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.active != nil {
		return fmt.Errorf("execution guard: a %s task is already running (%s: %s)",
			g.active.kind, g.active.runID, g.active.detail)
	}
	g.active = &activeTask{kind: kind, runID: id, detail: detail}
	return nil
}

// Release frees the execution slot. It must be called exactly once
// for each successful TryAcquire.
func (g *ExecutionGuard) Release() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.active = nil
}

// IsAnyTaskRunning reports whether a task currently holds the slot.
func (g *ExecutionGuard) IsAnyTaskRunning() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.active != nil
}

// RunningTask returns a description of the currently active task.
// If nothing is running the second return value is false.
func (g *ExecutionGuard) RunningTask() (kind, id, detail string, ok bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.active == nil {
		return "", "", "", false
	}
	return g.active.kind, g.active.runID, g.active.detail, true
}
