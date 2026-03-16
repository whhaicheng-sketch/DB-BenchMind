package bindings

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintask "github.com/whhaicheng/DB-BenchMind/internal/domain/task"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/collector"
)

func TestBuildPhaseExecutionConfig(t *testing.T) {
	base := map[string]interface{}{
		"time":    60,
		"threads": 8,
	}

	t.Run("prepare phase freezes runtime into prepare-only config", func(t *testing.T) {
		options, params := buildPhaseExecutionConfig(base, domaintask.PhasePrepare)
		if options.SkipPrepare {
			t.Fatalf("prepare phase should not skip prepare")
		}
		if !options.SkipCleanup {
			t.Fatalf("prepare phase should skip cleanup")
		}
		if got := params["time"]; got != 0 {
			t.Fatalf("prepare time = %v, want 0", got)
		}
		if got := params["_original_time"]; got != 60 {
			t.Fatalf("prepare _original_time = %v, want 60", got)
		}
	})

	t.Run("cleanup phase does not execute run workload", func(t *testing.T) {
		options, params := buildPhaseExecutionConfig(base, domaintask.PhaseCleanup)
		if !options.SkipPrepare {
			t.Fatalf("cleanup phase should skip prepare")
		}
		if options.SkipCleanup {
			t.Fatalf("cleanup phase should execute cleanup")
		}
		if got := params["time"]; got != 0 {
			t.Fatalf("cleanup time = %v, want 0", got)
		}
		if _, ok := params["_original_time"]; ok {
			t.Fatalf("cleanup phase should not carry _original_time")
		}
		if got := params["_cleanup_only"]; got != true {
			t.Fatalf("cleanup _cleanup_only = %v, want true", got)
		}
	})
}

func TestValidateReadiness(t *testing.T) {
	err := validateReadiness(domaintask.Readiness{
		TemplateSelected:   true,
		ConnectionSelected: true,
		ActionSupported:    true,
		RuntimeValid:       true,
		DBValid:            true,
	})
	if err != nil {
		t.Fatalf("validateReadiness() returned unexpected error: %v", err)
	}

	err = validateReadiness(domaintask.Readiness{
		TemplateSelected:   true,
		ConnectionSelected: true,
		ActionSupported:    false,
		RuntimeValid:       true,
		DBValid:            true,
	})
	if err == nil {
		t.Fatalf("validateReadiness() expected action support error")
	}
}

func TestTemplateSnapshot_SwingbenchPhasesDriveTaskActions(t *testing.T) {
	tmpl := &domaintemplate.Template{
		ID:             "tpl_swing",
		Name:           "Swingbench Oracle",
		Tool:           domaintemplate.ToolSwingbench,
		DBFamily:       "oracle",
		WorkloadFamily: "order-entry",
		Phases: domaintemplate.PhaseSet{
			Build:    domaintemplate.PhaseConfig{Enabled: true},
			Generate: domaintemplate.PhaseConfig{Enabled: true},
			Run:      domaintemplate.PhaseConfig{Enabled: true, Required: true},
		},
		Runtime: domaintemplate.Runtime{
			Concurrency:     domaintemplate.Concurrency{Mode: "users", Value: 8},
			DurationSeconds: 60,
		},
		ToolConfig: domaintemplate.ToolConfig{
			Swingbench: domaintemplate.SwingbenchConfig{
				Benchmark:      "orderEntry",
				UserCount:      8,
				RunTimeSeconds: 60,
			},
		},
	}
	tmpl.Normalize()

	snapshot := templateSnapshot(tmpl, map[string]interface{}{"time": 60})

	if !snapshot.Phases["prepare"] {
		t.Fatal("templateSnapshot() should expose prepare action for legacy swingbench phases")
	}
	if !snapshot.Phases["run"] {
		t.Fatal("templateSnapshot() should expose run action for swingbench")
	}
	if !snapshot.Phases["cleanup"] {
		t.Fatal("templateSnapshot() should expose cleanup action for legacy swingbench phases")
	}
	if !actionSupported(domaintask.ActionFullPipeline, snapshot.Phases) {
		t.Fatal("full pipeline should be supported when prepare/run/cleanup are enabled")
	}
}

