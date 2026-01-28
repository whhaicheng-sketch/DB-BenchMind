// Package repository provides settings repository implementation.
// Implements: Phase 7 - Settings Repository
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/config"
)

// SettingsRepository provides configuration persistence.
type SettingsRepository struct {
	configPath string
}

// NewSettingsRepository creates a new settings repository.
func NewSettingsRepository(configPath string) *SettingsRepository {
	return &SettingsRepository{
		configPath: configPath,
	}
}

// GetConfig loads the complete configuration.
func (r *SettingsRepository) GetConfig(ctx context.Context) (*config.Config, error) {
	// Check if config file exists
	if _, err := os.Stat(r.configPath); os.IsNotExist(err) {
		// Return default config
		return config.DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Parse JSON
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves the complete configuration.
func (r *SettingsRepository) SaveConfig(ctx context.Context, cfg *config.Config) error {
	// Validate before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(r.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(r.configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// GetToolPath returns the path for a specific tool.
func (r *SettingsRepository) GetToolPath(ctx context.Context, toolType config.ToolType) (string, error) {
	cfg, err := r.GetConfig(ctx)
	if err != nil {
		return "", err
	}

	path := cfg.GetToolPath(toolType)
	return path, nil
}

// SetToolPath sets the path for a specific tool.
func (r *SettingsRepository) SetToolPath(ctx context.Context, toolType config.ToolType, path string) error {
	cfg, err := r.GetConfig(ctx)
	if err != nil {
		return err
	}

	// Get existing tool config or create new
	toolCfg, ok := cfg.Tools[toolType]
	if !ok {
		toolCfg = config.ToolConfig{
			Type:    toolType,
			Enabled: false,
		}
	}

	// Update path
	toolCfg.Path = path

	// Save
	if err := cfg.SetToolConfig(toolCfg); err != nil {
		return err
	}

	return r.SaveConfig(ctx, cfg)
}

// IsToolEnabled checks if a tool is enabled.
func (r *SettingsRepository) IsToolEnabled(ctx context.Context, toolType config.ToolType) (bool, error) {
	cfg, err := r.GetConfig(ctx)
	if err != nil {
		return false, err
	}
	return cfg.IsToolEnabled(toolType), nil
}

// SetToolEnabled enables or disables a tool.
func (r *SettingsRepository) SetToolEnabled(ctx context.Context, toolType config.ToolType, enabled bool) error {
	cfg, err := r.GetConfig(ctx)
	if err != nil {
		return err
	}

	// Get existing tool config or create new
	toolCfg, ok := cfg.Tools[toolType]
	if !ok {
		toolCfg = config.ToolConfig{
			Type: toolType,
		}
	}

	// Update enabled status
	toolCfg.Enabled = enabled

	// Save
	if err := cfg.SetToolConfig(toolCfg); err != nil {
		return err
	}

	return r.SaveConfig(ctx, cfg)
}

// SetToolVersion sets the detected version for a tool.
func (r *SettingsRepository) SetToolVersion(ctx context.Context, toolType config.ToolType, version string) error {
	cfg, err := r.GetConfig(ctx)
	if err != nil {
		return err
	}

	// Get existing tool config or create new
	toolCfg, ok := cfg.Tools[toolType]
	if !ok {
		toolCfg = config.ToolConfig{
			Type: toolType,
		}
	}

	// Update version
	toolCfg.Version = version

	// Save
	if err := cfg.SetToolConfig(toolCfg); err != nil {
		return err
	}

	return r.SaveConfig(ctx, cfg)
}

// GetToolConfig returns the configuration for a specific tool.
func (r *SettingsRepository) GetToolConfig(ctx context.Context, toolType config.ToolType) (*config.ToolConfig, error) {
	cfg, err := r.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return cfg.GetToolConfig(toolType)
}

// ResetToDefaults resets configuration to defaults.
func (r *SettingsRepository) ResetToDefaults(ctx context.Context) error {
	// Delete config file to force defaults
	if err := os.Remove(r.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove config file: %w", err)
	}
	return nil
}

// GetConfigPath returns the configuration file path.
func (r *SettingsRepository) GetConfigPath() string {
	return r.configPath
}
