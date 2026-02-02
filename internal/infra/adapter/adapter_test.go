// Package adapter provides unit tests for benchmark adapter interface.
package adapter

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// mockBenchmarkAdapter is a mock implementation of BenchmarkAdapter for testing.
type mockBenchmarkAdapter struct {
	adapterType AdapterType
}

func (m *mockBenchmarkAdapter) Type() AdapterType {
	return m.adapterType
}

func (m *mockBenchmarkAdapter) BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error) {
	return &Command{CmdLine: "prepare", WorkDir: config.WorkDir}, nil
}

func (m *mockBenchmarkAdapter) BuildRunCommand(ctx context.Context, config *Config) (*Command, error) {
	return &Command{CmdLine: "run", WorkDir: config.WorkDir}, nil
}

func (m *mockBenchmarkAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	return &Command{CmdLine: "cleanup", WorkDir: config.WorkDir}, nil
}

func (m *mockBenchmarkAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*Result, error) {
	return &Result{TPS: 1000.0}, nil
}

func (m *mockBenchmarkAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader, stderr io.Reader) (<-chan Sample, <-chan error, *strings.Builder) {
	sampleCh := make(chan Sample, 1)
	errCh := make(chan error, 1)
	var buf strings.Builder
	close(sampleCh)
	close(errCh)
	return sampleCh, errCh, &buf
}

func (m *mockBenchmarkAdapter) ValidateConfig(ctx context.Context, config *Config) error {
	return nil
}

func (m *mockBenchmarkAdapter) SupportsDatabase(dbType connection.DatabaseType) bool {
	return true
}

func (m *mockBenchmarkAdapter) ParseFinalResults(ctx context.Context, stdout string) (*FinalResult, error) {
	return &FinalResult{
		TransactionsPerSec: 1000.0,
		TotalTransactions: 1000,
	}, nil
}

// TestAdapterRegistry_Register tests adapter registration.
func TestAdapterRegistry_Register(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter := &mockBenchmarkAdapter{adapterType: AdapterTypeSysbench}

	registry.Register(adapter)

	// Verify adapter is registered
	got := registry.Get(AdapterTypeSysbench)
	if got == nil {
		t.Error("Get() returned nil for registered adapter")
	}
	if got.Type() != AdapterTypeSysbench {
		t.Errorf("Get() type = %v, want %v", got.Type(), AdapterTypeSysbench)
	}
}

