// Package usecase provides benchmark execution business logic.
// Implements: REQ-EXEC-001 ~ REQ-EXEC-010
package usecase

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
	"golang.org/x/crypto/ssh"
)

var (
	// ErrBenchmarkNotFound is returned when a benchmark run is not found.
	ErrBenchmarkNotFound = errors.New("benchmark run not found")

	// ErrInvalidState is returned when an operation is invalid for the current state.
	ErrInvalidState = errors.New("invalid state for operation")

	// ErrPreCheckFailed is returned when pre-checks fail.
	ErrPreCheckFailed = errors.New("pre-check failed")

	// ErrExecutionFailed is returned when benchmark execution fails.
	ErrExecutionFailed = errors.New("execution failed")

	oracleSwingbenchPreflightPing = func(ctx context.Context, conn *connection.OracleConnection) error {
		result, err := conn.Test(ctx)
		if err != nil {
			return err
		}
		if result == nil {
			return fmt.Errorf("oracle preflight returned no result")
		}
		if !result.Success {
			if result.Error != "" {
				return fmt.Errorf("%s", result.Error)
			}
			return fmt.Errorf("oracle preflight login failed")
		}
		return nil
	}

	oracleSwingbenchPreflightUserExists = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		db, err := sql.Open("oracle", conn.GetDSNWithPassword())
		if err != nil {
			return false, err
		}
		defer db.Close()

		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var userCount sql.NullInt64
		err = db.QueryRowContext(checkCtx, `
			SELECT COUNT(*)
			FROM all_users
			WHERE username = UPPER(:1)
		`, workloadUser).Scan(&userCount)
		if err != nil {
			return false, err
		}
		return oracleSwingbenchSchemaCountValue(userCount) > 0, nil
	}

	oracleSwingbenchPreflightSchemaCheck = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		db, err := sql.Open("oracle", conn.GetDSNWithPassword())
		if err != nil {
			return false, err
		}
		defer db.Close()

		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var requiredTableCount sql.NullInt64
		var orderEntryPackageCount sql.NullInt64
		var sequenceCount sql.NullInt64
		err = db.QueryRowContext(checkCtx, `
			SELECT
				SUM(CASE
					WHEN object_type = 'TABLE' AND object_name IN ('CUSTOMERS', 'ORDERS', 'ORDER_ITEMS', 'WAREHOUSES')
					THEN 1 ELSE 0 END) AS required_tables,
				SUM(CASE
					WHEN object_name = 'ORDERENTRY' AND object_type IN ('PACKAGE', 'PACKAGE BODY') AND status = 'VALID'
					THEN 1 ELSE 0 END) AS orderentry_packages,
				SUM(CASE
					WHEN object_type = 'SEQUENCE'
					THEN 1 ELSE 0 END) AS sequences
			FROM all_objects
			WHERE owner = UPPER(:1)
			`, workloadUser).Scan(&requiredTableCount, &orderEntryPackageCount, &sequenceCount)
		if err != nil {
			return false, err
		}
		return oracleSwingbenchSchemaReady(
			oracleSwingbenchSchemaCountValue(requiredTableCount),
			oracleSwingbenchSchemaCountValue(orderEntryPackageCount),
			oracleSwingbenchSchemaCountValue(sequenceCount),
		), nil
	}

	benchmarkPrepareRemoteCPUCount = func(ctx context.Context, cfg *connection.SSHTunnelConfig) (int, error) {
		if cfg == nil {
			return 0, fmt.Errorf("ssh config is required")
		}
		if cfg.Port <= 0 || cfg.Port > 65535 {
			cfg.Port = 22
		}

		auth := make([]ssh.AuthMethod, 0, 2)
		if cfg.Password != "" {
			auth = append(auth, ssh.Password(cfg.Password))
		}
		if cfg.KeyPath != "" {
			keyBytes, err := os.ReadFile(cfg.KeyPath)
			if err != nil {
				keyBytes = []byte(cfg.KeyPath)
			}
			signer, err := ssh.ParsePrivateKey(keyBytes)
			if err != nil {
				return 0, fmt.Errorf("parse ssh private key: %w", err)
			}
			auth = append(auth, ssh.PublicKeys(signer))
		}
		if len(auth) == 0 {
			return 0, fmt.Errorf("ssh requires password or private key")
		}

		sshConfig := &ssh.ClientConfig{
			User:            cfg.Username,
			Auth:            auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
		}

		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), sshConfig)
		if err != nil {
			return 0, fmt.Errorf("ssh dial: %w", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			return 0, fmt.Errorf("ssh new session: %w", err)
		}
		defer session.Close()

		out, err := session.Output("nproc 2>/dev/null || getconf _NPROCESSORS_ONLN 2>/dev/null || grep -c ^processor /proc/cpuinfo 2>/dev/null")
		if err != nil {
			return 0, fmt.Errorf("ssh cpu probe: %w", err)
		}

		count, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err != nil {
			return 0, fmt.Errorf("parse remote cpu count %q: %w", strings.TrimSpace(string(out)), err)
		}
		return count, nil
	}
)

const (
	defaultPrepareThreads   = 4
	maxRemotePrepareThreads = 32
)

const (
	swingbenchInstallRoot = "/opt/benchtools/swingbench"
	swingbenchBinDir      = swingbenchInstallRoot + "/bin"
)

var (
	swingbenchDebugSearchRoot = swingbenchBinDir
	swingbenchDebugGlob       = filepath.Glob
)

var swingbenchRuntimeLinks = map[string]string{
	"sql":      filepath.Join(swingbenchInstallRoot, "sql"),
	"configs":  filepath.Join(swingbenchInstallRoot, "configs"),
	"launcher": filepath.Join(swingbenchInstallRoot, "launcher"),
	"lib":      filepath.Join(swingbenchInstallRoot, "lib"),
}

type oracleSwingbenchRunPreflightStatus string

type benchmarkRunPreflightStatus string

const (
	oracleSwingbenchRunPreflightOK                 oracleSwingbenchRunPreflightStatus = "ok"
	oracleSwingbenchRunPreflightUserMissing        oracleSwingbenchRunPreflightStatus = "user_missing"
	oracleSwingbenchRunPreflightUserLocked         oracleSwingbenchRunPreflightStatus = "user_locked"
	oracleSwingbenchRunPreflightInvalidCredentials oracleSwingbenchRunPreflightStatus = "invalid_credentials"
	oracleSwingbenchRunPreflightSchemaIncomplete   oracleSwingbenchRunPreflightStatus = "schema_incomplete"
	oracleSwingbenchRunPreflightCleanupInvalidated oracleSwingbenchRunPreflightStatus = "cleanup_invalidated"

	benchmarkRunPreflightOK                      benchmarkRunPreflightStatus = "ok"
	benchmarkRunPreflightInvalidCredentials      benchmarkRunPreflightStatus = "invalid_credentials"
	benchmarkRunPreflightDatabaseMissing         benchmarkRunPreflightStatus = "database_missing"
	benchmarkRunPreflightSchemaMissing           benchmarkRunPreflightStatus = "benchmark_schema_missing"
	benchmarkRunPreflightCleanupInvalidated      benchmarkRunPreflightStatus = "cleanup_invalidated"
	benchmarkRunPreflightBenchmarkObjectsMissing benchmarkRunPreflightStatus = "benchmark_objects_missing"
	benchmarkRunPreflightUnknownFailure          benchmarkRunPreflightStatus = "unknown_preflight_failure"
)

var (
	sysbenchMySQLRunPreflightCheck = func(ctx context.Context, conn *connection.MySQLConnection, dbName string) (benchmarkRunPreflightStatus, error) {
		adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?tls=false&timeout=5s&readTimeout=5s&writeTimeout=5s",
			conn.Username, conn.Password, conn.Host, conn.Port)
		db, err := sql.Open("mysql", adminDSN)
		if err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		defer db.Close()

		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := db.PingContext(checkCtx); err != nil {
			if isMySQLAuthenticationError(err.Error()) {
				return benchmarkRunPreflightInvalidCredentials, nil
			}
			return benchmarkRunPreflightUnknownFailure, err
		}

		var databaseCount int
		if err := db.QueryRowContext(checkCtx, "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?", dbName).Scan(&databaseCount); err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		if databaseCount == 0 {
			return benchmarkRunPreflightDatabaseMissing, nil
		}

		var tableCount int
		if err := db.QueryRowContext(checkCtx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'sbtest1'", dbName).Scan(&tableCount); err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		if tableCount == 0 {
			return benchmarkRunPreflightSchemaMissing, nil
		}
		return benchmarkRunPreflightOK, nil
	}

	sysbenchPostgreSQLRunPreflightCheck = func(ctx context.Context, conn *connection.PostgreSQLConnection, dbName string) (benchmarkRunPreflightStatus, error) {
		adminDSN := buildPostgreSQLRunPreflightDSN(conn, "postgres")
		db, err := sql.Open("postgres", adminDSN)
		if err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		defer db.Close()

		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := db.PingContext(checkCtx); err != nil {
			if isPostgreSQLAuthenticationError(err.Error()) {
				return benchmarkRunPreflightInvalidCredentials, nil
			}
			return benchmarkRunPreflightUnknownFailure, err
		}

		var databaseCount int
		if err := db.QueryRowContext(checkCtx, "SELECT COUNT(*) FROM pg_database WHERE datname = $1", dbName).Scan(&databaseCount); err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		if databaseCount == 0 {
			return benchmarkRunPreflightDatabaseMissing, nil
		}

		targetDB, err := sql.Open("postgres", buildPostgreSQLRunPreflightDSN(conn, dbName))
		if err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		defer targetDB.Close()

		if err := targetDB.PingContext(checkCtx); err != nil {
			if isPostgreSQLAuthenticationError(err.Error()) {
				return benchmarkRunPreflightInvalidCredentials, nil
			}
			if isPostgreSQLDatabaseMissingError(err.Error()) {
				return benchmarkRunPreflightDatabaseMissing, nil
			}
			return benchmarkRunPreflightUnknownFailure, err
		}

		var tableCount int
		if err := targetDB.QueryRowContext(checkCtx, "SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = 'sbtest1'").Scan(&tableCount); err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		if tableCount == 0 {
			return benchmarkRunPreflightSchemaMissing, nil
		}
		return benchmarkRunPreflightOK, nil
	}

	hammerDBSQLServerRunPreflightCheck = func(ctx context.Context, conn *connection.SQLServerConnection, databaseName string) (benchmarkRunPreflightStatus, error) {
		adminDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=master&encrypt=disable&trustservercertificate=%t",
			conn.Username, conn.Password, conn.Host, conn.Port, conn.TrustServerCertificate)
		db, err := sql.Open("sqlserver", adminDSN)
		if err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		defer db.Close()

		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := db.PingContext(checkCtx); err != nil {
			if isSQLServerAuthenticationError(err.Error()) {
				return benchmarkRunPreflightInvalidCredentials, nil
			}
			return benchmarkRunPreflightUnknownFailure, err
		}

		var databaseCount int
		if err := db.QueryRowContext(checkCtx, "SELECT COUNT(*) FROM sys.databases WHERE name = @p1", databaseName).Scan(&databaseCount); err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		if databaseCount == 0 {
			return benchmarkRunPreflightDatabaseMissing, nil
		}

		targetDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable&trustservercertificate=%t",
			conn.Username, conn.Password, conn.Host, conn.Port, databaseName, conn.TrustServerCertificate)
		targetDB, err := sql.Open("sqlserver", targetDSN)
		if err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		defer targetDB.Close()

		if err := targetDB.PingContext(checkCtx); err != nil {
			if isSQLServerAuthenticationError(err.Error()) {
				return benchmarkRunPreflightInvalidCredentials, nil
			}
			if isSQLServerDatabaseMissingError(err.Error()) {
				return benchmarkRunPreflightDatabaseMissing, nil
			}
			return benchmarkRunPreflightUnknownFailure, err
		}

		var objectCount int
		if err := targetDB.QueryRowContext(checkCtx, `
			SELECT COUNT(*)
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA = 'dbo'
			  AND TABLE_NAME IN ('warehouse', 'district', 'customer', 'history', 'orders', 'new_order', 'order_line', 'item', 'stock')
		`).Scan(&objectCount); err != nil {
			return benchmarkRunPreflightUnknownFailure, err
		}
		if objectCount < 9 {
			return benchmarkRunPreflightBenchmarkObjectsMissing, nil
		}
		return benchmarkRunPreflightOK, nil
	}
)

// RealtimeSampleCallback is called for each realtime sample during benchmark execution.
type RealtimeSampleCallback func(runID string, sample execution.MetricSample)

// BenchmarkUseCase provides benchmark execution business operations.
// Implements: REQ-EXEC-001 ~ REQ-EXEC-010
type BenchmarkUseCase struct {
	runRepo            RunRepository
	adapterReg         *adapter.AdapterRegistry
	connUseCase        *ConnectionUseCase
	templateUseCase    *TemplateUseCase
	realtimeCallback   RealtimeSampleCallback // Optional callback for realtime samples
	realtimeCallbackMu sync.RWMutex           // Protects realtimeCallback
	runningProcesses   map[string]*exec.Cmd   // Track running processes by run ID
	runningProcessesMu sync.RWMutex           // Protects runningProcesses
	reportCollector    ReportCollector        // Optional report collector for standalone reports
}

// NewBenchmarkUseCase creates a new benchmark use case.
func NewBenchmarkUseCase(
	runRepo RunRepository,
	adapterReg *adapter.AdapterRegistry,
	connUseCase *ConnectionUseCase,
	templateUseCase *TemplateUseCase,
) *BenchmarkUseCase {
	return &BenchmarkUseCase{
		runRepo:          runRepo,
		adapterReg:       adapterReg,
		connUseCase:      connUseCase,
		templateUseCase:  templateUseCase,
		runningProcesses: make(map[string]*exec.Cmd),
	}
}

// SetRealtimeCallback sets a callback function to receive realtime samples.
// The callback will be invoked for each sample as it's collected during benchmark execution.
func (uc *BenchmarkUseCase) SetRealtimeCallback(callback RealtimeSampleCallback) {
	uc.realtimeCallbackMu.Lock()
	defer uc.realtimeCallbackMu.Unlock()
	uc.realtimeCallback = callback
}

// WithReportCollector sets the report collector for standalone benchmark reports.
func WithReportCollector(collector ReportCollector) func(*BenchmarkUseCase) {
	return func(uc *BenchmarkUseCase) {
		uc.reportCollector = collector
	}
}

// SetReportCollector sets the report collector for standalone benchmark reports.
func (uc *BenchmarkUseCase) SetReportCollector(collector ReportCollector) {
	uc.reportCollector = collector
}

// =============================================================================
// Benchmark Execution
// Implements: REQ-EXEC-001 ~ REQ-EXEC-009
// =============================================================================

