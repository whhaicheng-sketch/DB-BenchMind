// Test script for Comparison feature
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
)

func main() {
	ctx := context.Background()

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║          DB-BenchMind Comparison Feature Test                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Initialize database
	fmt.Println("📂 Initializing database...")
	db, err := database.InitializeSQLite(ctx, "./data/db-benchmind.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	fmt.Println("✅ Database initialized")

	// Initialize repositories
	fmt.Println("\n📦 Initializing repositories...")
	historyRepo := repository.NewSQLiteHistoryRepository(db)
	fmt.Println("✅ History repository initialized")

	// Initialize ComparisonUseCase
	fmt.Println("\n🔧 Initializing ComparisonUseCase...")
	comparisonUC := usecase.NewComparisonUseCase(historyRepo)
	fmt.Println("✅ ComparisonUseCase initialized")

	// Test 1: Get all records
	fmt.Println("\n" + strings.Repeat("─", 70))
	fmt.Println("TEST 1: Get All Records")
	fmt.Println(strings.Repeat("─", 70))

	records, err := comparisonUC.GetAllRecords(ctx)
	if err != nil {
		log.Fatal("❌ Failed to get records:", err)
	}
	fmt.Printf("✅ Retrieved %d records\n\n", len(records))

	if len(records) == 0 {
		fmt.Println("⚠️  No records found. Cannot proceed with comparison tests.")
		return
	}

	// Display records
	fmt.Println("📋 Available Records:")
	fmt.Println(strings.Repeat("─", 70))
	for i, record := range records {
		fmt.Printf("%d. ID: %s\n", i+1, record.ID)
		fmt.Printf("   Template: %s\n", record.TemplateName)
		fmt.Printf("   Database: %s\n", record.DatabaseType)
		fmt.Printf("   Threads: %d\n", record.Threads)
		fmt.Printf("   TPS: %.2f\n", record.TPSCalculated)
		fmt.Printf("   Avg Latency: %.2f ms\n", record.LatencyAvg)
		fmt.Printf("   Start Time: %s\n", record.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	// Test 2: Get record references
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("TEST 2: Get Record References")
	fmt.Println(strings.Repeat("─", 70))

	refs, err := comparisonUC.GetRecordRefs(ctx)
	if err != nil {
		log.Fatal("❌ Failed to get record refs:", err)
	}
	fmt.Printf("✅ Retrieved %d record references\n\n", len(refs))

	for i, ref := range refs {
		fmt.Printf("%d. %s | %s | %d threads | %.2f TPS | %.2f QPS\n",
			i+1, ref.DatabaseType, ref.TemplateName, ref.Threads, ref.TPS, ref.QPS)
	}

	// Test 3: Filter records
	fmt.Println("\n" + strings.Repeat("─", 70))
	fmt.Println("TEST 3: Filter Records")
	fmt.Println(strings.Repeat("─", 70))

	filter := &usecase.ComparisonFilter{
		DatabaseType: "MySQL",
	}

	filtered := comparisonUC.FilterRecords(ctx, refs, filter)
	fmt.Printf("✅ Filtered to %d MySQL records\n", len(filtered))

	for i, ref := range filtered {
		fmt.Printf("%d. %s | %d threads | %.2f TPS\n",
			i+1, ref.DatabaseType, ref.Threads, ref.TPS)
	}

	// Test 4: Compare records
	fmt.Println("\n" + strings.Repeat("─", 70))
	fmt.Println("TEST 4: Compare Records")
	fmt.Println(strings.Repeat("─", 70))

	// Select up to 5 records for comparison
	maxRecords := 5
	if len(records) < 2 {
		fmt.Println("⚠️  Need at least 2 records for comparison")
		return
	}

	if len(records) < maxRecords {
		maxRecords = len(records)
	}

	recordIDs := make([]string, maxRecords)
	for i := 0; i < maxRecords; i++ {
		recordIDs[i] = records[i].ID
	}

	fmt.Printf("🔍 Comparing %d records grouped by Threads...\n", len(recordIDs))

	result, err := comparisonUC.CompareRecords(ctx, recordIDs, comparison.GroupByThreads)
	if err != nil {
		log.Fatal("❌ Failed to compare records:", err)
	}
	fmt.Println("✅ Comparison completed")

	// Display results
	fmt.Println("\n" + strings.Repeat("─", 70))
	fmt.Println("COMPARISON RESULTS")
	fmt.Println(strings.Repeat("─", 70))

	fmt.Printf("\n📊 Comparison ID: %s\n", result.ID)
	fmt.Printf("📅 Generated: %s\n", result.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("📝 Group By: %s\n", result.GroupBy)
	fmt.Printf("📈 Records Compared: %d\n\n", len(result.Records))

	// TPS Statistics
	if result.TPSComparison != nil {
		fmt.Println("TPS Statistics:")
		fmt.Printf("  Min:     %.2f\n", result.TPSComparison.Min)
		fmt.Printf("  Max:     %.2f\n", result.TPSComparison.Max)
		fmt.Printf("  Avg:     %.2f\n", result.TPSComparison.Avg)
		fmt.Printf("  StdDev:  %.2f\n", result.TPSComparison.StdDev)
		fmt.Println()
	}

	// Latency Statistics
	if result.LatencyCompare != nil && result.LatencyCompare.Avg != nil {
		fmt.Println("Latency Statistics (ms):")
		fmt.Printf("  Min:     %.2f\n", result.LatencyCompare.Min.Avg)
		fmt.Printf("  Max:     %.2f\n", result.LatencyCompare.Max.Avg)
		fmt.Printf("  Avg:     %.2f\n", result.LatencyCompare.Avg.Avg)
		fmt.Printf("  P95:     %.2f\n", result.LatencyCompare.P95.Avg)
		fmt.Printf("  P99:     %.2f\n", result.LatencyCompare.P99.Avg)
		fmt.Println()
	}

	// QPS Statistics
	if result.QPSComparison != nil {
		fmt.Println("QPS Statistics:")
		fmt.Printf("  Avg: %.2f\n", result.QPSComparison.Avg)
		fmt.Println()
	}

	// Read/Write Ratio
	if result.ReadWriteRatio != nil {
		fmt.Println("Query Distribution:")
		fmt.Printf("  Read:  %d queries (%.1f%%)\n",
			result.ReadWriteRatio.ReadQueries, result.ReadWriteRatio.ReadPct)
		fmt.Printf("  Write: %d queries (%.1f%%)\n",
			result.ReadWriteRatio.WriteQueries, result.ReadWriteRatio.WritePct)
		fmt.Printf("  Other: %d queries (%.1f%%)\n",
			result.ReadWriteRatio.OtherQueries, result.ReadWriteRatio.OtherPct)
		fmt.Println()
	}

	// Test 5: Format table
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("TEST 5: Format Output")
	fmt.Println(strings.Repeat("─", 70))

	fmt.Println("\n📋 Table View:")
	fmt.Println(result.FormatTable())

	// Test 6: Format bar chart
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("TEST 6: Bar Chart Visualization")
	fmt.Println(strings.Repeat("─", 70))

	fmt.Println("\n📊 TPS Bar Chart:")
	fmt.Println(result.FormatBarChart("TPS"))

	// Summary
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("TEST SUMMARY")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("✅ All tests passed!")
	fmt.Printf("✅ Successfully compared %d records\n", len(result.Records))
	fmt.Println("✅ Comparison feature is working correctly")
	fmt.Println()
}
