// Package usecase provides report query business logic.
package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
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

// GetReport retrieves a single report by ID.
func (uc *ReportUsecase) GetReport(ctx context.Context, id string) (*report.Report, error) {
	query := `
		SELECT id, suite_id, suite_item_id, source_type, connection_id, connection_name,
			database_type, template_id, template_name, started_at, ended_at, duration_ms,
			status, error_message, tpm, tps, qps, throughput, latency_avg_ms,
			latency_p95_ms, latency_p99_ms, error_count, metrics_json_path,
			monitoring_json_path, raw_json_path, report_html_path, summary_json_path,
			created_at, updated_at, tags
		FROM reports WHERE id = ?
	`

	r := &report.Report{}
	var endedAt, startedAt, createdAt, updatedAt *string
	var suiteItemID, connectionName, templateID, templateName, errorMessage *string
	var durationMs *int64
	var tpm, tps, qps, throughput, latencyAvgMs, latencyP95Ms, latencyP99Ms *float64
	var errorCount *int64
	var metricsJSONPath, monitoringJSONPath, rawJSONPath, reportHTMLPath, summaryJSONPath, tags *string
	var sourceType string
	var status string

	err := uc.db.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.SuiteID, &suiteItemID, &sourceType, &r.ConnectionID, &connectionName,
		&r.DatabaseType, &templateID, &templateName, &startedAt, &endedAt, &durationMs,
		&status, &errorMessage, &tpm, &tps, &qps, &throughput, &latencyAvgMs,
		&latencyP95Ms, &latencyP99Ms, &errorCount, &metricsJSONPath,
		&monitoringJSONPath, &rawJSONPath, &reportHTMLPath, &summaryJSONPath,
		&createdAt, &updatedAt, &tags,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("report not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
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

// GetExportFilePaths returns the file paths for all export files.
func (uc *ReportUsecase) GetExportFilePaths(ctx context.Context, id string) (metrics, monitoring, raw, html string, err error) {
	rpt, err := uc.GetReport(ctx, id)
	if err != nil {
		return "", "", "", "", fmt.Errorf("get report: %w", err)
	}

	return rpt.MetricsJSONPath, rpt.MonitoringJSONPath, rpt.RawJSONPath, rpt.ReportHTMLPath, nil
}
