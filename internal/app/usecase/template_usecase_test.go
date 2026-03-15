package usecase

import (
	"context"
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
		if tmpl.Scope == domaintemplate.ScopeBuiltin {
			clone, _ := tmpl.Clone()
			out = append(out, clone)
		}
	}
	return out, nil
}

func (m *mockTemplateRepository) FindCustom(ctx context.Context) ([]*domaintemplate.Template, error) {
	var out []*domaintemplate.Template
	for _, tmpl := range m.templates {
		if tmpl.Scope == domaintemplate.ScopeUser || tmpl.Scope == domaintemplate.ScopeProject || tmpl.Scope == domaintemplate.ScopeTest {
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
	for _, tmpl := range templates {
		clone, _ := tmpl.Clone()
		m.templates[tmpl.ID] = clone
	}
	return nil
}

func TestTemplateUseCase_CreateUpdateDeleteAndDuplicate(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	tmpl := newCanonicalTemplate("uc-user-1", domaintemplate.ScopeUser)
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
	if dup.ID == created.ID || dup.Scope != domaintemplate.ScopeUser {
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

	builtin := newCanonicalTemplate("uc-builtin-1", domaintemplate.ScopeBuiltin)
	shared := newCanonicalTemplate("uc-shared-1", domaintemplate.ScopeReadonlyShared)
	if err := repo.LoadBuiltinTemplates(ctx, []*domaintemplate.Template{builtin, shared}); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	builtin.Description = "mutated"
	if err := uc.UpdateTemplate(ctx, builtin); err != ErrReadonlyTemplateCannotBeEdited {
		t.Fatalf("UpdateTemplate() readonly err=%v want=%v", err, ErrReadonlyTemplateCannotBeEdited)
	}
	if err := uc.DeleteTemplate(ctx, builtin.ID); err != ErrBuiltinTemplateCannotBeDeleted {
		t.Fatalf("DeleteTemplate() builtin err=%v want=%v", err, ErrBuiltinTemplateCannotBeDeleted)
	}
	if err := uc.DeleteTemplate(ctx, shared.ID); err != ErrReadonlyTemplateCannotBeEdited {
		t.Fatalf("DeleteTemplate() shared err=%v want=%v", err, ErrReadonlyTemplateCannotBeEdited)
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
	if len(all) == 0 {
		t.Fatal("expected seeded templates")
	}
}

func TestTemplateUseCase_LoadBuiltinTemplates_SwingbenchSupportsPrepareRunCleanup(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	if err := uc.LoadBuiltinTemplates(ctx); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	tmpl, err := uc.GetTemplate(ctx, "tpl_swing_oe")
	if err != nil {
		t.Fatalf("GetTemplate() failed: %v", err)
	}

	if !tmpl.Phases.Prepare.Enabled {
		t.Fatal("swingbench builtin should enable prepare phase")
	}
	if !tmpl.Phases.Run.Enabled {
		t.Fatal("swingbench builtin should enable run phase")
	}
	if !tmpl.Phases.Cleanup.Enabled {
		t.Fatal("swingbench builtin should enable cleanup phase")
	}
}

func TestTemplateUseCase_LoadBuiltinTemplates_IncludesDatabaseTestCoverage(t *testing.T) {
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

	want := map[string]string{
		"mysql":      domaintemplate.ToolSysbench,
		"postgresql": domaintemplate.ToolSysbench,
		"oracle":     domaintemplate.ToolSwingbench,
		"sqlserver":  domaintemplate.ToolHammerDB,
	}

	found := map[string]*domaintemplate.Template{}
	for _, tmpl := range all {
		if tmpl.Scope != domaintemplate.ScopeTest {
			continue
		}
		found[tmpl.DBFamily] = tmpl
	}

	for dbFamily, tool := range want {
		tmpl := found[dbFamily]
		if tmpl == nil {
			t.Fatalf("missing test template for %s", dbFamily)
		}
		if tmpl.Tool != tool {
			t.Fatalf("test template for %s uses tool %s, want %s", dbFamily, tmpl.Tool, tool)
		}
		if tmpl.Scope != domaintemplate.ScopeTest {
			t.Fatalf("test template for %s scope = %s, want %s", dbFamily, tmpl.Scope, domaintemplate.ScopeTest)
		}
		if !containsString(tmpl.Tags, "test") {
			t.Fatalf("test template for %s must include tag 'test': %v", dbFamily, tmpl.Tags)
		}
		if !tmpl.SupportsDatabase(dbFamily) {
			t.Fatalf("test template for %s does not support its own database family", dbFamily)
		}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func newCanonicalTemplate(id, scope string) *domaintemplate.Template {
	tmpl := &domaintemplate.Template{
		ID:             id,
		Name:           id,
		Description:    "usecase test template",
		Tool:           domaintemplate.ToolSysbench,
		DBFamily:       "mysql",
		WorkloadFamily: "oltp-read-write",
		Scope:          scope,
		Status:         domaintemplate.StatusDraft,
		Tags:           []string{"test"},
		Version:        "0.1.0",
		Phases: domaintemplate.PhaseSet{
			Prepare: domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
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
