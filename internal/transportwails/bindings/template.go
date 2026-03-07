// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"context"
	"log/slog"
	"sort"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// TemplateBinding provides Wails bindings for template management.
type TemplateBinding struct {
	uc *usecase.TemplateUseCase
}

// NewTemplateBinding creates a new TemplateBinding.
func NewTemplateBinding(uc *usecase.TemplateUseCase) *TemplateBinding {
	return &TemplateBinding{uc: uc}
}

// TemplateDTO represents a template for JSON serialization to frontend.
type TemplateDTO struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Tool          string                 `json:"tool"`
	DatabaseTypes []string               `json:"database_types"`
	Version       string                 `json:"version"`
	Parameters    map[string]ParamDTO    `json:"parameters"`
	CustomData    map[string]interface{} `json:"custom_data,omitempty"`
}

// ParamDTO represents a template parameter for frontend.
type ParamDTO struct {
	Type    string        `json:"type"`
	Label   string        `json:"label"`
	Default interface{}   `json:"default"`
	Min     *int          `json:"min,omitempty"`
	Max     *int          `json:"max,omitempty"`
	Options []string      `json:"options,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// TemplateListResult represents the result of ListTemplates.
type TemplateListResult struct {
	Templates []TemplateDTO `json:"templates"`
	Error     string        `json:"error,omitempty"`
}

// TemplateParamsResult represents the result of GetTemplateParams.
type TemplateParamsResult struct {
	Params []ParamDefinition `json:"params"`
	Error  string            `json:"error,omitempty"`
}

// ParamDefinition represents a parameter definition for form rendering.
type ParamDefinition struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Label   string      `json:"label"`
	Default interface{} `json:"default"`
	Min     *int        `json:"min,omitempty"`
	Max     *int        `json:"max,omitempty"`
	Options []string    `json:"options,omitempty"`
}

// ListTemplates returns all templates (Wails binding).
func (b *TemplateBinding) ListTemplates() TemplateListResult {
	ctx := context.Background()
	templates, err := b.uc.ListTemplates(ctx)
	if err != nil {
		slog.Error("ListTemplates failed", "error", err)
		return TemplateListResult{
			Error: err.Error(),
		}
	}

	dtos := make([]TemplateDTO, 0, len(templates))
	for _, t := range templates {
		dtos = append(dtos, b.toDTO(t))
	}

	// Sort by name
	sort.Slice(dtos, func(i, j int) bool {
		return dtos[i].Name < dtos[j].Name
	})

	return TemplateListResult{
		Templates: dtos,
	}
}

// ListTemplatesByType returns templates filtered by database type (Wails binding).
func (b *TemplateBinding) ListTemplatesByType(dbType string) TemplateListResult {
	ctx := context.Background()
	allTemplates, err := b.uc.ListTemplates(ctx)
	if err != nil {
		slog.Error("ListTemplatesByType failed", "dbType", dbType, "error", err)
		return TemplateListResult{
			Error: err.Error(),
		}
	}

	// Filter by database type
	var filtered []*template.Template
	for _, t := range allTemplates {
		if t.SupportsDatabase(dbType) {
			filtered = append(filtered, t)
		}
	}

	dtos := make([]TemplateDTO, 0, len(filtered))
	for _, t := range filtered {
		dtos = append(dtos, b.toDTO(t))
	}

	// Sort by name
	sort.Slice(dtos, func(i, j int) bool {
		return dtos[i].Name < dtos[j].Name
	})

	return TemplateListResult{
		Templates: dtos,
	}
}

// GetTemplate returns a single template by ID (Wails binding).
func (b *TemplateBinding) GetTemplate(id string) *TemplateDTO {
	ctx := context.Background()
	tmpl, err := b.uc.GetTemplate(ctx, id)
	if err != nil {
		slog.Error("GetTemplate failed", "id", id, "error", err)
		return nil
	}
	if tmpl == nil {
		return nil
	}

	dto := b.toDTO(tmpl)
	return &dto
}

// GetTemplateParams returns parameter definitions for a template (Wails binding).
func (b *TemplateBinding) GetTemplateParams(id string) TemplateParamsResult {
	ctx := context.Background()
	tmpl, err := b.uc.GetTemplate(ctx, id)
	if err != nil {
		slog.Error("GetTemplateParams failed", "id", id, "error", err)
		return TemplateParamsResult{
			Error: err.Error(),
		}
	}

	params := make([]ParamDefinition, 0, len(tmpl.Parameters))
	for name, p := range tmpl.Parameters {
		def := ParamDefinition{
			Name:    name,
			Type:    string(p.Type),
			Label:   p.Label,
			Default: p.Default,
			Min:     p.Min,
			Max:     p.Max,
			Options: p.Options,
		}
		params = append(params, def)
	}

	// Sort by name for consistent ordering
	sort.Slice(params, func(i, j int) bool {
		return params[i].Name < params[j].Name
	})

	return TemplateParamsResult{
		Params: params,
	}
}

// ValidateTemplateForDB checks if a template supports a specific database type (Wails binding).
func (b *TemplateBinding) ValidateTemplateForDB(templateID, dbType string) bool {
	ctx := context.Background()
	err := b.uc.ValidateTemplateForDatabase(ctx, templateID, dbType)
	return err == nil
}

// GetTemplateMetadata returns metadata about a template (Wails binding).
func (b *TemplateBinding) GetTemplateMetadata(id string) *usecase.TemplateMetadata {
	ctx := context.Background()
	meta, err := b.uc.GetTemplateMetadata(ctx, id)
	if err != nil {
		slog.Error("GetTemplateMetadata failed", "id", id, "error", err)
		return nil
	}
	return meta
}

// toDTO converts a Template to TemplateDTO.
func (b *TemplateBinding) toDTO(t *template.Template) TemplateDTO {
	dto := TemplateDTO{
		ID:            t.ID,
		Name:          t.Name,
		Description:   t.Description,
		Tool:          t.Tool,
		DatabaseTypes: t.DatabaseTypes,
		Version:       t.Version,
		Parameters:    make(map[string]ParamDTO),
		CustomData:    t.CustomData,
	}

	// Convert parameters
	for name, p := range t.Parameters {
		dto.Parameters[name] = ParamDTO{
			Type:    string(p.Type),
			Label:   p.Label,
			Default: p.Default,
			Min:     p.Min,
			Max:     p.Max,
			Options: p.Options,
			Extra:   p.Extra,
		}
	}

	return dto
}
