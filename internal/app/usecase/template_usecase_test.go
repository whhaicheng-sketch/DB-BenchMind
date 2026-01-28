// Package usecase provides unit tests for template use case.
package usecase

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// mockTemplateRepository is a mock implementation of TemplateRepository for testing.
type mockTemplateRepository struct {
	templates map[string]*template.Template
	builtin   map[string]bool
}

func newMockTemplateRepository() *mockTemplateRepository {
	return &mockTemplateRepository{
		templates: make(map[string]*template.Template),
		builtin:   make(map[string]bool),
	}
}

func (m *mockTemplateRepository) Save(ctx context.Context, tmpl *template.Template) error {
	m.templates[tmpl.ID] = tmpl
	return nil
}

func (m *mockTemplateRepository) FindByID(ctx context.Context, id string) (*template.Template, error) {
	tmpl, ok := m.templates[id]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	return tmpl, nil
}

func (m *mockTemplateRepository) FindAll(ctx context.Context) ([]*template.Template, error) {
	var result []*template.Template
	for _, tmpl := range m.templates {
		result = append(result, tmpl)
	}
	return result, nil
}

func (m *mockTemplateRepository) FindBuiltin(ctx context.Context) ([]*template.Template, error) {
	var result []*template.Template
	for id, tmpl := range m.templates {
		if m.builtin[id] {
			result = append(result, tmpl)
		}
	}
	return result, nil
}

func (m *mockTemplateRepository) FindCustom(ctx context.Context) ([]*template.Template, error) {
	var result []*template.Template
	for id, tmpl := range m.templates {
		if !m.builtin[id] {
			result = append(result, tmpl)
		}
	}
	return result, nil
}

func (m *mockTemplateRepository) Delete(ctx context.Context, id string) error {
	if m.builtin[id] {
		return ErrBuiltinTemplateCannotBeDeleted
	}
	if _, ok := m.templates[id]; !ok {
		return ErrTemplateNotFound
	}
	delete(m.templates, id)
	return nil
}

func (m *mockTemplateRepository) LoadBuiltinTemplates(ctx context.Context, templates []*template.Template) error {
	for _, tmpl := range templates {
		m.templates[tmpl.ID] = tmpl
		m.builtin[tmpl.ID] = true
	}
	return nil
}

