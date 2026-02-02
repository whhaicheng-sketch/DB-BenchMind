// Test program for simplified report generation
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
)

func main() {
	// Initialize logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("Starting simplified report test")

	// Initialize database
	db, err := database.InitializeSQLite(context.Background(), "./data/db-benchmind.db")
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("Database initialized")

	// Initialize repositories
	historyRepo := repository.NewSQLiteHistoryRepository(db)
	runRepo := repository.NewSQLiteRunRepository(db)
	slog.Info("Repositories initialized")

	// Initialize use case
	comparisonUC := usecase.NewComparisonUseCase(historyRepo, runRepo)
	slog.Info("Use case initialized")

	// Get all records
	ctx := context.Background()
	refs, err := comparisonUC.GetRecordRefs(ctx)
	if err != nil {
		slog.Error("Failed to get records", "error", err)
		os.Exit(1)
	}

	if len(refs) < 2 {
		slog.Error("Need at least 2 records for comparison", "count", len(refs))
		os.Exit(1)
	}

	slog.Info("Records loaded", "count", len(refs))

	// Print record summary
	for i, ref := range refs {
		slog.Info("Record",
			"index", i,
			"id", ref.ID,
			"database", ref.DatabaseType,
			"threads", ref.Threads,
			"tps", ref.TPS,
			"qps", ref.QPS)
	}

	// Extract record IDs
	recordIDs := make([]string, len(refs))
	for i, ref := range refs {
		recordIDs[i] = ref.ID
	}

	// Generate simplified report
	slog.Info("Generating simplified report...")
	report, err := comparisonUC.GenerateSimplifiedReport(ctx, recordIDs, comparison.GroupByThreads)
	if err != nil {
		slog.Error("Failed to generate simplified report", "error", err)
		os.Exit(1)
	}

	slog.Info("Simplified report generated",
		"report_id", report.ReportID,
		"groups", len(report.ConfigGroups))

	// Print report summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("SIMPLIFIED REPORT SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Report ID: %s\n", report.ReportID)
	fmt.Printf("Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Records: %d\n", report.SelectedRecords)
	fmt.Printf("Group By: %s\n", report.GroupBy)
	fmt.Printf("Config Groups: %d\n", len(report.ConfigGroups))
	fmt.Printf("Sanity Checks: %d\n", len(report.SanityChecks))

	// Print sanity check results
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("SANITY CHECKS")
	fmt.Println(strings.Repeat("-", 70))
	passed := 0
	for _, check := range report.SanityChecks {
		status := "✅ PASS"
		if !check.Passed {
			status = "❌ FAIL"
		} else {
			passed++
		}
		fmt.Printf("%s | %s\n", status, check.Name)
		if check.Details != "" {
			fmt.Printf("         Details: %s\n", check.Details)
		}
	}
	fmt.Printf("\nTotal: %d/%d passed\n", passed, len(report.SanityChecks))

	// Print findings
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("FINDINGS")
	fmt.Println(strings.Repeat("-", 70))
	if report.Findings != nil {
		fmt.Printf("Best TPS: threads=%d (TPS=%.2f)\n",
			report.Findings.BestTPSThreads, report.Findings.BestTPSValue)
		if report.Findings.BestLatencyThreads > 0 {
			fmt.Printf("Best Latency: threads=%d (p95=%.2fms)\n",
				report.Findings.BestLatencyThreads, report.Findings.BestLatencyValue)
		}
		if report.Findings.ScalingKnee > 0 {
			fmt.Printf("Scaling Knee: threads=%d\n", report.Findings.ScalingKnee)
		}
		fmt.Printf("Recommendation: %s\n", report.Findings.Recommendation)
	}

	// Print thread groups
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("THREAD GROUPS")
	fmt.Println(strings.Repeat("-", 70))
	for _, group := range report.ConfigGroups {
		fmt.Printf("\nthreads=%d (N=%d)\n", group.Threads, group.Statistics.N)
		fmt.Printf("  TPS:  mean=%.2f, stddev=%.2f, min=%.2f, max=%.2f\n",
			group.Statistics.TPS.Mean, group.Statistics.TPS.StdDev,
			group.Statistics.TPS.Min, group.Statistics.TPS.Max)
		fmt.Printf("  QPS:  mean=%.2f, stddev=%.2f, min=%.2f, max=%.2f\n",
			group.Statistics.QPS.Mean, group.Statistics.QPS.StdDev,
			group.Statistics.QPS.Min, group.Statistics.QPS.Max)
		fmt.Printf("  Lat: avg=%.2f ms, p95=%.2f ms\n",
			group.Statistics.LatencyAvg.Mean, group.Statistics.LatencyP95.Mean)
		fmt.Printf("  Errors: %d, Reconnects: %d\n",
			group.Statistics.Errors, group.Statistics.Reconnects)
	}

	// Generate Markdown report
	markdown := report.FormatMarkdown()

	// Save to file
	timestamp := report.GeneratedAt.Format("20060102_150405")
	markdownFile := fmt.Sprintf("./exports/simplified_report_%s.md", timestamp)
	err = os.WriteFile(markdownFile, []byte(markdown), 0644)
	if err != nil {
		slog.Error("Failed to write markdown file", "error", err)
		os.Exit(1)
	}

	slog.Info("Markdown report saved", "file", markdownFile)

	// Generate TXT report
	txt := report.FormatTXT()
	txtFile := fmt.Sprintf("./exports/simplified_report_%s.txt", timestamp)
	err = os.WriteFile(txtFile, []byte(txt), 0644)
	if err != nil {
		slog.Error("Failed to write txt file", "error", err)
		os.Exit(1)
	}

	slog.Info("TXT report saved", "file", txtFile)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✅ TEST PASSED - Simplified report generated successfully!")
	fmt.Println(strings.Repeat("=", 70))
}