// StartBenchmark starts a new benchmark run.
// Implements: REQ-EXEC-001, REQ-EXEC-002
func (uc *BenchmarkUseCase) StartBenchmark(ctx context.Context, task *execution.BenchmarkTask) (*execution.Run, error) {
	// Validate task
	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPreCheckFailed, err)
	}

	// Get connection
	conn, err := uc.connUseCase.GetConnectionByID(ctx, task.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}

	// Get template
	tmpl, err := uc.templateUseCase.GetTemplate(ctx, task.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}

	// Get adapter
	adapt := uc.adapterReg.GetByTool(tmpl.Tool)
	if adapt == nil {
		return nil, fmt.Errorf("adapter not found for tool: %s", tmpl.Tool)
	}

	// Create run
	run := &execution.Run{
		ID:        uuid.New().String(),
		TaskID:    task.ID,
		State:     execution.StatePending,
		CreatedAt: time.Now(),
		WorkDir:   filepath.Join(os.TempDir(), fmt.Sprintf("db-benchmind-%s", uuid.New().String())),
	}

	// Save initial run
	if err := uc.runRepo.Save(ctx, run); err != nil {
		return nil, fmt.Errorf("save run: %w", err)
	}

	// Start execution in background
	go uc.executeBenchmark(context.Background(), run, conn, tmpl, adapt, task)

	return run, nil
}

// executeBenchmark executes the benchmark run.
// This runs in a goroutine.
func (uc *BenchmarkUseCase) executeBenchmark(
	ctx context.Context,
	run *execution.Run,
	conn connection.Connection,
	tmpl *domaintemplate.Template,
	adapt adapter.BenchmarkAdapter,
	task *execution.BenchmarkTask,
) {
	// Create work directory
	if err := os.MkdirAll(run.WorkDir, 0755); err != nil {
		uc.markAsFailed(ctx, run.ID, fmt.Sprintf("create work dir: %v", err))
		return
	}
	defer os.RemoveAll(run.WorkDir)

	// Build adapter config
	config := &adapter.Config{
		Connection: conn,
		Template:   tmpl,
		Parameters: task.Parameters,
		Options:    task.Options,
		WorkDir:    run.WorkDir,
	}
	prepareThreads, resolveErr := resolvePrepareThreads(ctx, conn)
	if resolveErr != nil {
		slog.Warn("Benchmark: Failed to resolve prepare threads, using default",
			"run_id", run.ID,
			"error", resolveErr,
			"default_prepare_threads", defaultPrepareThreads)
		prepareThreads = defaultPrepareThreads
	}
	config.PrepareThreads = prepareThreads

	slog.Info("Benchmark: executeBenchmark started",
		"run_id", run.ID,
		"skip_prepare", task.Options.SkipPrepare,
		"skip_cleanup", task.Options.SkipCleanup,
		"warmup_time", task.Options.WarmupTime)

	// Run pre-checks
	slog.Info("Benchmark: Running pre-checks", "run_id", run.ID)
	if err := uc.preChecks(ctx, run, adapt, config); err != nil {
		slog.Error("Benchmark: Pre-checks failed", "error", err, "run_id", run.ID)
		uc.markAsFailed(ctx, run.ID, fmt.Sprintf("pre-check: %v", err))
		return
	}
	slog.Info("Benchmark: Pre-checks passed", "run_id", run.ID)

	// Check if we should only execute prepare phase (time=0 indicates prepare-only)
	runTime := 0
	hasTime := false
	if timeVal, ok := task.Parameters["time"].(int); ok {
		runTime = timeVal
		hasTime = true
	}

	hasOriginalTime := false
	if _, ok := task.Parameters["_original_time"].(int); ok {
		hasOriginalTime = true
	}

	slog.Info("Benchmark: Checking execution mode",
		"run_id", run.ID,
		"hasTime", hasTime,
		"runTime", runTime,
		"hasOriginalTime", hasOriginalTime,
		"skipCleanup", task.Options.SkipCleanup)

	if hasTime && runTime == 0 && hasOriginalTime {
		// Prepare-only mode: execute prepare then mark as completed
		slog.Info("Benchmark: Prepare-only mode detected", "run_id", run.ID)

		// Create database if needed
		if err := uc.createDatabaseIfNeeded(ctx, run, adapt, config); err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("create database: %v", err))
			return
		}

		// Prepare already owns cleanup+init/create semantics inside the adapter command.
		slog.Info("Benchmark: Executing prepare phase (prepare-only mode)", "run_id", run.ID)
		if err := uc.executePhase(ctx, run, adapt, config, "prepare", execution.StatePreparing, execution.StatePrepared); err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("prepare: %v", err))
			return
		}

		msg1 := "✓ Prepare phase completed successfully"
		msg2 := "Info: Benchmark environment rebuilt with the current parameters."
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   strings.Repeat("=", 60),
		})
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   msg1,
		})
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   msg2,
		})
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   strings.Repeat("=", 60),
		})

		// For prepare-only mode, mark as completed directly (bypassing StatePrepared)
		uc.markAsCompleted(ctx, run.ID, 0)
		return
	}

	// Check if we should only execute cleanup phase
	if hasTime && runTime == 0 && !hasOriginalTime && !task.Options.SkipCleanup {
		// Cleanup-only mode
		slog.Info("Benchmark: Cleanup-only mode detected", "run_id", run.ID)

		// Cleanup phase
		// For cleanup-only mode, we bypass executePhase to avoid StatePrepared
		// and go directly to StateCompleted
		slog.Info("Benchmark: Executing cleanup phase (cleanup-only mode)", "run_id", run.ID)

		// Update state to running before executing command
		uc.updateState(ctx, run.ID, execution.StateRunning)

		cmd, err := adapt.BuildCleanupCommand(ctx, config)
		if err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("build cleanup command: %v", err))
			return
		}

		if err := uc.executeCommand(ctx, run, cmd); err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("cleanup: %v", err))
			return
		}

		// Cleanup completed successfully - add friendly message
		msg1 := "✓ Cleanup phase completed successfully"
		msg2 := "Info: All benchmark tables and data have been removed."
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   strings.Repeat("=", 60),
		})
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   msg1,
		})
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   msg2,
		})
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   strings.Repeat("=", 60),
		})

		// For cleanup-only mode, mark as completed directly (bypassing StatePrepared)
		uc.markAsCompleted(ctx, run.ID, 0)
		return
	}

	// Full benchmark execution (prepare + run + cleanup)

	// Create database if needed (before prepare phase)
	if !task.Options.SkipPrepare {
		if err := uc.createDatabaseIfNeeded(ctx, run, adapt, config); err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("create database: %v", err))
			return
		}
	}

	// Prepare phase
	if !task.Options.SkipPrepare {
		if err := uc.executePhase(ctx, run, adapt, config, "prepare", execution.StatePreparing, execution.StatePrepared); err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("prepare: %v", err))
			return
		}
	} else {
		uc.updateState(ctx, run.ID, execution.StatePrepared)
	}

	// Warmup phase
	if task.Options.WarmupTime > 0 {
		if err := uc.executeWarmup(ctx, run, adapt, config, task.Options.WarmupTime); err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("warmup: %v", err))
			return
		}
	}

	// Run phase
	startTime := time.Now()
	if err := uc.executeRun(ctx, run, adapt, config, task.Options.RunTimeout, conn, tmpl); err != nil {
		if uc.isRunStopped(ctx, run.ID) {
			return
		}
		uc.markAsFailed(ctx, run.ID, fmt.Sprintf("run: %v", err))
		return
	}
	duration := time.Since(startTime)

	// Cleanup phase
	if !task.Options.SkipCleanup {
		uc.executeCleanup(ctx, run, adapt, config)
	}

	// Mark as completed
	uc.markAsCompleted(ctx, run.ID, duration)

	// Collect and persist report (non-blocking)
	uc.collectStandaloneReport(ctx, run, conn, tmpl)
}

func resolvePrepareThreads(ctx context.Context, conn connection.Connection) (int, error) {
	sshCfg := sshConfigForConnection(conn)
	if sshCfg == nil || !sshCfg.Enabled || strings.TrimSpace(sshCfg.Host) == "" {
		return defaultPrepareThreads, nil
	}

	cpuCount, err := benchmarkPrepareRemoteCPUCount(ctx, sshCfg)
	if err != nil {
		return defaultPrepareThreads, err
	}

	threads := cpuCount / 2
	if threads < 1 {
		threads = 1
	}
	if threads > maxRemotePrepareThreads {
		threads = maxRemotePrepareThreads
	}
	return threads, nil
}

func sshConfigForConnection(conn connection.Connection) *connection.SSHTunnelConfig {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return c.SSH
	case *connection.PostgreSQLConnection:
		return c.SSH
	case *connection.OracleConnection:
		return c.SSH
	case *connection.SQLServerConnection:
		return c.SSH
	default:
		return nil
	}
}

// preChecks performs pre-execution checks.
// Implements: REQ-EXEC-001
func (uc *BenchmarkUseCase) preChecks(ctx context.Context, run *execution.Run, adapt adapter.BenchmarkAdapter, config *adapter.Config) error {
	// Validate config
	if err := adapt.ValidateConfig(ctx, config); err != nil {
		return fmt.Errorf("config validation: %w", err)
	}

	// Check tool availability
	if !uc.checkToolAvailable(ctx, adapt) {
		return fmt.Errorf("tool %s not available", adapt.Type())
	}

	// Check connection
	if err := uc.checkConnection(ctx, config.Connection); err != nil {
		return fmt.Errorf("connection check: %w", err)
	}

	// Check disk space
	if err := uc.checkDiskSpace(run.WorkDir, 1024*1024*1024); err != nil {
		return fmt.Errorf("disk space check: %w", err)
	}

	return nil
}

// createDatabaseIfNeeded creates the database if it doesn't exist.
// This runs before the prepare phase to ensure sysbench can connect to the database.
func (uc *BenchmarkUseCase) createDatabaseIfNeeded(ctx context.Context, run *execution.Run, adapt adapter.BenchmarkAdapter, config *adapter.Config) error {
	// Check if adapter supports database creation
	type DatabaseCreator interface {
		BuildCreateDatabaseCommand(ctx context.Context, config *adapter.Config) (*adapter.Command, error)
	}

	creator, ok := adapt.(DatabaseCreator)
	if !ok {
		// Adapter doesn't support database creation, skip
		slog.Info("Benchmark: Adapter does not support database creation, skipping", "adapter", adapt.Type())
		return nil
	}

	// Build create database command
	cmd, err := creator.BuildCreateDatabaseCommand(ctx, config)
	if err != nil {
		return fmt.Errorf("build create database command: %w", err)
	}

	// Execute command (ignore errors if database already exists)
	slog.Info("Benchmark: Creating database if not exists",
		"work_dir", run.WorkDir,
		"cmd_line", cmd.CmdLine,
		"env_vars", len(cmd.Env))
	if err := uc.executeCommand(ctx, run, cmd); err != nil {
		// Log error but don't fail - database might already exist
		// Get exit code if available
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			slog.Warn("Benchmark: Create database command failed",
				"error", err,
				"exit_code", exitErr.ExitCode(),
				"stderr", string(exitErr.Stderr))
		} else {
			slog.Warn("Benchmark: Create database command failed (database may already exist)", "error", err)
		}
	}

	return nil
}

// executePhase executes a single phase (prepare/cleanup).
func (uc *BenchmarkUseCase) executePhase(
	ctx context.Context,
	run *execution.Run,
	adapt adapter.BenchmarkAdapter,
	config *adapter.Config,
	phase string,
	targetState execution.RunState,
	successState execution.RunState,
) error {
	// Update state
	uc.updateState(ctx, run.ID, targetState)
	slog.Info("Benchmark: Starting phase", "phase", phase, "run_id", run.ID)
	uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Stream:    "info",
		Content:   fmt.Sprintf("Starting phase: %s", phase),
	})

	// Load password from keyring for Oracle connections
	// This is needed because OracleConnection.Password is not stored in JSON (security)
	conn := config.Connection
	if oracleConn, ok := conn.(*connection.OracleConnection); ok {
		if oracleConn.Password == "" {
			password, err := uc.connUseCase.GetPassword(ctx, oracleConn.ID)
			if err != nil {
				return fmt.Errorf("get password from keyring: %w", err)
			}
			oracleConn.Password = password
			slog.Info("Benchmark: Loaded password from keyring for Oracle phase",
				"phase", phase,
				"run_id", run.ID,
				"conn_id", oracleConn.ID,
				"username", oracleConn.Username)
		}
	}

	slog.Info("Benchmark: Starting phase with connection",
		"phase", phase,
		"run_id", run.ID,
		"connection_type", conn.GetType())
	uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Stream:    "info",
		Content:   fmt.Sprintf("Phase connection type: %s", conn.GetType()),
	})

	var cmd *adapter.Command
	var err error
	uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Stream:    "info",
		Content:   fmt.Sprintf("Building %s phase command", phase),
	})

	switch phase {
	case "prepare":
		cmd, err = adapt.BuildPrepareCommand(ctx, config)
	case "cleanup":
		cmd, err = adapt.BuildCleanupCommand(ctx, config)
	default:
		return fmt.Errorf("unknown phase: %s", phase)
	}

	if err != nil {
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "stderr",
			Content:   fmt.Sprintf("Failed to build %s phase command: %v", phase, err),
		})
		return fmt.Errorf("build %s command: %w", phase, err)
	}
	if cmd == nil {
		nilErr := fmt.Errorf("%s phase command builder returned nil command", phase)
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "stderr",
			Content:   nilErr.Error(),
		})
		return nilErr
	}
	if len(cmd.Commands) > 0 {
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   fmt.Sprintf("%s phase command constructed successfully as sequence with %d steps", phase, len(cmd.Commands)),
		})
	} else {
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   fmt.Sprintf("%s phase command constructed successfully", phase),
		})
	}

	slog.Info("Benchmark: Executing phase command",
		"phase", phase,
		"cmd", cmd.CmdLine,
		"run_id", run.ID)
	uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Stream:    "info",
		Content:   fmt.Sprintf("Executing phase command: %s", cmd.CmdLine),
	})

	// Execute command
	if err := uc.executeCommand(ctx, run, cmd); err != nil {
		slog.Warn("Benchmark: Phase command failed",
			"phase", phase,
			"error", err,
			"run_id", run.ID)
		return err
	}

	slog.Info("Benchmark: Phase completed successfully",
		"phase", phase,
		"run_id", run.ID)

	// Update to success state
	uc.updateState(ctx, run.ID, successState)
	return nil
}

// executeWarmup executes the warmup phase.
func (uc *BenchmarkUseCase) executeWarmup(
	ctx context.Context,
	run *execution.Run,
	adapt adapter.BenchmarkAdapter,
	config *adapter.Config,
	warmupTime int,
) error {
	uc.updateState(ctx, run.ID, execution.StateWarmingUp)

	// Build warmup command (same as run but with shorter time)
	cmd, err := adapt.BuildRunCommand(ctx, config)
	if err != nil {
		return err
	}

	// Modify time for warmup
	// This is a simplified version - real implementation would parse and modify the command
	_ = cmd
	_ = warmupTime

	// TODO: Execute warmup
	uc.updateState(ctx, run.ID, execution.StateRunning)
	return nil
}

