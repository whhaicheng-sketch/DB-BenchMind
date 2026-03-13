package bindings

import (
	"context"
	"log/slog"
	"sort"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

type TemplateBinding struct {
	uc *usecase.TemplateUseCase
}

func NewTemplateBinding(uc *usecase.TemplateUseCase) *TemplateBinding {
	return &TemplateBinding{uc: uc}
}

type TemplateDTO struct {
	ID             string                       `json:"id"`
	Name           string                       `json:"name"`
	Description    string                       `json:"description"`
	Tool           string                       `json:"tool"`
	DBFamily       string                       `json:"dbFamily"`
	WorkloadFamily string                       `json:"workloadFamily"`
	Scope          string                       `json:"scope"`
	Tags           []string                     `json:"tags"`
	Status         string                       `json:"status"`
	Version        string                       `json:"version"`
	CreatedAt      string                       `json:"createdAt"`
	UpdatedAt      string                       `json:"updatedAt"`
	Compatibility  domaintemplate.Compatibility `json:"compatibility"`
	Phases         domaintemplate.PhaseSet      `json:"phases"`
	Runtime        domaintemplate.Runtime       `json:"runtime"`
	ToolConfig     domaintemplate.ToolConfig    `json:"toolConfig"`
	DatabaseTypes  []string                     `json:"database_types"`
	Parameters     map[string]ParamDTO          `json:"parameters,omitempty"`
	CustomData     map[string]interface{}       `json:"custom_data,omitempty"`
}

type ParamDTO struct {
	Type    string                 `json:"type"`
	Label   string                 `json:"label"`
	Default interface{}            `json:"default"`
	Min     *int                   `json:"min,omitempty"`
	Max     *int                   `json:"max,omitempty"`
	Options []string               `json:"options,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

type TemplateListResult struct {
	Templates []TemplateDTO `json:"templates"`
	Error     string        `json:"error,omitempty"`
}

type TemplateResult struct {
	Template *TemplateDTO `json:"template,omitempty"`
	Error    string       `json:"error,omitempty"`
}

type TemplateDeleteResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type TemplateParamsResult struct {
	Params []ParamDefinition `json:"params"`
	Error  string            `json:"error,omitempty"`
}

type ParamDefinition struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Label   string      `json:"label"`
	Default interface{} `json:"default"`
	Min     *int        `json:"min,omitempty"`
	Max     *int        `json:"max,omitempty"`
	Options []string    `json:"options,omitempty"`
}

func (b *TemplateBinding) ListTemplates() TemplateListResult {
	ctx := context.Background()
	templates, err := b.uc.ListTemplates(ctx)
	if err != nil {
		slog.Error("ListTemplates failed", "error", err)
		return TemplateListResult{Error: err.Error()}
	}

	dtos := make([]TemplateDTO, 0, len(templates))
	for _, tmpl := range templates {
		dtos = append(dtos, b.toDTO(tmpl))
	}
	sort.Slice(dtos, func(i, j int) bool { return dtos[i].Name < dtos[j].Name })

	return TemplateListResult{Templates: dtos}
}

func (b *TemplateBinding) ListTemplatesByType(dbType string) TemplateListResult {
	ctx := context.Background()
	templates, err := b.uc.ListTemplates(ctx)
	if err != nil {
		slog.Error("ListTemplatesByType failed", "dbType", dbType, "error", err)
		return TemplateListResult{Error: err.Error()}
	}

	filtered := make([]TemplateDTO, 0)
	for _, tmpl := range templates {
		if tmpl.SupportsDatabase(dbType) {
			filtered = append(filtered, b.toDTO(tmpl))
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Name < filtered[j].Name })
	return TemplateListResult{Templates: filtered}
}

func (b *TemplateBinding) GetTemplate(id string) *TemplateDTO {
	ctx := context.Background()
	tmpl, err := b.uc.GetTemplate(ctx, id)
	if err != nil {
		slog.Error("GetTemplate failed", "id", id, "error", err)
		return nil
	}
	dto := b.toDTO(tmpl)
	return &dto
}

func (b *TemplateBinding) CreateTemplate(req TemplateDTO) TemplateResult {
	ctx := context.Background()
	tmpl := b.fromDTO(req)
	if err := b.uc.CreateTemplate(ctx, tmpl); err != nil {
		slog.Error("CreateTemplate failed", "error", err)
		return TemplateResult{Error: err.Error()}
	}
	created, err := b.uc.GetTemplate(ctx, tmpl.ID)
	if err != nil {
		return TemplateResult{Error: err.Error()}
	}
	dto := b.toDTO(created)
	return TemplateResult{Template: &dto}
}

func (b *TemplateBinding) UpdateTemplate(req TemplateDTO) TemplateResult {
	ctx := context.Background()
	tmpl := b.fromDTO(req)
	if err := b.uc.UpdateTemplate(ctx, tmpl); err != nil {
		slog.Error("UpdateTemplate failed", "id", req.ID, "error", err)
		return TemplateResult{Error: err.Error()}
	}
	updated, err := b.uc.GetTemplate(ctx, req.ID)
	if err != nil {
		return TemplateResult{Error: err.Error()}
	}
	dto := b.toDTO(updated)
	return TemplateResult{Template: &dto}
}

func (b *TemplateBinding) DeleteTemplate(id string) TemplateDeleteResult {
	ctx := context.Background()
	if err := b.uc.DeleteTemplate(ctx, id); err != nil {
		slog.Error("DeleteTemplate failed", "id", id, "error", err)
		return TemplateDeleteResult{Error: err.Error()}
	}
	return TemplateDeleteResult{Success: true}
}

func (b *TemplateBinding) DuplicateTemplate(id string) TemplateResult {
	ctx := context.Background()
	tmpl, err := b.uc.DuplicateTemplate(ctx, id)
	if err != nil {
		slog.Error("DuplicateTemplate failed", "id", id, "error", err)
		return TemplateResult{Error: err.Error()}
	}
	dto := b.toDTO(tmpl)
	return TemplateResult{Template: &dto}
}

func (b *TemplateBinding) GetTemplateParams(id string) TemplateParamsResult {
	ctx := context.Background()
	tmpl, err := b.uc.GetTemplate(ctx, id)
	if err != nil {
		slog.Error("GetTemplateParams failed", "id", id, "error", err)
		return TemplateParamsResult{Error: err.Error()}
	}

	params := make([]ParamDefinition, 0, len(tmpl.Parameters))
	for name, p := range tmpl.Parameters {
		params = append(params, ParamDefinition{
			Name:    name,
			Type:    string(p.Type),
			Label:   p.Label,
			Default: p.Default,
			Min:     p.Min,
			Max:     p.Max,
			Options: p.Options,
		})
	}
	sort.Slice(params, func(i, j int) bool { return params[i].Name < params[j].Name })
	return TemplateParamsResult{Params: params}
}

func (b *TemplateBinding) ValidateTemplateForDB(templateID, dbType string) bool {
	ctx := context.Background()
	err := b.uc.ValidateTemplateForDatabase(ctx, templateID, dbType)
	return err == nil
}

func (b *TemplateBinding) GetTemplateMetadata(id string) *usecase.TemplateMetadata {
	ctx := context.Background()
	meta, err := b.uc.GetTemplateMetadata(ctx, id)
	if err != nil {
		slog.Error("GetTemplateMetadata failed", "id", id, "error", err)
		return nil
	}
	return meta
}

func (b *TemplateBinding) toDTO(t *domaintemplate.Template) TemplateDTO {
	params := make(map[string]ParamDTO, len(t.Parameters))
	for name, p := range t.Parameters {
		params[name] = ParamDTO{
			Type:    string(p.Type),
			Label:   p.Label,
			Default: p.Default,
			Min:     p.Min,
			Max:     p.Max,
			Options: p.Options,
			Extra:   p.Extra,
		}
	}
	return TemplateDTO{
		ID:             t.ID,
		Name:           t.Name,
		Description:    t.Description,
		Tool:           t.Tool,
		DBFamily:       t.DBFamily,
		WorkloadFamily: t.WorkloadFamily,
		Scope:          t.Scope,
		Tags:           append([]string{}, t.Tags...),
		Status:         t.Status,
		Version:        t.Version,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		Compatibility:  t.Compatibility,
		Phases:         t.Phases,
		Runtime:        t.Runtime,
		ToolConfig:     t.ToolConfig,
		DatabaseTypes:  append([]string{}, t.DatabaseTypes...),
		Parameters:     params,
		CustomData:     t.CustomData,
	}
}

func (b *TemplateBinding) fromDTO(dto TemplateDTO) *domaintemplate.Template {
	params := make(map[string]domaintemplate.Parameter, len(dto.Parameters))
	for name, p := range dto.Parameters {
		params[name] = domaintemplate.Parameter{
			Type:    domaintemplate.ParameterType(p.Type),
			Label:   p.Label,
			Default: p.Default,
			Min:     p.Min,
			Max:     p.Max,
			Options: p.Options,
			Extra:   p.Extra,
		}
	}

	tmpl := &domaintemplate.Template{
		ID:             dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Tool:           dto.Tool,
		DBFamily:       dto.DBFamily,
		WorkloadFamily: dto.WorkloadFamily,
		Scope:          dto.Scope,
		Tags:           dto.Tags,
		Status:         dto.Status,
		Version:        dto.Version,
		CreatedAt:      dto.CreatedAt,
		UpdatedAt:      dto.UpdatedAt,
		Compatibility:  dto.Compatibility,
		Phases:         dto.Phases,
		Runtime:        dto.Runtime,
		ToolConfig:     dto.ToolConfig,
		DatabaseTypes:  dto.DatabaseTypes,
		Parameters:     params,
		CustomData:     dto.CustomData,
	}
	tmpl.Normalize()
	return tmpl
}
