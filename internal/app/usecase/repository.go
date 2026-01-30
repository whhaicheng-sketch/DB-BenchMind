// Package usecase defines repository interfaces for database operations.
// These interfaces are defined by the use case layer and implemented by the infrastructure layer.
// Implements: Clean Architecture - Dependency Inversion Principle
package usecase

import (
	"context"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/config"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// =============================================================================
// Connection Repository Interface
// Implements: REQ-CONN-001, REQ-CONN-008
// =============================================================================

// ConnectionRepository defines the interface for connection persistence operations.
// This interface is defined by the use case layer and implemented by the infrastructure layer.
type ConnectionRepository interface {
	// Save saves a connection to the database.
	// If the connection already exists (by ID), it will be updated.
	// Returns an error if the operation fails.
	Save(ctx context.Context, conn connection.Connection) error

	// FindByID finds a connection by its ID.
	// Returns the connection if found, or an error if not found or operation fails.
	FindByID(ctx context.Context, id string) (connection.Connection, error)

	// FindAll finds all connections in the database.
	// Returns a slice of connections, ordered by name.
	// Returns an empty slice if no connections exist.
	// Returns an error if the operation fails.
	FindAll(ctx context.Context) ([]connection.Connection, error)

	// Delete deletes a connection by its ID.
	// Returns an error if the connection is not found or operation fails.
	Delete(ctx context.Context, id string) error

	// ExistsByName checks if a connection with the given name exists.
	// Returns true if a connection with the name exists (excluding the given ID if provided).
	// Returns an error if the operation fails.
	ExistsByName(ctx context.Context, name string, excludeID string) (bool, error)
}

// =============================================================================
// Template Repository Interface
// Implements: REQ-TMPL-001, REQ-TMPL-002
// =============================================================================

// TemplateRepository defines the interface for template persistence operations.
// This interface is defined by the use case layer and implemented by the infrastructure layer.
type TemplateRepository interface {
	// Save saves a template to the database.
	// If the template already exists (by ID), it will be updated.
	Save(ctx context.Context, tmpl *template.Template) error

	// FindByID finds a template by its ID.
	FindByID(ctx context.Context, id string) (*template.Template, error)

	// FindAll finds all templates in the database.
	// Returns a slice of templates, ordered by tool and name.
	FindAll(ctx context.Context) ([]*template.Template, error)

	// FindBuiltin finds all builtin templates.
	FindBuiltin(ctx context.Context) ([]*template.Template, error)

	// FindCustom finds all user-defined (non-builtin) templates.
	FindCustom(ctx context.Context) ([]*template.Template, error)

	// Delete deletes a template by its ID.
	// Builtin templates cannot be deleted.
	Delete(ctx context.Context, id string) error

	// LoadBuiltinTemplates loads builtin templates into the database.
	LoadBuiltinTemplates(ctx context.Context, templates []*template.Template) error
}

// =============================================================================
// Run Repository Interface
// Implements: REQ-STORAGE-001, REQ-STORAGE-004, REQ-STORAGE-005
// =============================================================================

// RunRepository defines the interface for run persistence operations.
type RunRepository interface {
	// Save saves a run to the database.
	Save(ctx context.Context, run *execution.Run) error

	// FindByID finds a run by its ID.
	FindByID(ctx context.Context, id string) (*execution.Run, error)

	// FindAll finds runs with optional filtering and pagination.
	FindAll(ctx context.Context, opts FindOptions) ([]*execution.Run, error)

	// UpdateState updates the state of a run.
	UpdateState(ctx context.Context, id string, state execution.RunState) error

	// SaveMetricSample saves a metric sample for a run.
	SaveMetricSample(ctx context.Context, runID string, sample execution.MetricSample) error

	// GetMetricSamples retrieves all metric samples for a run.
	GetMetricSamples(ctx context.Context, runID string) ([]execution.MetricSample, error)

	// SaveLogEntry saves a log entry for a run.
	SaveLogEntry(ctx context.Context, runID string, entry LogEntry) error

	// Delete deletes a run by its ID.
	Delete(ctx context.Context, id string) error
}

// FindOptions defines options for finding runs.
type FindOptions struct {
	Limit       int                     // Maximum number of results
	Offset      int                     // Number of results to skip
	StateFilter *execution.RunState     // Filter by state
	TaskID      string                  // Filter by task ID
	SortBy      string                  // Sort field: created_at, started_at, duration
	SortOrder   string                  // Sort order: ASC, DESC
}

// LogEntry represents a log entry for a run.
// Implements: REQ-EXEC-005
type LogEntry struct {
	Timestamp string // ISO 8601 format
	Stream    string // "stdout" or "stderr"
	Content   string // Log content
}

// =============================================================================
// Settings Repository Interface
// Implements: Phase 7 - Settings Management
// =============================================================================

// SettingsRepository defines the interface for configuration persistence operations.
// This interface is defined by the use case layer and implemented by the infrastructure layer.
type SettingsRepository interface {
	// GetConfig retrieves the complete configuration.
	GetConfig(ctx context.Context) (*config.Config, error)

	// SaveConfig saves the complete configuration.
	SaveConfig(ctx context.Context, cfg *config.Config) error

	// GetToolPath returns the path for a specific tool.
	GetToolPath(ctx context.Context, toolType config.ToolType) (string, error)

	// SetToolPath sets the path for a specific tool.
	SetToolPath(ctx context.Context, toolType config.ToolType, path string) error

	// IsToolEnabled checks if a tool is enabled.
	IsToolEnabled(ctx context.Context, toolType config.ToolType) (bool, error)

	// SetToolEnabled enables or disables a tool.
	SetToolEnabled(ctx context.Context, toolType config.ToolType, enabled bool) error

	// SetToolVersion sets the detected version for a tool.
	SetToolVersion(ctx context.Context, toolType config.ToolType, version string) error

	// GetToolConfig returns the configuration for a specific tool.
	GetToolConfig(ctx context.Context, toolType config.ToolType) (*config.ToolConfig, error)

	// ResetToDefaults resets configuration to defaults.
	ResetToDefaults(ctx context.Context) error
}

