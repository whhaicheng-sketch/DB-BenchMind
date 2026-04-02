// Package usecase provides report query business logic.
package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// ListReportsOptions contains options for listing reports.
type ListReportsOptions struct {
	Page         int
	PageSize     int
	SuiteID      string
	Status       string
	ConnectionID string
}

// ListSuitesOptions contains options for listing suites.
type ListSuitesOptions struct {
	Page     int
	PageSize int
	Status   string
}

// ReportUsecase provides report query operations.
type ReportUsecase struct {
	db *sql.DB
}

// NewReportUsecase creates a new ReportUsecase.
func NewReportUsecase(db *sql.DB) *ReportUsecase {
	return &ReportUsecase{db: db}
}

// ListReports retrieves reports with pagination and filtering.
func (uc *ReportUsecase) ListReports(ctx context.Context, opts ListReportsOptions) ([]*report.Report, int, error) {
	// Apply defaults
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100
	}

	// Build count query
	countQuery := "SELECT COUNT(*) FROM reports WHERE 1=1"
	var countArgs []interface{}

	if opts.SuiteID != "" {
		countQuery += " AND suite_id = ?"
		countArgs = append(countArgs, opts.SuiteID)
	}
	if opts.Status != "" {
		countQuery += " AND status = ?"
		countArgs = append(countArgs, opts.Status)
	}
	if opts.ConnectionID != "" {
		countQuery += " AND connection_id = ?"
		countArgs = append(countArgs, opts.ConnectionID)
	}

	var total int
	if err := uc.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reports: %w", err)
	}

	// Build data query
	query := `
		SELECT id, suite_id, suite_item_id, source_type, connection_id, connection_name,
			database_type, template_id, template_name, started_at, ended_at, duration_ms,
			status, error_message, tpm, tps, qps, throughput, latency_avg_ms,
			latency_p95_ms, latency_p99_ms, error_count, metrics_json_path,
			monitoring_json_path, raw_json_path, report_html_path, summary_json_path,
			created_at, updated_at, tags
		FROM reports WHERE 1=1
	`
	var args []interface{}

	if opts.SuiteID != "" {
		query += " AND suite_id = ?"
		args = append(args, opts.SuiteID)
	}
	if opts.Status != "" {
		query += " AND status = ?"
		args = append(args, opts.Status)
	}
	if opts.ConnectionID != "" {
		query += " AND connection_id = ?"
		args = append(args, opts.ConnectionID)
	}

	query += " ORDER BY started_at DESC LIMIT ? OFFSET ?"
	args = append(args, opts.PageSize, (opts.Page-1)*opts.PageSize)

	rows, err := uc.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query reports: %w", err)
	}
	defer rows.Close()

	var reports []*report.Report
	for rows.Next() {
		r := &report.Report{}
		var endedAt, startedAt, createdAt, updatedAt *string
		var suiteItemID, connectionName, templateID, templateName, errorMessage *string
		var durationMs *int64
		var tpm, tps, qps, throughput, latencyAvgMs, latencyP95Ms, latencyP99Ms *float64
		var errorCount *int64
		var metricsJSONPath, monitoringJSONPath, rawJSONPath, reportHTMLPath, summaryJSONPath, tags *string
		var sourceType string
		var status string

		if err := rows.Scan(
			&r.ID, &r.SuiteID, &suiteItemID, &sourceType, &r.ConnectionID, &connectionName,
			&r.DatabaseType, &templateID, &templateName, &startedAt, &endedAt, &durationMs,
			&status, &errorMessage, &tpm, &tps, &qps, &throughput, &latencyAvgMs,
			&latencyP95Ms, &latencyP99Ms, &errorCount, &metricsJSONPath,
			&monitoringJSONPath, &rawJSONPath, &reportHTMLPath, &summaryJSONPath,
			&createdAt, &updatedAt, &tags,
		); err != nil {
			return nil, 0, fmt.Errorf("scan report: %w", err)
		}

		r.SourceType = report.SourceType(sourceType)
		r.Status = report.ReportStatus(status)
		if suiteItemID != nil {
			r.SuiteItemID = *suiteItemID
		}
		if connectionName != nil {
			r.ConnectionName = *connectionName
		}
		if templateID != nil {
			r.TemplateID = *templateID
		}
		if templateName != nil {
			r.TemplateName = *templateName
		}
		if startedAt != nil {
			t, err := time.Parse(time.RFC3339, *startedAt)
			if err == nil {
				r.StartedAt = t
			}
		}
		if endedAt != nil {
			t, err := time.Parse(time.RFC3339, *endedAt)
			if err == nil {
				r.EndedAt = &t
			}
		}
		if durationMs != nil {
			r.DurationMs = *durationMs
		}
		if errorMessage != nil {
			r.ErrorMessage = *errorMessage
		}
		if tpm != nil {
			r.TPM = *tpm
		}
		if tps != nil {
			r.TPS = *tps
		}
		if qps != nil {
			r.QPS = *qps
		}
		if throughput != nil {
			r.Throughput = *throughput
		}
		if latencyAvgMs != nil {
			r.LatencyAvgMs = *latencyAvgMs
		}
		if latencyP95Ms != nil {
			r.LatencyP95Ms = *latencyP95Ms
		}
		if latencyP99Ms != nil {
			r.LatencyP99Ms = *latencyP99Ms
		}
		if errorCount != nil {
			r.ErrorCount = *errorCount
		}
		if metricsJSONPath != nil {
			r.MetricsJSONPath = *metricsJSONPath
		}
		if monitoringJSONPath != nil {
			r.MonitoringJSONPath = *monitoringJSONPath
		}
		if rawJSONPath != nil {
			r.RawJSONPath = *rawJSONPath
		}
		if reportHTMLPath != nil {
			r.ReportHTMLPath = *reportHTMLPath
		}
		if summaryJSONPath != nil {
			r.SummaryJSONPath = *summaryJSONPath
		}
		if createdAt != nil {
			t, err := time.Parse(time.RFC3339, *createdAt)
			if err == nil {
				r.CreatedAt = t
			}
		}
		if updatedAt != nil {
			t, err := time.Parse(time.RFC3339, *updatedAt)
			if err == nil {
				r.UpdatedAt = t
			}
		}
		if tags != nil {
			r.Tags = *tags
		}

		reports = append(reports, r)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate reports: %w", err)
	}

	return reports, total, nil
}

