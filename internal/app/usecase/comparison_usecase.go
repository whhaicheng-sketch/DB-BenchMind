// Package usecase provides comparison business logic.
package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

// ComparisonUseCase provides comparison business logic.
type ComparisonUseCase struct {
	historyRepo repository.HistoryRepository
}

// NewComparisonUseCase creates a new comparison use case.
func NewComparisonUseCase(historyRepo repository.HistoryRepository) *ComparisonUseCase {
	return &ComparisonUseCase{
		historyRepo: historyRepo,
	}
}

// GetAllRecords retrieves all history records for comparison selection.
func (uc *ComparisonUseCase) GetAllRecords(ctx context.Context) ([]*history.Record, error) {
	return uc.historyRepo.GetAll(ctx)
}

// GetRecordRefs returns summary references of all history records.
func (uc *ComparisonUseCase) GetRecordRefs(ctx context.Context) ([]*comparison.RecordRef, error) {
	records, err := uc.historyRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	refs := make([]*comparison.RecordRef, len(records))
	for i, record := range records {
		durationSec := record.Duration.Seconds()
		qps := 0.0
		if durationSec > 0 && record.TotalQueries > 0 {
			qps = float64(record.TotalQueries) / durationSec
		}

		refs[i] = &comparison.RecordRef{
			ID:             record.ID,
			TemplateName:    record.TemplateName,
			DatabaseType:    record.DatabaseType,
			Threads:        record.Threads,
			ConnectionName: record.ConnectionName,
			StartTime:      record.StartTime,
			TPS:            record.TPSCalculated,
			LatencyAvg:     record.LatencyAvg,
			LatencyMin:     record.LatencyMin,
			LatencyMax:     record.LatencyMax,
			LatencyP95:     record.LatencyP95,
			LatencyP99:     record.LatencyP99,
			Duration:       record.Duration,
			QPS:            qps,
			ReadQueries:   record.ReadQueries,
			WriteQueries:  record.WriteQueries,
			OtherQueries:  record.OtherQueries,
		}
	}

	return refs, nil
}

// CompareRecords compares selected history records.
func (uc *ComparisonUseCase) CompareRecords(ctx context.Context, recordIDs []string, groupBy comparison.GroupByField) (*comparison.MultiConfigComparison, error) {
	if len(recordIDs) < 2 {
		return nil, fmt.Errorf("at least 2 records must be selected for comparison")
	}

	slog.Info("Comparison: Starting comparison", "record_count", len(recordIDs), "group_by", groupBy)

	// Fetch all records
	allRecords, err := uc.historyRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch records: %w", err)
	}

	// Filter selected records
	recordMap := make(map[string]*history.Record)
	for _, record := range allRecords {
		for _, id := range recordIDs {
			if record.ID == id {
				recordMap[id] = record
				break
			}
		}
	}

	if len(recordMap) != len(recordIDs) {
		return nil, fmt.Errorf("some records not found: expected %d, found %d", len(recordIDs), len(recordMap))
	}

	// Convert to slice in original order
	selectedRecords := make([]*history.Record, len(recordIDs))
	for i, id := range recordIDs {
		selectedRecords[i] = recordMap[id]
	}

	// Perform comparison
	result, err := comparison.CompareMultiConfig(selectedRecords, groupBy)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}

	slog.Info("Comparison: Completed successfully", "id", result.ID)
	return result, nil
}

// FilterRecords filters records by criteria.
func (uc *ComparisonUseCase) FilterRecords(ctx context.Context, refs []*comparison.RecordRef, filter *ComparisonFilter) []*comparison.RecordRef {
	if filter == nil {
		return refs
	}

	var filtered []*comparison.RecordRef
	for _, ref := range refs {
		if filter.DatabaseType != "" && ref.DatabaseType != filter.DatabaseType {
			continue
		}
		if filter.TemplateName != "" && ref.TemplateName != filter.TemplateName {
			continue
		}
		if filter.MinThreads > 0 && ref.Threads < filter.MinThreads {
			continue
		}
		if filter.MaxThreads > 0 && ref.Threads > filter.MaxThreads {
			continue
		}
		filtered = append(filtered, ref)
	}

	return filtered
}

// ComparisonFilter defines filter criteria for comparison records.
type ComparisonFilter struct {
	DatabaseType string
	TemplateName string
	MinThreads    int
	MaxThreads    int
}

