// Package usecase provides template management business logic.
package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

var (
	ErrBuiltinTemplateCannotBeDeleted = errors.New("builtin templates cannot be deleted")
	ErrReadonlyTemplateCannotBeEdited = errors.New("readonly templates cannot be updated or deleted")
	ErrTemplateInvalid                = errors.New("template validation failed")
	ErrTemplateNotFound               = errors.New("template not found")
	ErrTemplateIDRequired             = errors.New("template ID is required")
)

type TemplateUseCase struct {
	repo        TemplateRepository
	builtinPath string
}

func NewTemplateUseCase(repo TemplateRepository, builtinPath string) *TemplateUseCase {
	return &TemplateUseCase{repo: repo, builtinPath: builtinPath}
}

func (uc *TemplateUseCase) ListTemplates(ctx context.Context) ([]*domaintemplate.Template, error) {
	return uc.repo.FindAll(ctx)
}

func (uc *TemplateUseCase) GetTemplate(ctx context.Context, id string) (*domaintemplate.Template, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrTemplateIDRequired
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

func (uc *TemplateUseCase) ListBuiltinTemplates(ctx context.Context) ([]*domaintemplate.Template, error) {
	return uc.repo.FindBuiltin(ctx)
}

func (uc *TemplateUseCase) ListCustomTemplates(ctx context.Context) ([]*domaintemplate.Template, error) {
	return uc.repo.FindCustom(ctx)
}

func (uc *TemplateUseCase) ImportTemplate(ctx context.Context, filePath string) (*domaintemplate.Template, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read template file: %w", err)
	}
	tmpl, err := domaintemplate.FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("parse template JSON: %w", err)
	}
	if tmpl.ID == "" {
		tmpl.ID = generateTemplateID()
	}
	if err := uc.CreateTemplate(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (uc *TemplateUseCase) CreateTemplate(ctx context.Context, tmpl *domaintemplate.Template) error {
	if tmpl == nil {
		return fmt.Errorf("%w: template payload is required", ErrTemplateInvalid)
	}
	uc.prepareTemplateForCreate(tmpl)
	if err := tmpl.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateInvalid, err)
	}
	if err := uc.repo.Save(ctx, tmpl); err != nil {
		return fmt.Errorf("save template: %w", err)
	}
	return nil
}

func (uc *TemplateUseCase) UpdateTemplate(ctx context.Context, tmpl *domaintemplate.Template) error {
	if tmpl == nil {
		return fmt.Errorf("%w: template payload is required", ErrTemplateInvalid)
	}
	if strings.TrimSpace(tmpl.ID) == "" {
		return ErrTemplateIDRequired
	}

	existing, err := uc.repo.FindByID(ctx, tmpl.ID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get existing template: %w", err)
	}
	if existing.IsReadOnly() {
		return ErrReadonlyTemplateCannotBeEdited
	}

	uc.prepareTemplateForUpdate(existing, tmpl)
	if err := tmpl.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateInvalid, err)
	}
	if err := uc.repo.Save(ctx, tmpl); err != nil {
		return fmt.Errorf("save template: %w", err)
	}
	return nil
}

func (uc *TemplateUseCase) DeleteTemplate(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrTemplateIDRequired
	}
	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get template: %w", err)
	}
	if tmpl.IsBuiltin {
		return ErrBuiltinTemplateCannotBeDeleted
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

func (uc *TemplateUseCase) DuplicateTemplate(ctx context.Context, id string) (*domaintemplate.Template, error) {
	original, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	cloned, err := original.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone template: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	cloned.ID = generateTemplateID()
	cloned.Name = fmt.Sprintf("%s Copy", original.Name)
	cloned.IsBuiltin = false
	cloned.Version = "0.1.0"
	cloned.CreatedAt = now
	cloned.Normalize()

	if err := cloned.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTemplateInvalid, err)
	}
	if err := uc.repo.Save(ctx, cloned); err != nil {
		return nil, fmt.Errorf("save duplicated template: %w", err)
	}
	return cloned, nil
}

func (uc *TemplateUseCase) ExportTemplate(ctx context.Context, id, filePath string) error {
	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get template: %w", err)
	}
	data, err := tmpl.ToJSON()
	if err != nil {
		return fmt.Errorf("serialize template: %w", err)
	}
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (uc *TemplateUseCase) LoadBuiltinTemplates(ctx context.Context) error {
	seeds := domaintemplate.DefaultSeedTemplates()
	if uc.builtinPath != "" {
		if fileSeeds, err := uc.loadSeedTemplatesFromDir(); err == nil && len(fileSeeds) > 0 {
			seeds = fileSeeds
		}
	}
	if err := uc.repo.LoadBuiltinTemplates(ctx, seeds); err != nil {
		return fmt.Errorf("save builtin templates: %w", err)
	}
	return nil
}

