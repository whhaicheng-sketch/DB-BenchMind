// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
)

// ComparisonBinding provides Wails bindings for benchmark comparison.
type ComparisonBinding struct {
	historyUC    *usecase.HistoryUseCase
	comparisonUC *usecase.ComparisonUseCase
	ctx          context.Context
}

// NewComparisonBinding creates a new ComparisonBinding.
func NewComparisonBinding(historyUC *usecase.HistoryUseCase, comparisonUC *usecase.ComparisonUseCase) *ComparisonBinding {
	return &ComparisonBinding{
		historyUC:    historyUC,
		comparisonUC: comparisonUC,
	}
}

// SetContext sets the Wails context.
func (b *ComparisonBinding) SetContext(ctx context.Context) {
	b.ctx = ctx
}

// =============================================================================
// DTO Types
// =============================================================================

// ComparisonRecordRefDTO represents a record reference for comparison selection.
type ComparisonRecordRefDTO struct {
	ID             string  `json:"id"`
	TemplateName   string  `json:"template_name"`
	DatabaseType   string  `json:"database_type"`
	Threads        int     `json:"threads"`
	ConnectionName string  `json:"connection_name"`
	StartTime      string  `json:"start_time"`
	TPS            float64 `json:"tps"`
	LatencyAvg     float64 `json:"latency_avg_ms"`
	Duration       string  `json:"duration"`
}

// ComparisonRecordDetailDTO represents detailed record for comparison.
type ComparisonRecordDetailDTO struct {
	ComparisonRecordRefDTO
	QPS           float64 `json:"qps"`
	ReadQueries   int64   `json:"read_queries"`
	WriteQueries  int64   `json:"write_queries"`
	OtherQueries  int64   `json:"other_queries"`
	TotalQueries  int64   `json:"total_queries"`
}

// RecordRefsResult represents the result of GetRecordRefs.
type RecordRefsResult struct {
	Refs  []ComparisonRecordRefDTO `json:"refs"`
	Error string                   `json:"error,omitempty"`
}

