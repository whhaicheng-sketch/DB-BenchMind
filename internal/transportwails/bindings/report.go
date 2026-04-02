// Package bindings provides Wails bindings for the frontend.
package bindings

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// ReportBinding exposes report-related operations to the frontend.
type ReportBinding struct {
	uc *usecase.ReportUsecase
}

// NewReportBinding creates a new ReportBinding.
func NewReportBinding(uc *usecase.ReportUsecase) *ReportBinding {
	return &ReportBinding{uc: uc}
}

// ReportDTO is the data transfer object for reports.
type ReportDTO struct {
	ID             string  `json:"id"`
	SuiteID        string  `json:"suite_id"`
	SuiteItemID    string  `json:"suite_item_id,omitempty"`
	SourceType     string  `json:"source_type"`
	ConnectionID   string  `json:"connection_id"`
	ConnectionName string  `json:"connection_name,omitempty"`
	DatabaseType   string  `json:"database_type"`
	TemplateID     string  `json:"template_id,omitempty"`
	TemplateName   string  `json:"template_name,omitempty"`
	StartedAt      string  `json:"started_at"`
	EndedAt        string  `json:"ended_at,omitempty"`
	DurationMs     int64   `json:"duration_ms,omitempty"`
	Status         string  `json:"status"`
	ErrorMessage   string  `json:"error_message,omitempty"`
	TPM            float64 `json:"tpm,omitempty"`
	TPS            float64 `json:"tps,omitempty"`
	QPS            float64 `json:"qps,omitempty"`
	Throughput     float64 `json:"throughput,omitempty"`
	LatencyAvgMs   float64 `json:"latency_avg_ms,omitempty"`
	LatencyP95Ms   float64 `json:"latency_p95_ms,omitempty"`
	LatencyP99Ms   float64 `json:"latency_p99_ms,omitempty"`
	ErrorCount     int64   `json:"error_count,omitempty"`
	Tags           string  `json:"tags,omitempty"`
}

// SuiteDTO is the data transfer object for suites.
type SuiteDTO struct {
	ID                    string `json:"id"`
	Name                  string `json:"name,omitempty"`
	ExecutionMode         string `json:"execution_mode,omitempty"`
	FailurePolicy         string `json:"failure_policy,omitempty"`
	CleanupEnabled        bool   `json:"cleanup_enabled"`
	SuiteManifestJSONPath string `json:"suite_manifest_json_path,omitempty"`
	Status                string `json:"status"`
	StartedAt             string `json:"started_at,omitempty"`
	EndedAt               string `json:"ended_at,omitempty"`
	TotalItems            int    `json:"total_items"`
	CompletedItems        int    `json:"completed_items"`
	SuccessItems          int    `json:"success_items"`
	FailedItems           int    `json:"failed_items"`
	SkippedItems          int    `json:"skipped_items"`
	SuiteReportJSONPath   string `json:"suite_report_json_path,omitempty"`
	SuiteReportHTMLPath   string `json:"suite_report_html_path,omitempty"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

// ReportListResult is the result of ListReports.
type ReportListResult struct {
	Reports []ReportDTO `json:"reports"`
	Total   int         `json:"total"`
	Error   string      `json:"error,omitempty"`
}

// ReportResult is the result of GetReport.
type ReportResult struct {
	Report *ReportDTO `json:"report,omitempty"`
	Error  string     `json:"error,omitempty"`
}

// ReportMetricsResult is the result of GetReportMetrics.
type ReportMetricsResult struct {
	Metrics *report.MetricsData `json:"metrics,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// SuiteListResult is the result of ListSuites.
type SuiteListResult struct {
	Suites []SuiteDTO `json:"suites"`
	Total  int        `json:"total"`
	Error  string     `json:"error,omitempty"`
}

// SuiteResult is the result of GetSuite.
type SuiteResult struct {
	Suite *SuiteDTO `json:"suite,omitempty"`
	Error string    `json:"error,omitempty"`
}

// ListReportsOptionsDTO contains options for listing reports.
type ListReportsOptionsDTO struct {
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	SuiteID      string `json:"suite_id,omitempty"`
	Status       string `json:"status,omitempty"`
	ConnectionID string `json:"connection_id,omitempty"`
}

// ListSuitesOptionsDTO contains options for listing suites.
type ListSuitesOptionsDTO struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status,omitempty"`
}