func (uc *TemplateUseCase) loadSeedTemplatesFromDir() ([]*domaintemplate.Template, error) {
	files, err := filepath.Glob(filepath.Join(uc.builtinPath, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("find builtin templates: %w", err)
	}
	var templates []*domaintemplate.Template
	for _, file := range files {
		data, readErr := os.ReadFile(file)
		if readErr != nil {
			return nil, fmt.Errorf("read template file %s: %w", file, readErr)
		}
		tmpl, parseErr := domaintemplate.FromJSON(data)
		if parseErr != nil {
			continue
		}
		templates = append(templates, tmpl)
	}
	return templates, nil
}

func generateTemplateID() string {
	return fmt.Sprintf("tpl_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
}

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

// CloneTemplate is kept for backward compatibility and does not persist.
func (uc *TemplateUseCase) CloneTemplate(ctx context.Context, id string) (*domaintemplate.Template, error) {
	original, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	cloned, err := original.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone template: %w", err)
	}
	cloned.ID = generateTemplateID()
	cloned.Name = fmt.Sprintf("%s Copy", original.Name)
	cloned.IsBuiltin = false
	cloned.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	cloned.Normalize()
	return cloned, nil
}

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

type TemplateMetadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tool        string   `json:"tool"`
	IsBuiltin   bool     `json:"is_builtin"`
	ParamCount  int      `json:"param_count"`
	DBTypes     []string `json:"database_types"`
}

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
		IsBuiltin:   tmpl.IsBuiltin,
		ParamCount:  len(tmpl.Parameters),
		DBTypes:     tmpl.DatabaseTypes,
	}, nil
}

func (uc *TemplateUseCase) ValidateTemplateParameters(ctx context.Context, templateID string, params map[string]interface{}) error {
	tmpl, err := uc.repo.FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("get template: %w", err)
	}
	for name, value := range params {
		param, ok := tmpl.Parameters[name]
		if !ok {
			continue
		}
		switch param.Type {
		case domaintemplate.ParameterTypeInteger:
			switch val := value.(type) {
			case int:
				if param.Min != nil && val < *param.Min {
					return fmt.Errorf("parameter '%s': value %d < min %d", name, val, *param.Min)
				}
				if param.Max != nil && val > *param.Max {
					return fmt.Errorf("parameter '%s': value %d > max %d", name, val, *param.Max)
				}
			case float64:
				intVal := int(val)
				if param.Min != nil && intVal < *param.Min {
					return fmt.Errorf("parameter '%s': value %d < min %d", name, intVal, *param.Min)
				}
				if param.Max != nil && intVal > *param.Max {
					return fmt.Errorf("parameter '%s': value %d > max %d", name, intVal, *param.Max)
				}
			default:
				return fmt.Errorf("parameter '%s': expected integer, got %T", name, value)
			}
		case domaintemplate.ParameterTypeEnum:
			strVal, ok := value.(string)
			if !ok {
				return fmt.Errorf("parameter '%s': expected string, got %T", name, value)
			}
			valid := false
			for _, opt := range param.Options {
				if opt == strVal {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("parameter '%s': value '%s' is not in options", name, strVal)
			}
		}
	}
	return nil
}

func (uc *TemplateUseCase) SubstituteTemplateParams(cmdTemplate string, params map[string]interface{}, connectionStr string) (string, error) {
	result := strings.ReplaceAll(cmdTemplate, "{connection_string}", connectionStr)
	for key, value := range params {
		result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", key), fmt.Sprintf("%v", value))
	}
	return result, nil
}

func (uc *TemplateUseCase) GetTemplateJSONWithMetadata(ctx context.Context, id string) ([]byte, error) {
	metadata, err := uc.GetTemplateMetadata(ctx, id)
	if err != nil {
		return nil, err
	}
	tmpl, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	export := struct {
		*domaintemplate.Template
		Metadata *TemplateMetadata `json:"_metadata"`
	}{
		Template: tmpl,
		Metadata: metadata,
	}
	return json.MarshalIndent(export, "", "  ")
}

func (uc *TemplateUseCase) prepareTemplateForCreate(tmpl *domaintemplate.Template) {
	now := time.Now().UTC().Format(time.RFC3339)
	if tmpl.ID == "" {
		tmpl.ID = generateTemplateID()
	}
	tmpl.Normalize()
	tmpl.CreatedAt = now
	tmpl.IsBuiltin = false
	tmpl.Readonly = false
}

func (uc *TemplateUseCase) prepareTemplateForUpdate(existing, incoming *domaintemplate.Template) {
	incoming.Normalize()
	incoming.ID = existing.ID
	incoming.IsBuiltin = existing.IsBuiltin
	incoming.Readonly = existing.IsReadOnly()
	incoming.CreatedAt = existing.CreatedAt
}