func TestUpdateSystemMetricsIncludesCPUSteal(t *testing.T) {
	task := &domaintask.ExecutionTask{}
	points := []collector.SSHMetricPoint{
		{Timestamp: 1000, CPUUser: 10, CPUSys: 4, CPUIOWait: 1, CPUSteal: 0.5},
		{Timestamp: 2000, CPUUser: 12, CPUSys: 5, CPUIOWait: 2, CPUSteal: 1.25},
	}

	updateSystemMetrics(task, points)

	if got := task.Metrics.CPUSteal.Current; got != 1.25 {
		t.Fatalf("CPUSteal.Current = %v, want 1.25", got)
	}
	if len(task.Metrics.CPUSteal.Series) != len(points) {
		t.Fatalf("CPUSteal.Series length = %d, want %d", len(task.Metrics.CPUSteal.Series), len(points))
	}
	if got := task.Metrics.CPUSteal.Series[0].Value; got != 0.5 {
		t.Fatalf("CPUSteal.Series[0].Value = %v, want 0.5", got)
	}
}

func TestResolveParams_MapsToolSpecificDefaultsForExecution(t *testing.T) {
	t.Run("sysbench uses runtime and tool config defaults", func(t *testing.T) {
		tmpl := &domaintemplate.Template{
			ID:             "tpl-test-sysbench",
			Name:           "MySQL - Sysbench Test",
			Tool:           domaintemplate.ToolSysbench,
			DBFamily:       "mysql",
			WorkloadFamily: "oltp-read-write",
			Scope:          domaintemplate.ScopeTest,
			Status:         domaintemplate.StatusReady,
			Tags:           []string{"test"},
			Phases: domaintemplate.PhaseSet{
				Prepare: domaintemplate.PhaseConfig{Enabled: true},
				Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true},
				Cleanup: domaintemplate.PhaseConfig{Enabled: true},
			},
			Runtime: domaintemplate.Runtime{
				Concurrency:     domaintemplate.Concurrency{Mode: "threads", Value: 1},
				DurationSeconds: 30,
				RampUpSeconds:   3,
			},
			ToolConfig: domaintemplate.ToolConfig{
				Sysbench: domaintemplate.SysbenchConfig{
					DBDriver:   "mysql",
					ScriptType: "oltp_read_write",
					Tables:     1,
					TableSize:  1000,
				},
			},
		}
		tmpl.Normalize()

		params, err := resolveParams(tmpl, nil)
		if err != nil {
			t.Fatalf("resolveParams() failed: %v", err)
		}
		if got := params["tables"]; got != 1 {
			t.Fatalf("tables = %v, want 1", got)
		}
		if got := params["table_size"]; got != 1000 {
			t.Fatalf("table_size = %v, want 1000", got)
		}
		if got := params["threads"]; got != 1 {
			t.Fatalf("threads = %v, want 1", got)
		}
		if got := params["time"]; got != 30 {
			t.Fatalf("time = %v, want 30", got)
		}
	})

	t.Run("swingbench keeps scale and minimal run defaults", func(t *testing.T) {
		tmpl := &domaintemplate.Template{
			ID:             "tpl-test-swingbench",
			Name:           "Oracle - Swingbench Test",
			Tool:           domaintemplate.ToolSwingbench,
			DBFamily:       "oracle",
			WorkloadFamily: "order-entry",
			Scope:          domaintemplate.ScopeTest,
			Status:         domaintemplate.StatusReady,
			Tags:           []string{"test"},
			Parameters: map[string]domaintemplate.Parameter{
				"scale": {
					Type:    domaintemplate.ParameterTypeInteger,
					Label:   "Scale",
					Default: 0.1,
				},
			},
			Phases: domaintemplate.PhaseSet{
				Prepare: domaintemplate.PhaseConfig{Enabled: true},
				Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true},
				Cleanup: domaintemplate.PhaseConfig{Enabled: true},
			},
			Runtime: domaintemplate.Runtime{
				Concurrency:     domaintemplate.Concurrency{Mode: "users", Value: 1},
				DurationSeconds: 60,
			},
			ToolConfig: domaintemplate.ToolConfig{
				Swingbench: domaintemplate.SwingbenchConfig{
					Benchmark:      "orderEntry",
					UserCount:      1,
					RunTimeSeconds: 60,
				},
			},
		}
		tmpl.Normalize()

		params, err := resolveParams(tmpl, nil)
		if err != nil {
			t.Fatalf("resolveParams() failed: %v", err)
		}
		if got := params["scale"]; got != 0.1 {
			t.Fatalf("scale = %v, want 0.1", got)
		}
		if got := params["virtual_users"]; got != 1 {
			t.Fatalf("virtual_users = %v, want 1", got)
		}
		if got := params["time"]; got != 60 {
			t.Fatalf("time = %v, want 60", got)
		}
	})

	t.Run("hammerdb maps runtime to duration and rampup keys", func(t *testing.T) {
		tmpl := &domaintemplate.Template{
			ID:             "tpl-test-hammerdb",
			Name:           "SQL Server - HammerDB Test",
			Tool:           domaintemplate.ToolHammerDB,
			DBFamily:       "sqlserver",
			WorkloadFamily: "tproc-c",
			Scope:          domaintemplate.ScopeTest,
			Status:         domaintemplate.StatusReady,
			Tags:           []string{"test"},
			Phases: domaintemplate.PhaseSet{
				Prepare: domaintemplate.PhaseConfig{Enabled: true},
				Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true},
				Cleanup: domaintemplate.PhaseConfig{Enabled: true},
			},
			Runtime: domaintemplate.Runtime{
				Concurrency:     domaintemplate.Concurrency{Mode: "virtualUsers", Value: 1},
				DurationSeconds: 60,
				RampUpSeconds:   0,
				Iterations:      1,
			},
			ToolConfig: domaintemplate.ToolConfig{
				HammerDB: domaintemplate.HammerDBConfig{
					Benchmark:    "tproc-c",
					VirtualUsers: 1,
					Warehouses:   1,
					ScaleFactor:  1,
				},
			},
		}
		tmpl.Normalize()

		params, err := resolveParams(tmpl, nil)
		if err != nil {
			t.Fatalf("resolveParams() failed: %v", err)
		}
		if got := params["virtual_users"]; got != 1 {
			t.Fatalf("virtual_users = %v, want 1", got)
		}
		if got := params["warehouses"]; got != 1 {
			t.Fatalf("warehouses = %v, want 1", got)
		}
		if got := params["duration"]; got != 60 {
			t.Fatalf("duration = %v, want 60", got)
		}
		if got := params["rampup"]; got != 0 {
			t.Fatalf("rampup = %v, want 0", got)
		}
		if got := params["iterations"]; got != 1 {
			t.Fatalf("iterations = %v, want 1", got)
		}
	})
}

