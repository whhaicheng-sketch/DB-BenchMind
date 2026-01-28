// Package usecase provides unit tests for settings use case.
package usecase

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/config"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/tool"
)

// mockSettingsRepository is a mock implementation of SettingsRepository for testing.
type mockSettingsRepository struct {
	configPath string
}

func newMockSettingsRepository(configPath string) *mockSettingsRepository {
	return &mockSettingsRepository{configPath: configPath}
}

func (m *mockSettingsRepository) GetConfig(ctx context.Context) (*config.Config, error) {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return config.DefaultConfig(), nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (m *mockSettingsRepository) SaveConfig(ctx context.Context, cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

func (m *mockSettingsRepository) GetToolPath(ctx context.Context, toolType config.ToolType) (string, error) {
	cfg, err := m.GetConfig(ctx)
	if err != nil {
		return "", err
	}
	return cfg.GetToolPath(toolType), nil
}

func (m *mockSettingsRepository) SetToolPath(ctx context.Context, toolType config.ToolType, path string) error {
	cfg, err := m.GetConfig(ctx)
	if err != nil {
		return err
	}

	toolCfg := config.ToolConfig{
		Type:    toolType,
		Path:    path,
		Enabled: cfg.IsToolEnabled(toolType),
	}

	if err := cfg.SetToolConfig(toolCfg); err != nil {
		return err
	}

	return m.SaveConfig(ctx, cfg)
}

func (m *mockSettingsRepository) IsToolEnabled(ctx context.Context, toolType config.ToolType) (bool, error) {
	cfg, err := m.GetConfig(ctx)
	if err != nil {
		return false, err
	}
	return cfg.IsToolEnabled(toolType), nil
}

func (m *mockSettingsRepository) SetToolEnabled(ctx context.Context, toolType config.ToolType, enabled bool) error {
	cfg, err := m.GetConfig(ctx)
	if err != nil {
		return err
	}

	toolCfg := config.ToolConfig{
		Type:    toolType,
		Enabled: enabled,
	}

	if err := cfg.SetToolConfig(toolCfg); err != nil {
		return err
	}

	return m.SaveConfig(ctx, cfg)
}

func (m *mockSettingsRepository) SetToolVersion(ctx context.Context, toolType config.ToolType, version string) error {
	cfg, err := m.GetConfig(ctx)
	if err != nil {
		return err
	}

	toolCfg := config.ToolConfig{
		Type:    toolType,
		Version: version,
		Enabled: cfg.IsToolEnabled(toolType),
	}

	if err := cfg.SetToolConfig(toolCfg); err != nil {
		return err
	}

	return m.SaveConfig(ctx, cfg)
}

func (m *mockSettingsRepository) GetToolConfig(ctx context.Context, toolType config.ToolType) (*config.ToolConfig, error) {
	cfg, err := m.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return cfg.GetToolConfig(toolType)
}

func (m *mockSettingsRepository) ResetToDefaults(ctx context.Context) error {
	if err := os.Remove(m.configPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// setupSettingsTest creates a test settings use case.
func setupSettingsTest(t *testing.T) *SettingsUseCase {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	settingsRepo := newMockSettingsRepository(configPath)
	detector := tool.NewDetector()

	return NewSettingsUseCase(settingsRepo, detector)
}

// TestSettingsUseCase_GetConfig tests getting configuration.
func TestSettingsUseCase_GetConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	cfg, err := uc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Should return default config
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}
}

// TestSettingsUseCase_UpdateConfig tests updating configuration.
func TestSettingsUseCase_UpdateConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	cfg := config.DefaultConfig()
	cfg.UI.Theme = "dark"
	cfg.Database.MaxOpenConns = 100

	if err := uc.UpdateConfig(ctx, cfg); err != nil {
		t.Fatalf("UpdateConfig() failed: %v", err)
	}

	// Verify
	loaded, err := uc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() after update failed: %v", err)
	}

	if loaded.UI.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", loaded.UI.Theme)
	}

	if loaded.Database.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns = %d, want 100", loaded.Database.MaxOpenConns)
	}
}

