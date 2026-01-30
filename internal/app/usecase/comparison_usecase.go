// Package usecase provides comparison business logic.
package usecase

import (
	"context"
	"fmt"
	"log/slog"

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
