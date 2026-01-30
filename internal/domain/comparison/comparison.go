// Package comparison provides result comparison functionality.
// Implements: Phase 5 - Result comparison and analysis
package comparison

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

// GroupByField defines how to group comparison results.
type GroupByField string

const (
	// GroupByThreads groups results by thread count.
	GroupByThreads GroupByField = "threads"
	// GroupByDatabaseType groups results by database type.
	GroupByDatabaseType GroupByField = "database_type"
	// GroupByTemplate groups results by template name.
	GroupByTemplate GroupByField = "template"
	// GroupByDate groups results by date.
	GroupByDate GroupByField = "date"
)

// RecordRef is a reference to a history record with summary info.
type RecordRef struct {
	ID             string        `json:"id"`
	TemplateName   string        `json:"template_name"`
	DatabaseType   string        `json:"database_type"`
	Threads        int           `json:"threads"`
	ConnectionName string        `json:"connection_name"`
	StartTime      time.Time     `json:"start_time"`
	TPS            float64       `json:"tps"`
	LatencyAvg     float64       `json:"latency_avg_ms"`
	LatencyMin     float64       `json:"latency_min_ms"`
	LatencyMax     float64       `json:"latency_max_ms"`
	LatencyP95     float64       `json:"latency_p95_ms"`
	LatencyP99     float64       `json:"latency_p99_ms"`
	Duration       time.Duration `json:"duration"`
	QPS            float64       `json:"qps,omitempty"`
	ReadQueries    int64         `json:"read_queries,omitempty"`
	WriteQueries   int64         `json:"write_queries,omitempty"`
	OtherQueries   int64         `json:"other_queries,omitempty"`
}

// MetricStats contains statistical information about metrics.
type MetricStats struct {
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	Avg     float64   `json:"avg"`
	StdDev  float64   `json:"std_dev"`
	Median  float64   `json:"median,omitempty"`
	RunID   string    `json:"run_id,omitempty"`
	RunName string    `json:"run_name,omitempty"`
	Values  []float64 `json:"values,omitempty"`
	Labels  []string  `json:"labels,omitempty"`
}

// MultiConfigComparison represents a comparison of multiple configurations.
// This is the main structure for horizontal comparison across different configurations.
type MultiConfigComparison struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	CreatedAt     time.Time       `json:"created_at"`
	Records       []*RecordRef    `json:"records"`
	GroupBy       GroupByField    `json:"group_by"`
	GeneratedAt   time.Time       `json:"generated_at"`

	// Comparison results
	TPSComparison   *MetricStats   `json:"tps_comparison"`
	LatencyCompare  *LatencyStats  `json:"latency_comparison"`
	QPSComparison   *MetricStats   `json:"qps_comparison"`
	ReadWriteRatio  *ReadWriteRatio `json:"read_write_ratio"`
}

// LatencyStats contains detailed latency statistics.
type LatencyStats struct {
	Avg *MetricStats `json:"avg"`
	Min *MetricStats `json:"min"`
	Max *MetricStats `json:"max"`
	P95 *MetricStats `json:"p95"`
	P99 *MetricStats `json:"p99"`
}

// ReadWriteRatio represents read/write query distribution.
type ReadWriteRatio struct {
	ReadQueries  int64   `json:"read_queries"`
	WriteQueries int64   `json:"write_queries"`
	OtherQueries int64   `json:"other_queries"`
	ReadPct      float64 `json:"read_pct"`
	WritePct     float64 `json:"write_pct"`
	OtherPct     float64 `json:"other_pct"`
}

