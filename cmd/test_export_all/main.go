package main

import (
	"context"
	"fmt"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

func main() {
	// Create multiple test records
	now := time.Now()
	records := []*history.Record{}

	for i := 1; i <= 3; i++ {
		record := &history.Record{
			ID:        fmt.Sprintf("test-run-%03d", i),
			CreatedAt: now.Add(time.Duration(-i) * time.Minute),
			ConnectionName: "TestConnection",
			TemplateName:   "Sysbench OLTP Read-Write",
			DatabaseType:   "MySQL",
			Threads:        4,
			StartTime:      now.Add(time.Duration(-i) * time.Minute).Add(-6 * time.Second),
			Duration:       6 * time.Second,

			TPSCalculated: 344.74 + float64(i)*10,
			LatencyAvg:    11.56 + float64(i)*0.5,
			LatencyMin:    8.49,
			LatencyMax:    22.90,
			LatencyP95:    13.46,
			LatencyP99:    18.23,
			LatencySum:    20030.44,

			ReadQueries:       24248 * int64(i),
			WriteQueries:      6928 * int64(i),
			OtherQueries:      3464 * int64(i),
			TotalQueries:      34640 * int64(i),
			TotalTransactions: 1732 * int64(i),

			IgnoredErrors: 0,
			Reconnects:    0,

			TotalTime:   5.0221,
			TotalEvents: 1732 * int64(i),

			EventsAvg:      433.0 * float64(i),
			EventsStddev:   1.58,
			ExecTimeAvg:    5.0076,
			ExecTimeStddev: 0.01,

			TimeSeries: []history.MetricSample{
				{
					Timestamp:  now.Add(time.Duration(-i) * time.Minute).Add(-5 * time.Second),
					Phase:      "run",
					TPS:        341.28,
					QPS:        6871.52,
					LatencyAvg: 11.23,
					LatencyP95:  13.46,
					ErrorRate:  0.0,
				},
				{
					Timestamp:  now.Add(time.Duration(-i) * time.Minute).Add(-1 * time.Second),
					Phase:      "run",
					TPS:        338.96,
					QPS:        6762.19,
					LatencyAvg: 11.78,
					LatencyP95:  13.95,
					ErrorRate:  0.0,
				},
			},
		}
		records = append(records, record)
	}

	// Create export use case
	exportUC := usecase.NewExportUseCase("./exports")
	ctx := context.Background()

	fmt.Println("=== Testing Export All Functionality ===\n")
	fmt.Printf("Total records to export: %d\n\n", len(records))

	// Test 1: Export all to TXT
	fmt.Println("Test 1: Exporting all records to TXT format...")
	count, exportDir, err := exportUC.ExportAllRecords(ctx, records, usecase.FormatTXT)
	if err != nil {
		fmt.Printf("❌ TXT export failed: %v\n", err)
	} else {
		fmt.Printf("✓ TXT export successful: %d records exported to %s\n", count, exportDir)
	}

	// Test 2: Export all to Markdown
	fmt.Println("\nTest 2: Exporting all records to Markdown format...")
	count, exportDir, err = exportUC.ExportAllRecords(ctx, records, usecase.FormatMarkdown)
	if err != nil {
		fmt.Printf("❌ Markdown export failed: %v\n", err)
	} else {
		fmt.Printf("✓ Markdown export successful: %d records exported to %s\n", count, exportDir)
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("Check ./exports/ directory for exported files")
}
