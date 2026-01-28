// Package config provides configuration domain models.
// Implements: Phase 7 - Settings and Optimization
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	// ErrInvalidConfiguration is returned when configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrToolNotFound is returned when a tool is not found.
	ErrToolNotFound = errors.New("tool not found")

	// ErrInvalidToolPath is returned when a tool path is invalid.
	ErrInvalidToolPath = errors.New("invalid tool path")
)

// ToolType represents a benchmark tool type.
type ToolType string

const (
	ToolTypeSysbench  ToolType = "sysbench"
	ToolTypeSwingbench ToolType = "swingbench"
	ToolTypeHammerDB   ToolType = "hammerdb"
)

// String returns the string representation of the tool type.
func (t ToolType) String() string {
	return string(t)
}

// Validate checks if the tool type is valid.
func (t ToolType) Validate() error {
	switch t {
	case ToolTypeSysbench, ToolTypeSwingbench, ToolTypeHammerDB:
		return nil
	default:
		return fmt.Errorf("%w: unknown tool type: %s", ErrInvalidConfiguration, t)
	}
}

// ToolConfig represents configuration for a benchmark tool.
type ToolConfig struct {
	// Type is the tool type.
	Type ToolType `json:"type"`

	// Path is the custom path to the tool executable.
	// If empty, the tool will be searched in PATH.
	Path string `json:"path,omitempty"`

	// Version is the detected tool version.
	Version string `json:"version,omitempty"`

	// Enabled indicates if the tool is enabled for use.
	Enabled bool `json:"enabled"`
}

// Validate validates the tool configuration.
func (c *ToolConfig) Validate() error {
	if err := c.Type.Validate(); err != nil {
		return err
	}

	if c.Path != "" {
		if !filepath.IsAbs(c.Path) {
			return fmt.Errorf("%w: tool path must be absolute", ErrInvalidToolPath)
		}

		// Check if path exists and is executable
		info, err := os.Stat(c.Path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%w: tool path does not exist: %s", ErrToolNotFound, c.Path)
			}
			return fmt.Errorf("check tool path: %w", err)
		}

		if info.IsDir() {
			return fmt.Errorf("%w: tool path is a directory, not an executable", ErrInvalidToolPath)
		}

		// Check if file is executable
		if info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("%w: tool path is not executable", ErrInvalidToolPath)
		}
	}

	return nil
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	// Path is the path to the SQLite database file.
	Path string `json:"path"`

	// MaxOpenConns is the maximum number of open connections.
	MaxOpenConns int `json:"max_open_conns"`

	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int `json:"max_idle_conns"`

	// ConnMaxLifetime is the maximum connection lifetime in seconds.
	ConnMaxLifetime int `json:"conn_max_lifetime"`
}

// Validate validates the database configuration.
func (c *DatabaseConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("%w: database path is required", ErrInvalidConfiguration)
	}

	// Ensure parent directory exists or can be created
	dir := filepath.Dir(c.Path)
	if dir != "" && dir != "." {
		if _, err := os.Stat(dir); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("check database directory: %w", err)
			}
		}
	}

	if c.MaxOpenConns < 1 {
		return fmt.Errorf("%w: max_open_conns must be at least 1", ErrInvalidConfiguration)
	}

	if c.MaxIdleConns < 0 {
		return fmt.Errorf("%w: max_idle_conns cannot be negative", ErrInvalidConfiguration)
	}

	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("%w: max_idle_conns cannot exceed max_open_conns", ErrInvalidConfiguration)
	}

	return nil
}

// ReportConfig represents report generation configuration.
type ReportConfig struct {
	// DefaultFormat is the default report format.
	DefaultFormat string `json:"default_format"`

	// IncludeCharts enables chart generation by default.
	IncludeCharts bool `json:"include_charts"`

	// IncludeLogs enables log inclusion in reports by default.
	IncludeLogs bool `json:"include_logs"`

	// ChartWidth is the default width for text-based charts.
	ChartWidth int `json:"chart_width"`

	// ChartHeight is the default height for text-based charts.
	ChartHeight int `json:"chart_height"`

	// OutputDir is the default directory for report output.
	OutputDir string `json:"output_dir"`
}