// executeRun executes the main benchmark run with realtime monitoring.
// Implements: REQ-EXEC-002, REQ-EXEC-004, REQ-EXEC-005
func (uc *BenchmarkUseCase) executeRun(
	ctx context.Context,
	run *execution.Run,
	adapt adapter.BenchmarkAdapter,
	config *adapter.Config,
	timeout time.Duration,
	conn connection.Connection,
	tmpl *domaintemplate.Template,
) error {
	slog.Info("Benchmark: executeRun ENTER", "run_id", run.ID)

	// Load password from keyring for Oracle connections before any run preflight.
	if oracleConn, ok := conn.(*connection.OracleConnection); ok {
		if oracleConn.Password == "" {
			password, err := uc.connUseCase.GetPassword(ctx, oracleConn.ID)
			if err != nil {
				return fmt.Errorf("get password from keyring: %w", err)
			}
			oracleConn.Password = password
			slog.Info("Benchmark: Loaded password from keyring for Oracle",
				"run_id", run.ID,
				"conn_id", oracleConn.ID,
				"username", oracleConn.Username)
		}
	}

	// Check if tables exist and configuration matches before run
	// This ensures run uses the data prepared by prepare phase
	if adapt.Type() == adapter.AdapterTypeSysbench {
		slog.Info("Benchmark: Checking tables configuration before run", "run_id", run.ID)

		// Direct check without calling adapter method (not in interface)
		if err := uc.checkTablesConfigForRun(ctx, run, conn, config); err != nil {
			userMsg := fmt.Sprintf("✗ Error: %s\n\nPlease run Prepare first to create tables with the correct configuration.", err.Error())
			run.Message = userMsg
			run.ErrorMessage = userMsg
			uc.runRepo.Save(ctx, run)
			uc.markAsFailed(ctx, run.ID, userMsg)
			return fmt.Errorf("tables config check failed: %w", err)
		}
		slog.Info("Benchmark: Tables configuration check passed", "run_id", run.ID)
	}

	if err := benchmarkRunPreflight(ctx, config); err != nil {
		slog.Warn("Benchmark: benchmark run preflight failed", "run_id", run.ID, "error", err)
		return err
	}

	// Update state
	uc.updateState(ctx, run.ID, execution.StateRunning)

	// Update started_at
	now := time.Now()
	run.StartedAt = &now
	uc.runRepo.Save(ctx, run)

	// Log connection type
	slog.Info("Benchmark: Starting benchmark with connection",
		"run_id", run.ID,
		"connection_type", conn.GetType())

	// Build run command
	slog.Info("Benchmark: Building run command", "run_id", run.ID)
	cmd, err := adapt.BuildRunCommand(ctx, config)
	if err != nil {
		slog.Error("Benchmark: BuildRunCommand failed", "run_id", run.ID, "error", err)
		return err
	}
	slog.Info("Benchmark: Run command built successfully", "run_id", run.ID, "cmd", cmd.CmdLine)

	// Create context with timeout
	runCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		slog.Info("Benchmark: Created context with timeout", "run_id", run.ID, "timeout", timeout)
	} else {
		slog.Warn("Benchmark: timeout is 0 or negative, will use no deadline", "run_id", run.ID, "timeout", timeout)
	}

	// Start command
	slog.Info("Benchmark: Starting command", "run_id", run.ID)
	process, stdout, stderr, err := uc.startCommand(runCtx, cmd)
	if err != nil {
		slog.Error("Benchmark: startCommand failed", "run_id", run.ID, "error", err)
		return fmt.Errorf("start command: %w", err)
	}
	slog.Info("Benchmark: Command started successfully", "run_id", run.ID, "cmd", cmd.CmdLine)

	// Save process reference for later stop operations
	uc.runningProcessesMu.Lock()
	uc.runningProcesses[run.ID] = process
	uc.runningProcessesMu.Unlock()

	// Clean up process reference when done
	defer func() {
		uc.runningProcessesMu.Lock()
		delete(uc.runningProcesses, run.ID)
		uc.runningProcessesMu.Unlock()
	}()

	mirroredStdout := uc.mirrorOutputStream(runCtx, run.ID, "stdout", stdout)
	defer mirroredStdout.Close()

	// Start realtime collection from stdout only
	sampleCh, errCh, stdoutBuf := adapt.StartRealtimeCollection(runCtx, mirroredStdout)

	var stderrBuf strings.Builder
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), "\r")
			stderrBuf.WriteString(line)
			stderrBuf.WriteString("\n")
			uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "stderr",
				Content:   line,
			})
		}
	}()

	// Monitor process
	done := make(chan error, 1)
	go func() {
		done <- process.Wait()
	}()

	// Collect samples and monitor for completion
	zeroThroughputSamples := 0
	seenPositiveThroughput := false
	for {
		select {
		case sample, ok := <-sampleCh:
			if !ok {
				// Channel closed - wait briefly for any remaining samples to be processed
				// This ensures the final second's data is captured before we exit
				slog.Info("Benchmark: Sample channel closed, waiting for final samples", "run_id", run.ID)
				time.Sleep(500 * time.Millisecond)

				// Now wait for process to complete
				slog.Info("Benchmark: Waiting for process to complete", "run_id", run.ID)
				processErr := <-done
				slog.Info("Benchmark: Process completed, processErr", "run_id", run.ID, "error", processErr)
				<-stderrDone
				if processErr != nil {
					if uc.isRunStopped(ctx, run.ID) {
						return fmt.Errorf("run stopped")
					}
					stderrStr := stderrBuf.String()

					// Also check stdoutBuf - sysbench sometimes outputs errors to stdout
					stdoutStr := stdoutBuf.String()

					errMsg := processErr.Error()

					slog.Info("Benchmark: Run process failed",
						"run_id", run.ID,
						"error", errMsg,
						"stderr", stderrStr,
						"stdout", stdoutStr)

					// Parse error based on tool type - try stderr first, then stdout
					errorOutput := stderrStr
					if errorOutput == "" {
						errorOutput = stdoutStr
					}

					var userMsg string
					if tmpl.Tool == "swingbench" {
						userMsg = uc.parseSwingbenchError(errorOutput, stdoutBuf.String())
					} else if tmpl.Tool == "hammerdb" {
						userMsg = uc.parseHammerDBError(errorOutput)
					} else {
						// Default to sysbench error parsing
						userMsg = uc.parseSysbenchError(errorOutput)
					}

					// Set both Message (for UI dialog) and ErrorMessage (for error tracking)
					run.Message = userMsg
					run.ErrorMessage = userMsg
					uc.runRepo.Save(ctx, run)

					// Mark run as failed
					uc.markAsFailed(ctx, run.ID, userMsg)

					// Save error to logs
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "error",
						Content:   "============================================================",
					})
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "error",
						Content:   userMsg,
					})
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "error",
						Content:   "============================================================",
					})

					return fmt.Errorf("run failed: %w", processErr)
				}

				// Process completed successfully, parse final results
				slog.Info("Benchmark: Process completed successfully, parsing final results", "run_id", run.ID)
				stdoutStr := stdoutBuf.String()
				slog.Info("Benchmark: Sysbench output length", "run_id", run.ID, "length", len(stdoutStr))
				if len(stdoutStr) > 0 {
					slog.Info("Benchmark: Sysbench output preview", "run_id", run.ID, "output_preview", stdoutStr[:min(500, len(stdoutStr))])
				}
				finalResult, err := adapt.ParseFinalResults(ctx, withSwingbenchResultFileHint(stdoutStr, tmpl, cmd))
				slog.Info("Benchmark: ParseFinalResults returned", "run_id", run.ID, "err", err, "finalResult_nil", finalResult == nil)
				if err != nil {
					slog.Error("Benchmark: Failed to parse final results", "run_id", run.ID, "error", err)
				} else if swingbenchGuardErr := oracleSwingbenchZeroThroughputFailure(stdoutStr, finalResult); swingbenchGuardErr != nil {
					run.Message = swingbenchGuardErr.Error()
					run.ErrorMessage = swingbenchGuardErr.Error()
					uc.runRepo.Save(ctx, run)
					uc.markAsFailed(ctx, run.ID, swingbenchGuardErr.Error())
					return swingbenchGuardErr
				} else {
					slog.Info("Benchmark: Final result parsed",
						"run_id", run.ID,
						"transactions", finalResult.TotalTransactions,
						"tps", finalResult.TransactionsPerSec,
						"queries", finalResult.TotalQueries,
						"qps", finalResult.QueriesPerSec,
						"latency_min", finalResult.LatencyMin,
						"latency_avg", finalResult.LatencyAvg,
						"latency_max", finalResult.LatencyMax,
						"latency_p95", finalResult.LatencyP95)

					// Get threads/users count from parameters
					// Note: Oracle Swingbench uses "virtual_users", Sysbench uses "threads"
					threads := 0
					if t, ok := config.Parameters["threads"].(int); ok {
						threads = t
					} else if u, ok := config.Parameters["virtual_users"].(int); ok {
						// Swingbench uses "virtual_users" parameter
						threads = u
					}

					// Convert finalResult to BenchmarkResult and save to run
					slog.Info("Benchmark: Creating BenchmarkResult", "run_id", run.ID)
					result := &execution.BenchmarkResult{
						RunID:             run.ID,
						TPSCalculated:     finalResult.TransactionsPerSec,
						TPMCalculated:     finalResult.AvgTPM,
						MaxTPS:            finalResult.MaxTPS,
						AvgTPS:            finalResult.AvgTPS,
						MaxTPM:            finalResult.MaxTPM,
						AvgTPM:            finalResult.AvgTPM,
						LatencyAvg:        finalResult.LatencyAvg,
						LatencyMin:        finalResult.LatencyMin,
						LatencyMax:        finalResult.LatencyMax,
						LatencyP95:        finalResult.LatencyP95,
						LatencyP99:        finalResult.LatencyP99,
						LatencySum:        finalResult.LatencySum,
						TotalTransactions: finalResult.TotalTransactions,
						TotalQueries:      finalResult.TotalQueries,
						Duration:          time.Duration(finalResult.TotalTime) * time.Second,

						// SQL Statistics
						ReadQueries:   finalResult.ReadQueries,
						WriteQueries:  finalResult.WriteQueries,
						OtherQueries:  finalResult.OtherQueries,
						IgnoredErrors: finalResult.IgnoredErrors,
						Reconnects:    finalResult.Reconnects,

						// Oracle DML Statistics (Swingbench)
						SelectStatements:   finalResult.SelectStatements,
						InsertStatements:   finalResult.InsertStatements,
						UpdateStatements:   finalResult.UpdateStatements,
						DeleteStatements:   finalResult.DeleteStatements,
						CommitStatements:   finalResult.CommitStatements,
						RollbackStatements: finalResult.RollbackStatements,

						// General Statistics
						TotalTime:   finalResult.TotalTime,
						TotalEvents: finalResult.TotalEvents,

						// Threads Fairness
						EventsAvg:      finalResult.EventsAvg,
						EventsStddev:   finalResult.EventsStddev,
						ExecTimeAvg:    finalResult.ExecTimeAvg,
						ExecTimeStddev: finalResult.ExecTimeStddev,

						// Connection and Template Info (for History)
						ConnectionName: conn.GetName(),
						TemplateName:   tmpl.Name,
						DatabaseType:   string(conn.GetType()),
						Threads:        threads,
						StartTime:      *run.StartedAt,
					}

					slog.Info("Benchmark: Saving result to run", "run_id", run.ID)
					// Save result to run
					run.Result = result
					if err := uc.runRepo.Save(ctx, run); err != nil {
						slog.Error("Benchmark: Failed to save final result to run", "run_id", run.ID, "error", err)
					} else {
						slog.Info("Benchmark: Final result saved successfully", "run_id", run.ID)
					}
				}
				return nil
			}
			if tmpl.Tool == domaintemplate.ToolSwingbench && isOracleSwingbenchConfig(config) {
				if sample.TPS > 0 || sample.TPM > 0 {
					seenPositiveThroughput = true
					zeroThroughputSamples = 0
				} else if !seenPositiveThroughput && strings.Contains(sample.RawLine, "[0/") {
					zeroThroughputSamples++
					if zeroThroughputSamples >= 3 {
						errMsg := "Oracle Swingbench run failed: workload never established completed transactions and remained at 0 TPS/TPM ([0/N]). Run Prepare first."
						run.Message = errMsg
						run.ErrorMessage = errMsg
						uc.runRepo.Save(ctx, run)
						uc.markAsFailed(ctx, run.ID, errMsg)
						_ = terminateProcess(process, true)
						return fmt.Errorf("%s", errMsg)
					}
				}
			}

			// Save metric sample with error handling
			func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("Benchmark: Panic in SaveMetricSample", "run_id", run.ID, "panic", r)
					}
				}()
				metricSample := execution.MetricSample{
					Timestamp:  sample.Timestamp,
					Phase:      "run",
					TPS:        sample.TPS,
					QPS:        sample.QPS,
					TPM:        sample.TPM,
					LatencyAvg: sample.LatencyAvg,
					LatencyP95: sample.LatencyP95,
					LatencyP99: sample.LatencyP99,
					LatencyMax: sample.LatencyMax,
					ErrorRate:  sample.ErrorRate,
					Errors:     sample.Errors,
					Percentage: sample.Percentage,
					RawLine:    sample.RawLine,
				}
				if err := uc.runRepo.SaveMetricSample(ctx, run.ID, metricSample); err != nil {
					slog.Error("Benchmark: Failed to save metric sample", "run_id", run.ID, "error", err)
				}

				// Invoke realtime callback if set (for UI streaming)
				uc.realtimeCallbackMu.RLock()
				callback := uc.realtimeCallback
				uc.realtimeCallbackMu.RUnlock()

				if callback != nil {
					// Call callback in goroutine to avoid blocking sample processing
					go func() {
						defer func() {
							if r := recover(); r != nil {
								slog.Error("Benchmark: Panic in realtime callback", "run_id", run.ID, "panic", r)
							}
						}()
						callback(run.ID, metricSample)
					}()
				}
			}()

		case err, ok := <-errCh:
			if !ok {
				// Channel closed
				continue
			}
			// Log error with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("Benchmark: Panic in SaveLogEntry", "run_id", run.ID, "panic", r)
					}
				}()
				uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
					Timestamp: time.Now().Format(time.RFC3339),
					Stream:    "stderr",
					Content:   err.Error(),
				})
			}()

		case err := <-done:
			slog.Info("Benchmark: case <-done: executed", "run_id", run.ID, "err", err)
			if err != nil {
				if uc.isRunStopped(ctx, run.ID) {
					return fmt.Errorf("run stopped")
				}
				// Check if error is "table does not exist"
				errMsg := err.Error()
				slog.Info("Benchmark: Run command failed, checking error type", "run_id", run.ID, "error", errMsg)

				if strings.Contains(errMsg, "1146") || // Table doesn't exist
					strings.Contains(errMsg, "Table.*doesn't exist") ||
					strings.Contains(errMsg, "Table.*not exist") ||
					strings.Contains(errMsg, "no such table") {
					// Table does not exist - set user-friendly message
					slog.Info("Benchmark: Run phase - tables do not exist", "run_id", run.ID)
					run.Message = "✗ Error: Benchmark tables do not exist\n\nPlease run the Prepare phase first to create the tables and load data.\n\nGo to Task Configuration and click the '📦 Prepare' button."
					uc.runRepo.Save(ctx, run)

					// Save log entries
					msg1 := "✗ Error: Benchmark tables do not exist"
					msg2 := "Please run the Prepare phase first to create the tables and load data."
					msg3 := "Go to Task Configuration and click the '📦 Prepare' button."
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "error",
						Content:   strings.Repeat("=", 60),
					})
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "error",
						Content:   msg1,
					})
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "info",
						Content:   msg2,
					})
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "info",
						Content:   msg3,
					})
					uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
						Timestamp: time.Now().Format(time.RFC3339),
						Stream:    "error",
						Content:   strings.Repeat("=", 60),
					})
				}
				return fmt.Errorf("process error: %w", err)
			}
			// Process completed successfully, parse final results
			slog.Info("Benchmark: Process completed successfully (from done channel), parsing final results", "run_id", run.ID)
			stdoutStr := stdoutBuf.String()
			slog.Info("Benchmark: Sysbench output length", "run_id", run.ID, "length", len(stdoutStr))
			if len(stdoutStr) > 0 {
				slog.Info("Benchmark: Sysbench output preview", "run_id", run.ID, "output_preview", stdoutStr[:min(500, len(stdoutStr))])
			}
			finalResult, err := adapt.ParseFinalResults(ctx, withSwingbenchResultFileHint(stdoutStr, tmpl, cmd))
			slog.Info("Benchmark: ParseFinalResults returned", "run_id", run.ID, "err", err, "finalResult_nil", finalResult == nil)
			if err != nil {
				slog.Error("Benchmark: Failed to parse final results", "run_id", run.ID, "error", err)
			} else if swingbenchGuardErr := oracleSwingbenchZeroThroughputFailure(stdoutStr, finalResult); swingbenchGuardErr != nil {
				run.Message = swingbenchGuardErr.Error()
				run.ErrorMessage = swingbenchGuardErr.Error()
				uc.runRepo.Save(ctx, run)
				uc.markAsFailed(ctx, run.ID, swingbenchGuardErr.Error())
				return swingbenchGuardErr
			} else {
				slog.Info("Benchmark: Final result parsed",
					"run_id", run.ID,
					"transactions", finalResult.TotalTransactions,
					"tps", finalResult.TransactionsPerSec,
					"queries", finalResult.TotalQueries,
					"qps", finalResult.QueriesPerSec,
					"latency_min", finalResult.LatencyMin,
					"latency_avg", finalResult.LatencyAvg,
					"latency_max", finalResult.LatencyMax,
					"latency_p95", finalResult.LatencyP95)

				// Get threads/users count from parameters
				threads := 0
				if t, ok := config.Parameters["threads"].(int); ok {
					threads = t
				} else if u, ok := config.Parameters["virtual_users"].(int); ok {
					threads = u
				}

				// Convert finalResult to BenchmarkResult and save to run
				slog.Info("Benchmark: Creating BenchmarkResult", "run_id", run.ID)
				result := &execution.BenchmarkResult{
					RunID:              run.ID,
					TPSCalculated:      finalResult.TransactionsPerSec,
					TPMCalculated:      finalResult.AvgTPM,
					MaxTPS:             finalResult.MaxTPS,
					AvgTPS:             finalResult.AvgTPS,
					MaxTPM:             finalResult.MaxTPM,
					AvgTPM:             finalResult.AvgTPM,
					LatencyAvg:         finalResult.LatencyAvg,
					LatencyMin:         finalResult.LatencyMin,
					LatencyMax:         finalResult.LatencyMax,
					LatencyP95:         finalResult.LatencyP95,
					LatencyP99:         finalResult.LatencyP99,
					LatencySum:         finalResult.LatencySum,
					TotalTransactions:  finalResult.TotalTransactions,
					TotalQueries:       finalResult.TotalQueries,
					Duration:           time.Duration(finalResult.TotalTime) * time.Second,
					ReadQueries:        finalResult.ReadQueries,
					WriteQueries:       finalResult.WriteQueries,
					OtherQueries:       finalResult.OtherQueries,
					IgnoredErrors:      finalResult.IgnoredErrors,
					Reconnects:         finalResult.Reconnects,
					SelectStatements:   finalResult.SelectStatements,
					InsertStatements:   finalResult.InsertStatements,
					UpdateStatements:   finalResult.UpdateStatements,
					DeleteStatements:   finalResult.DeleteStatements,
					CommitStatements:   finalResult.CommitStatements,
					RollbackStatements: finalResult.RollbackStatements,
					TotalTime:          finalResult.TotalTime,
					TotalEvents:        finalResult.TotalEvents,
					EventsAvg:          finalResult.EventsAvg,
					EventsStddev:       finalResult.EventsStddev,
					ExecTimeAvg:        finalResult.ExecTimeAvg,
					ExecTimeStddev:     finalResult.ExecTimeStddev,
					ConnectionName:     conn.GetName(),
					TemplateName:       tmpl.Name,
					DatabaseType:       string(conn.GetType()),
					Threads:            threads,
					StartTime:          *run.StartedAt,
				}

				slog.Info("Benchmark: Saving result to run", "run_id", run.ID)
				run.Result = result
				if err := uc.runRepo.Save(ctx, run); err != nil {
					slog.Error("Benchmark: Failed to save final result to run", "run_id", run.ID, "error", err)
				} else {
					slog.Info("Benchmark: Final result saved successfully", "run_id", run.ID)
				}
			}
			return nil

		case <-runCtx.Done():
			// Timeout or cancellation
			if process.Process != nil {
				process.Process.Signal(syscall.SIGTERM)
				select {
				case <-time.After(30 * time.Second):
					// Force kill after 30 seconds
					process.Process.Signal(syscall.SIGKILL)
				case <-done:
				}
			}
			return ctx.Err()
		}
	}
}

