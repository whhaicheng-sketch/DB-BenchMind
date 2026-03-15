package bindings

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintask "github.com/whhaicheng/DB-BenchMind/internal/domain/task"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/collector"
)

const taskLogTailLimit = 500

type TaskBinding struct {
	ctx         context.Context
	benchmarkUC *usecase.BenchmarkUseCase
	connUC      *usecase.ConnectionUseCase
	templateUC  *usecase.TemplateUseCase
	runRepo     *usecase.MemoryRunRepository

	mu           sync.RWMutex
	tasks        map[string]*domaintask.ExecutionTask
	previews     map[string]*domaintask.ExecutionTask
	activeTaskID string
	executions   map[string]*taskExecutionContext
}

type taskExecutionContext struct {
	currentRunID  string
	logSeen       map[string]int
	stopRequested bool
	sshCollector  *collector.SSHMetricsCollector
	connection    connection.Connection
}

type TaskDraftRequest struct {
	TaskName     string                 `json:"task_name"`
	DatabaseType string                 `json:"database_type,omitempty"`
	TemplateID   string                 `json:"template_id"`
	ConnectionID string                 `json:"connection_id"`
	Action       string                 `json:"action"`
	PreviewToken string                 `json:"preview_token,omitempty"`
	Overrides    map[string]interface{} `json:"overrides"`
}

type TaskLogsRequest struct {
	TaskID string `json:"task_id"`
	Limit  int    `json:"limit"`
	Query  string `json:"query"`
	Phase  string `json:"phase"`
}

type TaskListResult struct {
	Tasks []domaintask.ExecutionTask `json:"tasks"`
	Error string                     `json:"error,omitempty"`
}

type TaskResult struct {
	Task  *domaintask.ExecutionTask `json:"task,omitempty"`
	Error string                    `json:"error,omitempty"`
}

type TaskLogsResult struct {
	Lines []domaintask.LogLine `json:"lines"`
	Error string               `json:"error,omitempty"`
}

type TaskActionResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func NewTaskBinding(
	benchmarkUC *usecase.BenchmarkUseCase,
	connUC *usecase.ConnectionUseCase,
	templateUC *usecase.TemplateUseCase,
	runRepo *usecase.MemoryRunRepository,
) *TaskBinding {
	return &TaskBinding{
		benchmarkUC: benchmarkUC,
		connUC:      connUC,
		templateUC:  templateUC,
		runRepo:     runRepo,
		tasks:       make(map[string]*domaintask.ExecutionTask),
		previews:    make(map[string]*domaintask.ExecutionTask),
		executions:  make(map[string]*taskExecutionContext),
	}
}

func (b *TaskBinding) SetContext(ctx context.Context) {
	b.ctx = ctx
}

func (b *TaskBinding) ValidateDraft(req TaskDraftRequest) TaskResult {
	task, err := b.buildTask(req)
	if err != nil {
		return TaskResult{Error: err.Error()}
	}
	readiness := b.evaluateReadiness(task)
	task.Readiness = readiness
	if readiness.DBValid {
		appendTaskEvent(task, domaintask.PhaseNone, "DB validation passed")
	} else {
		appendTaskEvent(task, domaintask.PhaseNone, fmt.Sprintf("DB validation failed: %s", readiness.DBMessage))
	}
	if readiness.SSHAvailable {
		appendTaskEvent(task, domaintask.PhaseNone, readiness.SSHMessage)
	} else if readiness.SSHMessage != "" {
		appendTaskEvent(task, domaintask.PhaseNone, readiness.SSHMessage)
	}
	if hint := preparePrivilegeHint(task); hint != "" {
		appendTaskEvent(task, domaintask.PhaseNone, hint)
	}
	b.mu.Lock()
	b.previews[task.PreviewToken] = cloneTask(task)
	b.mu.Unlock()
	return TaskResult{Task: task}
}

func (b *TaskBinding) CreateTask(req TaskDraftRequest) TaskResult {
	if strings.TrimSpace(req.PreviewToken) == "" {
		return TaskResult{Error: "preview confirmation required before creating task"}
	}
	b.mu.Lock()
	task, ok := b.previews[req.PreviewToken]
	if !ok {
		b.mu.Unlock()
		return TaskResult{Error: "preview expired, please reopen preview"}
	}
	delete(b.previews, req.PreviewToken)
	task = cloneTask(task)
	task.Readiness = b.evaluateReadiness(task)
	if err := validateReadiness(task.Readiness); err != nil {
		b.mu.Unlock()
		return TaskResult{Error: err.Error()}
	}
	defer b.mu.Unlock()
	if b.activeTaskID != "" {
		return TaskResult{Error: "an active task is already running"}
	}
	appendTaskEvent(task, domaintask.PhaseNone, "Preview confirmed")
	appendTaskEvent(task, domaintask.PhaseNone, fmt.Sprintf("Task created: %s", task.Name))
	task.Status = domaintask.StatusStarting
	b.activeTaskID = task.ID
	b.tasks[task.ID] = task
	b.executions[task.ID] = &taskExecutionContext{logSeen: make(map[string]int)}
	go b.executeTask(task.ID)
	return TaskResult{Task: cloneTask(task)}
}