func TestSyncTaskTimingFromPhaseHistory(t *testing.T) {
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	prepareEnd := base.Add(33 * time.Second)
	runStart := prepareEnd
	runEnd := runStart.Add(60 * time.Second)
	cleanupEnd := runEnd.Add(5 * time.Second)
	completed := cleanupEnd

	task := &domaintask.ExecutionTask{
		CreatedAt:   base.Add(-2 * time.Second),
		StartedAt:   &base,
		CompletedAt: &completed,
		ResolvedParams: map[string]interface{}{
			"time": 60,
		},
		PhaseHistory: []domaintask.PhaseRecord{
			{Phase: domaintask.PhasePrepare, Status: "prepared", StartedAt: base, EndedAt: &prepareEnd},
			{Phase: domaintask.PhaseRun, Status: "success", StartedAt: runStart, EndedAt: &runEnd},
			{Phase: domaintask.PhaseCleanup, Status: "success", StartedAt: runEnd, EndedAt: &cleanupEnd},
		},
	}

	syncTaskTiming(task, completed)

	if got := task.Timing.PrepareMs; got != 33_000 {
		t.Fatalf("PrepareMs = %d, want 33000", got)
	}
	if got := task.Timing.RunMs; got != 60_000 {
		t.Fatalf("RunMs = %d, want 60000", got)
	}
	if got := task.Timing.CleanupMs; got != 5_000 {
		t.Fatalf("CleanupMs = %d, want 5000", got)
	}
	if got := task.Timing.TotalMs; got != 98_000 {
		t.Fatalf("TotalMs = %d, want 98000", got)
	}
	if got := task.Timing.RunDurationInputMs; got != 60_000 {
		t.Fatalf("RunDurationInputMs = %d, want 60000", got)
	}
}

func TestSyncTaskTimingUsesRequestedRunDurationFromResolvedParams(t *testing.T) {
	now := time.Date(2026, 3, 15, 13, 0, 0, 0, time.UTC)
	task := &domaintask.ExecutionTask{
		ResolvedParams: map[string]interface{}{
			"time": 120,
		},
	}

	syncTaskTiming(task, now)

	if got := task.Timing.RunDurationInputMs; got != 120_000 {
		t.Fatalf("RunDurationInputMs = %d, want 120000", got)
	}
}

