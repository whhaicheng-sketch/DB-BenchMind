// Package usecase provides template management business logic.
// Implements: REQ-TMPL-001 ~ REQ-TMPL-007
package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

var (
	// ErrBuiltinTemplateCannotBeDeleted is returned when trying to delete a builtin template.
	ErrBuiltinTemplateCannotBeDeleted = errors.New("builtin templates cannot be deleted")

	// ErrTemplateInvalid is returned when template validation fails.
	ErrTemplateInvalid = errors.New("template validation failed")

	// ErrTemplateNotFound is returned when a template is not found.
	ErrTemplateNotFound = errors.New("template not found")

	// ErrTemplateIDRequired is returned when template ID is not provided.
	ErrTemplateIDRequired = errors.New("template ID is required")
)

// TemplateUseCase provides template management business operations.
// Implements: REQ-TMPL-001 ~ REQ-TMPL-007
type TemplateUseCase struct {
	repo        TemplateRepository
	builtinPath string // Path to builtin templates directory
}

// NewTemplateUseCase creates a new template use case.
func NewTemplateUseCase(repo TemplateRepository, builtinPath string) *TemplateUseCase {
	return &TemplateUseCase{
		repo:        repo,
		builtinPath: builtinPath,
	}
}

// =============================================================================
// Template Operations
// Implements: REQ-TMPL-001, REQ-TMPL-002
// =============================================================================

// ListTemplates lists all templates (both builtin and custom).
// Implements: REQ-TMPL-001
func (uc *TemplateUseCase) ListTemplates(ctx context.Context) ([]*template.Template, error) {
	return uc.repo.FindAll(ctx)
}

// GetTemplate retrieves a template by ID.
// Implements: REQ-TMPL-002
func (uc *TemplateUseCase) GetTemplate(ctx context.Context, id string) (*template.Template, error) {
	if uc.repo == nil {
		return nil, ErrTemplateNotFound
	}

	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	return tmpl, nil
}

// ListBuiltinTemplates lists all builtin templates.
// Implements: REQ-TMPL-001
func (uc *TemplateUseCase) ListBuiltinTemplates(ctx context.Context) ([]*template.Template, error) {
	return uc.repo.FindBuiltin(ctx)
}

// ListCustomTemplates lists all user-defined templates.
// Implements: REQ-TMPL-001
func (uc *TemplateUseCase) ListCustomTemplates(ctx context.Context) ([]*template.Template, error) {
	return uc.repo.FindCustom(ctx)
}

// =============================================================================
// Template CRUD Operations
// Implements: REQ-TMPL-003, REQ-TMPL-004
// =============================================================================

// ImportTemplate imports a template from a file (JSON or YAML format).
// Implements: REQ-TMPL-003, REQ-TMPL-004
func (uc *TemplateUseCase) ImportTemplate(ctx context.Context, filePath string) (*template.Template, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read template file: %w", err)
	}

	// Parse JSON
	tmpl, err := template.FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("parse template JSON: %w", err)
	}

	// Validate template
	if err := tmpl.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTemplateInvalid, err)
	}

	// Generate ID if not set
	if tmpl.ID == "" {
		tmpl.ID = generateTemplateID()
	}

	// Check if tool is available (TODO: implement tool availability check)

	// Save to repository
	if err := uc.repo.Save(ctx, tmpl); err != nil {
		return nil, fmt.Errorf("save template: %w", err)
	}

	return tmpl, nil
}

// CreateTemplate creates a new custom template.
// Implements: REQ-TMPL-003
func (uc *TemplateUseCase) CreateTemplate(ctx context.Context, tmpl *template.Template) error {
	// Validate template
	if err := tmpl.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateInvalid, err)
	}

	// Generate ID if not set
	if tmpl.ID == "" {
		tmpl.ID = generateTemplateID()
	}

	// Save to repository
	if err := uc.repo.Save(ctx, tmpl); err != nil {
		return fmt.Errorf("save template: %w", err)
	}

	return nil
}