// executeCleanup executes the cleanup phase (non-blocking).
func (uc *BenchmarkUseCase) executeCleanup(
	ctx context.Context,
	run *execution.Run,
	adapt adapter.BenchmarkAdapter,
	config *adapter.Config,
) {
	cmd, err := adapt.BuildCleanupCommand(ctx, config)
	if err != nil {
		return
	}

	// Execute without blocking
	go func() {
		uc.executeCommand(context.Background(), run, cmd)
	}()
}

// executeCommand executes a command and saves logs.
// For Swingbench, uses startCommand+Wait to properly handle background Java processes.
// For other tools, uses CombinedOutput for simpler synchronous execution.
// If the command contains a sequence of sub-commands, executes them in order.
func (uc *BenchmarkUseCase) executeCommand(ctx context.Context, run *execution.Run, cmd *adapter.Command) error {
	// Check if this is a command sequence
	if len(cmd.Commands) > 0 {
		slog.Info("Benchmark: Executing command sequence",
			"run_id", run.ID,
			"steps", len(cmd.Commands),
			"description", cmd.Description)
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   fmt.Sprintf("Executing command sequence (%d steps)", len(cmd.Commands)),
		})
		return uc.executeCommandSequence(ctx, run, cmd)
	}

	// Check if this is a Swingbench command (uses java with LauncherBootstrap)
	// We detect Swingbench by looking for LauncherBootstrap class name
	isSwingbench := isSwingbenchCommandLine(cmd.CmdLine)

	if isSwingbench {
		// Use startCommand+Wait for Swingbench to properly track background processes
		return uc.executeCommandSwingbench(ctx, run, cmd)
	}

	// For other tools (sysbench, hammerdb), use CombinedOutput
	return uc.executeCommandSync(ctx, run, cmd)
}

// executeCommandSync executes a command with realtime output streaming.
// Used for sysbench, hammerdb prepare/cleanup phases to show progress in UI.
// If cmd.Retry is configured, implements retry logic with exponential backoff.
func (uc *BenchmarkUseCase) executeCommandSync(ctx context.Context, run *execution.Run, cmd *adapter.Command) error {
	var lastErr error

	// Retry loop
	maxRetries := 0
	if cmd.Retry != nil && cmd.Retry.MaxRetries > 0 {
		maxRetries = cmd.Retry.MaxRetries
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// If not the first attempt, log retry and wait
		if attempt > 0 {
			backoffDelay := uc.calculateBackoffDelay(cmd.Retry, attempt-1)
			slog.Warn("Benchmark: Retrying command",
				"run_id", run.ID,
				"attempt", attempt,
				"max_retries", maxRetries,
				"delay", backoffDelay,
				"last_error", lastErr)

			// Add log entry for retry
			uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "warn",
				Content:   fmt.Sprintf("Retrying command (attempt %d/%d) after %v...", attempt, maxRetries, backoffDelay),
			})

			// Wait before retry
			select {
			case <-time.After(backoffDelay):
				// Continue with retry
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Execute the command
		err := uc.executeCommandSyncOnce(ctx, run, cmd, attempt)
		if err == nil {
			// Success
			if attempt > 0 {
				// Log successful retry
				uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
					Timestamp: time.Now().Format(time.RFC3339),
					Stream:    "info",
					Content:   fmt.Sprintf("✓ Command succeeded on retry attempt %d", attempt),
				})
			}
			return nil
		}

		lastErr = err

		// Check if this is a retryable error
		if attempt < maxRetries && uc.isRetryableError(err) {
			// This is retryable, will retry in next iteration
			slog.Warn("Benchmark: Command failed with retryable error",
				"run_id", run.ID,
				"attempt", attempt,
				"error", err)
			continue
		}

		// Either not retryable or no more retries left
		slog.Error("Benchmark: Command failed with non-retryable error or max retries exceeded",
			"run_id", run.ID,
			"attempt", attempt,
			"max_retries", maxRetries,
			"error", err)
		break
	}

	return lastErr
}

// executeCommandSyncOnce executes a command once (no retry logic).
// Internal method used by executeCommandSync.
func (uc *BenchmarkUseCase) executeCommandSyncOnce(ctx context.Context, run *execution.Run, cmd *adapter.Command, attempt int) error {
	var execCmd *exec.Cmd

	// Check if command contains pipe operator - if so, use shell to execute
	// This is necessary because exec.CommandContext doesn't interpret shell operators
	if strings.Contains(cmd.CmdLine, "|") || strings.Contains(cmd.CmdLine, "&&") || strings.Contains(cmd.CmdLine, "||") {
		// Use bash -c to execute the full command with shell operators
		execCmd = exec.CommandContext(ctx, "bash", "-c", cmd.CmdLine)
	} else {
		// Parse command line for simple commands without shell operators
		parts, err := parseCommandLine(cmd.CmdLine)
		if err != nil {
			return err
		}
		execCmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
	}

	execCmd.Dir = cmd.WorkDir
	execCmd.Env = append(os.Environ(), cmd.Env...)
	execCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Debug: Log command execution with environment details
	hasMYSQL_PWD := false
	hasPGPASSWORD := false
	for _, env := range execCmd.Env {
		if strings.HasPrefix(env, "MYSQL_PWD=") {
			hasMYSQL_PWD = true
		}
		if strings.HasPrefix(env, "PGPASSWORD=") {
			hasPGPASSWORD = true
		}
	}

	// Log the actual command that will be executed
	slog.Info("Benchmark: === EXECUTING COMMAND (SYNC WITH REALTIME) ===",
		"run_id", run.ID,
		"cmd_line", cmd.CmdLine,
		"work_dir", execCmd.Dir,
		"env_count", len(execCmd.Env),
		"has_mysql_pwd", hasMYSQL_PWD,
		"has_pgpassword", hasPGPASSWORD)

	// Create pipes for stdout and stderr
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	// Start the command
	if err := execCmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("start command: %w", err)
	}

	// Determine adapter type for realtime collection
	var adapt adapter.BenchmarkAdapter
	if strings.Contains(cmd.CmdLine, "sysbench") {
		adapt = uc.adapterReg.Get(adapter.AdapterTypeSysbench)
	} else if strings.Contains(cmd.CmdLine, "hammerdb") {
		adapt = uc.adapterReg.Get(adapter.AdapterTypeHammerDB)
	}

	// Start realtime collection if adapter available
	var sampleCh <-chan adapter.Sample
	var errCh <-chan error
	var stdoutBuf *strings.Builder
	stdoutDone := make(chan struct{})
	mirroredStdout := uc.mirrorOutputStream(ctx, run.ID, "stdout", stdout)
	defer mirroredStdout.Close()

	if adapt != nil {
		sampleCh, errCh, stdoutBuf = adapt.StartRealtimeCollection(ctx, mirroredStdout)
		close(stdoutDone)
	} else {
		// No adapter, just collect output
		stdoutBuf = &strings.Builder{}
		go func() {
			defer close(stdoutDone)
			io.Copy(stdoutBuf, mirroredStdout)
		}()
	}

	// Collect stderr to log entries
	var stderrBuf strings.Builder
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")

			// Save stderr to log
			uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "stderr",
				Content:   line,
			})

			slog.Info("Benchmark: command stderr", "run_id", run.ID, "line", line)
		}
	}()

	// Process samples from realtime collection
	// Start goroutine to wait for process completion
	done := make(chan error, 1)
	go func() {
		err := execCmd.Wait()
		slog.Info("Benchmark: Process completed", "run_id", run.ID, "error", err)
		done <- err
	}()

	// Process realtime samples in a separate goroutine
	// This doesn't block the process waiting
	sampleLoopDone := make(chan struct{})
	go func() {
		defer close(sampleLoopDone)

		if adapt != nil && sampleCh != nil {
			for {
				select {
				case sample, ok := <-sampleCh:
					if !ok {
						// Channel closed - output collection finished
						slog.Info("Benchmark: Sample channel closed", "run_id", run.ID)
						return
					}

					// Save sample to repository
					metricSample := execution.MetricSample{
						Timestamp:  sample.Timestamp,
						Phase:      "prepare", // Default to prepare for prepare/cleanup
						TPS:        sample.TPS,
						QPS:        sample.QPS,
						LatencyAvg: sample.LatencyAvg,
						LatencyP95: sample.LatencyP95,
						ErrorRate:  sample.ErrorRate,
						RawLine:    sample.RawLine,
					}

					if err := uc.runRepo.SaveMetricSample(ctx, run.ID, metricSample); err != nil {
						slog.Error("Benchmark: Failed to save metric sample", "run_id", run.ID, "error", err)
					}

					// Invoke realtime callback if set (for UI streaming)
					uc.realtimeCallbackMu.RLock()
					callback := uc.realtimeCallback
					uc.realtimeCallbackMu.RUnlock()

					if callback != nil {
						go func() {
							defer func() {
								if r := recover(); r != nil {
									slog.Error("Benchmark: Panic in realtime callback", "run_id", run.ID, "panic", r)
								}
							}()
							callback(run.ID, metricSample)
						}()
					}

				case err, ok := <-errCh:
					if !ok {
						continue
					}
					if isIgnorableSampleCollectorError(err) {
						slog.Debug("Benchmark: Sample collector finished during shutdown", "run_id", run.ID, "error", err)
						continue
					}
					slog.Error("Benchmark: Error from sample collector", "run_id", run.ID, "error", err)
				}
			}
		}
	}()

	// Wait for command to complete (this is the MAIN wait)
	slog.Info("Benchmark: Waiting for command to complete", "run_id", run.ID)
	processErr := <-done
	slog.Info("Benchmark: Command wait finished", "run_id", run.ID, "error", processErr)

	// Wait for sample loop to finish
	<-sampleLoopDone
	<-stdoutDone
	<-stderrDone

	// Close pipes
	stdout.Close()
	stderr.Close()

	// If command failed, return error with output
	if processErr != nil {
		slog.Error("Benchmark: Command failed", "run_id", run.ID, "exit_error", processErr)
		var details []string
		if stdoutBuf != nil && stdoutBuf.Len() > 0 {
			details = append(details, fmt.Sprintf("stdout:\n%s", stdoutBuf.String()))
		}
		if stderrBuf.Len() > 0 {
			details = append(details, fmt.Sprintf("stderr:\n%s", stderrBuf.String()))
		}
		if len(details) == 0 {
			details = append(details, "stderr:\n")
		}
		return fmt.Errorf("command failed with exit status %v: %s", processErr, strings.Join(details, "\n"))
	}

	if hammerDBErr := detectHammerDBCommandFailure(cmd.CmdLine, stdoutBuf.String(), stderrBuf.String()); hammerDBErr != nil {
		slog.Error("Benchmark: HammerDB command reported failure despite zero exit status",
			"run_id", run.ID,
			"error", hammerDBErr)
		return hammerDBErr
	}

	return nil
}

