// Package repository provides SQLite implementation of repository interfaces.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
)

// SQLiteSuiteRepository implements SuiteRepository interface using SQLite.
type SQLiteSuiteRepository struct {
	db *sql.DB
}

// NewSQLiteSuiteRepository creates a new SQLite suite repository.
func NewSQLiteSuiteRepository(db *sql.DB) usecase.SuiteRepository {
	return &SQLiteSuiteRepository{db: db}
}

// Save saves a suite to the database.
func (r *SQLiteSuiteRepository) Save(ctx context.Context, suite domainautobench.Suite) error {
	now := time.Now().Format(time.RFC3339)

	var startedAt, endedAt interface{}
	if suite.StartedAt != nil {
		startedAt = suite.StartedAt.Format(time.RFC3339)
	}
	if suite.EndedAt != nil {
		endedAt = suite.EndedAt.Format(time.RFC3339)
	}

	cleanupEnabled := 0
	if suite.ExecutionPolicy.CleanupEnabled {
		cleanupEnabled = 1
	}

	query := `
		INSERT INTO suites (
			id, name, execution_mode, failure_policy, cleanup_enabled,
			suite_manifest_json_path, status, started_at, ended_at,
			total_items, completed_items, success_items, failed_items, skipped_items,
			suite_report_json_path, suite_report_html_path, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			execution_mode = excluded.execution_mode,
			failure_policy = excluded.failure_policy,
			cleanup_enabled = excluded.cleanup_enabled,
			suite_manifest_json_path = excluded.suite_manifest_json_path,
			status = excluded.status,
			started_at = excluded.started_at,
			ended_at = excluded.ended_at,
			total_items = excluded.total_items,
			completed_items = excluded.completed_items,
			success_items = excluded.success_items,
			failed_items = excluded.failed_items,
			skipped_items = excluded.skipped_items,
			suite_report_json_path = excluded.suite_report_json_path,
			suite_report_html_path = excluded.suite_report_html_path,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		suite.ID,
		suite.Name,
		string(suite.ExecutionPolicy.Mode),
		string(suite.ExecutionPolicy.FailurePolicy),
		cleanupEnabled,
		suite.SuiteManifestJSONPath,
		string(suite.Status),
		startedAt,
		endedAt,
		len(suite.Items),
		countItemsByStatus(suite.Items, domainautobench.SuiteItemStatusSuccess) +
			countItemsByStatus(suite.Items, domainautobench.SuiteItemStatusFailed) +
			countItemsByStatus(suite.Items, domainautobench.SuiteItemStatusSkipped),
		countItemsByStatus(suite.Items, domainautobench.SuiteItemStatusSuccess),
		countItemsByStatus(suite.Items, domainautobench.SuiteItemStatusFailed),
		countItemsByStatus(suite.Items, domainautobench.SuiteItemStatusSkipped),
		suite.ReportJSONPath,
		suite.ReportPath,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("save suite: %w", err)
	}

	return nil
}

// FindByID finds a suite by its ID.
func (r *SQLiteSuiteRepository) FindByID(ctx context.Context, id string) (domainautobench.Suite, error) {
	query := `
		SELECT id, name, execution_mode, failure_policy, cleanup_enabled,
			suite_manifest_json_path, status, started_at, ended_at,
			total_items, completed_items, success_items, failed_items, skipped_items,
			suite_report_json_path, suite_report_html_path, created_at, updated_at
		FROM suites WHERE id = ?
	`

	suite := domainautobench.Suite{}
	var name, executionMode, failurePolicy, suiteManifestJSONPath *string
	var suiteReportJSONPath, suiteReportHTMLPath *string
	var startedAt, endedAt, createdAt, updatedAt *string
	var cleanupEnabled *int
	var status string
	var totalItems, completedItems, successItems, failedItems, skippedItems int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&suite.ID, &name, &executionMode, &failurePolicy, &cleanupEnabled,
		&suiteManifestJSONPath, &status, &startedAt, &endedAt,
		&totalItems, &completedItems, &successItems, &failedItems, &skippedItems,
		&suiteReportJSONPath, &suiteReportHTMLPath, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return domainautobench.Suite{}, &SuiteNotFoundError{ID: id}
	}
	if err != nil {
		return domainautobench.Suite{}, fmt.Errorf("query suite: %w", err)
	}

	suite.Status = domainautobench.SuiteStatus(status)
	if name != nil {
		suite.Name = *name
	}
	if executionMode != nil {
		suite.ExecutionPolicy.Mode = domainautobench.ExecutionMode(*executionMode)
	}
	if failurePolicy != nil {
		suite.ExecutionPolicy.FailurePolicy = domainautobench.FailurePolicy(*failurePolicy)
	}
	if cleanupEnabled != nil {
		suite.ExecutionPolicy.CleanupEnabled = *cleanupEnabled != 0
	}
	if suiteManifestJSONPath != nil {
		suite.SuiteManifestJSONPath = *suiteManifestJSONPath
	}
	if startedAt != nil {
		t, err := time.Parse(time.RFC3339, *startedAt)
		if err == nil {
			suite.StartedAt = &t
		}
	}
	if endedAt != nil {
		t, err := time.Parse(time.RFC3339, *endedAt)
		if err == nil {
			suite.EndedAt = &t
		}
	}
	if suiteReportJSONPath != nil {
		suite.ReportJSONPath = *suiteReportJSONPath
	}
	if suiteReportHTMLPath != nil {
		suite.ReportPath = *suiteReportHTMLPath
	}

	return suite, nil
}

// FindAll finds all suites in the database.
func (r *SQLiteSuiteRepository) FindAll(ctx context.Context) ([]domainautobench.Suite, error) {
	query := `
		SELECT id, name, execution_mode, failure_policy, cleanup_enabled,
			suite_manifest_json_path, status, started_at, ended_at,
			total_items, completed_items, success_items, failed_items, skipped_items,
			suite_report_json_path, suite_report_html_path, created_at, updated_at
		FROM suites ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query suites: %w", err)
	}
	defer rows.Close()

	var suites []domainautobench.Suite
	for rows.Next() {
		suite := domainautobench.Suite{}
		var name, executionMode, failurePolicy, suiteManifestJSONPath *string
		var suiteReportJSONPath, suiteReportHTMLPath *string
		var startedAt, endedAt, createdAt, updatedAt *string
		var cleanupEnabled *int
		var status string
		var totalItems, completedItems, successItems, failedItems, skippedItems int

		if err := rows.Scan(
			&suite.ID, &name, &executionMode, &failurePolicy, &cleanupEnabled,
			&suiteManifestJSONPath, &status, &startedAt, &endedAt,
			&totalItems, &completedItems, &successItems, &failedItems, &skippedItems,
			&suiteReportJSONPath, &suiteReportHTMLPath, &createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan suite: %w", err)
		}

		suite.Status = domainautobench.SuiteStatus(status)
		if name != nil {
			suite.Name = *name
		}
		if executionMode != nil {
			suite.ExecutionPolicy.Mode = domainautobench.ExecutionMode(*executionMode)
		}
		if failurePolicy != nil {
			suite.ExecutionPolicy.FailurePolicy = domainautobench.FailurePolicy(*failurePolicy)
		}
		if cleanupEnabled != nil {
			suite.ExecutionPolicy.CleanupEnabled = *cleanupEnabled != 0
		}
		if suiteManifestJSONPath != nil {
			suite.SuiteManifestJSONPath = *suiteManifestJSONPath
		}
		if startedAt != nil {
			t, err := time.Parse(time.RFC3339, *startedAt)
			if err == nil {
				suite.StartedAt = &t
			}
		}
		if endedAt != nil {
			t, err := time.Parse(time.RFC3339, *endedAt)
			if err == nil {
				suite.EndedAt = &t
			}
		}
		if suiteReportJSONPath != nil {
			suite.ReportJSONPath = *suiteReportJSONPath
		}
		if suiteReportHTMLPath != nil {
			suite.ReportPath = *suiteReportHTMLPath
		}

		suites = append(suites, suite)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate suites: %w", err)
	}

	return suites, nil
}

// UpdateStatus updates the status of a suite.
func (r *SQLiteSuiteRepository) UpdateStatus(ctx context.Context, id string, status domainautobench.SuiteStatus, manifestPath string) error {
	now := time.Now().Format(time.RFC3339)

	query := `
		UPDATE suites SET status = ?, suite_manifest_json_path = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, string(status), manifestPath, now, id)
	if err != nil {
		return fmt.Errorf("update suite status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return &SuiteNotFoundError{ID: id}
	}

	return nil
}

// Delete deletes a suite by its ID.
func (r *SQLiteSuiteRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM suites WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete suite: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return &SuiteNotFoundError{ID: id}
	}

	return nil
}

// SuiteNotFoundError is returned when a suite is not found.
type SuiteNotFoundError struct {
	ID string
}

func (e *SuiteNotFoundError) Error() string {
	return fmt.Sprintf("suite not found: %s", e.ID)
}

// countItemsByStatus counts suite items by status.
func countItemsByStatus(items []domainautobench.SuiteItem, status domainautobench.SuiteItemStatus) int {
	count := 0
	for _, item := range items {
		if item.Status == status {
			count++
		}
	}
	return count
}
