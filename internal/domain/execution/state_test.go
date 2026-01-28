// Package execution provides unit tests for run state machine.
package execution

import (
	"testing"
)

// TestRunState_IsValid tests valid state detection.
func TestRunState_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		state RunState
		want  bool
	}{
		{"pending is valid", StatePending, true},
		{"preparing is valid", StatePreparing, true},
		{"prepared is valid", StatePrepared, true},
		{"warming_up is valid", StateWarmingUp, true},
		{"running is valid", StateRunning, true},
		{"completed is valid", StateCompleted, true},
		{"failed is valid", StateFailed, true},
		{"cancelled is valid", StateCancelled, true},
		{"timeout is valid", StateTimeout, true},
		{"force_stopped is valid", StateForceStopped, true},
		{"invalid state", RunState("invalid"), false},
		{"empty state", RunState(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsValid(); got != tt.want {
				t.Errorf("RunState.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRunState_IsTerminal tests terminal state detection.
func TestRunState_IsTerminal(t *testing.T) {
	tests := []struct {
		name  string
		state RunState
		want  bool
	}{
		{"completed is terminal", StateCompleted, true},
		{"failed is terminal", StateFailed, true},
		{"cancelled is terminal", StateCancelled, true},
		{"timeout is terminal", StateTimeout, true},
		{"force_stopped is terminal", StateForceStopped, true},
		{"pending is not terminal", StatePending, false},
		{"preparing is not terminal", StatePreparing, false},
		{"prepared is not terminal", StatePrepared, false},
		{"warming_up is not terminal", StateWarmingUp, false},
		{"running is not terminal", StateRunning, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsTerminal(); got != tt.want {
				t.Errorf("RunState.IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRunState_CanTransitionTo tests valid state transitions.
func TestRunState_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		from   RunState
		to     RunState
		wantOk bool
	}{
		// Happy path: pending -> preparing -> prepared -> warming_up -> running -> completed
		{"pending -> preparing", StatePending, StatePreparing, true},
		{"preparing -> prepared", StatePreparing, StatePrepared, true},
		{"prepared -> warming_up", StatePrepared, StateWarmingUp, true},
		{"warming_up -> running", StateWarmingUp, StateRunning, true},
		{"running -> completed", StateRunning, StateCompleted, true},

		// Cancellation from any state
		{"pending -> cancelled", StatePending, StateCancelled, true},
		{"preparing -> cancelled", StatePreparing, StateCancelled, true},
		{"prepared -> cancelled", StatePrepared, StateCancelled, true},
		{"warming_up -> cancelled", StateWarmingUp, StateCancelled, true},
		{"running -> cancelled", StateRunning, StateCancelled, true},

		// Failure transitions
		{"preparing -> failed", StatePreparing, StateFailed, true},
		{"warming_up -> failed", StateWarmingUp, StateFailed, true},
		{"running -> failed", StateRunning, StateFailed, true},

		// Timeout transitions
		{"preparing -> timeout", StatePreparing, StateTimeout, true},
		{"warming_up -> timeout", StateWarmingUp, StateTimeout, true},
		{"running -> timeout", StateRunning, StateTimeout, true},

		// Force stop (only from running)
		{"running -> force_stopped", StateRunning, StateForceStopped, true},

		// Invalid transitions
		{"pending -> completed (skip)", StatePending, StateCompleted, false},
		{"pending -> running (skip)", StatePending, StateRunning, false},
		{"completed -> running (terminal)", StateCompleted, StateRunning, false},
		{"failed -> running (terminal)", StateFailed, StateRunning, false},
		{"preparing -> running (skip)", StatePreparing, StateRunning, false},
		{"pending -> pending (no change)", StatePending, StatePending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk := tt.from.CanTransitionTo(tt.to)
			if gotOk != tt.wantOk {
				t.Errorf("RunState.CanTransitionTo() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

// TestRunState_String tests string representation.
func TestRunState_String(t *testing.T) {
	tests := []struct {
		state RunState
		want  string
	}{
		{StatePending, "pending"},
		{StatePreparing, "preparing"},
		{StatePrepared, "prepared"},
		{StateWarmingUp, "warming_up"},
		{StateRunning, "running"},
		{StateCompleted, "completed"},
		{StateFailed, "failed"},
		{StateCancelled, "cancelled"},
		{StateTimeout, "timeout"},
		{StateForceStopped, "force_stopped"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("RunState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
