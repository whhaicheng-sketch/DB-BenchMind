// Package usecase provides unit tests for benchmark use case.
package usecase

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
)

func TestOracleSwingbenchRunPreflight_FailsForInvalidWorkloadCredentials(t *testing.T) {
	restoreUserExists := oracleSwingbenchPreflightUserExists
	restorePing := oracleSwingbenchPreflightPing
	restoreSchema := oracleSwingbenchPreflightSchemaCheck
	defer func() {
		oracleSwingbenchPreflightUserExists = restoreUserExists
		oracleSwingbenchPreflightPing = restorePing
		oracleSwingbenchPreflightSchemaCheck = restoreSchema
	}()

	oracleSwingbenchPreflightUserExists = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		return true, nil
	}
	oracleSwingbenchPreflightPing = func(ctx context.Context, conn *connection.OracleConnection) error {
		if conn.Username == "soe" {
			return errors.New("ORA-01017: invalid username/password; logon denied")
		}
		return nil
	}
	oracleSwingbenchPreflightSchemaCheck = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		return true, nil
	}

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "system",
		Password:       "manager",
	}
	config := &adapter.Config{
		Connection: conn,
		Template:   &domaintemplate.Template{Tool: domaintemplate.ToolSwingbench},
		Parameters: map[string]interface{}{},
	}

	err := oracleSwingbenchRunPreflight(context.Background(), config)
	if err == nil {
		t.Fatal("oracleSwingbenchRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"Invalid SOE workload username/password"}) {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestOracleSwingbenchRunPreflight_FailsWhenWorkloadUserDoesNotExist(t *testing.T) {
	restoreUserExists := oracleSwingbenchPreflightUserExists
	restorePing := oracleSwingbenchPreflightPing
	restoreSchema := oracleSwingbenchPreflightSchemaCheck
	defer func() {
		oracleSwingbenchPreflightUserExists = restoreUserExists
		oracleSwingbenchPreflightPing = restorePing
		oracleSwingbenchPreflightSchemaCheck = restoreSchema
	}()

	loginCalled := false
	schemaCalled := false
	oracleSwingbenchPreflightUserExists = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		return false, nil
	}
	oracleSwingbenchPreflightPing = func(ctx context.Context, conn *connection.OracleConnection) error {
		loginCalled = true
		return nil
	}
	oracleSwingbenchPreflightSchemaCheck = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		schemaCalled = true
		return true, nil
	}

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "system",
		Password:       "manager",
	}
	config := &adapter.Config{
		Connection: conn,
		Template:   &domaintemplate.Template{Tool: domaintemplate.ToolSwingbench},
		Parameters: map[string]interface{}{},
	}

	err := oracleSwingbenchRunPreflight(context.Background(), config)
	if err == nil {
		t.Fatal("oracleSwingbenchRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"SOE workload user does not exist", "Run Prepare first or recreate Swingbench schema"}) {
		t.Fatalf("unexpected error: %s", got)
	}
	if loginCalled {
		t.Fatal("login check should not run when workload user is missing")
	}
	if schemaCalled {
		t.Fatal("schema check should not run when workload user is missing")
	}
}

func TestOracleSwingbenchRunPreflight_FailsWhenWorkloadUserIsLocked(t *testing.T) {
	restoreUserExists := oracleSwingbenchPreflightUserExists
	restorePing := oracleSwingbenchPreflightPing
	restoreSchema := oracleSwingbenchPreflightSchemaCheck
	defer func() {
		oracleSwingbenchPreflightUserExists = restoreUserExists
		oracleSwingbenchPreflightPing = restorePing
		oracleSwingbenchPreflightSchemaCheck = restoreSchema
	}()

	schemaCalled := false
	oracleSwingbenchPreflightUserExists = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		return true, nil
	}
	oracleSwingbenchPreflightPing = func(ctx context.Context, conn *connection.OracleConnection) error {
		return errors.New("ORA-28000: the account is locked")
	}
	oracleSwingbenchPreflightSchemaCheck = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		schemaCalled = true
		return true, nil
	}

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "system",
		Password:       "manager",
	}
	config := &adapter.Config{
		Connection: conn,
		Template:   &domaintemplate.Template{Tool: domaintemplate.ToolSwingbench},
		Parameters: map[string]interface{}{},
	}

	err := oracleSwingbenchRunPreflight(context.Background(), config)
	if err == nil {
		t.Fatal("oracleSwingbenchRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"SOE workload user is locked"}) {
		t.Fatalf("unexpected error: %s", got)
	}
	if schemaCalled {
		t.Fatal("schema check should not run when workload user is locked")
	}
}

