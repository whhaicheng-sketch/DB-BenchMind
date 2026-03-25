package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

func TestAutoBenchSuiteRunner_RunSuiteExecutesItemsSequentiallyUsingExistingBenchmarkAbility(t *testing.T) {
	ctx := context.Background()
	suiteUC := NewAutoBenchSuiteUseCase()

	suite, err := suiteUC.CreateSuite(ctx, CreateSuiteInput{
		Name:          "nightly-suite",
		ConnectionIDs: []string{"conn-mysql", "conn-oracle"},
		Profiles: []domainautobench.ProfileType{
			domainautobench.ProfileCPU,
			domainautobench.ProfileTest,
		},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	runner := NewAutoBenchSuiteRunner(
		suiteUC,
		&stubAutoBenchBenchmarkRunner{
			statusesByRunID: map[string][]execution.RunState{
				"run-1": {execution.StateRunning, execution.StateCompleted},
				"run-2": {execution.StateRunning, execution.StateCompleted},
				"run-3": {execution.StateRunning, execution.StateCompleted},
				"run-4": {execution.StateRunning, execution.StateCompleted},
			},
		},
		&stubAutoBenchConnectionProvider{
			connections: map[string]connection.Connection{
				"conn-mysql":  &connection.MySQLConnection{BaseConnection: connection.BaseConnection{ID: "conn-mysql", Name: "MySQL"}, Host: "127.0.0.1", Port: 3306, Database: "bench", Username: "root"},
				"conn-oracle": &connection.OracleConnection{BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"}, Host: "127.0.0.1", Port: 1521, ServiceName: "ORCL", Username: "system"},
			},
		},
		&stubAutoBenchTemplateProvider{
			templates: []*domaintemplate.Template{
				{ID: "oracle_cpu_bound", Name: "Oracle CPU", DBFamily: "oracle", ProfileType: "cpu_bound", IsBuiltin: true},
				{ID: "mysql_test", Name: "MySQL Test", DBFamily: "mysql", ProfileType: "test", IsBuiltin: true},
				{ID: "oracle_test", Name: "Oracle Test", DBFamily: "oracle", ProfileType: "test", IsBuiltin: true},
				{ID: "mysql_cpu_bound", Name: "MySQL CPU", DBFamily: "mysql", ProfileType: "cpu_bound", IsBuiltin: true},
			},
		},
	)
	runner.waitInterval = 0
	runner.waitForNextPoll = func(context.Context, time.Duration) error { return nil }

	if err := runner.RunSuite(ctx, suite.ID); err != nil {
		t.Fatalf("RunSuite() failed: %v", err)
	}

	benchmark := runner.benchmark.(*stubAutoBenchBenchmarkRunner)
	if len(benchmark.startedTasks) != 4 {
		t.Fatalf("len(startedTasks) = %d, want 4", len(benchmark.startedTasks))
	}

	expectedLaunches := []struct {
		connectionID string
		templateID   string
	}{
		{"conn-mysql", "mysql_test"},
		{"conn-mysql", "mysql_cpu_bound"},
		{"conn-oracle", "oracle_test"},
		{"conn-oracle", "oracle_cpu_bound"},
	}
	for i, want := range expectedLaunches {
		task := benchmark.startedTasks[i]
		if task.ConnectionID != want.connectionID {
			t.Fatalf("startedTasks[%d].ConnectionID = %q, want %q", i, task.ConnectionID, want.connectionID)
		}
		if task.TemplateID != want.templateID {
			t.Fatalf("startedTasks[%d].TemplateID = %q, want %q", i, task.TemplateID, want.templateID)
		}
	}

	status, err := suiteUC.GetSuiteStatus(ctx, suite.ID)
	if err != nil {
		t.Fatalf("GetSuiteStatus() failed: %v", err)
	}
	if status.Status != domainautobench.SuiteStatusSuccess {
		t.Fatalf("Status = %q, want %q", status.Status, domainautobench.SuiteStatusSuccess)
	}
	for i, item := range status.Items {
		if item.Status != domainautobench.SuiteItemStatusSuccess {
			t.Fatalf("Items[%d].Status = %q, want %q", i, item.Status, domainautobench.SuiteItemStatusSuccess)
		}
		if item.TemplateID == "" {
			t.Fatalf("Items[%d].TemplateID should be populated", i)
		}
		if item.LinkedTaskID == "" {
			t.Fatalf("Items[%d].LinkedTaskID should be populated", i)
		}
	}
}

func TestAutoBenchSuiteRunner_RunSuiteAppliesContinueByConnectionFailurePolicy(t *testing.T) {
	ctx := context.Background()
	suiteUC := NewAutoBenchSuiteUseCase()

	suite, err := suiteUC.CreateSuite(ctx, CreateSuiteInput{
		Name:          "continue-by-connection",
		ConnectionIDs: []string{"conn-mysql", "conn-oracle"},
		Profiles: []domainautobench.ProfileType{
			domainautobench.ProfileTest,
			domainautobench.ProfileCPU,
		},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	runner := NewAutoBenchSuiteRunner(
		suiteUC,
		&stubAutoBenchBenchmarkRunner{
			statusesByRunID: map[string][]execution.RunState{
				"run-1": {execution.StateRunning, execution.StateFailed},
				"run-2": {execution.StateRunning, execution.StateCompleted},
				"run-3": {execution.StateRunning, execution.StateCompleted},
			},
		},
		&stubAutoBenchConnectionProvider{
			connections: map[string]connection.Connection{
				"conn-mysql":  &connection.MySQLConnection{BaseConnection: connection.BaseConnection{ID: "conn-mysql", Name: "MySQL"}, Host: "127.0.0.1", Port: 3306, Database: "bench", Username: "root"},
				"conn-oracle": &connection.OracleConnection{BaseConnection: connection.BaseConnection{ID: "conn-oracle", Name: "Oracle"}, Host: "127.0.0.1", Port: 1521, ServiceName: "ORCL", Username: "system"},
			},
		},
		&stubAutoBenchTemplateProvider{
			templates: []*domaintemplate.Template{
				{ID: "mysql_test", Name: "MySQL Test", DBFamily: "mysql", ProfileType: "test", IsBuiltin: true},
				{ID: "mysql_cpu_bound", Name: "MySQL CPU", DBFamily: "mysql", ProfileType: "cpu_bound", IsBuiltin: true},
				{ID: "oracle_test", Name: "Oracle Test", DBFamily: "oracle", ProfileType: "test", IsBuiltin: true},
				{ID: "oracle_cpu_bound", Name: "Oracle CPU", DBFamily: "oracle", ProfileType: "cpu_bound", IsBuiltin: true},
			},
		},
	)
	runner.waitInterval = 0
	runner.waitForNextPoll = func(context.Context, time.Duration) error { return nil }

	if err := runner.RunSuite(ctx, suite.ID); err != nil {
		t.Fatalf("RunSuite() failed: %v", err)
	}

	benchmark := runner.benchmark.(*stubAutoBenchBenchmarkRunner)
	if len(benchmark.startedTasks) != 3 {
		t.Fatalf("len(startedTasks) = %d, want 3", len(benchmark.startedTasks))
	}
	expectedLaunches := []struct {
		connectionID string
		templateID   string
	}{
		{"conn-mysql", "mysql_test"},
		{"conn-oracle", "oracle_test"},
		{"conn-oracle", "oracle_cpu_bound"},
	}
	for i, want := range expectedLaunches {
		task := benchmark.startedTasks[i]
		if task.ConnectionID != want.connectionID {
			t.Fatalf("startedTasks[%d].ConnectionID = %q, want %q", i, task.ConnectionID, want.connectionID)
		}
		if task.TemplateID != want.templateID {
			t.Fatalf("startedTasks[%d].TemplateID = %q, want %q", i, task.TemplateID, want.templateID)
		}
	}

	status, statusErr := suiteUC.GetSuiteStatus(ctx, suite.ID)
	if statusErr != nil {
		t.Fatalf("GetSuiteStatus() failed: %v", statusErr)
	}
	if status.Status != domainautobench.SuiteStatusPartialSuccess {
		t.Fatalf("Status = %q, want %q", status.Status, domainautobench.SuiteStatusPartialSuccess)
	}
	if status.Items[0].Status != domainautobench.SuiteItemStatusFailed {
		t.Fatalf("Items[0].Status = %q, want %q", status.Items[0].Status, domainautobench.SuiteItemStatusFailed)
	}
	if status.Items[1].Status != domainautobench.SuiteItemStatusSkipped {
		t.Fatalf("Items[1].Status = %q, want %q", status.Items[1].Status, domainautobench.SuiteItemStatusSkipped)
	}
	if status.Items[2].Status != domainautobench.SuiteItemStatusSuccess {
		t.Fatalf("Items[2].Status = %q, want %q", status.Items[2].Status, domainautobench.SuiteItemStatusSuccess)
	}
	if status.Items[3].Status != domainautobench.SuiteItemStatusSuccess {
		t.Fatalf("Items[3].Status = %q, want %q", status.Items[3].Status, domainautobench.SuiteItemStatusSuccess)
	}
}

func TestAutoBenchSuiteRunner_RunSuiteMarksSuiteCancelledWhenContextInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	suiteUC := NewAutoBenchSuiteUseCase()

	suite, err := suiteUC.CreateSuite(ctx, CreateSuiteInput{
		Name:          "cancelled-suite",
		ConnectionIDs: []string{"conn-mysql"},
		Profiles:      []domainautobench.ProfileType{domainautobench.ProfileTest},
	})
	if err != nil {
		t.Fatalf("CreateSuite() failed: %v", err)
	}

	runner := NewAutoBenchSuiteRunner(
		suiteUC,
		&stubAutoBenchBenchmarkRunner{
			statusesByRunID: map[string][]execution.RunState{
				"run-1": {execution.StateRunning, execution.StateRunning},
			},
		},
		&stubAutoBenchConnectionProvider{
			connections: map[string]connection.Connection{
				"conn-mysql": &connection.MySQLConnection{BaseConnection: connection.BaseConnection{ID: "conn-mysql", Name: "MySQL"}, Host: "127.0.0.1", Port: 3306, Database: "bench", Username: "root"},
			},
		},
		&stubAutoBenchTemplateProvider{
			templates: []*domaintemplate.Template{
				{ID: "mysql_test", Name: "MySQL Test", DBFamily: "mysql", ProfileType: "test", IsBuiltin: true},
			},
		},
	)
	runner.waitInterval = 0
	runner.waitForNextPoll = func(context.Context, time.Duration) error {
		cancel()
		return context.Canceled
	}

	err = runner.RunSuite(ctx, suite.ID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunSuite() error = %v, want context.Canceled", err)
	}

	status, statusErr := suiteUC.GetSuiteStatus(context.Background(), suite.ID)
	if statusErr != nil {
		t.Fatalf("GetSuiteStatus() failed: %v", statusErr)
	}
	if status.Status != domainautobench.SuiteStatusCancelled {
		t.Fatalf("Status = %q, want %q", status.Status, domainautobench.SuiteStatusCancelled)
	}
}

func TestAutoBenchSuiteRunner_SelectTemplateIDForItemPrefersBuiltinCandidate(t *testing.T) {
	runner := NewAutoBenchSuiteRunner(
		NewAutoBenchSuiteUseCase(),
		&stubAutoBenchBenchmarkRunner{},
		&stubAutoBenchConnectionProvider{},
		&stubAutoBenchTemplateProvider{
			templates: []*domaintemplate.Template{
				{ID: "mysql_test_custom", Name: "Custom MySQL Test", DBFamily: "mysql", ProfileType: "test", IsBuiltin: false},
				{ID: "mysql_test_builtin", Name: "Builtin MySQL Test", DBFamily: "mysql", ProfileType: "test", IsBuiltin: true},
			},
		},
	)

	templateID, err := runner.selectTemplateIDForItem(context.Background(), "mysql", domainautobench.ProfileTest)
	if err != nil {
		t.Fatalf("selectTemplateIDForItem() failed: %v", err)
	}
	if templateID != "mysql_test_builtin" {
		t.Fatalf("templateID = %q, want %q", templateID, "mysql_test_builtin")
	}
}

type stubAutoBenchBenchmarkRunner struct {
	startedTasks    []*execution.BenchmarkTask
	statusesByRunID map[string][]execution.RunState
	runIndex        int
}

func (s *stubAutoBenchBenchmarkRunner) StartBenchmark(ctx context.Context, task *execution.BenchmarkTask) (*execution.Run, error) {
	_ = ctx
	s.runIndex++
	runID := fmt.Sprintf("run-%d", s.runIndex)
	s.startedTasks = append(s.startedTasks, task)
	if _, ok := s.statusesByRunID[runID]; !ok {
		return nil, errors.New("missing stub run state sequence")
	}
	return &execution.Run{
		ID:     runID,
		TaskID: task.ID,
		State:  execution.StatePending,
	}, nil
}

func (s *stubAutoBenchBenchmarkRunner) GetBenchmarkStatus(ctx context.Context, runID string) (*execution.Run, error) {
	_ = ctx
	sequence, ok := s.statusesByRunID[runID]
	if !ok || len(sequence) == 0 {
		return nil, errors.New("missing run status")
	}
	state := sequence[0]
	if len(sequence) > 1 {
		s.statusesByRunID[runID] = sequence[1:]
	}
	return &execution.Run{ID: runID, State: state}, nil
}

type stubAutoBenchConnectionProvider struct {
	connections map[string]connection.Connection
}

func (s *stubAutoBenchConnectionProvider) GetConnectionByID(ctx context.Context, id string) (connection.Connection, error) {
	_ = ctx
	conn, ok := s.connections[id]
	if !ok {
		return nil, errors.New("connection not found")
	}
	return conn, nil
}

type stubAutoBenchTemplateProvider struct {
	templates []*domaintemplate.Template
}

func (s *stubAutoBenchTemplateProvider) ListTemplates(ctx context.Context) ([]*domaintemplate.Template, error) {
	_ = ctx
	return s.templates, nil
}
