// Package usecase provides benchmark execution business logic.
// Implements: REQ-EXEC-001 ~ REQ-EXEC-010
package usecase

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
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

// BenchmarkUseCase provides benchmark execution business operations.
// Implements: REQ-EXEC-001 ~ REQ-EXEC-010
type BenchmarkUseCase struct {
	runRepo       RunRepository
	adapterReg    *adapter.AdapterRegistry
	connUseCase   *ConnectionUseCase
	templateUseCase *TemplateUseCase
}

// NewBenchmarkUseCase creates a new benchmark use case.
func NewBenchmarkUseCase(
	runRepo RunRepository,
	adapterReg *adapter.AdapterRegistry,
	connUseCase *ConnectionUseCase,
	templateUseCase *TemplateUseCase,
) *BenchmarkUseCase {
	return &BenchmarkUseCase{
		runRepo:         runRepo,
		adapterReg:      adapterReg,
		connUseCase:     connUseCase,
		templateUseCase: templateUseCase,
	}
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
		Options:    execution.TaskOptions{}, // TODO: Get from task
		WorkDir:    run.WorkDir,
	}

	// Run pre-checks
	if err := uc.preChecks(ctx, run, adapt, config); err != nil {
		uc.markAsFailed(ctx, run.ID, fmt.Sprintf("pre-check: %v", err))
		return
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
	if err := uc.executeRun(ctx, run, adapt, config, task.Options.RunTimeout); err != nil {
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
		return err
	}

	// Execute command
	if err := uc.executeCommand(ctx, run, cmd); err != nil {
		return err
	}

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
// Implements: REQ-EXEC-002, REQ-EXEC-004
func (uc *BenchmarkUseCase) executeRun(
	ctx context.Context,
	run *execution.Run,
	adapt adapter.BenchmarkAdapter,
	config *adapter.Config,
	timeout time.Duration,
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
	process, stdout, stderr, err := uc.startCommand(runCtx, cmd)
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}
	defer process.Wait()
	defer stdout.Close()
	defer stderr.Close()

	// Start realtime collection
	sampleCh, errCh := adapt.StartRealtimeCollection(runCtx, stdout, stderr)

	// Monitor process
	done := make(chan error, 1)
	go func() {
		done <- process.Wait()
	}()

	// Collect samples and monitor for completion
	for {
		select {
		case sample := <-sampleCh:
			// Save metric sample
			metricSample := execution.MetricSample{
				Timestamp:   sample.Timestamp,
				Phase:       "run",
				TPS:         sample.TPS,
				LatencyAvg:  sample.LatencyAvg,
				LatencyP95:  sample.LatencyP95,
				LatencyP99:  sample.LatencyP99,
				ErrorRate:   sample.ErrorRate,
			}
			uc.runRepo.SaveMetricSample(ctx, run.ID, metricSample)

		case err := <-errCh:
			// Log error
			uc.runRepo.SaveLogEntry(ctx, run.ID, LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Stream:    "stderr",
				Content:   err.Error(),
			})

		case err := <-done:
			if err != nil {
				return fmt.Errorf("process error: %w", err)
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

	// Get pipes
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start command
	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	// Capture output
	go uc.captureOutput(ctx, run.ID, "stdout", stdout)
	go uc.captureOutput(ctx, run.ID, "stderr", stderr)

	// Wait for completion
	return execCmd.Wait()
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
	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("get run: %w", err)
	}

	// Check state
	if run.State != execution.StateRunning && run.State != execution.StateWarmingUp {
		return fmt.Errorf("%w: run is not running", ErrInvalidState)
	}

	// TODO: Send SIGTERM to process
	// For now, just update state
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
func (uc *BenchmarkUseCase) markAsCompleted(ctx context.Context, runID string, duration time.Duration) {
	if uc.runRepo == nil {
		return
	}
	run, err := uc.runRepo.FindByID(ctx, runID)
	if err != nil {
		return
	}

	now := time.Now()
	if err := run.SetState(execution.StateCompleted); err == nil {
		run.State = execution.StateCompleted
		run.CompletedAt = &now
		run.Duration = &duration
		uc.runRepo.Save(ctx, run)
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

// parseCommandLine parses a command line string into parts.
func parseCommandLine(cmdLine string) ([]string, error) {
	// Simple implementation - split on spaces
	// TODO: Use proper shell parsing for quoted strings
	return strings.Fields(cmdLine), nil
}

// GetRunLogs retrieves log entries for a run.
func (uc *BenchmarkUseCase) GetRunLogs(ctx context.Context, runID string, stream string, limit int) ([]LogEntry, error) {
	// TODO: Implement log retrieval from run_logs table
	return []LogEntry{}, nil
}

// GetMetricSamples retrieves metric samples for a run.
func (uc *BenchmarkUseCase) GetMetricSamples(ctx context.Context, runID string) ([]execution.MetricSample, error) {
	// TODO: Implement metric sample retrieval
	return []execution.MetricSample{}, nil
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