func TestOracleSwingbenchRunPreflight_InvalidCredentialsDoNotFallThroughToSchemaCheck(t *testing.T) {
	restoreUserExists := oracleSwingbenchPreflightUserExists
	restorePing := oracleSwingbenchPreflightPing
	restoreSchema := oracleSwingbenchPreflightSchemaCheck
	defer func() {
		oracleSwingbenchPreflightUserExists = restoreUserExists
		oracleSwingbenchPreflightPing = restorePing
		oracleSwingbenchPreflightSchemaCheck = restoreSchema
	}()

	oracleSwingbenchPreflightUserExists = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		return true, nil
	}
	schemaCalled := false
	oracleSwingbenchPreflightPing = func(ctx context.Context, conn *connection.OracleConnection) error {
		return errors.New("ORA-01017: invalid username/password; logon denied")
	}
	oracleSwingbenchPreflightSchemaCheck = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		schemaCalled = true
		return false, nil
	}

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "system",
		Password:       "manager",
	}
	config := &adapter.Config{
		Connection: conn,
		Template:   &domaintemplate.Template{Tool: domaintemplate.ToolSwingbench},
		Parameters: map[string]interface{}{},
	}

	err := oracleSwingbenchRunPreflight(context.Background(), config)
	if err == nil {
		t.Fatal("oracleSwingbenchRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"Invalid SOE workload username/password"}) {
		t.Fatalf("unexpected error: %s", got)
	}
	if schemaCalled {
		t.Fatal("schema check should not run when workload credentials are invalid")
	}
}