// TestSettingsUseCase_UpdateConfig_Invalid tests updating with invalid config.
func TestSettingsUseCase_UpdateConfig_Invalid(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	cfg := &config.Config{
		Version: 999, // Invalid
	}

	if err := uc.UpdateConfig(ctx, cfg); err == nil {
		t.Error("UpdateConfig() with invalid config should fail")
	}
}

// TestSettingsUseCase_GetToolConfig tests getting tool configuration.
func TestSettingsUseCase_GetToolConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	toolCfg, err := uc.GetToolConfig(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("GetToolConfig() failed: %v", err)
	}

	if toolCfg.Type != config.ToolTypeSysbench {
		t.Errorf("Tool type = %v, want sysbench", toolCfg.Type)
	}
}

// TestSettingsUseCase_SetToolPath tests setting tool path.
func TestSettingsUseCase_SetToolPath(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	// Create a temp executable
	tmpFile := t.TempDir() + "/sysbench"
	if err := os.WriteFile(tmpFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := uc.SetToolPath(ctx, config.ToolTypeSysbench, tmpFile); err != nil {
		t.Fatalf("SetToolPath() failed: %v", err)
	}

	// Verify
	path, err := uc.GetToolPath(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("GetToolPath() failed: %v", err)
	}

	if path != tmpFile {
		t.Errorf("Tool path = %s, want %s", path, tmpFile)
	}
}

// TestSettingsUseCase_SetToolEnabled tests enabling/disabling tools.
func TestSettingsUseCase_SetToolEnabled(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	// Disable sysbench
	if err := uc.SetToolEnabled(ctx, config.ToolTypeSysbench, false); err != nil {
		t.Fatalf("SetToolEnabled() failed: %v", err)
	}

	// Verify
	enabled, err := uc.IsToolEnabled(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("IsToolEnabled() failed: %v", err)
	}

	if enabled {
		t.Error("Sysbench should be disabled")
	}
}

// TestSettingsUseCase_DetectTools tests detecting all tools.
func TestSettingsUseCase_DetectTools(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	toolInfos := uc.DetectTools(ctx)

	if len(toolInfos) != 3 {
		t.Errorf("DetectTools() returned %d results, want 3", len(toolInfos))
	}

	for toolType, info := range toolInfos {
		if info.Type != toolType {
			t.Errorf("ToolInfo.Type = %v, want %v", info.Type, toolType)
		}

		if info.Found {
			t.Logf("%s: found at %s, version %s", toolType, info.Path, info.Version)
		} else {
			t.Logf("%s: not found", toolType)
		}
	}
}

// TestSettingsUseCase_DetectTool tests detecting a specific tool.
func TestSettingsUseCase_DetectTool(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	info, err := uc.DetectTool(ctx, config.ToolTypeSysbench)
	if err != nil {
		t.Fatalf("DetectTool() failed: %v", err)
	}

	if info.Type != config.ToolTypeSysbench {
		t.Errorf("Tool type = %v, want sysbench", info.Type)
	}

	if info.Found {
		t.Logf("Sysbench found at %s, version %s", info.Path, info.Version)
	} else {
		t.Log("Sysbench not found (may not be installed)")
	}
}

// TestSettingsUseCase_DetectAndSaveTools tests detecting and saving tools.
func TestSettingsUseCase_DetectAndSaveTools(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	toolInfos, err := uc.DetectAndSaveTools(ctx)
	if err != nil {
		t.Fatalf("DetectAndSaveTools() failed: %v", err)
	}

	if len(toolInfos) != 3 {
		t.Errorf("DetectAndSaveTools() returned %d results, want 3", len(toolInfos))
	}

	// Verify config was updated
	cfg, err := uc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// At least sysbench should be in the config
	if _, ok := cfg.Tools[config.ToolTypeSysbench]; !ok {
		t.Error("Sysbench should be in config")
	}
}

// TestSettingsUseCase_ResetSettings tests resetting to defaults.
func TestSettingsUseCase_ResetSettings(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	// Save custom config
	cfg := config.DefaultConfig()
	cfg.UI.Theme = "dark"
	if err := uc.UpdateConfig(ctx, cfg); err != nil {
		t.Fatalf("UpdateConfig() failed: %v", err)
	}

	// Reset
	if err := uc.ResetSettings(ctx); err != nil {
		t.Fatalf("ResetSettings() failed: %v", err)
	}

	// Verify
	loaded, err := uc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() after reset failed: %v", err)
	}

	if loaded.UI.Theme != "auto" {
		t.Errorf("Theme after reset = %s, want auto", loaded.UI.Theme)
	}
}

