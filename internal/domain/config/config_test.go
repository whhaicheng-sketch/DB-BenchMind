// Package config provides unit tests for configuration domain models.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestToolType_Validate tests tool type validation.
func TestToolType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tool    ToolType
		wantErr bool
	}{
		{"valid sysbench", ToolTypeSysbench, false},
		{"valid swingbench", ToolTypeSwingbench, false},
		{"valid hammerdb", ToolTypeHammerDB, false},
		{"invalid tool", ToolType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tool.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ToolType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestToolConfig_Validate tests tool configuration validation.
func TestToolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ToolConfig
		wantErr bool
	}{
		{
			name: "valid config without path",
			config: ToolConfig{
				Type:    ToolTypeSysbench,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "valid config with path",
			config: ToolConfig{
				Type:    ToolTypeSysbench,
				Path:    "/usr/bin/sysbench",
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			config: ToolConfig{
				Type: ToolType("invalid"),
			},
			wantErr: true,
		},
		{
			name: "relative path",
			config: ToolConfig{
				Type: ToolTypeSysbench,
				Path: "relative/path",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ToolConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestToolConfig_Validate_WithRealFile tests validation with real file.
func TestToolConfig_Validate_WithRealFile(t *testing.T) {
	// Create a temporary executable file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-tool")

	// Create and make executable
	if err := os.WriteFile(tmpFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := ToolConfig{
		Type:    ToolTypeSysbench,
		Path:    tmpFile,
		Enabled: true,
	}

	if err := config.Validate(); err != nil {
		t.Errorf("ToolConfig.Validate() with valid executable failed: %v", err)
	}

	// Test with non-executable file
	configNotExec := ToolConfig{
		Type: ToolTypeSysbench,
		Path: tmpFile,
	}
	if err := os.Chmod(tmpFile, 0644); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}

	if err := configNotExec.Validate(); err == nil {
		t.Error("ToolConfig.Validate() should fail with non-executable file")
	}
}

// TestDatabaseConfig_Validate tests database configuration validation.
func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: DatabaseConfig{
				Path:            "/tmp/test.db",
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 300,
			},
			wantErr: false,
		},
		{
			name: "missing path",
			config: DatabaseConfig{
				MaxOpenConns: 25,
			},
			wantErr: true,
		},
		{
			name: "max_open_conns too small",
			config: DatabaseConfig{
				Path:         "/tmp/test.db",
				MaxOpenConns: 0,
			},
			wantErr: true,
		},
		{
			name: "negative max_idle_conns",
			config: DatabaseConfig{
				Path:         "/tmp/test.db",
				MaxOpenConns: 25,
				MaxIdleConns: -1,
			},
			wantErr: true,
		},
		{
			name: "max_idle exceeds max_open",
			config: DatabaseConfig{
				Path:         "/tmp/test.db",
				MaxOpenConns: 5,
				MaxIdleConns: 10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("DatabaseConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReportConfig_Validate tests report configuration validation.
func TestReportConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ReportConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ReportConfig{
				DefaultFormat: "markdown",
				IncludeCharts: true,
				ChartWidth:    60,
				ChartHeight:   10,
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			config: ReportConfig{
				DefaultFormat: "invalid",
			},
			wantErr: true,
		},
		{
			name: "chart_width too small",
			config: ReportConfig{
				DefaultFormat: "markdown",
				ChartWidth:    10,
			},
			wantErr: true,
		},
		{
			name: "chart_height too large",
			config: ReportConfig{
				DefaultFormat: "markdown",
				ChartWidth:    60,
				ChartHeight:   100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ReportConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUIConfig_Validate tests UI configuration validation.
func TestUIConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  UIConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: UIConfig{
				Theme:           "auto",
				Language:        "en",
				AutoSave:        true,
				RefreshInterval: 5,
			},
			wantErr: false,
		},
		{
			name: "invalid theme",
			config: UIConfig{
				Theme: "invalid",
			},
			wantErr: true,
		},
		{
			name: "refresh_interval too small",
			config: UIConfig{
				Theme:           "auto",
				RefreshInterval: 0,
			},
			wantErr: true,
		},
		{
			name: "refresh_interval too large",
			config: UIConfig{
				Theme:           "auto",
				RefreshInterval: 100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("UIConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAdvancedConfig_Validate tests advanced configuration validation.
func TestAdvancedConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  AdvancedConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AdvancedConfig{
				LogLevel:    "info",
				MaxLogFiles: 10,
				WorkDir:     "/tmp/work",
				Timeout:     60,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: AdvancedConfig{
				LogLevel: "invalid",
			},
			wantErr: true,
		},
		{
			name: "max_log_files too large",
			config: AdvancedConfig{
				LogLevel:    "info",
				MaxLogFiles: 200,
			},
			wantErr: true,
		},
		{
			name: "timeout too small",
			config: AdvancedConfig{
				LogLevel: "info",
				Timeout:  0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("AdvancedConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfig_Validate tests complete configuration validation.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "custom valid config",
			config: &Config{
				Version: 1,
				Database: DatabaseConfig{
					Path:         "/tmp/test.db",
					MaxOpenConns: 25,
				},
				Tools: map[ToolType]ToolConfig{
					ToolTypeSysbench: {
						Type:    ToolTypeSysbench,
						Enabled: true,
					},
				},
				Reports: ReportConfig{
					DefaultFormat: "markdown",
					ChartWidth:    60,
					ChartHeight:   10,
				},
				UI: UIConfig{
					Theme:           "auto",
					RefreshInterval: 5,
				},
				Advanced: AdvancedConfig{
					LogLevel: "info",
					Timeout:  60,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			config: &Config{
				Version: 2,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultConfig tests default configuration.
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}

	if config.Database.Path == "" {
		t.Error("Database path should not be empty")
	}

	if config.Database.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %d, want 25", config.Database.MaxOpenConns)
	}

	if len(config.Tools) != 3 {
		t.Errorf("Tools count = %d, want 3", len(config.Tools))
	}

	if config.Reports.DefaultFormat != "markdown" {
		t.Errorf("DefaultFormat = %s, want markdown", config.Reports.DefaultFormat)
	}

	if !config.UI.AutoSave {
		t.Error("AutoSave should be true by default")
	}

	if config.Advanced.Timeout != 60 {
		t.Errorf("Timeout = %d, want 60", config.Advanced.Timeout)
	}
}

// TestConfig_GetToolConfig tests getting tool configuration.
func TestConfig_GetToolConfig(t *testing.T) {
	config := DefaultConfig()

	// Get existing tool
	tool, err := config.GetToolConfig(ToolTypeSysbench)
	if err != nil {
		t.Fatalf("GetToolConfig(sysbench) failed: %v", err)
	}

	if tool.Type != ToolTypeSysbench {
		t.Errorf("Tool type = %v, want sysbench", tool.Type)
	}

	// Get non-existing tool
	_, err = config.GetToolConfig(ToolType("invalid"))
	if err == nil {
		t.Error("GetToolConfig(invalid) should return error")
	}
}

// TestConfig_SetToolConfig tests setting tool configuration.
func TestConfig_SetToolConfig(t *testing.T) {
	config := DefaultConfig()

	// Create a temporary file
	tmpFile := t.TempDir() + "/test-tool"
	if err := os.WriteFile(tmpFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	newConfig := ToolConfig{
		Type:    ToolTypeSysbench,
		Path:    tmpFile,
		Enabled: false,
	}

	if err := config.SetToolConfig(newConfig); err != nil {
		t.Fatalf("SetToolConfig failed: %v", err)
	}

	tool, _ := config.GetToolConfig(ToolTypeSysbench)
	if tool.Path != tmpFile {
		t.Errorf("Tool path = %s, want %s", tool.Path, tmpFile)
	}
	if tool.Enabled {
		t.Error("Tool should be disabled")
	}
}

// TestConfig_GetToolPath tests getting tool path.
func TestConfig_GetToolPath(t *testing.T) {
	config := DefaultConfig()

	// Sysbench is enabled by default
	path := config.GetToolPath(ToolTypeSysbench)
	if path != "" {
		t.Error("Default sysbench path should be empty")
	}

	// Swingbench is disabled
	path = config.GetToolPath(ToolTypeSwingbench)
	if path != "" {
		t.Error("Disabled tool should return empty path")
	}
}

// TestConfig_IsToolEnabled tests checking if tool is enabled.
func TestConfig_IsToolEnabled(t *testing.T) {
	config := DefaultConfig()

	if !config.IsToolEnabled(ToolTypeSysbench) {
		t.Error("Sysbench should be enabled by default")
	}

	if config.IsToolEnabled(ToolTypeSwingbench) {
		t.Error("Swingbench should be disabled by default")
	}
}