// reportSelectColumns is the standard column list used for single-report queries.
const reportSelectColumns = `
	id, suite_id, suite_item_id, source_type, connection_id, connection_name,
	database_type, template_id, template_name, started_at, ended_at, duration_ms,
	status, error_message, tpm, tps, qps, throughput, latency_avg_ms,
	latency_p95_ms, latency_p99_ms, error_count, metrics_json_path,
	monitoring_json_path, raw_json_path, report_html_path, summary_json_path,
	created_at, updated_at, tags`

// scanReportRow scans a single report row into a Report struct.
func scanReportRow(row *sql.Row) (*report.Report, error) {
	r := &report.Report{}
	var endedAt, startedAt, createdAt, updatedAt *string
	var suiteItemID, connectionName, templateID, templateName, errorMessage *string
	var durationMs *int64
	var tpm, tps, qps, throughput, latencyAvgMs, latencyP95Ms, latencyP99Ms *float64
	var errorCount *int64
	var metricsJSONPath, monitoringJSONPath, rawJSONPath, reportHTMLPath, summaryJSONPath, tags *string
	var sourceType string
	var status string

	err := row.Scan(
		&r.ID, &r.SuiteID, &suiteItemID, &sourceType, &r.ConnectionID, &connectionName,
		&r.DatabaseType, &templateID, &templateName, &startedAt, &endedAt, &durationMs,
		&status, &errorMessage, &tpm, &tps, &qps, &throughput, &latencyAvgMs,
		&latencyP95Ms, &latencyP99Ms, &errorCount, &metricsJSONPath,
		&monitoringJSONPath, &rawJSONPath, &reportHTMLPath, &summaryJSONPath,
		&createdAt, &updatedAt, &tags,
	)
	if err != nil {
		return nil, err
	}

	r.SourceType = report.SourceType(sourceType)
	r.Status = report.ReportStatus(status)
	if suiteItemID != nil {
		r.SuiteItemID = *suiteItemID
	}
	if connectionName != nil {
		r.ConnectionName = *connectionName
	}
	if templateID != nil {
		r.TemplateID = *templateID
	}
	if templateName != nil {
		r.TemplateName = *templateName
	}
	if startedAt != nil {
		t, err := time.Parse(time.RFC3339, *startedAt)
		if err == nil {
			r.StartedAt = t
		}
	}
	if endedAt != nil {
		t, err := time.Parse(time.RFC3339, *endedAt)
		if err == nil {
			r.EndedAt = &t
		}
	}
	if durationMs != nil {
		r.DurationMs = *durationMs
	}
	if errorMessage != nil {
		r.ErrorMessage = *errorMessage
	}
	if tpm != nil {
		r.TPM = *tpm
	}
	if tps != nil {
		r.TPS = *tps
	}
	if qps != nil {
		r.QPS = *qps
	}
	if throughput != nil {
		r.Throughput = *throughput
	}
	if latencyAvgMs != nil {
		r.LatencyAvgMs = *latencyAvgMs
	}
	if latencyP95Ms != nil {
		r.LatencyP95Ms = *latencyP95Ms
	}
	if latencyP99Ms != nil {
		r.LatencyP99Ms = *latencyP99Ms
	}
	if errorCount != nil {
		r.ErrorCount = *errorCount
	}
	if metricsJSONPath != nil {
		r.MetricsJSONPath = *metricsJSONPath
	}
	if monitoringJSONPath != nil {
		r.MonitoringJSONPath = *monitoringJSONPath
	}
	if rawJSONPath != nil {
		r.RawJSONPath = *rawJSONPath
	}
	if reportHTMLPath != nil {
		r.ReportHTMLPath = *reportHTMLPath
	}
	if summaryJSONPath != nil {
		r.SummaryJSONPath = *summaryJSONPath
	}
	if createdAt != nil {
		t, err := time.Parse(time.RFC3339, *createdAt)
		if err == nil {
			r.CreatedAt = t
		}
	}
	if updatedAt != nil {
		t, err := time.Parse(time.RFC3339, *updatedAt)
		if err == nil {
			r.UpdatedAt = t
		}
	}
	if tags != nil {
		r.Tags = *tags
	}

	return r, nil
}

