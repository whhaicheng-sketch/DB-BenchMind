// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

var (
	// ErrTemplateNotFound is returned when a template is not found.
	ErrTemplateNotFound = errors.New("template not found")

	// ErrBuiltinTemplateCannotBeDeleted is returned when trying to delete a builtin template.
	ErrBuiltinTemplateCannotBeDeleted = errors.New("builtin templates cannot be deleted")
)

// TemplateRepository implements the TemplateRepository interface using SQLite.
// Implements: REQ-TMPL-001, REQ-TMPL-002
type TemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new SQLite template repository.
func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Save saves a template to the database.
// If the template already exists (by ID), it will be updated.
func (r *TemplateRepository) Save(ctx context.Context, tmpl *template.Template) error {
	configJSON, err := json.Marshal(tmpl)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	query := `
		INSERT INTO templates (id, name, description, tool, database_types, version, config_json, is_builtin)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			tool = excluded.tool,
			database_types = excluded.database_types,
			version = excluded.version,
			config_json = excluded.config_json
	`

	dbTypesJSON, err := json.Marshal(tmpl.DatabaseTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal database types: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tmpl.ID,
		tmpl.Name,
		tmpl.Description,
		tmpl.Tool,
		string(dbTypesJSON),
		tmpl.Version,
		string(configJSON),
		false, // is_builtin - user templates are not builtin
	)
	if err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// FindByID finds a template by its ID.
func (r *TemplateRepository) FindByID(ctx context.Context, id string) (*template.Template, error) {
	query := `
		SELECT id, name, description, tool, database_types, version, config_json
		FROM templates
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var tmpl template.Template
	var dbTypesJSON string
	var configJSON string

	err := row.Scan(
		&tmpl.ID,
		&tmpl.Name,
		&tmpl.Description,
		&tmpl.Tool,
		&dbTypesJSON,
		&tmpl.Version,
		&configJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to scan template: %w", err)
	}

	// Unmarshal database types
	if err := json.Unmarshal([]byte(dbTypesJSON), &tmpl.DatabaseTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database types: %w", err)
	}

	// Unmarshal config (which contains the full template)
	if err := json.Unmarshal([]byte(configJSON), &tmpl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template config: %w", err)
	}

	return &tmpl, nil
}

// FindAll finds all templates in the database.
// Returns templates ordered by tool and name.
func (r *TemplateRepository) FindAll(ctx context.Context) ([]*template.Template, error) {
	query := `
		SELECT id, name, description, tool, database_types, version, config_json
		FROM templates
		ORDER BY tool, name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*template.Template
	for rows.Next() {
		tmpl, err := r.scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, tmpl)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

// FindBuiltin finds all builtin templates.
func (r *TemplateRepository) FindBuiltin(ctx context.Context) ([]*template.Template, error) {
	query := `
		SELECT id, name, description, tool, database_types, version, config_json
		FROM templates
		WHERE is_builtin = 1
		ORDER BY tool, name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query builtin templates: %w", err)
	}
	defer rows.Close()

	var templates []*template.Template
	for rows.Next() {
		tmpl, err := r.scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, tmpl)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating builtin templates: %w", err)
	}

	return templates, nil
}

// FindCustom finds all user-defined (non-builtin) templates.
func (r *TemplateRepository) FindCustom(ctx context.Context) ([]*template.Template, error) {
	query := `
		SELECT id, name, description, tool, database_types, version, config_json
		FROM templates
		WHERE is_builtin = 0
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query custom templates: %w", err)
	}
	defer rows.Close()

	var templates []*template.Template
	for rows.Next() {
		tmpl, err := r.scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, tmpl)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating custom templates: %w", err)
	}

	return templates, nil
}

// Delete deletes a template by its ID.
// Builtin templates cannot be deleted.
func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	// Check if template is builtin
	var isBuiltin bool
	err := r.db.QueryRowContext(ctx, "SELECT is_builtin FROM templates WHERE id = ?", id).Scan(&isBuiltin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("failed to check if template is builtin: %w", err)
	}

	if isBuiltin {
		return ErrBuiltinTemplateCannotBeDeleted
	}

	// Delete the template
	result, err := r.db.ExecContext(ctx, "DELETE FROM templates WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

// LoadBuiltinTemplates loads builtin templates from a directory and saves them to the database.
// If a builtin template already exists, it will be updated.
func (r *TemplateRepository) LoadBuiltinTemplates(ctx context.Context, templates []*template.Template) error {
	for _, tmpl := range templates {
		configJSON, err := json.Marshal(tmpl)
		if err != nil {
			return fmt.Errorf("failed to marshal template %s: %w", tmpl.ID, err)
		}

		dbTypesJSON, err := json.Marshal(tmpl.DatabaseTypes)
		if err != nil {
			return fmt.Errorf("failed to marshal database types for %s: %w", tmpl.ID, err)
		}

		query := `
			INSERT INTO templates (id, name, description, tool, database_types, version, config_json, is_builtin)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				description = excluded.description,
				tool = excluded.tool,
				database_types = excluded.database_types,
				version = excluded.version,
				config_json = excluded.config_json,
				is_builtin = excluded.is_builtin
		`

		_, err = r.db.ExecContext(ctx, query,
			tmpl.ID,
			tmpl.Name,
			tmpl.Description,
			tmpl.Tool,
			string(dbTypesJSON),
			tmpl.Version,
			string(configJSON),
			true, // is_builtin
		)
		if err != nil {
			return fmt.Errorf("failed to save builtin template %s: %w", tmpl.ID, err)
		}
	}

	return nil
}

// scanTemplate scans a template from a database row.
func (r *TemplateRepository) scanTemplate(rows *sql.Rows) (*template.Template, error) {
	var tmpl template.Template
	var dbTypesJSON string
	var configJSON string

	err := rows.Scan(
		&tmpl.ID,
		&tmpl.Name,
		&tmpl.Description,
		&tmpl.Tool,
		&dbTypesJSON,
		&tmpl.Version,
		&configJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan template: %w", err)
	}

	// Unmarshal database types
	if err := json.Unmarshal([]byte(dbTypesJSON), &tmpl.DatabaseTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database types: %w", err)
	}

	// Unmarshal config (which contains the full template)
	if err := json.Unmarshal([]byte(configJSON), &tmpl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template config: %w", err)
	}

	return &tmpl, nil
}
