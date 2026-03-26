// Package usecase provides report collector business logic.
package usecase

import (
	"context"
	"fmt"
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
}

// ReportCollectorOption configures the collector.
type ReportCollectorOption func(*DefaultReportCollector)

// WithReportsDir sets the reports directory.
func WithReportsDir(dir string) ReportCollectorOption {
	return func(c *DefaultReportCollector) {
		c.reportsDir = dir
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

	// Persist files (will be implemented in next task)
	paths := report.ReportPaths{
		MetricsJSON:    filepath.Join(reportDir, "metrics.json"),
		MonitoringJSON: filepath.Join(reportDir, "monitoring.json"),
		RawJSON:        filepath.Join(reportDir, "raw.json"),
		ReportHTML:     filepath.Join(reportDir, "report.html"),
		SummaryJSON:    filepath.Join(reportDir, "summary.json"),
	}

	return &report.ReportResult{
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
	}, nil
}