// GetReport retrieves a single report by ID.
func (uc *ReportUsecase) GetReport(ctx context.Context, id string) (*report.Report, error) {
	row := uc.db.QueryRowContext(ctx, "SELECT "+reportSelectColumns+" FROM reports WHERE id = ?", id)
	r, err := scanReportRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("report not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}
	return r, nil
}

// GetReportMetrics retrieves the full metrics data for a report.
func (uc *ReportUsecase) GetReportMetrics(ctx context.Context, id string) (*report.MetricsData, error) {
	rpt, err := uc.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}

	if rpt.MetricsJSONPath == "" {
		return nil, fmt.Errorf("report has no metrics path: %s", id)
	}

	data, err := os.ReadFile(rpt.MetricsJSONPath)
	if err != nil {
		return nil, fmt.Errorf("read metrics file: %w", err)
	}

	var metrics report.MetricsData
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, fmt.Errorf("parse metrics json: %w", err)
	}

	return &metrics, nil
}

// ListSuites retrieves suites with pagination and filtering.
func (uc *ReportUsecase) ListSuites(ctx context.Context, opts ListSuitesOptions) ([]*report.Suite, int, error) {
	// Apply defaults
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100
	}

	// Build count query
	countQuery := "SELECT COUNT(*) FROM suites WHERE 1=1"
	var countArgs []interface{}

	if opts.Status != "" {
		countQuery += " AND status = ?"
		countArgs = append(countArgs, opts.Status)
	}

	var total int
	if err := uc.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count suites: %w", err)
	}

	// Build data query
	query := `
		SELECT id, name, execution_mode, failure_policy, cleanup_enabled,
			suite_manifest_json_path, status, started_at, ended_at,
			total_items, completed_items, success_items, failed_items, skipped_items,
			suite_report_json_path, suite_report_html_path, created_at, updated_at
		FROM suites WHERE 1=1
	`
	var args []interface{}

	if opts.Status != "" {
		query += " AND status = ?"
		args = append(args, opts.Status)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, opts.PageSize, (opts.Page-1)*opts.PageSize)

	rows, err := uc.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query suites: %w", err)
	}
	defer rows.Close()

	var suites []*report.Suite
	for rows.Next() {
		s := &report.Suite{}
		var name, executionMode, failurePolicy, suiteManifestJSONPath *string
		var suiteReportJSONPath, suiteReportHTMLPath *string
		var startedAt, endedAt, createdAt, updatedAt *string
		var cleanupEnabled *int
		var status string

		if err := rows.Scan(
			&s.ID, &name, &executionMode, &failurePolicy, &cleanupEnabled,
			&suiteManifestJSONPath, &status, &startedAt, &endedAt,
			&s.TotalItems, &s.CompletedItems, &s.SuccessItems, &s.FailedItems, &s.SkippedItems,
			&suiteReportJSONPath, &suiteReportHTMLPath, &createdAt, &updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan suite: %w", err)
		}

		s.Status = report.SuiteStatus(status)
		if name != nil {
			s.Name = *name
		}
		if executionMode != nil {
			s.ExecutionMode = *executionMode
		}
		if failurePolicy != nil {
			s.FailurePolicy = *failurePolicy
		}
		if cleanupEnabled != nil {
			s.CleanupEnabled = *cleanupEnabled != 0
		}
		if suiteManifestJSONPath != nil {
			s.SuiteManifestJSONPath = *suiteManifestJSONPath
		}
		if startedAt != nil {
			t, err := time.Parse(time.RFC3339, *startedAt)
			if err == nil {
				s.StartedAt = &t
			}
		}
		if endedAt != nil {
			t, err := time.Parse(time.RFC3339, *endedAt)
			if err == nil {
				s.EndedAt = &t
			}
		}
		if suiteReportJSONPath != nil {
			s.SuiteReportJSONPath = *suiteReportJSONPath
		}
		if suiteReportHTMLPath != nil {
			s.SuiteReportHTMLPath = *suiteReportHTMLPath
		}
		if createdAt != nil {
			t, err := time.Parse(time.RFC3339, *createdAt)
			if err == nil {
				s.CreatedAt = t
			}
		}
		if updatedAt != nil {
			t, err := time.Parse(time.RFC3339, *updatedAt)
			if err == nil {
				s.UpdatedAt = t
			}
		}

		suites = append(suites, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate suites: %w", err)
	}

	return suites, total, nil
}

