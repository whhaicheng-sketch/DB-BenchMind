// Package usecase provides report generation business logic.
// Implements: Phase 5 - Report Generation and Export
package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
	infrareport "github.com/whhaicheng/DB-BenchMind/internal/infra/report"
)

// ReportUseCase provides report generation business operations.
type ReportUseCase struct {
	runRepo       RunRepository
	connUseCase   *ConnectionUseCase
	templateUseCase *TemplateUseCase

	// Registered generators
	generators map[report.ReportFormat]report.Generator
}

// NewReportUseCase creates a new report use case.
func NewReportUseCase(
	runRepo RunRepository,
	connUseCase *ConnectionUseCase,
	templateUseCase *TemplateUseCase,
) *ReportUseCase {
	uc := &ReportUseCase{
		runRepo:         runRepo,
		connUseCase:     connUseCase,
		templateUseCase: templateUseCase,
		generators:      make(map[report.ReportFormat]report.Generator),
	}

	// Register default generators
	uc.RegisterGenerator(infrareport.NewMarkdownGenerator())
	uc.RegisterGenerator(infrareport.NewJSONGenerator())
	uc.RegisterGenerator(infrareport.NewHTMLGenerator())

	return uc
}

// RegisterGenerator registers a report generator.
func (uc *ReportUseCase) RegisterGenerator(generator report.Generator) {
	uc.generators[generator.Format()] = generator
}

// GenerateReport generates a report for a run.
func (uc *ReportUseCase) GenerateReport(ctx context.Context, runID string, format report.ReportFormat, config *report.ReportConfig) (*report.Report, error) {
	// Validate format
	if err := format.Validate(); err != nil {
		return nil, fmt.Errorf("invalid format: %w", err)
	}

	// Set default config if nil
	if config == nil {
		config = report.DefaultConfig(format)
	}
	config.Format = format

	// Get run
	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("find run: %w", err)
	}

	return uc.GenerateReportFromRun(ctx, run, format, config)
}

// GenerateReportFromRun generates a report from a run object.
func (uc *ReportUseCase) GenerateReportFromRun(ctx context.Context, run *execution.Run, format report.ReportFormat, config *report.ReportConfig) (*report.Report, error) {
	// Set default config if nil
	if config == nil {
		config = report.DefaultConfig(format)
	}
	config.Format = format

	// Get generator
	generator, ok := uc.generators[format]
	if !ok {
		return nil, fmt.Errorf("no generator registered for format: %s", format)
	}

	// Build generate context
	genCtx, err := uc.buildGenerateContext(ctx, run, config)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	// Generate report
	rpt, err := generator.Generate(genCtx)
	if err != nil {
		return nil, fmt.Errorf("generate report: %w", err)
	}

	// Save to file if output path specified
	if config.OutputPath != "" {
		if err := uc.saveReport(rpt, config.OutputPath); err != nil {
			return nil, fmt.Errorf("save report: %w", err)
		}
		rpt.FilePath = config.OutputPath
	} else {
		// Generate default path
		defaultPath := uc.getDefaultReportPath(run.ID, format)
		if err := uc.saveReport(rpt, defaultPath); err != nil {
			return nil, fmt.Errorf("save report: %w", err)
		}
		rpt.FilePath = defaultPath
	}

	return rpt, nil
}

// buildGenerateContext builds the generate context from a run.
func (uc *ReportUseCase) buildGenerateContext(ctx context.Context, run *execution.Run, config *report.ReportConfig) (*report.GenerateContext, error) {
	genCtx := report.NewGenerateContext(run.ID, config)

	// Basic info
	genCtx.TaskID = run.TaskID
	genCtx.State = run.State.String()
	genCtx.CreatedAt = run.CreatedAt
	genCtx.StartedAt = run.StartedAt
	genCtx.CompletedAt = run.CompletedAt
	genCtx.Duration = run.Duration
	genCtx.ErrorMessage = run.ErrorMessage

	// Get connection info if available
	// Note: We'd need to store connection_id in runs to do this properly
	// For now, use placeholder values
	genCtx.ConnectionName = "Unknown"
	genCtx.ConnectionType = "unknown"

	// Get template info
	// Note: We'd need to store template_id in runs to do this properly
	genCtx.TemplateName = "Unknown"
	genCtx.Tool = "unknown"

	// Get parameters
	// Note: We'd need to store task parameters
	genCtx.Parameters = make(map[string]interface{})

	// Get metrics from result
	if run.Result != nil {
		genCtx.TPS = run.Result.TPSCalculated
		genCtx.LatencyAvg = run.Result.LatencyAvg
		genCtx.LatencyP95 = run.Result.LatencyP95
		genCtx.LatencyP99 = run.Result.LatencyP99
		genCtx.TotalTransactions = run.Result.TotalTransactions
		genCtx.TotalQueries = run.Result.TotalQueries
		genCtx.ErrorCount = run.Result.ErrorCount
		genCtx.ErrorRate = run.Result.ErrorRate
	}

	// Get time series samples
	genCtx.Samples = make([]report.MetricSample, len(run.Result.TimeSeries))
	for i, s := range run.Result.TimeSeries {
		genCtx.Samples[i] = report.MetricSample{
			Timestamp:   s.Timestamp,
			TPS:         s.TPS,
			LatencyAvg:  s.LatencyAvg,
			LatencyP95:  s.LatencyP95,
			LatencyP99:  s.LatencyP99,
			ErrorRate:   s.ErrorRate,
		}
	}

	// Get logs if requested
	if config.IncludeLogs {
		// TODO: Implement log retrieval from run_logs table
		genCtx.Logs = []report.LogEntry{}
	}

	return genCtx, nil
}

// saveReport saves a report to a file.
func (uc *ReportUseCase) saveReport(rpt *report.Report, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, rpt.Content, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// getDefaultReportPath generates a default report file path.
func (uc *ReportUseCase) getDefaultReportPath(runID string, format report.ReportFormat) string {
	// Use reports directory in current directory
	baseDir := "reports"
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s%s", runID, date, format.FileExtension())
	return filepath.Join(baseDir, filename)
}

// ExportReport exports a report to a specific path.
func (uc *ReportUseCase) ExportReport(ctx context.Context, runID string, format report.ReportFormat, outputPath string) error {
	config := report.DefaultConfig(format)
	config.OutputPath = outputPath

	_, err := uc.GenerateReport(ctx, runID, format, config)
	return err
}

// ListSupportedFormats returns a list of supported report formats.
func (uc *ReportUseCase) ListSupportedFormats() []report.ReportFormat {
	formats := make([]report.ReportFormat, 0, len(uc.generators))
	for format := range uc.generators {
		formats = append(formats, format)
	}
	return formats
}

// IsFormatSupported checks if a format is supported.
func (uc *ReportUseCase) IsFormatSupported(format report.ReportFormat) bool {
	_, ok := uc.generators[format]
	return ok
}