// executeCommandSwingbench executes a Swingbench command using startCommand+Wait.
// This properly handles Swingbench's background Java process architecture.
func (uc *BenchmarkUseCase) executeCommandSwingbench(ctx context.Context, run *execution.Run, cmd *adapter.Command) error {
	slog.Info("Benchmark: === EXECUTING COMMAND (SWINGBENCH) ===",
		"run_id", run.ID,
		"cmd_line", cmd.CmdLine,
		"work_dir", cmd.WorkDir)
	startedAt := time.Now()
	defer uc.ingestSwingbenchDebugLogs(ctx, run.ID, cmd, startedAt)

	// Get swingbench adapter for realtime collection
	adapt := uc.adapterReg.Get(adapter.AdapterTypeSwingbench)
	if adapt == nil {
		return fmt.Errorf("swingbench adapter not found")
	}

	if shouldCleanupResidualSwingbenchProcesses(cmd) {
		uc.cleanupResidualSwingbenchProcesses(ctx, run.ID)
	}

	// Start the command
	process, stdout, stderr, err := uc.startCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	// Save process reference for potential stop operation
	uc.runningProcessesMu.Lock()
	uc.runningProcesses[run.ID] = process
	uc.runningProcessesMu.Unlock()

	// Clean up process reference when done
	defer func() {
		uc.runningProcessesMu.Lock()
		delete(uc.runningProcesses, run.ID)
		uc.runningProcessesMu.Unlock()
	}()

	defer stderr.Close()
	mirroredStdout := uc.mirrorOutputStream(ctx, run.ID, "stdout", stdout)
	defer mirroredStdout.Close()

	// Start realtime sample collection from stdout
	sampleCh, errCh, _ := adapt.StartRealtimeCollection(ctx, mirroredStdout)

	// Also capture stderr to log entries
	var stderrBuf strings.Builder
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")

			uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "stderr",
				Content:   line,
			})

			slog.Info("Benchmark: swingbench stderr", "run_id", run.ID, "line", line)
		}
	}()

	// Wait for process to complete
	done := make(chan error, 1)
	go func() {
		done <- process.Wait()
	}()

	// Collect samples and monitor for completion
	for {
		select {
		case sample, ok := <-sampleCh:
			if !ok {
				// Channel closed - wait briefly for any remaining samples
				slog.Info("Benchmark: Sample channel closed", "run_id", run.ID)
				time.Sleep(500 * time.Millisecond)
				// Wait for process to complete
				processErr := <-done
				<-stderrDone
				if processErr != nil {
					return processErr
				}
				return nil
			}

			// Save sample to repository
			metricSample := execution.MetricSample{
				Timestamp:  sample.Timestamp,
				Phase:      "prepare", // Swingbench prepare/cleanup phase
				TPS:        sample.TPS,
				QPS:        sample.QPS,
				TPM:        sample.TPM,
				LatencyAvg: sample.LatencyAvg,
				LatencyP95: sample.LatencyP95,
				LatencyP99: sample.LatencyP99,
				ErrorRate:  sample.ErrorRate,
				Errors:     sample.Errors,
				Percentage: sample.Percentage,
				RawLine:    sample.RawLine,
			}

			if err := uc.runRepo.SaveMetricSample(ctx, run.ID, metricSample); err != nil {
				slog.Error("Benchmark: Failed to save metric sample", "run_id", run.ID, "error", err)
			}

			// Invoke realtime callback if set (for UI streaming)
			uc.realtimeCallbackMu.RLock()
			callback := uc.realtimeCallback
			uc.realtimeCallbackMu.RUnlock()

			if callback != nil {
				go func() {
					defer func() {
						if r := recover(); r != nil {
							slog.Error("Benchmark: Panic in realtime callback", "run_id", run.ID, "panic", r)
						}
					}()
					callback(run.ID, metricSample)
				}()
			}

		case err, ok := <-errCh:
			if !ok {
				continue
			}
			if isIgnorableSampleCollectorError(err) {
				slog.Debug("Benchmark: Sample collector finished during shutdown", "run_id", run.ID, "error", err)
				continue
			}
			slog.Error("Benchmark: Error from sample collector", "run_id", run.ID, "error", err)

		case <-ctx.Done():
			slog.Info("Benchmark: Context cancelled", "run_id", run.ID)
			if process.Process != nil {
				_ = terminateProcess(process, true)
			}
			return ctx.Err()
		}
	}
}

// executeCommandSequence executes a sequence of commands in order.
// Each step must succeed before the next one runs.
// Implements multi-phase operations like Oracle prepare (tablespace creation + probe + oewizard).
func (uc *BenchmarkUseCase) executeCommandSequence(ctx context.Context, run *execution.Run, cmd *adapter.Command) error {
	for i, step := range cmd.Commands {
		stepNum := i + 1
		totalSteps := len(cmd.Commands)

		// Log step start
		stepDesc := step.StepName
		if stepDesc == "" {
			stepDesc = fmt.Sprintf("Step %d", stepNum)
		}
		slog.Info("Benchmark: === EXECUTING SEQUENCE STEP ===",
			"run_id", run.ID,
			"step", stepNum,
			"total", totalSteps,
			"step_name", stepDesc,
			"description", step.Description)
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   fmt.Sprintf("Starting sequence step=%d/%d: %s", stepNum, totalSteps, stepDesc),
		})

		// Add separator log entry for UI
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   fmt.Sprintf("=== %s (%d/%d) ===", stepDesc, stepNum, totalSteps),
		})
		if step.Description != "" {
			uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "info",
				Content:   step.Description,
			})
		}
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   fmt.Sprintf("Command: %s", step.CmdLine),
		})
		for _, debugPath := range extractSwingbenchDebugPaths(step.CmdLine) {
			uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "info",
				Content:   fmt.Sprintf("Swingbench debug log: %s", debugPath),
			})
		}

		// Execute the step command
		var err error
		isSwingbench := isSwingbenchCommandLine(step.CmdLine)

		if isSwingbench {
			err = uc.executeCommandSwingbench(ctx, run, step)
		} else {
			err = uc.executeCommandSync(ctx, run, step)
		}

		if err != nil {
			// Step failed - log and return error
			slog.Error("Benchmark: Command sequence step failed",
				"run_id", run.ID,
				"step", stepNum,
				"total", totalSteps,
				"step_name", stepDesc,
				"error", err)
			return fmt.Errorf("step %d (%s) failed: %w", stepNum, stepDesc, err)
		}

		// Step succeeded
		slog.Info("Benchmark: Command sequence step completed",
			"run_id", run.ID,
			"step", stepNum,
			"total", totalSteps,
			"step_name", stepDesc)

		// Add success log entry
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "info",
			Content:   fmt.Sprintf("✓ %s completed", stepDesc),
		})
	}

	// All steps succeeded
	slog.Info("Benchmark: Command sequence completed successfully",
		"run_id", run.ID,
		"total_steps", len(cmd.Commands))

	return nil
}

// isRetryableError checks if an error is retryable.
// Transient errors like network timeouts, deadlocks, and connection issues are retryable.
func (uc *BenchmarkUseCase) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Network timeout errors
	retryablePatterns := []string{
		"timeout",
		"timed out",
		"deadline exceeded",
		"connection was terminated",
		"connection reset",
		"broken pipe",
		"temporarily unavailable",
		"deadlock",
		"lock wait timeout",
		"connection lost",
		"server closed the connection",
		"try again",
		"transient",
		"resource temporarily unavailable",
		"network is unreachable",
		"no route to host",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	// Check exit code for some common transient errors
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// Exit code 1 can indicate transient issues in some cases
		// We'll rely on the error message patterns above
	}

	return false
}

func detectHammerDBCommandFailure(cmdLine, stdout, stderr string) error {
	combined := strings.TrimSpace(stdout + "\n" + stderr)
	if combined == "" {
		return nil
	}

	lowerCmd := strings.ToLower(cmdLine)
	lowerOutput := strings.ToLower(combined)
	if !strings.Contains(lowerCmd, "hammerdb") && !strings.Contains(lowerOutput, "hammerdb cli") {
		return nil
	}

	for _, line := range strings.Split(combined, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "Error:") {
			return fmt.Errorf("hammerdb command reported failure: %s", trimmed)
		}
	}

	return nil
}

// calculateBackoffDelay calculates the delay for the next retry attempt using exponential backoff.
func (uc *BenchmarkUseCase) calculateBackoffDelay(retryConfig *adapter.RetryConfig, attempt int) time.Duration {
	if retryConfig == nil || attempt == 0 {
		return 0
	}

	// Calculate delay: initialDelay * (backoffFactor ^ attempt)
	delay := float64(retryConfig.InitialDelay) * float64(int(1)<<uint(attempt))
	if retryConfig.BackoffFactor != 0 {
		delay = float64(retryConfig.InitialDelay) * float64(retryConfig.BackoffFactor)
		for i := 1; i < attempt; i++ {
			delay *= float64(retryConfig.BackoffFactor)
		}
	}

	// Cap at max delay
	if delay > float64(retryConfig.MaxDelay) {
		delay = float64(retryConfig.MaxDelay)
	}

	return time.Duration(delay)
}

// startCommand starts a command and returns the process and pipes.
func (uc *BenchmarkUseCase) startCommand(ctx context.Context, cmd *adapter.Command) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	parts, err := parseCommandLine(cmd.CmdLine)
	if err != nil {
		return nil, nil, nil, err
	}
	workDir := cmd.WorkDir
	if isSwingbenchCommandLine(cmd.CmdLine) {
		parts = normalizeSwingbenchCommandParts(parts)
		workDir, err = prepareSwingbenchRuntimeDir(cmd.WorkDir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("prepare swingbench runtime dir: %w", err)
		}
	}

	execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	execCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if isSwingbenchCommandLine(cmd.CmdLine) {
		execCmd.Dir = workDir
		slog.Info("Benchmark: Setting Swingbench working directory",
			"work_dir", execCmd.Dir,
			"reason", "Swingbench needs a bin-relative runtime layout while logs stay in the run workspace")
	} else {
		execCmd.Dir = workDir
	}

	execCmd.Env = append(os.Environ(), cmd.Env...)

	// Debug: Log command execution with environment details
	hasMYSQL_PWD := false
	for _, env := range execCmd.Env {
		if strings.HasPrefix(env, "MYSQL_PWD=") {
			hasMYSQL_PWD = true
			break
		}
	}
	slog.Info("Benchmark: Starting command",
		"cmd", execCmd.String(),
		"work_dir", execCmd.Dir,
		"env_count", len(execCmd.Env),
		"has_mysql_pwd", hasMYSQL_PWD)

	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return nil, nil, nil, err
	}

	if err := execCmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return nil, nil, nil, fmt.Errorf("start command: %w", err)
	}

	return execCmd, stdout, stderr, nil
}

func prepareSwingbenchRuntimeDir(workDir string) (string, error) {
	root := filepath.Join(workDir, "swingbench")
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", err
	}

	for name, target := range swingbenchRuntimeLinks {
		linkPath := filepath.Join(root, name)
		if err := ensureSwingbenchRuntimeLink(linkPath, target); err != nil {
			return "", err
		}
	}
	if err := ensureSwingbenchRuntimeLink(filepath.Join(binDir, "data"), filepath.Join(swingbenchBinDir, "data")); err != nil {
		return "", err
	}

	return binDir, nil
}

func ensureSwingbenchRuntimeLink(linkPath, target string) error {
	info, err := os.Lstat(linkPath)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			currentTarget, readErr := os.Readlink(linkPath)
			if readErr == nil && currentTarget == target {
				return nil
			}
		}
		if removeErr := os.RemoveAll(linkPath); removeErr != nil {
			return removeErr
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return os.Symlink(target, linkPath)
}

// captureOutput captures and saves command output.
func (uc *BenchmarkUseCase) captureOutput(ctx context.Context, runID, stream string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		uc.runRepo.SaveLogEntry(ctx, runID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    stream,
			Content:   line,
		})
	}
}

func isSwingbenchCommandLine(cmdLine string) bool {
	lower := strings.ToLower(cmdLine)
	return strings.Contains(lower, "launcherbootstrap") ||
		strings.Contains(lower, "charbench") ||
		strings.Contains(lower, "oewizard") ||
		strings.Contains(lower, "minibench")
}

func extractSwingbenchDebugPaths(cmdLine string) []string {
	parts, err := parseCommandLine(cmdLine)
	if err != nil {
		return nil
	}

	var paths []string
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "-debugf" {
			candidate := strings.TrimSpace(parts[i+1])
			if candidate != "" && !strings.HasPrefix(candidate, "-") {
				paths = append(paths, candidate)
			}
		}
	}
	return paths
}