// CompareMultiConfig performs multi-config comparison on history records.
func CompareMultiConfig(records []*history.Record, groupBy GroupByField) (*MultiConfigComparison, error) {
	if len(records) < 2 {
		return nil, fmt.Errorf("at least 2 records are required for comparison")
	}

	// Sort records by group field
	sortedRecords := make([]*history.Record, len(records))
	copy(sortedRecords, records)

	switch groupBy {
	case GroupByThreads:
		sort.Slice(sortedRecords, func(i, j int) bool {
			return sortedRecords[i].Threads < sortedRecords[j].Threads
		})
	case GroupByDatabaseType:
		sort.Slice(sortedRecords, func(i, j int) bool {
			return sortedRecords[i].DatabaseType < sortedRecords[j].DatabaseType
		})
	case GroupByTemplate:
		sort.Slice(sortedRecords, func(i, j int) bool {
			return sortedRecords[i].TemplateName < sortedRecords[j].TemplateName
		})
	case GroupByDate:
		sort.Slice(sortedRecords, func(i, j int) bool {
			return sortedRecords[i].StartTime.Before(sortedRecords[j].StartTime)
		})
	}

	// Create record references
	refs := make([]*RecordRef, len(sortedRecords))
	for i, record := range sortedRecords {
		durationSec := record.Duration.Seconds()
		qps := 0.0
		if durationSec > 0 && record.TotalQueries > 0 {
			qps = float64(record.TotalQueries) / durationSec
		}

		refs[i] = &RecordRef{
			ID:             record.ID,
			TemplateName:   record.TemplateName,
			DatabaseType:   record.DatabaseType,
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
			ReadQueries:    record.ReadQueries,
			WriteQueries:   record.WriteQueries,
			OtherQueries:   record.OtherQueries,
		}
	}

	// Calculate TPS comparison
	tpsStats := calculateMetricStats(refs, func(r *RecordRef) float64 {
		return r.TPS
	})

	// Calculate latency comparison
	latencyStats := &LatencyStats{
		Avg: calculateMetricStats(refs, func(r *RecordRef) float64 { return r.LatencyAvg }),
		Min: calculateMetricStats(refs, func(r *RecordRef) float64 { return r.LatencyMin }),
		Max: calculateMetricStats(refs, func(r *RecordRef) float64 { return r.LatencyMax }),
		P95: calculateMetricStats(refs, func(r *RecordRef) float64 { return r.LatencyP95 }),
		P99: calculateMetricStats(refs, func(r *RecordRef) float64 { return r.LatencyP99 }),
	}

	// Calculate QPS comparison
	qpsStats := calculateMetricStats(refs, func(r *RecordRef) float64 {
		return r.QPS
	})

	// Calculate read/write ratio
	rwRatio := &ReadWriteRatio{}
	for _, ref := range refs {
		rwRatio.ReadQueries += ref.ReadQueries
		rwRatio.WriteQueries += ref.WriteQueries
		rwRatio.OtherQueries += ref.OtherQueries
	}
	totalQueries := rwRatio.ReadQueries + rwRatio.WriteQueries + rwRatio.OtherQueries
	if totalQueries > 0 {
		rwRatio.ReadPct = float64(rwRatio.ReadQueries) / float64(totalQueries) * 100
		rwRatio.WritePct = float64(rwRatio.WriteQueries) / float64(totalQueries) * 100
		rwRatio.OtherPct = float64(rwRatio.OtherQueries) / float64(totalQueries) * 100
	}

	return &MultiConfigComparison{
		ID:              generateComparisonID(),
		Name:            fmt.Sprintf("Comparison - %d records", len(records)),
		CreatedAt:       time.Now(),
		Records:         refs,
		GroupBy:         groupBy,
		GeneratedAt:     time.Now(),
		TPSComparison:   tpsStats,
		LatencyCompare:  latencyStats,
		QPSComparison:   qpsStats,
		ReadWriteRatio:  rwRatio,
	}, nil
}

// calculateMetricStats calculates statistics for a metric across all records.
func calculateMetricStats(records []*RecordRef, extractor func(*RecordRef) float64) *MetricStats {
	if len(records) == 0 {
		return nil
	}

	values := make([]float64, len(records))
	labels := make([]string, len(records))
	for i, record := range records {
		values[i] = extractor(record)
		labels[i] = fmt.Sprintf("%s (%d threads)", record.DatabaseType, record.Threads)
	}

	sort.Float64s(values)
	min := values[0]
	max := values[len(values)-1]

	// Calculate average
	var sum float64
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))

	// Calculate standard deviation
	var varianceSum float64
	for _, v := range values {
		diff := v - avg
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(len(values)))

	return &MetricStats{
		Min:     min,
		Max:     max,
		Avg:     avg,
		StdDev:  stdDev,
		Values:  values,
		Labels:  labels,
	}
}