// Validate validates the report configuration.
func (c *ReportConfig) Validate() error {
	validFormats := map[string]bool{
		"markdown": true,
		"json":     true,
		"html":     true,
		"pdf":      true,
	}

	if !validFormats[c.DefaultFormat] {
		return fmt.Errorf("%w: invalid default format: %s", ErrInvalidConfiguration, c.DefaultFormat)
	}

	if c.ChartWidth < 20 || c.ChartWidth > 200 {
		return fmt.Errorf("%w: chart_width must be between 20 and 200", ErrInvalidConfiguration)
	}

	if c.ChartHeight < 5 || c.ChartHeight > 50 {
		return fmt.Errorf("%w: chart_height must be between 5 and 50", ErrInvalidConfiguration)
	}

	if c.OutputDir != "" {
		if !filepath.IsAbs(c.OutputDir) {
			return fmt.Errorf("%w: output_dir must be an absolute path", ErrInvalidConfiguration)
		}
	}

	return nil
}

// UIConfig represents UI configuration.
type UIConfig struct {
	// Theme is the UI theme (light, dark, auto).
	Theme string `json:"theme"`

	// Language is the UI language.
	Language string `json:"language"`

	// AutoSave indicates if changes should be auto-saved.
	AutoSave bool `json:"auto_save"`

	// RefreshInterval is the refresh interval for live updates in seconds.
	RefreshInterval int `json:"refresh_interval"`
}

// Validate validates the UI configuration.
func (c *UIConfig) Validate() error {
	validThemes := map[string]bool{
		"light": true,
		"dark":  true,
		"auto":  true,
	}

	if !validThemes[c.Theme] {
		return fmt.Errorf("%w: invalid theme: %s", ErrInvalidConfiguration, c.Theme)
	}

	if c.RefreshInterval < 1 || c.RefreshInterval > 60 {
		return fmt.Errorf("%w: refresh_interval must be between 1 and 60 seconds", ErrInvalidConfiguration)
	}

	return nil
}

// AdvancedConfig represents advanced configuration.
type AdvancedConfig struct {
	// LogLevel is the logging level (debug, info, warn, error).
	LogLevel string `json:"log_level"`

	// MaxLogFiles is the maximum number of log files to keep.
	MaxLogFiles int `json:"max_log_files"`

	// EnableTelemetry enables anonymous usage telemetry.
	EnableTelemetry bool `json:"enable_telemetry"`

	// CheckUpdates enables automatic update checks.
	CheckUpdates bool `json:"check_updates"`

	// WorkDir is the working directory for benchmark execution.
	WorkDir string `json:"work_dir"`

	// Timeout is the default timeout for benchmark execution in minutes.
	Timeout int `json:"timeout"`
}

// Validate validates the advanced configuration.
func (c *AdvancedConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[c.LogLevel] {
		return fmt.Errorf("%w: invalid log level: %s", ErrInvalidConfiguration, c.LogLevel)
	}

	if c.MaxLogFiles < 0 || c.MaxLogFiles > 100 {
		return fmt.Errorf("%w: max_log_files must be between 0 and 100", ErrInvalidConfiguration)
	}

	if c.WorkDir != "" {
		if !filepath.IsAbs(c.WorkDir) {
			return fmt.Errorf("%w: work_dir must be an absolute path", ErrInvalidConfiguration)
		}
	}

	if c.Timeout < 1 || c.Timeout > 1440 {
		return fmt.Errorf("%w: timeout must be between 1 and 1440 minutes", ErrInvalidConfiguration)
	}

	return nil
}