// GetSuite retrieves a single suite by ID.
func (uc *ReportUsecase) GetSuite(ctx context.Context, id string) (*report.Suite, error) {
	query := `
		SELECT id, name, execution_mode, failure_policy, cleanup_enabled,
			suite_manifest_json_path, status, started_at, ended_at,
			total_items, completed_items, success_items, failed_items, skipped_items,
			suite_report_json_path, suite_report_html_path, created_at, updated_at
		FROM suites WHERE id = ?
	`

	s := &report.Suite{}
	var name, executionMode, failurePolicy, suiteManifestJSONPath *string
	var suiteReportJSONPath, suiteReportHTMLPath *string
	var startedAt, endedAt, createdAt, updatedAt *string
	var cleanupEnabled *int
	var status string

	err := uc.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &name, &executionMode, &failurePolicy, &cleanupEnabled,
		&suiteManifestJSONPath, &status, &startedAt, &endedAt,
		&s.TotalItems, &s.CompletedItems, &s.SuccessItems, &s.FailedItems, &s.SkippedItems,
		&suiteReportJSONPath, &suiteReportHTMLPath, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("suite not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get suite: %w", err)
	}

	s.Status = report.SuiteStatus(status)
	if name != nil {
		s.Name = *name
	}
	if executionMode != nil {
		s.ExecutionMode = *executionMode
	}
	if failurePolicy != nil {
		s.FailurePolicy = *failurePolicy
	}
	if cleanupEnabled != nil {
		s.CleanupEnabled = *cleanupEnabled != 0
	}
	if suiteManifestJSONPath != nil {
		s.SuiteManifestJSONPath = *suiteManifestJSONPath
	}
	if startedAt != nil {
		t, err := time.Parse(time.RFC3339, *startedAt)
		if err == nil {
			s.StartedAt = &t
		}
	}
	if endedAt != nil {
		t, err := time.Parse(time.RFC3339, *endedAt)
		if err == nil {
			s.EndedAt = &t
		}
	}
	if suiteReportJSONPath != nil {
		s.SuiteReportJSONPath = *suiteReportJSONPath
	}
	if suiteReportHTMLPath != nil {
		s.SuiteReportHTMLPath = *suiteReportHTMLPath
	}
	if createdAt != nil {
		t, err := time.Parse(time.RFC3339, *createdAt)
		if err == nil {
			s.CreatedAt = t
		}
	}
	if updatedAt != nil {
		t, err := time.Parse(time.RFC3339, *updatedAt)
		if err == nil {
			s.UpdatedAt = t
		}
	}

	return s, nil
}

