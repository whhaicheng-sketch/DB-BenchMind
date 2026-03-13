package bindings

import (
	"testing"

	domaintask "github.com/whhaicheng/DB-BenchMind/internal/domain/task"
)

func TestBuildPhaseExecutionConfig(t *testing.T) {
	base := map[string]interface{}{
		"time":    60,
		"threads": 8,
	}

	t.Run("prepare phase freezes runtime into prepare-only config", func(t *testing.T) {
		options, params := buildPhaseExecutionConfig(base, domaintask.PhasePrepare)
		if options.SkipPrepare {
			t.Fatalf("prepare phase should not skip prepare")
		}
		if !options.SkipCleanup {
			t.Fatalf("prepare phase should skip cleanup")
		}
		if got := params["time"]; got != 0 {
			t.Fatalf("prepare time = %v, want 0", got)
		}
		if got := params["_original_time"]; got != 60 {
			t.Fatalf("prepare _original_time = %v, want 60", got)
		}
	})

	t.Run("cleanup phase does not execute run workload", func(t *testing.T) {
		options, params := buildPhaseExecutionConfig(base, domaintask.PhaseCleanup)
		if !options.SkipPrepare {
			t.Fatalf("cleanup phase should skip prepare")
		}
		if options.SkipCleanup {
			t.Fatalf("cleanup phase should execute cleanup")
		}
		if got := params["time"]; got != 0 {
			t.Fatalf("cleanup time = %v, want 0", got)
		}
		if _, ok := params["_original_time"]; ok {
			t.Fatalf("cleanup phase should not carry _original_time")
		}
		if got := params["_cleanup_only"]; got != true {
			t.Fatalf("cleanup _cleanup_only = %v, want true", got)
		}
	})
}

func TestValidateReadiness(t *testing.T) {
	err := validateReadiness(domaintask.Readiness{
		TemplateSelected:   true,
		ConnectionSelected: true,
		ActionSupported:    true,
		RuntimeValid:       true,
		DBValid:            true,
	})
	if err != nil {
		t.Fatalf("validateReadiness() returned unexpected error: %v", err)
	}

	err = validateReadiness(domaintask.Readiness{
		TemplateSelected:   true,
		ConnectionSelected: true,
		ActionSupported:    false,
		RuntimeValid:       true,
		DBValid:            true,
	})
	if err == nil {
		t.Fatalf("validateReadiness() expected action support error")
	}
}
