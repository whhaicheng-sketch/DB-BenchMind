// Package usecase provides report collector business logic.
package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// ReportCollector collects benchmark results and persists them as reports.
type ReportCollector interface {
	CollectAndPersist(
		ctx context.Context,
		runFn func() (*execution.Run, error),
		rptCtx report.ReportContext,
	) (*report.ReportResult, error)
}

// DefaultReportCollector is the default implementation.
type DefaultReportCollector struct {
	reportsDir string
	db         *sql.DB
}

// ReportCollectorOption configures the collector.
type ReportCollectorOption func(*DefaultReportCollector)

// WithReportsDir sets the reports directory.
func WithReportsDir(dir string) ReportCollectorOption {
	return func(c *DefaultReportCollector) {
		c.reportsDir = dir
	}
}

// WithDB sets the database connection for persisting report records.
func WithDB(db *sql.DB) ReportCollectorOption {
	return func(c *DefaultReportCollector) {
		c.db = db
	}
}

// NewDefaultReportCollector creates a new collector.
func NewDefaultReportCollector(opts ...ReportCollectorOption) *DefaultReportCollector {
	c := &DefaultReportCollector{
		reportsDir: "./reports",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CollectAndPersist executes runFn and persists the result.
func (c *DefaultReportCollector) CollectAndPersist(
	ctx context.Context,
	runFn func() (*execution.Run, error),
	rptCtx report.ReportContext,
) (*report.ReportResult, error) {
	reportID := uuid.New().String()
	startTime := time.Now()

	// Execute benchmark
	run, err := runFn()
	if err != nil {
		return nil, fmt.Errorf("run benchmark: %w", err)
	}

	endTime := time.Now()
	durationMs := endTime.Sub(startTime).Milliseconds()

	// Build report
	rpt := &report.Report{
		ID:             reportID,
		SuiteID:        rptCtx.SuiteID,
		SuiteItemID:    rptCtx.SuiteItemID,
		SourceType:     rptCtx.SourceType,
		ConnectionID:   rptCtx.ConnectionID,
		ConnectionName: rptCtx.ConnectionName,
		DatabaseType:   rptCtx.DatabaseType,
		TemplateID:     rptCtx.TemplateID,
		TemplateName:   rptCtx.TemplateName,
		StartedAt:      startTime,
		EndedAt:        &endTime,
		DurationMs:     durationMs,
		Status:         report.StatusCompleted,
		Tags:           rptCtx.Tags,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Extract metrics from run result
	if run.Result != nil {
		rpt.TPM = run.Result.TPMCalculated
		rpt.TPS = run.Result.TPSCalculated
		rpt.LatencyAvgMs = run.Result.LatencyAvg
		rpt.LatencyP95Ms = run.Result.LatencyP95
		rpt.LatencyP99Ms = run.Result.LatencyP99
		rpt.ErrorCount = run.Result.ErrorCount
	}

	// Handle run state
	if run.State == execution.StateFailed {
		rpt.Status = report.StatusFailed
		rpt.ErrorMessage = run.ErrorMessage
	} else if run.State == execution.StateCancelled || run.State == execution.StateForceStopped {
		rpt.Status = report.StatusCancelled
	}

	// Create directory
	suiteDir := filepath.Join(c.reportsDir, rptCtx.SuiteID)
	reportDir := filepath.Join(suiteDir, reportID)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return nil, fmt.Errorf("create report directory: %w", err)
	}

	// Persist files
	paths, err := c.persistFiles(reportDir, reportID, rpt, run)
	if err != nil {
		return nil, fmt.Errorf("persist files: %w", err)
	}

	// Set file paths on report before DB insert
	rpt.MetricsJSONPath = paths.MetricsJSON
	rpt.MonitoringJSONPath = paths.MonitoringJSON
	rpt.RawJSONPath = paths.RawJSON
	rpt.ReportHTMLPath = paths.ReportHTML
	rpt.SummaryJSONPath = paths.SummaryJSON

	// Persist report record to database
	result := &report.ReportResult{
		ReportID:    reportID,
		ReportPaths: paths,
		Summary: report.ReportSummary{
			Status:       rpt.Status,
			TPM:          rpt.TPM,
			TPS:          rpt.TPS,
			LatencyAvgMs: rpt.LatencyAvgMs,
			LatencyP95Ms: rpt.LatencyP95Ms,
			ErrorCount:   rpt.ErrorCount,
		},
	}
	if err := c.persistToDB(ctx, rpt); err != nil {
		// Log the error but don't fail the overall result — files are already written.
		slog.Error("persist report to database", "report_id", reportID, "error", err)
		result.PersistError = fmt.Errorf("persist report to database: %w", err)
	}

	return result, nil
}

// persistFiles writes all report files to the report directory.
func (c *DefaultReportCollector) persistFiles(
	dir string,
	rptID string,
	rpt *report.Report,
	run *execution.Run,
) (report.ReportPaths, error) {
	paths := report.ReportPaths{
		MetricsJSON:    filepath.Join(dir, "metrics.json"),
		MonitoringJSON: filepath.Join(dir, "monitoring.json"),
		RawJSON:        filepath.Join(dir, "raw.json"),
		SummaryJSON:    filepath.Join(dir, "summary.json"),
		ReportHTML:     filepath.Join(dir, "report.html"),
	}

	// Write metrics.json
	metrics := c.buildMetricsJSON(rpt, run)
	if err := writeJSON(paths.MetricsJSON, metrics); err != nil {
		return paths, fmt.Errorf("write metrics.json: %w", err)
	}

	// Write monitoring.json
	monitoring := c.buildMonitoringJSON(rptID, run)
	if err := writeJSON(paths.MonitoringJSON, monitoring); err != nil {
		return paths, fmt.Errorf("write monitoring.json: %w", err)
	}

	// Write raw.json
	raw := c.buildRawJSON(rptID, run)
	if err := writeJSON(paths.RawJSON, raw); err != nil {
		return paths, fmt.Errorf("write raw.json: %w", err)
	}

	// Write summary.json
	summary := c.buildSummaryJSON(rpt)
	if err := writeJSON(paths.SummaryJSON, summary); err != nil {
		return paths, fmt.Errorf("write summary.json: %w", err)
	}

	// Write report.html (minimal for Phase 1)
	html := c.buildReportHTML(rpt)
	if err := os.WriteFile(paths.ReportHTML, []byte(html), 0644); err != nil {
		return paths, fmt.Errorf("write report.html: %w", err)
	}

	return paths, nil
}

// persistToDB inserts the report record into the SQLite reports table.
// If a row with the same suite_item_id already exists (from a running report),
// it updates the existing row instead.
func (c *DefaultReportCollector) persistToDB(ctx context.Context, rpt *report.Report) error {
	if c.db == nil {
		return nil // No database configured, skip DB persistence
	}

	// Check if a running report row exists for this suite_item_id
	if rpt.SuiteItemID != "" {
		var existingID string
		err := c.db.QueryRowContext(ctx,
			"SELECT id FROM reports WHERE suite_item_id = ?",
			rpt.SuiteItemID,
		).Scan(&existingID)
		if err == nil {
			// Row exists — update it with final data
			rpt.ID = existingID
			rpt.UpdatedAt = time.Now()
			return c.updateReportRow(ctx, rpt)
		}
	}

	var endedAt *string
	if rpt.EndedAt != nil {
		s := rpt.EndedAt.Format(time.RFC3339)
		endedAt = &s
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

	_, err := c.db.ExecContext(ctx, query,
		rpt.ID, rpt.SuiteID, nilIfEmpty(rpt.SuiteItemID), string(rpt.SourceType),
		rpt.ConnectionID, nilIfEmpty(rpt.ConnectionName), rpt.DatabaseType,
		nilIfEmpty(rpt.TemplateID), nilIfEmpty(rpt.TemplateName),
		rpt.StartedAt.Format(time.RFC3339), endedAt, rpt.DurationMs,
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
		return fmt.Errorf("insert report row: %w", err)
	}

	return nil
}

// updateReportRow updates an existing report row with final data.
func (c *DefaultReportCollector) updateReportRow(ctx context.Context, rpt *report.Report) error {
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
		WHERE id = ?
	`

	_, err := c.db.ExecContext(ctx, query,
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
		rpt.ID,
	)
	if err != nil {
		return fmt.Errorf("update report row: %w", err)
	}
	return nil
}

// writeJSON writes data to path as formatted JSON (atomic write).
func writeJSON(path string, data interface{}) error {
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, path) // Atomic
}

// buildMetricsJSON creates the metrics.json structure.
func (c *DefaultReportCollector) buildMetricsJSON(rpt *report.Report, run *execution.Run) map[string]interface{} {
	metrics := map[string]interface{}{
		"schema_version": "v1",
		"report_id":      rpt.ID,
		"suite_id":       rpt.SuiteID,
		"suite_item_id":  rpt.SuiteItemID,
		"generated_at":   time.Now().Format(time.RFC3339),
		"benchmark": map[string]interface{}{
			"connection_id":   rpt.ConnectionID,
			"connection_name": rpt.ConnectionName,
			"database_type":   rpt.DatabaseType,
			"template_id":     rpt.TemplateID,
			"template_name":   rpt.TemplateName,
		},
		"execution": map[string]interface{}{
			"status":      string(rpt.Status),
			"started_at":  rpt.StartedAt.Format(time.RFC3339),
			"duration_ms": rpt.DurationMs,
		},
		"summary": map[string]interface{}{
			"tpm":            rpt.TPM,
			"tps":            rpt.TPS,
			"latency_avg_ms": rpt.LatencyAvgMs,
			"latency_p95_ms": rpt.LatencyP95Ms,
			"latency_p99_ms": rpt.LatencyP99Ms,
			"error_count":    rpt.ErrorCount,
		},
	}

	if run.Result != nil {
		metrics["percentiles"] = map[string]interface{}{
			"p50": run.Result.LatencyAvg, // Approximation
			"p90": run.Result.LatencyP95,
			"p95": run.Result.LatencyP95,
			"p99": run.Result.LatencyP99,
		}
	}

	return metrics
}

// buildMonitoringJSON creates the monitoring.json structure.
func (c *DefaultReportCollector) buildMonitoringJSON(reportID string, run *execution.Run) map[string]interface{} {
	monitoring := map[string]interface{}{
		"schema_version":     "v1",
		"report_id":          reportID,
		"generated_at":       time.Now().Format(time.RFC3339),
		"sample_interval_ms": 1000,
	}

	// Build time series from run result
	if run.Result != nil && len(run.Result.TimeSeries) > 0 {
		timestamps := make([]int64, len(run.Result.TimeSeries))
		tps := make([]float64, len(run.Result.TimeSeries))
		latency := make([]float64, len(run.Result.TimeSeries))

		for i, sample := range run.Result.TimeSeries {
			timestamps[i] = sample.Timestamp.Unix()
			tps[i] = sample.TPS
			latency[i] = sample.LatencyAvg
		}

		monitoring["time_series"] = map[string]interface{}{
			"timestamps": timestamps,
			"tps":        tps,
			"latency_ms": latency,
		}
		monitoring["total_samples"] = len(run.Result.TimeSeries)
	} else {
		monitoring["total_samples"] = 0
		monitoring["time_series"] = map[string]interface{}{
			"timestamps": []int64{},
			"tps":        []float64{},
			"latency_ms": []float64{},
		}
	}

	// Phase 2 reserved fields
	monitoring["system"] = map[string]interface{}{
		"timestamps":     []int64{},
		"cpu_percent":    []float64{},
		"disk_read_bps":  []float64{},
		"disk_write_bps": []float64{},
	}

	return monitoring
}

// buildRawJSON creates the raw.json structure.
func (c *DefaultReportCollector) buildRawJSON(reportID string, run *execution.Run) map[string]interface{} {
	raw := map[string]interface{}{
		"schema_version": "v1",
		"report_id":      reportID,
		"generated_at":   time.Now().Format(time.RFC3339),
		"adapter_type":   "sysbench", // Default
		"stdout":         "",
		"stderr":         "",
		"exit_code":      0,
	}

	if run.Result != nil {
		raw["parsed_result"] = map[string]interface{}{
			"raw": map[string]interface{}{
				"total_transactions": run.Result.TotalTransactions,
				"total_queries":      run.Result.TotalQueries,
				"read_queries":       run.Result.ReadQueries,
				"write_queries":      run.Result.WriteQueries,
				"errors":             run.Result.ErrorCount,
			},
		}
	}

	if run.State == execution.StateFailed {
		raw["exit_code"] = 1
		raw["stderr"] = run.ErrorMessage
	}

	return raw
}

// buildSummaryJSON creates the summary.json structure.
func (c *DefaultReportCollector) buildSummaryJSON(rpt *report.Report) map[string]interface{} {
	duration := ""
	if rpt.DurationMs > 0 {
		d := time.Duration(rpt.DurationMs) * time.Millisecond
		if d >= time.Minute {
			duration = fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
		} else {
			duration = fmt.Sprintf("%.1fs", d.Seconds())
		}
	}

	return map[string]interface{}{
		"schema_version": "v1",
		"report_id":      rpt.ID,
		"suite_id":       rpt.SuiteID,
		"generated_at":   time.Now().Format(time.RFC3339),
		"display": map[string]interface{}{
			"title":    fmt.Sprintf("%s - %s", rpt.ConnectionName, rpt.TemplateName),
			"subtitle": duration,
			"status":   string(rpt.Status),
		},
		"key_metrics": map[string]interface{}{
			"tpm":            rpt.TPM,
			"tps":            rpt.TPS,
			"latency_avg_ms": rpt.LatencyAvgMs,
			"latency_p95_ms": rpt.LatencyP95Ms,
		},
		"duration_display": duration,
		"database_type":    rpt.DatabaseType,
		"connection_name":  rpt.ConnectionName,
	}
}

// buildReportHTML creates a minimal HTML report.
func (c *DefaultReportCollector) buildReportHTML(rpt *report.Report) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Report %s</title>
    <style>
        body { font-family: system-ui, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .metric { margin: 10px 0; }
        .label { color: #666; }
        .value { font-weight: bold; }
    </style>
</head>
<body>
    <h1>Benchmark Report</h1>
    <div class="metric"><span class="label">Status:</span> <span class="value">%s</span></div>
    <div class="metric"><span class="label">TPM:</span> <span class="value">%.2f</span></div>
    <div class="metric"><span class="label">TPS:</span> <span class="value">%.2f</span></div>
    <div class="metric"><span class="label">Latency Avg:</span> <span class="value">%.2f ms</span></div>
    <div class="metric"><span class="label">Latency P95:</span> <span class="value">%.2f ms</span></div>
</body>
</html>`, rpt.ID, rpt.Status, rpt.TPM, rpt.TPS, rpt.LatencyAvgMs, rpt.LatencyP95Ms)
}