// ExportData contains all data for JSON export.
type ExportData struct {
	SchemaVersion string                  `json:"schema_version"`
	ReportID      string                  `json:"report_id"`
	Report        *report.Report          `json:"report"`
	Metrics       *report.MetricsData     `json:"metrics,omitempty"`
	Monitoring    *report.MonitoringData  `json:"monitoring,omitempty"`
	Raw           *report.RawData         `json:"raw,omitempty"`
	ExportedAt    string                  `json:"exported_at"`
}

// ExportReportJSON exports all report data as a single JSON structure.
func (uc *ReportUsecase) ExportReportJSON(ctx context.Context, id string) (*ExportData, error) {
	// Get report metadata
	rpt, err := uc.GetReport(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}

	export := &ExportData{
		SchemaVersion: "v1",
		ReportID:      id,
		Report:        rpt,
		ExportedAt:    time.Now().Format(time.RFC3339),
	}

	// Load metrics data
	if rpt.MetricsJSONPath != "" {
		if data, err := os.ReadFile(rpt.MetricsJSONPath); err == nil {
			var metrics report.MetricsData
			if err := json.Unmarshal(data, &metrics); err == nil {
				export.Metrics = &metrics
			}
		}
	}

	// Load monitoring data
	if rpt.MonitoringJSONPath != "" {
		if data, err := os.ReadFile(rpt.MonitoringJSONPath); err == nil {
			var monitoring report.MonitoringData
			if err := json.Unmarshal(data, &monitoring); err == nil {
				export.Monitoring = &monitoring
			}
		}
	}

	// Load raw data
	if rpt.RawJSONPath != "" {
		if data, err := os.ReadFile(rpt.RawJSONPath); err == nil {
			var raw report.RawData
			if err := json.Unmarshal(data, &raw); err == nil {
				export.Raw = &raw
			}
		}
	}

	return export, nil
}

// ExportReportHTML returns the HTML content for a report.
func (uc *ReportUsecase) ExportReportHTML(ctx context.Context, id string) (string, error) {
	rpt, err := uc.GetReport(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get report: %w", err)
	}

	if rpt.ReportHTMLPath == "" {
		return "", fmt.Errorf("report has no HTML path: %s", id)
	}

	data, err := os.ReadFile(rpt.ReportHTMLPath)
	if err != nil {
		return "", fmt.Errorf("read HTML file: %w", err)
	}

	return string(data), nil
}