func TestOracleSwingbenchRunPreflight_FailsWhenLoginSucceedsButObjectsAreMissing(t *testing.T) {
	restoreUserExists := oracleSwingbenchPreflightUserExists
	restorePing := oracleSwingbenchPreflightPing
	restoreSchema := oracleSwingbenchPreflightSchemaCheck
	defer func() {
		oracleSwingbenchPreflightUserExists = restoreUserExists
		oracleSwingbenchPreflightPing = restorePing
		oracleSwingbenchPreflightSchemaCheck = restoreSchema
	}()

	oracleSwingbenchPreflightUserExists = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		return true, nil
	}
	oracleSwingbenchPreflightPing = func(ctx context.Context, conn *connection.OracleConnection) error {
		return nil
	}
	oracleSwingbenchPreflightSchemaCheck = func(ctx context.Context, conn *connection.OracleConnection, workloadUser string) (bool, error) {
		return false, nil
	}

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "system",
		Password:       "manager",
	}
	config := &adapter.Config{
		Connection: conn,
		Template:   &domaintemplate.Template{Tool: domaintemplate.ToolSwingbench},
		Parameters: map[string]interface{}{},
	}

	err := oracleSwingbenchRunPreflight(context.Background(), config)
	if err == nil {
		t.Fatal("oracleSwingbenchRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"Run requires prepared SOE schema. Please run Prepare first."}) {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestBenchmarkRunPreflight_FailsForMySQLSysbenchWhenDatabaseIsMissing(t *testing.T) {
	restore := sysbenchMySQLRunPreflightCheck
	defer func() {
		sysbenchMySQLRunPreflightCheck = restore
	}()

	sysbenchMySQLRunPreflightCheck = func(ctx context.Context, conn *connection.MySQLConnection, dbName string) (benchmarkRunPreflightStatus, error) {
		return benchmarkRunPreflightDatabaseMissing, nil
	}

	err := benchmarkRunPreflight(context.Background(), &adapter.Config{
		Connection: &connection.MySQLConnection{
			BaseConnection: connection.BaseConnection{ID: "mysql-1", Name: "MySQL Sysbench"},
			Host:           "127.0.0.1",
			Port:           3306,
			Database:       "sbtest",
			Username:       "root",
			Password:       "secret",
		},
		Template: &domaintemplate.Template{
			Tool:     domaintemplate.ToolSysbench,
			DBFamily: "mysql",
		},
		Parameters: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("benchmarkRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"Sysbench run failed", "benchmark database does not exist", "Prepare first"}) {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestBenchmarkRunPreflight_FailsForPostgreSQLSysbenchWhenTablesAreMissing(t *testing.T) {
	restore := sysbenchPostgreSQLRunPreflightCheck
	defer func() {
		sysbenchPostgreSQLRunPreflightCheck = restore
	}()

	sysbenchPostgreSQLRunPreflightCheck = func(ctx context.Context, conn *connection.PostgreSQLConnection, dbName string) (benchmarkRunPreflightStatus, error) {
		return benchmarkRunPreflightSchemaMissing, nil
	}

	err := benchmarkRunPreflight(context.Background(), &adapter.Config{
		Connection: &connection.PostgreSQLConnection{
			BaseConnection: connection.BaseConnection{ID: "pg-1", Name: "PostgreSQL Sysbench"},
			Host:           "127.0.0.1",
			Port:           5432,
			Database:       "sbtest",
			Username:       "postgres",
			Password:       "secret",
			SSLMode:        "disable",
		},
		Template: &domaintemplate.Template{
			Tool:     domaintemplate.ToolSysbench,
			DBFamily: "postgresql",
		},
		Parameters: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("benchmarkRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"Sysbench run failed", "benchmark tables are not prepared", "Prepare first"}) {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestBenchmarkRunPreflight_FailsForSQLServerHammerDBWhenObjectsAreMissing(t *testing.T) {
	restore := hammerDBSQLServerRunPreflightCheck
	defer func() {
		hammerDBSQLServerRunPreflightCheck = restore
	}()

	hammerDBSQLServerRunPreflightCheck = func(ctx context.Context, conn *connection.SQLServerConnection, databaseName string) (benchmarkRunPreflightStatus, error) {
		return benchmarkRunPreflightBenchmarkObjectsMissing, nil
	}

	err := benchmarkRunPreflight(context.Background(), &adapter.Config{
		Connection: &connection.SQLServerConnection{
			BaseConnection: connection.BaseConnection{ID: "sqlserver-1", Name: "SQL Server HammerDB"},
			Host:           "127.0.0.1",
			Port:           1433,
			Database:       "tpcc",
			Username:       "sa",
			Password:       "secret",
		},
		Template: &domaintemplate.Template{
			Tool:     domaintemplate.ToolHammerDB,
			DBFamily: "sqlserver",
		},
		Parameters: map[string]interface{}{
			"database_name": "tpcc",
		},
	})
	if err == nil {
		t.Fatal("benchmarkRunPreflight() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"HammerDB run failed", "benchmark objects are missing", "Prepare first"}) {
		t.Fatalf("unexpected error: %s", got)
	}
}

func containsBenchmarkUseCaseSubs(value string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(value, sub) {
			return false
		}
	}
	return true
}

// mockRunRepository is a mock implementation of RunRepository for testing.
type mockRunRepository struct {
	runs map[string]*execution.Run
}

var (
	// ErrRunNotFound is returned when a run is not found.
	ErrRunNotFound = errors.New("run not found")
)

func newMockRunRepository() *mockRunRepository {
	return &mockRunRepository{
		runs: make(map[string]*execution.Run),
	}
}

func (m *mockRunRepository) Save(ctx context.Context, run *execution.Run) error {
	m.runs[run.ID] = run
	return nil
}

func (m *mockRunRepository) FindByID(ctx context.Context, id string) (*execution.Run, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, ErrRunNotFound
	}
	return run, nil
}

func (m *mockRunRepository) FindAll(ctx context.Context, opts FindOptions) ([]*execution.Run, error) {
	var result []*execution.Run
	for _, run := range m.runs {
		result = append(result, run)
	}
	return result, nil
}

func (m *mockRunRepository) UpdateState(ctx context.Context, id string, state execution.RunState) error {
	run, ok := m.runs[id]
	if !ok {
		return ErrRunNotFound
	}
	if err := run.SetState(state); err != nil {
		return err
	}
	return nil
}

func (m *mockRunRepository) SaveMetricSample(ctx context.Context, runID string, sample execution.MetricSample) error {
	return nil // Ignore for mock
}

func (m *mockRunRepository) GetMetricSamples(ctx context.Context, runID string) ([]execution.MetricSample, error) {
	return []execution.MetricSample{}, nil // Return empty slice for mock
}

func (m *mockRunRepository) SaveLogEntry(ctx context.Context, runID string, entry LogEntry) error {
	return nil // Ignore for mock
}

func (m *mockRunRepository) Delete(ctx context.Context, id string) error {
	delete(m.runs, id)
	return nil
}

// TestBenchmarkUseCase_StartBenchmark tests starting a benchmark.
func TestBenchmarkUseCase_StartBenchmark(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	runRepo := newMockRunRepository()
	adapterReg := adapter.NewAdapterRegistry()
	// Register sysbench adapter for testing
	adapterReg.Register(adapter.NewSysbenchAdapter())

	// Create mock connection repository with a test connection
	connRepo := newMockConnectionRepository()
	testConn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-1",
			Name: "Test Connection",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
	}
	connRepo.Save(ctx, testConn)

	// Create mock template repository with a test template
	templateRepo := newMockTemplateRepositoryForBenchmark()
	testTmpl := &domaintemplate.Template{
		ID:            "sysbench-oltp-read-write",
		Name:          "Sysbench OLTP",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: domaintemplate.CommandTemplate{
			Run: "run",
		},
		OutputParser: domaintemplate.OutputParser{
			Type: domaintemplate.ParserTypeRegex,
		},
	}
	templateRepo.Save(ctx, testTmpl)

	// Create use cases
	connUseCase := NewConnectionUseCase(connRepo, nil)
	templateUseCase := NewTemplateUseCase(templateRepo, "")

	uc := NewBenchmarkUseCase(runRepo, adapterReg, connUseCase, templateUseCase)

	// Create a test task
	task := &execution.BenchmarkTask{
		ID:           "test-task-1",
		Name:         "Test Benchmark",
		ConnectionID: "test-conn-1",
		TemplateID:   "sysbench-oltp-read-write",
		Parameters: map[string]interface{}{
			"threads": 8,
			"time":    60,
		},
		CreatedAt: time.Now(),
	}

	// Start benchmark (will fail pre-checks since we don't have a real connection)
	// This test mainly verifies the structure is correct
	run, err := uc.StartBenchmark(ctx, task)

	// We expect this to fail during pre-checks in the goroutine
	// but the run object should be created and returned immediately
	if err != nil {
		t.Fatalf("StartBenchmark() failed immediately: %v", err)
	}

	if run.ID == "" {
		t.Error("Run ID should not be empty")
	}
	if run.State != execution.StatePending {
		t.Errorf("Initial state should be pending, got %s", run.State)
	}
}

// TestBenchmarkUseCase_StopBenchmark tests stopping a benchmark.
func TestBenchmarkUseCase_StopBenchmark(t *testing.T) {
	ctx := context.Background()

	runRepo := newMockRunRepository()
	adapterReg := adapter.NewAdapterRegistry()
	templateRepo := newMockTemplateRepositoryForBenchmark()
	templateUseCase := NewTemplateUseCase(templateRepo, "")
	connRepo := newMockConnectionRepository()
	connUseCase := NewConnectionUseCase(connRepo, nil)

	uc := NewBenchmarkUseCase(runRepo, adapterReg, connUseCase, templateUseCase)

	// Create a running run
	run := &execution.Run{
		ID:        "test-run-1",
		TaskID:    "test-task-1",
		State:     execution.StateRunning,
		CreatedAt: time.Now(),
	}
	runRepo.Save(ctx, run)

	// Stop the benchmark
	err := uc.StopBenchmark(ctx, run.ID, false)
	if err != nil {
		t.Fatalf("StopBenchmark() failed: %v", err)
	}

	// Verify state was updated
	stopped, _ := runRepo.FindByID(ctx, run.ID)
	if stopped.State != execution.StateCancelled {
		t.Errorf("State should be cancelled, got %s", stopped.State)
	}
	if stopped.CompletedAt == nil {
		t.Error("CompletedAt should be set after StopBenchmark")
	}
}

// TestBenchmarkUseCase_StopBenchmark_InvalidState tests stopping a non-running benchmark.
func TestBenchmarkUseCase_StopBenchmark_InvalidState(t *testing.T) {
	ctx := context.Background()

	runRepo := newMockRunRepository()
	adapterReg := adapter.NewAdapterRegistry()
	templateRepo := newMockTemplateRepositoryForBenchmark()
	templateUseCase := NewTemplateUseCase(templateRepo, "")
	connRepo := newMockConnectionRepository()
	connUseCase := NewConnectionUseCase(connRepo, nil)

	uc := NewBenchmarkUseCase(runRepo, adapterReg, connUseCase, templateUseCase)

	// Create a completed run
	run := &execution.Run{
		ID:        "test-run-1",
		TaskID:    "test-task-1",
		State:     execution.StateCompleted,
		CreatedAt: time.Now(),
	}
	runRepo.Save(ctx, run)

	// Try to stop - should fail
	err := uc.StopBenchmark(ctx, run.ID, false)
	if err == nil {
		t.Error("StopBenchmark() should return error for completed run")
	}
}

func TestTerminateProcess_KillsProcessGroupChildren(t *testing.T) {
	tempDir := t.TempDir()
	childPIDPath := filepath.Join(tempDir, "child.pid")

	cmd := exec.Command("bash", "-lc", "sleep 30 & echo $! > "+childPIDPath+"; wait")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer cmd.Wait()

	var childPID int
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(childPIDPath)
		if err == nil && strings.TrimSpace(string(data)) != "" {
			value, convErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if convErr != nil {
				t.Fatalf("Atoi(child pid) failed: %v", convErr)
			}
			childPID = value
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if childPID == 0 {
		t.Fatal("child pid was not captured before terminateProcess()")
	}

	if err := terminateProcess(cmd, true); err != nil {
		t.Fatalf("terminateProcess() failed: %v", err)
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Fatal("process group did not exit after terminateProcess()")
	}

	if err := syscall.Kill(childPID, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("child process still alive after terminateProcess(), err = %v", err)
	}
}

// TestBenchmarkUseCase_GetBenchmarkStatus tests getting benchmark status.
func TestBenchmarkUseCase_GetBenchmarkStatus(t *testing.T) {
	ctx := context.Background()

	runRepo := newMockRunRepository()
	adapterReg := adapter.NewAdapterRegistry()
	templateRepo := newMockTemplateRepositoryForBenchmark()
	templateUseCase := NewTemplateUseCase(templateRepo, "")
	connRepo := newMockConnectionRepository()
	connUseCase := NewConnectionUseCase(connRepo, nil)

	uc := NewBenchmarkUseCase(runRepo, adapterReg, connUseCase, templateUseCase)

	// Create a run
	run := &execution.Run{
		ID:        "test-run-1",
		TaskID:    "test-task-1",
		State:     execution.StateRunning,
		CreatedAt: time.Now(),
	}
	runRepo.Save(ctx, run)

	// Get status
	status, err := uc.GetBenchmarkStatus(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetBenchmarkStatus() failed: %v", err)
	}

	if status.State != execution.StateRunning {
		t.Errorf("State = %s, want %s", status.State, execution.StateRunning)
	}
}

// TestBenchmarkUseCase_ListBenchmarks tests listing benchmarks.
func TestBenchmarkUseCase_ListBenchmarks(t *testing.T) {
	ctx := context.Background()

	runRepo := newMockRunRepository()
	adapterReg := adapter.NewAdapterRegistry()
	templateRepo := newMockTemplateRepositoryForBenchmark()
	templateUseCase := NewTemplateUseCase(templateRepo, "")
	connRepo := newMockConnectionRepository()
	connUseCase := NewConnectionUseCase(connRepo, nil)

	uc := NewBenchmarkUseCase(runRepo, adapterReg, connUseCase, templateUseCase)

	// Create multiple runs
	runs := []*execution.Run{
		{
			ID:        "run-1",
			TaskID:    "task-1",
			State:     execution.StateCompleted,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "run-2",
			TaskID:    "task-2",
			State:     execution.StateRunning,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, run := range runs {
		runRepo.Save(ctx, run)
	}

	// List all
	all, err := uc.ListBenchmarks(ctx, FindOptions{})
	if err != nil {
		t.Fatalf("ListBenchmarks() failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("ListBenchmarks() count = %d, want 2", len(all))
	}
}

// TestBenchmarkExecutor_Stop tests executor stop functionality.
func TestBenchmarkExecutor_Stop(t *testing.T) {
	executor := &BenchmarkExecutor{
		runID: "test-run-1",
	}

	// Test force stop
	err := executor.Stop(true)
	if err != nil {
		t.Errorf("Stop(force=true) failed: %v", err)
	}

	if !executor.stopping {
		t.Error("Executor should be marked as stopping")
	}
}

// TestParseCommandLine tests command line parsing.
func TestParseCommandLine(t *testing.T) {
	tests := []struct {
		name    string
		cmdLine string
		wantLen int
		wantErr bool
	}{
		{
			name:    "simple command",
			cmdLine: "sysbench mysql run",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "command with flags",
			cmdLine: "sysbench mysql --threads=8 --time=60 run",
			wantLen: 5,
			wantErr: false,
		},
		{
			name:    "empty command",
			cmdLine: "",
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := parseCommandLine(tt.cmdLine)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCommandLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(parts) != tt.wantLen {
				t.Errorf("parseCommandLine() len = %d, want %d", len(parts), tt.wantLen)
			}
		})
	}
}

func TestIsSwingbenchCommandLine(t *testing.T) {
	tests := []struct {
		name    string
		cmdLine string
		want    bool
	}{
		{name: "charbench script", cmdLine: "./charbench -c ../configs/server_side_soe_v2.xml", want: true},
		{name: "oewizard launcher", cmdLine: "java -cp ../launcher LauncherBootstrap -executablename oewizard oewizard", want: true},
		{name: "minibench", cmdLine: "./minibench -c config.xml", want: true},
		{name: "sysbench", cmdLine: "sysbench oltp_read_write run", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSwingbenchCommandLine(tt.cmdLine); got != tt.want {
				t.Fatalf("isSwingbenchCommandLine(%q) = %v, want %v", tt.cmdLine, got, tt.want)
			}
		})
	}
}

// TestCheckDiskSpace tests disk space checking.
func TestCheckDiskSpace(t *testing.T) {
	uc := &BenchmarkUseCase{}

	// Test with temp directory (should have enough space)
	err := uc.checkDiskSpace("/tmp", 1024)
	if err != nil {
		// This might fail on some systems, so we'll just log it
		t.Logf("checkDiskSpace() failed (might be OK on some systems): %v", err)
	}
}

// TestMarkAsFailed tests marking a run as failed.
func TestMarkAsFailed(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := &BenchmarkUseCase{runRepo: runRepo}

	// Create a run
	run := &execution.Run{
		ID:        "test-run-1",
		TaskID:    "test-task-1",
		State:     execution.StateRunning,
		CreatedAt: time.Now(),
	}
	runRepo.Save(ctx, run)

	// Mark as failed
	uc.markAsFailed(ctx, run.ID, "test error")

	// Verify
	failed, _ := runRepo.FindByID(ctx, run.ID)
	if failed.State != execution.StateFailed {
		t.Errorf("State = %s, want %s", failed.State, execution.StateFailed)
	}
	if failed.ErrorMessage != "test error" {
		t.Errorf("ErrorMessage = %s, want 'test error'", failed.ErrorMessage)
	}
}

func TestMarkAsFailed_AllowsPreparedRunToFailAfterRunPreflight(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := &BenchmarkUseCase{runRepo: runRepo}

	now := time.Now()
	run := &execution.Run{
		ID:        "test-run-prepared",
		TaskID:    "test-task-prepared",
		State:     execution.StatePrepared,
		CreatedAt: now.Add(-time.Minute),
	}
	runRepo.Save(ctx, run)

	uc.markAsFailed(ctx, run.ID, "run: Oracle Swingbench run failed: SOE workload user does not exist")

	failed, _ := runRepo.FindByID(ctx, run.ID)
	if failed.State != execution.StateFailed {
		t.Fatalf("State = %s, want %s", failed.State, execution.StateFailed)
	}
	if failed.ErrorMessage == "" {
		t.Fatal("ErrorMessage should be populated")
	}
	if failed.CompletedAt == nil {
		t.Fatal("CompletedAt should be set for prepared -> failed reconciliation")
	}
}

// TestMarkAsCompleted tests marking a run as completed.
func TestMarkAsCompleted(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := &BenchmarkUseCase{runRepo: runRepo}

	now := time.Now()
	run := &execution.Run{
		ID:        "test-run-1",
		TaskID:    "test-task-1",
		State:     execution.StateRunning,
		CreatedAt: now,
		StartedAt: &now,
	}
	runRepo.Save(ctx, run)

	duration := 60 * time.Second
	uc.markAsCompleted(ctx, run.ID, duration)

	// Verify
	completed, _ := runRepo.FindByID(ctx, run.ID)
	if completed.State != execution.StateCompleted {
		t.Errorf("State = %s, want %s", completed.State, execution.StateCompleted)
	}
	if completed.Duration == nil {
		t.Error("Duration should be set")
	} else if *completed.Duration != duration {
		t.Errorf("Duration = %v, want %v", *completed.Duration, duration)
	}
}

// ErrConnectionNotFound is returned when a connection is not found.
var ErrConnectionNotFound = errors.New("connection not found")

// mockConnectionRepository is a mock connection repository.
type mockConnectionRepository struct {
	connections map[string]connection.Connection
}

func newMockConnectionRepository() *mockConnectionRepository {
	return &mockConnectionRepository{
		connections: make(map[string]connection.Connection),
	}
}

func (m *mockConnectionRepository) Save(ctx context.Context, conn connection.Connection) error {
	m.connections[conn.GetID()] = conn
	return nil
}

func (m *mockConnectionRepository) FindByID(ctx context.Context, id string) (connection.Connection, error) {
	conn, ok := m.connections[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}
	return conn, nil
}

func (m *mockConnectionRepository) FindAll(ctx context.Context) ([]connection.Connection, error) {
	var result []connection.Connection
	for _, conn := range m.connections {
		result = append(result, conn)
	}
	return result, nil
}

func (m *mockConnectionRepository) Delete(ctx context.Context, id string) error {
	delete(m.connections, id)
	return nil
}

func (m *mockConnectionRepository) ExistsByName(ctx context.Context, name string, excludeID string) (bool, error) {
	for id, conn := range m.connections {
		if conn.GetName() == name && id != excludeID {
			return true, nil
		}
	}
	return false, nil
}

// mockTemplateRepositoryForBenchmark is a simplified mock template repository.
type mockTemplateRepositoryForBenchmark struct {
	templates map[string]*domaintemplate.Template
}

func newMockTemplateRepositoryForBenchmark() *mockTemplateRepositoryForBenchmark {
	return &mockTemplateRepositoryForBenchmark{
		templates: make(map[string]*domaintemplate.Template),
	}
}

func (m *mockTemplateRepositoryForBenchmark) Save(ctx context.Context, tmpl *domaintemplate.Template) error {
	m.templates[tmpl.ID] = tmpl
	return nil
}

func (m *mockTemplateRepositoryForBenchmark) FindByID(ctx context.Context, id string) (*domaintemplate.Template, error) {
	tmpl, ok := m.templates[id]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	return tmpl, nil
}

func (m *mockTemplateRepositoryForBenchmark) FindAll(ctx context.Context) ([]*domaintemplate.Template, error) {
	var result []*domaintemplate.Template
	for _, tmpl := range m.templates {
		result = append(result, tmpl)
	}
	return result, nil
}

func (m *mockTemplateRepositoryForBenchmark) FindBuiltin(ctx context.Context) ([]*domaintemplate.Template, error) {
	return m.FindAll(ctx)
}

func (m *mockTemplateRepositoryForBenchmark) FindCustom(ctx context.Context) ([]*domaintemplate.Template, error) {
	return m.FindAll(ctx)
}

func (m *mockTemplateRepositoryForBenchmark) Delete(ctx context.Context, id string) error {
	delete(m.templates, id)
	return nil
}

func (m *mockTemplateRepositoryForBenchmark) LoadBuiltinTemplates(ctx context.Context, templates []*domaintemplate.Template) error {
	for _, tmpl := range templates {
		m.templates[tmpl.ID] = tmpl
	}
	return nil
}
