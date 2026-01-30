// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

var (
	// ErrRunNotFound is returned when a run is not found.
	ErrRunNotFound = errors.New("run not found")
)

// SQLiteRunRepository implements the RunRepository interface using SQLite.
// Implements: REQ-STORAGE-001, REQ-STORAGE-004, REQ-STORAGE-005
type SQLiteRunRepository struct {
	db *sql.DB
}

// NewSQLiteRunRepository creates a new SQLite run repository.
func NewSQLiteRunRepository(db *sql.DB) *SQLiteRunRepository {
	return &SQLiteRunRepository{db: db}
}

// Save saves a run to the database.
// If the run already exists (by ID), it will be updated.
func (r *SQLiteRunRepository) Save(ctx context.Context, run *execution.Run) error {
	// Serialize result to JSON
	var resultSummaryJSON, resultDetailJSON []byte
	var err error

	if run.Result != nil {
		resultSummaryJSON, err = json.Marshal(run.Result)
		if err != nil {
			return fmt.Errorf("marshal result summary: %w", err)
		}
		resultDetailJSON, err = json.Marshal(run.Result)
		if err != nil {
			return fmt.Errorf("marshal result detail: %w", err)
		}
	}

	// Prepare duration
	var durationSeconds *float64
	if run.Duration != nil {
		d := run.Duration.Seconds()
		durationSeconds = &d
	}

	// Prepare timestamps
	var startedAt, completedAt *string
	if run.StartedAt != nil {
		s := run.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}
	if run.CompletedAt != nil {
		c := run.CompletedAt.Format(time.RFC3339)
		completedAt = &c
	}

	query := `
		INSERT INTO runs (
			id, task_id, state, created_at, started_at, completed_at,
			duration_seconds, result_summary_json, result_detail_json,
			error_message, config_snapshot_path
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			state = excluded.state,
			started_at = excluded.started_at,
			completed_at = excluded.completed_at,
			duration_seconds = excluded.duration_seconds,
			result_summary_json = excluded.result_summary_json,
			result_detail_json = excluded.result_detail_json,
			error_message = excluded.error_message
	`

	_, err = r.db.ExecContext(ctx, query,
		run.ID,
		run.TaskID,
		string(run.State),
		run.CreatedAt.Format(time.RFC3339),
		startedAt,
		completedAt,
		durationSeconds,
		string(resultSummaryJSON),
		string(resultDetailJSON),
		run.ErrorMessage,
		run.WorkDir,
	)
	if err != nil {
		return fmt.Errorf("save run: %w", err)
	}

	return nil
}

// FindByID finds a run by its ID.
func (r *SQLiteRunRepository) FindByID(ctx context.Context, id string) (*execution.Run, error) {
	query := `
		SELECT id, task_id, state, created_at, started_at, completed_at,
		       duration_seconds, result_summary_json, error_message, config_snapshot_path
		FROM runs
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var run execution.Run
	var stateStr, createdAtStr string
	var startedAtStr, completedAtStr *string
	var durationSeconds *float64
	var resultSummaryJSON *string
	var errMsg *string

	err := row.Scan(
		&run.ID,
		&run.TaskID,
		&stateStr,
		&createdAtStr,
		&startedAtStr,
		&completedAtStr,
		&durationSeconds,
		&resultSummaryJSON,
		&errMsg,
		&run.WorkDir,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRunNotFound
		}
		return nil, fmt.Errorf("scan run: %w", err)
	}

	// Parse state
	run.State = execution.RunState(stateStr)

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	run.CreatedAt = createdAt

	if startedAtStr != nil {
		t, err := time.Parse(time.RFC3339, *startedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse started_at: %w", err)
		}
		run.StartedAt = &t
	}

	if completedAtStr != nil {
		t, err := time.Parse(time.RFC3339, *completedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
		run.CompletedAt = &t
	}

	// Parse duration
	if durationSeconds != nil {
		d := time.Duration(*durationSeconds * float64(time.Second))
		run.Duration = &d
	}

	// Parse result
	if resultSummaryJSON != nil && *resultSummaryJSON != "" {
		var result execution.BenchmarkResult
		if err := json.Unmarshal([]byte(*resultSummaryJSON), &result); err == nil {
			run.Result = &result
		}
	}

	// Parse error message
	if errMsg != nil {
		run.ErrorMessage = *errMsg
	}

	return &run, nil
}

// FindAll finds runs with optional filtering and pagination.
func (r *SQLiteRunRepository) FindAll(ctx context.Context, opts usecase.FindOptions) ([]*execution.Run, error) {
	query := `
		SELECT id, task_id, state, created_at, started_at, completed_at,
		       duration_seconds, result_summary_json, error_message, config_snapshot_path
		FROM runs
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply filters
	if opts.StateFilter != nil {
		query += " AND state = ?"
		args = append(args, string(*opts.StateFilter))
	}
	if opts.TaskID != "" {
		query += " AND task_id = ?"
		args = append(args, opts.TaskID)
	}

	// Apply sorting
	sortBy := "created_at"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	sortOrder := "DESC"
	if opts.SortOrder == "ASC" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Apply pagination
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
		if opts.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, opts.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	var runs []*execution.Run
	for rows.Next() {
		run, err := r.scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runs: %w", err)
	}

	return runs, nil
}

