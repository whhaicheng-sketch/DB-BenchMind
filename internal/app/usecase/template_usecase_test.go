package usecase

import (
	"context"
	"strings"
	"testing"

	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

type mockTemplateRepository struct {
	templates map[string]*domaintemplate.Template
}

func newMockTemplateRepository() *mockTemplateRepository {
	return &mockTemplateRepository{templates: map[string]*domaintemplate.Template{}}
}

func (m *mockTemplateRepository) Save(ctx context.Context, tmpl *domaintemplate.Template) error {
	clone, _ := tmpl.Clone()
	m.templates[tmpl.ID] = clone
	return nil
}

func (m *mockTemplateRepository) FindByID(ctx context.Context, id string) (*domaintemplate.Template, error) {
	tmpl, ok := m.templates[id]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	return tmpl.Clone()
}

func (m *mockTemplateRepository) FindAll(ctx context.Context) ([]*domaintemplate.Template, error) {
	var out []*domaintemplate.Template
	for _, tmpl := range m.templates {
		clone, _ := tmpl.Clone()
		out = append(out, clone)
	}
	return out, nil
}

func (m *mockTemplateRepository) FindBuiltin(ctx context.Context) ([]*domaintemplate.Template, error) {
	var out []*domaintemplate.Template
	for _, tmpl := range m.templates {
		if tmpl.IsBuiltin {
			clone, _ := tmpl.Clone()
			out = append(out, clone)
		}
	}
	return out, nil
}

func (m *mockTemplateRepository) FindCustom(ctx context.Context) ([]*domaintemplate.Template, error) {
	var out []*domaintemplate.Template
	for _, tmpl := range m.templates {
		if !tmpl.IsBuiltin {
			clone, _ := tmpl.Clone()
			out = append(out, clone)
		}
	}
	return out, nil
}

func (m *mockTemplateRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.templates[id]; !ok {
		return ErrTemplateNotFound
	}
	delete(m.templates, id)
	return nil
}

func (m *mockTemplateRepository) LoadBuiltinTemplates(ctx context.Context, templates []*domaintemplate.Template) error {
	keep := map[string]struct{}{}
	for _, tmpl := range templates {
		clone, _ := tmpl.Clone()
		m.templates[tmpl.ID] = clone
		keep[tmpl.ID] = struct{}{}
	}
	for id, tmpl := range m.templates {
		if !tmpl.IsBuiltin {
			continue
		}
		if _, ok := keep[id]; !ok {
			delete(m.templates, id)
		}
	}
	return nil
}

func TestTemplateUseCase_CreateUpdateDeleteAndDuplicate(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	tmpl := newCanonicalTemplate("uc-user-1", false)
	if err := uc.CreateTemplate(ctx, tmpl); err != nil {
		t.Fatalf("CreateTemplate() failed: %v", err)
	}

	created, err := uc.GetTemplate(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("GetTemplate() failed: %v", err)
	}
	created.Description = "updated"
	if err := uc.UpdateTemplate(ctx, created); err != nil {
		t.Fatalf("UpdateTemplate() failed: %v", err)
	}

	dup, err := uc.DuplicateTemplate(ctx, created.ID)
	if err != nil {
		t.Fatalf("DuplicateTemplate() failed: %v", err)
	}
	if dup.ID == created.ID || dup.IsBuiltin {
		t.Fatalf("duplicate not normalized correctly: %+v", dup)
	}

	if err := uc.DeleteTemplate(ctx, created.ID); err != nil {
		t.Fatalf("DeleteTemplate() failed: %v", err)
	}
}

func TestTemplateUseCase_ReadonlyProtection(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	builtin := newCanonicalTemplate("uc-builtin-1", true)
	if err := repo.LoadBuiltinTemplates(ctx, []*domaintemplate.Template{builtin}); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	builtin.Description = "mutated"
	if err := uc.UpdateTemplate(ctx, builtin); err != ErrReadonlyTemplateCannotBeEdited {
		t.Fatalf("UpdateTemplate() readonly err=%v want=%v", err, ErrReadonlyTemplateCannotBeEdited)
	}
	if err := uc.DeleteTemplate(ctx, builtin.ID); err != ErrBuiltinTemplateCannotBeDeleted {
		t.Fatalf("DeleteTemplate() builtin err=%v want=%v", err, ErrBuiltinTemplateCannotBeDeleted)
	}
}

func TestTemplateUseCase_LoadBuiltinTemplates(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	if err := uc.LoadBuiltinTemplates(ctx); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	all, err := uc.ListTemplates(ctx)
	if err != nil {
		t.Fatalf("ListTemplates() failed: %v", err)
	}
	if len(all) != 12 {
		t.Fatalf("expected 12 seeded templates, got %d", len(all))
	}
	for _, id := range []string{"oracle_cpu_bound", "mysql_test", "sqlserver_io_bound", "postgresql_cpu_bound"} {
		if _, err := uc.GetTemplate(ctx, id); err != nil {
			t.Fatalf("GetTemplate(%s) failed: %v", id, err)
		}
	}
}

func TestTemplateUseCase_LoadBuiltinTemplates_KeepsCustomTemplatesAndRefreshesBuiltinSet(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	custom := newCanonicalTemplate("user-template", false)
	if err := uc.CreateTemplate(ctx, custom); err != nil {
		t.Fatalf("CreateTemplate() failed: %v", err)
	}

	if err := uc.LoadBuiltinTemplates(ctx); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	all, err := uc.ListTemplates(ctx)
	if err != nil {
		t.Fatalf("ListTemplates() failed: %v", err)
	}

	if len(all) != 13 {
		t.Fatalf("expected 13 templates after loading seeds with one custom template, got %d", len(all))
	}

	if _, err := uc.GetTemplate(ctx, "user-template"); err != nil {
		t.Fatalf("custom template should survive builtin reload: %v", err)
	}
}

func TestTemplateUseCase_CreateTemplate_DoesNotRequireRemovedMetadataFields(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	tmpl := newCanonicalTemplate("uc-minimal-1", false)
	tmpl.IsBuiltin = false

	if err := uc.CreateTemplate(ctx, tmpl); err != nil {
		t.Fatalf("CreateTemplate() failed without scope/status/tags/updatedAt: %v", err)
	}

	saved, err := uc.GetTemplate(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("GetTemplate() failed: %v", err)
	}
	if saved.IsBuiltin {
		t.Fatal("saved template should stay editable")
	}
}

func TestTemplateUseCase_DuplicateBuiltinTemplateCreatesEditableCopy(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	if err := uc.LoadBuiltinTemplates(ctx); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	copy, err := uc.DuplicateTemplate(ctx, "oracle_cpu_bound")
	if err != nil {
		t.Fatalf("DuplicateTemplate() failed: %v", err)
	}
	if copy.IsBuiltin {
		t.Fatal("duplicate of builtin template must become custom")
	}
	if copy.ProfileType != "cpu_bound" {
		t.Fatalf("duplicate profileType = %s, want cpu_bound", copy.ProfileType)
	}
	if !strings.Contains(copy.Name, "Copy") {
		t.Fatalf("duplicate name = %s, want Copy suffix", copy.Name)
	}

	copy.Description = "customized copy"
	if err := uc.UpdateTemplate(ctx, copy); err != nil {
		t.Fatalf("UpdateTemplate() for duplicated builtin failed: %v", err)
	}
}

func newCanonicalTemplate(id string, isBuiltin bool) *domaintemplate.Template {
	tmpl := &domaintemplate.Template{
		ID:             id,
		Name:           id,
		Description:    "usecase test template",
		Tool:           domaintemplate.ToolSysbench,
		DBFamily:       "mysql",
		WorkloadFamily: "oltp-read-write",
		IsBuiltin:      isBuiltin,
		Version:        "0.1.0",
		Phases: domaintemplate.PhaseSet{
			Prepare: domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			Warmup:  domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
			Cleanup: domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
		},
		Runtime: domaintemplate.Runtime{
			Concurrency:           domaintemplate.Concurrency{Mode: "threads", Value: 4},
			DurationSeconds:       60,
			ReportIntervalSeconds: 10,
			Percentile:            95,
		},
		ToolConfig: domaintemplate.ToolConfig{
			Sysbench: domaintemplate.SysbenchConfig{
				DBDriver:   "mysql",
				ScriptType: "oltp_read_write",
				Tables:     1,
				TableSize:  1000,
			},
		},
	}
	tmpl.Normalize()
	return tmpl
}