// TestSettingsUseCase_GetDatabaseConfig tests getting database config.
func TestSettingsUseCase_GetDatabaseConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	dbCfg, err := uc.GetDatabaseConfig(ctx)
	if err != nil {
		t.Fatalf("GetDatabaseConfig() failed: %v", err)
	}

	if dbCfg.Path == "" {
		t.Error("Database path should not be empty")
	}

	if dbCfg.MaxOpenConns == 0 {
		t.Error("MaxOpenConns should not be 0")
	}
}

// TestSettingsUseCase_UpdateDatabaseConfig tests updating database config.
func TestSettingsUseCase_UpdateDatabaseConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	newCfg := config.DatabaseConfig{
		Path:            "/tmp/test.db",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 600,
	}

	if err := uc.UpdateDatabaseConfig(ctx, newCfg); err != nil {
		t.Fatalf("UpdateDatabaseConfig() failed: %v", err)
	}

	// Verify
	dbCfg, err := uc.GetDatabaseConfig(ctx)
	if err != nil {
		t.Fatalf("GetDatabaseConfig() failed: %v", err)
	}

	if dbCfg.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns = %d, want 50", dbCfg.MaxOpenConns)
	}
}

// TestSettingsUseCase_UpdateDatabaseConfig_Invalid tests invalid database config.
func TestSettingsUseCase_UpdateDatabaseConfig_Invalid(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	invalidCfg := config.DatabaseConfig{
		Path:         "", // Missing path
		MaxOpenConns: 0, // Invalid
	}

	if err := uc.UpdateDatabaseConfig(ctx, invalidCfg); err == nil {
		t.Error("UpdateDatabaseConfig() with invalid config should fail")
	}
}

// TestSettingsUseCase_GetReportConfig tests getting report config.
func TestSettingsUseCase_GetReportConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	reportCfg, err := uc.GetReportConfig(ctx)
	if err != nil {
		t.Fatalf("GetReportConfig() failed: %v", err)
	}

	if reportCfg.DefaultFormat == "" {
		t.Error("Default format should not be empty")
	}
}

// TestSettingsUseCase_UpdateReportConfig tests updating report config.
func TestSettingsUseCase_UpdateReportConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	newCfg := config.ReportConfig{
		DefaultFormat: "json",
		IncludeCharts: false,
		ChartWidth:    80,
		ChartHeight:   15,
	}

	if err := uc.UpdateReportConfig(ctx, newCfg); err != nil {
		t.Fatalf("UpdateReportConfig() failed: %v", err)
	}

	// Verify
	reportCfg, err := uc.GetReportConfig(ctx)
	if err != nil {
		t.Fatalf("GetReportConfig() failed: %v", err)
	}

	if reportCfg.DefaultFormat != "json" {
		t.Errorf("DefaultFormat = %s, want json", reportCfg.DefaultFormat)
	}

	if reportCfg.ChartWidth != 80 {
		t.Errorf("ChartWidth = %d, want 80", reportCfg.ChartWidth)
	}
}

// TestSettingsUseCase_GetUIConfig tests getting UI config.
func TestSettingsUseCase_GetUIConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	uiCfg, err := uc.GetUIConfig(ctx)
	if err != nil {
		t.Fatalf("GetUIConfig() failed: %v", err)
	}

	if uiCfg.Theme == "" {
		t.Error("Theme should not be empty")
	}
}