// DeleteAllReports deletes all reports, cleaning up associated files.
func (uc *ReportUsecase) DeleteAllReports(ctx context.Context) (int, error) {
	rows, err := uc.db.QueryContext(ctx, "SELECT id, metrics_json_path, monitoring_json_path, raw_json_path, report_html_path, summary_json_path FROM reports")
	if err != nil {
		return 0, fmt.Errorf("query reports for delete all: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id, metrics, monitoring, raw, html, summary string
		if err := rows.Scan(&id, &metrics, &monitoring, &raw, &html, &summary); err != nil {
			continue
		}
		for _, p := range []string{metrics, monitoring, raw, html, summary} {
			if p != "" {
				os.Remove(p)
			}
		}
	}

	_, err = uc.db.ExecContext(ctx, "DELETE FROM reports")
	if err != nil {
		return count, fmt.Errorf("delete all reports: %w", err)
	}
	return count, nil
}

// DeleteReport deletes a report by ID, cleaning up associated files.
func (uc *ReportUsecase) DeleteReport(ctx context.Context, id string) error {
	rpt, err := uc.GetReport(ctx, id)
	if err != nil {
		return fmt.Errorf("get report for delete: %w", err)
	}
	// Best-effort file cleanup
	for _, p := range []string{rpt.MetricsJSONPath, rpt.MonitoringJSONPath, rpt.RawJSONPath, rpt.ReportHTMLPath, rpt.SummaryJSONPath} {
		if p != "" {
			os.Remove(p)
		}
	}
	_, err = uc.db.ExecContext(ctx, "DELETE FROM reports WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete report %s: %w", id, err)
	}
	return nil
}

// nilIfEmpty returns nil for empty strings, otherwise returns a pointer to the string.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// GetExportFilePaths returns the file paths for all export files.
func (uc *ReportUsecase) GetExportFilePaths(ctx context.Context, id string) (metrics, monitoring, raw, html string, err error) {
	rpt, err := uc.GetReport(ctx, id)
	if err != nil {
		return "", "", "", "", fmt.Errorf("get report: %w", err)
	}

	return rpt.MetricsJSONPath, rpt.MonitoringJSONPath, rpt.RawJSONPath, rpt.ReportHTMLPath, nil
}

// GetPreviousReportForComparison finds the most recent completed report for the same
// connection and template (excluding the given report ID).
func (uc *ReportUsecase) GetPreviousReportForComparison(ctx context.Context, currentID string) (*report.Report, error) {
	current, err := uc.GetReport(ctx, currentID)
	if err != nil {
		return nil, fmt.Errorf("get current report: %w", err)
	}

	row := uc.db.QueryRowContext(ctx, `
		SELECT `+reportSelectColumns+`
		FROM reports
		WHERE id != ?
		  AND connection_id = ?
		  AND template_id = ?
		  AND status = 'completed'
		ORDER BY started_at DESC
		LIMIT 1`,
		currentID, current.ConnectionID, current.TemplateID,
	)

	prev, err := scanReportRow(row)
	if err == sql.ErrNoRows {
		return nil, nil // No previous report
	}
	if err != nil {
		return nil, fmt.Errorf("query previous report: %w", err)
	}
	return prev, nil
}

