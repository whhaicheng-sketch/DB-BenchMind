package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

var (
	ErrTemplateNotFound               = errors.New("template not found")
	ErrBuiltinTemplateCannotBeDeleted = errors.New("builtin templates cannot be deleted")
)

// TemplateRepository implements usecase.TemplateRepository using SQLite.
type TemplateRepository struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) usecase.TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Save(ctx context.Context, tmpl *domaintemplate.Template) error {
	if tmpl == nil {
		return fmt.Errorf("template is required")
	}
	_, err := r.saveTemplate(ctx, r.db, tmpl)
	return err
}

func (r *TemplateRepository) saveTemplate(ctx context.Context, execer interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}, tmpl *domaintemplate.Template) (sql.Result, error) {
	if tmpl == nil {
		return nil, fmt.Errorf("template is required")
	}
	tmpl.Normalize()

	configJSON, err := json.Marshal(tmpl)
	if err != nil {
		return nil, fmt.Errorf("marshal template: %w", err)
	}
	dbTypesJSON, err := json.Marshal(tmpl.DatabaseTypes)
	if err != nil {
		return nil, fmt.Errorf("marshal database types: %w", err)
	}
	parametersJSON := "[]"
	commandJSON := "{}"
	parserJSON := "{}"
	if tmpl.Parameters != nil {
		if b, marshalErr := json.Marshal(tmpl.Parameters); marshalErr == nil {
			parametersJSON = string(b)
		}
	}
	if b, marshalErr := json.Marshal(tmpl.CommandTemplate); marshalErr == nil {
		commandJSON = string(b)
	}
	if b, marshalErr := json.Marshal(tmpl.OutputParser); marshalErr == nil {
		parserJSON = string(b)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	createdAt := tmpl.CreatedAt
	if createdAt == "" {
		createdAt = now
	}
	isBuiltin := tmpl.IsBuiltin

	query := `
		INSERT INTO templates (
			id, name, description, tool, database_types, version,
			parameters_json, command_template_json, output_parser_json,
			config_json, is_builtin, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			tool = excluded.tool,
			database_types = excluded.database_types,
			version = excluded.version,
			parameters_json = excluded.parameters_json,
			command_template_json = excluded.command_template_json,
			output_parser_json = excluded.output_parser_json,
			config_json = excluded.config_json,
			is_builtin = excluded.is_builtin,
			updated_at = excluded.updated_at
	`
	result, err := execer.ExecContext(ctx, query,
		tmpl.ID,
		tmpl.Name,
		tmpl.Description,
		tmpl.Tool,
		string(dbTypesJSON),
		tmpl.Version,
		parametersJSON,
		commandJSON,
		parserJSON,
		string(configJSON),
		isBuiltin,
		createdAt,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("save template: %w", err)
	}
	return result, nil
}

func (r *TemplateRepository) FindByID(ctx context.Context, id string) (*domaintemplate.Template, error) {
	query := `
		SELECT config_json, is_builtin
		FROM templates
		WHERE id = ? AND config_json IS NOT NULL AND config_json != ''
	`
	var configJSON string
	var isBuiltin bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&configJSON, &isBuiltin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("query template: %w", err)
	}
	return r.unmarshalTemplate(configJSON, isBuiltin)
}

func (r *TemplateRepository) FindAll(ctx context.Context) ([]*domaintemplate.Template, error) {
	return r.findByQuery(ctx, `
		SELECT config_json, is_builtin
		FROM templates
		WHERE config_json IS NOT NULL AND config_json != ''
		ORDER BY name
	`)
}

func (r *TemplateRepository) FindBuiltin(ctx context.Context) ([]*domaintemplate.Template, error) {
	return r.findByQuery(ctx, `
		SELECT config_json, is_builtin
		FROM templates
		WHERE is_builtin = 1 AND config_json IS NOT NULL AND config_json != ''
		ORDER BY name
	`)
}

func (r *TemplateRepository) FindCustom(ctx context.Context) ([]*domaintemplate.Template, error) {
	return r.findByQuery(ctx, `
		SELECT config_json, is_builtin
		FROM templates
		WHERE is_builtin = 0 AND config_json IS NOT NULL AND config_json != ''
		ORDER BY name
	`)
}

func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	var isBuiltin bool
	err := r.db.QueryRowContext(ctx, "SELECT is_builtin FROM templates WHERE id = ?", id).Scan(&isBuiltin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("query template builtin flag: %w", err)
	}
	if isBuiltin {
		return ErrBuiltinTemplateCannotBeDeleted
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM templates WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrTemplateNotFound
	}
	return nil
}

func (r *TemplateRepository) LoadBuiltinTemplates(ctx context.Context, templates []*domaintemplate.Template) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin template sync: %w", err)
	}
	defer tx.Rollback()

	keepIDs := make([]string, 0, len(templates))
	for _, tmpl := range templates {
		if tmpl == nil {
			continue
		}
		keepIDs = append(keepIDs, tmpl.ID)
		if _, err := r.saveTemplate(ctx, tx, tmpl); err != nil {
			return err
		}
	}

	if len(keepIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(keepIDs)), ",")
		args := make([]interface{}, len(keepIDs))
		for i, id := range keepIDs {
			args[i] = id
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM templates WHERE is_builtin = 1 AND id NOT IN ("+placeholders+")", args...); err != nil {
			return fmt.Errorf("delete stale builtin templates: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit template sync: %w", err)
	}
	return nil
}

func (r *TemplateRepository) findByQuery(ctx context.Context, query string) ([]*domaintemplate.Template, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	var templates []*domaintemplate.Template
	for rows.Next() {
		var configJSON string
		var isBuiltin bool
		if scanErr := rows.Scan(&configJSON, &isBuiltin); scanErr != nil {
			return nil, fmt.Errorf("scan template: %w", scanErr)
		}
		tmpl, unmarshalErr := r.unmarshalTemplate(configJSON, isBuiltin)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		templates = append(templates, tmpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate templates: %w", err)
	}
	return templates, nil
}

func (r *TemplateRepository) unmarshalTemplate(configJSON string, isBuiltin bool) (*domaintemplate.Template, error) {
	var tmpl domaintemplate.Template
	if err := json.Unmarshal([]byte(configJSON), &tmpl); err != nil {
		return nil, fmt.Errorf("unmarshal template config: %w", err)
	}
	tmpl.IsBuiltin = isBuiltin
	tmpl.Normalize()
	return &tmpl, nil
}
