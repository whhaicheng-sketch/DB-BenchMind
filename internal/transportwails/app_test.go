package transportwails

import (
	"context"
	"testing"
)

func TestBenchmarkCleanupCommandContainsManagedProcessPatterns(t *testing.T) {
	cmd := benchmarkCleanupCommand()

	for _, want := range []string{
		"bin/db-benchmind",
		"LauncherBootstrap.*oewizard",
		"LauncherBootstrap.*charbench",
		"com\\.dom\\.benchmarking\\.swingbench\\.wizards\\.Wizard",
		"com\\.dom\\.benchmarking\\.swingbench\\.CharBench",
		"sysbench",
		"hammerdbcli",
		"pkill -TERM",
		"pkill -KILL",
	} {
		if !contains(cmd, want) {
			t.Fatalf("benchmarkCleanupCommand() missing %q in %q", want, cmd)
		}
	}
}

func TestAppShutdownInvokesBenchmarkCleanup(t *testing.T) {
	called := false
	gotCtx := context.Background()
	gotCmd := ""

	restore := runBenchmarkCleanup
	runBenchmarkCleanup = func(ctx context.Context, cmd string) error {
		called = true
		gotCtx = ctx
		gotCmd = cmd
		return nil
	}
	defer func() {
		runBenchmarkCleanup = restore
	}()

	app := NewApp()
	ctx := context.Background()
	app.Shutdown(ctx)

	if !called {
		t.Fatal("Shutdown() did not invoke benchmark cleanup")
	}
	if gotCtx != ctx {
		t.Fatal("Shutdown() passed unexpected context to cleanup")
	}
	if gotCmd != benchmarkCleanupCommand() {
		t.Fatalf("Shutdown() cleanup cmd = %q, want %q", gotCmd, benchmarkCleanupCommand())
	}
}

func contains(text, target string) bool {
	return len(target) == 0 || (len(text) >= len(target) && func() bool {
		for i := 0; i+len(target) <= len(text); i++ {
			if text[i:i+len(target)] == target {
				return true
			}
		}
		return false
	}())
}