func (uc *BenchmarkUseCase) ingestSwingbenchDebugLogs(ctx context.Context, runID string, cmd *adapter.Command, startedAt time.Time) {
	if uc == nil || uc.runRepo == nil || cmd == nil || !isSwingbenchCommandLine(cmd.CmdLine) {
		return
	}

	seen := make(map[string]struct{})
	paths := append([]string(nil), extractSwingbenchDebugPaths(cmd.CmdLine)...)
	fallbackPattern := filepath.Join(swingbenchDebugSearchRoot, "debug.log*")
	if matches, err := swingbenchDebugGlob(fallbackPattern); err == nil {
		for _, match := range matches {
			if match != "" {
				paths = append(paths, match)
			}
		}
	}

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}

		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if info.ModTime().Before(startedAt.Add(-2 * time.Second)) {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		uc.runRepo.SaveLogEntry(ctx, runID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    "debug",
			Content:   fmt.Sprintf("Swingbench debug output from %s", path),
		})
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimRight(line, "\r")
			if strings.TrimSpace(line) == "" {
				continue
			}
			uc.runRepo.SaveLogEntry(ctx, runID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "debug",
				Content:   line,
			})
		}
	}
}

func normalizeSwingbenchCommandParts(parts []string) []string {
	if len(parts) == 0 {
		return nil
	}

	normalized := append([]string(nil), parts...)
	switch normalized[0] {
	case "./charbench", "charbench":
		normalized = directSwingbenchLauncherCommand("charbench", normalized[1:])
	case "./minibench", "minibench":
		normalized = directSwingbenchLauncherCommand("minibench", normalized[1:])
	case "./oewizard", "oewizard":
		normalized = directSwingbenchLauncherCommand("oewizard", normalized[1:])
	}

	for i := 0; i < len(normalized)-1; i++ {
		switch normalized[i] {
		case "-cp":
			if normalized[i+1] == "../launcher" {
				normalized[i+1] = filepath.Join(swingbenchInstallRoot, "launcher")
			}
		case "-c":
			if rewritten, ok := resolveSwingbenchConfigPath(normalized[i+1]); ok {
				normalized[i+1] = rewritten
			}
		}
	}

	return normalized
}

func directSwingbenchLauncherCommand(executable string, args []string) []string {
	normalized := []string{
		"java",
		"-cp", filepath.Join(swingbenchInstallRoot, "launcher"),
		"LauncherBootstrap",
		"-executablename", executable,
		executable,
	}
	if executable == "oewizard" && !hasSwingbenchFlag(args, "-c") {
		normalized = append(normalized, "-c", filepath.Join(swingbenchBinDir, "oewizard.xml"))
	}
	return append(normalized, args...)
}

func hasSwingbenchFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func resolveSwingbenchConfigPath(path string) (string, bool) {
	if path == "" || filepath.IsAbs(path) {
		return "", false
	}

	clean := filepath.Clean(path)
	switch {
	case clean == "oewizard.xml":
		return filepath.Join(swingbenchBinDir, clean), true
	case strings.HasPrefix(clean, "../configs/"):
		return filepath.Join(swingbenchInstallRoot, strings.TrimPrefix(clean, "../")), true
	case strings.HasPrefix(clean, "configs/"):
		return filepath.Join(swingbenchInstallRoot, clean), true
	default:
		return "", false
	}
}

func withSwingbenchResultFileHint(stdout string, tmpl *domaintemplate.Template, cmd *adapter.Command) string {
	if tmpl == nil || tmpl.Tool != domaintemplate.ToolSwingbench || cmd == nil || strings.TrimSpace(cmd.ResultFile) == "" {
		return stdout
	}
	const hintPrefix = "__DB_BENCHMIND_SWINGBENCH_RESULT_FILE__="
	if strings.Contains(stdout, hintPrefix) {
		return stdout
	}
	if stdout != "" && !strings.HasSuffix(stdout, "\n") {
		stdout += "\n"
	}
	return stdout + hintPrefix + strings.TrimSpace(cmd.ResultFile) + "\n"
}

func (uc *BenchmarkUseCase) mirrorOutputStream(ctx context.Context, runID string, stream string, source io.ReadCloser) io.ReadCloser {
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer source.Close()
		defer pipeWriter.Close()

		scanner := bufio.NewScanner(source)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), "\r")
			uc.runRepo.SaveLogEntry(ctx, runID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    stream,
				Content:   line,
			})
			if _, err := pipeWriter.Write(append([]byte(line), '\n')); err != nil {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			pipeWriter.CloseWithError(err)
		}
	}()

	return pipeReader
}

func isIgnorableSampleCollectorError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, fs.ErrClosed) || errors.Is(err, os.ErrClosed) {
		return true
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	lowerMessage := strings.ToLower(err.Error())
	return strings.Contains(lowerMessage, "file already closed") ||
		strings.Contains(lowerMessage, "use of closed file") ||
		strings.Contains(lowerMessage, "context canceled")
}

// =============================================================================
// Run Control
// Implements: REQ-EXEC-006, REQ-EXEC-007, REQ-EXEC-009
// =============================================================================

// StopBenchmark stops a running benchmark.
// Implements: REQ-EXEC-006 (graceful stop)
func (uc *BenchmarkUseCase) StopBenchmark(ctx context.Context, runID string, force bool) error {
	slog.Info("Benchmark: StopBenchmark called", "run_id", runID, "force", force)

	uc.runningProcessesMu.RLock()
	processCount := len(uc.runningProcesses)
	uc.runningProcessesMu.RUnlock()
	slog.Info("Benchmark: Current running processes", "count", processCount)

	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("get run: %w", err)
	}

	slog.Info("Benchmark: Run state", "run_id", runID, "state", run.State)

	// Check state
	if run.State != execution.StateRunning && run.State != execution.StateWarmingUp && run.State != execution.StatePreparing {
		return fmt.Errorf("%w: run is not running", ErrInvalidState)
	}

	// Get the running process and kill it
	uc.runningProcessesMu.Lock()
	process := uc.runningProcesses[runID]
	uc.runningProcessesMu.Unlock()

	slog.Info("Benchmark: Retrieved process from map", "run_id", runID, "process_found", process != nil, "process_nil", process == nil)

	if process != nil && process.Process != nil {
		slog.Info("Benchmark: Stopping process", "run_id", runID, "force", force, "pid", process.Process.Pid)

		// Send SIGTERM first (graceful shutdown)
		if err := terminateProcess(process, force); err != nil {
			slog.Error("Benchmark: Failed to send SIGTERM", "run_id", runID, "error", err)
		} else {
			slog.Info("Benchmark: SIGTERM sent successfully", "run_id", runID)
		}
	} else {
		slog.Error("Benchmark: Process not found in map or Process is nil", "run_id", runID)
	}

	if force {
		if err := uc.updateState(ctx, runID, execution.StateForceStopped); err != nil {
			return err
		}
		return uc.markRunCompletedNow(ctx, runID)
	}
	if err := uc.updateState(ctx, runID, execution.StateCancelled); err != nil {
		return err
	}
	return uc.markRunCompletedNow(ctx, runID)
}

// GetBenchmarkStatus returns the current status of a benchmark run.
func (uc *BenchmarkUseCase) GetBenchmarkStatus(ctx context.Context, runID string) (*execution.Run, error) {
	return uc.runRepo.FindByID(ctx, runID)
}

// ListBenchmarks lists benchmark runs with optional filtering.
func (uc *BenchmarkUseCase) ListBenchmarks(ctx context.Context, opts FindOptions) ([]*execution.Run, error) {
	return uc.runRepo.FindAll(ctx, opts)
}

// =============================================================================
// Helper Methods
// =============================================================================

// updateState updates the state of a run.
func (uc *BenchmarkUseCase) updateState(ctx context.Context, runID string, state execution.RunState) error {
	return uc.runRepo.UpdateState(ctx, runID, state)
}

func (uc *BenchmarkUseCase) isRunStopped(ctx context.Context, runID string) bool {
	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil || run == nil {
		return false
	}
	return run.State == execution.StateCancelled || run.State == execution.StateForceStopped
}

func (uc *BenchmarkUseCase) markRunCompletedNow(ctx context.Context, runID string) error {
	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil {
		return err
	}
	now := time.Now()
	run.CompletedAt = &now
	run.CalculateDuration()
	return uc.runRepo.Save(ctx, run)
}

// collectStandaloneReport collects and persists a report for standalone benchmark runs.
// This is called after benchmark execution completes (success or failure).
// The report collection runs in a goroutine to not block the response.
// Reports are only collected for actual benchmark runs (with results), not for
// prepare-only or cleanup-only operations.
func (uc *BenchmarkUseCase) collectStandaloneReport(
	ctx context.Context,
	run *execution.Run,
	conn connection.Connection,
	tmpl *domaintemplate.Template,
) {
	if uc.reportCollector == nil {
		return
	}

	// Reload run to get final state
	finalRun, err := uc.runRepo.FindByID(ctx, run.ID)
	if err != nil {
		slog.Warn("Benchmark: failed to reload run for report collection",
			"run_id", run.ID,
			"error", err)
		return
	}

	// Only collect report if run is in terminal state
	if !finalRun.State.IsTerminal() {
		slog.Debug("Benchmark: skipping report collection, run not in terminal state",
			"run_id", run.ID,
			"state", finalRun.State)
		return
	}

	// Skip report collection if there's no benchmark result
	// (e.g., prepare-only or cleanup-only operations)
	if finalRun.Result == nil {
		slog.Debug("Benchmark: skipping report collection, no benchmark result",
			"run_id", run.ID,
			"state", finalRun.State)
		return
	}

	rptCtx := report.ReportContext{
		SuiteID:        report.StandaloneSuiteID,
		SourceType:     report.SourceTypeBenchmark,
		ConnectionID:   conn.GetID(),
		ConnectionName: conn.GetName(),
		DatabaseType:   string(conn.GetType()),
		TemplateID:     tmpl.ID,
		TemplateName:   tmpl.Name,
	}

	// Use goroutine to not block response
	go func() {
		// Use background context to avoid cancellation when original context is done
		_, collectErr := uc.reportCollector.CollectAndPersist(context.Background(), func() (*execution.Run, error) {
			return finalRun, nil
		}, rptCtx)
		if collectErr != nil {
			slog.Warn("Benchmark: failed to collect standalone report",
				"run_id", run.ID,
				"error", collectErr)
		} else {
			slog.Info("Benchmark: collected standalone report",
				"run_id", run.ID,
				"suite_id", report.StandaloneSuiteID)
		}
	}()
}

// markAsFailed marks a run as failed with an error message.
func (uc *BenchmarkUseCase) markAsFailed(ctx context.Context, runID string, errMsg string) {
	if uc.runRepo == nil {
		return
	}
	now := time.Now()
	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil {
		return
	}

	// Update state and error message
	if err := run.SetState(execution.StateFailed); err == nil {
		run.State = execution.StateFailed
		run.ErrorMessage = errMsg
		if run.CompletedAt == nil {
			run.CompletedAt = &now
		}
		run.CalculateDuration()
		uc.runRepo.Save(ctx, run)
	}
}

// markAsCompleted marks a run as completed.
// For prepare-only and cleanup-only modes, this bypasses normal state machine validation.
func (uc *BenchmarkUseCase) markAsCompleted(ctx context.Context, runID string, duration time.Duration) {
	if uc.runRepo == nil {
		slog.Error("Benchmark: markAsCompleted failed - runRepo is nil", "run_id", runID)
		return
	}
	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil {
		slog.Error("Benchmark: markAsCompleted failed - cannot find run", "run_id", runID, "error", err)
		return
	}

	slog.Info("Benchmark: markAsCompleted called", "run_id", runID, "current_state", run.State, "duration", duration)

	now := time.Now()

	// Prepare-only completes after the prepare chain without entering the run phase.
	// Bypass the normal state machine so the run can terminate cleanly from preparing/prepared.
	if run.State == execution.StatePreparing || run.State == execution.StatePrepared {
		slog.Info("Benchmark: Prepare-only mode completed, forcing terminal completion", "run_id", runID, "current_state", run.State)
		run.State = execution.StateCompleted
		run.CompletedAt = &now
		run.Duration = &duration
		if err := uc.runRepo.Save(ctx, run); err != nil {
			slog.Error("Benchmark: markAsCompleted failed to save", "run_id", runID, "error", err)
		} else {
			slog.Info("Benchmark: markAsCompleted saved successfully (prepare-only)", "run_id", runID, "state", run.State)
		}
		return
	}

	// For cleanup-only mode: StatePending should transition to StateCompleted
	// This bypasses normal state machine validation for cleanup-only operations
	if run.State == execution.StatePending {
		slog.Info("Benchmark: Cleanup-only mode completed, forcing state transition", "run_id", runID)
		run.State = execution.StateCompleted
		run.CompletedAt = &now
		run.Duration = &duration
		if err := uc.runRepo.Save(ctx, run); err != nil {
			slog.Error("Benchmark: markAsCompleted failed to save", "run_id", runID, "error", err)
		} else {
			slog.Info("Benchmark: markAsCompleted saved successfully (forced transition)", "run_id", runID, "state", run.State)
		}
		return
	}

	// Normal path: use SetState with validation
	// For StateRunning -> StateCompleted transitions (full benchmark execution)
	if err := run.SetState(execution.StateCompleted); err == nil {
		run.State = execution.StateCompleted
		run.CompletedAt = &now
		run.Duration = &duration
		if err := uc.runRepo.Save(ctx, run); err != nil {
			slog.Error("Benchmark: markAsCompleted failed to save", "run_id", runID, "error", err)
		} else {
			slog.Info("Benchmark: markAsCompleted saved successfully", "run_id", runID, "state", run.State)
		}
	} else {
		slog.Error("Benchmark: markAsCompleted - SetState failed", "run_id", runID, "error", err)
	}
}

// checkToolAvailable checks if the benchmark tool is available.
func (uc *BenchmarkUseCase) checkToolAvailable(ctx context.Context, adapt adapter.BenchmarkAdapter) bool {
	// TODO: Implement tool availability check
	// For now, return true
	return true
}

// checkConnection checks if the database connection is working.
func (uc *BenchmarkUseCase) checkConnection(ctx context.Context, conn connection.Connection) error {
	// Use connection's Test method
	_, err := conn.Test(ctx)
	return err
}

// checkDiskSpace checks if there's enough disk space.
func (uc *BenchmarkUseCase) checkDiskSpace(path string, requiredBytes int64) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}

	// Calculate available space in bytes
	available := stat.Bavail * uint64(stat.Bsize)

	if uint64(requiredBytes) > available {
		return fmt.Errorf("insufficient disk space: need %d bytes, available %d bytes", requiredBytes, available)
	}

	return nil
}

// checkTablesExist checks if the benchmark tables exist in the database
func (uc *BenchmarkUseCase) checkTablesExist(ctx context.Context, conn connection.Connection, params map[string]interface{}) bool {
	// Get database name
	dbName := "sbtest"
	if db, ok := params["db_name"].(string); ok && db != "" {
		dbName = db
	}

	// Check based on connection type
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return uc.checkMySQLTablesExist(ctx, c, dbName)
	case *connection.PostgreSQLConnection:
		return uc.checkPostgreSQLTablesExist(ctx, c, dbName)
	default:
		// Assume tables exist for other database types
		return true
	}
}