func (b *TaskBinding) ListTasks() TaskListResult {
	b.mu.RLock()
	defer b.mu.RUnlock()
	tasks := make([]domaintask.ExecutionTask, 0, len(b.tasks))
	now := time.Now()
	for _, task := range b.tasks {
		cloned := cloneTask(task)
		syncTaskTiming(cloned, now)
		tasks = append(tasks, *cloned)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
	return TaskListResult{Tasks: tasks}
}

func (b *TaskBinding) CancelTask(taskID string) TaskActionResult {
	return TaskActionResult{Error: "queue is disabled in single-task mode"}
}

func (b *TaskBinding) StopTask(taskID string) TaskActionResult {
	b.mu.Lock()
	task, ok := b.tasks[taskID]
	execCtx := b.executions[taskID]
	if !ok || execCtx == nil {
		b.mu.Unlock()
		return TaskActionResult{Error: "task not running"}
	}
	task.Status = domaintask.StatusStopping
	execCtx.stopRequested = true
	appendTaskEvent(task, task.CurrentPhase, "Stop requested")
	runID := execCtx.currentRunID
	b.mu.Unlock()
	if runID == "" {
		return TaskActionResult{Success: true}
	}
	if err := b.benchmarkUC.StopBenchmark(context.Background(), runID, false); err != nil {
		return TaskActionResult{Error: err.Error()}
	}
	return TaskActionResult{Success: true}
}

func (b *TaskBinding) GetTaskLogs(req TaskLogsRequest) TaskLogsResult {
	b.mu.RLock()
	task, ok := b.tasks[req.TaskID]
	b.mu.RUnlock()
	if !ok {
		return TaskLogsResult{Error: "task not found"}
	}
	limit := req.Limit
	if limit <= 0 {
		limit = taskLogTailLimit
	}
	lines := b.collectTaskLogs(task)
	if len(lines) == 0 {
		lines = append([]domaintask.LogLine(nil), task.LogTail...)
	}
	lines = filterLogLines(lines, req.Query, req.Phase)
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return TaskLogsResult{Lines: lines}
}

func (b *TaskBinding) collectTaskLogs(task *domaintask.ExecutionTask) []domaintask.LogLine {
	lines := make([]domaintask.LogLine, 0, len(task.LogTail))
	seenRuns := make(map[string]struct{})
	for _, record := range task.PhaseHistory {
		if record.RunID == "" {
			continue
		}
		if _, ok := seenRuns[record.RunID]; ok {
			continue
		}
		seenRuns[record.RunID] = struct{}{}
		entries, err := b.runRepo.GetPersistedLogs(record.RunID)
		if err != nil || len(entries) == 0 {
			entries = b.runRepo.GetLogs(record.RunID)
		}
		for _, entry := range entries {
			lines = append(lines, domaintask.LogLine{
				Timestamp: entry.Timestamp,
				Phase:     record.Phase,
				Stream:    entry.Stream,
				Content:   entry.Content,
			})
		}
	}
	for _, line := range task.LogTail {
		if line.Stream != "event" {
			continue
		}
		lines = append(lines, line)
	}
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].Timestamp < lines[j].Timestamp
	})
	return lines
}

