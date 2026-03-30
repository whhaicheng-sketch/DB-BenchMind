package bindings

import (
	"testing"
)

func TestBenchmarkBinding_GetRunningTask(t *testing.T) {
	guard := NewExecutionGuard()
	binding := &BenchmarkBinding{guard: guard}

	tests := []struct {
		name       string
		setup      func()
		wantKind   string
		wantID     string
		wantDetail string
		wantOK     bool
	}{
		{
			name:       "no task running",
			setup:      func() {},
			wantKind:   "",
			wantID:     "",
			wantDetail: "",
			wantOK:     false,
		},
		{
			name: "task acquired",
			setup: func() {
				_ = guard.TryAcquire("benchmark", "run-123", "Test Benchmark")
			},
			wantKind:   "benchmark",
			wantID:     "run-123",
			wantDetail: "Test Benchmark",
			wantOK:     true,
		},
		{
			name: "task released",
			setup: func() {
				guard.Release()
			},
			wantKind:   "",
			wantID:     "",
			wantDetail: "",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := binding.GetRunningTask()
			if result.Kind != tt.wantKind {
				t.Errorf("Kind = %q, want %q", result.Kind, tt.wantKind)
			}
			if result.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", result.ID, tt.wantID)
			}
			if result.Detail != tt.wantDetail {
				t.Errorf("Detail = %q, want %q", result.Detail, tt.wantDetail)
			}
			if result.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v", result.OK, tt.wantOK)
			}
		})
	}
}