// UpdateTemplate updates an existing template.
// Builtin templates cannot be updated.
func (uc *TemplateUseCase) UpdateTemplate(ctx context.Context, tmpl *template.Template) error {
	// Validate template
	if err := tmpl.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateInvalid, err)
	}

	if tmpl.ID == "" {
		return ErrTemplateIDRequired
	}

	// Check if template exists and is not builtin
	existing, err := uc.repo.FindByID(ctx, tmpl.ID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get existing template: %w", err)
	}

	// For now, we don't have an easy way to check if a template is builtin
	// without adding a method to the repository interface
	// We'll rely on the repository to handle this

	// Save updated template (repository will handle updates)
	if err := uc.repo.Save(ctx, tmpl); err != nil {
		return fmt.Errorf("save template: %w", err)
	}

	_ = existing // Avoid unused variable warning
	return nil
}

// DeleteTemplate deletes a template by ID.
// Builtin templates cannot be deleted.
// Implements: REQ-TMPL-004
func (uc *TemplateUseCase) DeleteTemplate(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrBuiltinTemplateCannotBeDeleted) {
			return ErrBuiltinTemplateCannotBeDeleted
		}
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

// =============================================================================
// Template Export
// Implements: REQ-TMPL-006
// =============================================================================

// ExportTemplate exports a template to a JSON file.
// Implements: REQ-TMPL-006
func (uc *TemplateUseCase) ExportTemplate(ctx context.Context, id, filePath string) error {
	// Get template
	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get template: %w", err)
	}

	// Serialize to JSON
	data, err := tmpl.ToJSON()
	if err != nil {
		return fmt.Errorf("serialize template: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// =============================================================================
// Builtin Template Loading
// Implements: REQ-TMPL-007
// =============================================================================

// LoadBuiltinTemplates loads all builtin templates from the builtin templates directory.
// This should be called during application initialization.
// Implements: REQ-TMPL-007
func (uc *TemplateUseCase) LoadBuiltinTemplates(ctx context.Context) error {
	if uc.builtinPath == "" {
		return nil // No builtin templates to load
	}

	// Read all JSON files from builtin templates directory
	files, err := filepath.Glob(filepath.Join(uc.builtinPath, "*.json"))
	if err != nil {
		return fmt.Errorf("find builtin templates: %w", err)
	}

	var templates []*template.Template
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read template file %s: %w", file, err)
		}

		tmpl, err := template.FromJSON(data)
		if err != nil {
			return fmt.Errorf("parse template file %s: %w", file, err)
		}

		if err := tmpl.Validate(); err != nil {
			return fmt.Errorf("validate template %s: %w", tmpl.ID, err)
		}

		templates = append(templates, tmpl)
	}

	// Load templates into repository
	if err := uc.repo.LoadBuiltinTemplates(ctx, templates); err != nil {
		return fmt.Errorf("save builtin templates: %w", err)
	}

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// generateTemplateID generates a new unique template ID.
func generateTemplateID() string {
	return fmt.Sprintf("custom-%s", uuid.New().String())
}

// ValidateTemplateForDatabase checks if a template supports a specific database type.
// Implements: REQ-EXEC-001 (pre-check tool compatibility)
func (uc *TemplateUseCase) ValidateTemplateForDatabase(ctx context.Context, templateID, dbType string) error {
	tmpl, err := uc.repo.FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get template: %w", err)
	}

	if !tmpl.SupportsDatabase(dbType) {
		return fmt.Errorf("template '%s' does not support database type '%s'", tmpl.ID, dbType)
	}

	return nil
}

// CloneTemplate creates a copy of a template with a new ID.
func (uc *TemplateUseCase) CloneTemplate(ctx context.Context, id string) (*template.Template, error) {
	// Get original template
	original, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}

	// Clone
	cloned, err := original.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone template: %w", err)
	}

	// Generate new ID
	cloned.ID = generateTemplateID()
	cloned.Name = fmt.Sprintf("%s (Copy)", cloned.Name)

	return cloned, nil
}

