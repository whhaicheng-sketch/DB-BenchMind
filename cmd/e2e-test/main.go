package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: e2e-test <command>")
		fmt.Println("Commands: test-connections, run-suite, run-all")
		os.Exit(1)
	}

	ctx := context.Background()
	dataDir := "./data"

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Println("Failed to create data directory:", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.InitializeSQLite(ctx, filepath.Join(dataDir, "db-benchmind.db"))
	if err != nil {
		fmt.Println("DB init error:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize components (mirrors main.go wiring)
	connRepo := repository.NewSQLiteConnectionRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	keyringProvider, err := keyring.NewFileFallback(dataDir, "")
	if err != nil {
		fmt.Println("Keyring init error:", err)
		os.Exit(1)
	}

	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)
	templateUC := usecase.NewTemplateUseCase(templateRepo, "")

	// Load built-in templates
	if err := templateUC.LoadBuiltinTemplates(ctx); err != nil {
		slog.Warn("Failed to load built-in templates", "error", err)
	}

	adapterReg := adapter.NewAdapterRegistry()
	adapterReg.Register(adapter.NewSysbenchAdapter())
	adapterReg.Register(adapter.NewSwingbenchAdapter())
	adapterReg.Register(adapter.NewHammerDBAdapter())

	runRepo := usecase.NewMemoryRunRepository(filepath.Join(dataDir, "run-logs"))
	benchmarkUC := usecase.NewBenchmarkUseCase(runRepo, adapterReg, connUC, templateUC)

	reportUC := usecase.NewReportUsecase(db)
	reportCollector := usecase.NewDefaultReportCollector(
		usecase.WithReportsDir(dataDir),
		usecase.WithDB(db),
	)
	benchmarkUC.SetReportCollector(reportCollector)

	manifestWriter := usecase.NewSuiteManifestWriter(dataDir)
	suiteRepo := repository.NewSQLiteSuiteRepository(db)
	autobenchUC := usecase.NewAutoBenchSuiteUseCase(usecase.WithSuiteRepository(suiteRepo))
	autobenchRunner := usecase.NewAutoBenchSuiteRunner(
		autobenchUC, benchmarkUC, connUC, templateUC,
		usecase.WithManifestWriter(manifestWriter),
	)

	switch os.Args[1] {
	case "test-connections":
		testConnections(ctx, connUC)
	case "run-suite":
		runSuite(ctx, connUC, templateUC, autobenchUC, autobenchRunner, reportUC, db)
	case "run-all":
		fmt.Println("=== Step 1: Test Connections ===")
		testConnections(ctx, connUC)
		fmt.Println("\n=== Step 2: Run Suite ===")
		runSuite(ctx, connUC, templateUC, autobenchUC, autobenchRunner, reportUC, db)
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func testConnections(ctx context.Context, connUC *usecase.ConnectionUseCase) {
	conns, err := connUC.ListConnections(ctx)
	if err != nil {
		fmt.Println("List connections error:", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d connections\n\n", len(conns))
	allOK := true

	for _, conn := range conns {
		fmt.Printf("=== Testing %s (%s) ===\n", conn.GetName(), conn.GetType())
		fmt.Printf("    ID: %s\n", conn.GetID())

		fullConn, err := connUC.GetConnectionByID(ctx, conn.GetID())
		if err != nil {
			fmt.Printf("    GetConnectionByID error: %v\n", err)
			allOK = false
			continue
		}

		result, err := fullConn.Test(ctx)
		if err != nil {
			fmt.Printf("    Test error: %v\n", err)
			allOK = false
		} else {
			fmt.Printf("    DB Test: success=%v latency=%dms version=%s error=%s\n",
				result.Success, result.LatencyMs, result.DatabaseVersion, result.Error)
			if !result.Success {
				allOK = false
			}
		}
		fmt.Println()
	}

	if allOK {
		fmt.Println("ALL CONNECTIONS OK")
	} else {
		fmt.Println("SOME CONNECTIONS FAILED")
		os.Exit(1)
	}
}

func runSuite(ctx context.Context, connUC *usecase.ConnectionUseCase, templateUC *usecase.TemplateUseCase, autoUC *usecase.AutoBenchSuiteUseCase, runner *usecase.AutoBenchSuiteRunner, reportUC *usecase.ReportUsecase, db *sql.DB) {
	// List all connections
	conns, err := connUC.ListConnections(ctx)
	if err != nil {
		fmt.Println("List connections error:", err)
		os.Exit(1)
	}

	var connIDs []string
	for _, c := range conns {
		connIDs = append(connIDs, c.GetID())
		fmt.Printf("Connection: %s (%s) - %s\n", c.GetName(), c.GetType(), c.GetID())
	}

	// List all templates
	templates, err := templateUC.ListTemplates(ctx)
	if err != nil {
		fmt.Println("List templates error:", err)
		os.Exit(1)
	}
	fmt.Printf("\nTemplates: %d total\n", len(templates))
	for _, t := range templates {
		fmt.Printf("  %s - %s (tool=%s, profile=%s)\n", t.ID, t.Name, t.Tool, t.ProfileType)
	}

	// Create suite with "test" profile (smoke tests, ~60 seconds each)
	fmt.Println("\n=== Creating Suite ===")
	cleanupEnabled := false
	suite, err := autoUC.CreateSuite(ctx, usecase.CreateSuiteInput{
		Name:           "E2E Test Suite",
		ConnectionIDs:  connIDs,
		Profiles:       []domainautobench.ProfileType{domainautobench.ProfileTest},
		CleanupEnabled: &cleanupEnabled,
	})
	if err != nil {
		fmt.Println("Create suite error:", err)
		os.Exit(1)
	}
	fmt.Printf("Suite created: %s\n", suite.ID)

	// Get execution plan
	plan, err := autoUC.BuildExecutionPlan(ctx, suite.ID)
	if err != nil {
		fmt.Println("Build execution plan error:", err)
		os.Exit(1)
	}
	fmt.Printf("\nExecution plan (%d items):\n", len(plan.Items))
	for i, item := range plan.Items {
		fmt.Printf("  %d. conn_id=%s profile=%s suite_item_id=%s\n", i+1, item.ConnectionID, item.ProfileType, item.SuiteItemID)
	}

	// Start the suite
	fmt.Println("\n=== Starting Suite Execution ===")
	startTime := time.Now()

	err = runner.RunSuite(ctx, suite.ID)
	elapsed := time.Since(startTime)

	if err != nil {
		fmt.Printf("Suite execution error: %v (elapsed: %v)\n", err, elapsed)
	}

	// Get final status
	status, err := autoUC.GetSuiteStatus(ctx, suite.ID)
	if err != nil {
		fmt.Printf("Get suite status error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== Suite Results (elapsed: %v) ===\n", elapsed.Round(time.Second))
	fmt.Printf("Suite ID: %s\n", suite.ID)
	fmt.Printf("Status: %s\n", status.Status)
	fmt.Printf("Total: %d, Completed: %d, Pending: %d, Running: %d\n",
		status.TotalItems, status.CompletedItems, status.PendingItems, status.RunningItems)

	// Count success/fail from items
	var successCount, failCount int
	for _, item := range status.Items {
		if item.Status == domainautobench.SuiteItemStatusSuccess {
			successCount++
		} else if item.Status == domainautobench.SuiteItemStatusFailed {
			failCount++
		}
	}
	fmt.Printf("Items: success=%d, failed=%d\n", successCount, failCount)

	for _, item := range status.Items {
		icon := "OK  "
		if item.Status != domainautobench.SuiteItemStatusSuccess {
			icon = "FAIL"
		}
		fmt.Printf("  [%s] conn_id=%s db_type=%s profile=%s template=%s",
			icon, item.ConnectionID, item.DatabaseType, item.ProfileType, item.TemplateID)
		if item.ErrorSummary != "" {
			fmt.Printf(" error=%s", item.ErrorSummary)
		}
		if item.ReportID != "" {
			fmt.Printf(" report=%s", item.ReportID)
		}
		fmt.Println()
	}

	// Check reports in database
	fmt.Printf("\n=== Checking Reports in Database ===\n")
	reports, total, err := reportUC.ListReports(ctx, usecase.ListReportsOptions{})
	if err != nil {
		fmt.Printf("List reports error: %v\n", err)
	} else {
		fmt.Printf("Reports: %d total\n", total)
		for _, r := range reports {
			fmt.Printf("  %s | %s | %s | %s | status=%s", r.ID, r.ConnectionName, r.DatabaseType, r.TemplateName, r.Status)
			if r.TPS > 0 {
				fmt.Printf(" tps=%.1f lat=%.1fms", r.TPS, r.LatencyAvgMs)
			}
			fmt.Println()
			if r.MetricsJSONPath != "" {
				if _, err := os.Stat(r.MetricsJSONPath); err == nil {
					fmt.Printf("    metrics.json: %s [EXISTS]\n", r.MetricsJSONPath)
				} else {
					fmt.Printf("    metrics.json: %s [MISSING]\n", r.MetricsJSONPath)
				}
			}
			if r.MonitoringJSONPath != "" {
				if _, err := os.Stat(r.MonitoringJSONPath); err == nil {
					fmt.Printf("    monitoring.json: %s [EXISTS]\n", r.MonitoringJSONPath)
				} else {
					fmt.Printf("    monitoring.json: %s [MISSING]\n", r.MonitoringJSONPath)
				}
			}
			if r.RawJSONPath != "" {
				if _, err := os.Stat(r.RawJSONPath); err == nil {
					fmt.Printf("    raw.json: %s [EXISTS]\n", r.RawJSONPath)
				} else {
					fmt.Printf("    raw.json: %s [MISSING]\n", r.RawJSONPath)
				}
			}
		}
	}

	// Verify connections still exist
	fmt.Printf("\n=== Verifying Connections Preserved ===\n")
	connsAfter, _ := connUC.ListConnections(ctx)
	fmt.Printf("Connections before: %d, after: %d\n", len(conns), len(connsAfter))
	connMapBefore := map[string]string{}
	for _, c := range conns {
		connMapBefore[c.GetID()] = c.GetName()
	}
	for _, c := range connsAfter {
		if _, ok := connMapBefore[c.GetID()]; ok {
			fmt.Printf("  [OK] %s (%s) - preserved\n", c.GetName(), c.GetType())
		} else {
			fmt.Printf("  [NEW] %s (%s) - unexpected\n", c.GetName(), c.GetType())
		}
	}
	for id, name := range connMapBefore {
		found := false
		for _, c := range connsAfter {
			if c.GetID() == id {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  [DELETED] %s (%s) - CONNECTION WAS DELETED!\n", name, id)
		}
	}

	// Final verdict
	fmt.Println("\n=== FINAL VERDICT ===")
	if status.Status == domainautobench.SuiteStatusSuccess && int64(len(conns)) == int64(len(connsAfter)) {
		fmt.Println("ALL TESTS PASSED - ALL CONNECTIONS PRESERVED")
	} else {
		if status.Status != domainautobench.SuiteStatusSuccess {
			fmt.Printf("SUITE STATUS: %s (%d failed)\n", status.Status, failCount)
		}
		if int64(len(conns)) != int64(len(connsAfter)) {
			fmt.Println("CONNECTION COUNT MISMATCH!")
		}
		os.Exit(1)
	}
}
