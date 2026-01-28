// Package main is the CLI entry point for DB-BenchMind.
// A simple CLI tool for database benchmark management.
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
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/tool"
)

const Version = "1.0.0"

func main() {
	// Setup logging to both file and console
	logDir := "./data/logs"
	os.MkdirAll(logDir, 0755)

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("db-benchmind-cli-%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create multi-writer for both file and console
	logger := slog.New(newMultiHandler(os.Stdout, file))
	slog.SetDefault(logger)

	slog.Info("DB-BenchMind CLI started", "version", Version, "log_file", logFile)

	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Simple command routing
	switch cmd {
	case "version", "-v", "--version":
		fmt.Printf("DB-BenchMind CLI v%s\n", Version)
	case "help", "-h", "--help":
		showHelp()
	case "list":
		listConnections()
	case "detect":
		detectTools()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf(`DB-BenchMind CLI v%s - Database Benchmark Management Tool

USAGE:
    db-benchmind-cli <command>

COMMANDS:
    list        List all database connections
    detect      Detect benchmark tools (sysbench, swingbench, hammerdb)
    version     Show version information
    help        Show this help message

EXAMPLES:
    # List connections
    db-benchmind-cli list

    # Detect tools
    db-benchmind-cli detect

For more information: https://github.com/whhaicheng/DB-BenchMind
`, Version)
}

func listConnections() {
	slog.Info("Listing connections", "command", "list")
	ctx := context.Background()

	// Initialize database
	os.MkdirAll("./data", 0755)
	db, err := database.InitializeSQLite(ctx, "./data/db-benchmind.db")
	if err != nil {
		slog.Error("Database init failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repository
	connRepo := repository.NewSQLiteConnectionRepository(db)

	// Initialize usecase
	keyringProvider, err := keyring.NewFileFallback("./data", "")
	if err != nil {
		slog.Error("Keyring init failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize keyring: %v\n", err)
		os.Exit(1)
	}
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// List connections
	conns, err := connUC.ListConnections(ctx)
	if err != nil {
		slog.Error("List connections failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: Failed to list connections: %v\n", err)
		os.Exit(1)
	}

	if len(conns) == 0 {
		fmt.Println("No connections found.")
		fmt.Println("\nTo add a connection, use the database API or CLI:")
		fmt.Println("  mysql - Add MySQL connection")
		fmt.Println("  postgresql - Add PostgreSQL connection")
		fmt.Println("  oracle - Add Oracle connection")
		fmt.Println("  sqlserver - Add SQL Server connection")
		return
	}

	fmt.Printf("\nFound %d connection(s):\n", len(conns))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i, conn := range conns {
		fmt.Printf("\n[%d] %s\n", i+1, conn.GetName())
		fmt.Printf("    ID:   %s\n", conn.GetID())
		fmt.Printf("    Type: %s\n", conn.GetType())
		fmt.Printf("    Host: %s\n", getHostInfo(conn))
	}
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func detectTools() {
	slog.Info("Detecting benchmark tools", "command", "detect")
	ctx := context.Background()

	// Initialize settings
	settingsRepo := repository.NewSettingsRepository("./data/db-benchmind.db")
	detector := tool.NewDetector()
	settingsUC := usecase.NewSettingsUseCase(settingsRepo, detector)

	fmt.Println("\nDetecting benchmark tools...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	tools := settingsUC.DetectTools(ctx)

	for toolType, info := range tools {
		if info.Found {
			fmt.Printf("✓ %s\n", toolType)
			fmt.Printf("  Path:    %s\n", info.Path)
			if info.Version != "" {
				fmt.Printf("  Version: %s\n", info.Version)
			}
		} else {
			fmt.Printf("✗ %s (not found)\n", toolType)
		}
		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\nTip: To install tools:")
	fmt.Println("  Sysbench:   apt-get install sysbench")
	fmt.Println("  Swingbench: Download from https://www.swingbench.com")
	fmt.Println("  HammerDB:   Download from https://www.hammerdb.com")
}

func getHostInfo(conn connection.Connection) string {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return fmt.Sprintf("%s:%d/%s", c.Host, c.Port, c.Database)
	case *connection.PostgreSQLConnection:
		return fmt.Sprintf("%s:%d/%s", c.Host, c.Port, c.Database)
	case *connection.OracleConnection:
		if c.ServiceName != "" {
			return fmt.Sprintf("%s:%d/%s", c.Host, c.Port, c.ServiceName)
		}
		return fmt.Sprintf("%s:%d:%s", c.Host, c.Port, c.SID)
	case *connection.SQLServerConnection:
		return fmt.Sprintf("%s:%d/%s", c.Host, c.Port, c.Database)
	default:
		return "unknown"
	}
}

// multiHandler writes log records to multiple handlers.
type multiHandler struct {
	handlers []slog.Handler
}

// newMultiHandler creates a new multi-handler that writes to all provided writers.
func newMultiHandler(writers ...io.Writer) slog.Handler {
	var handlers []slog.Handler
	for _, w := range writers {
		handlers = append(handlers, slog.NewTextHandler(w, nil))
	}
	return &multiHandler{handlers: handlers}
}

// Handle handles the log record by forwarding to all handlers.
func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// Enabled reports whether the handler is enabled for the given level.
func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// WithAttrs returns a new handler with the given attributes.
func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var newHandlers []slog.Handler
	for _, h := range m.handlers {
		newHandlers = append(newHandlers, h.WithAttrs(attrs))
	}
	return &multiHandler{handlers: newHandlers}
}

// WithGroup returns a new handler with the given group name.
func (m *multiHandler) WithGroup(name string) slog.Handler {
	var newHandlers []slog.Handler
	for _, h := range m.handlers {
		newHandlers = append(newHandlers, h.WithGroup(name))
	}
	return &multiHandler{handlers: newHandlers}
}
