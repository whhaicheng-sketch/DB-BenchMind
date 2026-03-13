package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
)

func TestTemplateRepository_SaveFindAndDelete(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	tmpl := newTemplate("repo-user-1", domaintemplate.ScopeUser)

	if err := repo.Save(ctx, tmpl); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	found, err := repo.FindByID(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if found.Name != tmpl.Name || found.DBFamily != tmpl.DBFamily {
		t.Fatalf("FindByID() returned unexpected template: %+v", found)
	}

	if err := repo.Delete(ctx, tmpl.ID); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	if _, err := repo.FindByID(ctx, tmpl.ID); err != ErrTemplateNotFound {
		t.Fatalf("FindByID() after delete error = %v, want %v", err, ErrTemplateNotFound)
	}
}

func TestTemplateRepository_FindFiltersAndBuiltinProtection(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	builtin := newTemplate("repo-builtin-1", domaintemplate.ScopeBuiltin)
	shared := newTemplate("repo-shared-1", domaintemplate.ScopeReadonlyShared)
	user := newTemplate("repo-user-2", domaintemplate.ScopeUser)

	if err := repo.LoadBuiltinTemplates(ctx, []*domaintemplate.Template{builtin, shared}); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	all, err := repo.FindAll(ctx)
	if err != nil || len(all) != 3 {
		t.Fatalf("FindAll() err=%v len=%d", err, len(all))
	}

	builtins, err := repo.FindBuiltin(ctx)
	if err != nil || len(builtins) != 1 {
		t.Fatalf("FindBuiltin() err=%v len=%d", err, len(builtins))
	}

	custom, err := repo.FindCustom(ctx)
	if err != nil || len(custom) != 1 {
		t.Fatalf("FindCustom() err=%v len=%d", err, len(custom))
	}

	if err := repo.Delete(ctx, builtin.ID); err != ErrBuiltinTemplateCannotBeDeleted {
		t.Fatalf("Delete builtin err=%v want=%v", err, ErrBuiltinTemplateCannotBeDeleted)
	}
}

func TestTemplateRepository_PersistsAcrossRestart(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "templates.db")

	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("InitializeSQLite() failed: %v", err)
	}
	repo := NewTemplateRepository(db)
	tmpl := newTemplate("repo-persist-1", domaintemplate.ScopeUser)
	if err := repo.Save(ctx, tmpl); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	db.Close()

	reopened, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("reopen sqlite failed: %v", err)
	}
	defer reopened.Close()

	reopenedRepo := NewTemplateRepository(reopened)
	found, err := reopenedRepo.FindByID(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("FindByID() after reopen failed: %v", err)
	}
	if found.ID != tmpl.ID {
		t.Fatalf("persisted template id = %s, want %s", found.ID, tmpl.ID)
	}
}

func setupTemplateTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE templates (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			tool TEXT NOT NULL,
			database_types TEXT NOT NULL,
			version TEXT NOT NULL,
			parameters_json TEXT NOT NULL,
			command_template_json TEXT NOT NULL,
			output_parser_json TEXT NOT NULL,
			scope TEXT NOT NULL DEFAULT 'builtin',
			status TEXT NOT NULL DEFAULT 'ready',
			tags_json TEXT NOT NULL DEFAULT '[]',
			config_json TEXT NOT NULL DEFAULT '',
			is_builtin INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("create templates table: %v", err)
	}

	return db
}

func newTemplate(id, scope string) *domaintemplate.Template {
	tmpl := &domaintemplate.Template{
		ID:             id,
		Name:           id,
		Description:    "repository test template",
		Tool:           domaintemplate.ToolSysbench,
		DBFamily:       "mysql",
		WorkloadFamily: "oltp-read-write",
		Scope:          scope,
		Status:         domaintemplate.StatusReady,
		Tags:           []string{"repo"},
		Version:        "1.0.0",
		Phases: domaintemplate.PhaseSet{
			Prepare: domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
		},
		Runtime: domaintemplate.Runtime{
			Concurrency:           domaintemplate.Concurrency{Mode: "threads", Value: 8},
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
