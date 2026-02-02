// Package comparison provides result comparison functionality.
// This file implements grouping logic for multi-run configuration analysis.
package comparison

import (
	"fmt"
	"sort"
	"strings"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

// GroupRecordsByConfig groups history records by configuration.
// It identifies runs with the same configuration within the time window
// and returns ConfigGroup objects with aggregated statistics.
//
// Parameters:
//   - records: History records to group
//   - groupBy: Primary grouping dimension (threads, database_type, etc.)
//   - config: Similarity detection configuration (time window, etc.)
//
// Returns:
//   - []*ConfigGroup: Groups sorted by the primary dimension (e.g., threads)
//   - error: If grouping fails
func GroupRecordsByConfig(
	records []*history.Record,
	groupBy GroupByField,
	config *SimilarityConfig,
) ([]*ConfigGroup, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records to group")
	}

	// Use default config if not provided
	if config == nil {
		config = DefaultSimilarityConfig()
		config.GroupBy = groupBy
	}

	// Group records by config spec
	groupsMap := make(map[string]*ConfigGroup)

	for _, record := range records {
		// Create config spec for this record
		spec := createConfigSpecFromRecord(record, config)

		// Create group key
		groupKey := configSpecToString(spec)

		// Create or get group
		group, exists := groupsMap[groupKey]
		if !exists {
			group = &ConfigGroup{
				Config:  spec,
				Runs:    []*Run{},
				GroupID: "", // Will be assigned later
			}
			groupsMap[groupKey] = group
		}

		// Convert record to run and add to group
		run := convertRecordToRun(record)
		group.Runs = append(group.Runs, run)
	}

	// Convert map to slice
	var groups []*ConfigGroup
	for _, group := range groupsMap {
		groups = append(groups, group)
	}

	// Sort groups by the primary dimension
	sortGroups(groups, groupBy)

	// Assign group IDs (C1, C2, C3...)
	for i, group := range groups {
		group.GroupID = fmt.Sprintf("C%d", i+1)
	}

	// Calculate statistics for each group
	for _, group := range groups {
		stats := CalculateRunStats(group.Runs)
		group.Statistics = stats
	}

	return groups, nil
}

// convertRecordToRun converts a single history record to a Run object.
func convertRecordToRun(record *history.Record) *Run {
	run := &Run{
		RunID:      record.ID,
		StartTime:  record.StartTime,
		Duration:   record.Duration,

		TPS:        record.TPSCalculated,
		LatencyAvg: record.LatencyAvg,
		LatencyMin: record.LatencyMin,
		LatencyMax: record.LatencyMax,
		LatencyP95: record.LatencyP95,
		LatencyP99: record.LatencyP99,

		ReadQueries:        record.ReadQueries,
		WriteQueries:       record.WriteQueries,
		OtherQueries:       record.OtherQueries,
		TotalQueries:       record.TotalQueries,
		TotalTransactions:  record.TotalTransactions,

		Errors:      record.IgnoredErrors,
		Reconnects:  record.Reconnects,

		TotalTime:   record.TotalTime,
		TotalEvents: record.TotalEvents,
	}

	// Calculate QPS
	if record.TotalQueries > 0 && record.Duration.Seconds() > 0 {
		run.QPS = float64(record.TotalQueries) / record.Duration.Seconds()
	}

	// Calculate queries per transaction
	if record.TotalTransactions > 0 && record.TotalQueries > 0 {
		run.QueriesPerTransaction = float64(record.TotalQueries) / float64(record.TotalTransactions)
	}

	return run
}

// createConfigSpecFromRecord creates a ConfigSpec from a history record.
func createConfigSpecFromRecord(record *history.Record, config *SimilarityConfig) ConfigSpec {
	spec := ConfigSpec{
		Threads:        record.Threads,
		DatabaseType:   record.DatabaseType,
		TemplateName:   record.TemplateName,
		ConnectionName: record.ConnectionName,
	}

	// If not considering connection, clear it
	if !config.ConsiderConnection {
		spec.ConnectionName = ""
	}

	return spec
}

// configSpecToString creates a string key for ConfigSpec.
func configSpecToString(spec ConfigSpec) string {
	parts := []string{
		fmt.Sprintf("threads=%d", spec.Threads),
		fmt.Sprintf("db=%s", spec.DatabaseType),
		fmt.Sprintf("template=%s", spec.TemplateName),
	}
	if spec.ConnectionName != "" {
		parts = append(parts, fmt.Sprintf("conn=%s", spec.ConnectionName))
	}
	return strings.Join(parts, "|")
}

// sortGroups sorts config groups by the primary dimension.
func sortGroups(groups []*ConfigGroup, groupBy GroupByField) {
	switch groupBy {
	case GroupByThreads:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Config.Threads < groups[j].Config.Threads
		})
	case GroupByDatabaseType:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Config.DatabaseType < groups[j].Config.DatabaseType
		})
	case GroupByTemplate:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Config.TemplateName < groups[j].Config.TemplateName
		})
	case GroupByConnection:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Config.ConnectionName < groups[j].Config.ConnectionName
		})
	default:
		// Default to threads
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Config.Threads < groups[j].Config.Threads
		})
	}
}

// convertRecordsToRuns converts history records to Run objects.
// DEPRECATED: Use convertRecordToRun for single record conversion.
func convertRecordsToRuns(records []*history.Record) []*Run {
	runs := make([]*Run, 0, len(records))

	for _, record := range records {
		run := convertRecordToRun(record)
		runs = append(runs, run)
	}

	return runs
}