// CompareReports generates a comparison between two reports.
func CompareReports(current, previous *report.Report) *report.ComparisonResult {
	if previous == nil {
		return nil
	}

	result := &report.ComparisonResult{
		PreviousReportID: previous.ID,
		Deltas:           make([]report.ComparisonDelta, 0),
	}

	comparisons := []struct {
		metric  string
		current float64
		other   float64
	}{
		{"TPS", current.TPS, previous.TPS},
		{"TPM", current.TPM, previous.TPM},
		{"P95 Latency (ms)", current.LatencyP95Ms, previous.LatencyP95Ms},
		{"P99 Latency (ms)", current.LatencyP99Ms, previous.LatencyP99Ms},
	}

	for _, c := range comparisons {
		if c.current == 0 && c.other == 0 {
			continue
		}
		delta := c.current - c.other
		pctChange := float64(0)
		if c.other != 0 {
			pctChange = (delta / c.other) * 100
		}
		result.Deltas = append(result.Deltas, report.ComparisonDelta{
			Metric:    c.metric,
			Current:   c.current,
			Previous:  c.other,
			Delta:     delta,
			PctChange: pctChange,
		})
	}

	// Determine trend direction using percentage-based thresholds (5%)
	improved := 0
	degraded := 0
	for _, d := range result.Deltas {
		if d.Metric == "P95 Latency (ms)" || d.Metric == "P99 Latency (ms)" {
			// For latency, decrease is improvement
			if d.PctChange < -5 {
				improved++
			} else if d.PctChange > 5 {
				degraded++
			}
		} else {
			// For throughput, increase is improvement
			if d.PctChange > 5 {
				improved++
			} else if d.PctChange < -5 {
				degraded++
			}
		}
	}

	if improved > degraded {
		result.TrendDirection = "improved"
	} else if degraded > improved {
		result.TrendDirection = "degraded"
	} else {
		result.TrendDirection = "stable"
	}

	return result
}

// GetReportBySuiteItemID retrieves a report by its suite_item_id.
func (uc *ReportUsecase) GetReportBySuiteItemID(ctx context.Context, suiteItemID string) (*report.Report, error) {
	row := uc.db.QueryRowContext(ctx, "SELECT "+reportSelectColumns+" FROM reports WHERE suite_item_id = ?", suiteItemID)
	r, err := scanReportRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("report not found for suite_item_id: %s", suiteItemID)
	}
	if err != nil {
		return nil, fmt.Errorf("get report by suite_item_id: %w", err)
	}
	return r, nil
}

// GetReportIDBySuiteItemID retrieves only the report ID for a given suite_item_id.
func (uc *ReportUsecase) GetReportIDBySuiteItemID(ctx context.Context, suiteItemID string) (string, error) {
	var id string
	err := uc.db.QueryRowContext(ctx, "SELECT id FROM reports WHERE suite_item_id = ?", suiteItemID).Scan(&id)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("report not found for suite_item_id: %s", suiteItemID)
	}
	if err != nil {
		return "", fmt.Errorf("get report id by suite_item_id: %w", err)
	}
	return id, nil
}

