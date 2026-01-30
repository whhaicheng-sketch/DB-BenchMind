// Package repository provides history record repository interfaces and implementations.
package repository

import (
	"context"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

// HistoryRepository defines the interface for history record persistence.
type HistoryRepository interface {
	// Save saves a history record.
	Save(ctx context.Context, record *history.Record) error

	// GetByID retrieves a history record by ID.
	GetByID(ctx context.Context, id string) (*history.Record, error)

	// GetAll retrieves all history records.
	GetAll(ctx context.Context) ([]*history.Record, error)

	// Delete deletes a history record by ID.
	Delete(ctx context.Context, id string) error

	// List retrieves history records with pagination and filtering options.
	List(ctx context.Context, opts *ListOptions) ([]*history.Record, error)
}

// ListOptions defines options for listing history records.
type ListOptions struct {
	// Limit limits the number of records returned.
	Limit int

	// Offset skips the first N records.
	Offset int

	// OrderBy specifies the sort order (e.g., "start_time DESC", "tps ASC").
	OrderBy string

	// ConnectionName filters by connection name.
	ConnectionName string

	// TemplateName filters by template name.
	TemplateName string

	// DatabaseType filters by database type.
	DatabaseType string

	// StartTimeAfter filters records with start time after this value.
	StartTimeAfter *time.Time

	// StartTimeBefore filters records with start time before this value.
	StartTimeBefore *time.Time
}