func TestSyncTaskTiming_DoesNotCountRunMsWhileOracleSwingbenchIsStillPreflighting(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 30, 0, time.UTC)
	runStarted := now.Add(-30 * time.Second)
	task := &domaintask.ExecutionTask{
		Status:        domaintask.StatusStarting,
		CurrentPhase:  domaintask.PhaseNone,
		BenchmarkTool: "swingbench",
		ConnectionSnapshot: domaintask.ConnectionSnapshot{
			Type: "oracle",
		},
		PhaseHistory: []domaintask.PhaseRecord{
			{Phase: domaintask.PhaseRun, Status: "started", StartedAt: runStarted},
		},
	}

	syncTaskTiming(task, now)

	if got := task.Timing.RunMs; got != 0 {
		t.Fatalf("RunMs = %d, want 0 while task is still preflighting", got)
	}
}

func TestClassifyTaskExecutionError_OraclePreparePermission(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ConnectionSnapshot: domaintask.ConnectionSnapshot{Type: "oracle", Username: "app_user"},
		BenchmarkTool:      "swingbench",
		CurrentPhase:       domaintask.PhasePrepare,
	}

	err := classifyTaskExecutionError(task, domaintask.PhasePrepare, errors.New("GRANT EXECUTE ON dbms_lock TO soe : ORA-01031: insufficient privileges"))
	if err == nil {
		t.Fatal("classifyTaskExecutionError() returned nil")
	}
	if got := err.Error(); got == "GRANT EXECUTE ON dbms_lock TO soe : ORA-01031: insufficient privileges" {
		t.Fatalf("error was not enriched: %s", got)
	}
	if got := err.Error(); !containsAll(got, []string{"Oracle Swingbench prepare requires a higher-privilege account", "dbms_lock", "prepare", "Run can use a lower-privilege SOE workload account"}) {
		t.Fatalf("unexpected enriched error: %s", got)
	}
}

func TestClassifyTaskExecutionError_OracleRunAuthenticationFailure(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ConnectionSnapshot: domaintask.ConnectionSnapshot{Type: "oracle", Username: "soe"},
		BenchmarkTool:      "swingbench",
		CurrentPhase:       domaintask.PhaseRun,
	}

	err := classifyTaskExecutionError(task, domaintask.PhaseRun, errors.New("Could not establish/maintain connection: ORA-01017: invalid username/password; logon denied"))
	if err == nil {
		t.Fatal("classifyTaskExecutionError() returned nil")
	}
	if got := err.Error(); !containsAll(got, []string{"Oracle Swingbench run failed", "Invalid SOE workload username/password"}) {
		t.Fatalf("unexpected authentication error: %s", got)
	}
}

func TestClassifyTaskExecutionError_OracleRunMissingWorkloadUser(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ConnectionSnapshot: domaintask.ConnectionSnapshot{Type: "oracle", Username: "soe"},
		BenchmarkTool:      "swingbench",
		CurrentPhase:       domaintask.PhaseRun,
	}

	err := classifyTaskExecutionError(task, domaintask.PhaseRun, errors.New("ORA-01918: user 'SOE' does not exist"))
	if err == nil {
		t.Fatal("classifyTaskExecutionError() returned nil")
	}
	if got := err.Error(); !containsAll(got, []string{"Oracle Swingbench run failed", "SOE workload user does not exist"}) {
		t.Fatalf("unexpected missing-user error: %s", got)
	}
}

func TestClassifyTaskExecutionError_OracleRunLockedWorkloadUser(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ConnectionSnapshot: domaintask.ConnectionSnapshot{Type: "oracle", Username: "soe"},
		BenchmarkTool:      "swingbench",
		CurrentPhase:       domaintask.PhaseRun,
	}

	err := classifyTaskExecutionError(task, domaintask.PhaseRun, errors.New("ORA-28000: the account is locked"))
	if err == nil {
		t.Fatal("classifyTaskExecutionError() returned nil")
	}
	if got := err.Error(); !containsAll(got, []string{"Oracle Swingbench run failed", "SOE workload user is locked"}) {
		t.Fatalf("unexpected locked-user error: %s", got)
	}
}

func TestClassifyTaskExecutionError_OracleRunMissingPreparedSchema(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ConnectionSnapshot: domaintask.ConnectionSnapshot{Type: "oracle", Username: "soe"},
		BenchmarkTool:      "swingbench",
		CurrentPhase:       domaintask.PhaseRun,
	}

	err := classifyTaskExecutionError(task, domaintask.PhaseRun, errors.New("ORA-00942: table or view does not exist"))
	if err == nil {
		t.Fatal("classifyTaskExecutionError() returned nil")
	}
	if got := err.Error(); !containsAll(got, []string{"Oracle Swingbench run failed", "Run requires prepared SOE schema. Please run Prepare first."}) {
		t.Fatalf("unexpected schema error: %s", got)
	}
}