// ListReports retrieves reports with pagination and filtering.
func (b *ReportBinding) ListReports(opts ListReportsOptionsDTO) ReportListResult {
	ctx := context.Background()
	reports, total, err := b.uc.ListReports(ctx, usecase.ListReportsOptions{
		Page:         opts.Page,
		PageSize:     opts.PageSize,
		SuiteID:      opts.SuiteID,
		Status:       opts.Status,
		ConnectionID: opts.ConnectionID,
	})
	if err != nil {
		slog.Error("ListReports failed", "error", err)
		return ReportListResult{Error: err.Error()}
	}

	dtos := make([]ReportDTO, 0, len(reports))
	for _, r := range reports {
		dtos = append(dtos, reportToDTO(r))
	}

	return ReportListResult{Reports: dtos, Total: total}
}

// GetReport retrieves a single report by ID.
func (b *ReportBinding) GetReport(id string) ReportResult {
	ctx := context.Background()
	r, err := b.uc.GetReport(ctx, id)
	if err != nil {
		slog.Error("GetReport failed", "id", id, "error", err)
		return ReportResult{Error: err.Error()}
	}
	dto := reportToDTO(r)
	return ReportResult{Report: &dto}
}

// GetReportMetrics retrieves the full metrics data for a report.
func (b *ReportBinding) GetReportMetrics(id string) ReportMetricsResult {
	ctx := context.Background()
	metrics, err := b.uc.GetReportMetrics(ctx, id)
	if err != nil {
		slog.Error("GetReportMetrics failed", "id", id, "error", err)
		return ReportMetricsResult{Error: err.Error()}
	}
	return ReportMetricsResult{Metrics: metrics}
}

// ListSuites retrieves suites with pagination and filtering.
func (b *ReportBinding) ListSuites(opts ListSuitesOptionsDTO) SuiteListResult {
	ctx := context.Background()
	suites, total, err := b.uc.ListSuites(ctx, usecase.ListSuitesOptions{
		Page:     opts.Page,
		PageSize: opts.PageSize,
		Status:   opts.Status,
	})
	if err != nil {
		slog.Error("ListSuites failed", "error", err)
		return SuiteListResult{Error: err.Error()}
	}

	dtos := make([]SuiteDTO, 0, len(suites))
	for _, s := range suites {
		dtos = append(dtos, suiteToDTO(s))
	}

	return SuiteListResult{Suites: dtos, Total: total}
}

// GetSuite retrieves a single suite by ID.
func (b *ReportBinding) GetSuite(id string) SuiteResult {
	ctx := context.Background()
	s, err := b.uc.GetSuite(ctx, id)
	if err != nil {
		slog.Error("GetSuite failed", "id", id, "error", err)
		return SuiteResult{Error: err.Error()}
	}
	dto := suiteToDTO(s)
	return SuiteResult{Suite: &dto}
}

