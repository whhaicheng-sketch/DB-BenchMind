// Package repository provides unit tests for template repository.
package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

func TestTemplateRepository_Save_FindByID(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	// Create test template
	tmpl := &template.Template{
		ID:            "test-template",
		Name:          "Test Template",
		Description:   "A test template",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql", "postgresql"},
		Version:       "1.0.0",
		Parameters: map[string]template.Parameter{
			"threads": {
				Type:    template.ParameterTypeInteger,
				Label:   "Threads",
				Default: 8,
			},
		},
		CommandTemplate: template.CommandTemplate{
			Run: "sysbench run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}

	// Save
	err := repo.Save(ctx, tmpl)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(ctx, "test-template")
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}

	if found.ID != tmpl.ID {
		t.Errorf("ID = %s, want %s", found.ID, tmpl.ID)
	}
	if found.Name != tmpl.Name {
		t.Errorf("Name = %s, want %s", found.Name, tmpl.Name)
	}
	if found.Tool != tmpl.Tool {
		t.Errorf("Tool = %s, want %s", found.Tool, tmpl.Tool)
	}
}

func TestTemplateRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	_, err := repo.FindByID(ctx, "nonexistent")
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound, got: %v", err)
	}
}

func TestTemplateRepository_Save_Update(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	// Save original template
	tmpl := &template.Template{
		ID:            "test-template",
		Name:          "Original Name",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run command",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}

	err := repo.Save(ctx, tmpl)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Update
	tmpl.Name = "Updated Name"
	tmpl.DatabaseTypes = []string{"mysql", "postgresql"}
	err = repo.Save(ctx, tmpl)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	// Verify
	found, err := repo.FindByID(ctx, "test-template")
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}

	if found.Name != "Updated Name" {
		t.Errorf("Name = %s, want 'Updated Name'", found.Name)
	}
	if len(found.DatabaseTypes) != 2 {
		t.Errorf("DatabaseTypes count = %d, want 2", len(found.DatabaseTypes))
	}
}

func TestTemplateRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	// Save multiple templates
	templates := []*template.Template{
		{
			ID:            "template-2",
			Name:          "B Template",
			Tool:          "sysbench",
			DatabaseTypes: []string{"mysql"},
			CommandTemplate: template.CommandTemplate{
				Run: "run",
			},
			OutputParser: template.OutputParser{
				Type: template.ParserTypeRegex,
			},
		},
		{
			ID:            "template-1",
			Name:          "A Template",
			Tool:          "hammerdb",
			DatabaseTypes: []string{"postgresql"},
			CommandTemplate: template.CommandTemplate{
				Run: "run",
			},
			OutputParser: template.OutputParser{
				Type: template.ParserTypeRegex,
			},
		},
	}

	for _, tmpl := range templates {
		if err := repo.Save(ctx, tmpl); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
	}

	// FindAll
	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("FindAll() count = %d, want 2", len(all))
	}

	// Verify ordering (by tool, then name)
	// hammerdb comes before sysbench
	if all[0].ID != "template-1" {
		t.Errorf("First template ID = %s, want 'template-1'", all[0].ID)
	}
	if all[1].ID != "template-2" {
		t.Errorf("Second template ID = %s, want 'template-2'", all[1].ID)
	}
}

func TestTemplateRepository_FindBuiltin(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	// Load builtin templates
	builtin := []*template.Template{
		{
			ID:            "builtin-1",
			Name:          "Builtin 1",
			Tool:          "sysbench",
			DatabaseTypes: []string{"mysql"},
			CommandTemplate: template.CommandTemplate{
				Run: "run",
			},
			OutputParser: template.OutputParser{
				Type: template.ParserTypeRegex,
			},
		},
	}

	err := repo.LoadBuiltinTemplates(ctx, builtin)
	if err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	// Save a custom template
	custom := &template.Template{
		ID:            "custom-1",
		Name:          "Custom 1",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	err = repo.Save(ctx, custom)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// FindBuiltin should only return builtin templates
	builtinFound, err := repo.FindBuiltin(ctx)
	if err != nil {
		t.Fatalf("FindBuiltin() failed: %v", err)
	}

	if len(builtinFound) != 1 {
		t.Fatalf("FindBuiltin() count = %d, want 1", len(builtinFound))
	}
	if builtinFound[0].ID != "builtin-1" {
		t.Errorf("Builtin template ID = %s, want 'builtin-1'", builtinFound[0].ID)
	}
}

func TestTemplateRepository_FindCustom(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	// Load builtin templates
	builtin := []*template.Template{
		{
			ID:            "builtin-1",
			Name:          "Builtin 1",
			Tool:          "sysbench",
			DatabaseTypes: []string{"mysql"},
			CommandTemplate: template.CommandTemplate{
				Run: "run",
			},
			OutputParser: template.OutputParser{
				Type: template.ParserTypeRegex,
			},
		},
	}
	err := repo.LoadBuiltinTemplates(ctx, builtin)
	if err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	// Save a custom template
	custom := &template.Template{
		ID:            "custom-1",
		Name:          "Custom 1",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	err = repo.Save(ctx, custom)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// FindCustom should only return custom templates
	customFound, err := repo.FindCustom(ctx)
	if err != nil {
		t.Fatalf("FindCustom() failed: %v", err)
	}

	if len(customFound) != 1 {
		t.Fatalf("FindCustom() count = %d, want 1", len(customFound))
	}
	if customFound[0].ID != "custom-1" {
		t.Errorf("Custom template ID = %s, want 'custom-1'", customFound[0].ID)
	}
}

func TestTemplateRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	// Save a custom template
	tmpl := &template.Template{
		ID:            "test-template",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	err := repo.Save(ctx, tmpl)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Delete
	err = repo.Delete(ctx, "test-template")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deleted
	_, err = repo.FindByID(ctx, "test-template")
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound after Delete(), got: %v", err)
	}
}

func TestTemplateRepository_Delete_Builtin(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	// Load builtin template
	builtin := []*template.Template{
		{
			ID:            "builtin-1",
			Name:          "Builtin",
			Tool:          "sysbench",
			DatabaseTypes: []string{"mysql"},
			CommandTemplate: template.CommandTemplate{
				Run: "run",
			},
			OutputParser: template.OutputParser{
				Type: template.ParserTypeRegex,
			},
		},
	}
	err := repo.LoadBuiltinTemplates(ctx, builtin)
	if err != nil {
		t.Fatalf("LoadBuiltinTemplates() failed: %v", err)
	}

	// Try to delete builtin template
	err = repo.Delete(ctx, "builtin-1")
	if err != ErrBuiltinTemplateCannotBeDeleted {
		t.Errorf("Expected ErrBuiltinTemplateCannotBeDeleted, got: %v", err)
	}
}

func TestTemplateRepository_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupTemplateTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)

	err := repo.Delete(ctx, "nonexistent")
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound, got: %v", err)
	}
}

// setupTemplateTestDB creates an in-memory SQLite database for template testing.
func setupTemplateTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// Create templates table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS templates (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			tool TEXT NOT NULL,
			database_types TEXT NOT NULL,
			version TEXT NOT NULL,
			config_json TEXT NOT NULL,
			is_builtin INTEGER DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_templates_tool ON templates(tool);
		CREATE INDEX IF NOT EXISTS idx_templates_is_builtin ON templates(is_builtin);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("create tables: %v", err)
	}

	return db
}