func (b *TaskBinding) buildTask(req TaskDraftRequest) (*domaintask.ExecutionTask, error) {
	action := domaintask.Action(req.Action)
	if action == "" {
		action = domaintask.ActionFullPipeline
	}
	tmpl, err := b.templateUC.GetTemplate(context.Background(), req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	conn, err := b.connUC.GetConnectionByID(context.Background(), req.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	if req.DatabaseType != "" && !strings.EqualFold(req.DatabaseType, string(conn.GetType())) {
		return nil, fmt.Errorf("connection does not match selected database type")
	}
	if req.DatabaseType != "" && !tmpl.SupportsDatabase(req.DatabaseType) {
		return nil, fmt.Errorf("template does not support selected database type")
	}
	if !tmpl.SupportsDatabase(string(conn.GetType())) {
		return nil, fmt.Errorf("template does not support selected connection type")
	}
	params, err := resolveParams(tmpl, req.Overrides)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.TaskName)
	if name == "" {
		name = fmt.Sprintf("%s-%s", tmpl.Name, time.Now().Format("20060102-150405"))
	}
	task := &domaintask.ExecutionTask{
		ID:                 uuid.NewString(),
		PreviewToken:       uuid.NewString(),
		Name:               name,
		Action:             action,
		Status:             domaintask.StatusStarting,
		CurrentPhase:       domaintask.PhaseNone,
		TemplateSnapshot:   templateSnapshot(tmpl, params),
		ConnectionSnapshot: connectionSnapshot(conn),
		ResolvedParams:     params,
		BenchmarkTool:      tmpl.Tool,
		CreatedAt:          time.Now(),
		RunLogPaths:        make(map[string]string),
		SystemLogPaths:     make(map[string]string),
	}
	syncTaskTiming(task, task.CreatedAt)
	return task, nil
}

func (b *TaskBinding) evaluateReadiness(task *domaintask.ExecutionTask) domaintask.Readiness {
	readiness := domaintask.Readiness{
		TemplateSelected:   task.TemplateSnapshot.ID != "",
		ConnectionSelected: task.ConnectionSnapshot.ID != "",
		ActionSupported:    actionSupported(task.Action, task.TemplateSnapshot.Phases),
		RuntimeValid:       task.ResolvedParams != nil,
	}
	if readiness.ConnectionSelected {
		if result, err := b.connUC.TestConnection(context.Background(), task.ConnectionSnapshot.ID); err != nil {
			readiness.DBMessage = err.Error()
		} else if result != nil {
			readiness.DBValid = result.Success
			if result.Success {
				readiness.DBMessage = fmt.Sprintf("DB ok (%d ms)", result.LatencyMs)
			} else {
				readiness.DBMessage = result.Error
			}
		}
	}
	conn, err := b.connUC.GetConnectionByID(context.Background(), task.ConnectionSnapshot.ID)
	if err == nil {
		if sshConfig := sshConfigFromConnection(conn); sshConfig == nil || !sshConfig.Enabled {
			readiness.SSHChecked = true
			readiness.SSHAvailable = false
			readiness.SSHMessage = "SSH required"
		} else if ok, latency, sshErr := connection.TestSSHConnection(context.Background(), sshConfig); sshErr != nil {
			readiness.SSHChecked = true
			readiness.SSHAvailable = false
			readiness.SSHMessage = fmt.Sprintf("SSH unavailable: %v", sshErr)
		} else {
			readiness.SSHChecked = true
			readiness.SSHAvailable = ok
			readiness.SSHMessage = fmt.Sprintf("SSH ok (%d ms)", latency)
		}
	}
	return readiness
}

func (b *TaskBinding) executeTask(taskID string) {
	b.mu.Lock()
	task := b.tasks[taskID]
	execCtx := b.executions[taskID]
	if task == nil || execCtx == nil {
		b.mu.Unlock()
		return
	}
	now := time.Now()
	task.StartedAt = &now
	syncTaskTiming(task, now)
	task.Status = domaintask.StatusStarting
	task.Readiness = b.evaluateReadiness(task)
	appendTaskEvent(task, domaintask.PhaseNone, "Task execution started")
	b.mu.Unlock()

	if !task.Readiness.DBValid {
		b.mu.Lock()
		appendTaskEvent(task, domaintask.PhaseNone, fmt.Sprintf("DB validation failed: %s", task.Readiness.DBMessage))
		b.mu.Unlock()
		b.completeTask(taskID, domaintask.StatusFailed, domaintask.PhaseNone, task.Readiness.DBMessage)
		return
	}
	b.mu.Lock()
	appendTaskEvent(task, domaintask.PhaseNone, "DB validation passed")
	if task.Readiness.SSHAvailable {
		appendTaskEvent(task, domaintask.PhaseNone, task.Readiness.SSHMessage)
	} else if task.Readiness.SSHMessage != "" {
		appendTaskEvent(task, domaintask.PhaseNone, task.Readiness.SSHMessage)
	}
	if hint := preparePrivilegeHint(task); hint != "" {
		appendTaskEvent(task, domaintask.PhaseNone, hint)
	}
	b.mu.Unlock()

	if task.Readiness.SSHAvailable {
		conn, err := b.connUC.GetConnectionByID(context.Background(), task.ConnectionSnapshot.ID)
		if err == nil {
			execCtx.connection = conn
			if sshConfig := sshConfigFromConnection(conn); sshConfig != nil && sshConfig.Enabled {
				execCtx.sshCollector = collector.NewSSHMetricsCollector(sshConfig, time.Second)
				if err := execCtx.sshCollector.Start(); err != nil {
					task.Readiness.SSHAvailable = false
					task.Readiness.SSHMessage = fmt.Sprintf("SSH unavailable: %v", err)
				}
			}
		}
	}
	if execCtx.connection == nil {
		if conn, err := b.connUC.GetConnectionByID(context.Background(), task.ConnectionSnapshot.ID); err == nil {
			execCtx.connection = conn
		}
	}

	var err error
	switch task.Action {
	case domaintask.ActionPrepare:
		err = b.runPhase(taskID, domaintask.PhasePrepare)
	case domaintask.ActionRun:
		err = b.runPhase(taskID, domaintask.PhaseRun)
	case domaintask.ActionCleanup:
		err = b.runPhase(taskID, domaintask.PhaseCleanup)
	default:
		err = b.runPhase(taskID, domaintask.PhasePrepare)
		if err == nil {
			err = b.runPhase(taskID, domaintask.PhaseRun)
		}
		if err == nil {
			err = b.runPhase(taskID, domaintask.PhaseCleanup)
		}
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if execCtx.sshCollector != nil {
		execCtx.sshCollector.Stop()
	}
	delete(b.executions, taskID)
	if err != nil {
		if execCtx.stopRequested {
			task.Status = domaintask.StatusStopped
			task.ErrorMessage = ""
		} else {
			task.Status = domaintask.StatusFailed
			task.ErrorMessage = err.Error()
		}
	} else {
		task.Status = domaintask.StatusSuccess
	}
	task.CurrentPhase = domaintask.PhaseNone
	done := time.Now()
	task.CompletedAt = &done
	syncTaskTiming(task, done)
	b.activeTaskID = ""
}

func (b *TaskBinding) runPhase(taskID string, phase domaintask.Phase) error {
	b.mu.Lock()
	task := b.tasks[taskID]
	execCtx := b.executions[taskID]
	if task == nil || execCtx == nil {
		b.mu.Unlock()
		return fmt.Errorf("task not found")
	}
	switch phase {
	case domaintask.PhasePrepare:
		task.Status = domaintask.StatusPreparing
	case domaintask.PhaseRun:
		task.Status = domaintask.StatusRunning
	case domaintask.PhaseCleanup:
		task.Status = domaintask.StatusCleaning
	}
	task.CurrentPhase = phase
	appendTaskEvent(task, phase, fmt.Sprintf("Phase started: %s", phase))
	syncTaskTiming(task, time.Now())
	run, err := b.startPhaseRun(task, phase)
	if err != nil {
		b.mu.Unlock()
		return err
	}
	execCtx.currentRunID = run.ID
	task.RunLogPaths[string(phase)] = b.runRepo.GetLogPath(run.ID)
	task.SystemLogPaths[string(phase)] = b.runRepo.GetLogPath(run.ID)
	started := time.Now()
	task.PhaseHistory = append(task.PhaseHistory, domaintask.PhaseRecord{
		Phase:     phase,
		Status:    "started",
		RunID:     run.ID,
		StartedAt: started,
	})
	b.mu.Unlock()

	return b.waitForRun(taskID, phase, run.ID)
}

func (b *TaskBinding) startPhaseRun(task *domaintask.ExecutionTask, phase domaintask.Phase) (*execution.Run, error) {
	options, params := buildPhaseExecutionConfig(task.ResolvedParams, phase)
	runTask := &execution.BenchmarkTask{
		ID:           uuid.NewString(),
		Name:         task.Name,
		ConnectionID: task.ConnectionSnapshot.ID,
		TemplateID:   task.TemplateSnapshot.ID,
		Parameters:   params,
		Options:      options,
		CreatedAt:    time.Now(),
	}
	return b.benchmarkUC.StartBenchmark(context.Background(), runTask)
}

func (b *TaskBinding) waitForRun(taskID string, phase domaintask.Phase, runID string) error {
	for {
		time.Sleep(time.Second)
		run, err := b.benchmarkUC.GetBenchmarkStatus(context.Background(), runID)
		if err != nil {
			return err
		}
		b.refreshMetricsAndLogs(taskID, phase, runID)
		state := run.State
		if phase == domaintask.PhasePrepare && state == execution.StatePrepared {
			b.finishPhase(taskID, phase, runID, "prepared")
			return nil
		}
		if state == execution.StateCompleted {
			b.finishPhase(taskID, phase, runID, "success")
			return nil
		}
		if state == execution.StateCancelled || state == execution.StateForceStopped {
			b.finishPhase(taskID, phase, runID, "stopped")
			return fmt.Errorf("task stopped")
		}
		if state == execution.StateFailed || state == execution.StateTimeout {
			msg := run.ErrorMessage
			if msg == "" {
				msg = run.Message
			}
			b.finishPhase(taskID, phase, runID, "failed")
			b.mu.RLock()
			task := cloneTask(b.tasks[taskID])
			b.mu.RUnlock()
			return classifyTaskExecutionError(task, phase, fmt.Errorf("%s", msg))
		}
	}
}

func (b *TaskBinding) refreshMetricsAndLogs(taskID string, phase domaintask.Phase, runID string) {
	samples, _ := b.benchmarkUC.GetMetricSamples(context.Background(), runID)
	logs := b.runRepo.GetLogs(runID)

	b.mu.Lock()
	defer b.mu.Unlock()
	task := b.tasks[taskID]
	execCtx := b.executions[taskID]
	if task == nil || execCtx == nil {
		return
	}
	updateMetrics(task, samples)
	syncTaskTiming(task, time.Now())
	if execCtx.sshCollector != nil {
		updateSystemMetrics(task, execCtx.sshCollector.Snapshot())
		task.Metrics.SystemEnabled = true
		task.Metrics.SystemMessage = ""
	} else {
		task.Metrics.SystemEnabled = false
		task.Metrics.SystemMessage = task.Readiness.SSHMessage
	}

	start := execCtx.logSeen[runID]
	if start < len(logs) {
		for _, line := range logs[start:] {
			task.LogTail = append(task.LogTail, domaintask.LogLine{
				Timestamp: line.Timestamp,
				Phase:     phase,
				Stream:    line.Stream,
				Content:   line.Content,
			})
		}
		if len(task.LogTail) > taskLogTailLimit {
			task.LogTail = append([]domaintask.LogLine(nil), task.LogTail[len(task.LogTail)-taskLogTailLimit:]...)
		}
		execCtx.logSeen[runID] = len(logs)
	}
}

func (b *TaskBinding) finishPhase(taskID string, phase domaintask.Phase, runID string, status string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	task := b.tasks[taskID]
	execCtx := b.executions[taskID]
	if task == nil || execCtx == nil {
		return
	}
	for i := len(task.PhaseHistory) - 1; i >= 0; i-- {
		if task.PhaseHistory[i].RunID == runID {
			now := time.Now()
			task.PhaseHistory[i].Status = status
			task.PhaseHistory[i].EndedAt = &now
			break
		}
	}
	appendTaskEvent(task, phase, fmt.Sprintf("Phase finished: %s (%s)", phase, status))
	syncTaskTiming(task, time.Now())
	execCtx.currentRunID = ""
}

func (b *TaskBinding) completeTask(taskID string, status domaintask.Status, phase domaintask.Phase, errMsg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	task := b.tasks[taskID]
	if task == nil {
		return
	}
	task.Status = status
	task.CurrentPhase = phase
	task.ErrorMessage = errMsg
	if errMsg != "" {
		appendTaskEvent(task, phase, errMsg)
	}
	done := time.Now()
	task.CompletedAt = &done
	syncTaskTiming(task, done)
	b.activeTaskID = ""
	delete(b.executions, taskID)
}

func actionSupported(action domaintask.Action, phases map[string]bool) bool {
	switch action {
	case domaintask.ActionPrepare:
		return phases["prepare"]
	case domaintask.ActionRun:
		return phases["run"]
	case domaintask.ActionCleanup:
		return phases["cleanup"]
	default:
		return phases["prepare"] && phases["run"] && phases["cleanup"]
	}
}

func validateReadiness(readiness domaintask.Readiness) error {
	if !readiness.TemplateSelected {
		return fmt.Errorf("template is required")
	}
	if !readiness.ConnectionSelected {
		return fmt.Errorf("connection is required")
	}
	if !readiness.ActionSupported {
		return fmt.Errorf("selected action is not supported by template")
	}
	if !readiness.RuntimeValid {
		return fmt.Errorf("runtime overrides are invalid")
	}
	if !readiness.DBValid {
		if readiness.DBMessage != "" {
			return fmt.Errorf("db validation failed: %s", readiness.DBMessage)
		}
		return fmt.Errorf("db validation failed")
	}
	return nil
}

func buildPhaseExecutionConfig(baseParams map[string]interface{}, phase domaintask.Phase) (execution.TaskOptions, map[string]interface{}) {
	options := execution.TaskOptions{RunTimeout: 24 * time.Hour}
	params := cloneParams(baseParams)
	switch phase {
	case domaintask.PhasePrepare:
		options.SkipPrepare = false
		options.SkipCleanup = true
		params["_prepare_only"] = true
		params["_original_time"] = params["time"]
		params["time"] = 0
	case domaintask.PhaseRun:
		options.SkipPrepare = true
		options.SkipCleanup = true
	case domaintask.PhaseCleanup:
		options.SkipPrepare = true
		options.SkipCleanup = false
		params["_cleanup_only"] = true
		params["time"] = 0
		delete(params, "_original_time")
	}
	return options, params
}

func templateSnapshot(tmpl *domaintemplate.Template, params map[string]interface{}) domaintask.TemplateSnapshot {
	phases := map[string]bool{
		"prepare": tmpl.Phases.Prepare.Enabled || tmpl.CommandTemplate.Prepare != "",
		"run":     tmpl.Phases.Run.Enabled || tmpl.CommandTemplate.Run != "",
		"cleanup": tmpl.Phases.Cleanup.Enabled || tmpl.CommandTemplate.Cleanup != "",
	}
	return domaintask.TemplateSnapshot{
		ID:             tmpl.ID,
		Name:           tmpl.Name,
		Tool:           tmpl.Tool,
		DBFamily:       tmpl.DBFamily,
		WorkloadFamily: tmpl.WorkloadFamily,
		Phases:         phases,
		Parameters:     params,
	}
}

func connectionSnapshot(conn connection.Connection) domaintask.ConnectionSnapshot {
	snapshot := domaintask.ConnectionSnapshot{
		ID:   conn.GetID(),
		Name: conn.GetName(),
		Type: string(conn.GetType()),
	}
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		snapshot.Host = c.Host
		snapshot.Port = c.Port
		snapshot.Database = c.Database
		snapshot.Username = c.Username
	case *connection.PostgreSQLConnection:
		snapshot.Host = c.Host
		snapshot.Port = c.Port
		snapshot.Database = c.Database
		snapshot.Username = c.Username
	case *connection.OracleConnection:
		snapshot.Host = c.Host
		snapshot.Port = c.Port
		snapshot.Database = c.ServiceName
		snapshot.Username = c.Username
	case *connection.SQLServerConnection:
		snapshot.Host = c.Host
		snapshot.Port = c.Port
		snapshot.Database = c.Database
		snapshot.Username = c.Username
	}
	if ssh := sshConfigFromConnection(conn); ssh != nil {
		snapshot.SSHEnabled = ssh.Enabled
		snapshot.SSHHost = ssh.Host
		snapshot.SSHPort = ssh.Port
		snapshot.SSHUsername = ssh.Username
	}
	return snapshot
}

func sshConfigFromConnection(conn connection.Connection) *connection.SSHTunnelConfig {
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

func resolveParams(tmpl *domaintemplate.Template, overrides map[string]interface{}) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	for key, def := range tmpl.Parameters {
		params[key] = def.Default
	}
	if tmpl.Runtime.DurationSeconds > 0 {
		params["time"] = tmpl.Runtime.DurationSeconds
	}
	switch tmpl.Tool {
	case domaintemplate.ToolSysbench:
		if tmpl.ToolConfig.Sysbench.Tables > 0 {
			params["tables"] = tmpl.ToolConfig.Sysbench.Tables
		}
		if tmpl.ToolConfig.Sysbench.TableSize > 0 {
			params["table_size"] = tmpl.ToolConfig.Sysbench.TableSize
		}
		if tmpl.Runtime.Concurrency.Value > 0 {
			params["threads"] = tmpl.Runtime.Concurrency.Value
		}
	case domaintemplate.ToolSwingbench:
		if tmpl.ToolConfig.Swingbench.UserCount > 0 {
			params["virtual_users"] = tmpl.ToolConfig.Swingbench.UserCount
		}
		if tmpl.ToolConfig.Swingbench.RunTimeSeconds > 0 {
			params["time"] = tmpl.ToolConfig.Swingbench.RunTimeSeconds
		}
		if tmpl.ToolConfig.Swingbench.XMLOverrides != "" {
			params["config_file"] = tmpl.ToolConfig.Swingbench.XMLOverrides
		}
	case domaintemplate.ToolHammerDB:
		if tmpl.ToolConfig.HammerDB.VirtualUsers > 0 {
			params["virtual_users"] = tmpl.ToolConfig.HammerDB.VirtualUsers
		}
		if tmpl.ToolConfig.HammerDB.Warehouses > 0 {
			params["warehouses"] = tmpl.ToolConfig.HammerDB.Warehouses
		}
		if tmpl.ToolConfig.HammerDB.ScaleFactor > 0 {
			params["scale"] = tmpl.ToolConfig.HammerDB.ScaleFactor
		}
		if tmpl.Runtime.DurationSeconds > 0 {
			params["duration"] = tmpl.Runtime.DurationSeconds
		}
		params["rampup"] = tmpl.Runtime.RampUpSeconds
		if tmpl.Runtime.Iterations > 0 {
			params["iterations"] = tmpl.Runtime.Iterations
		}
	}
	for key, value := range overrides {
		params[key] = value
	}
	for _, key := range []string{"threads", "virtual_users", "time", "duration", "tables", "table_size", "warehouses", "scale"} {
		if value, ok := params[key]; ok {
			intValue, ok := toInt(value)
			if !ok {
				return nil, fmt.Errorf("%s must be an integer", key)
			}
			if intValue <= 0 && key != "duration" {
				return nil, fmt.Errorf("%s must be positive", key)
			}
			params[key] = intValue
		}
	}
	if tmpl.Tool != domaintemplate.ToolHammerDB {
		if duration, ok := params["duration"]; ok {
			params["time"] = duration
			delete(params, "duration")
		}
	}
	return params, nil
}

func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func updateMetrics(task *domaintask.ExecutionTask, samples []execution.MetricSample) {
	tpsSeries := make([]domaintask.MetricSeriesPoint, 0, len(samples))
	tpmSeries := make([]domaintask.MetricSeriesPoint, 0, len(samples))
	var tpsSum float64
	var tpmSum float64
	for _, sample := range samples {
		tps := sample.TPS
		tpm := sample.TPM
		if tpm == 0 && tps > 0 {
			tpm = tps * 60
		}
		ts := sample.Timestamp.UnixMilli()
		tpsSeries = append(tpsSeries, domaintask.MetricSeriesPoint{Timestamp: ts, Value: tps})
		tpmSeries = append(tpmSeries, domaintask.MetricSeriesPoint{Timestamp: ts, Value: tpm})
		tpsSum += tps
		tpmSum += tpm
		if tps > task.Metrics.TPS.Max {
			task.Metrics.TPS.Max = tps
		}
		if tpm > task.Metrics.TPM.Max {
			task.Metrics.TPM.Max = tpm
		}
		task.Metrics.TPS.Current = tps
		task.Metrics.TPM.Current = tpm
	}
	if len(samples) > 0 {
		task.Metrics.TPS.Avg = tpsSum / float64(len(samples))
		task.Metrics.TPM.Avg = tpmSum / float64(len(samples))
	}
	if len(tpsSeries) > 300 {
		tpsSeries = tpsSeries[len(tpsSeries)-300:]
	}
	if len(tpmSeries) > 300 {
		tpmSeries = tpmSeries[len(tpmSeries)-300:]
	}
	task.Metrics.TPS.Series = tpsSeries
	task.Metrics.TPM.Series = tpmSeries
}

func updateSystemMetrics(task *domaintask.ExecutionTask, points []collector.SSHMetricPoint) {
	if len(points) == 0 {
		return
	}
	task.Metrics.CPUUser.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.CPUUser })
	task.Metrics.CPUSys.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.CPUSys })
	task.Metrics.CPUIOWait.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.CPUIOWait })
	task.Metrics.CPUSteal.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.CPUSteal })
	task.Metrics.DiskReadBps.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.DiskReadBps })
	task.Metrics.DiskWriteBps.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.DiskWriteBps })
	task.Metrics.DiskReadLatencyMs.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.DiskReadLatencyMs })
	task.Metrics.DiskWriteLatencyMs.Series = shrinkSystemSeries(points, func(point collector.SSHMetricPoint) float64 { return point.DiskWriteLatencyMs })
	last := points[len(points)-1]
	task.Metrics.CPUUser.Current = last.CPUUser
	task.Metrics.CPUSys.Current = last.CPUSys
	task.Metrics.CPUIOWait.Current = last.CPUIOWait
	task.Metrics.CPUSteal.Current = last.CPUSteal
	task.Metrics.DiskReadBps.Current = last.DiskReadBps
	task.Metrics.DiskWriteBps.Current = last.DiskWriteBps
	task.Metrics.DiskReadLatencyMs.Current = last.DiskReadLatencyMs
	task.Metrics.DiskWriteLatencyMs.Current = last.DiskWriteLatencyMs
}