func TestOracleSwingbenchRuntimeFailure_DetectsCleanupInvalidatedZeroThroughput(t *testing.T) {
	task := &domaintask.ExecutionTask{
		Status:        domaintask.StatusRunning,
		CurrentPhase:  domaintask.PhaseRun,
		BenchmarkTool: "swingbench",
		ConnectionSnapshot: domaintask.ConnectionSnapshot{
			Type: "oracle",
		},
		PhaseHistory: []domaintask.PhaseRecord{
			{Phase: domaintask.PhaseCleanup, Status: "success", StartedAt: time.Now().Add(-2 * time.Minute)},
		},
		LogTail: []domaintask.LogLine{
			{Phase: domaintask.PhaseRun, Content: "10:58:35 [0/10]      0        0       0"},
			{Phase: domaintask.PhaseRun, Content: "10:58:36 [0/10]      0        0       0"},
			{Phase: domaintask.PhaseRun, Content: "10:58:37 [0/10]      0        0       0"},
		},
		Metrics: domaintask.UnifiedMetrics{
			TPS: domaintask.MetricSummary{Current: 0},
			TPM: domaintask.MetricSummary{Current: 0},
		},
	}
	run := &execution.Run{
		ID:      "run-zero",
		State:   execution.StateRunning,
		Message: "",
		Result: &execution.BenchmarkResult{
			TotalTransactions: 0,
		},
	}

	err := detectOracleSwingbenchRuntimeFailure(task, run)
	if err == nil {
		t.Fatal("detectOracleSwingbenchRuntimeFailure() expected error")
	}
	if got := err.Error(); !containsAll(got, []string{"Cleanup removed required SOE objects", "Prepare first"}) {
		t.Fatalf("unexpected runtime failure: %s", got)
	}
}

func TestOracleSwingbenchDirectRunGuard_BlocksRunAfterCleanup(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ID:            "task-run",
		Action:        domaintask.ActionRun,
		BenchmarkTool: "swingbench",
		ConnectionSnapshot: domaintask.ConnectionSnapshot{
			ID:   "conn-oracle",
			Type: "oracle",
		},
		TemplateSnapshot: domaintask.TemplateSnapshot{
			WorkloadFamily: "order-entry",
		},
	}
	tasks := map[string]*domaintask.ExecutionTask{
		"task-run": task,
		"task-cleanup": {
			ID:            "task-cleanup",
			BenchmarkTool: "swingbench",
			ConnectionSnapshot: domaintask.ConnectionSnapshot{
				ID:   "conn-oracle",
				Type: "oracle",
			},
			TemplateSnapshot: domaintask.TemplateSnapshot{
				WorkloadFamily: "order-entry",
			},
			PhaseHistory: []domaintask.PhaseRecord{
				{Phase: domaintask.PhaseCleanup, Status: "success", StartedAt: time.Now().Add(-time.Minute)},
			},
		},
	}

	err := oracleSwingbenchDirectRunGuard(task, tasks)
	if err == nil {
		t.Fatal("oracleSwingbenchDirectRunGuard() expected error")
	}
	if got := err.Error(); !containsAll(got, []string{"Cleanup removed required SOE objects", "Prepare first"}) {
		t.Fatalf("unexpected direct run guard error: %s", got)
	}
}

func TestOracleSwingbenchDirectRunGuard_AllowsRunAfterLaterPrepare(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ID:            "task-run",
		Action:        domaintask.ActionRun,
		BenchmarkTool: "swingbench",
		ConnectionSnapshot: domaintask.ConnectionSnapshot{
			ID:   "conn-oracle",
			Type: "oracle",
		},
		TemplateSnapshot: domaintask.TemplateSnapshot{
			WorkloadFamily: "order-entry",
		},
	}
	now := time.Now()
	tasks := map[string]*domaintask.ExecutionTask{
		"task-run": task,
		"task-cleanup": {
			ID:            "task-cleanup",
			BenchmarkTool: "swingbench",
			ConnectionSnapshot: domaintask.ConnectionSnapshot{
				ID:   "conn-oracle",
				Type: "oracle",
			},
			TemplateSnapshot: domaintask.TemplateSnapshot{
				WorkloadFamily: "order-entry",
			},
			PhaseHistory: []domaintask.PhaseRecord{
				{Phase: domaintask.PhaseCleanup, Status: "success", StartedAt: now.Add(-2 * time.Minute)},
			},
		},
		"task-prepare": {
			ID:            "task-prepare",
			BenchmarkTool: "swingbench",
			ConnectionSnapshot: domaintask.ConnectionSnapshot{
				ID:   "conn-oracle",
				Type: "oracle",
			},
			TemplateSnapshot: domaintask.TemplateSnapshot{
				WorkloadFamily: "order-entry",
			},
			PhaseHistory: []domaintask.PhaseRecord{
				{Phase: domaintask.PhasePrepare, Status: "prepared", StartedAt: now.Add(-time.Minute)},
			},
		},
	}

	if err := oracleSwingbenchDirectRunGuard(task, tasks); err != nil {
		t.Fatalf("oracleSwingbenchDirectRunGuard() unexpected error: %v", err)
	}
}