// CompareResult represents the result of CompareRecords.
type CompareResult struct {
	Success bool                   `json:"success"`
	Result  *ComparisonResultDTO   `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ComparisonResultDTO represents comparison result for frontend.
type ComparisonResultDTO struct {
	ReportID    string                   `json:"report_id"`
	GeneratedAt string                   `json:"generated_at"`
	GroupBy     comparison.GroupByField  `json:"group_by"`
	Groups      []ConfigGroupDTO         `json:"groups"`
}

// ConfigGroupDTO represents a config group for frontend.
type ConfigGroupDTO struct {
	Threads     int                       `json:"threads"`
	N           int                       `json:"n"`
	Records     []ComparisonRecordDetailDTO `json:"records"`
	TPSMean     float64                   `json:"tps_mean"`
	TPSStdDev   float64                   `json:"tps_std_dev"`
	TPSMin      float64                   `json:"tps_min"`
	TPSMax      float64                   `json:"tps_max"`
	LatencyMean float64                   `json:"latency_mean_ms"`
	LatencyP95  float64                   `json:"latency_p95_ms"`
}

// ExportResult represents the result of export operations.
type ExportResult struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Filepath string `json:"filepath,omitempty"`
}

// =============================================================================
// Binding Methods
// =============================================================================

// GetRecordRefs returns summary references of all history records.
func (b *ComparisonBinding) GetRecordRefs() RecordRefsResult {
	ctx := context.Background()

	refs, err := b.comparisonUC.GetRecordRefs(ctx)
	if err != nil {
		slog.Error("GetRecordRefs failed", "error", err)
		return RecordRefsResult{
			Error: err.Error(),
		}
	}

	dtos := make([]ComparisonRecordRefDTO, 0, len(refs))
	for _, ref := range refs {
		dtos = append(dtos, *b.toRefDTO(ref))
	}

	return RecordRefsResult{
		Refs: dtos,
	}
}

// CompareRecords compares selected history records.
func (b *ComparisonBinding) CompareRecords(recordIDs []string) CompareResult {
	ctx := context.Background()

	if len(recordIDs) < 2 {
		return CompareResult{
			Success: false,
			Error:   "at least 2 records must be selected for comparison",
		}
	}

	result, err := b.comparisonUC.GenerateSimplifiedReport(ctx, recordIDs, comparison.GroupByThreads)
	if err != nil {
		slog.Error("CompareRecords failed", "error", err)
		return CompareResult{
			Error: err.Error(),
		}
	}

	// Convert to DTO
	groupDtos := make([]ConfigGroupDTO, 0, len(result.ConfigGroups))
	for _, group := range result.ConfigGroups {
		records := make([]ComparisonRecordDetailDTO, 0, len(group.Records))
		for _, rec := range group.Records {
			records = append(records, *b.toDetailDTO(rec))
		}
		groupDtos = append(groupDtos, ConfigGroupDTO{
			Threads:     group.Threads,
			N:           group.Statistics.N,
			Records:     records,
			TPSMean:     group.Statistics.TPS.Mean,
			TPSStdDev:   group.Statistics.TPS.StdDev,
			TPSMin:      group.Statistics.TPS.Min,
			TPSMax:      group.Statistics.TPS.Max,
			LatencyMean: group.Statistics.LatencyAvg.Mean,
			LatencyP95:  group.Statistics.LatencyP95.Mean,
		})
	}

	return CompareResult{
		Success: true,
		Result: &ComparisonResultDTO{
			ReportID:    result.ReportID,
			GeneratedAt: result.GeneratedAt.Format(time.RFC3339),
			GroupBy:     result.GroupBy,
			Groups:      groupDtos,
		},
	}
}

// ExportComparison exports comparison result as markdown or text.
func (b *ComparisonBinding) ExportComparison(reportID string, format string) ExportResult {
	ctx := context.Background()

	// Use runtime.SaveFileDialog to get save path
	savePath, err := runtime.SaveFileDialog(b.ctx, runtime.SaveDialogOptions{
		Title:           "Export Comparison",
		DefaultFilename: "comparison_" + reportID + ".md",
	})
	if err != nil {
		return ExportResult{Error: err.Error()}
	}

	if savePath == "" {
		return ExportResult{Error: "no file selected"}
	}

	// Get records for export (we need record IDs)
	refs, err := b.comparisonUC.GetRecordRefs(ctx)
	if err != nil {
		return ExportResult{Error: err.Error()}
	}

	// Generate content based on format
	var content string
	content = "# Comparison Report\n\nGenerated: " + time.Now().Format(time.RFC3339) + "\n"

	for _, ref := range refs {
		content += ref.TemplateName + " | TPS: " + fmt.Sprintf("%.2f", ref.TPS) + "\n"
	}

	// Write file
	if err := os.WriteFile(savePath, []byte(content), 0644); err != nil {
		return ExportResult{Error: err.Error()}
	}

	return ExportResult{Success: true, Filepath: savePath}
}

// =============================================================================
// Helper Methods
// =============================================================================

// toRefDTO converts a RecordRef to DTO.
func (b *ComparisonBinding) toRefDTO(ref *comparison.RecordRef) *ComparisonRecordRefDTO {
	return &ComparisonRecordRefDTO{
		ID:             ref.ID,
		TemplateName:   ref.TemplateName,
		DatabaseType:   ref.DatabaseType,
		Threads:        ref.Threads,
		ConnectionName: ref.ConnectionName,
		StartTime:      ref.StartTime.Format(time.RFC3339),
		TPS:            ref.TPS,
		LatencyAvg:     ref.LatencyAvg,
		Duration:       ref.Duration.String(),
	}
}

// toDetailDTO converts a RecordRef with details to DTO.
func (b *ComparisonBinding) toDetailDTO(ref *comparison.RecordRef) *ComparisonRecordDetailDTO {
	return &ComparisonRecordDetailDTO{
		ComparisonRecordRefDTO: ComparisonRecordRefDTO{
			ID:             ref.ID,
			TemplateName:   ref.TemplateName,
			DatabaseType:   ref.DatabaseType,
			Threads:        ref.Threads,
			ConnectionName: ref.ConnectionName,
			StartTime:      ref.StartTime.Format(time.RFC3339),
			TPS:            ref.TPS,
			LatencyAvg:     ref.LatencyAvg,
			Duration:       ref.Duration.String(),
		},
		QPS:          ref.QPS,
		ReadQueries:  ref.ReadQueries,
		WriteQueries: ref.WriteQueries,
		OtherQueries: ref.OtherQueries,
		TotalQueries: ref.TotalQueries,
	}
}