// checkMySQLTablesExist checks if sbtest tables exist in MySQL
func (uc *BenchmarkUseCase) checkMySQLTablesExist(ctx context.Context, conn *connection.MySQLConnection, dbName string) bool {
	// Build connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		conn.Username,
		conn.Password,
		conn.Host,
		conn.Port,
		dbName)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Warn("checkMySQLTablesExist: Failed to open database", "error", err)
		return true // Assume tables exist if we can't check
	}
	defer db.Close()

	// Check if first benchmark table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'sbtest1'", dbName).Scan(&count)
	if err != nil {
		// Check if error is "database doesn't exist" (Error 1049)
		if strings.Contains(err.Error(), "Error 1049") || strings.Contains(err.Error(), "Unknown database") {
			slog.Info("checkMySQLTablesExist: Database does not exist", "database", dbName, "error", err)
			return false // Database doesn't exist, so tables don't exist
		}
		slog.Warn("checkMySQLTablesExist: Failed to query table", "error", err)
		return true // Assume tables exist for other errors
	}

	return count > 0
}

// checkPostgreSQLTablesExist checks if sbtest tables exist in PostgreSQL
func (uc *BenchmarkUseCase) checkPostgreSQLTablesExist(ctx context.Context, conn *connection.PostgreSQLConnection, dbName string) bool {
	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		conn.Host,
		conn.Port,
		dbName,
		conn.Username,
		conn.Password,
		conn.SSLMode)

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Warn("checkPostgreSQLTablesExist: Failed to open database", "error", err)
		return false // Cannot connect - assume tables don't exist
	}
	defer db.Close()

	// Ping to verify connection works
	err = db.Ping()
	if err != nil {
		slog.Warn("checkPostgreSQLTablesExist: Database ping failed", "error", err)
		// Check if error is "database does not exist"
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "3D000") {
			slog.Info("checkPostgreSQLTablesExist: Database does not exist", "database", dbName)
			return false // Database doesn't exist, so tables don't exist
		}
		return false // Connection failed for other reasons - assume tables don't exist
	}

	// Check if first benchmark table exists (PostgreSQL uses pg_tables)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = 'sbtest1'").Scan(&count)
	if err != nil {
		slog.Warn("checkPostgreSQLTablesExist: Failed to query table", "error", err)
		return false // Query failed - assume tables don't exist
	}

	slog.Info("checkPostgreSQLTablesExist: Table check result", "database", dbName, "sbtest1_exists", count > 0)
	return count > 0
}

// parseCommandLine parses a command line string into parts.
// Handles quoted strings (both single and double quotes) and backticks.
func parseCommandLine(cmdLine string) ([]string, error) {
	var parts []string
	var current strings.Builder
	var inSingleQuote, inDoubleQuote, inBacktick bool
	var escapeNext bool

	for i, r := range cmdLine {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			continue
		}

		switch r {
		case '\\':
			escapeNext = true
		case '\'':
			if !inDoubleQuote && !inBacktick {
				inSingleQuote = !inSingleQuote
			} else {
				current.WriteRune(r)
			}
		case '"':
			if !inSingleQuote && !inBacktick {
				inDoubleQuote = !inDoubleQuote
			} else {
				current.WriteRune(r)
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote {
				inBacktick = !inBacktick
			} else {
				current.WriteRune(r)
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote || inBacktick {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}

		// Check for unclosed quotes at end
		if i == len(cmdLine)-1 && (inSingleQuote || inDoubleQuote || inBacktick) {
			return nil, fmt.Errorf("unclosed quote at position %d", i)
		}
	}

	// Add last part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	// Handle empty command
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return parts, nil
}

// GetRunLogs retrieves log entries for a run.
func (uc *BenchmarkUseCase) GetRunLogs(ctx context.Context, runID string, stream string, limit int) ([]LogEntry, error) {
	// TODO: Implement log retrieval from run_logs table
	return []LogEntry{}, nil
}

// GetMetricSamples retrieves metric samples for a run.
func (uc *BenchmarkUseCase) GetMetricSamples(ctx context.Context, runID string) ([]execution.MetricSample, error) {
	return uc.runRepo.GetMetricSamples(ctx, runID)
}

// BenchmarkExecutor manages an active benchmark execution.
type BenchmarkExecutor struct {
	runID    string
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	mu       sync.Mutex
	stopping bool
}

// Stop stops the benchmark execution gracefully.
// Implements: REQ-EXEC-006
func (e *BenchmarkExecutor) Stop(force bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stopping = true

	if e.cmd == nil || e.cmd.Process == nil {
		return nil
	}

	if force {
		return e.cmd.Process.Signal(syscall.SIGKILL)
	}

	// Graceful shutdown: SIGTERM
	if err := e.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	// Wait up to 30 seconds for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- e.cmd.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(30 * time.Second):
		// Force kill after timeout
		return e.cmd.Process.Signal(syscall.SIGKILL)
	}
}

// GetStatus returns the current status.
func (e *BenchmarkExecutor) GetStatus() execution.RunState {
	// TODO: Return actual status
	return execution.StateRunning
}

// GetResult returns the final result.
func (e *BenchmarkExecutor) GetResult() (*adapter.Result, error) {
	// TODO: Parse and return result
	return &adapter.Result{}, nil
}

// parseSysbenchError parses sysbench stderr and generates a user-friendly error message.
// This function extracts the actual error from sysbench output and translates it to
// a clear action message for the user.
func (uc *BenchmarkUseCase) parseSysbenchError(stderr string) string {
	// Common sysbench error patterns and their translations
	// Order matters: more specific patterns should come first
	errorPatterns := []struct {
		pattern    string
		userMsg    string
		suggestion string
	}{
		{
			pattern:    "1146",
			userMsg:    "✗ Error: Benchmark tables do not exist",
			suggestion: "The benchmark tables do not exist (Table 'sbtest.sbtest1' doesn't exist).\n\nPlease run the Prepare phase first to create the tables and load data.\n\nGo to Task Configuration and click the '📦 Prepare' button.",
		},
		{
			pattern:    "Table.*doesn't exist",
			userMsg:    "✗ Error: Benchmark tables do not exist",
			suggestion: "The benchmark tables have not been created yet.\n\nPlease run the Prepare phase first to create the tables and load data.\n\nGo to Task Configuration and click the '📦 Prepare' button.",
		},
		{
			pattern:    "Unknown database",
			userMsg:    "✗ Error: Database does not exist",
			suggestion: "The database does not exist.\n\nPlease run the Prepare phase first to create the database and tables.\n\nGo to Task Configuration and click the '📦 Prepare' button.",
		},
		{
			pattern:    "1049",
			userMsg:    "✗ Error: Database does not exist",
			suggestion: "The database does not exist (Error 1049).\n\nPlease run the Prepare phase first to create the database and tables.\n\nGo to Task Configuration and click the '📦 Prepare' button.",
		},
		{
			pattern:    "Table.*Unknown",
			userMsg:    "✗ Error: Benchmark tables do not exist",
			suggestion: "The benchmark tables do not exist.\n\nPlease run the Prepare phase first to create the tables and load data.\n\nGo to Task Configuration and click the '📦 Prepare' button.",
		},
		{
			pattern:    "Access denied",
			userMsg:    "✗ Error: Authentication failed",
			suggestion: "Could not connect to the database - access denied.\n\nPlease check your username and password in the connection settings.\n\nGo to Connections page and verify the credentials.",
		},
		{
			pattern:    "Can't connect to MySQL server",
			userMsg:    "✗ Error: Cannot connect to database",
			suggestion: "Could not connect to the database server.\n\nPlease check:\n1. The database is running\n2. Host and port are correct\n3. Network connectivity\n\nGo to Connections page and verify the connection settings.",
		},
		{
			pattern:    "connection refused",
			userMsg:    "✗ Error: Connection refused",
			suggestion: "The database server refused the connection.\n\nPlease check:\n1. The database is running\n2. Host and port are correct\n3. Firewall settings\n\nGo to Connections page and verify the connection settings.",
		},
		{
			pattern:    "no such file or directory",
			userMsg:    "✗ Error: File not found",
			suggestion: "A required file could not be found.\n\nThis may be a configuration issue. Please check your benchmark settings.",
		},
	}

	// Try to match error patterns
	for _, ep := range errorPatterns {
		if strings.Contains(stderr, ep.pattern) {
			// Return the user-friendly message (no format strings, just direct text)
			return ep.userMsg + "\n\n" + ep.suggestion
		}
	}

	// If no pattern matched, try to extract the actual error from FATAL lines
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FATAL:") {
			// Found a FATAL error line
			if len(line) > 200 {
				line = line[:200] + "..."
			}
			return "✗ Error: " + strings.TrimPrefix(line, "FATAL:") + "\n\nCheck the logs for more details."
		}
		if line != "" && !strings.HasPrefix(line, "sysbench") && !strings.Contains(line, "Running the test") {
			// Found the actual error line
			if len(line) > 200 {
				line = line[:200] + "..."
			}
			return "✗ Error: " + line + "\n\nCheck the logs for more details."
		}
	}

	// Fallback to generic error message
	return "✗ Error: Benchmark execution failed\n\nPlease check the logs for more details."
}

func (uc *BenchmarkUseCase) parseHammerDBError(stderr string) string {
	errorOutput := strings.TrimSpace(stderr)
	lower := strings.ToLower(errorOutput)

	switch {
	case strings.Contains(lower, "login failed"), strings.Contains(lower, "18456"), strings.Contains(lower, "authentication failed"):
		return "✗ Error: SQL Server login failed\n\nHammerDB could not log in to the benchmark database.\n\nPlease check the workload credentials in the connection settings before running again."
	case strings.Contains(lower, "cannot open database"), strings.Contains(lower, "does not exist"), strings.Contains(lower, "unknown database"):
		return "✗ Error: Benchmark database does not exist\n\nHammerDB requires the benchmark database to exist before Run.\n\nPlease run the Prepare phase first."
	case strings.Contains(lower, "invalid object name"), strings.Contains(lower, "could not find stored procedure"), strings.Contains(lower, "object"), strings.Contains(lower, "table"):
		return "✗ Error: Benchmark objects are missing\n\nHammerDB could not find the required benchmark schema objects.\n\nPlease run the Prepare phase first."
	default:
		if errorOutput == "" {
			return "✗ Error: HammerDB execution failed\n\nPlease check the logs for more details."
		}
		return "✗ Error: HammerDB execution failed\n\n" + errorOutput
	}
}

// parseSwingbenchError parses Swingbench/oewizard stderr and generates a user-friendly error message.
// This function extracts Oracle/Swingbench specific errors and translates them to
// clear action messages for the user.
func (uc *BenchmarkUseCase) parseSwingbenchError(stderr, stdout string) string {
	// Combine both stderr and stdout for error detection
	// Swingbench may output errors to either stream
	errorOutput := stderr
	if errorOutput == "" {
		errorOutput = stdout
	}

	// Swingbench/oewizard error patterns
	errorPatterns := []struct {
		pattern    string
		userMsg    string
		suggestion string
	}{
		{
			pattern:    "ORA-01920", // user name 'SOE' conflicts with another user
			userMsg:    "✗ Error: SOE schema already exists",
			suggestion: "The SOE user/schema already exists in the database.\n\nIf you want to recreate the schema, run the Cleanup phase first to drop the existing schema.\n\nGo to Task Configuration and click the '🗑️ Cleanup' button, then run Prepare again.",
		},
		{
			pattern:    "schema already exists",
			userMsg:    "✗ Error: SOE schema already exists",
			suggestion: "The SOE user/schema already exists in the database.\n\nIf you want to recreate the schema, run the Cleanup phase first to drop the existing schema.\n\nGo to Task Configuration and click the '🗑️ Cleanup' button, then run Prepare again.",
		},
		{
			pattern:    "ORA-00942", // table or view does not exist
			userMsg:    "✗ Error: SOE schema tables do not exist",
			suggestion: "The SOE schema tables have not been created yet.\n\nPlease run the Prepare phase first to create the SOE user, tablespace, and tables.\n\nGo to Task Configuration and click the '📦 Prepare' button.",
		},
		{
			pattern:    "ORA-01017", // invalid username/password
			userMsg:    "✗ Error: Authentication failed",
			suggestion: "Could not connect to the database - invalid username or password.\n\nPlease check:\n1. The connection credentials are correct\n2. The SOE user exists (run Prepare first)\n\nGo to Connections page and verify the connection settings.",
		},
		{
			pattern:    "ORA-12154", // TNS:could not resolve the connect identifier
			userMsg:    "✗ Error: Cannot connect to database",
			suggestion: "Could not resolve the database connection identifier.\n\nPlease check:\n1. The connection string (SID/Service Name) is correct\n2. Host and port are correct\n\nGo to Connections page and verify the connection settings.",
		},
		{
			pattern:    "ORA-12541", // TNS:no listener
			userMsg:    "✗ Error: Database listener not running",
			suggestion: "The Oracle database listener is not running.\n\nPlease start the Oracle database and listener on the target server.",
		},
		{
			pattern:    "ORA-02236", // invalid file name
			userMsg:    "✗ Error: Tablespace datafile creation failed",
			suggestion: "Could not create the tablespace datafile.\n\nThe datafile path may be incorrect or the directory may not exist.\n\nPlease check the Oracle data directory configuration.",
		},
		{
			pattern:    "Unable to parse config file",
			userMsg:    "✗ Error: Invalid Swingbench configuration",
			suggestion: "The Swingbench configuration file is invalid.\n\nPlease check the template configuration or try a different template.",
		},
		{
			pattern:    "NumberFormatException",
			userMsg:    "✗ Error: Invalid parameter value",
			suggestion: "An invalid parameter was passed to Swingbench.\n\nPlease check the template parameters and try again.",
		},
	}

	// Try to match error patterns
	for _, ep := range errorPatterns {
		if strings.Contains(errorOutput, ep.pattern) {
			return ep.userMsg + "\n\n" + ep.suggestion
		}
	}

	// Check for ORA errors specifically
	lines := strings.Split(errorOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for ORA- errors
		if strings.Contains(line, "ORA-") {
			// Extract the ORA error code and message
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if len(line) > 300 {
					line = line[:300] + "..."
				}
				return "✗ Oracle Error: " + line + "\n\nPlease check the logs for more details."
			}
		}
		// Look for ERROR lines
		if strings.HasPrefix(line, "ERROR :") || strings.HasPrefix(line, "SEVERE:") {
			if len(line) > 300 {
				line = line[:300] + "..."
			}
			return "✗ Error: " + line + "\n\nPlease check the logs for more details."
		}
	}

	// Fallback to generic error message
	return "✗ Error: Swingbench execution failed\n\nPlease check the logs for more details."
}