func TestTaskBindingRunPhase_OracleSwingbenchCleanupGuardDoesNotOptimisticallyEnterRunning(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ID:            "task-run",
		Action:        domaintask.ActionRun,
		Status:        domaintask.StatusStarting,
		CurrentPhase:  domaintask.PhaseNone,
		BenchmarkTool: "swingbench",
		ConnectionSnapshot: domaintask.ConnectionSnapshot{
			ID:   "conn-oracle",
			Type: "oracle",
		},
		TemplateSnapshot: domaintask.TemplateSnapshot{
			WorkloadFamily: "order-entry",
		},
	}
	binding := &TaskBinding{
		tasks: map[string]*domaintask.ExecutionTask{
			"task-run": task,
			"task-cleanup": {
				ID:            "task-cleanup",
				BenchmarkTool: "swingbench",
				ConnectionSnapshot: domaintask.ConnectionSnapshot{
					ID:   "conn-oracle",
					Type: "oracle",
				},
				TemplateSnapshot: domaintask.TemplateSnapshot{
					WorkloadFamily: "order-entry",
				},
				PhaseHistory: []domaintask.PhaseRecord{
					{Phase: domaintask.PhaseCleanup, Status: "success", StartedAt: time.Now().Add(-time.Minute)},
				},
			},
		},
		executions: map[string]*taskExecutionContext{
			"task-run": {
				logSeen: make(map[string]int),
			},
		},
	}

	err := binding.runPhase("task-run", domaintask.PhaseRun)
	if err == nil {
		t.Fatal("runPhase() expected cleanup guard error")
	}
	if got := task.Status; got != domaintask.StatusStarting {
		t.Fatalf("task status = %s, want %s", got, domaintask.StatusStarting)
	}
	if got := task.CurrentPhase; got != domaintask.PhaseNone {
		t.Fatalf("task current phase = %s, want %s", got, domaintask.PhaseNone)
	}
	if len(task.PhaseHistory) != 0 {
		t.Fatalf("phase history length = %d, want 0", len(task.PhaseHistory))
	}
	if !containsAll(task.LogTail[len(task.LogTail)-1].Content, []string{"Cleanup removed required SOE objects", "Prepare first"}) {
		t.Fatalf("unexpected log tail: %+v", task.LogTail)
	}
}

func TestTaskBindingRunPhase_MySQLSysbenchCleanupGuardDoesNotOptimisticallyEnterRunning(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ID:            "task-run",
		Action:        domaintask.ActionRun,
		Status:        domaintask.StatusStarting,
		CurrentPhase:  domaintask.PhaseNone,
		BenchmarkTool: "sysbench",
		ConnectionSnapshot: domaintask.ConnectionSnapshot{
			ID:   "conn-mysql",
			Type: "mysql",
		},
		TemplateSnapshot: domaintask.TemplateSnapshot{
			WorkloadFamily: "oltp-read-write",
		},
	}
	binding := &TaskBinding{
		tasks: map[string]*domaintask.ExecutionTask{
			"task-run": task,
			"task-cleanup": {
				ID:            "task-cleanup",
				BenchmarkTool: "sysbench",
				ConnectionSnapshot: domaintask.ConnectionSnapshot{
					ID:   "conn-mysql",
					Type: "mysql",
				},
				TemplateSnapshot: domaintask.TemplateSnapshot{
					WorkloadFamily: "oltp-read-write",
				},
				PhaseHistory: []domaintask.PhaseRecord{
					{Phase: domaintask.PhaseCleanup, Status: "success", StartedAt: time.Now().Add(-time.Minute)},
				},
			},
		},
		executions: map[string]*taskExecutionContext{
			"task-run": {
				logSeen: make(map[string]int),
			},
		},
	}

	err := binding.runPhase("task-run", domaintask.PhaseRun)
	if err == nil {
		t.Fatal("runPhase() expected cleanup guard error")
	}
	if got := task.Status; got != domaintask.StatusStarting {
		t.Fatalf("task status = %s, want %s", got, domaintask.StatusStarting)
	}
	if got := task.CurrentPhase; got != domaintask.PhaseNone {
		t.Fatalf("task current phase = %s, want %s", got, domaintask.PhaseNone)
	}
	if len(task.PhaseHistory) != 0 {
		t.Fatalf("phase history length = %d, want 0", len(task.PhaseHistory))
	}
	if !containsAll(task.LogTail[len(task.LogTail)-1].Content, []string{"Cleanup removed required benchmark objects", "Prepare first"}) {
		t.Fatalf("unexpected log tail: %+v", task.LogTail)
	}
}