// GenerateComprehensiveReport generates a comprehensive comparison report
// with grouped configurations and statistical analysis.
//
// This is the new method that replaces CompareRecords for professional
// multi-configuration analysis with N runs per config.
//
// Parameters:
//   - ctx: Context
//   - recordIDs: IDs of history records to include (or empty for all records)
//   - groupBy: Primary grouping dimension
//   - similarityConfig: Auto-detection settings (optional, uses defaults if nil)
//
// Returns:
//   - *comparison.ComparisonReport: Comprehensive report with all analysis
//   - error: If report generation fails
func (uc *ComparisonUseCase) GenerateComprehensiveReport(
	ctx context.Context,
	recordIDs []string,
	groupBy comparison.GroupByField,
	similarityConfig *comparison.SimilarityConfig,
) (*comparison.ComparisonReport, error) {
	slog.Info("Comparison: Generating comprehensive report",
		"record_ids_count", len(recordIDs), "group_by", groupBy)

	// Fetch records
	var records []*history.Record
	var err error

	if len(recordIDs) > 0 {
		// Fetch specific records
		records, err = uc.getRecordsByID(ctx, recordIDs)
	} else {
		// Fetch all records
		records, err = uc.historyRepo.GetAll(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("fetch records: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("need at least 2 records for comparison, got %d", len(records))
	}

	slog.Info("Comparison: Records loaded", "count", len(records))

	// Use default similarity config if not provided
	if similarityConfig == nil {
		similarityConfig = comparison.DefaultSimilarityConfig()
		similarityConfig.GroupBy = groupBy
	}

	// Group records by configuration
	configGroups, err := comparison.GroupRecordsByConfig(records, groupBy, similarityConfig)
	if err != nil {
		return nil, fmt.Errorf("group records: %w", err)
	}

	slog.Info("Comparison: Records grouped", "groups", len(configGroups))

	// Create report
	report := &comparison.ComparisonReport{
		GeneratedAt:     time.Now(),
		ReportID:        comparison.FormatReportID(),
		GroupBy:         groupBy,
		ConfigGroups:    configGroups,
		SimilarityConfig: similarityConfig,
	}

	// Perform scaling analysis
	report.ScalingAnalysis = uc.performScalingAnalysis(configGroups)

	// Perform sanity checks
	report.SanityChecks = comparison.ValidateReport(report)

	// Generate findings
	report.Findings = comparison.GenerateReportFindings(report)

	slog.Info("Comparison: Report generated successfully",
		"report_id", report.ReportID,
		"groups", len(configGroups))

	return report, nil
}

// getRecordsByID fetches specific records by their IDs.
func (uc *ComparisonUseCase) getRecordsByID(ctx context.Context, recordIDs []string) ([]*history.Record, error) {
	if len(recordIDs) == 0 {
		return []*history.Record{}, nil
	}

	allRecords, err := uc.historyRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Create ID map for efficient lookup
	idMap := make(map[string]bool)
	for _, id := range recordIDs {
		idMap[id] = true
	}

	// Filter records by ID
	var filtered []*history.Record
	for _, record := range allRecords {
		if idMap[record.ID] {
			filtered = append(filtered, record)
		}
	}

	// Verify all requested records were found
	if len(filtered) != len(recordIDs) {
		return nil, fmt.Errorf("some records not found: expected %d, found %d",
			len(recordIDs), len(filtered))
	}

	return filtered, nil
}

// performScalingAnalysis performs scaling analysis on config groups.
func (uc *ComparisonUseCase) performScalingAnalysis(groups []*comparison.ConfigGroup) *comparison.ScalingAnalysis {
	if len(groups) == 0 {
		return nil
	}

	// Find baseline (threads=1)
	var baseline *comparison.ConfigGroup
	for _, group := range groups {
		if group.Config.Threads == 1 {
			baseline = group
			break
		}
	}

	// Perform analysis
	return comparison.AnalyzeScaling(groups, baseline)
}

// ExportReport exports a comparison report to file.
// Supported formats: "markdown", "txt"
func (uc *ComparisonUseCase) ExportReport(
	ctx context.Context,
	report *comparison.ComparisonReport,
	format string,
	filepath string,
) error {
	if report == nil {
		return fmt.Errorf("report is nil")
	}

	var content string
	switch format {
	case "markdown", "md":
		content = report.FormatMarkdown()
	case "txt":
		content = report.FormatTXT()
	default:
		return fmt.Errorf("unsupported format: %s (supported: markdown, txt)", format)
	}

	// Write to file
	err := os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	slog.Info("Comparison: Report exported",
		"format", format,
		"filepath", filepath,
		"report_id", report.ReportID)

	return nil
}
