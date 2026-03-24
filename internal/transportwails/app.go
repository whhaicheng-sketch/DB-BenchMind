// Package transportwails provides the Wails transport layer for DB-BenchMind.
package transportwails

import (
	"context"
	"log/slog"
	"os/exec"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/bindings"
)

var runBenchmarkCleanup = func(ctx context.Context, cmd string) error {
	return exec.CommandContext(ctx, "bash", "-lc", cmd).Run()
}

const benchmarkCleanupTimeout = 3 * time.Second

// App provides the main application bindings for Wails.
type App struct {
	ctx context.Context
	holder bindingsHolder
}

// bindingsHolder holds bindings that need context injection
type bindingsHolder struct {
	benchmarkBinding *bindings.BenchmarkBinding
	monitorBinding   *bindings.MonitorBinding
	taskBinding      *bindings.TaskBinding
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// SetBenchmarkBinding stores the benchmark binding for context injection.
func (a *App) SetBenchmarkBinding(b *bindings.BenchmarkBinding) {
	a.holder.benchmarkBinding = b
}

// SetMonitorBinding stores the monitor binding for context injection.
func (a *App) SetMonitorBinding(m *bindings.MonitorBinding) {
	a.holder.monitorBinding = m
}

// SetTaskBinding stores the task binding for context injection.
func (a *App) SetTaskBinding(t *bindings.TaskBinding) {
	a.holder.taskBinding = t
}

// Startup is called when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("Wails App: startup called")

	// Inject context into benchmark binding
	if a.holder.benchmarkBinding != nil {
		a.holder.benchmarkBinding.SetContext(ctx)
		slog.Info("BenchmarkBinding context injected")
	}

	// Inject context into monitor binding
	if a.holder.monitorBinding != nil {
		a.holder.monitorBinding.SetContext(ctx)
		slog.Info("MonitorBinding context injected")
	}

	if a.holder.taskBinding != nil {
		a.holder.taskBinding.SetContext(ctx)
		slog.Info("TaskBinding context injected")
	}
}

// Shutdown is called when the app is shutting down.
func (a *App) Shutdown(ctx context.Context) {
	slog.Info("Wails App: shutdown called")

	cleanupCtx, cancel := context.WithTimeout(context.Background(), benchmarkCleanupTimeout)
	defer cancel()

	if err := runBenchmarkCleanup(cleanupCtx, benchmarkCleanupCommand()); err != nil {
		slog.Warn("Wails App: benchmark cleanup failed during shutdown", "error", err)
	}
}

// GetAppVersion returns the application version.
func (a *App) GetAppVersion() string {
	return "1.0.0-wails"
}

// GetCurrentContext returns the current application context.
// This is useful for Wails runtime operations.
func (a *App) GetCurrentContext() context.Context {
	return a.ctx
}

// EmitEvent emits a custom event to the frontend.
func (a *App) EmitEvent(eventName string, data interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, eventName, data)
	}
}

func benchmarkCleanupCommand() string {
	return `pkill -f "bin/db-benchmind" >/dev/null 2>&1 || true; ` +
		`pkill -TERM -f 'LauncherBootstrap.*oewizard|LauncherBootstrap.*charbench|com\.dom\.benchmarking\.swingbench\.wizards\.Wizard|com\.dom\.benchmarking\.swingbench\.CharBench|(^|/)sysbench([[:space:]]|$)|(^|/)hammerdbcli([[:space:]]|$)' >/dev/null 2>&1 || true; ` +
		`sleep 0.5; ` +
		`pkill -KILL -f 'LauncherBootstrap.*oewizard|LauncherBootstrap.*charbench|com\.dom\.benchmarking\.swingbench\.wizards\.Wizard|com\.dom\.benchmarking\.swingbench\.CharBench|(^|/)sysbench([[:space:]]|$)|(^|/)hammerdbcli([[:space:]]|$)' >/dev/null 2>&1 || true`
}