func benchmarkRunPreflight(ctx context.Context, config *adapter.Config) error {
	if config == nil || config.Template == nil || config.Connection == nil {
		return nil
	}

	if err := oracleSwingbenchRunPreflight(ctx, config); err != nil {
		return err
	}

	switch {
	case config.Template.Tool == domaintemplate.ToolSysbench && config.Connection.GetType() == connection.DatabaseTypeMySQL:
		conn, ok := config.Connection.(*connection.MySQLConnection)
		if !ok {
			return nil
		}
		status, err := sysbenchMySQLRunPreflightCheck(ctx, conn, resolveSysbenchDatabaseName(config))
		return benchmarkRunPreflightError("Sysbench", status, err)
	case config.Template.Tool == domaintemplate.ToolSysbench && config.Connection.GetType() == connection.DatabaseTypePostgreSQL:
		conn, ok := config.Connection.(*connection.PostgreSQLConnection)
		if !ok {
			return nil
		}
		status, err := sysbenchPostgreSQLRunPreflightCheck(ctx, conn, resolveSysbenchDatabaseName(config))
		return benchmarkRunPreflightError("Sysbench", status, err)
	case config.Template.Tool == domaintemplate.ToolHammerDB && config.Connection.GetType() == connection.DatabaseTypeSQLServer:
		conn, ok := config.Connection.(*connection.SQLServerConnection)
		if !ok {
			return nil
		}
		status, err := hammerDBSQLServerRunPreflightCheck(ctx, conn, resolveHammerDBDatabaseName(config))
		return benchmarkRunPreflightError("HammerDB", status, err)
	default:
		return nil
	}
}

func resolveSysbenchDatabaseName(config *adapter.Config) string {
	if config == nil {
		return "sbtest"
	}
	if value, ok := config.Parameters["db_name"].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	switch conn := config.Connection.(type) {
	case *connection.MySQLConnection:
		if strings.TrimSpace(conn.Database) != "" {
			return strings.TrimSpace(conn.Database)
		}
	case *connection.PostgreSQLConnection:
		if strings.TrimSpace(conn.Database) != "" {
			return strings.TrimSpace(conn.Database)
		}
	}
	return "sbtest"
}

func resolveHammerDBDatabaseName(config *adapter.Config) string {
	if config == nil {
		return "tpcc"
	}
	if value, ok := config.Parameters["database_name"].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	if conn, ok := config.Connection.(*connection.SQLServerConnection); ok && strings.TrimSpace(conn.Database) != "" {
		return strings.TrimSpace(conn.Database)
	}
	return "tpcc"
}

func benchmarkRunPreflightError(toolName string, status benchmarkRunPreflightStatus, cause error) error {
	switch status {
	case benchmarkRunPreflightOK:
		return nil
	case benchmarkRunPreflightInvalidCredentials:
		if cause == nil {
			return fmt.Errorf("%s run failed: workload login failed. Check the benchmark credentials before running again.", toolName)
		}
		return fmt.Errorf("%s run failed: workload login failed. Check the benchmark credentials before running again. Original error: %s", toolName, cause)
	case benchmarkRunPreflightDatabaseMissing:
		return fmt.Errorf("%s run failed: benchmark database does not exist. Please run Prepare first.", toolName)
	case benchmarkRunPreflightSchemaMissing:
		return fmt.Errorf("%s run failed: benchmark tables are not prepared. Please run Prepare first.", toolName)
	case benchmarkRunPreflightCleanupInvalidated:
		return fmt.Errorf("%s run failed: Cleanup removed required benchmark objects. Please run Prepare first.", toolName)
	case benchmarkRunPreflightBenchmarkObjectsMissing:
		return fmt.Errorf("%s run failed: benchmark objects are missing. Please run Prepare first.", toolName)
	case benchmarkRunPreflightUnknownFailure:
		if cause != nil {
			return fmt.Errorf("%s run failed: unable to complete run preflight. Original error: %s", toolName, cause)
		}
		return fmt.Errorf("%s run failed: unable to complete run preflight", toolName)
	default:
		if cause != nil {
			return cause
		}
		return fmt.Errorf("%s run failed", toolName)
	}
}

func buildPostgreSQLRunPreflightDSN(conn *connection.PostgreSQLConnection, dbName string) string {
	sslMode := conn.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		conn.Host, conn.Port, dbName, conn.Username, conn.Password, sslMode)
}

func isMySQLAuthenticationError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "access denied") || strings.Contains(lower, "error 1045")
}

func isPostgreSQLAuthenticationError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "password authentication failed") || strings.Contains(lower, "28p01")
}

func isPostgreSQLDatabaseMissingError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "does not exist") || strings.Contains(lower, "3d000")
}

func isSQLServerAuthenticationError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "login failed") || strings.Contains(lower, "18456")
}

func isSQLServerDatabaseMissingError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "cannot open database") || strings.Contains(lower, "does not exist")
}

// checkTablesConfigForRun checks if existing tables match the expected configuration before run.
// This ensures run uses the data prepared by prepare phase.
// NOTE: This check is disabled for now to allow testing. Re-enable after fixing SQL execution.
func (uc *BenchmarkUseCase) checkTablesConfigForRun(ctx context.Context, run *execution.Run, conn connection.Connection, config *adapter.Config) error {
	// Get expected tables count
	expectedTables := 10 // Default
	if t, ok := config.Parameters["tables"].(int); ok {
		expectedTables = t
	}

	slog.Info("checkTablesConfigForRun: Table config check skipped (validation disabled for testing)",
		"run_id", run.ID,
		"expected_tables", expectedTables)

	// TODO: Re-enable proper table validation using database/sql instead of command line
	// For now, skip the check to allow testing the Save/OK dialog functionality
	return nil
}

func oracleSwingbenchRunPreflight(ctx context.Context, config *adapter.Config) error {
	if !isOracleSwingbenchConfig(config) {
		return nil
	}

	baseConn, ok := config.Connection.(*connection.OracleConnection)
	if !ok {
		return nil
	}

	workloadUser, workloadPassword := resolveOracleSwingbenchRunCredentials(config)
	workloadConn := cloneOracleConnectionWithCredentials(baseConn, workloadUser, workloadPassword)

	userExists, userExistsErr := oracleSwingbenchPreflightUserExists(ctx, baseConn, workloadUser)
	if userExistsErr != nil {
		return fmt.Errorf("Oracle Swingbench run failed: unable to verify SOE workload user before Run. Original error: %s", userExistsErr)
	}
	if !userExists {
		return oracleSwingbenchRunPreflightError(oracleSwingbenchRunPreflightUserMissing, nil)
	}

	if err := oracleSwingbenchPreflightPing(ctx, workloadConn); err != nil {
		if isOracleAccountLockedError(err.Error()) {
			return oracleSwingbenchRunPreflightError(oracleSwingbenchRunPreflightUserLocked, err)
		}
		if isOracleAuthenticationError(err.Error()) {
			return oracleSwingbenchRunPreflightError(oracleSwingbenchRunPreflightInvalidCredentials, err)
		}
		return fmt.Errorf("Oracle Swingbench run failed: Oracle connection/login failed while starting the workload. Original error: %s", err)
	}

	schemaReady, schemaErr := oracleSwingbenchSchemaReadyForRun(ctx, baseConn, workloadConn, workloadUser)
	if schemaErr != nil {
		return fmt.Errorf("Oracle Swingbench run failed: unable to verify required SOE workload objects before Run. Original error: %s", schemaErr)
	}
	if !schemaReady {
		return oracleSwingbenchRunPreflightError(oracleSwingbenchRunPreflightSchemaIncomplete, nil)
	}
	return nil
}

func isOracleSwingbenchConfig(config *adapter.Config) bool {
	if config == nil || config.Template == nil || config.Connection == nil {
		return false
	}
	return config.Template.Tool == domaintemplate.ToolSwingbench && config.Connection.GetType() == connection.DatabaseTypeOracle
}

func resolveOracleSwingbenchRunCredentials(config *adapter.Config) (string, string) {
	user := "soe"
	password := "soe"
	if config == nil {
		return user, password
	}
	if value, ok := config.Parameters["benchmark_user"].(string); ok && strings.TrimSpace(value) != "" {
		user = strings.TrimSpace(value)
	}
	if value, ok := config.Parameters["benchmark_password"].(string); ok && strings.TrimSpace(value) != "" {
		password = strings.TrimSpace(value)
	}
	return user, password
}

func cloneOracleConnectionWithCredentials(base *connection.OracleConnection, username, password string) *connection.OracleConnection {
	if base == nil {
		return nil
	}
	cloned := *base
	cloned.Username = username
	cloned.Password = password
	return &cloned
}

func oracleSwingbenchSchemaReadyForRun(ctx context.Context, baseConn, workloadConn *connection.OracleConnection, workloadUser string) (bool, error) {
	if workloadConn != nil {
		if ready, err := oracleSwingbenchPreflightSchemaCheck(ctx, workloadConn, workloadUser); err == nil {
			return ready, nil
		}
	}
	if baseConn == nil {
		return false, fmt.Errorf("no Oracle connection available for schema preflight")
	}
	return oracleSwingbenchPreflightSchemaCheck(ctx, baseConn, workloadUser)
}

func oracleSwingbenchSchemaReady(requiredTableCount, orderEntryPackageCount, sequenceCount int) bool {
	return requiredTableCount == 4 && orderEntryPackageCount == 2 && sequenceCount > 0
}

func oracleSwingbenchSchemaCountValue(value sql.NullInt64) int {
	if !value.Valid {
		return 0
	}
	return int(value.Int64)
}

func isOracleAuthenticationError(message string) bool {
	upper := strings.ToUpper(message)
	lower := strings.ToLower(message)
	return strings.Contains(upper, "ORA-01017") ||
		strings.Contains(lower, "invalid username/password") ||
		strings.Contains(lower, "logon denied")
}

func isOracleAccountLockedError(message string) bool {
	upper := strings.ToUpper(message)
	lower := strings.ToLower(message)
	return strings.Contains(upper, "ORA-28000") || strings.Contains(lower, "account is locked")
}

func oracleSwingbenchRunPreflightError(status oracleSwingbenchRunPreflightStatus, cause error) error {
	switch status {
	case oracleSwingbenchRunPreflightUserMissing:
		return fmt.Errorf("Oracle Swingbench run failed: SOE workload user does not exist.\nRun Prepare first or recreate Swingbench schema.")
	case oracleSwingbenchRunPreflightUserLocked:
		if cause == nil {
			return fmt.Errorf("Oracle Swingbench run failed: SOE workload user is locked.")
		}
		return fmt.Errorf("Oracle Swingbench run failed: SOE workload user is locked. Original error: %s", cause)
	case oracleSwingbenchRunPreflightInvalidCredentials:
		if cause == nil {
			return fmt.Errorf("Oracle Swingbench run failed: Invalid SOE workload username/password.")
		}
		return fmt.Errorf("Oracle Swingbench run failed: Invalid SOE workload username/password. Original error: %s", cause)
	case oracleSwingbenchRunPreflightSchemaIncomplete:
		return fmt.Errorf("Oracle Swingbench run failed: Run requires prepared SOE schema. Please run Prepare first.")
	case oracleSwingbenchRunPreflightCleanupInvalidated:
		return fmt.Errorf("Oracle Swingbench run failed: Cleanup removed required SOE objects. Please run Prepare first.")
	case oracleSwingbenchRunPreflightOK:
		return nil
	default:
		if cause != nil {
			return cause
		}
		return fmt.Errorf("Oracle Swingbench run failed")
	}
}

func oracleSwingbenchHistoryInvalidatesDirectRun(runs []*execution.Run, connectionName, databaseType, tool, workloadFamily string) (bool, error) {
	_ = tool
	_ = workloadFamily

	latestPhase := ""
	latestAt := time.Time{}
	for _, run := range runs {
		if run == nil || run.Result == nil {
			continue
		}
		if !strings.EqualFold(run.Result.ConnectionName, connectionName) || !strings.EqualFold(run.Result.DatabaseType, databaseType) {
			continue
		}

		lowerMessage := strings.ToLower(strings.TrimSpace(run.Message))
		phase := ""
		switch {
		case strings.Contains(lowerMessage, "phase=cleanup"):
			phase = "cleanup"
		case strings.Contains(lowerMessage, "phase=prepare"):
			phase = "prepare"
		default:
			continue
		}

		at := run.CreatedAt
		if run.CompletedAt != nil {
			at = *run.CompletedAt
		}
		if latestAt.IsZero() || !at.Before(latestAt) {
			latestAt = at
			latestPhase = phase
		}
	}

	return latestPhase == "cleanup", nil
}

func oracleSwingbenchZeroThroughputFailure(stdout string, finalResult *adapter.FinalResult) error {
	if finalResult == nil {
		return nil
	}
	if finalResult.TotalTransactions > 0 || finalResult.TransactionsPerSec > 0 || finalResult.AvgTPS > 0 || finalResult.AvgTPM > 0 {
		return nil
	}

	zeroUserSamples := 0
	for _, line := range strings.Split(stdout, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "[0/") && strings.Contains(trimmed, " 0") {
			zeroUserSamples++
		}
	}
	if zeroUserSamples < 3 {
		return nil
	}

	return fmt.Errorf("Oracle Swingbench run failed: workload never advanced beyond zero active users and zero TPS/TPM (%d stalled samples, e.g. [0/N]). Cleanup removes workload objects, so direct Run must be preceded by Prepare. Run Prepare first.", zeroUserSamples)
}

func terminateProcess(cmd *exec.Cmd, force bool) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	target := cmd.Process.Pid
	if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil && pgid > 0 {
		target = -pgid
	}

	if err := syscall.Kill(target, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	if force {
		time.Sleep(200 * time.Millisecond)
		if err := syscall.Kill(target, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			return err
		}
	}
	return nil
}

func swingbenchNoOutputTimeoutForStep(step *adapter.Command) time.Duration {
	return 0
}

func shouldCleanupResidualSwingbenchProcesses(step *adapter.Command) bool {
	if step == nil {
		return false
	}
	lower := strings.ToLower(step.CmdLine)
	return strings.Contains(lower, "launcherbootstrap") && strings.Contains(lower, "oewizard")
}

func (uc *BenchmarkUseCase) cleanupResidualSwingbenchProcesses(ctx context.Context, runID string) {
	cmd := exec.CommandContext(ctx, "bash", "-lc", "pkill -TERM -f 'LauncherBootstrap.*oewizard|com\\.dom\\.benchmarking\\.swingbench\\.wizards\\.Wizard' >/dev/null 2>&1 || true; sleep 1; pkill -KILL -f 'LauncherBootstrap.*oewizard|com\\.dom\\.benchmarking\\.swingbench\\.wizards\\.Wizard' >/dev/null 2>&1 || true")
	if err := cmd.Run(); err != nil {
		slog.Warn("Benchmark: Failed to cleanup residual Swingbench processes", "run_id", runID, "error", err)
		return
	}
	slog.Info("Benchmark: Cleaned up residual Swingbench processes", "run_id", runID)
}

func tickerChan(t *time.Ticker) <-chan time.Time {
	if t == nil {
		return nil
	}
	return t.C
}
