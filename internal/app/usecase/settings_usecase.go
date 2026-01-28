// Package usecase provides settings management business logic.
// Implements: Phase 7 - Settings Management
package usecase

import (
	"context"
	"fmt"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/config"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/tool"
)

// SettingsUseCase provides settings management business operations.
type SettingsUseCase struct {
	settingsRepo SettingsRepository
	detector     *tool.Detector
}

// NewSettingsUseCase creates a new settings use case.
func NewSettingsUseCase(
	settingsRepo SettingsRepository,
	detector *tool.Detector,
) *SettingsUseCase {
	return &SettingsUseCase{
		settingsRepo: settingsRepo,
		detector:     detector,
	}
}

// GetConfig retrieves the current configuration.
func (uc *SettingsUseCase) GetConfig(ctx context.Context) (*config.Config, error) {
	return uc.settingsRepo.GetConfig(ctx)
}

// UpdateConfig updates the configuration.
func (uc *SettingsUseCase) UpdateConfig(ctx context.Context, cfg *config.Config) error {
	// Validate before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	return uc.settingsRepo.SaveConfig(ctx, cfg)
}

// GetToolConfig retrieves configuration for a specific tool.
func (uc *SettingsUseCase) GetToolConfig(ctx context.Context, toolType config.ToolType) (*config.ToolConfig, error) {
	return uc.settingsRepo.GetToolConfig(ctx, toolType)
}

// SetToolPath sets the path for a specific tool.
func (uc *SettingsUseCase) SetToolPath(ctx context.Context, toolType config.ToolType, path string) error {
	return uc.settingsRepo.SetToolPath(ctx, toolType, path)
}

// SetToolEnabled enables or disables a tool.
func (uc *SettingsUseCase) SetToolEnabled(ctx context.Context, toolType config.ToolType, enabled bool) error {
	return uc.settingsRepo.SetToolEnabled(ctx, toolType, enabled)
}

// DetectTools detects all benchmark tools on the system.
func (uc *SettingsUseCase) DetectTools(ctx context.Context) map[config.ToolType]*tool.ToolInfo {
	detector := tool.NewDetector()
	return detector.DetectAllTools(ctx)
}

// DetectTool detects a specific tool on the system.
func (uc *SettingsUseCase) DetectTool(ctx context.Context, toolType config.ToolType) (*tool.ToolInfo, error) {
	detector := tool.NewDetector()

	info := &tool.ToolInfo{
		Type:  toolType,
		Found: false,
	}

	// Check if tool exists
	path, err := detector.DetectTool(ctx, toolType)
	if err == nil {
		info.Found = true
		info.Path = path

		// Try to get version
		version, err := detector.GetToolVersion(ctx, toolType)
		if err == nil {
			info.Version = version
		}
	} else {
		info.Error = err.Error()
	}

	return info, nil
}

// DetectAndSaveTools detects all tools and saves their information to config.
func (uc *SettingsUseCase) DetectAndSaveTools(ctx context.Context) (map[config.ToolType]*tool.ToolInfo, error) {
	// Detect all tools
	toolInfos := uc.DetectTools(ctx)

	// Save to config
	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	// Update tool configs
	for toolType, info := range toolInfos {
		toolCfg := config.ToolConfig{
			Type:    toolType,
			Path:    info.Path,
			Version: info.Version,
			Enabled: info.Found, // Auto-enable detected tools
		}

		if err := cfg.SetToolConfig(toolCfg); err != nil {
			// Log but continue with other tools
			continue
		}

		// Save version if detected
		if info.Version != "" {
			if err := uc.settingsRepo.SetToolVersion(ctx, toolType, info.Version); err != nil {
				// Non-critical, continue
			}
		}
	}

	// Save updated config
	if err := uc.settingsRepo.SaveConfig(ctx, cfg); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	return toolInfos, nil
}

// ResetSettings resets all settings to defaults.
func (uc *SettingsUseCase) ResetSettings(ctx context.Context) error {
	return uc.settingsRepo.ResetToDefaults(ctx)
}

