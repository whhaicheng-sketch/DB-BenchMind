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
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
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

		// Prepare phase
		// For prepare-only mode, we bypass executePhase to avoid StatePrepared
		// and go directly to StateCompleted
		slog.Info("Benchmark: Executing prepare phase (prepare-only mode)", "run_id", run.ID)

		cmd, err := adapt.BuildPrepareCommand(ctx, config)
		if err != nil {
			uc.markAsFailed(ctx, run.ID, fmt.Sprintf("build prepare command: %v", err))
			return
		}

		if err := uc.executeCommand(ctx, run, cmd); err != nil {
			// Check if error is "table already exists" (MySQL error 1050)
			errMsg := err.Error()
			slog.Info("Benchmark: Prepare command failed, checking error type", "run_id", run.ID, "error", errMsg)

			if strings.Contains(errMsg, "1050") || strings.Contains(errMsg, "already exists") ||
				strings.Contains(errMsg, "Duplicate key") || strings.Contains(errMsg, "Table.*already exists") ||
				strings.Contains(errMsg, "Table '") && strings.Contains(errMsg, "already exists") {
				slog.Info("Benchmark: Prepare phase - data already exists, treating as success",
					"error", err, "run_id", run.ID)

				// Set user-friendly message for UI popup
				run.Message = "✓ Table data already exists\n\nThe benchmark tables are already prepared and ready to use."
				uc.runRepo.Save(ctx, run)

				// Save log entries
				msg1 := "✓ Table data already exists - skipping prepare phase"
				msg2 := "Info: The benchmark tables are already prepared and ready to use."
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
				// Data already exists, this is OK for prepare phase - continue to mark as completed
			} else {
				uc.markAsFailed(ctx, run.ID, fmt.Sprintf("prepare: %v", err))
				return
			}
		} else {
			// Prepare completed successfully
			msg1 := "✓ Prepare phase completed successfully"
			msg2 := "Info: All tables created and data loaded successfully."
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
		}

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
			// Check if error is "table already exists" (MySQL error 1050)
			// This is OK - means data was already prepared, we can continue
			if strings.Contains(err.Error(), "1050") || strings.Contains(err.Error(), "already exists") {
				slog.Warn("Benchmark: Prepare phase failed with 'table already exists', continuing",
					"error", err, "run_id", run.ID)
				// Continue to run phase anyway
				uc.updateState(ctx, run.ID, execution.StatePrepared)
			} else {
				// For other errors, fail the benchmark
				uc.markAsFailed(ctx, run.ID, fmt.Sprintf("prepare: %v", err))
				return
			}
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

	var cmd *adapter.Command
	var err error

	switch phase {
	case "prepare":
		cmd, err = adapt.BuildPrepareCommand(ctx, config)
	case "cleanup":
		cmd, err = adapt.BuildCleanupCommand(ctx, config)
	default:
		return fmt.Errorf("unknown phase: %s", phase)
	}

	if err != nil {
		return fmt.Errorf("build %s command: %w", phase, err)
	}

	slog.Info("Benchmark: Executing phase command",
		"phase", phase,
		"cmd", cmd.CmdLine,
		"run_id", run.ID)

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
	// Update state
	uc.updateState(ctx, run.ID, execution.StateRunning)

	// Update started_at
	now := time.Now()
	run.StartedAt = &now
	uc.runRepo.Save(ctx, run)

	// Build run command
	cmd, err := adapt.BuildRunCommand(ctx, config)
	if err != nil {
		return err
	}

	// Create context with timeout
	runCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Start command
	process, stdout, _, err := uc.startCommand(runCtx, cmd)
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

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

	// We'll read stderr after process completes
	// Don't close stderr here - we'll read it after process.Wait()
	defer stdout.Close()

	// Start realtime collection from stdout only
	sampleCh, errCh, stdoutBuf := adapt.StartRealtimeCollection(runCtx, stdout)

	// Monitor process
	done := make(chan error, 1)
	go func() {
		done <- process.Wait()
	}()

	// Collect samples and monitor for completion
	for {
		select {
		case sample, ok := <-sampleCh:
			if !ok {
				// Channel closed - wait briefly for any remaining samples to be processed
				// This ensures the final second's data is captured before we exit
				slog.Info("Benchmark: Sample channel closed, waiting for final samples", "run_id", run.ID)
				time.Sleep(500 * time.Millisecond)

				// Now wait for process to complete
				processErr := <-done
				if processErr != nil {
					errMsg := processErr.Error()
					slog.Info("Benchmark: Run process failed", "run_id", run.ID, "error", errMsg)

					// Check if tables exist by querying the database
					// This is more reliable than parsing stderr
					tablesExist := uc.checkTablesExist(ctx, config.Connection, config.Parameters)

					if !tablesExist {
						// Table does not exist - set user-friendly message
						slog.Info("Benchmark: Run phase - tables do not exist", "run_id", run.ID)
						run.Message = "✗ Error: Benchmark tables do not exist\n\nPlease run the Prepare phase first to create the tables and load data.\n\nGo to Task Configuration and click the '📦 Prepare' button."
						uc.runRepo.Save(ctx, run)

						// Save error to logs
						uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
							Timestamp: time.Now().Format(time.RFC3339),
							Stream:    "error",
							Content:   "============================================================",
						})
						uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
							Timestamp: time.Now().Format(time.RFC3339),
							Stream:    "error",
							Content:   run.Message,
						})
						uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
							Timestamp: time.Now().Format(time.RFC3339),
							Stream:    "error",
							Content:   "============================================================",
						})
					}
					return fmt.Errorf("process error: %w", processErr)
				}

				// Process completed successfully, parse final results
				slog.Info("Benchmark: Process completed successfully, parsing final results", "run_id", run.ID)
				finalResult, err := adapt.ParseFinalResults(ctx, stdoutBuf.String())
				slog.Info("Benchmark: ParseFinalResults returned", "run_id", run.ID, "err", err, "finalResult_nil", finalResult == nil)
				if err != nil {
					slog.Error("Benchmark: Failed to parse final results", "run_id", run.ID, "error", err)
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

					// Get threads count from parameters
					threads := 0
					if t, ok := config.Parameters["threads"].(int); ok {
						threads = t
					}

					// Convert finalResult to BenchmarkResult and save to run
					slog.Info("Benchmark: Creating BenchmarkResult", "run_id", run.ID)
					result := &execution.BenchmarkResult{
						RunID:             run.ID,
						TPSCalculated:     finalResult.TransactionsPerSec,
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
					LatencyAvg: sample.LatencyAvg,
					LatencyP95: sample.LatencyP95,
					LatencyP99: sample.LatencyP99,
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
			if err != nil {
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
			finalResult, err := adapt.ParseFinalResults(ctx, stdoutBuf.String())
			if err != nil {
				slog.Warn("Benchmark: Failed to parse final results", "run_id", run.ID, "error", err)
			} else {
				// Save final results to run
				slog.Info("Benchmark: Final results parsed",
					"run_id", run.ID,
					"tps", finalResult.TransactionsPerSec,
					"qps", finalResult.QueriesPerSec,
					"latency_avg", finalResult.LatencyAvg,
					"latency_p95", finalResult.LatencyP95)
				// TODO: Save finalResult to run object or database
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
func (uc *BenchmarkUseCase) executeCommand(ctx context.Context, run *execution.Run, cmd *adapter.Command) error {
	// Parse command line
	parts, err := parseCommandLine(cmd.CmdLine)
	if err != nil {
		return err
	}

	// Create command
	execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	execCmd.Dir = cmd.WorkDir
	execCmd.Env = append(os.Environ(), cmd.Env...)

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
	slog.Info("Benchmark: === EXECUTING COMMAND ===",
		"run_id", run.ID,
		"binary", parts[0],
		"arguments", parts[1:],
		"work_dir", execCmd.Dir,
		"env_count", len(execCmd.Env),
		"has_mysql_pwd", hasMYSQL_PWD,
		"has_pgpassword", hasPGPASSWORD)

	// Use CombinedOutput to capture both stdout and stderr
	// This avoids the race condition of reading both pipes concurrently
	output, err := execCmd.CombinedOutput()

	// Split output into lines and save to repository
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Determine if this is stderr (error messages) by checking content
		stream := "stdout"
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "error") ||
			strings.Contains(lineLower, "failed") ||
			strings.Contains(lineLower, "fatal") ||
			strings.Contains(lineLower, "warning") {
			stream = "stderr"
		}

		// Save to repository
		uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Stream:    stream,
			Content:   line,
		})

		// Also log important messages to slog
		if stream == "stderr" {
			slog.Info("Benchmark: command output", "run_id", run.ID, "stream", stream, "line", line)
		}
	}

	// If command failed, return error with output
	if err != nil {
		slog.Error("Benchmark: Command failed", "run_id", run.ID, "exit_error", err, "output", string(output))
		// Return error that includes output information
		return fmt.Errorf("command failed with exit status %v: %w", err, fmt.Errorf("output:\n%s", string(output)))
	}

	return nil
}