func reportToDTO(r *report.Report) ReportDTO {
	dto := ReportDTO{
		ID:           r.ID,
		SuiteID:      r.SuiteID,
		SuiteItemID:  r.SuiteItemID,
		SourceType:   string(r.SourceType),
		ConnectionID: r.ConnectionID,
		DatabaseType: r.DatabaseType,
		TemplateID:   r.TemplateID,
		TemplateName: r.TemplateName,
		Status:       string(r.Status),
		ErrorMessage: r.ErrorMessage,
		TPM:          r.TPM,
		TPS:          r.TPS,
		QPS:          r.QPS,
		Throughput:   r.Throughput,
		LatencyAvgMs: r.LatencyAvgMs,
		LatencyP95Ms: r.LatencyP95Ms,
		LatencyP99Ms: r.LatencyP99Ms,
		ErrorCount:   r.ErrorCount,
		Tags:         r.Tags,
	}

	if !r.StartedAt.IsZero() {
		dto.StartedAt = r.StartedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if r.EndedAt != nil {
		dto.EndedAt = r.EndedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	dto.DurationMs = r.DurationMs

	return dto
}

func suiteToDTO(s *report.Suite) SuiteDTO {
	dto := SuiteDTO{
		ID:                    s.ID,
		Name:                  s.Name,
		ExecutionMode:         s.ExecutionMode,
		FailurePolicy:         s.FailurePolicy,
		CleanupEnabled:        s.CleanupEnabled,
		SuiteManifestJSONPath: s.SuiteManifestJSONPath,
		Status:                string(s.Status),
		TotalItems:            s.TotalItems,
		CompletedItems:        s.CompletedItems,
		SuccessItems:          s.SuccessItems,
		FailedItems:           s.FailedItems,
		SkippedItems:          s.SkippedItems,
		SuiteReportJSONPath:   s.SuiteReportJSONPath,
		SuiteReportHTMLPath:   s.SuiteReportHTMLPath,
	}

	if s.StartedAt != nil {
		dto.StartedAt = s.StartedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if s.EndedAt != nil {
		dto.EndedAt = s.EndedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !s.CreatedAt.IsZero() {
		dto.CreatedAt = s.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !s.UpdatedAt.IsZero() {
		dto.UpdatedAt = s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return dto
}

// DeleteAllResult is the result of DeleteAllReports.
type DeleteAllResult struct {
	Count int    `json:"count"`
	Error string `json:"error,omitempty"`
}

// DeleteAllReports deletes all reports.
func (b *ReportBinding) DeleteAllReports() DeleteAllResult {
	count, err := b.uc.DeleteAllReports(context.Background())
	if err != nil {
		slog.Error("DeleteAllReports failed", "error", err)
		return DeleteAllResult{Error: err.Error()}
	}
	return DeleteAllResult{Count: count}
}

// DeleteReportResult is the result of DeleteReport.
type DeleteReportResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DeleteReport deletes a report by ID.
func (b *ReportBinding) DeleteReport(id string) DeleteReportResult {
	if err := b.uc.DeleteReport(context.Background(), id); err != nil {
		slog.Error("DeleteReport failed", "id", id, "error", err)
		return DeleteReportResult{Error: err.Error()}
	}
	return DeleteReportResult{Success: true}
}

// ExportJSONResult is the result of ExportReportJSON.
type ExportJSONResult struct {
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// ExportHTMLResult is the result of ExportReportHTML.
type ExportHTMLResult struct {
	HTML  string `json:"html,omitempty"`
	Error string `json:"error,omitempty"`
}

// ExportFilePathsResult is the result of GetExportFilePaths.
type ExportFilePathsResult struct {
	Metrics    string `json:"metrics,omitempty"`
	Monitoring string `json:"monitoring,omitempty"`
	Raw        string `json:"raw,omitempty"`
	HTML       string `json:"html,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ExportReportJSON exports all report data as a JSON string.
func (b *ReportBinding) ExportReportJSON(id string) ExportJSONResult {
	ctx := context.Background()
	export, err := b.uc.ExportReportJSON(ctx, id)
	if err != nil {
		slog.Error("ExportReportJSON failed", "id", id, "error", err)
		return ExportJSONResult{Error: err.Error()}
	}

	data, err := json.Marshal(export)
	if err != nil {
		slog.Error("Marshal export data failed", "id", id, "error", err)
		return ExportJSONResult{Error: err.Error()}
	}

	return ExportJSONResult{Data: string(data)}
}

// ExportReportHTML returns the HTML content for a report.
func (b *ReportBinding) ExportReportHTML(id string) ExportHTMLResult {
	ctx := context.Background()
	html, err := b.uc.ExportReportHTML(ctx, id)
	if err != nil {
		slog.Error("ExportReportHTML failed", "id", id, "error", err)
		return ExportHTMLResult{Error: err.Error()}
	}
	return ExportHTMLResult{HTML: html}
}

// GetExportFilePaths returns the file paths for all export files.
func (b *ReportBinding) GetExportFilePaths(id string) ExportFilePathsResult {
	ctx := context.Background()
	metrics, monitoring, raw, html, err := b.uc.GetExportFilePaths(ctx, id)
	if err != nil {
		slog.Error("GetExportFilePaths failed", "id", id, "error", err)
		return ExportFilePathsResult{Error: err.Error()}
	}
	return ExportFilePathsResult{
		Metrics:    metrics,
		Monitoring: monitoring,
		Raw:        raw,
		HTML:       html,
	}
}

// DetailedDataResult is the result of GetDetailedData.
type DetailedDataResult struct {
	Preview  *report.DetailedDataPreview `json:"preview,omitempty"`
	Markdown string                      `json:"markdown,omitempty"`
	Error    string                      `json:"error,omitempty"`
}

// GetDetailedData retrieves the detailed data (AI Bundle preview + markdown) for a report.
func (b *ReportBinding) GetDetailedData(id string) DetailedDataResult {
	ctx := context.Background()
	result, err := b.uc.GetDetailedData(ctx, id)
	if err != nil {
		slog.Error("GetDetailedData failed", "id", id, "error", err)
		return DetailedDataResult{Error: err.Error()}
	}
	return DetailedDataResult{
		Preview:  result.Preview,
		Markdown: result.Markdown,
	}
}
