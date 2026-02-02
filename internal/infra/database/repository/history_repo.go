// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

var (
	// ErrHistoryRecordNotFound is returned when a history record is not found.
	ErrHistoryRecordNotFound = errors.New("history record not found")
)

// SQLiteHistoryRepository implements the HistoryRepository interface using SQLite.
type SQLiteHistoryRepository struct {
	db *sql.DB
}

// NewSQLiteHistoryRepository creates a new SQLite history repository.
func NewSQLiteHistoryRepository(db *sql.DB) *SQLiteHistoryRepository {
	return &SQLiteHistoryRepository{db: db}
}

// Save saves a history record to the database.
// If the record already exists (by ID), it will be updated.
func (r *SQLiteHistoryRepository) Save(ctx context.Context, record *history.Record) error {
	// Serialize the record to JSON for storage
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	// Check if record already exists
	var existingID string
	err = r.db.QueryRowContext(ctx, "SELECT id FROM history_records WHERE id = ?", record.ID).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check existing record: %w", err)
	}

	// Log the operation
	if existingID != "" {
		// Record exists - this should not happen with unique run IDs
		return fmt.Errorf("record with id %s already exists (should not happen)", record.ID)
	}

	query := `
		INSERT INTO history_records (
			id, created_at, connection_name, template_name, database_type,
			threads, start_time, duration_seconds, tps, record_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.CreatedAt.Format(time.RFC3339),
		record.ConnectionName,
		record.TemplateName,
		record.DatabaseType,
		record.Threads,
		record.StartTime.Format(time.RFC3339),
		record.Duration.Seconds(),
		record.TPSCalculated,
		string(recordJSON),
	)
	if err != nil {
		return fmt.Errorf("insert history record: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		return fmt.Errorf("expected 1 row affected, got %d", rowsAffected)
	}

	return nil
}

// GetByID retrieves a history record by ID.
func (r *SQLiteHistoryRepository) GetByID(ctx context.Context, id string) (*history.Record, error) {
	query := `SELECT id, created_at, connection_name, template_name, database_type,
	          threads, start_time, duration_seconds, tps, record_json
	          FROM history_records WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var record history.Record
	var createdAtStr, startTimeStr string
	var durationSeconds, tps float64
	var recordJSON string

	err := row.Scan(
		&record.ID,
		&createdAtStr,
		&record.ConnectionName,
		&record.TemplateName,
		&record.DatabaseType,
		&record.Threads,
		&startTimeStr,
		&durationSeconds,
		&tps,
		&recordJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHistoryRecordNotFound
		}
		return nil, fmt.Errorf("scan history record: %w", err)
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	record.CreatedAt = createdAt

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("parse start_time: %w", err)
	}
	record.StartTime = startTime

	record.Duration = time.Duration(durationSeconds) * time.Second
	record.TPSCalculated = tps

	// Unmarshal the full record JSON to get all fields
	if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
		return nil, fmt.Errorf("unmarshal record JSON: %w", err)
	}

	return &record, nil
}

// GetAll retrieves all history records ordered by start time (newest first).
func (r *SQLiteHistoryRepository) GetAll(ctx context.Context) ([]*history.Record, error) {
	query := `SELECT id, created_at, connection_name, template_name, database_type,
	          threads, start_time, duration_seconds, tps, record_json
	          FROM history_records ORDER BY start_time DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query history records: %w", err)
	}
	defer rows.Close()

	var records []*history.Record
	for rows.Next() {
		var record history.Record
		var createdAtStr, startTimeStr string
		var durationSeconds, tps float64
		var recordJSON string

		err := rows.Scan(
			&record.ID,
			&createdAtStr,
			&record.ConnectionName,
			&record.TemplateName,
			&record.DatabaseType,
			&record.Threads,
			&startTimeStr,
			&durationSeconds,
			&tps,
			&recordJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan history record: %w", err)
		}

		// Parse timestamps
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		record.CreatedAt = createdAt

		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("parse start_time: %w", err)
		}
		record.StartTime = startTime

		record.Duration = time.Duration(durationSeconds) * time.Second

		// Unmarshal the full record JSON to get all fields
		if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
			return nil, fmt.Errorf("unmarshal record JSON: %w", err)
		}

		// ⭐ 关键修复：在Unmarshal之后设置TPS，确保使用数据库列中的值
		record.TPSCalculated = tps

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history records: %w", err)
	}

	return records, nil
}

// Delete deletes a history record by ID.
func (r *SQLiteHistoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM history_records WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete history record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrHistoryRecordNotFound
	}

	return nil
}

// List retrieves history records with pagination and filtering options.
func (r *SQLiteHistoryRepository) List(ctx context.Context, opts *repository.ListOptions) ([]*history.Record, error) {
	if opts == nil {
		opts = &repository.ListOptions{}
	}

	// Build query with filters
	query := `SELECT id, created_at, connection_name, template_name, database_type,
	          threads, start_time, duration_seconds, tps, record_json
	          FROM history_records WHERE 1=1`
	args := []interface{}{}

	// Add filters
	if opts.ConnectionName != "" {
		query += " AND connection_name = ?"
		args = append(args, opts.ConnectionName)
	}
	if opts.TemplateName != "" {
		query += " AND template_name = ?"
		args = append(args, opts.TemplateName)
	}
	if opts.DatabaseType != "" {
		query += " AND database_type = ?"
		args = append(args, opts.DatabaseType)
	}
	if opts.StartTimeAfter != nil {
		query += " AND start_time >= ?"
		args = append(args, opts.StartTimeAfter.Format(time.RFC3339))
	}
	if opts.StartTimeBefore != nil {
		query += " AND start_time <= ?"
		args = append(args, opts.StartTimeBefore.Format(time.RFC3339))
	}

	// Add ordering
	orderClause := "start_time DESC"
	if opts.OrderBy != "" {
		orderClause = opts.OrderBy
	}
	query += " ORDER BY " + orderClause

	// Add pagination
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query history records: %w", err)
	}
	defer rows.Close()

	var records []*history.Record
	for rows.Next() {
		var record history.Record
		var createdAtStr, startTimeStr string
		var durationSeconds, tps float64
		var recordJSON string

		err := rows.Scan(
			&record.ID,
			&createdAtStr,
			&record.ConnectionName,
			&record.TemplateName,
			&record.DatabaseType,
			&record.Threads,
			&startTimeStr,
			&durationSeconds,
			&tps,
			&recordJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan history record: %w", err)
		}

		// Parse timestamps
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		record.CreatedAt = createdAt

		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("parse start_time: %w", err)
		}
		record.StartTime = startTime

		record.Duration = time.Duration(durationSeconds) * time.Second

		// Unmarshal the full record JSON to get all fields
		if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
			return nil, fmt.Errorf("unmarshal record JSON: %w", err)
		}

		// ⭐ 关键修复：在Unmarshal之后设置TPS，确保使用数据库列中的值
		record.TPSCalculated = tps

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history records: %w", err)
	}

	return records, nil
}
