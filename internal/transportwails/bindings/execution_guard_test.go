package bindings

import (
	"sync"
	"testing"
)

func TestExecutionGuard_TryAcquire(t *testing.T) {
	tests := []struct {
		name      string
		kind      string
		id        string
		detail    string
		wantErr   bool
	}{
		{
			name:    "acquire idle slot succeeds",
			kind:    "benchmark",
			id:      "run-001",
			detail:  "test benchmark",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewExecutionGuard()
			err := g.TryAcquire(tt.kind, tt.id, tt.detail)
			if (err != nil) != tt.wantErr {
				t.Errorf("TryAcquire() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutionGuard_TryAcquire_AlreadyRunning(t *testing.T) {
	g := NewExecutionGuard()

	err := g.TryAcquire("benchmark", "run-001", "first task")
	if err != nil {
		t.Fatalf("first TryAcquire failed: %v", err)
	}

	err = g.TryAcquire("benchmark", "run-002", "second task")
	if err == nil {
		t.Error("expected error when acquiring already-held slot")
	}
}

func TestExecutionGuard_Release(t *testing.T) {
	g := NewExecutionGuard()

	err := g.TryAcquire("benchmark", "run-001", "test")
	if err != nil {
		t.Fatalf("TryAcquire failed: %v", err)
	}

	g.Release()

	if g.IsAnyTaskRunning() {
		t.Error("expected no task running after Release()")
	}

	// Should be able to acquire again after release
	err = g.TryAcquire("autobench", "suite-001", "suite test")
	if err != nil {
		t.Errorf("TryAcquire after Release failed: %v", err)
	}
}

func TestExecutionGuard_IsAnyTaskRunning(t *testing.T) {
	g := NewExecutionGuard()

	if g.IsAnyTaskRunning() {
		t.Error("new guard should report no task running")
	}

	_ = g.TryAcquire("benchmark", "run-001", "test")

	if !g.IsAnyTaskRunning() {
		t.Error("expected task running after TryAcquire")
	}

	g.Release()

	if g.IsAnyTaskRunning() {
		t.Error("expected no task running after Release")
	}
}

func TestExecutionGuard_RunningTask(t *testing.T) {
	g := NewExecutionGuard()

	// Nothing running
	_, _, _, ok := g.RunningTask()
	if ok {
		t.Error("expected ok=false when nothing is running")
	}

	// Acquire
	_ = g.TryAcquire("benchmark", "run-001", "my benchmark")

	kind, id, detail, ok := g.RunningTask()
	if !ok {
		t.Error("expected ok=true after TryAcquire")
	}
	if kind != "benchmark" {
		t.Errorf("kind = %q, want %q", kind, "benchmark")
	}
	if id != "run-001" {
		t.Errorf("id = %q, want %q", id, "run-001")
	}
	if detail != "my benchmark" {
		t.Errorf("detail = %q, want %q", detail, "my benchmark")
	}

	g.Release()

	_, _, _, ok = g.RunningTask()
	if ok {
		t.Error("expected ok=false after Release")
	}
}

func TestExecutionGuard_ConcurrentAccess(t *testing.T) {
	g := NewExecutionGuard()

	const goroutines = 50
	var wg sync.WaitGroup
	acquired := make(chan struct{}, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := g.TryAcquire("benchmark", "run", "concurrent test")
			if err == nil {
				acquired <- struct{}{}
				g.Release()
			}
		}(i)
	}

	wg.Wait()
	close(acquired)

	count := 0
	for range acquired {
		count++
	}

	// Exactly one goroutine should have acquired at a time,
	// but overall multiple should succeed since they release.
	// Just verify at least one succeeded.
	if count == 0 {
		t.Error("expected at least one successful acquisition")
	}
}

func TestExecutionGuard_ReleaseWithoutAcquire(t *testing.T) {
	g := NewExecutionGuard()
	// Release on an idle guard should not panic
	g.Release()
	// No assertion needed; this test passes if it does not panic.
}
