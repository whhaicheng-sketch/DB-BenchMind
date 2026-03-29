// Package main provides the Wails application entry point.
// This is the primary (and only) GUI framework for DB-BenchMind.
package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/bindings"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	slog.Info("DB-BenchMind starting...")

	// Create application context
	ctx := context.Background()

	// Ensure data directory exists
	ensureDataDir()

	// Initialize database
	dbPath := "./data/db-benchmind.db"
	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Database initialized", "path", dbPath)

	// Initialize repositories
	connRepo := repository.NewSQLiteConnectionRepository(db)
	templateRepo := repository.NewTemplateRepository(db)

	// Initialize keyring - use file fallback for GUI
	dataDir := "./data"
	keyringProvider, err := keyring.NewFileFallback(dataDir, "")
	if err != nil {
		slog.Error("Failed to initialize keyring", "error", err)
		os.Exit(1)
	}
	slog.Info("Keyring initialized")

	// Initialize use cases
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)
	templateUC := usecase.NewTemplateUseCase(templateRepo, "")

	// Load built-in templates
	if err := templateUC.LoadBuiltinTemplates(ctx); err != nil {
		slog.Warn("Failed to load built-in templates", "error", err)
	} else {
		templates, _ := templateUC.ListBuiltinTemplates(ctx)
		slog.Info("Built-in templates loaded", "count", len(templates))
	}

	// Initialize adapter registry and register adapters
	adapterReg := adapter.NewAdapterRegistry()
	adapterReg.Register(adapter.NewSysbenchAdapter())
	adapterReg.Register(adapter.NewSwingbenchAdapter())
	adapterReg.Register(adapter.NewHammerDBAdapter())
	slog.Info("Benchmark adapters registered")

	// Initialize run repository (in-memory for now)
	runRepo := usecase.NewMemoryRunRepository(filepath.Join(dataDir, "run-logs"))

	// Initialize benchmark use case
	benchmarkUC := usecase.NewBenchmarkUseCase(runRepo, adapterReg, connUC, templateUC)

	// Initialize report use case
	reportUC := usecase.NewReportUsecase(db)

	// Initialize report collector with DB for persisting report records
	reportCollector := usecase.NewDefaultReportCollector(
		usecase.WithReportsDir(filepath.Join(dataDir, "reports")),
		usecase.WithDB(db),
	)
	benchmarkUC.SetReportCollector(reportCollector)

	// Initialize AutoBench suite use case and runner
	suiteRepo := repository.NewSQLiteSuiteRepository(db)
	autobenchUC := usecase.NewAutoBenchSuiteUseCase(usecase.WithSuiteRepository(suiteRepo))
	autobenchRunner := usecase.NewAutoBenchSuiteRunner(
		autobenchUC,
		benchmarkUC,
		connUC,
		templateUC,
		usecase.WithManifestWriter(usecase.NewSuiteManifestWriter(dataDir)),
	)

	// Create application with basic options
	app := transportwails.NewApp()

	// Create Wails bindings
	execGuard := bindings.NewExecutionGuard()
	connBinding := bindings.NewConnectionBinding(connUC)
	templateBinding := bindings.NewTemplateBinding(templateUC)
	benchmarkBinding := bindings.NewBenchmarkBinding(benchmarkUC, connUC, templateUC, execGuard)
	monitorBinding := bindings.NewMonitorBinding()
	taskBinding := bindings.NewTaskBinding(benchmarkUC, connUC, templateUC, runRepo)
	reportBinding := bindings.NewReportBinding(reportUC)
	autobenchBinding := bindings.NewAutoBenchBinding(autobenchUC, autobenchRunner, execGuard)

	// Store benchmark binding for context injection
	app.SetBenchmarkBinding(benchmarkBinding)
	app.SetMonitorBinding(monitorBinding)
	app.SetTaskBinding(taskBinding)
	app.SetAutoBenchBinding(autobenchBinding)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "DB-BenchMind",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
			connBinding,
			templateBinding,
			benchmarkBinding,
			monitorBinding,
			taskBinding,
			reportBinding,
			autobenchBinding,
		},
	})

	if err != nil {
		slog.Error("Failed to start Wails application", "error", err)
	}
}

// ensureDataDir ensures the data directory exists.
func ensureDataDir() string {
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		slog.Error("Failed to create data directory", "error", err)
		os.Exit(1)
	}
	absPath, _ := filepath.Abs(dataDir)
	return absPath
}