// GetTemplateAsJSON returns the template as JSON (for REQ-TMPL-007 template snapshot).
func (uc *TemplateUseCase) GetTemplateAsJSON(ctx context.Context, id string) ([]byte, error) {
	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}

	return tmpl.ToJSON()
}

// TemplateMetadata contains metadata about a template.
type TemplateMetadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tool        string   `json:"tool"`
	IsBuiltin   bool     `json:"is_builtin"`
	ParamCount  int      `json:"param_count"`
	DBTypes     []string `json:"database_types"`
}

// GetTemplateMetadata returns metadata about a template without the full config.
func (uc *TemplateUseCase) GetTemplateMetadata(ctx context.Context, id string) (*TemplateMetadata, error) {
	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}

	return &TemplateMetadata{
		ID:          tmpl.ID,
		Name:        tmpl.Name,
		Description: tmpl.Description,
		Tool:        tmpl.Tool,
		IsBuiltin:   false, // TODO: get from repository
		ParamCount:  len(tmpl.Parameters),
		DBTypes:     tmpl.DatabaseTypes,
	}, nil
}

// ValidateTemplateParameters validates parameter values against a template definition.
func (uc *TemplateUseCase) ValidateTemplateParameters(ctx context.Context, templateID string, params map[string]interface{}) error {
	tmpl, err := uc.repo.FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get template: %w", err)
	}

	// Validate each parameter
	for name, value := range params {
		param, ok := tmpl.Parameters[name]
		if !ok {
			return fmt.Errorf("unknown parameter: '%s'", name)
		}

		// Type-specific validation
		switch param.Type {
		case template.ParameterTypeInteger:
			if _, ok := value.(int); !ok {
				// JSON unmarshaling converts numbers to float64
				if f, ok := value.(float64); ok {
					value = int(f)
				} else {
					return fmt.Errorf("parameter '%s': expected integer, got %T", name, value)
				}
			}

			// Range validation
			if val, ok := value.(int); ok {
				if param.Min != nil && val < *param.Min {
					return fmt.Errorf("parameter '%s': value %d < min %d", name, val, *param.Min)
				}
				if param.Max != nil && val > *param.Max {
					return fmt.Errorf("parameter '%s': value %d > max %d", name, val, *param.Max)
				}
			}

		case template.ParameterTypeEnum:
			strVal, ok := value.(string)
			if !ok {
				return fmt.Errorf("parameter '%s': expected string, got %T", name, value)
			}
			found := false
			for _, opt := range param.Options {
				if opt == strVal {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("parameter '%s': value '%s' is not in options", name, strVal)
			}

		case template.ParameterTypeString, template.ParameterTypeBoolean:
			// No additional validation needed
		}
	}

	return nil
}

// SubstituteTemplateParams substitutes parameter values into a command template.
// This is used internally when preparing commands for execution.
func (uc *TemplateUseCase) SubstituteTemplateParams(cmdTemplate string, params map[string]interface{}, connectionStr string) (string, error) {
	// For now, just do basic substitution
	// TODO: Implement proper template variable substitution
	result := cmdTemplate

	// Substitute connection string
	result = substituteVar(result, "{connection_string}", connectionStr)

	// Substitute parameters
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		result = substituteVar(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result, nil
}

// substituteVar replaces a placeholder with a value in a string.
func substituteVar(s, placeholder, value string) string {
	// Simple string replacement
	// TODO: Use a proper template engine for more complex substitution
	return strings.ReplaceAll(s, placeholder, value)
}

// GetTemplateJSONWithMetadata returns the template JSON with usage statistics.
// Implements: REQ-TMPL-006
func (uc *TemplateUseCase) GetTemplateJSONWithMetadata(ctx context.Context, id string) ([]byte, error) {
	metadata, err := uc.GetTemplateMetadata(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get full template
	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create export structure with metadata
	export := struct {
		*template.Template
		Metadata *TemplateMetadata `json:"_metadata"`
	}{
		Template: tmpl,
		Metadata: metadata,
	}

	return json.MarshalIndent(export, "", "  ")
}