// GetDatabaseConfig retrieves database configuration.
func (uc *SettingsUseCase) GetDatabaseConfig(ctx context.Context) (*config.DatabaseConfig, error) {
	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &cfg.Database, nil
}

// UpdateDatabaseConfig updates database configuration.
func (uc *SettingsUseCase) UpdateDatabaseConfig(ctx context.Context, dbCfg config.DatabaseConfig) error {
	if err := dbCfg.Validate(); err != nil {
		return fmt.Errorf("validate database config: %w", err)
	}

	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	cfg.Database = dbCfg
	return uc.settingsRepo.SaveConfig(ctx, cfg)
}

// GetReportConfig retrieves report configuration.
func (uc *SettingsUseCase) GetReportConfig(ctx context.Context) (*config.ReportConfig, error) {
	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &cfg.Reports, nil
}

// UpdateReportConfig updates report configuration.
func (uc *SettingsUseCase) UpdateReportConfig(ctx context.Context, reportCfg config.ReportConfig) error {
	if err := reportCfg.Validate(); err != nil {
		return fmt.Errorf("validate report config: %w", err)
	}

	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	cfg.Reports = reportCfg
	return uc.settingsRepo.SaveConfig(ctx, cfg)
}

// GetUIConfig retrieves UI configuration.
func (uc *SettingsUseCase) GetUIConfig(ctx context.Context) (*config.UIConfig, error) {
	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &cfg.UI, nil
}

// UpdateUIConfig updates UI configuration.
func (uc *SettingsUseCase) UpdateUIConfig(ctx context.Context, uiCfg config.UIConfig) error {
	if err := uiCfg.Validate(); err != nil {
		return fmt.Errorf("validate UI config: %w", err)
	}

	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	cfg.UI = uiCfg
	return uc.settingsRepo.SaveConfig(ctx, cfg)
}

// GetAdvancedConfig retrieves advanced configuration.
func (uc *SettingsUseCase) GetAdvancedConfig(ctx context.Context) (*config.AdvancedConfig, error) {
	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &cfg.Advanced, nil
}

// UpdateAdvancedConfig updates advanced configuration.
func (uc *SettingsUseCase) UpdateAdvancedConfig(ctx context.Context, advCfg config.AdvancedConfig) error {
	if err := advCfg.Validate(); err != nil {
		return fmt.Errorf("validate advanced config: %w", err)
	}

	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	cfg.Advanced = advCfg
	return uc.settingsRepo.SaveConfig(ctx, cfg)
}

// IsToolEnabled checks if a tool is enabled.
func (uc *SettingsUseCase) IsToolEnabled(ctx context.Context, toolType config.ToolType) (bool, error) {
	return uc.settingsRepo.IsToolEnabled(ctx, toolType)
}

// GetToolPath retrieves the path for a specific tool.
func (uc *SettingsUseCase) GetToolPath(ctx context.Context, toolType config.ToolType) (string, error) {
	return uc.settingsRepo.GetToolPath(ctx, toolType)
}

// VerifyTool verifies that a tool at the configured path is available.
func (uc *SettingsUseCase) VerifyTool(ctx context.Context, toolType config.ToolType) error {
	path, err := uc.GetToolPath(ctx, toolType)
	if err != nil {
		return err
	}

	if path == "" {
		return fmt.Errorf("%w: no path configured for %s", config.ErrToolNotFound, toolType)
	}

	detector := tool.NewDetector()
	return detector.CheckAvailability(path)
}

// GetEnabledTools returns a list of enabled tool types.
func (uc *SettingsUseCase) GetEnabledTools(ctx context.Context) ([]config.ToolType, error) {
	cfg, err := uc.settingsRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	var enabled []config.ToolType
	for toolType, toolCfg := range cfg.Tools {
		if toolCfg.Enabled {
			enabled = append(enabled, toolType)
		}
	}

	return enabled, nil
}
