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
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
	"github.com/whhaicheng/DB-BenchMind/internal/transport/ui"
)

func main() {
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
	slog.Info("Use cases initialized")

	// 5. Start GUI
	slog.Info("Starting GUI")
	app := ui.NewApplication(connUC)
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
