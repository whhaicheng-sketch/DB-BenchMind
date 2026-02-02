// Package execution provides benchmark execution domain models.
// Implements: REQ-EXEC-001 ~ REQ-EXEC-010
package execution

// RunState represents the state of a benchmark run.
// Implements: spec.md 3.4.2
type RunState string

const (
	StatePending      RunState = "pending"       // Created, waiting to execute
	StatePreparing    RunState = "preparing"     // Preparing data
	StatePrepared     RunState = "prepared"      // Preparation complete
	StateWarmingUp    RunState = "warming_up"    // Warming up
	StateRunning      RunState = "running"       // Running
	StateCompleted    RunState = "completed"     // Completed successfully
	StateFailed       RunState = "failed"        // Failed
	StateCancelled    RunState = "cancelled"     // Cancelled by user
	StateTimeout      RunState = "timeout"       // Timeout
	StateForceStopped RunState = "force_stopped" // Force stopped
)

// IsValid checks if the state is valid.
func (s RunState) IsValid() bool {
	switch s {
	case StatePending, StatePreparing, StatePrepared, StateWarmingUp,
		StateRunning, StateCompleted, StateFailed, StateCancelled,
		StateTimeout, StateForceStopped:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the state is a terminal state (no further transitions possible).
// Implements: REQ-EXEC-008
func (s RunState) IsTerminal() bool {
	return s == StateCompleted || s == StateFailed ||
		s == StateCancelled || s == StateTimeout || s == StateForceStopped
}

// CanTransitionTo checks if a transition from current state to target state is valid.
// Implements: spec.md 3.4.2 state transition rules
func (s RunState) CanTransitionTo(target RunState) bool {
	// Define valid state transitions
	transitions := map[RunState][]RunState{
		StatePending:   {StatePreparing, StateCancelled},
		StatePreparing: {StatePrepared, StateFailed, StateCancelled, StateTimeout},
		StatePrepared:  {StateWarmingUp, StateCancelled},
		StateWarmingUp: {StateRunning, StateFailed, StateCancelled, StateTimeout},
		StateRunning:   {StateCompleted, StateFailed, StateCancelled, StateTimeout, StateForceStopped},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}

	for _, state := range allowed {
		if state == target {
			return true
		}
	}
	return false
}

// String implements Stringer interface.
func (s RunState) String() string {
	return string(s)
}