// Config represents the complete application configuration.
type Config struct {
	// Version is the configuration version.
	Version int `json:"version"`

	// Database is the database configuration.
	Database DatabaseConfig `json:"database"`

	// Tools maps tool types to their configurations.
	Tools map[ToolType]ToolConfig `json:"tools"`

	// Reports is the report configuration.
	Reports ReportConfig `json:"reports"`

	// UI is the UI configuration.
	UI UIConfig `json:"ui"`

	// Advanced is the advanced configuration.
	Advanced AdvancedConfig `json:"advanced"`
}

// Validate validates the complete configuration.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("%w: unsupported configuration version: %d", ErrInvalidConfiguration, c.Version)
	}

	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database: %w", err)
	}

	for toolType, toolConfig := range c.Tools {
		if toolConfig.Type != toolType {
			return fmt.Errorf("%w: tool config type mismatch: expected %s, got %s",
				ErrInvalidConfiguration, toolType, toolConfig.Type)
		}
		if err := toolConfig.Validate(); err != nil {
			return fmt.Errorf("tool %s: %w", toolType, err)
		}
	}

	if err := c.Reports.Validate(); err != nil {
		return fmt.Errorf("reports: %w", err)
	}

	if err := c.UI.Validate(); err != nil {
		return fmt.Errorf("ui: %w", err)
	}

	if err := c.Advanced.Validate(); err != nil {
		return fmt.Errorf("advanced: %w", err)
	}

	return nil
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	// Get default database path
	userHomeDir, _ := os.UserHomeDir()
	defaultDBPath := filepath.Join(userHomeDir, ".db-benchmind", "benchmarks.db")
	defaultWorkDir := filepath.Join(os.TempDir(), "db-benchmind")
	defaultOutputDir := filepath.Join(userHomeDir, ".db-benchmind", "reports")

	return &Config{
		Version: 1,
		Database: DatabaseConfig{
			Path:            defaultDBPath,
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300, // 5 minutes
		},
		Tools: map[ToolType]ToolConfig{
			ToolTypeSysbench: {
				Type:    ToolTypeSysbench,
				Path:    "",
				Enabled: true,
			},
			ToolTypeSwingbench: {
				Type:    ToolTypeSwingbench,
				Path:    "",
				Enabled: false, // Disabled by default (requires Oracle)
			},
			ToolTypeHammerDB: {
				Type:    ToolTypeHammerDB,
				Path:    "",
				Enabled: false, // Disabled by default
			},
		},
		Reports: ReportConfig{
			DefaultFormat: "markdown",
			IncludeCharts: true,
			IncludeLogs:   false,
			ChartWidth:    60,
			ChartHeight:   10,
			OutputDir:     defaultOutputDir,
		},
		UI: UIConfig{
			Theme:           "auto",
			Language:        "en",
			AutoSave:        true,
			RefreshInterval: 5,
		},
		Advanced: AdvancedConfig{
			LogLevel:       "info",
			MaxLogFiles:    10,
			EnableTelemetry: false,
			CheckUpdates:   true,
			WorkDir:        defaultWorkDir,
			Timeout:        60, // 1 hour
		},
	}
}

// GetToolConfig returns the configuration for a specific tool.
func (c *Config) GetToolConfig(toolType ToolType) (*ToolConfig, error) {
	config, ok := c.Tools[toolType]
	if !ok {
		return nil, fmt.Errorf("%w: tool %s", ErrToolNotFound, toolType)
	}
	return &config, nil
}

// SetToolConfig sets the configuration for a specific tool.
func (c *Config) SetToolConfig(config ToolConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	if c.Tools == nil {
		c.Tools = make(map[ToolType]ToolConfig)
	}
	c.Tools[config.Type] = config
	return nil
}

// GetToolPath returns the executable path for a tool.
// Returns empty string if tool is disabled or not configured.
func (c *Config) GetToolPath(toolType ToolType) string {
	config, ok := c.Tools[toolType]
	if !ok || !config.Enabled {
		return ""
	}
	return config.Path
}

// IsToolEnabled checks if a tool is enabled.
func (c *Config) IsToolEnabled(toolType ToolType) bool {
	config, ok := c.Tools[toolType]
	return ok && config.Enabled
}
