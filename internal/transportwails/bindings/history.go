// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

// HistoryBinding provides Wails bindings for history record management.
type HistoryBinding struct {
	uc  *usecase.HistoryUseCase
	ctx context.Context
}

// NewHistoryBinding creates a new HistoryBinding.
func NewHistoryBinding(uc *usecase.HistoryUseCase) *HistoryBinding {
	return &HistoryBinding{uc: uc}
}

// SetContext sets the Wails context.
func (b *HistoryBinding) SetContext(ctx context.Context) {
	b.ctx = ctx
}

// =============================================================================
// DTO Types
// =============================================================================

// HistoryRecordDTO represents a history record for frontend.
type HistoryRecordDTO struct {
	ID             string  `json:"id"`
	CreatedAt      string  `json:"created_at"`
	ConnectionName string  `json:"connection_name"`
	TemplateName   string  `json:"template_name"`
	DatabaseType   string  `json:"database_type"`
	Threads        int     `json:"threads"`
	StartTime      string  `json:"start_time"`
	Duration       string  `json:"duration"`
	DurationMs     int64   `json:"duration_ms"`
	TPS            float64 `json:"tps"`
	TPM            float64 `json:"tpm"`
	LatencyAvg     float64 `json:"latency_avg_ms"`
	LatencyMin     float64 `json:"latency_min_ms"`
	LatencyMax     float64 `json:"latency_max_ms"`
	LatencyP95     float64 `json:"latency_p95_ms"`
	LatencyP99     float64 `json:"latency_p99_ms"`
	TotalQueries   int64   `json:"total_queries"`
	ErrorCount     int64   `json:"error_count"`
}

// HistoryRecordDetailDTO represents detailed history record for frontend.
type HistoryRecordDetailDTO struct {
	HistoryRecordDTO
	// SQL Statistics
	ReadQueries  int64 `json:"read_queries"`
	WriteQueries int64 `json:"write_queries"`
	OtherQueries int64 `json:"other_queries"`

	// Errors and Reconnects
	IgnoredErrors int64 `json:"ignored_errors"`
	Reconnects    int64 `json:"reconnects"`

	// Oracle DML Statistics
	SelectStatements    int64 `json:"select_statements"`
	InsertStatements    int64 `json:"insert_statements"`
	UpdateStatements    int64 `json:"update_statements"`
	DeleteStatements    int64 `json:"delete_statements"`
	CommitStatements    int64 `json:"commit_statements"`
	RollbackStatements  int64 `json:"rollback_statements"`

	// General Statistics
	TotalTime   float64 `json:"total_time"`
	TotalEvents int64   `json:"total_events"`

	// Threads Fairness
	EventsAvg    float64 `json:"events_avg"`
	EventsStddev float64 `json:"events_stddev"`
	ExecTimeAvg  float64 `json:"exec_time_avg"`
	ExecTimeStddev float64 `json:"exec_time_stddev"`
}

// HistoryListResult represents the result of GetRecords.
type HistoryListResult struct {
	Records []HistoryRecordDTO `json:"records"`
	Total   int                `json:"total"`
	Error   string             `json:"error,omitempty"`
}

// HistoryDetailResult represents the result of GetRecordByID.
type HistoryDetailResult struct {
	Record *HistoryRecordDetailDTO `json:"record"`
	Error  string                  `json:"error,omitempty"`
}

// HistoryDeleteResult represents the result of DeleteRecord.
type HistoryDeleteResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// HistoryExportResult represents the result of export operations.
type HistoryExportResult struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// =============================================================================
// Binding Methods
// =============================================================================

// GetRecords retrieves all history records.
func (b *HistoryBinding) GetRecords() HistoryListResult {
	ctx := context.Background()

	records, err := b.uc.GetAllRecords(ctx)
	if err != nil {
		slog.Error("GetRecords failed", "error", err)
		return HistoryListResult{
			Error: err.Error(),
		}
	}

	dtos := make([]HistoryRecordDTO, 0, len(records))
	for _, record := range records {
		dtos = append(dtos, *b.toDTO(record))
	}

	return HistoryListResult{
		Records: dtos,
		Total:   len(dtos),
	}
}

// GetRecordByID retrieves a single history record by ID.
func (b *HistoryBinding) GetRecordByID(id string) HistoryDetailResult {
	ctx := context.Background()

	record, err := b.uc.GetRecordByID(ctx, id)
	if err != nil {
		slog.Error("GetRecordByID failed", "id", id, "error", err)
		return HistoryDetailResult{
			Error: err.Error(),
		}
	}

	return HistoryDetailResult{
		Record: b.toDetailDTO(record),
	}
}