// FormatTable formats the comparison result as a table.
func (c *MultiConfigComparison) FormatTable() string {
	var builder strings.Builder

	builder.WriteString("╔════════════════════════════════════════════════════════════════════════════╗\n")
	builder.WriteString("║                      Multi-Configuration Comparison Results                         ║\n")
	builder.WriteString("╠════════════════════════════════════════════════════════════════════════════╣\n")
	builder.WriteString(fmt.Sprintf("║ Generated: %s                                                                  ║\n", c.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString("╠════════════════════════════════════════════════════════════════════════════╣\n\n")

	// Summary
	builder.WriteString("## Summary\n\n")
	builder.WriteString(fmt.Sprintf("Total Records: %d\n", len(c.Records)))
	builder.WriteString(fmt.Sprintf("Group By: %s\n\n", c.GroupBy))

	// TPS Comparison
	if c.TPSComparison != nil {
		builder.WriteString("## TPS Comparison (Transactions Per Second)\n\n")
		formatMetricTable(&builder, "Configuration", c.TPSComparison)
	}

	// Latency Comparison
	if c.LatencyCompare != nil {
		builder.WriteString("\n## Latency Comparison (ms)\n\n")
		if c.LatencyCompare.Avg != nil {
			formatMetricTable(&builder, "Configuration (Avg Latency)", c.LatencyCompare.Avg)
		}
	}

	// QPS Comparison
	if c.QPSComparison != nil {
		builder.WriteString("\n## QPS Comparison (Queries Per Second)\n\n")
		formatMetricTable(&builder, "Configuration", c.QPSComparison)
	}

	// Read/Write Ratio
	if c.ReadWriteRatio != nil {
		builder.WriteString("\n## Query Distribution\n\n")
		builder.WriteString(fmt.Sprintf("  Read:  %d queries (%.1f%%)\n", c.ReadWriteRatio.ReadQueries, c.ReadWriteRatio.ReadPct))
		builder.WriteString(fmt.Sprintf("  Write: %d queries (%.1f%%)\n", c.ReadWriteRatio.WriteQueries, c.ReadWriteRatio.WritePct))
		if c.ReadWriteRatio.OtherQueries > 0 {
			builder.WriteString(fmt.Sprintf("  Other: %d queries (%.1f%%)\n", c.ReadWriteRatio.OtherQueries, c.ReadWriteRatio.OtherPct))
		}
	}

	return builder.String()
}

// formatMetricTable formats a metric stats as a table.
func formatMetricTable(builder *strings.Builder, label string, stats *MetricStats) {
	builder.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	builder.WriteString(fmt.Sprintf("│ %-65s │\n", label))
	builder.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))
	builder.WriteString(fmt.Sprintf("│ %-20s │ %10s │ %10s │ %10s │ %10s │\n", "Config", "Min", "Avg", "Max", "StdDev"))
	builder.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	for i := 0; i < len(stats.Values) && i < len(stats.Labels); i++ {
		label := stats.Labels[i]
		if len(label) > 20 {
			label = label[:17] + "..."
		}
		builder.WriteString(fmt.Sprintf("│ %-20s │ %10.2f │ %10.2f │ %10.2f │ %10.2f │\n",
			label, stats.Values[i], stats.Avg, stats.Max, stats.StdDev))
	}

	builder.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))
}

// FormatBarChart formats a simple ASCII bar chart.
func (c *MultiConfigComparison) FormatBarChart(metric string) string {
	var builder strings.Builder

	var stats *MetricStats
	switch metric {
	case "TPS":
		stats = c.TPSComparison
	case "QPS":
		stats = c.QPSComparison
	default:
		return "Unknown metric"
	}

	if stats == nil || len(stats.Values) == 0 {
		return "No data available"
	}

	builder.WriteString(fmt.Sprintf("\n## %s Bar Chart\n\n", metric))

	// Find max for scaling
	max := stats.Max
	if max == 0 {
		max = 1
	}

	// Calculate bar width (max 50 chars)
	barWidth := 50
	for i, val := range stats.Values {
		label := stats.Labels[i]
		if len(label) > 20 {
			label = label[:17] + "..."
		}

		// Calculate bar length
		length := int((val / max) * float64(barWidth))
		if length < 1 {
			length = 1
		}
		if length > barWidth {
			length = barWidth
		}

		bar := strings.Repeat("█", length)
		spaces := strings.Repeat(" ", barWidth-length)

		builder.WriteString(fmt.Sprintf("%-20s │%s %10.2f\n", label, bar+spaces, val))
	}

	return builder.String()
}

func generateComparisonID() string {
	return fmt.Sprintf("cmp-%d", time.Now().UnixNano())
}
