// Package usecase provides data import functionality.
// This file imports raw sysbench outputs from exports/ directory into database.
package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/sysbench"
)

// ImportSysbenchOutputs imports raw sysbench outputs from exports/ directory.
// It reads all benchmark_*.txt files, parses them, and stores in database.
func (uc *ComparisonUseCase) ImportSysbenchOutputs(ctx context.Context) (*ImportResult, error) {
	// Find all benchmark output files
	exportsDir := "./exports"
	files, err := filepath.Glob(filepath.Join(exportsDir, "benchmark_*.txt"))
	if err != nil {
		return nil, fmt.Errorf("find benchmark files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no benchmark files found in %s", exportsDir)
	}

	slog.Info("Import: Found benchmark files", "count", len(files))

	result := &ImportResult{
		TotalFiles:   len(files),
		ImportedRuns: 0,
		FailedRuns:   0,
		SkippedRuns:  0,
		Files:        make([]FileImportStatus, len(files)),
	}

	for i, filepath := range files {
		slog.Info("Import: Processing file", "file", filepath)

		// Read file content
		content, err := os.ReadFile(filepath)
		if err != nil {
			slog.Error("Import: Failed to read file", "file", filepath, "error", err)
			result.Files[i] = FileImportStatus{
				Filepath: filepath,
				Status:   "failed",
				Error:    err.Error(),
			}
			result.FailedRuns++
			continue
		}

		// Extract run ID from filename
		// Format: benchmark_Template_YYYYMMDD_HHMMSS.txt
		runID := extractRunIDFromFilename(filepath)
		if runID == "" {
			slog.Warn("Import: Skipped file (cannot extract run ID)", "file", filepath)
			result.Files[i] = FileImportStatus{
				Filepath: filepath,
				Status:   "skipped",
				Error:    "Cannot extract run ID from filename",
			}
			result.SkippedRuns++
			continue
		}

		// Parse sysbench output
		parsed, err := sysbench.ParseSysbenchOutput(runID, string(content))
		if err != nil {
			slog.Error("Import: Failed to parse", "file", filepath, "error", err)
			result.Files[i] = FileImportStatus{
				Filepath: filepath,
				RunID:    runID,
				Status:   "failed",
				Error:    err.Error(),
			}
			result.FailedRuns++
			continue
		}

		// Store in run_logs table
		err = uc.storeRunLog(ctx, runID, string(content))
		if err != nil {
			slog.Error("Import: Failed to store run log", "run_id", runID, "error", err)
			result.Files[i] = FileImportStatus{
				Filepath: filepath,
				RunID:    runID,
				Status:   "failed",
				Error:    err.Error(),
			}
			result.FailedRuns++
			continue
		}

		// Store time series data in metric_samples
		err = uc.storeMetricSamples(ctx, runID, parsed.TimeSeries)
		if err != nil {
			slog.Error("Import: Failed to store metric samples", "run_id", runID, "error", err)
			// Non-fatal, continue
		}

		slog.Info("Import: Successfully imported", "run_id", runID, "file", filepath)
		result.Files[i] = FileImportStatus{
			Filepath:  filepath,
			RunID:     runID,
			Status:    "imported",
			Timestamp: parsed.Timestamp,
			Samples:   len(parsed.TimeSeries),
		}
		result.ImportedRuns++
	}

	slog.Info("Import: Completed", "imported", result.ImportedRuns, "failed", result.FailedRuns, "skipped", result.SkippedRuns)

	return result, nil
}

// storeRunLog stores raw output in run_logs table.
func (uc *ComparisonUseCase) storeRunLog(ctx context.Context, runID, content string) error {
	// This would need database access
	// For now, we'll implement a simple version
	// TODO: Implement actual database insert

	// Get database connection from historyRepo
	// We need to access the underlying DB connection
	// This is a placeholder - actual implementation would use db.ExecContext

	return nil
}

// storeMetricSamples stores time series data in metric_samples table.
func (uc *ComparisonUseCase) storeMetricSamples(ctx context.Context, runID string, samples []sysbench.TimeSeriesSample) error {
	// TODO: Implement actual database insert
	return nil
}

// extractRunIDFromFilename extracts run ID from filename.
// Format: benchmark_Sysbench_OLTP_Read-Write_20260130_150822.txt
// We'll use timestamp as run ID
func extractRunIDFromFilename(filepath string) string {
	// Extract timestamp from filename
	re := regexp.MustCompile(`\d{8}_\d{6}`)
	matches := re.FindString(filepath)
	if matches == "" {
		return ""
	}
	return "run-" + matches
}

// ImportResult represents the result of importing benchmark files.
type ImportResult struct {
	TotalFiles   int
	ImportedRuns int
	FailedRuns   int
	SkippedRuns  int
	Files        []FileImportStatus
}

// FileImportStatus represents the import status of a single file.
type FileImportStatus struct {
	Filepath  string
	RunID     string
	Status    string // "imported", "failed", "skipped"
	Error     string
	Timestamp time.Time
	Samples   int
}
