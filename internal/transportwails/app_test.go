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
	var gotCtx context.Context
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
	if gotCtx == nil {
		t.Fatal("Shutdown() did not pass a context to cleanup")
	}
	if gotCmd != benchmarkCleanupCommand() {
		t.Fatalf("Shutdown() cleanup cmd = %q, want %q", gotCmd, benchmarkCleanupCommand())
	}
}

func TestAppShutdownUsesIndependentCleanupContextWhenShutdownContextCancelled(t *testing.T) {
	called := false
	errAtCleanup := error(nil)
	hadDeadline := false

	restore := runBenchmarkCleanup
	runBenchmarkCleanup = func(ctx context.Context, cmd string) error {
		called = true
		errAtCleanup = ctx.Err()
		_, hadDeadline = ctx.Deadline()
		return nil
	}
	defer func() {
		runBenchmarkCleanup = restore
	}()

	app := NewApp()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	app.Shutdown(ctx)

	if !called {
		t.Fatal("Shutdown() did not invoke benchmark cleanup")
	}
	if errAtCleanup != nil {
		t.Fatalf("Shutdown() passed cancelled cleanup context: %v", errAtCleanup)
	}
	if !hadDeadline {
		t.Fatal("Shutdown() cleanup context should have a deadline")
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
