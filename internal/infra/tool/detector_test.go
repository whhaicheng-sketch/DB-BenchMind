// Package tool provides unit tests for tool detection.
package tool

import (
	"context"
	"os/exec"
	"runtime"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/config"
)

// TestDetector_getExecutableName tests executable name mapping.
func TestDetector_getExecutableName(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name     string
		toolType config.ToolType
		want     string
	}{
		{"sysbench", config.ToolTypeSysbench, "sysbench"},
		{"swingbench", config.ToolTypeSwingbench, "swingbench"},
		{"hammerdb linux", config.ToolTypeHammerDB, "hammerdbcli"},
	}

	// Add special case for Windows
	if runtime.GOOS == "windows" {
		tests[2].name = "hammerdb windows"
		tests[2].want = "hammerdbcli.bat"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.getExecutableName(tt.toolType); got != tt.want {
				t.Errorf("getExecutableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetector_parseVersion tests version parsing.
func TestDetector_parseVersion(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name     string
		toolType config.ToolType
		output   string
		want     string
	}{
		{
			name:     "sysbench version",
			toolType: config.ToolTypeSysbench,
			output:   "sysbench 1.0.20",
			want:     "1.0.20",
		},
		{
			name:     "sysbench with extra text",
			toolType: config.ToolTypeSysbench,
			output:   "sysbench 1.0.20 (compiled with ...)",
			want:     "1.0.20",
		},
		{
			name:     "hammerdb version",
			toolType: config.ToolTypeHammerDB,
			output:   "HammerDB CLI v4.6",
			want:     "4.6",
		},
		{
			name:     "swingbench version",
			toolType: config.ToolTypeSwingbench,
			output:   "Swingbench V2.5.1234",
			want:     "2.5.1234",
		},
		{
			name:     "empty output",
			toolType: config.ToolTypeSysbench,
			output:   "",
			want:     "",
		},
		{
			name:     "malformed output",
			toolType: config.ToolTypeSysbench,
			output:   "invalid output",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.parseVersion(tt.toolType, tt.output); got != tt.want {
				t.Errorf("parseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetector_DetectTool tests tool detection.
func TestDetector_DetectTool(t *testing.T) {
	ctx := context.Background()
	d := NewDetector()

	// Test with a tool that should exist on most systems
	// Using "ls" as a proxy since benchmark tools might not be installed
	path, err := d.DetectTool(ctx, config.ToolTypeSysbench)

	// We expect this might fail if sysbench is not installed
	// Just verify the function doesn't crash
	if err != nil {
		t.Logf("Sysbench not found (expected if not installed): %v", err)
	} else {
		if path == "" {
			t.Error("Path should not be empty when tool is found")
		}
		t.Logf("Found sysbench at: %s", path)
	}
}

// TestDetector_GetToolVersion tests version detection.
func TestDetector_GetToolVersion(t *testing.T) {
	ctx := context.Background()
	d := NewDetector()

	// Test with sysbench if available
	version, err := d.GetToolVersion(ctx, config.ToolTypeSysbench)

	if err != nil {
		t.Logf("Sysbench version detection failed (expected if not installed): %v", err)
	} else {
		if version == "" {
			t.Error("Version should not be empty when detection succeeds")
		}
		t.Logf("Sysbench version: %s", version)
	}
}

// TestDetector_DetectAllTools tests detecting all tools.
func TestDetector_DetectAllTools(t *testing.T) {
	ctx := context.Background()
	d := NewDetector()

	results := d.DetectAllTools(ctx)

	if len(results) != 3 {
		t.Errorf("DetectAllTools() returned %d results, want 3", len(results))
	}

	for toolType, info := range results {
		if info.Type != toolType {
			t.Errorf("ToolInfo.Type = %v, want %v", info.Type, toolType)
		}

		if info.Found {
			if info.Path == "" {
				t.Errorf("Path should not be empty when %s is found", toolType)
			}
			t.Logf("%s: found at %s, version %s", toolType, info.Path, info.Version)
		} else {
			t.Logf("%s: not found", toolType)
		}
	}
}

// TestDetector_CheckAvailability tests checking availability of a tool path.
func TestDetector_CheckAvailability(t *testing.T) {
	d := NewDetector()

	// Test with a system command that should exist
	if runtime.GOOS == "windows" {
		if err := d.CheckAvailability("cmd.exe"); err != nil {
			t.Errorf("CheckAvailability(cmd.exe) failed: %v", err)
		}
	} else {
		if err := d.CheckAvailability("/bin/sh"); err != nil {
			t.Errorf("CheckAvailability(/bin/sh) failed: %v", err)
		}
	}

	// Test with non-existent path
	err := d.CheckAvailability("/nonexistent/tool")
	if err == nil {
		t.Error("CheckAvailability() should fail for non-existent tool")
	}
}

// TestDetector_DetectToolsAsync tests async tool detection.
func TestDetector_DetectToolsAsync(t *testing.T) {
	ctx := context.Background()
	d := NewDetector()

	resultCh := d.DetectToolsAsync(ctx)

	count := 0
	for info := range resultCh {
		count++
		if info.Type == "" {
			t.Error("ToolInfo.Type should not be empty")
		}

		if info.Found && info.Path == "" {
			t.Errorf("Path should not be empty when %s is found", info.Type)
		}

		t.Logf("Tool: %s, Found: %v, Path: %s, Version: %s",
			info.Type, info.Found, info.Path, info.Version)
	}

	if count != 3 {
		t.Errorf("Received %d results, want 3", count)
	}
}

// TestDetector_DetectRealCommand tests detection with a real command.
func TestDetector_DetectRealCommand(t *testing.T) {
	d := NewDetector()

	// Use "echo" as a test command that should always exist
	// We'll temporarily treat it as a tool
	path, err := exec.LookPath("echo")
	if err != nil {
		t.Skip("echo command not found")
	}

	if path == "" {
		t.Error("echo path should not be empty")
	}

	// Verify the path is executable
	if err := d.CheckAvailability(path); err != nil {
		t.Errorf("CheckAvailability() failed for echo: %v", err)
	}
}

// TestDetector_ParseVersion_EdgeCases tests edge cases in version parsing.
func TestDetector_ParseVersion_EdgeCases(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name     string
		toolType config.ToolType
		output   string
		want     string
	}{
		{
			name:     "sysbench with trailing newline",
			toolType: config.ToolTypeSysbench,
			output:   "sysbench 1.0.20\n",
			want:     "1.0.20",
		},
		{
			name:     "sysbench with spaces",
			toolType: config.ToolTypeSysbench,
			output:   "  sysbench   1.0.20  ",
			want:     "1.0.20",
		},
		{
			name:     "swingbench multiline",
			toolType: config.ToolTypeSwingbench,
			output:   "Help text...\nSwingbench V2.5.1234 - CLI tool\nMore text...",
			want:     "2.5.1234",
		},
		{
			name:     "hammerdb with extra text",
			toolType: config.ToolTypeHammerDB,
			output:   "Welcome to HammerDB CLI v4.6 - Type help for commands",
			want:     "4.6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.parseVersion(tt.toolType, tt.output); got != tt.want {
				t.Errorf("parseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetector_DetectInvalidTool tests detecting invalid tool type.
func TestDetector_DetectInvalidTool(t *testing.T) {
	ctx := context.Background()
	d := NewDetector()

	// Try to detect an invalid tool type
	_, err := d.DetectTool(ctx, config.ToolType("invalid"))
	if err == nil {
		t.Error("DetectTool() should fail for invalid tool type")
	}
}

// BenchmarkDetector_DetectAllTools benchmarks detecting all tools.
func BenchmarkDetector_DetectAllTools(b *testing.B) {
	ctx := context.Background()
	d := NewDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.DetectAllTools(ctx)
	}
}

// BenchmarkDetector_ParseVersion benchmarks version parsing.
func BenchmarkDetector_ParseVersion(b *testing.B) {
	d := NewDetector()
	output := "sysbench 1.0.20 (compiled with GCC 9.3.0)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.parseVersion(config.ToolTypeSysbench, output)
	}
}
