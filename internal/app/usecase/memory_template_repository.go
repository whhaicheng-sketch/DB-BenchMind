// Package usecase provides in-memory template repository for testing and development.
// TODO: Replace with persistent implementation for production
package usecase

import (
	"context"
	"log/slog"
	"sync"

	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// MemoryTemplateRepository provides an in-memory implementation of TemplateRepository.
// This is a temporary implementation for development.
type MemoryTemplateRepository struct {
	templates          map[string]*domaintemplate.Template
	builtinTemplateIDs map[string]bool // Track which templates are builtin
	mu                 sync.RWMutex
}

// NewMemoryTemplateRepository creates a new in-memory template repository.
func NewMemoryTemplateRepository() *MemoryTemplateRepository {
	return &MemoryTemplateRepository{
		templates:          make(map[string]*domaintemplate.Template),
		builtinTemplateIDs: make(map[string]bool),
	}
}

// Save saves a template to the repository.
func (r *MemoryTemplateRepository) Save(ctx context.Context, tmpl *domaintemplate.Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.templates[tmpl.ID] = tmpl
	slog.Debug("MemoryTemplateRepository: Saved template", "id", tmpl.ID, "name", tmpl.Name)
	return nil
}

// FindByID finds a template by its ID.
func (r *MemoryTemplateRepository) FindByID(ctx context.Context, id string) (*domaintemplate.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tmpl, ok := r.templates[id]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	return tmpl, nil
}

// FindAll finds all templates.
func (r *MemoryTemplateRepository) FindAll(ctx context.Context) ([]*domaintemplate.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var templates []*domaintemplate.Template
	for _, tmpl := range r.templates {
		templates = append(templates, tmpl)
	}
	return templates, nil
}

// FindBuiltin finds all builtin templates.
func (r *MemoryTemplateRepository) FindBuiltin(ctx context.Context) ([]*domaintemplate.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var templates []*domaintemplate.Template
	for id := range r.builtinTemplateIDs {
		if tmpl, ok := r.templates[id]; ok {
			templates = append(templates, tmpl)
		}
	}
	return templates, nil
}

// FindCustom finds all user-defined (non-builtin) templates.
func (r *MemoryTemplateRepository) FindCustom(ctx context.Context) ([]*domaintemplate.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var templates []*domaintemplate.Template
	for _, tmpl := range r.templates {
		templates = append(templates, tmpl)
	}
	return templates, nil
}

// Update updates a template.
func (r *MemoryTemplateRepository) Update(ctx context.Context, tmpl *domaintemplate.Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.templates[tmpl.ID]; !ok {
		return ErrTemplateNotFound
	}
	r.templates[tmpl.ID] = tmpl
	return nil
}

// Delete deletes a template by its ID.
func (r *MemoryTemplateRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.templates, id)
	return nil
}

// LoadBuiltinTemplates loads builtin templates into the repository.
func (r *MemoryTemplateRepository) LoadBuiltinTemplates(ctx context.Context, templates []*domaintemplate.Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, tmpl := range templates {
		r.templates[tmpl.ID] = tmpl
		r.builtinTemplateIDs[tmpl.ID] = true
	}
	slog.Info("MemoryTemplateRepository: Loaded builtin templates", "count", len(templates))
	return nil
}