// UpdateState updates the state of a run.
func (r *SQLiteRunRepository) UpdateState(ctx context.Context, id string, state execution.RunState) error {
	// Get current run to validate transition
	run, err := r.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRunNotFound) {
			return ErrRunNotFound
		}
		return fmt.Errorf("get run: %w", err)
	}

	// Validate state transition
	if err := run.SetState(state); err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	// Update in database
	query := `UPDATE runs SET state = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, string(state), id)
	if err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrRunNotFound
	}

	return nil
}

// SaveMetricSample saves a metric sample for a run.
func (r *SQLiteRunRepository) SaveMetricSample(ctx context.Context, runID string, sample execution.MetricSample) error {
	query := `
		INSERT INTO metric_samples (
			run_id, timestamp, phase, tps, qps, latency_avg, latency_p95, latency_p99, error_rate, raw_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		runID,
		sample.Timestamp.Format(time.RFC3339),
		sample.Phase,
		sample.TPS,
		sample.QPS,
		sample.LatencyAvg,
		sample.LatencyP95,
		sample.LatencyP99,
		sample.ErrorRate,
		sample.RawLine,
	)
	if err != nil {
		return fmt.Errorf("save metric sample: %w", err)
	}

	return nil
}

// GetMetricSamples retrieves all metric samples for a run.
func (r *SQLiteRunRepository) GetMetricSamples(ctx context.Context, runID string) ([]execution.MetricSample, error) {
	query := `
		SELECT timestamp, phase, tps, qps, latency_avg, latency_p95, latency_p99, error_rate, raw_line
		FROM metric_samples
		WHERE run_id = ?
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, runID)
	if err != nil {
		return nil, fmt.Errorf("query metric samples: %w", err)
	}
	defer rows.Close()

	var samples []execution.MetricSample
	for rows.Next() {
		var sample execution.MetricSample
		var timestampStr string

		err := rows.Scan(
			&timestampStr,
			&sample.Phase,
			&sample.TPS,
			&sample.QPS,
			&sample.LatencyAvg,
			&sample.LatencyP95,
			&sample.LatencyP99,
			&sample.ErrorRate,
			&sample.RawLine,
		)
		if err != nil {
			return nil, fmt.Errorf("scan metric sample: %w", err)
		}

		// Parse timestamp
		sample.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("parse timestamp: %w", err)
		}

		samples = append(samples, sample)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate metric samples: %w", err)
	}

	return samples, nil
}

// SaveLogEntry saves a log entry for a run.
func (r *SQLiteRunRepository) SaveLogEntry(ctx context.Context, runID string, entry usecase.LogEntry) error {
	query := `
		INSERT INTO run_logs (run_id, timestamp, stream, content)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		runID,
		entry.Timestamp,
		entry.Stream,
		entry.Content,
	)
	if err != nil {
		return fmt.Errorf("save log entry: %w", err)
	}

	return nil
}

// Delete deletes a run by its ID.
func (r *SQLiteRunRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM runs WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete run: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrRunNotFound
	}

	return nil
}

// scanRun scans a run from a database row.
func (r *SQLiteRunRepository) scanRun(rows *sql.Rows) (*execution.Run, error) {
	var run execution.Run
	var stateStr, createdAtStr string
	var startedAtStr, completedAtStr *string
	var durationSeconds *float64
	var resultSummaryJSON *string
	var errMsg *string

	err := rows.Scan(
		&run.ID,
		&run.TaskID,
		&stateStr,
		&createdAtStr,
		&startedAtStr,
		&completedAtStr,
		&durationSeconds,
		&resultSummaryJSON,
		&errMsg,
		&run.WorkDir,
	)
	if err != nil {
		return nil, fmt.Errorf("scan run: %w", err)
	}

	// Parse state
	run.State = execution.RunState(stateStr)

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	run.CreatedAt = createdAt

	if startedAtStr != nil {
		t, err := time.Parse(time.RFC3339, *startedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse started_at: %w", err)
		}
		run.StartedAt = &t
	}

	if completedAtStr != nil {
		t, err := time.Parse(time.RFC3339, *completedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
		run.CompletedAt = &t
	}

	// Parse duration
	if durationSeconds != nil {
		d := time.Duration(*durationSeconds * float64(time.Second))
		run.Duration = &d
	}

	// Parse result
	if resultSummaryJSON != nil && *resultSummaryJSON != "" {
		var result execution.BenchmarkResult
		if err := json.Unmarshal([]byte(*resultSummaryJSON), &result); err == nil {
			run.Result = &result
		}
	}

	// Parse error message
	if errMsg != nil {
		run.ErrorMessage = *errMsg
	}

	return &run, nil
}