// InsertRunningReport inserts a report with "running" status before a benchmark completes.
func (uc *ReportUsecase) InsertRunningReport(ctx context.Context, rpt *report.Report) error {
	if uc.db == nil {
		return nil
	}

	query := `
		INSERT INTO reports (
			id, suite_id, suite_item_id, source_type, connection_id, connection_name,
			database_type, template_id, template_name, started_at, ended_at, duration_ms,
			status, error_message, tpm, tps, qps, throughput, latency_avg_ms,
			latency_p95_ms, latency_p99_ms, error_count, metrics_json_path,
			monitoring_json_path, raw_json_path, report_html_path, summary_json_path,
			created_at, updated_at, tags
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := uc.db.ExecContext(ctx, query,
		rpt.ID, rpt.SuiteID, nilIfEmpty(rpt.SuiteItemID), string(rpt.SourceType),
		rpt.ConnectionID, nilIfEmpty(rpt.ConnectionName), rpt.DatabaseType,
		nilIfEmpty(rpt.TemplateID), nilIfEmpty(rpt.TemplateName),
		rpt.StartedAt.Format(time.RFC3339), nil, rpt.DurationMs,
		string(rpt.Status), nilIfEmpty(rpt.ErrorMessage),
		rpt.TPM, rpt.TPS, rpt.QPS, rpt.Throughput,
		rpt.LatencyAvgMs, rpt.LatencyP95Ms, rpt.LatencyP99Ms, rpt.ErrorCount,
		nilIfEmpty(rpt.MetricsJSONPath), nilIfEmpty(rpt.MonitoringJSONPath),
		nilIfEmpty(rpt.RawJSONPath), nilIfEmpty(rpt.ReportHTMLPath),
		nilIfEmpty(rpt.SummaryJSONPath),
		rpt.CreatedAt.Format(time.RFC3339), rpt.UpdatedAt.Format(time.RFC3339),
		nilIfEmpty(rpt.Tags),
	)
	if err != nil {
		return fmt.Errorf("insert running report: %w", err)
	}
	return nil
}

// UpdateReportByItemID updates an existing report identified by suite_item_id.
func (uc *ReportUsecase) UpdateReportByItemID(ctx context.Context, suiteItemID string, rpt *report.Report) error {
	if uc.db == nil {
		return nil
	}

	var endedAt *string
	if rpt.EndedAt != nil {
		s := rpt.EndedAt.Format(time.RFC3339)
		endedAt = &s
	}

	query := `
		UPDATE reports SET
			source_type = ?, connection_id = ?, connection_name = ?,
			database_type = ?, template_id = ?, template_name = ?,
			ended_at = ?, duration_ms = ?, status = ?, error_message = ?,
			tpm = ?, tps = ?, qps = ?, throughput = ?,
			latency_avg_ms = ?, latency_p95_ms = ?, latency_p99_ms = ?, error_count = ?,
			metrics_json_path = ?, monitoring_json_path = ?, raw_json_path = ?,
			report_html_path = ?, summary_json_path = ?,
			updated_at = ?, tags = ?
		WHERE suite_item_id = ?
	`

	_, err := uc.db.ExecContext(ctx, query,
		string(rpt.SourceType),
		rpt.ConnectionID, nilIfEmpty(rpt.ConnectionName), rpt.DatabaseType,
		nilIfEmpty(rpt.TemplateID), nilIfEmpty(rpt.TemplateName),
		endedAt, rpt.DurationMs, string(rpt.Status), nilIfEmpty(rpt.ErrorMessage),
		rpt.TPM, rpt.TPS, rpt.QPS, rpt.Throughput,
		rpt.LatencyAvgMs, rpt.LatencyP95Ms, rpt.LatencyP99Ms, rpt.ErrorCount,
		nilIfEmpty(rpt.MetricsJSONPath), nilIfEmpty(rpt.MonitoringJSONPath),
		nilIfEmpty(rpt.RawJSONPath), nilIfEmpty(rpt.ReportHTMLPath),
		nilIfEmpty(rpt.SummaryJSONPath),
		rpt.UpdatedAt.Format(time.RFC3339), nilIfEmpty(rpt.Tags),
		suiteItemID,
	)
	if err != nil {
		return fmt.Errorf("update report by suite_item_id: %w", err)
	}
	return nil
}

// DetailedDataResult contains the preview and markdown for a report's detailed data.
type DetailedDataResult struct {
	Preview  *report.DetailedDataPreview `json:"preview"`
	Markdown string                      `json:"markdown"`
}

// GetDetailedData reads the persisted bundle and markdown for a report.
func (uc *ReportUsecase) GetDetailedData(ctx context.Context, id string) (*DetailedDataResult, error) {
	rpt, err := uc.GetReport(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}

	if rpt.MetricsJSONPath == "" {
		return &DetailedDataResult{}, nil
	}

	reportDir := filepath.Dir(rpt.MetricsJSONPath)

	// Read bundle
	bundlePath := filepath.Join(reportDir, "report_bundle.json.gz")
	result := &DetailedDataResult{}

	compressed, err := os.ReadFile(bundlePath)
	if err == nil && len(compressed) > 0 {
		bundle, decompressErr := DecompressBundle(compressed)
		if decompressErr == nil {
			gen := NewBundleGenerator()
			result.Preview = gen.BuildPreviewFromBundle(bundle, compressed, id)
		}
	}

	// Read markdown
	mdPath := filepath.Join(reportDir, "report.md")
	if mdData, err := os.ReadFile(mdPath); err == nil {
		result.Markdown = string(mdData)
	}

	return result, nil
}