func TestClassifyTaskExecutionError_SysbenchMissingTables(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ConnectionSnapshot: domaintask.ConnectionSnapshot{Type: "mysql"},
		BenchmarkTool:      "sysbench",
		CurrentPhase:       domaintask.PhaseRun,
	}

	err := classifyTaskExecutionError(task, domaintask.PhaseRun, errors.New(`run: ✗ Error: Benchmark tables do not exist

The benchmark tables do not exist (Table 'sbtest.sbtest1' doesn't exist).

Please run the Prepare phase first to create the tables and load data.`))
	if err == nil {
		t.Fatal("classifyTaskExecutionError() returned nil")
	}
	if got := err.Error(); !containsAll(got, []string{"Sysbench run failed", "benchmark tables are not prepared", "Prepare first"}) {
		t.Fatalf("unexpected sysbench error: %s", got)
	}
}

func TestClassifyTaskExecutionError_SQLServerHammerDBMissingObjects(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ConnectionSnapshot: domaintask.ConnectionSnapshot{Type: "sqlserver"},
		BenchmarkTool:      "hammerdb",
		CurrentPhase:       domaintask.PhaseRun,
	}

	err := classifyTaskExecutionError(task, domaintask.PhaseRun, errors.New("Error: Invalid object name 'dbo.warehouse'"))
	if err == nil {
		t.Fatal("classifyTaskExecutionError() returned nil")
	}
	if got := err.Error(); !containsAll(got, []string{"HammerDB run failed", "benchmark objects are missing", "Prepare first"}) {
		t.Fatalf("unexpected HammerDB object error: %s", got)
	}
}

func TestFinishPhase_FailedRunPersistsTerminalReasonToTaskState(t *testing.T) {
	task := &domaintask.ExecutionTask{
		ID: "task-run",
		PhaseHistory: []domaintask.PhaseRecord{
			{
				Phase:     domaintask.PhaseRun,
				Status:    "started",
				RunID:     "run-1",
				StartedAt: time.Now().Add(-5 * time.Second),
			},
		},
		LogTail: []domaintask.LogLine{
			{Phase: domaintask.PhaseRun, Stream: "event", Content: "Phase started: run"},
		},
	}
	binding := &TaskBinding{
		tasks: map[string]*domaintask.ExecutionTask{
			task.ID: task,
		},
		executions: map[string]*taskExecutionContext{
			task.ID: {
				currentRunID: "run-1",
				logSeen:      make(map[string]int),
			},
		},
	}

	message := "Sysbench run failed: benchmark tables are not prepared. Please run Prepare first."
	binding.finishPhase(task.ID, domaintask.PhaseRun, "run-1", "failed", message)

	if got := task.PhaseHistory[0].Status; got != "failed" {
		t.Fatalf("phase status = %s, want failed", got)
	}
	if got := task.PhaseHistory[0].Message; got != message {
		t.Fatalf("phase message = %q, want %q", got, message)
	}
	if len(task.LogTail) < 2 {
		t.Fatalf("log tail length = %d, want >= 2", len(task.LogTail))
	}
	if got := task.LogTail[len(task.LogTail)-1].Content; got != message {
		t.Fatalf("last log line = %q, want %q", got, message)
	}
}

