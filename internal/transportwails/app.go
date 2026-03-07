// Package transportwails provides the Wails transport layer for DB-BenchMind.
package transportwails

import (
	"context"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/bindings"
)

// App provides the main application bindings for Wails.
type App struct {
	ctx context.Context
	holder bindingsHolder
}

// bindingsHolder holds bindings that need context injection
type bindingsHolder struct {
	benchmarkBinding *bindings.BenchmarkBinding
	monitorBinding   *bindings.MonitorBinding
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
}

// Shutdown is called when the app is shutting down.
func (a *App) Shutdown(ctx context.Context) {
	slog.Info("Wails App: shutdown called")

	// Cleanup resources here
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