// DeleteRecord deletes a history record by ID.
func (b *HistoryBinding) DeleteRecord(id string) HistoryDeleteResult {
	ctx := context.Background()

	err := b.uc.DeleteRecord(ctx, id)
	if err != nil {
		slog.Error("DeleteRecord failed", "id", id, "error", err)
		return HistoryDeleteResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	slog.Info("History record deleted", "id", id)
	return HistoryDeleteResult{
		Success: true,
	}
}

// DeleteAllRecords deletes all history records.
func (b *HistoryBinding) DeleteAllRecords() HistoryDeleteResult {
	ctx := context.Background()

	records, err := b.uc.GetAllRecords(ctx)
	if err != nil {
		slog.Error("DeleteAllRecords: GetRecords failed", "error", err)
		return HistoryDeleteResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	for _, record := range records {
		if err := b.uc.DeleteRecord(ctx, record.ID); err != nil {
			slog.Error("DeleteAllRecords: DeleteRecord failed", "id", record.ID, "error", err)
			return HistoryDeleteResult{
				Success: false,
				Error:   err.Error(),
			}
		}
	}

	slog.Info("All history records deleted", "count", len(records))
	return HistoryDeleteResult{
		Success: true,
	}
}

// ListRecords retrieves history records with options.
func (b *HistoryBinding) ListRecords(limit int, offset int, databaseType string) HistoryListResult {
	ctx := context.Background()

	opts := &repository.ListOptions{
		Limit:       limit,
		Offset:      offset,
		DatabaseType: databaseType,
		OrderBy:     "start_time DESC",
	}

	if limit <= 0 {
		opts.Limit = 50
	}

	records, err := b.uc.ListRecords(ctx, opts)
	if err != nil {
		slog.Error("ListRecords failed", "error", err)
		return HistoryListResult{
			Error: err.Error(),
		}
	}

	// Get total count
	allRecords, _ := b.uc.GetAllRecords(ctx)
	total := len(allRecords)

	dtos := make([]HistoryRecordDTO, 0, len(records))
	for _, record := range records {
		dtos = append(dtos, *b.toDTO(record))
	}

	return HistoryListResult{
		Records: dtos,
		Total:   total,
	}
}

// ExportRecord exports a single record to text format.
func (b *HistoryBinding) ExportRecord(id string, format string) HistoryExportResult {
	ctx := context.Background()

	record, err := b.uc.GetRecordByID(ctx, id)
	if err != nil {
		slog.Error("ExportRecord: GetRecordByID failed", "id", id, "error", err)
		return HistoryExportResult{
			Error: err.Error(),
		}
	}

	var content string
	switch format {
	case "markdown", "md":
		content = b.formatMarkdown(record)
	case "txt":
		content = b.formatTXT(record)
	default:
		content = b.formatTXT(record)
	}

	return HistoryExportResult{
		Content: content,
	}
}

// ExportAllRecords exports all records to the specified format.
func (b *HistoryBinding) ExportAllRecords(format string) HistoryExportResult {
	ctx := context.Background()

	records, err := b.uc.GetAllRecords(ctx)
	if err != nil {
		slog.Error("ExportAllRecords: GetRecords failed", "error", err)
		return HistoryExportResult{
			Error: err.Error(),
		}
	}

	var content string
	switch format {
	case "markdown", "md":
		content = b.formatAllMarkdown(records)
	case "txt":
		content = b.formatAllTXT(records)
	default:
		content = b.formatAllTXT(records)
	}

	return HistoryExportResult{
		Content: content,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// toDTO converts a Record to HistoryRecordDTO.
func (b *HistoryBinding) toDTO(record *history.Record) *HistoryRecordDTO {
	if record == nil {
		return nil
	}

	return &HistoryRecordDTO{
		ID:             record.ID,
		CreatedAt:      record.CreatedAt.Format(time.RFC3339),
		ConnectionName: record.ConnectionName,
		TemplateName:   record.TemplateName,
		DatabaseType:   record.DatabaseType,
		Threads:        record.Threads,
		StartTime:      record.StartTime.Format(time.RFC3339),
		Duration:       record.Duration.String(),
		DurationMs:     record.Duration.Milliseconds(),
		TPS:            record.TPSCalculated,
		TPM:            record.TPM,
		LatencyAvg:     record.LatencyAvg,
		LatencyMin:     record.LatencyMin,
		LatencyMax:     record.LatencyMax,
		LatencyP95:     record.LatencyP95,
		LatencyP99:     record.LatencyP99,
		TotalQueries:   record.TotalQueries,
		ErrorCount:     record.ErrorCount,
	}
}

// toDetailDTO converts a Record to HistoryRecordDetailDTO.
func (b *HistoryBinding) toDetailDTO(record *history.Record) *HistoryRecordDetailDTO {
	if record == nil {
		return nil
	}

	dto := &HistoryRecordDetailDTO{
		HistoryRecordDTO: HistoryRecordDTO{
			ID:             record.ID,
			CreatedAt:      record.CreatedAt.Format(time.RFC3339),
			ConnectionName: record.ConnectionName,
			TemplateName:   record.TemplateName,
			DatabaseType:   record.DatabaseType,
			Threads:        record.Threads,
			StartTime:      record.StartTime.Format(time.RFC3339),
			Duration:       record.Duration.String(),
			DurationMs:     record.Duration.Milliseconds(),
			TPS:            record.TPSCalculated,
			TPM:            record.TPM,
			LatencyAvg:     record.LatencyAvg,
			LatencyMin:     record.LatencyMin,
			LatencyMax:     record.LatencyMax,
			LatencyP95:     record.LatencyP95,
			LatencyP99:     record.LatencyP99,
			TotalQueries:   record.TotalQueries,
			ErrorCount:     record.ErrorCount,
		},
		ReadQueries:         record.ReadQueries,
		WriteQueries:        record.WriteQueries,
		OtherQueries:        record.OtherQueries,
		IgnoredErrors:       record.IgnoredErrors,
		Reconnects:          record.Reconnects,
		SelectStatements:    record.SelectStatements,
		InsertStatements:    record.InsertStatements,
		UpdateStatements:    record.UpdateStatements,
		DeleteStatements:    record.DeleteStatements,
		CommitStatements:    record.CommitStatements,
		RollbackStatements:  record.RollbackStatements,
		TotalTime:           record.TotalTime,
		TotalEvents:         record.TotalEvents,
		EventsAvg:           record.EventsAvg,
		EventsStddev:        record.EventsStddev,
		ExecTimeAvg:         record.ExecTimeAvg,
		ExecTimeStddev:      record.ExecTimeStddev,
	}

	return dto
}

// formatMarkdown formats a record as Markdown.
func (b *HistoryBinding) formatMarkdown(record *history.Record) string {
	return fmt.Sprintf(`# Benchmark Result - %s

## Configuration
- **Connection**: %s
- **Template**: %s
- **Database**: %s
- **Threads**: %d
- **Duration**: %s

## Performance Metrics
- **TPS**: %.2f
- **TPM**: %.2f

## Latency (ms)
- **Average**: %.2f
- **Min**: %.2f
- **Max**: %.2f
- **P95**: %.2f
- **P99**: %.2f

## Query Statistics
- **Total Queries**: %d
- **Read Queries**: %d
- **Write Queries**: %d
- **Error Count**: %d

---
Generated at: %s
`,
		record.ID,
		record.ConnectionName,
		record.TemplateName,
		record.DatabaseType,
		record.Threads,
		record.Duration,
		record.TPSCalculated,
		record.TPM,
		record.LatencyAvg,
		record.LatencyMin,
		record.LatencyMax,
		record.LatencyP95,
		record.LatencyP99,
		record.TotalQueries,
		record.ReadQueries,
		record.WriteQueries,
		record.ErrorCount,
		time.Now().Format(time.RFC3339),
	)
}

// formatTXT formats a record as plain text.
func (b *HistoryBinding) formatTXT(record *history.Record) string {
	return fmt.Sprintf(`Benchmark Result - %s
=====================================

Configuration:
  Connection: %s
  Template: %s
  Database: %s
  Threads: %d
  Duration: %s

Performance Metrics:
  TPS: %.2f
  TPM: %.2f

Latency (ms):
  Average: %.2f
  Min: %.2f
  Max: %.2f
  P95: %.2f
  P99: %.2f

Query Statistics:
  Total Queries: %d
  Read Queries: %d
  Write Queries: %d
  Error Count: %d

---
Generated at: %s
`,
		record.ID,
		record.ConnectionName,
		record.TemplateName,
		record.DatabaseType,
		record.Threads,
		record.Duration,
		record.TPSCalculated,
		record.TPM,
		record.LatencyAvg,
		record.LatencyMin,
		record.LatencyMax,
		record.LatencyP95,
		record.LatencyP99,
		record.TotalQueries,
		record.ReadQueries,
		record.WriteQueries,
		record.ErrorCount,
		time.Now().Format(time.RFC3339),
	)
}

// formatAllMarkdown formats all records as Markdown.
func (b *HistoryBinding) formatAllMarkdown(records []*history.Record) string {
	var result string
	result = "# Benchmark History Records\n\n"
	result += fmt.Sprintf("Total Records: %d\n\n", len(records))
	result += "---\n\n"

	for _, record := range records {
		result += b.formatMarkdown(record)
		result += "\n\n"
	}

	return result
}

// formatAllTXT formats all records as plain text.
func (b *HistoryBinding) formatAllTXT(records []*history.Record) string {
	var result string
	result = "Benchmark History Records\n"
	result += "=========================\n\n"
	result += fmt.Sprintf("Total Records: %d\n\n", len(records))

	for _, record := range records {
		result += b.formatTXT(record)
		result += "\n" + strings.Repeat("-", 40) + "\n\n"
	}

	return result
}
