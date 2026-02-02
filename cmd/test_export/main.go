package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

func main() {
	// Create a test history record with all fields
	now := time.Now()
	record := &history.Record{
		ID:        "test-run-001",
		CreatedAt: now,
		ConnectionName: "TestConnection",
		TemplateName:   "Sysbench OLTP Read-Write",
		DatabaseType:   "MySQL",
		Threads:        4,
		StartTime:      now.Add(-6 * time.Second),
		Duration:       6 * time.Second,

		// Core metrics
		TPSCalculated: 344.74,
		LatencyAvg:    11.56,
		LatencyMin:    8.49,
		LatencyMax:    22.90,
		LatencyP95:    13.46,
		LatencyP99:    18.23,
		LatencySum:    20030.44,

		// SQL Statistics
		ReadQueries:       24248,
		WriteQueries:      6928,
		OtherQueries:      3464,
		TotalQueries:      34640,
		TotalTransactions: 1732,

		// Errors and Reconnects
		IgnoredErrors: 0,
		Reconnects:    0,

		// General Statistics
		TotalTime:   5.0221,
		TotalEvents: 1732,

		// Threads Fairness
		EventsAvg:      433.0,
		EventsStddev:   1.58,
		ExecTimeAvg:    5.0076,
		ExecTimeStddev: 0.01,

		// Time series data (5 samples)
		TimeSeries: []history.MetricSample{
			{
				Timestamp:  now.Add(-5 * time.Second),
				Phase:      "run",
				TPS:        341.28,
				QPS:        6871.52,
				LatencyAvg: 11.23,
				LatencyP95:  13.46,
				LatencyP99:  18.12,
				ErrorRate:  0.0,
			},
			{
				Timestamp:  now.Add(-4 * time.Second),
				Phase:      "run",
				TPS:        348.11,
				QPS:        6934.23,
				LatencyAvg: 11.45,
				LatencyP95:  13.22,
				LatencyP99:  17.89,
				ErrorRate:  0.0,
			},
			{
				Timestamp:  now.Add(-3 * time.Second),
				Phase:      "run",
				TPS:        348.00,
				QPS:        6998.92,
				LatencyAvg: 11.48,
				LatencyP95:  13.22,
				LatencyP99:  18.01,
				ErrorRate:  0.0,
			},
			{
				Timestamp:  now.Add(-2 * time.Second),
				Phase:      "run",
				TPS:        351.00,
				QPS:        7023.09,
				LatencyAvg: 11.38,
				LatencyP95:  13.22,
				LatencyP99:  17.95,
				ErrorRate:  0.0,
			},
			{
				Timestamp:  now.Add(-1 * time.Second),
				Phase:      "run",
				TPS:        338.96,
				QPS:        6762.19,
				LatencyAvg: 11.78,
				LatencyP95:  13.95,
				LatencyP99:  18.45,
				ErrorRate:  0.0,
			},
		},
	}

	// Create export use case
	exportUC := usecase.NewExportUseCase("./exports")
	ctx := context.Background()

	fmt.Println("=== Testing Export Functionality ===\n")

	// Test 1: Export to TXT
	fmt.Println("Test 1: Exporting to TXT format...")
	txtPath, err := exportUC.ExportRecord(ctx, record, usecase.FormatTXT)
	if err != nil {
		fmt.Printf("❌ TXT export failed: %v\n", err)
	} else {
		fmt.Printf("✓ TXT export successful: %s\n", txtPath)
	}

	// Test 2: Export to Markdown
	fmt.Println("\nTest 2: Exporting to Markdown format...")
	mdPath, err := exportUC.ExportRecord(ctx, record, usecase.FormatMarkdown)
	if err != nil {
		fmt.Printf("❌ Markdown export failed: %v\n", err)
	} else {
		fmt.Printf("✓ Markdown export successful: %s\n", mdPath)
	}

	// Display TXT export
	if txtPath != "" {
		fmt.Println("\n=== TXT Export Content ===")
		content, _ := os.ReadFile(txtPath)
		fmt.Println(string(content))
	}

	// Verify key fields
	if txtPath != "" {
		content, _ := os.ReadFile(txtPath)
		contentStr := string(content)
		
		fmt.Println("\n=== Verification ===")
		
		checks := map[string]string{
			"sysbench header":                   "sysbench 1.0.20",
			"threads info":                      "Number of threads: 4",
			"time series data":                  "[ 1s ]",
			"SQL statistics":                    "SQL statistics:",
			"read queries":                      "read:",
			"transactions with per sec":         "transactions:",
			"general statistics":                "General statistics:",
			"latency section":                   "Latency (ms):",
			"latency min":                       "min:",
			"latency avg":                       "avg:",
			"latency max":                       "max:",
			"latency 95th percentile":           "95th percentile:",
			"latency sum":                       "sum:",
			"threads fairness":                  "Threads fairness:",
			"events avg/stddev":                 "events (avg/stddev):",
			"execution time avg/stddev":         "execution time (avg/stddev):",
		}
		
		allPassed := true
		for checkName, checkStr := range checks {
			if strings.Contains(contentStr, checkStr) {
				fmt.Printf("✓ %s\n", checkName)
			} else {
				fmt.Printf("❌ %s - NOT FOUND\n", checkName)
				allPassed = false
			}
		}
		
		if allPassed {
			fmt.Println("\n✓✓✓ All TXT export checks passed!")
		}
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("Export files saved to: ./exports/")
}