func TestTaskBindingStopTask_IgnoresDuplicateStopRequests(t *testing.T) {
	binding := &TaskBinding{
		tasks: map[string]*domaintask.ExecutionTask{
			"task-1": {
				ID:           "task-1",
				Status:       domaintask.StatusRunning,
				CurrentPhase: domaintask.PhaseRun,
				LogTail: []domaintask.LogLine{
					{Stream: "event", Content: "Phase started: run"},
				},
			},
		},
		executions: map[string]*taskExecutionContext{
			"task-1": {
				logSeen: make(map[string]int),
			},
		},
	}

	first := binding.StopTask("task-1")
	second := binding.StopTask("task-1")
	if !first.Success {
		t.Fatalf("first StopTask() success = false, error = %s", first.Error)
	}
	if !second.Success {
		t.Fatalf("second StopTask() success = false, error = %s", second.Error)
	}

	task := binding.tasks["task-1"]
	stopEvents := 0
	for _, line := range task.LogTail {
		if line.Stream == "event" && line.Content == "Stop requested" {
			stopEvents++
		}
	}
	if stopEvents != 1 {
		t.Fatalf("Stop requested events = %d, want 1", stopEvents)
	}
}

func TestTaskBindingStopTask_RollsBackStoppingStateWhenStopRequestFails(t *testing.T) {
	runRepo := useTestRunRepoWithRun(&execution.Run{
		ID:    "run-completed",
		State: execution.StateCompleted,
	})
	benchmarkUC := &usecase.BenchmarkUseCase{}
	*benchmarkUC = *usecase.NewBenchmarkUseCase(runRepo, nil, nil, nil)

	binding := &TaskBinding{
		benchmarkUC: benchmarkUC,
		tasks: map[string]*domaintask.ExecutionTask{
			"task-1": {
				ID:           "task-1",
				Status:       domaintask.StatusRunning,
				CurrentPhase: domaintask.PhaseRun,
			},
		},
		executions: map[string]*taskExecutionContext{
			"task-1": {
				currentRunID: "run-completed",
				logSeen:      make(map[string]int),
			},
		},
	}

	result := binding.StopTask("task-1")
	if result.Error == "" {
		t.Fatal("StopTask() expected error")
	}

	task := binding.tasks["task-1"]
	if got := task.Status; got != domaintask.StatusRunning {
		t.Fatalf("task status = %s, want %s", got, domaintask.StatusRunning)
	}
	if binding.executions["task-1"].stopRequested {
		t.Fatal("stopRequested should rollback to false on stop failure")
	}
}

func TestTaskBindingStopTask_ReconcilesPreparedOracleSwingbenchRunWithoutProcess(t *testing.T) {
	runRepo := useTestRunRepoWithRun(&execution.Run{
		ID:           "run-prepared",
		State:        execution.StatePrepared,
		ErrorMessage: "Oracle Swingbench run failed: SOE workload user does not exist. Run Prepare first or recreate Swingbench schema.",
	})
	benchmarkUC := &usecase.BenchmarkUseCase{}
	*benchmarkUC = *usecase.NewBenchmarkUseCase(runRepo, nil, nil, nil)

	binding := &TaskBinding{
		benchmarkUC:  benchmarkUC,
		activeTaskID: "task-1",
		tasks: map[string]*domaintask.ExecutionTask{
			"task-1": {
				ID:            "task-1",
				Status:        domaintask.StatusStarting,
				CurrentPhase:  domaintask.PhaseNone,
				BenchmarkTool: "swingbench",
				ConnectionSnapshot: domaintask.ConnectionSnapshot{
					Type: "oracle",
				},
			},
		},
		executions: map[string]*taskExecutionContext{
			"task-1": {
				currentRunID: "run-prepared",
				logSeen:      make(map[string]int),
			},
		},
	}

	result := binding.StopTask("task-1")
	if result.Error != "" {
		t.Fatalf("StopTask() error = %q, want quick reconciliation", result.Error)
	}

	task := binding.tasks["task-1"]
	if task.Status != domaintask.StatusFailed {
		t.Fatalf("task status = %s, want %s", task.Status, domaintask.StatusFailed)
	}
	if task.CompletedAt == nil {
		t.Fatal("CompletedAt should be set")
	}
	if binding.activeTaskID != "" {
		t.Fatalf("activeTaskID = %q, want empty", binding.activeTaskID)
	}
	if _, ok := binding.executions["task-1"]; ok {
		t.Fatal("stale execution context should be cleared")
	}
	if !containsAll(task.ErrorMessage, []string{"No active benchmark process to stop", "never entered run"}) {
		t.Fatalf("unexpected error message: %s", task.ErrorMessage)
	}
}

func useTestRunRepoWithRun(run *execution.Run) *usecase.MemoryRunRepository {
	repo := usecase.NewMemoryRunRepository()
	if run != nil {
		_ = repo.Save(context.Background(), run)
	}
	return repo
}

func containsAll(value string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(value, sub) {
			return false
		}
	}
	return true
}
