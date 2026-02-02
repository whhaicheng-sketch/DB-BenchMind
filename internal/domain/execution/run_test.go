// Package execution provides unit tests for benchmark run model.
package execution

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestRun_SetState tests state setting with validation.
func TestRun_SetState(t *testing.T) {
	now := time.Now()
	run := &Run{
		ID:        uuid.New().String(),
		TaskID:    uuid.New().String(),
		State:     StatePending,
		CreatedAt: now,
	}

	tests := []struct {
		name      string
		fromState RunState
		toState   RunState
		wantErr   bool
	}{
		{"valid: pending -> preparing", StatePending, StatePreparing, false},
		{"valid: preparing -> prepared", StatePreparing, StatePrepared, false},
		{"valid: running -> completed", StateRunning, StateCompleted, false},
		{"invalid: pending -> completed", StatePending, StateCompleted, true},
		{"invalid: completed -> running", StateCompleted, StateRunning, true},
		{"invalid: preparing -> running", StatePreparing, StateRunning, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run.State = tt.fromState
			err := run.SetState(tt.toState)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && run.State != tt.toState {
				t.Errorf("SetState() state = %v, want %v", run.State, tt.toState)
			}
		})
	}
}

// TestRun_IsCompleted tests terminal state detection.
func TestRun_IsCompleted(t *testing.T) {
	tests := []struct {
		name  string
		state RunState
		want  bool
	}{
		{"running is not completed", StateRunning, false},
		{"pending is not completed", StatePending, false},
		{"completed is completed", StateCompleted, true},
		{"failed is completed", StateFailed, true},
		{"cancelled is completed", StateCancelled, true},
		{"timeout is completed", StateTimeout, true},
		{"force_stopped is completed", StateForceStopped, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := &Run{State: tt.state}
			if got := run.IsCompleted(); got != tt.want {
				t.Errorf("Run.IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRun_CalculateDuration tests duration calculation.
func TestRun_CalculateDuration(t *testing.T) {
	now := time.Now()
	started := now.Add(1 * time.Second)
	completed := now.Add(61 * time.Second) // 60 seconds later

	run := &Run{
		StartedAt:   &started,
		CompletedAt: &completed,
	}

	run.CalculateDuration()

	if run.Duration == nil {
		t.Fatal("Duration should not be nil after CalculateDuration()")
	}

	expected := 60 * time.Second
	if *run.Duration != expected {
		t.Errorf("Duration = %v, want %v", *run.Duration, expected)
	}
}

// TestRun_CalculateDuration_NoTimestamps tests duration calculation with missing timestamps.
func TestRun_CalculateDuration_NoTimestamps(t *testing.T) {
	run := &Run{}

	run.CalculateDuration()

	if run.Duration != nil {
		t.Error("Duration should remain nil when timestamps are missing")
	}
}

// TestRun_CalculateDuration_OnlyStarted tests duration calculation with only started_at.
func TestRun_CalculateDuration_OnlyStarted(t *testing.T) {
	now := time.Now()
	started := now.Add(1 * time.Second)

	run := &Run{
		StartedAt: &started,
	}

	run.CalculateDuration()

	if run.Duration != nil {
		t.Error("Duration should remain nil when completed_at is missing")
	}
}

// TestBenchmarkTask_Validate tests task validation.
func TestBenchmarkTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    BenchmarkTask
		wantErr bool
	}{
		{
			name: "valid task",
			task: BenchmarkTask{
				ID:           uuid.New().String(),
				Name:         "Test Task",
				ConnectionID: uuid.New().String(),
				TemplateID:   "sysbench-oltp-read-write",
				CreatedAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing id",
			task: BenchmarkTask{
				Name:         "Test Task",
				ConnectionID: uuid.New().String(),
				TemplateID:   "sysbench-oltp-read-write",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			task: BenchmarkTask{
				ID:           uuid.New().String(),
				ConnectionID: uuid.New().String(),
				TemplateID:   "sysbench-oltp-read-write",
			},
			wantErr: true,
		},
		{
			name: "missing connection_id",
			task: BenchmarkTask{
				ID:         uuid.New().String(),
				Name:       "Test Task",
				TemplateID: "sysbench-oltp-read-write",
			},
			wantErr: true,
		},
		{
			name: "missing template_id",
			task: BenchmarkTask{
				ID:           uuid.New().String(),
				Name:         "Test Task",
				ConnectionID: uuid.New().String(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BenchmarkTask.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestInvalidStateTransitionError tests error formatting.
func TestInvalidStateTransitionError(t *testing.T) {
	err := &InvalidStateTransitionError{
		From: StatePending,
		To:   StateCompleted,
	}

	expected := "invalid state transition: pending -> completed"
	if got := err.Error(); got != expected {
		t.Errorf("InvalidStateTransitionError.Error() = %v, want %v", got, expected)
	}
}

// TestMetricSample tests metric sample structure.
func TestMetricSample(t *testing.T) {
	now := time.Now()
	sample := MetricSample{
		Timestamp:  now,
		Phase:      "run",
		TPS:        1000.5,
		QPS:        5000.0,
		LatencyAvg: 5.2,
		LatencyP95: 10.5,
		LatencyP99: 25.0,
		ErrorRate:  0.1,
	}

	if sample.TPS != 1000.5 {
		t.Errorf("TPS = %v, want %v", sample.TPS, 1000.5)
	}
	if sample.Phase != "run" {
		t.Errorf("Phase = %v, want %v", sample.Phase, "run")
	}
}

// TestBenchmarkResult tests benchmark result structure.
func TestBenchmarkResult(t *testing.T) {
	result := BenchmarkResult{
		RunID:             "test-run-id",
		TPSCalculated:     1000.5,
		LatencyAvg:        5.2,
		LatencyP95:        10.5,
		LatencyP99:        25.0,
		ErrorCount:        10,
		ErrorRate:         0.1,
		Duration:          60 * time.Second,
		TotalTransactions: 60000,
		TotalQueries:      300000,
	}

	if result.RunID != "test-run-id" {
		t.Errorf("RunID = %v, want %v", result.RunID, "test-run-id")
	}
	if result.TPSCalculated != 1000.5 {
		t.Errorf("TPSCalculated = %v, want %v", result.TPSCalculated, 1000.5)
	}
}

// TestTaskOptions tests task options structure.
func TestTaskOptions(t *testing.T) {
	options := TaskOptions{
		SkipPrepare:    false,
		SkipCleanup:    true,
		WarmupTime:     30,
		SampleInterval: 1 * time.Second,
		DryRun:         false,
		PrepareTimeout: 30 * time.Minute,
		RunTimeout:     24 * time.Hour,
	}

	if options.WarmupTime != 30 {
		t.Errorf("WarmupTime = %v, want %v", options.WarmupTime, 30)
	}
	if options.SkipCleanup != true {
		t.Errorf("SkipCleanup = %v, want %v", options.SkipCleanup, true)
	}
}