// TestAdapterRegistry_Get tests getting adapters.
func TestAdapterRegistry_Get(t *testing.T) {
	registry := NewAdapterRegistry()

	// Register multiple adapters
	sysbench := &mockBenchmarkAdapter{adapterType: AdapterTypeSysbench}
	hammerdb := &mockBenchmarkAdapter{adapterType: AdapterTypeHammerDB}
	registry.Register(sysbench)
	registry.Register(hammerdb)

	tests := []struct {
		name        string
		adapterType AdapterType
		wantNil     bool
	}{
		{"get sysbench", AdapterTypeSysbench, false},
		{"get hammerdb", AdapterTypeHammerDB, false},
		{"get unregistered adapter", AdapterTypeSwingbench, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.Get(tt.adapterType)
			if (got == nil) != tt.wantNil {
				t.Errorf("Get() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

// TestAdapterRegistry_GetByTool tests getting adapter by tool name.
func TestAdapterRegistry_GetByTool(t *testing.T) {
	registry := NewAdapterRegistry()

	// Register adapters
	sysbench := &mockBenchmarkAdapter{adapterType: AdapterTypeSysbench}
	swingbench := &mockBenchmarkAdapter{adapterType: AdapterTypeSwingbench}
	hammerdb := &mockBenchmarkAdapter{adapterType: AdapterTypeHammerDB}
	registry.Register(sysbench)
	registry.Register(swingbench)
	registry.Register(hammerdb)

	tests := []struct {
		name    string
		tool    string
		wantNil bool
	}{
		{"get sysbench", "sysbench", false},
		{"get swingbench", "swingbench", false},
		{"get hammerdb", "hammerdb", false},
		{"get tpcc (not registered)", "tpcc", true},
		{"get unknown tool", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.GetByTool(tt.tool)
			if (got == nil) != tt.wantNil {
				t.Errorf("GetByTool() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

// TestAdapterRegistry_List tests listing all adapter types.
func TestAdapterRegistry_List(t *testing.T) {
	registry := NewAdapterRegistry()

	// Initially empty
	types := registry.List()
	if len(types) != 0 {
		t.Errorf("List() length = %d, want 0", len(types))
	}

	// Register adapters
	sysbench := &mockBenchmarkAdapter{adapterType: AdapterTypeSysbench}
	hammerdb := &mockBenchmarkAdapter{adapterType: AdapterTypeHammerDB}
	registry.Register(sysbench)
	registry.Register(hammerdb)

	// List should return 2 adapters
	types = registry.List()
	if len(types) != 2 {
		t.Errorf("List() length = %d, want 2", len(types))
	}

	// Verify types are present
	typeMap := make(map[AdapterType]bool)
	for _, typ := range types {
		typeMap[typ] = true
	}

	if !typeMap[AdapterTypeSysbench] {
		t.Error("List() missing sysbench adapter")
	}
	if !typeMap[AdapterTypeHammerDB] {
		t.Error("List() missing hammerdb adapter")
	}
}

// TestBenchmarkAdapter_BuildPrepareCommand tests command building.
func TestBenchmarkAdapter_BuildPrepareCommand(t *testing.T) {
	ctx := context.Background()
	adapter := &mockBenchmarkAdapter{}

	config := &Config{
		WorkDir: "/tmp/work",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, config)
	if err != nil {
		t.Fatalf("BuildPrepareCommand() failed: %v", err)
	}

	if cmd.CmdLine != "prepare" {
		t.Errorf("CmdLine = %s, want 'prepare'", cmd.CmdLine)
	}
	if cmd.WorkDir != "/tmp/work" {
		t.Errorf("WorkDir = %s, want '/tmp/work'", cmd.WorkDir)
	}
}

// TestConfig tests Config structure.
func TestConfig(t *testing.T) {
	config := Config{
		WorkDir: "/tmp/work",
		Parameters: map[string]interface{}{
			"threads": 8,
			"time":    60,
		},
		Options: execution.TaskOptions{
			WarmupTime:     30,
			SampleInterval: 1 * time.Second,
		},
	}

	if config.WorkDir != "/tmp/work" {
		t.Errorf("WorkDir = %s, want '/tmp/work'", config.WorkDir)
	}
	if config.Parameters["threads"] != 8 {
		t.Errorf("Parameters[threads] = %v, want 8", config.Parameters["threads"])
	}
}

// TestCommand tests Command structure.
func TestCommand(t *testing.T) {
	cmd := Command{
		CmdLine: "sysbench oltp run",
		WorkDir: "/tmp/work",
		Env:     []string{"PATH=/usr/bin", "HOME=/root"},
	}

	if cmd.CmdLine != "sysbench oltp run" {
		t.Errorf("CmdLine = %s, want 'sysbench oltp run'", cmd.CmdLine)
	}
	if len(cmd.Env) != 2 {
		t.Errorf("Env length = %d, want 2", len(cmd.Env))
	}
}

// TestResult tests Result structure.
func TestResult(t *testing.T) {
	result := Result{
		TPS:            1000.5,
		LatencyAvg:     5.2,
		LatencyP95:     10.5,
		LatencyP99:     25.0,
		TotalQueries:   50000,
		TotalErrors:    10,
		ErrorRate:      0.02,
		Duration:       60 * time.Second,
		RawOutput:      "test output",
	}

	if result.TPS != 1000.5 {
		t.Errorf("TPS = %v, want 1000.5", result.TPS)
	}
	if result.LatencyAvg != 5.2 {
		t.Errorf("LatencyAvg = %v, want 5.2", result.LatencyAvg)
	}
	if result.RawOutput != "test output" {
		t.Errorf("RawOutput = %s, want 'test output'", result.RawOutput)
	}
}

// TestSample tests Sample structure.
func TestSample(t *testing.T) {
	sample := Sample{
		Timestamp:   time.Now(),
		TPS:         1000.0,
		LatencyAvg:  5.0,
		LatencyP95:  10.0,
		LatencyP99:  20.0,
		ErrorRate:   0.1,
		ThreadCount: 8,
	}

	if sample.TPS != 1000.0 {
		t.Errorf("TPS = %v, want 1000.0", sample.TPS)
	}
	if sample.ThreadCount != 8 {
		t.Errorf("ThreadCount = %d, want 8", sample.ThreadCount)
	}
}

// TestProgressUpdate tests ProgressUpdate structure.
func TestProgressUpdate(t *testing.T) {
	update := ProgressUpdate{
		Phase:      "run",
		Timestamp:  time.Now(),
		Percentage: 50.5,
		Message:    "Half complete",
	}

	if update.Phase != "run" {
		t.Errorf("Phase = %s, want 'run'", update.Phase)
	}
	if update.Percentage != 50.5 {
		t.Errorf("Percentage = %v, want 50.5", update.Percentage)
	}
}

// TestAdapterType tests AdapterType constants.
func TestAdapterType(t *testing.T) {
	tests := []struct {
		adapterType AdapterType
		want        string
	}{
		{AdapterTypeSysbench, "sysbench"},
		{AdapterTypeSwingbench, "swingbench"},
		{AdapterTypeHammerDB, "hammerdb"},
		{AdapterTypeTPCC, "tpcc"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.adapterType) != tt.want {
				t.Errorf("AdapterType = %s, want %s", tt.adapterType, tt.want)
			}
		})
	}
}