func shrinkSystemSeries(points []collector.SSHMetricPoint, fn func(point collector.SSHMetricPoint) float64) []domaintask.MetricSeriesPoint {
	start := 0
	if len(points) > 300 {
		start = len(points) - 300
	}
	series := make([]domaintask.MetricSeriesPoint, 0, len(points)-start)
	for _, point := range points[start:] {
		series = append(series, domaintask.MetricSeriesPoint{Timestamp: point.Timestamp, Value: fn(point)})
	}
	return series
}

func filterLogLines(lines []domaintask.LogLine, query string, phase string) []domaintask.LogLine {
	filtered := make([]domaintask.LogLine, 0, len(lines))
	query = strings.ToLower(strings.TrimSpace(query))
	for _, line := range lines {
		if phase != "" && phase != string(line.Phase) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(line.Content), query) {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}

func syncTaskTiming(task *domaintask.ExecutionTask, now time.Time) {
	if task == nil {
		return
	}

	var prepareMs int64
	var runMs int64
	var cleanupMs int64
	for _, record := range task.PhaseHistory {
		end := now
		if record.EndedAt != nil {
			end = *record.EndedAt
		}
		if end.Before(record.StartedAt) {
			continue
		}
		durationMs := end.Sub(record.StartedAt).Milliseconds()
		switch record.Phase {
		case domaintask.PhasePrepare:
			prepareMs += durationMs
		case domaintask.PhaseRun:
			runMs += durationMs
		case domaintask.PhaseCleanup:
			cleanupMs += durationMs
		}
	}

	totalMs := int64(0)
	if task.StartedAt != nil {
		end := now
		if task.CompletedAt != nil {
			end = *task.CompletedAt
		}
		if end.After(*task.StartedAt) || end.Equal(*task.StartedAt) {
			totalMs = end.Sub(*task.StartedAt).Milliseconds()
		}
	}

	task.Timing = domaintask.TaskTiming{
		PrepareMs:          prepareMs,
		RunMs:              runMs,
		CleanupMs:          cleanupMs,
		TotalMs:            totalMs,
		RunDurationInputMs: resolveRequestedRunDuration(task.ResolvedParams),
	}
}

func resolveRequestedRunDuration(params map[string]interface{}) int64 {
	if len(params) == 0 {
		return 0
	}
	for _, key := range []string{"time", "duration"} {
		value, ok := params[key]
		if !ok {
			continue
		}
		seconds, ok := toInt(value)
		if !ok || seconds <= 0 {
			continue
		}
		return int64(seconds) * 1000
	}
	return 0
}

func classifyTaskExecutionError(task *domaintask.ExecutionTask, phase domaintask.Phase, err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if task == nil {
		return err
	}
	if strings.EqualFold(task.ConnectionSnapshot.Type, "oracle") &&
		strings.EqualFold(task.BenchmarkTool, string(domaintemplate.ToolSwingbench)) &&
		phase == domaintask.PhasePrepare &&
		strings.Contains(strings.ToUpper(message), "ORA-01031") &&
		strings.Contains(strings.ToLower(message), "dbms_lock") {
		username := strings.TrimSpace(task.ConnectionSnapshot.Username)
		if username == "" {
			username = "current connection account"
		}
		return fmt.Errorf("Oracle Swingbench prepare requires a higher-privilege account. The prepare flow ran post-schema setup as %s and failed to grant EXECUTE on sys.dbms_lock to SOE (ORA-01031). Use a DBA/SYSDBA-style account for prepare, or grant the required privilege before prepare. Run can use a lower-privilege SOE workload account after schema build. Original error: %s", username, message)
	}
	return err
}

func preparePrivilegeHint(task *domaintask.ExecutionTask) string {
	if task == nil {
		return ""
	}
	if !strings.EqualFold(task.ConnectionSnapshot.Type, "oracle") || !strings.EqualFold(task.BenchmarkTool, string(domaintemplate.ToolSwingbench)) {
		return ""
	}
	if task.Action != domaintask.ActionPrepare && task.Action != domaintask.ActionFullPipeline {
		return ""
	}
	return "Oracle Swingbench prepare uses the configured connection account for schema setup. Prepare requires DBA/SYSDBA-style privileges for schema build and DBMS_LOCK grant; run can use the lower-privilege SOE workload account after prepare succeeds."
}

func cloneParams(params map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}

func cloneTask(task *domaintask.ExecutionTask) *domaintask.ExecutionTask {
	if task == nil {
		return nil
	}
	cloned := *task
	cloned.ResolvedParams = cloneParams(task.ResolvedParams)
	cloned.LogTail = append([]domaintask.LogLine(nil), task.LogTail...)
	cloned.PhaseHistory = append([]domaintask.PhaseRecord(nil), task.PhaseHistory...)
	cloned.RunLogPaths = cloneParamsString(task.RunLogPaths)
	cloned.SystemLogPaths = cloneParamsString(task.SystemLogPaths)
	cloned.Metrics = cloneMetrics(task.Metrics)
	return &cloned
}

func appendTaskEvent(task *domaintask.ExecutionTask, phase domaintask.Phase, content string) {
	if task == nil || strings.TrimSpace(content) == "" {
		return
	}
	task.LogTail = append(task.LogTail, domaintask.LogLine{
		Timestamp: time.Now().Format(time.RFC3339),
		Phase:     phase,
		Stream:    "event",
		Content:   content,
	})
	if len(task.LogTail) > taskLogTailLimit {
		task.LogTail = append([]domaintask.LogLine(nil), task.LogTail[len(task.LogTail)-taskLogTailLimit:]...)
	}
}

func cloneParamsString(params map[string]string) map[string]string {
	cloned := make(map[string]string, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}

func cloneMetrics(metrics domaintask.UnifiedMetrics) domaintask.UnifiedMetrics {
	metrics.TPS.Series = cloneSeries(metrics.TPS.Series)
	metrics.TPM.Series = cloneSeries(metrics.TPM.Series)
	metrics.CPUUser.Series = cloneSeries(metrics.CPUUser.Series)
	metrics.CPUSys.Series = cloneSeries(metrics.CPUSys.Series)
	metrics.CPUIOWait.Series = cloneSeries(metrics.CPUIOWait.Series)
	metrics.CPUSteal.Series = cloneSeries(metrics.CPUSteal.Series)
	metrics.DiskReadBps.Series = cloneSeries(metrics.DiskReadBps.Series)
	metrics.DiskWriteBps.Series = cloneSeries(metrics.DiskWriteBps.Series)
	metrics.DiskReadLatencyMs.Series = cloneSeries(metrics.DiskReadLatencyMs.Series)
	metrics.DiskWriteLatencyMs.Series = cloneSeries(metrics.DiskWriteLatencyMs.Series)
	return metrics
}

func cloneSeries(series []domaintask.MetricSeriesPoint) []domaintask.MetricSeriesPoint {
	return append([]domaintask.MetricSeriesPoint(nil), series...)
}