// TestSettingsUseCase_UpdateUIConfig tests updating UI config.
func TestSettingsUseCase_UpdateUIConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	newCfg := config.UIConfig{
		Theme:           "dark",
		Language:        "en",
		AutoSave:        false,
		RefreshInterval: 10,
	}

	if err := uc.UpdateUIConfig(ctx, newCfg); err != nil {
		t.Fatalf("UpdateUIConfig() failed: %v", err)
	}

	// Verify
	uiCfg, err := uc.GetUIConfig(ctx)
	if err != nil {
		t.Fatalf("GetUIConfig() failed: %v", err)
	}

	if uiCfg.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", uiCfg.Theme)
	}
}

// TestSettingsUseCase_GetAdvancedConfig tests getting advanced config.
func TestSettingsUseCase_GetAdvancedConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	advCfg, err := uc.GetAdvancedConfig(ctx)
	if err != nil {
		t.Fatalf("GetAdvancedConfig() failed: %v", err)
	}

	if advCfg.LogLevel == "" {
		t.Error("Log level should not be empty")
	}
}

// TestSettingsUseCase_UpdateAdvancedConfig tests updating advanced config.
func TestSettingsUseCase_UpdateAdvancedConfig(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	newCfg := config.AdvancedConfig{
		LogLevel:     "debug",
		MaxLogFiles:  20,
		WorkDir:      "/tmp/dbbench",
		Timeout:      120,
	}

	if err := uc.UpdateAdvancedConfig(ctx, newCfg); err != nil {
		t.Fatalf("UpdateAdvancedConfig() failed: %v", err)
	}

	// Verify
	advCfg, err := uc.GetAdvancedConfig(ctx)
	if err != nil {
		t.Fatalf("GetAdvancedConfig() failed: %v", err)
	}

	if advCfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %s, want debug", advCfg.LogLevel)
	}

	if advCfg.MaxLogFiles != 20 {
		t.Errorf("MaxLogFiles = %d, want 20", advCfg.MaxLogFiles)
	}
}

// TestSettingsUseCase_GetEnabledTools tests getting enabled tools list.
func TestSettingsUseCase_GetEnabledTools(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	tools, err := uc.GetEnabledTools(ctx)
	if err != nil {
		t.Fatalf("GetEnabledTools() failed: %v", err)
	}

	// By default, only sysbench is enabled
	if len(tools) != 1 {
		t.Errorf("Enabled tools count = %d, want 1", len(tools))
	}

	if tools[0] != config.ToolTypeSysbench {
		t.Errorf("Expected sysbench to be enabled, got %s", tools[0])
	}
}

// TestSettingsUseCase_VerifyTool tests verifying a tool.
func TestSettingsUseCase_VerifyTool(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	// Create a temp executable
	tmpFile := t.TempDir() + "/test-tool"
	if err := os.WriteFile(tmpFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set tool path
	if err := uc.SetToolPath(ctx, config.ToolTypeSysbench, tmpFile); err != nil {
		t.Fatalf("SetToolPath() failed: %v", err)
	}

	// Verify
	if err := uc.VerifyTool(ctx, config.ToolTypeSysbench); err != nil {
		t.Errorf("VerifyTool() failed: %v", err)
	}
}

// TestSettingsUseCase_VerifyTool_NotConfigured tests verifying unconfigured tool.
func TestSettingsUseCase_VerifyTool_NotConfigured(t *testing.T) {
	ctx := context.Background()
	uc := setupSettingsTest(t)

	// Disable tool to remove path
	if err := uc.SetToolEnabled(ctx, config.ToolTypeSwingbench, false); err != nil {
		t.Fatalf("SetToolEnabled() failed: %v", err)
	}

	// Try to verify - should fail
	err := uc.VerifyTool(ctx, config.ToolTypeSwingbench)
	if err == nil {
		t.Error("VerifyTool() should fail for tool with no path")
	}
}
