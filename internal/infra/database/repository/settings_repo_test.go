// Package repository provides unit tests for settings repository.
package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/config"
)

// setupSettingsTestDB creates a test database for settings tests.
func setupSettingsTestDB(t *testing.T) string {
	t.Helper()

	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	return configPath
}

// TestSettingsRepository_GetConfig_Default tests getting default config.
func TestSettingsRepository_GetConfig_Default(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)

	repo := NewSettingsRepository(configPath)

	cfg, err := repo.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Should return default config
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}

	if cfg.Database.Path == "" {
		t.Error("Database path should not be empty in default config")
	}

	if len(cfg.Tools) != 3 {
		t.Errorf("Tools count = %d, want 3", len(cfg.Tools))
	}
}

// TestSettingsRepository_SaveConfig tests saving configuration.
func TestSettingsRepository_SaveConfig(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)

	repo := NewSettingsRepository(configPath)

	// Save custom config
	cfg := config.DefaultConfig()
	cfg.Database.MaxOpenConns = 50
	cfg.UI.Theme = "dark"

	if err := repo.SaveConfig(ctx, cfg); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify
	loaded, err := repo.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() after save failed: %v", err)
	}

	if loaded.Database.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns = %d, want 50", loaded.Database.MaxOpenConns)
	}

	if loaded.UI.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", loaded.UI.Theme)
	}
}

// TestSettingsRepository_SaveConfig_Invalid tests saving invalid config.
func TestSettingsRepository_SaveConfig_Invalid(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)

	repo := NewSettingsRepository(configPath)

	// Create invalid config
	cfg := &config.Config{
		Version: 999, // Invalid version
	}

	if err := repo.SaveConfig(ctx, cfg); err == nil {
		t.Error("SaveConfig() with invalid config should fail")
	}

	// Verify file was not created
	if _, err := os.Stat(configPath); err == nil {
		t.Error("Config file should not be created for invalid config")
	}
}

// TestSettingsRepository_GetToolPath tests getting tool path.
func TestSettingsRepository_GetToolPath(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)
	repo := NewSettingsRepository(configPath)

	// Default config has empty paths
	path, err := repo.GetToolPath(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("GetToolPath() failed: %v", err)
	}

	if path != "" {
		t.Errorf("Default tool path should be empty, got %s", path)
	}
}

// TestSettingsRepository_SetToolPath tests setting tool path.
func TestSettingsRepository_SetToolPath(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)
	repo := NewSettingsRepository(configPath)

	// Create a temp executable file
	tmpFile := t.TempDir() + "/sysbench"
	if err := os.WriteFile(tmpFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set tool path
	if err := repo.SetToolPath(ctx, config.ToolTypeSysbench, tmpFile); err != nil {
		t.Fatalf("SetToolPath() failed: %v", err)
	}

	// Verify
	path, err := repo.GetToolPath(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("GetToolPath() after SetToolPath failed: %v", err)
	}

	if path != tmpFile {
		t.Errorf("Tool path = %s, want %s", path, tmpFile)
	}
}

// TestSettingsRepository_IsToolEnabled tests checking if tool is enabled.
func TestSettingsRepository_IsToolEnabled(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)
	repo := NewSettingsRepository(configPath)

	// Sysbench is enabled by default
	enabled, err := repo.IsToolEnabled(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("IsToolEnabled() failed: %v", err)
	}

	if !enabled {
		t.Error("Sysbench should be enabled by default")
	}

	// Swingbench is disabled by default
	enabled, err = repo.IsToolEnabled(ctx, config.ToolTypeSwingbench)
	if err != nil {
		t.Fatalf("IsToolEnabled() for swingbench failed: %v", err)
	}

	if enabled {
		t.Error("Swingbench should be disabled by default")
	}
}

// TestSettingsRepository_SetToolEnabled tests enabling/disabling tools.
func TestSettingsRepository_SetToolEnabled(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)
	repo := NewSettingsRepository(configPath)

	// Disable sysbench
	if err := repo.SetToolEnabled(ctx, config.ToolTypeSysbench, false); err != nil {
		t.Fatalf("SetToolEnabled() failed: %v", err)
	}

	// Verify
	enabled, err := repo.IsToolEnabled(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("IsToolEnabled() after disable failed: %v", err)
	}

	if enabled {
		t.Error("Sysbench should be disabled")
	}

	// Enable swingbench
	if err := repo.SetToolEnabled(ctx, config.ToolTypeSwingbench, true); err != nil {
		t.Fatalf("SetToolEnabled() for swingbench failed: %v", err)
	}

	// Verify
	enabled, err = repo.IsToolEnabled(ctx, config.ToolTypeSwingbench)
	if err != nil {
		t.Fatalf("IsToolEnabled() after enable failed: %v", err)
	}

	if !enabled {
		t.Error("Swingbench should be enabled")
	}
}

// TestSettingsRepository_SetToolVersion tests setting tool version.
func TestSettingsRepository_SetToolVersion(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)
	repo := NewSettingsRepository(configPath)

	// Set version
	if err := repo.SetToolVersion(ctx, config.ToolTypeSysbench, "1.0.20"); err != nil {
		t.Fatalf("SetToolVersion() failed: %v", err)
	}

	// Verify
	toolCfg, err := repo.GetToolConfig(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("GetToolConfig() failed: %v", err)
	}

	if toolCfg.Version != "1.0.20" {
		t.Errorf("Tool version = %s, want 1.0.20", toolCfg.Version)
	}
}

// TestSettingsRepository_ResetToDefaults tests resetting to defaults.
func TestSettingsRepository_ResetToDefaults(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)
	repo := NewSettingsRepository(configPath)

	// Save custom config
	cfg := config.DefaultConfig()
	cfg.UI.Theme = "dark"
	if err := repo.SaveConfig(ctx, cfg); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Reset
	if err := repo.ResetToDefaults(ctx); err != nil {
		t.Fatalf("ResetToDefaults() failed: %v", err)
	}

	// Verify config file was deleted
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Config file should be deleted after reset")
	}

	// Verify we get defaults back
	loaded, err := repo.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() after reset failed: %v", err)
	}

	if loaded.UI.Theme != "auto" {
		t.Errorf("Theme after reset = %s, want auto", loaded.UI.Theme)
	}
}

// TestSettingsRepository_Persistence tests config persistence across reopens.
func TestSettingsRepository_Persistence(t *testing.T) {
	ctx := context.Background()
	configPath := setupSettingsTestDB(t)
	repo := NewSettingsRepository(configPath)

	// Save custom config
	cfg := config.DefaultConfig()
	cfg.Database.MaxOpenConns = 100
	cfg.Reports.ChartWidth = 80
	cfg.Advanced.LogLevel = "debug"

	if err := repo.SaveConfig(ctx, cfg); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Create new repository instance (simulating reopen)
	repo2 := NewSettingsRepository(configPath)

	// Load and verify
	loaded, err := repo2.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() after reopen failed: %v", err)
	}

	if loaded.Database.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns = %d, want 100", loaded.Database.MaxOpenConns)
	}

	if loaded.Reports.ChartWidth != 80 {
		t.Errorf("ChartWidth = %d, want 80", loaded.Reports.ChartWidth)
	}

	if loaded.Advanced.LogLevel != "debug" {
		t.Errorf("LogLevel = %s, want debug", loaded.Advanced.LogLevel)
	}
}
