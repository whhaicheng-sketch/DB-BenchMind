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
	tmpl := newTemplate("repo-user-1", false)

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
	builtin := newTemplate("repo-builtin-1", true)
	user := newTemplate("repo-user-2", false)
	testTemplate := newTemplate("repo-test-1", false)

	if err := repo.LoadBuiltinTemplates(ctx, []*domaintemplate.Template{builtin}); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	if err := repo.Save(ctx, testTemplate); err != nil {
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
	if err != nil || len(custom) != 2 {
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
	tmpl := newTemplate("repo-persist-1", false)
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

func TestTemplateRepository_LoadBuiltinTemplates_SyncsToSingleDefaultTestTemplate(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	legacyBuiltin := newTemplate("legacy-builtin", true)
	legacyUser := newTemplate("legacy-user", false)
	legacyUser.ID = "legacy-user"
	legacyUser.Name = "legacy-user"
	legacyTest := newTemplate("tpl_test_postgresql_sysbench", false)
	legacyTest.DBFamily = "postgresql"
	legacyTest.DatabaseTypes = []string{"postgresql"}
	legacyTest.ToolConfig.Sysbench.DBDriver = "pgsql"

	for _, tmpl := range []*domaintemplate.Template{legacyBuiltin, legacyUser, legacyTest} {
		tmpl.Normalize()
		if err := repo.Save(ctx, tmpl); err != nil {
			t.Fatalf("seed legacy template %s: %v", tmpl.ID, err)
		}
	}

	if err := repo.LoadBuiltinTemplates(ctx, domaintemplate.DefaultSeedTemplates()); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 template after sync, got %d", len(all))
	}
	if all[0].ID != "tpl_test_mysql_sysbench" {
		t.Fatalf("remaining template id = %s, want tpl_test_mysql_sysbench", all[0].ID)
	}
	if !all[0].IsBuiltin {
		t.Fatal("remaining template should be builtin")
	}
}

func TestTemplateRepository_LoadBuiltinTemplates_ExistingDefaultTemplateKeepsStableRecord(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	defaultTemplate := newTemplate("tpl_test_mysql_sysbench", true)
	defaultTemplate.Name = "stale default"
	defaultTemplate.CreatedAt = "2026-01-01T00:00:00Z"
	defaultTemplate.Normalize()
	if err := repo.Save(ctx, defaultTemplate); err != nil {
		t.Fatalf("seed default template: %v", err)
	}
	if err := repo.Save(ctx, newTemplate("legacy-user", false)); err != nil {
		t.Fatalf("seed legacy user template: %v", err)
	}
	if err := repo.Save(ctx, newTemplate("legacy-builtin", true)); err != nil {
		t.Fatalf("seed legacy builtin template: %v", err)
	}

	beforeRowID := templateRowID(t, db, "tpl_test_mysql_sysbench")

	if err := repo.LoadBuiltinTemplates(ctx, domaintemplate.DefaultSeedTemplates()); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	afterRowID := templateRowID(t, db, "tpl_test_mysql_sysbench")
	if beforeRowID != afterRowID {
		t.Fatalf("default template rowid changed from %d to %d; template was recreated instead of updated", beforeRowID, afterRowID)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 template after sync, got %d", len(all))
	}
	if all[0].Name != "MySQL - Sysbench Test" {
		t.Fatalf("default template was not refreshed to latest seed content: %s", all[0].Name)
	}
}

func TestTemplateRepository_LoadBuiltinTemplates_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	for i := 0; i < 3; i++ {
		if err := repo.LoadBuiltinTemplates(ctx, domaintemplate.DefaultSeedTemplates()); err != nil {
			t.Fatalf("LoadBuiltinTemplates() run %d failed: %v", i+1, err)
		}
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 template after repeated sync, got %d", len(all))
	}
	if all[0].ID != "tpl_test_mysql_sysbench" {
		t.Fatalf("remaining template id = %s, want tpl_test_mysql_sysbench", all[0].ID)
	}
}

func TestTemplateRepository_LoadBuiltinTemplates_FreshInstallCreatesOnlyDefaultTemplate(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	if err := repo.LoadBuiltinTemplates(ctx, domaintemplate.DefaultSeedTemplates()); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 template after fresh install sync, got %d", len(all))
	}
	if all[0].ID != "tpl_test_mysql_sysbench" {
		t.Fatalf("remaining template id = %s, want tpl_test_mysql_sysbench", all[0].ID)
	}
}

func TestTemplateRepository_LoadBuiltinTemplates_PreservesDependenciesForDefaultTemplate(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	defaultTemplate := newTemplate("tpl_test_mysql_sysbench", true)
	defaultTemplate.Name = "stale default"
	defaultTemplate.Normalize()
	if err := repo.Save(ctx, defaultTemplate); err != nil {
		t.Fatalf("seed default template: %v", err)
	}
	if err := repo.Save(ctx, newTemplate("legacy-user", false)); err != nil {
		t.Fatalf("seed legacy user template: %v", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO tasks (id, name, connection_id, template_id, parameters_json, options_json, tags, created_at, updated_at)
		VALUES ('task-default', 'Task Default', 'conn-1', 'tpl_test_mysql_sysbench', '{}', '{}', '', '2026-03-20T00:00:00Z', '2026-03-20T00:00:00Z')
	`); err != nil {
		t.Fatalf("seed task referencing default template: %v", err)
	}

	if err := repo.LoadBuiltinTemplates(ctx, domaintemplate.DefaultSeedTemplates()); err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = 'task-default'").Scan(&count); err != nil {
		t.Fatalf("count dependent tasks: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected dependent task to survive sync, got count %d", count)
	}
	if templateRowID(t, db, "tpl_test_mysql_sysbench") == 0 {
		t.Fatal("default template should still exist after sync")
	}
}

func setupTemplateTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		db.Close()
		t.Fatalf("enable foreign keys: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE connections (
			id TEXT PRIMARY KEY
		);
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
			config_json TEXT NOT NULL DEFAULT '',
			is_builtin INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE tasks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			connection_id TEXT NOT NULL,
			template_id TEXT NOT NULL,
			parameters_json TEXT NOT NULL,
			options_json TEXT,
			tags TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (connection_id) REFERENCES connections(id) ON DELETE CASCADE,
			FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE
		);
		INSERT INTO connections (id) VALUES ('conn-1');
	`)
	if err != nil {
		db.Close()
		t.Fatalf("create templates table: %v", err)
	}

	return db
}

func templateRowID(t *testing.T, db *sql.DB, id string) int64 {
	t.Helper()

	var rowID int64
	if err := db.QueryRow("SELECT rowid FROM templates WHERE id = ?", id).Scan(&rowID); err != nil {
		t.Fatalf("query template rowid for %s: %v", id, err)
	}
	return rowID
}

func newTemplate(id string, isBuiltin bool) *domaintemplate.Template {
	tmpl := &domaintemplate.Template{
		ID:             id,
		Name:           id,
		Description:    "repository test template",
		Tool:           domaintemplate.ToolSysbench,
		DBFamily:       "mysql",
		WorkloadFamily: "oltp-read-write",
		IsBuiltin:      isBuiltin,
		Version:        "1.0.0",
		Phases: domaintemplate.PhaseSet{
			Prepare: domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			Warmup:  domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
			Cleanup: domaintemplate.PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
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