// TestTemplateUseCase_ListTemplates tests listing all templates.
func TestTemplateUseCase_ListTemplates(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	// Create test templates
	tmpl1 := &template.Template{
		ID:            "test-1",
		Name:          "Test 1",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	tmpl2 := &template.Template{
		ID:            "test-2",
		Name:          "Test 2",
		Tool:          "hammerdb",
		DatabaseTypes: []string{"postgresql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}

	repo.Save(ctx, tmpl1)
	repo.Save(ctx, tmpl2)

	// List templates
	all, err := uc.ListTemplates(ctx)
	if err != nil {
		t.Fatalf("ListTemplates() failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("ListTemplates() count = %d, want 2", len(all))
	}
}

// TestTemplateUseCase_GetTemplate tests getting a template by ID.
func TestTemplateUseCase_GetTemplate(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	tmpl := &template.Template{
		ID:            "test-1",
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
	repo.Save(ctx, tmpl)

	// Get existing template
	found, err := uc.GetTemplate(ctx, "test-1")
	if err != nil {
		t.Fatalf("GetTemplate() failed: %v", err)
	}
	if found.ID != "test-1" {
		t.Errorf("GetTemplate() ID = %s, want 'test-1'", found.ID)
	}

	// Get non-existing template
	_, err = uc.GetTemplate(ctx, "nonexistent")
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound, got: %v", err)
	}
}

// TestTemplateUseCase_ImportTemplate tests importing a template from file.
func TestTemplateUseCase_ImportTemplate(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	// Create a test template file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-template.json")

	tmpl := &template.Template{
		ID:            "imported-template",
		Name:          "Imported Template",
		Description:   "A test template",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
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

	data, err := tmpl.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize template: %v", err)
	}

	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Import template
	imported, err := uc.ImportTemplate(ctx, testFile)
	if err != nil {
		t.Fatalf("ImportTemplate() failed: %v", err)
	}

	if imported.ID != "imported-template" {
		t.Errorf("ImportTemplate() ID = %s, want 'imported-template'", imported.ID)
	}
	if imported.Name != "Imported Template" {
		t.Errorf("ImportTemplate() Name = %s, want 'Imported Template'", imported.Name)
	}

	// Verify template was saved to repository
	saved, err := uc.GetTemplate(ctx, "imported-template")
	if err != nil {
		t.Errorf("Template was not saved to repository: %v", err)
	}
	if saved.Name != imported.Name {
		t.Errorf("Saved template Name = %s, want %s", saved.Name, imported.Name)
	}
}

// TestTemplateUseCase_ImportTemplate_InvalidJSON tests importing invalid JSON.
func TestTemplateUseCase_ImportTemplate_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	// Create invalid JSON file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(testFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Import should fail
	_, err := uc.ImportTemplate(ctx, testFile)
	if err == nil {
		t.Error("ImportTemplate() with invalid JSON should return error")
	}
}

// TestTemplateUseCase_ValidateTemplateForDatabase tests database compatibility validation.
func TestTemplateUseCase_ValidateTemplateForDatabase(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	tmpl := &template.Template{
		ID:            "test-1",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql", "postgresql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	repo.Save(ctx, tmpl)

	// Valid database type
	err := uc.ValidateTemplateForDatabase(ctx, "test-1", "mysql")
	if err != nil {
		t.Errorf("ValidateTemplateForDatabase() with valid type failed: %v", err)
	}

	// Invalid database type
	err = uc.ValidateTemplateForDatabase(ctx, "test-1", "oracle")
	if err == nil {
		t.Error("ValidateTemplateForDatabase() with invalid type should return error")
	}
}

// TestTemplateUseCase_DeleteTemplate tests deleting templates.
func TestTemplateUseCase_DeleteTemplate(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	// Create custom template
	tmpl := &template.Template{
		ID:            "custom-1",
		Name:          "Custom",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	repo.Save(ctx, tmpl)

	// Delete custom template
	err := uc.DeleteTemplate(ctx, "custom-1")
	if err != nil {
		t.Fatalf("DeleteTemplate() failed: %v", err)
	}

	// Verify deleted
	_, err = uc.GetTemplate(ctx, "custom-1")
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound after Delete(), got: %v", err)
	}
}

// TestTemplateUseCase_DeleteTemplate_Builtin tests that builtin templates cannot be deleted.
func TestTemplateUseCase_DeleteTemplate_Builtin(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	// Create builtin template
	tmpl := &template.Template{
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
	}
	repo.LoadBuiltinTemplates(ctx, []*template.Template{tmpl})

	// Try to delete builtin template
	err := uc.DeleteTemplate(ctx, "builtin-1")
	if err != ErrBuiltinTemplateCannotBeDeleted {
		t.Errorf("Expected ErrBuiltinTemplateCannotBeDeleted, got: %v", err)
	}
}

// TestTemplateUseCase_ValidateTemplateParameters tests parameter validation.
func TestTemplateUseCase_ValidateTemplateParameters(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	tmpl := &template.Template{
		ID:            "test-1",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		Parameters: map[string]template.Parameter{
			"threads": {
				Type:    template.ParameterTypeInteger,
				Label:   "Threads",
				Default: 8,
				Min:     intPtr(1),
				Max:     intPtr(1024),
			},
			"mode": {
				Type:    parameterTypeEnum,
				Label:   "Mode",
				Default: "fast",
				Options: []string{"fast", "slow"},
			},
		},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	repo.Save(ctx, tmpl)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"threads": 8,
				"mode":    "fast",
			},
			wantErr: false,
		},
		{
			name: "unknown parameter",
			params: map[string]interface{}{
				"unknown": "value",
			},
			wantErr: true,
		},
		{
			name: "threads below minimum",
			params: map[string]interface{}{
				"threads": 0,
			},
			wantErr: true,
		},
		{
			name: "threads above maximum",
			params: map[string]interface{}{
				"threads": 2000,
			},
			wantErr: true,
		},
		{
			name: "invalid enum value",
			params: map[string]interface{}{
				"mode": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.ValidateTemplateParameters(ctx, "test-1", tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplateParameters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTemplateUseCase_CloneTemplate tests cloning templates.
func TestTemplateUseCase_CloneTemplate(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	original := &template.Template{
		ID:            "test-1",
		Name:          "Original",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}
	repo.Save(ctx, original)

	// Clone
	cloned, err := uc.CloneTemplate(ctx, "test-1")
	if err != nil {
		t.Fatalf("CloneTemplate() failed: %v", err)
	}

	if cloned.ID == original.ID {
		t.Error("CloneTemplate() should generate new ID")
	}
	if cloned.Name != "Original (Copy)" {
		t.Errorf("CloneTemplate() Name = %s, want 'Original (Copy)'", cloned.Name)
	}
}

// TestTemplateUseCase_ExportTemplate tests exporting templates.
func TestTemplateUseCase_ExportTemplate(t *testing.T) {
	ctx := context.Background()
	repo := newMockTemplateRepository()
	uc := NewTemplateUseCase(repo, "")

	tmpl := &template.Template{
		ID:            "test-1",
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
	repo.Save(ctx, tmpl)

	// Export to file
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "exported.json")

	err := uc.ExportTemplate(ctx, "test-1", exportPath)
	if err != nil {
		t.Fatalf("ExportTemplate() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("ExportTemplate() did not create file")
	}

	// Verify file can be read back
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	restored, err := template.FromJSON(data)
	if err != nil {
		t.Fatalf("Failed to parse exported file: %v", err)
	}

	if restored.ID != "test-1" {
		t.Errorf("Exported template ID = %s, want 'test-1'", restored.ID)
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

// Need to define this constant for the test
const parameterTypeEnum template.ParameterType = template.ParameterTypeEnum
