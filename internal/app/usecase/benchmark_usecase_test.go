// Package usecase provides unit tests for benchmark use case.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
)

func TestIsIgnorableSampleCollectorError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "closed file error is ignored",
			err:  fmt.Errorf("scan stdout: %w", fs.ErrClosed),
			want: true,
		},
		{
			name: "context canceled is ignored",
			err:  context.Canceled,
			want: true,
		},
		{
			name: "deadline exceeded is ignored",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "real scanner failure is not ignored",
			err:  errors.New("scan stdout: token too long"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIgnorableSampleCollectorError(tt.err); got != tt.want {
				t.Fatalf("isIgnorableSampleCollectorError() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	logs map[string][]LogEntry
}

var (
	// ErrRunNotFound is returned when a run is not found.
	ErrRunNotFound = errors.New("run not found")
)

func newMockRunRepository() *mockRunRepository {
	return &mockRunRepository{
		runs: make(map[string]*execution.Run),
		logs: make(map[string][]LogEntry),
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
	m.logs[runID] = append(m.logs[runID], entry)
	return nil
}

func (m *mockRunRepository) Delete(ctx context.Context, id string) error {
	delete(m.runs, id)
	return nil
}

type successfulTestConnection struct {
	id   string
	name string
}

func (c *successfulTestConnection) GetID() string       { return c.id }
func (c *successfulTestConnection) GetName() string     { return c.name }
func (c *successfulTestConnection) SetName(name string) { c.name = name }
func (c *successfulTestConnection) GetType() connection.DatabaseType {
	return connection.DatabaseTypeMySQL
}
func (c *successfulTestConnection) Validate() error { return nil }
func (c *successfulTestConnection) Test(ctx context.Context) (*connection.TestResult, error) {
	return &connection.TestResult{Success: true}, nil
}
func (c *successfulTestConnection) GetDSN() string             { return "test-dsn" }
func (c *successfulTestConnection) GetDSNWithPassword() string { return "test-dsn-with-password" }
func (c *successfulTestConnection) Redact() string             { return "test-redacted" }
func (c *successfulTestConnection) ToJSON() ([]byte, error)    { return []byte(`{}`), nil }
func (c *successfulTestConnection) GetAIAssistants() []connection.AIAssistantConfig {
	return nil
}
func (c *successfulTestConnection) SetAIAssistants(assistants []connection.AIAssistantConfig) {}

type prepareOnlySequenceAdapter struct {
	prepare *adapter.Command
	cleanup *adapter.Command
}

func (a *prepareOnlySequenceAdapter) Type() adapter.AdapterType { return adapter.AdapterTypeSysbench }
func (a *prepareOnlySequenceAdapter) BuildPrepareCommand(ctx context.Context, config *adapter.Config) (*adapter.Command, error) {
	return a.prepare, nil
}
func (a *prepareOnlySequenceAdapter) BuildRunCommand(ctx context.Context, config *adapter.Config) (*adapter.Command, error) {
	return &adapter.Command{CmdLine: "bash -lc 'exit 0'", WorkDir: config.WorkDir}, nil
}
func (a *prepareOnlySequenceAdapter) BuildCleanupCommand(ctx context.Context, config *adapter.Config) (*adapter.Command, error) {
	return a.cleanup, nil
}
func (a *prepareOnlySequenceAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*adapter.Result, error) {
	return &adapter.Result{}, nil
}
func (a *prepareOnlySequenceAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader) (<-chan adapter.Sample, <-chan error, *strings.Builder) {
	sampleCh := make(chan adapter.Sample)
	errCh := make(chan error)
	var buf strings.Builder
	close(sampleCh)
	close(errCh)
	return sampleCh, errCh, &buf
}
func (a *prepareOnlySequenceAdapter) ValidateConfig(ctx context.Context, config *adapter.Config) error {
	return nil
}
func (a *prepareOnlySequenceAdapter) ParseFinalResults(ctx context.Context, stdout string) (*adapter.FinalResult, error) {
	return &adapter.FinalResult{}, nil
}
func (a *prepareOnlySequenceAdapter) SupportsDatabase(dbType connection.DatabaseType) bool {
	return true
}

type failingPrepareAdapter struct {
	err error
}

func (a *failingPrepareAdapter) Type() adapter.AdapterType { return adapter.AdapterTypeSwingbench }
func (a *failingPrepareAdapter) BuildPrepareCommand(ctx context.Context, config *adapter.Config) (*adapter.Command, error) {
	return nil, a.err
}
func (a *failingPrepareAdapter) BuildRunCommand(ctx context.Context, config *adapter.Config) (*adapter.Command, error) {
	return nil, errors.New("not implemented")
}
func (a *failingPrepareAdapter) BuildCleanupCommand(ctx context.Context, config *adapter.Config) (*adapter.Command, error) {
	return nil, errors.New("not implemented")
}
func (a *failingPrepareAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*adapter.Result, error) {
	return &adapter.Result{}, nil
}
func (a *failingPrepareAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader) (<-chan adapter.Sample, <-chan error, *strings.Builder) {
	sampleCh := make(chan adapter.Sample)
	errCh := make(chan error)
	var buf strings.Builder
	close(sampleCh)
	close(errCh)
	return sampleCh, errCh, &buf
}
func (a *failingPrepareAdapter) ValidateConfig(ctx context.Context, config *adapter.Config) error {
	return nil
}
func (a *failingPrepareAdapter) ParseFinalResults(ctx context.Context, stdout string) (*adapter.FinalResult, error) {
	return &adapter.FinalResult{}, nil
}
func (a *failingPrepareAdapter) SupportsDatabase(dbType connection.DatabaseType) bool {
	return true
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

func TestResolvePrepareThreads_DefaultsToFourWithoutSSH(t *testing.T) {
	threads, err := resolvePrepareThreads(context.Background(), &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-no-ssh"},
		Host:           "localhost",
		Port:           3306,
		Database:       "db",
		Username:       "root",
	})
	if err != nil {
		t.Fatalf("resolvePrepareThreads() failed: %v", err)
	}
	if threads != 4 {
		t.Fatalf("resolvePrepareThreads() = %d, want 4", threads)
	}
}

func TestResolvePrepareThreads_UsesHalfRemoteCPUCappedAt32(t *testing.T) {
	restore := benchmarkPrepareRemoteCPUCount
	benchmarkPrepareRemoteCPUCount = func(ctx context.Context, cfg *connection.SSHTunnelConfig) (int, error) {
		return 96, nil
	}
	defer func() {
		benchmarkPrepareRemoteCPUCount = restore
	}()

	threads, err := resolvePrepareThreads(context.Background(), &connection.SQLServerConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-sqlserver-ssh"},
		Host:           "db.internal",
		Port:           1433,
		Database:       "tpcc",
		Username:       "sa",
		SSH: &connection.SSHTunnelConfig{
			Enabled:  true,
			Host:     "ssh.internal",
			Port:     22,
			Username: "ops",
			Password: "secret",
		},
	})
	if err != nil {
		t.Fatalf("resolvePrepareThreads() failed: %v", err)
	}
	if threads != 32 {
		t.Fatalf("resolvePrepareThreads() = %d, want 32", threads)
	}
}

func TestResolvePrepareThreads_UsesHalfRemoteCPUMinimumOne(t *testing.T) {
	restore := benchmarkPrepareRemoteCPUCount
	benchmarkPrepareRemoteCPUCount = func(ctx context.Context, cfg *connection.SSHTunnelConfig) (int, error) {
		return 1, nil
	}
	defer func() {
		benchmarkPrepareRemoteCPUCount = restore
	}()

	threads, err := resolvePrepareThreads(context.Background(), &connection.PostgreSQLConnection{
		BaseConnection: connection.BaseConnection{ID: "conn-pg-ssh"},
		Host:           "db.internal",
		Port:           5432,
		Database:       "bench",
		Username:       "postgres",
		SSH: &connection.SSHTunnelConfig{
			Enabled:  true,
			Host:     "ssh.internal",
			Port:     22,
			Username: "ops",
			Password: "secret",
		},
	})
	if err != nil {
		t.Fatalf("resolvePrepareThreads() failed: %v", err)
	}
	if threads != 1 {
		t.Fatalf("resolvePrepareThreads() = %d, want 1", threads)
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

func TestExecuteBenchmark_PrepareOnlyRunsCleanupExactlyOnce(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	tempDir := t.TempDir()
	traceFile := filepath.Join(tempDir, "prepare-trace.log")
	run := &execution.Run{
		ID:        "run-prepare-only",
		TaskID:    "task-prepare-only",
		State:     execution.StatePending,
		CreatedAt: time.Now(),
		WorkDir:   filepath.Join(tempDir, "run-workdir"),
	}
	if err := runRepo.Save(ctx, run); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	adapt := &prepareOnlySequenceAdapter{
		cleanup: &adapter.Command{
			CmdLine: "bash -lc 'printf \"cleanup-from-direct\\n\" >> \"" + traceFile + "\"'",
			WorkDir: tempDir,
		},
		prepare: &adapter.Command{
			CmdLine: "prepare-sequence",
			Commands: []*adapter.Command{
				{
					StepName: "Cleanup Existing Environment",
					CmdLine:  "bash -lc 'printf \"cleanup-from-sequence\\n\" >> \"" + traceFile + "\"'",
					WorkDir:  tempDir,
				},
				{
					StepName: "Init/Create",
					CmdLine:  "bash -lc 'printf \"create\\n\" >> \"" + traceFile + "\"'",
					WorkDir:  tempDir,
				},
			},
		},
	}

	task := &execution.BenchmarkTask{
		ID: "task-prepare-only",
		Parameters: map[string]interface{}{
			"time":           0,
			"_original_time": 60,
		},
		Options: execution.TaskOptions{SkipCleanup: true},
	}

	uc.executeBenchmark(ctx, run, &successfulTestConnection{id: "conn-1", name: "Test Conn"}, &domaintemplate.Template{
		ID:   "tpl-prepare-only",
		Name: "Prepare Only",
		Tool: domaintemplate.ToolSysbench,
	}, adapt, task)

	data, err := os.ReadFile(traceFile)
	if err != nil {
		t.Fatalf("ReadFile(trace) failed: %v", err)
	}
	lines := strings.Fields(string(data))
	cleanupCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "cleanup-") {
			cleanupCount++
		}
	}
	if cleanupCount != 1 {
		t.Fatalf("cleanup executions = %d, want 1, trace = %q", cleanupCount, string(data))
	}

	storedRun, err := runRepo.FindByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if storedRun.State != execution.StateCompleted {
		t.Fatalf("prepare-only final state = %s, want %s", storedRun.State, execution.StateCompleted)
	}
}

func TestExecutePhase_LogsPrepareCommandConstructionFailureReason(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	run := &execution.Run{
		ID:        "run-build-prepare-fail",
		TaskID:    "task-build-prepare-fail",
		State:     execution.StatePending,
		CreatedAt: time.Now(),
		WorkDir:   t.TempDir(),
	}
	if err := runRepo.Save(ctx, run); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	config := &adapter.Config{
		Connection: &connection.OracleConnection{
			BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"},
			Host:           "localhost",
			Port:           1521,
			ServiceName:    "ORCL",
			Username:       "system",
			Password:       "manager",
		},
		Parameters: map[string]interface{}{},
		WorkDir:    run.WorkDir,
	}

	err := uc.executePhase(ctx, run, &failingPrepareAdapter{
		err: errors.New("missing sysdba_password for bootstrap override"),
	}, config, "prepare", execution.StatePreparing, execution.StatePrepared)
	if err == nil {
		t.Fatal("executePhase() expected error")
	}

	logText := flattenLogEntries(runRepo.logs[run.ID])
	for _, want := range []string{
		"Starting phase: prepare",
		"Building prepare phase command",
		"Failed to build prepare phase command",
		"missing sysdba_password for bootstrap override",
	} {
		if !strings.Contains(logText, want) {
			t.Fatalf("expected %q in logs, got: %s", want, logText)
		}
	}
}

func TestExecutePhase_PrepareSequenceLogsExecutionMilestones(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	tempDir := t.TempDir()
	run := &execution.Run{
		ID:        "run-prepare-sequence-log",
		TaskID:    "task-prepare-sequence-log",
		State:     execution.StatePending,
		CreatedAt: time.Now(),
		WorkDir:   tempDir,
	}
	if err := runRepo.Save(ctx, run); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := uc.executePhase(ctx, run, &prepareOnlySequenceAdapter{
		prepare: &adapter.Command{
			CmdLine: "oracle_prepare_sequence",
			Commands: []*adapter.Command{
				{
					StepName: "Cleanup Existing Environment",
					CmdLine:  "bash -lc 'echo cleanup-started'",
					WorkDir:  tempDir,
				},
				{
					StepName: "Create Schema",
					CmdLine:  "bash -lc 'echo create-started'",
					WorkDir:  tempDir,
				},
			},
		},
		cleanup: &adapter.Command{CmdLine: "bash -lc 'exit 0'", WorkDir: tempDir},
	}, &adapter.Config{
		Connection: &connection.OracleConnection{
			BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"},
			Host:           "localhost",
			Port:           1521,
			ServiceName:    "ORCL",
			Username:       "system",
			Password:       "manager",
		},
		Parameters: map[string]interface{}{},
		WorkDir:    tempDir,
	}, "prepare", execution.StatePreparing, execution.StatePrepared)
	if err != nil {
		t.Fatalf("executePhase() unexpected error: %v", err)
	}

	logText := flattenLogEntries(runRepo.logs[run.ID])
	for _, want := range []string{
		"Starting phase: prepare",
		"Executing phase command: oracle_prepare_sequence",
		"Executing command sequence (2 steps)",
		"Starting sequence step=1/2: Cleanup Existing Environment",
	} {
		if !strings.Contains(logText, want) {
			t.Fatalf("expected %q in logs, got: %s", want, logText)
		}
	}
}

func TestExecuteCommandSequence_BlocksCreateWhenCleanupVerificationFailsAndLogsReason(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	tempDir := t.TempDir()
	createMarker := filepath.Join(tempDir, "create.marker")
	run := &execution.Run{
		ID:        "run-verify-fail",
		TaskID:    "task-verify-fail",
		State:     execution.StatePreparing,
		CreatedAt: time.Now(),
		WorkDir:   tempDir,
	}
	if err := runRepo.Save(ctx, run); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := uc.executeCommandSequence(ctx, run, &adapter.Command{
		CmdLine: "oracle_prepare_sequence",
		Commands: []*adapter.Command{
			{StepName: "Cleanup Existing Environment", CmdLine: "bash -lc 'echo cleanup-complete'", WorkDir: tempDir},
			{StepName: "Verify Cleanup State", CmdLine: "bash -lc 'echo \"SOE cleanup verification failed: residual tablespaces remain\" 1>&2; exit 1'", WorkDir: tempDir},
			{StepName: "Create Schema", CmdLine: "bash -lc 'touch \"" + createMarker + "\"'", WorkDir: tempDir},
		},
	})

	if err == nil {
		t.Fatal("executeCommandSequence() expected verification failure")
	}
	if !strings.Contains(err.Error(), "Verify Cleanup State") {
		t.Fatalf("error = %q, want Verify Cleanup State context", err.Error())
	}
	if _, statErr := os.Stat(createMarker); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("create marker should not exist, stat err = %v", statErr)
	}

	logText := flattenLogEntries(runRepo.logs[run.ID])
	if !strings.Contains(logText, "SOE cleanup verification failed: residual tablespaces remain") {
		t.Fatalf("expected verification reason in logs, got: %s", logText)
	}
}

func TestExecuteCommandSequence_AllowsCreateAfterCleanupVerificationPasses(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	tempDir := t.TempDir()
	createMarker := filepath.Join(tempDir, "create.marker")
	run := &execution.Run{
		ID:        "run-verify-pass",
		TaskID:    "task-verify-pass",
		State:     execution.StatePreparing,
		CreatedAt: time.Now(),
		WorkDir:   tempDir,
	}
	if err := runRepo.Save(ctx, run); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := uc.executeCommandSequence(ctx, run, &adapter.Command{
		CmdLine: "oracle_prepare_sequence",
		Commands: []*adapter.Command{
			{StepName: "Cleanup Existing Environment", CmdLine: "bash -lc 'echo cleanup-complete'", WorkDir: tempDir},
			{StepName: "Verify Cleanup State", CmdLine: "bash -lc 'echo SOE cleanup verification passed'", WorkDir: tempDir},
			{StepName: "Create Schema", CmdLine: "bash -lc 'touch \"" + createMarker + "\"'", WorkDir: tempDir},
		},
	})

	if err != nil {
		t.Fatalf("executeCommandSequence() unexpected error: %v", err)
	}
	if _, statErr := os.Stat(createMarker); statErr != nil {
		t.Fatalf("create marker should exist after verification passed: %v", statErr)
	}
}

func TestExecuteCommandSequence_CreateFailureLogsDiagnosticOutput(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	tempDir := t.TempDir()
	run := &execution.Run{
		ID:        "run-create-fail",
		TaskID:    "task-create-fail",
		State:     execution.StatePreparing,
		CreatedAt: time.Now(),
		WorkDir:   tempDir,
	}
	if err := runRepo.Save(ctx, run); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := uc.executeCommandSequence(ctx, run, &adapter.Command{
		CmdLine: "oracle_prepare_sequence",
		Commands: []*adapter.Command{
			{StepName: "Cleanup Existing Environment", CmdLine: "bash -lc 'echo cleanup-complete'", WorkDir: tempDir},
			{StepName: "Verify Cleanup State", CmdLine: "bash -lc 'echo SOE cleanup verification passed'", WorkDir: tempDir},
			{StepName: "Create Schema", CmdLine: "bash -lc 'echo \"create stdout: started\"; echo \"ORA-01017: invalid username/password; logon denied\" 1>&2; echo \"debugf: " + filepath.Join(tempDir, "create-debug.log") + "\" 1>&2; exit 255'", WorkDir: tempDir},
		},
	})

	if err == nil {
		t.Fatal("executeCommandSequence() expected create failure")
	}

	logText := flattenLogEntries(runRepo.logs[run.ID])
	for _, want := range []string{
		"create stdout: started",
		"ORA-01017: invalid username/password; logon denied",
		"create-debug.log",
	} {
		if !strings.Contains(logText, want) {
			t.Fatalf("expected %q in logs, got: %s", want, logText)
		}
	}
}

func TestExecuteCommandSync_IncludesStdoutForSqlplusStyleFailures(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	tempDir := t.TempDir()
	run := &execution.Run{
		ID:        "run-sqlplus-stdout-fail",
		TaskID:    "task-sqlplus-stdout-fail",
		State:     execution.StatePreparing,
		CreatedAt: time.Now(),
		WorkDir:   tempDir,
	}
	require.NoError(t, runRepo.Save(ctx, run))

	err := uc.executeCommandSync(ctx, run, &adapter.Command{
		StepName: "Bootstrap SOE",
		CmdLine:  "bash -lc 'echo \"ORA-01031: insufficient privileges\"; exit 7'",
		WorkDir:  tempDir,
	})

	require.Error(t, err)
	assertErr := err.Error()
	if !strings.Contains(assertErr, "ORA-01031: insufficient privileges") {
		t.Fatalf("expected stdout diagnostics in error, got: %s", assertErr)
	}
	if !strings.Contains(assertErr, "stdout:") {
		t.Fatalf("expected stdout label in error, got: %s", assertErr)
	}
}

func TestExecuteCommandSync_FailsWhenHammerDBReportsStdoutError(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	tempDir := t.TempDir()
	run := &execution.Run{
		ID:        "run-hammerdb-stdout-fail",
		TaskID:    "task-hammerdb-stdout-fail",
		State:     execution.StatePreparing,
		CreatedAt: time.Now(),
		WorkDir:   tempDir,
	}
	require.NoError(t, runRepo.Save(ctx, run))

	err := uc.executeCommandSync(ctx, run, &adapter.Command{
		StepName: "Build schema and load data",
		CmdLine:  "bash -lc 'printf \"HammerDB CLI v4.10\\nError:Build virtual users must be less than or equal to number of warehouses\\n\"' && true # hammerdbcli",
		WorkDir:  tempDir,
	})

	require.Error(t, err)
	if !strings.Contains(err.Error(), "Build virtual users must be less than or equal to number of warehouses") {
		t.Fatalf("expected HammerDB stdout failure reason in error, got: %s", err)
	}
}

func TestIngestSwingbenchDebugLogs_CollectsExplicitAndFallbackLogs(t *testing.T) {
	ctx := context.Background()
	runRepo := newMockRunRepository()
	uc := NewBenchmarkUseCase(runRepo, adapter.NewAdapterRegistry(), nil, nil)

	runID := "run-debug-logs"
	explicitDir := t.TempDir()
	explicitPath := filepath.Join(explicitDir, "oewizard-create-debug.log")
	fallbackDir := t.TempDir()
	fallbackPath := filepath.Join(fallbackDir, "debug.log")

	require.NoError(t, os.WriteFile(explicitPath, []byte("explicit line 1\nexplicit line 2\n"), 0o644))
	require.NoError(t, os.WriteFile(fallbackPath, []byte("fallback line 1\nfallback line 2\n"), 0o644))

	restoreRoot := swingbenchDebugSearchRoot
	restoreGlob := swingbenchDebugGlob
	swingbenchDebugSearchRoot = fallbackDir
	swingbenchDebugGlob = func(pattern string) ([]string, error) {
		return []string{fallbackPath}, nil
	}
	defer func() {
		swingbenchDebugSearchRoot = restoreRoot
		swingbenchDebugGlob = restoreGlob
	}()

	startedAt := time.Now().Add(-time.Second)
	uc.ingestSwingbenchDebugLogs(ctx, runID, &adapter.Command{
		StepName: "Create Schema",
		CmdLine:  "./oewizard -debugf " + explicitPath,
	}, startedAt)

	logText := flattenLogEntries(runRepo.logs[runID])
	for _, want := range []string{
		"Swingbench debug output from " + explicitPath,
		"explicit line 1",
		"explicit line 2",
		"Swingbench debug output from " + fallbackPath,
		"fallback line 1",
		"fallback line 2",
	} {
		if !strings.Contains(logText, want) {
			t.Fatalf("expected %q in logs, got: %s", want, logText)
		}
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

func flattenLogEntries(entries []LogEntry) string {
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString(entry.Stream)
		builder.WriteString(":")
		builder.WriteString(entry.Content)
		builder.WriteString("\n")
	}
	return builder.String()
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
		{name: "sbutil shell wrapper", cmdLine: "bash -lc \"/opt/benchtools/swingbench/bin/sbutil -soe -cs //localhost:1521/ORCL -u soe -p soe -val\"", want: false},
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

func TestNormalizeSwingbenchCommandParts_RewritesBundledPathsOnly(t *testing.T) {
	parts := []string{
		"./oewizard",
		"-cl",
		"-c", "oewizard.xml",
		"-r", "results.xml",
	}

	got := normalizeSwingbenchCommandParts(parts)

	wantLauncher := filepath.Join(swingbenchInstallRoot, "launcher")
	wantConfig := filepath.Join(swingbenchBinDir, "oewizard.xml")
	if got[0] != "java" {
		t.Fatalf("entrypoint = %q, want java", got[0])
	}
	if got[2] != wantLauncher {
		t.Fatalf("launcher path = %q, want %q", got[2], wantLauncher)
	}
	configIndex := -1
	resultIndex := -1
	for i := 0; i < len(got)-1; i++ {
		switch got[i] {
		case "-c":
			configIndex = i + 1
		case "-r":
			resultIndex = i + 1
		}
	}
	if configIndex == -1 || got[configIndex] != wantConfig {
		t.Fatalf("config path = %q, want %q", got[configIndex], wantConfig)
	}
	if resultIndex == -1 || got[resultIndex] != "results.xml" {
		t.Fatalf("result file path = %q, want relative workdir output", got[resultIndex])
	}
	if parts[0] != "./oewizard" || parts[3] != "oewizard.xml" {
		t.Fatal("normalizeSwingbenchCommandParts should not mutate input slice")
	}
}

func TestNormalizeSwingbenchCommandParts_InsertsDefaultOewizardConfigWhenMissing(t *testing.T) {
	parts := []string{
		"./oewizard",
		"-cl",
		"-cs", "//db-host:1521/ORCL",
	}

	got := normalizeSwingbenchCommandParts(parts)

	wantConfig := filepath.Join(swingbenchBinDir, "oewizard.xml")
	configIndex := slices.Index(got, "-c")
	if configIndex == -1 || configIndex+1 >= len(got) {
		t.Fatalf("expected -c in normalized args, got %v", got)
	}
	if got[configIndex+1] != wantConfig {
		t.Fatalf("config path = %q, want %q", got[configIndex+1], wantConfig)
	}
	if parts[0] != "./oewizard" {
		t.Fatal("normalizeSwingbenchCommandParts should not mutate input slice")
	}
}

func TestPrepareSwingbenchRuntimeDir_CreatesSandboxLayout(t *testing.T) {
	workDir := t.TempDir()

	runtimeDir, err := prepareSwingbenchRuntimeDir(workDir)
	if err != nil {
		t.Fatalf("prepareSwingbenchRuntimeDir() error = %v", err)
	}

	wantDir := filepath.Join(workDir, "swingbench", "bin")
	if runtimeDir != wantDir {
		t.Fatalf("runtimeDir = %q, want %q", runtimeDir, wantDir)
	}

	for _, name := range []string{"sql", "configs", "launcher", "lib"} {
		target := filepath.Join(workDir, "swingbench", name)
		info, statErr := os.Lstat(target)
		if statErr != nil {
			t.Fatalf("Lstat(%q) error = %v", target, statErr)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("%q is not a symlink", target)
		}
	}

	binData := filepath.Join(workDir, "swingbench", "bin", "data")
	info, err := os.Lstat(binData)
	if err != nil {
		t.Fatalf("Lstat(%q) error = %v", binData, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%q is not a symlink", binData)
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

func TestSwingbenchNoOutputTimeoutForStep(t *testing.T) {
	tests := []struct {
		name string
		step *adapter.Command
		want time.Duration
	}{
		{
			name: "generate data has no inactivity kill timeout",
			step: &adapter.Command{StepName: "Generate Data", CmdLine: "java -cp ../launcher LauncherBootstrap -executablename oewizard oewizard"},
			want: 0,
		},
		{
			name: "create schema has no inactivity kill timeout",
			step: &adapter.Command{StepName: "Create Schema", CmdLine: "java -cp ../launcher LauncherBootstrap -executablename oewizard oewizard"},
			want: 0,
		},
		{
			name: "build indexes has no inactivity kill timeout",
			step: &adapter.Command{StepName: "Build Indexes", CmdLine: "bash -lc \"/opt/benchtools/swingbench/bin/sbutil -soe -cs //localhost:1521/ORCL -u soe -p soe -val\""},
			want: 0,
		},
		{
			name: "post schema setup has no inactivity timeout",
			step: &adapter.Command{StepName: "Post-Schema Setup", CmdLine: "sqlplus -L ..."},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := swingbenchNoOutputTimeoutForStep(tt.step); got != tt.want {
				t.Fatalf("swingbenchNoOutputTimeoutForStep() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestShouldCleanupResidualSwingbenchProcesses(t *testing.T) {
	tests := []struct {
		name string
		step *adapter.Command
		want bool
	}{
		{
			name: "oewizard generate triggers cleanup",
			step: &adapter.Command{StepName: "Generate Data", CmdLine: "java -cp ../launcher LauncherBootstrap -executablename oewizard oewizard"},
			want: true,
		},
		{
			name: "sqlplus does not trigger cleanup",
			step: &adapter.Command{StepName: "Bootstrap SOE", CmdLine: "sqlplus -L ..."},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCleanupResidualSwingbenchProcesses(tt.step); got != tt.want {
				t.Fatalf("shouldCleanupResidualSwingbenchProcesses() = %v, want %v", got, tt.want)
			}
		})
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
