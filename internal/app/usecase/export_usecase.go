// Package usecase provides export business logic.
package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

// ExportFormat represents the export format type.
type ExportFormat string

const (
	FormatTXT      ExportFormat = "txt"
	FormatMarkdown ExportFormat = "markdown"
)

// ExportUseCase provides export business logic.
type ExportUseCase struct {
	exportDir string // Default export directory
}

// NewExportUseCase creates a new export use case.
func NewExportUseCase(exportDir string) *ExportUseCase {
	if exportDir == "" {
		exportDir = "./exports"
	}
	return &ExportUseCase{
		exportDir: exportDir,
	}
}

// ExportRecord exports a single history record to the specified format.
func (uc *ExportUseCase) ExportRecord(ctx context.Context, record *history.Record, format ExportFormat) (string, error) {
	// Ensure export directory exists
	if err := os.MkdirAll(uc.exportDir, 0755); err != nil {
		return "", fmt.Errorf("create export directory: %w", err)
	}

	// Generate filename
	filename := uc.generateFilename(record, format)
	filepath := filepath.Join(uc.exportDir, filename)

	// Export based on format
	switch format {
	case FormatTXT:
		if err := uc.exportToTXT(record, filepath); err != nil {
			return "", err
		}
	case FormatMarkdown:
		if err := uc.exportToMarkdown(record, filepath); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	return filepath, nil
}

// ExportAllRecords exports all history records to the specified format.
// Returns the count of successfully exported records and the directory path.
func (uc *ExportUseCase) ExportAllRecords(ctx context.Context, records []*history.Record, format ExportFormat) (int, string, error) {
	if len(records) == 0 {
		return 0, "", fmt.Errorf("no records to export")
	}

	// Ensure export directory exists
	if err := os.MkdirAll(uc.exportDir, 0755); err != nil {
		return 0, "", fmt.Errorf("create export directory: %w", err)
	}

	successCount := 0
	failedRecords := []string{}

	for i, record := range records {
		// Generate filename for this record
		filename := uc.generateFilename(record, format)
		filepath := filepath.Join(uc.exportDir, filename)

		// Export based on format
		var err error
		switch format {
		case FormatTXT:
			err = uc.exportToTXT(record, filepath)
		case FormatMarkdown:
			err = uc.exportToMarkdown(record, filepath)
		default:
			err = fmt.Errorf("unsupported format: %s", format)
		}

		if err != nil {
			slog.Error("Failed to export record", "index", i, "id", record.ID, "error", err)
			failedRecords = append(failedRecords, record.ID)
		} else {
			successCount++
		}
	}

	if len(failedRecords) > 0 {
		return successCount, uc.exportDir, fmt.Errorf("failed to export %d records: %v", len(failedRecords), failedRecords)
	}

	return successCount, uc.exportDir, nil
}

// generateFilename generates a filename for the exported record.
func (uc *ExportUseCase) generateFilename(record *history.Record, format ExportFormat) string {
	// Format: benchmark_{template_name}_{timestamp}.{ext}
	templateName := strings.ReplaceAll(record.TemplateName, " ", "_")
	templateName = strings.ReplaceAll(templateName, "/", "_")
	timestamp := record.StartTime.Format("20060102_150405")

	ext := string(format)
	if format == FormatMarkdown {
		ext = "md"
	}

	return fmt.Sprintf("benchmark_%s_%s.%s", templateName, timestamp, ext)
}

// exportToTXT exports record to plain text format (exact sysbench format).
func (uc *ExportUseCase) exportToTXT(record *history.Record, filepath string) error {
	var builder strings.Builder

	// Build sysbench-style output
	builder.WriteString(fmt.Sprintf("sysbench 1.0.20 (using bundled LuaJIT 2.1.0-beta3)\n\n"))
	builder.WriteString(fmt.Sprintf("Running the test with following options:\n"))
	builder.WriteString(fmt.Sprintf("Number of threads: %d\n", record.Threads))
	builder.WriteString(fmt.Sprintf("Initializing random number generator from current time\n\n"))
	builder.WriteString(fmt.Sprintf("\nInitializing worker threads...\n\n"))
	builder.WriteString(fmt.Sprintf("Threads started!\n\n"))

	// Build time series data (intermediate results)
	if len(record.TimeSeries) > 0 {
		for _, sample := range record.TimeSeries {
			if sample.Phase == "run" {
				// Format: [ 1s ] thds: 4 tps: 341.28 qps: 6871.52 (r/w/o: 4817.85/1367.12/686.55) lat (ms,95%): 13.46 err/s: 0.00 reconn/s: 0.00
				second := int(sample.Timestamp.Sub(record.StartTime).Seconds())
				builder.WriteString(fmt.Sprintf("[%3ds ] thds: %d tps: %.2f qps: %.2f lat (ms,95%%): %.2f err/s: %.2f reconn/s: %.2f\n",
					second,
					record.Threads,
					sample.TPS,
					sample.QPS,
					sample.LatencyP95,
					sample.ErrorRate,
					0.0, // reconnects per second - not in time series
				))
			}
		}
		builder.WriteString("\n")
	}

	// SQL statistics
	builder.WriteString(fmt.Sprintf("SQL statistics:\n"))
	builder.WriteString(fmt.Sprintf("    queries performed:\n"))
	builder.WriteString(fmt.Sprintf("        read:                            %d\n", record.ReadQueries))
	builder.WriteString(fmt.Sprintf("        write:                           %d\n", record.WriteQueries))
	builder.WriteString(fmt.Sprintf("        other:                           %d\n", record.OtherQueries))
	builder.WriteString(fmt.Sprintf("        total:                           %d\n", record.TotalQueries))

	durationSec := record.Duration.Seconds()
	ignoredErrorsPerSec := 0.0
	if durationSec > 0 {
		ignoredErrorsPerSec = float64(record.IgnoredErrors) / durationSec
	}
	reconnectsPerSec := 0.0
	if durationSec > 0 {
		reconnectsPerSec = float64(record.Reconnects) / durationSec
	}

	builder.WriteString(fmt.Sprintf("    transactions:                        %d  (%.2f per sec.)\n", record.TotalTransactions, record.TPSCalculated))
	builder.WriteString(fmt.Sprintf("    queries:                             %d (%.2f per sec.)\n", record.TotalQueries, float64(record.TotalQueries)/durationSec))
	builder.WriteString(fmt.Sprintf("    ignored errors:                      %d      (%.2f per sec.)\n", record.IgnoredErrors, ignoredErrorsPerSec))
	builder.WriteString(fmt.Sprintf("    reconnects:                          %d      (%.2f per sec.)\n\n", record.Reconnects, reconnectsPerSec))

	// General statistics
	builder.WriteString(fmt.Sprintf("General statistics:\n"))
	builder.WriteString(fmt.Sprintf("    total time:                          %.4fs\n", record.TotalTime))
	builder.WriteString(fmt.Sprintf("    total number of events:              %d\n\n", record.TotalEvents))

	// Latency statistics
	builder.WriteString(fmt.Sprintf("Latency (ms):\n"))
	builder.WriteString(fmt.Sprintf("         min:                                    %.2f\n", record.LatencyMin))
	builder.WriteString(fmt.Sprintf("         avg:                                   %.2f\n", record.LatencyAvg))
	builder.WriteString(fmt.Sprintf("         max:                                   %.2f\n", record.LatencyMax))
	builder.WriteString(fmt.Sprintf("         95th percentile:                       %.2f\n", record.LatencyP95))
	if record.LatencySum > 0 {
		builder.WriteString(fmt.Sprintf("         sum:                                %.2f\n", record.LatencySum))
	}
	builder.WriteString("\n")

	// Threads fairness
	builder.WriteString(fmt.Sprintf("Threads fairness:\n"))
	builder.WriteString(fmt.Sprintf("    events (avg/stddev):           %.4f/%.2f\n", record.EventsAvg, record.EventsStddev))
	builder.WriteString(fmt.Sprintf("    execution time (avg/stddev):   %.4f/%.2f\n", record.ExecTimeAvg, record.ExecTimeStddev))
	builder.WriteString("\n")

	// Write to file
	if err := os.WriteFile(filepath, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// exportToMarkdown exports record to Markdown format.
func (uc *ExportUseCase) exportToMarkdown(record *history.Record, filepath string) error {
	var builder strings.Builder

	// Build header
	builder.WriteString("# DB-BenchMind Benchmark Results\n\n")
	builder.WriteString("## Run Information\n\n")
	builder.WriteString("| Field | Value |\n")
	builder.WriteString("|-------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Run ID | `%s` |\n", record.ID))
	builder.WriteString(fmt.Sprintf("| Connection | %s |\n", record.ConnectionName))
	builder.WriteString(fmt.Sprintf("| Template | %s |\n", record.TemplateName))
	builder.WriteString(fmt.Sprintf("| Database Type | %s |\n", record.DatabaseType))
	builder.WriteString(fmt.Sprintf("| Threads | %d |\n", record.Threads))
	builder.WriteString(fmt.Sprintf("| Start Time | %s |\n", record.StartTime.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("| Duration | %s |\n", record.Duration))
	builder.WriteString("\n")

	// Build core metrics
	builder.WriteString("## Core Metrics\n\n")
	builder.WriteString("| Metric | Value |\n")
	builder.WriteString("|--------|-------|\n")
	builder.WriteString(fmt.Sprintf("| **TPS** | **%.2f** |\n", record.TPSCalculated))
	builder.WriteString(fmt.Sprintf("| Latency Avg | %.2f ms |\n", record.LatencyAvg))
	builder.WriteString(fmt.Sprintf("| Latency Min | %.2f ms |\n", record.LatencyMin))
	builder.WriteString(fmt.Sprintf("| Latency Max | %.2f ms |\n", record.LatencyMax))
	builder.WriteString(fmt.Sprintf("| Latency P95 | %.2f ms |\n", record.LatencyP95))
	if record.LatencySum > 0 {
		builder.WriteString(fmt.Sprintf("| Latency Sum | %.2f ms |\n", record.LatencySum))
	}
	builder.WriteString("\n")

	// Build SQL statistics
	builder.WriteString("## SQL Statistics\n\n")
	builder.WriteString("| Category | Count |\n")
	builder.WriteString("|----------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Read Queries | %d |\n", record.ReadQueries))
	builder.WriteString(fmt.Sprintf("| Write Queries | %d |\n", record.WriteQueries))
	builder.WriteString(fmt.Sprintf("| Other Queries | %d |\n", record.OtherQueries))
	builder.WriteString(fmt.Sprintf("| **Total Queries** | **%d** |\n", record.TotalQueries))
	builder.WriteString(fmt.Sprintf("| **Total Transactions** | **%d** |\n", record.TotalTransactions))
	builder.WriteString(fmt.Sprintf("| Ignored Errors | %d |\n", record.IgnoredErrors))
	builder.WriteString(fmt.Sprintf("| Reconnects | %d |\n", record.Reconnects))
	builder.WriteString("\n")

	durationSec := record.Duration.Seconds()
	qps := 0.0
	if durationSec > 0 && record.TotalQueries > 0 {
		qps = float64(record.TotalQueries) / durationSec
	}

	builder.WriteString("**Rates:**\n")
	builder.WriteString(fmt.Sprintf("- Transactions: %.2f/sec\n", record.TPSCalculated))
	builder.WriteString(fmt.Sprintf("- Queries: %.2f/sec\n", qps))
	builder.WriteString("\n")

	// Build general statistics
	builder.WriteString("## General Statistics\n\n")
	builder.WriteString("| Metric | Value |\n")
	builder.WriteString("|--------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Total Time | %.4f s |\n", record.TotalTime))
	builder.WriteString(fmt.Sprintf("| Total Events | %d |\n", record.TotalEvents))
	builder.WriteString("\n")

	// Build threads fairness
	builder.WriteString("## Threads Fairness\n\n")
	builder.WriteString("| Metric | Avg | Stddev |\n")
	builder.WriteString("|--------|-----|--------|\n")
	builder.WriteString(fmt.Sprintf("| Events | %.4f | %.2f |\n", record.EventsAvg, record.EventsStddev))
	builder.WriteString(fmt.Sprintf("| Execution Time | %.4f | %.2f |\n", record.ExecTimeAvg, record.ExecTimeStddev))
	builder.WriteString("\n")

	// Build time series if available
	if len(record.TimeSeries) > 0 {
		builder.WriteString("## Time Series Data\n\n")
		builder.WriteString(fmt.Sprintf("**Total samples:** %d\n\n", len(record.TimeSeries)))

		// Count run phase samples
		runSamples := 0
		for _, s := range record.TimeSeries {
			if s.Phase == "run" {
				runSamples++
			}
		}

		builder.WriteString(fmt.Sprintf("**Run phase samples:** %d\n\n", runSamples))

		// Show first 20 samples
		displayCount := 20
		if runSamples < displayCount {
			displayCount = runSamples
		}

		builder.WriteString(fmt.Sprintf("### First %d Samples\n\n", displayCount))
		builder.WriteString("| Time | TPS | QPS | Latency P95 (ms) | Error Rate (%) |\n")
		builder.WriteString("|------|-----|-----|------------------|---------------|\n")

		count := 0
		for _, sample := range record.TimeSeries {
			if sample.Phase == "run" {
				second := int(sample.Timestamp.Sub(record.StartTime).Seconds())
				builder.WriteString(fmt.Sprintf("| [%3ds] | %.2f | %.2f | %.2f | %.2f |\n",
					second, sample.TPS, sample.QPS, sample.LatencyP95, sample.ErrorRate))
				count++
				if count >= displayCount {
					break
				}
			}
		}

		if runSamples > displayCount {
			builder.WriteString("\n...")
			builder.WriteString(fmt.Sprintf("\n... (%d samples omitted) ...\n\n", runSamples-displayCount))

			// Show last 10 samples
			builder.WriteString("### Last 10 Samples\n\n")
			builder.WriteString("| Time | TPS | QPS | Latency P95 (ms) | Error Rate (%) |\n")
			builder.WriteString("|------|-----|-----|------------------|---------------|\n")

			shown := 0
			for i := len(record.TimeSeries) - 1; i >= 0; i-- {
				sample := record.TimeSeries[i]
				if sample.Phase == "run" {
					second := int(sample.Timestamp.Sub(record.StartTime).Seconds())
					builder.WriteString(fmt.Sprintf("| [%3ds] | %.2f | %.2f | %.2f | %.2f |\n",
						second, sample.TPS, sample.QPS, sample.LatencyP95, sample.ErrorRate))
					shown++
					if shown >= 10 {
						break
					}
				}
			}
		}
		builder.WriteString("\n")
	}

	// Write to file
	if err := os.WriteFile(filepath, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
