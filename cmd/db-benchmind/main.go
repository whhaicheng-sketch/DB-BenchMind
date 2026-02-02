// Package main is the entry point for DB-BenchMind GUI application.
// Implements: Library-First Principle (constitution.md Article I)
// - cmd/ only does assembly and I/O
// - All business logic is in internal/ and pkg/
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
	"github.com/whhaicheng/DB-BenchMind/internal/transport/ui"
)

func main() {
	// Check working directory - MUST be project root!
	checkWorkingDirectory()

	// Set locale to avoid Fyne warning
	if os.Getenv("LANG") == "" || os.Getenv("LANG") == "C" {
		os.Setenv("LANG", "en_US.UTF-8")
	}

	// Setup logging to both file and console
	logDir := "./data/logs"
	os.MkdirAll(logDir, 0755)

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("db-benchmind-%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create multi-writer for both file and console
	// Use a custom handler that writes to both
	logger := slog.New(NewMultiHandler(os.Stdout, file))
	slog.SetDefault(logger)

	slog.Info("Starting DB-BenchMind", "log_file", logFile)

	// 1. Initialize database
	dbPath := "./data/db-benchmind.db"
	db, err := database.InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Database initialized", "path", dbPath)

	// 2. Initialize repositories
	connRepo := repository.NewSQLiteConnectionRepository(db)
	slog.Info("Repositories initialized")

	// 3. Initialize keyring - use file fallback for GUI
	dataDir := "./data"
	keyringProvider, err := keyring.NewFileFallback(dataDir, "")
	if err != nil {
		slog.Error("Failed to initialize keyring", "error", err)
		os.Exit(1)
	}
	slog.Info("Keyring initialized")

	// 4. Initialize use cases
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// Create template repository and use case
	templateRepo := usecase.NewMemoryTemplateRepository()
	templateUC := usecase.NewTemplateUseCase(templateRepo, "contracts/templates")

	// Load built-in templates
	if err := templateUC.LoadBuiltinTemplates(context.Background()); err != nil {
		slog.Warn("Failed to load built-in templates", "error", err)
	} else {
		// Get templates to verify loading
		templates, _ := templateUC.ListBuiltinTemplates(context.Background())
		slog.Info("Built-in templates loaded", "count", len(templates))
	}

	// Create adapter registry
	adapterReg := adapter.NewAdapterRegistry()
	adapterReg.Register(adapter.NewSysbenchAdapter())
	// Register other adapters as needed

	// Create run repository
	runRepo := usecase.NewMemoryRunRepository()

	// Create benchmark use case
	benchmarkUC := usecase.NewBenchmarkUseCase(runRepo, adapterReg, connUC, templateUC)

	// Create history repository and use case
	historyRepo := repository.NewSQLiteHistoryRepository(db)
	historyUC := usecase.NewHistoryUseCase(historyRepo)

	// Create export use case
	exportUC := usecase.NewExportUseCase("./exports")

	// Create comparison use case
	comparisonUC := usecase.NewComparisonUseCase(historyRepo, runRepo)

	slog.Info("Use cases initialized")

	// 5. Start GUI
	slog.Info("Starting GUI")
	app := ui.NewApplication(connUC, benchmarkUC, templateUC, historyUC, exportUC, comparisonUC)
	app.Run()
}

// MultiHandler writes log records to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a new multi-handler that writes to all provided handlers.
func NewMultiHandler(writers ...io.Writer) slog.Handler {
	var handlers []slog.Handler
	for _, w := range writers {
		handlers = append(handlers, slog.NewTextHandler(w, nil))
	}
	return &MultiHandler{handlers: handlers}
}

// Handle handles the log record by forwarding to all handlers.
func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// Enabled reports whether the handler is enabled for the given level.
func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// WithAttrs returns a new handler with the given attributes.
func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var newHandlers []slog.Handler
	for _, h := range m.handlers {
		newHandlers = append(newHandlers, h.WithAttrs(attrs))
	}
	return &MultiHandler{handlers: newHandlers}
}

// WithGroup returns a new handler with the given group name.
func (m *MultiHandler) WithGroup(name string) slog.Handler {
	var newHandlers []slog.Handler
	for _, h := range m.handlers {
		newHandlers = append(newHandlers, h.WithGroup(name))
	}
	return &MultiHandler{handlers: newHandlers}
}

// checkWorkingDirectory verifies that the application is running from the project root directory.
// This is critical because the application uses relative paths for:
// - Database: ./data/db-benchmind.db
// - Logs: ./data/logs/
// - Templates: contracts/templates/
func checkWorkingDirectory() {
	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine executable path: %v\n", err)
		return
	}

	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not get working directory: %v\n", err)
		os.Exit(1)
	}

	// Check for key files/directories that should exist in project root
	requiredPaths := []string{
		"bin/db-benchmind",    // Executable (if built with make)
		"contracts/templates", // Template directory
		"Makefile",            // Makefile
		"cmd/db-benchmind",    // Source directory
	}

	missingCount := 0
	var missingPaths []string
	for _, path := range requiredPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missingCount++
			missingPaths = append(missingPaths, path)
		}
	}

	// If more than half of the required paths are missing, we're likely in the wrong directory
	if missingCount > len(requiredPaths)/2 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "❌ ERROR: Not running from project root directory!\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Current working directory: %s\n", wd)
		fmt.Fprintf(os.Stderr, "Executable path: %s\n", execPath)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Required files/directories are missing:\n")
		for _, path := range missingPaths {
			fmt.Fprintf(os.Stderr, "  ❌ %s\n", path)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "SOLUTION:\n")
		fmt.Fprintf(os.Stderr, "  cd /path/to/DB-BenchMind  # Replace with your actual project path\n")
		fmt.Fprintf(os.Stderr, "  ./bin/db-benchmind gui\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "See: docs/OPERATION.md for detailed instructions.\n")
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	// Log the working directory (will be visible after logging is initialized)
	// We can't log here because slog is not initialized yet
}