// startCommand starts a command and returns the process and pipes.
func (uc *BenchmarkUseCase) startCommand(ctx context.Context, cmd *adapter.Command) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	parts, err := parseCommandLine(cmd.CmdLine)
	if err != nil {
		return nil, nil, nil, err
	}

	execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	execCmd.Dir = cmd.WorkDir
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
	if run.State != execution.StateRunning && run.State != execution.StateWarmingUp {
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
		if err := process.Process.Signal(syscall.SIGTERM); err != nil {
			slog.Error("Benchmark: Failed to send SIGTERM", "run_id", runID, "error", err)
		} else {
			slog.Info("Benchmark: SIGTERM sent successfully", "run_id", runID)
		}

		// If force stopping, wait a bit then send SIGKILL if needed
		if force {
			time.Sleep(2 * time.Second)
			if err := process.Process.Signal(syscall.SIGKILL); err != nil {
				slog.Error("Benchmark: Failed to send SIGKILL", "run_id", runID, "error", err)
			} else {
				slog.Info("Benchmark: SIGKILL sent successfully", "run_id", runID)
			}
		}
	} else {
		slog.Error("Benchmark: Process not found in map or Process is nil", "run_id", runID)
	}

	if force {
		return uc.updateState(ctx, runID, execution.StateForceStopped)
	}
	return uc.updateState(ctx, runID, execution.StateCancelled)
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

	// For prepare-only and cleanup-only modes, we bypass normal state machine
	// because StatePending cannot directly transition to StateCompleted
	// We force the state transition for these special cases
	if run.State == execution.StatePending {
		slog.Info("Benchmark: Forcing state transition from pending to completed (prepare/cleanup-only mode)", "run_id", runID)
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
		slog.Warn("checkMySQLTablesExist: Failed to query table", "error", err)
		return true // Assume tables exist if query fails
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
