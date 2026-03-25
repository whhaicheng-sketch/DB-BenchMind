package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

var ErrAutoBenchTemplateNotFound = errors.New("autobench template not found")
var ErrAutoBenchBenchmarkRunnerRequired = errors.New("autobench benchmark runner is required")
var ErrAutoBenchConnectionProviderRequired = errors.New("autobench connection provider is required")
var ErrAutoBenchTemplateProviderRequired = errors.New("autobench template provider is required")

type autoBenchBenchmarkRunner interface {
	StartBenchmark(ctx context.Context, task *execution.BenchmarkTask) (*execution.Run, error)
	GetBenchmarkStatus(ctx context.Context, runID string) (*execution.Run, error)
}

type autoBenchConnectionProvider interface {
	GetConnectionByID(ctx context.Context, id string) (connection.Connection, error)
}

type autoBenchTemplateProvider interface {
	ListTemplates(ctx context.Context) ([]*domaintemplate.Template, error)
}

type AutoBenchSuiteRunner struct {
	suites          *AutoBenchSuiteUseCase
	benchmark       autoBenchBenchmarkRunner
	connections     autoBenchConnectionProvider
	templates       autoBenchTemplateProvider
	waitInterval    time.Duration
	waitForNextPoll func(ctx context.Context, interval time.Duration) error
}

func NewAutoBenchSuiteRunner(
	suites *AutoBenchSuiteUseCase,
	benchmark autoBenchBenchmarkRunner,
	connections autoBenchConnectionProvider,
	templates autoBenchTemplateProvider,
) *AutoBenchSuiteRunner {
	return &AutoBenchSuiteRunner{
		suites:          suites,
		benchmark:       benchmark,
		connections:     connections,
		templates:       templates,
		waitInterval:    100 * time.Millisecond,
		waitForNextPoll: waitForNextAutoBenchPoll,
	}
}

func (r *AutoBenchSuiteRunner) RunSuite(ctx context.Context, suiteID string) error {
	if r.benchmark == nil {
		return ErrAutoBenchBenchmarkRunnerRequired
	}
	if r.connections == nil {
		return ErrAutoBenchConnectionProviderRequired
	}
	if r.templates == nil {
		return ErrAutoBenchTemplateProviderRequired
	}

	suite, err := r.suites.getSuite(suiteID)
	if err != nil {
		return err
	}
	connectionFailure := map[string]bool{}

	if err := r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
		suite.Status = domainautobench.SuiteStatusRunning
		return nil
	}); err != nil {
		return err
	}

	for _, item := range suite.Items {
		if err := ctx.Err(); err != nil {
			_ = r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
				suite.Status = domainautobench.SuiteStatusCancelled
				return nil
			})
			return err
		}
		if connectionFailure[item.ConnectionID] {
			if err := r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
				target, findErr := findSuiteItemByID(suite.Items, item.ID)
				if findErr != nil {
					return findErr
				}
				target.Status = domainautobench.SuiteItemStatusSkipped
				target.ErrorSummary = "skipped after earlier connection failure"
				return nil
			}); err != nil {
				return err
			}
			continue
		}

		conn, err := r.connections.GetConnectionByID(ctx, item.ConnectionID)
		if err != nil {
			connectionFailure[item.ConnectionID] = true
			if failErr := r.markSuiteItemFailed(suiteID, item.ID, fmt.Sprintf("get connection: %v", err)); failErr != nil {
				return failErr
			}
			if suite.ExecutionPolicy.FailurePolicy != domainautobench.FailurePolicyContinueByConnection {
				return err
			}
			continue
		}

		dbType := string(conn.GetType())
		templateID, err := r.selectTemplateIDForItem(ctx, dbType, item.ProfileType)
		if err != nil {
			connectionFailure[item.ConnectionID] = true
			if failErr := r.markSuiteItemFailed(suiteID, item.ID, err.Error()); failErr != nil {
				return failErr
			}
			if suite.ExecutionPolicy.FailurePolicy != domainautobench.FailurePolicyContinueByConnection {
				return err
			}
			continue
		}

		if err := r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
			target, findErr := findSuiteItemByID(suite.Items, item.ID)
			if findErr != nil {
				return findErr
			}
			target.DatabaseType = dbType
			target.TemplateID = templateID
			target.Status = domainautobench.SuiteItemStatusRunning
			target.PhaseStatus = execution.StatePending.String()
			return nil
		}); err != nil {
			return err
		}

		task := &execution.BenchmarkTask{
			ID:           uuid.New().String(),
			Name:         fmt.Sprintf("AutoBench %s %s %s", strings.TrimSpace(suite.Name), item.ConnectionID, item.ProfileType),
			ConnectionID: item.ConnectionID,
			TemplateID:   templateID,
			Parameters:   map[string]interface{}{},
			Options: execution.TaskOptions{
				SkipCleanup: !suite.ExecutionPolicy.CleanupEnabled,
			},
			Tags: []string{
				"autobench",
				fmt.Sprintf("suite:%s", suiteID),
				fmt.Sprintf("item:%s", item.ID),
				fmt.Sprintf("profile:%s", item.ProfileType),
			},
			CreatedAt: time.Now(),
		}

		run, err := r.benchmark.StartBenchmark(ctx, task)
		if err != nil {
			connectionFailure[item.ConnectionID] = true
			if failErr := r.markSuiteItemFailed(suiteID, item.ID, fmt.Sprintf("start benchmark: %v", err)); failErr != nil {
				return failErr
			}
			if suite.ExecutionPolicy.FailurePolicy != domainautobench.FailurePolicyContinueByConnection {
				return err
			}
			continue
		}

		if err := r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
			target, findErr := findSuiteItemByID(suite.Items, item.ID)
			if findErr != nil {
				return findErr
			}
			target.LinkedTaskID = task.ID
			target.PhaseStatus = run.State.String()
			return nil
		}); err != nil {
			return err
		}

		finalRun, err := r.waitForRunCompletion(ctx, run.ID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				_ = r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
					suite.Status = domainautobench.SuiteStatusCancelled
					target, findErr := findSuiteItemByID(suite.Items, item.ID)
					if findErr != nil {
						return findErr
					}
					target.PhaseStatus = "cancelled"
					return nil
				})
				return err
			}
			connectionFailure[item.ConnectionID] = true
			if failErr := r.markSuiteItemFailed(suiteID, item.ID, err.Error()); failErr != nil {
				return failErr
			}
			if suite.ExecutionPolicy.FailurePolicy != domainautobench.FailurePolicyContinueByConnection {
				return err
			}
			continue
		}

		if finalRun.State == execution.StateCompleted {
			if err := r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
				target, findErr := findSuiteItemByID(suite.Items, item.ID)
				if findErr != nil {
					return findErr
				}
				target.Status = domainautobench.SuiteItemStatusSuccess
				target.PhaseStatus = finalRun.State.String()
				return nil
			}); err != nil {
				return err
			}
			continue
		}

		err = fmt.Errorf("benchmark run ended with state %s", finalRun.State)
		connectionFailure[item.ConnectionID] = true
		if failErr := r.markSuiteItemFailed(suiteID, item.ID, err.Error()); failErr != nil {
			return failErr
		}
		if suite.ExecutionPolicy.FailurePolicy != domainautobench.FailurePolicyContinueByConnection {
			return err
		}
	}

	return r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
		suite.Status = summarizeSuiteStatus(suite.Items)
		return nil
	})
}

func (r *AutoBenchSuiteRunner) waitForRunCompletion(ctx context.Context, runID string) (*execution.Run, error) {
	for {
		run, err := r.benchmark.GetBenchmarkStatus(ctx, runID)
		if err != nil {
			return nil, err
		}
		if run.IsCompleted() {
			return run, nil
		}
		if err := r.waitForNextPoll(ctx, r.waitInterval); err != nil {
			return nil, err
		}
	}
}

func waitForNextAutoBenchPoll(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return nil
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (r *AutoBenchSuiteRunner) selectTemplateIDForItem(ctx context.Context, dbType string, profile domainautobench.ProfileType) (string, error) {
	templates, err := r.templates.ListTemplates(ctx)
	if err != nil {
		return "", err
	}

	candidates := make([]*domaintemplate.Template, 0, len(templates))
	for _, tmpl := range templates {
		if tmpl == nil {
			continue
		}
		if !tmpl.SupportsDatabase(dbType) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(tmpl.ProfileType), string(profile)) {
			continue
		}
		candidates = append(candidates, tmpl)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("%w: %s/%s", ErrAutoBenchTemplateNotFound, dbType, profile)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].IsBuiltin != candidates[j].IsBuiltin {
			return candidates[i].IsBuiltin
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0].ID, nil
}

func (r *AutoBenchSuiteRunner) markSuiteItemFailed(suiteID, itemID, summary string) error {
	return r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
		suite.Status = domainautobench.SuiteStatusFailed
		target, err := findSuiteItemByID(suite.Items, itemID)
		if err != nil {
			return err
		}
		target.Status = domainautobench.SuiteItemStatusFailed
		target.ErrorSummary = summary
		return nil
	})
}

func (uc *AutoBenchSuiteUseCase) mutateSuite(suiteID string, mutator func(*domainautobench.Suite) error) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	suite, ok := uc.suites[strings.TrimSpace(suiteID)]
	if !ok {
		return fmt.Errorf("%w: %s", ErrAutoBenchSuiteNotFound, suiteID)
	}
	suite = cloneSuite(suite)
	if err := mutator(&suite); err != nil {
		return err
	}
	uc.suites[suiteID] = cloneSuite(suite)
	return nil
}

func findSuiteItemByID(items []domainautobench.SuiteItem, itemID string) (*domainautobench.SuiteItem, error) {
	for i := range items {
		if items[i].ID == itemID {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("suite item not found: %s", itemID)
}

func summarizeSuiteStatus(items []domainautobench.SuiteItem) domainautobench.SuiteStatus {
	hasFailedOrSkipped := false
	hasSuccess := false

	for _, item := range items {
		switch item.Status {
		case domainautobench.SuiteItemStatusSuccess:
			hasSuccess = true
		case domainautobench.SuiteItemStatusFailed, domainautobench.SuiteItemStatusSkipped:
			hasFailedOrSkipped = true
		case domainautobench.SuiteItemStatusPending, domainautobench.SuiteItemStatusRunning, domainautobench.SuiteItemStatusPreparing, domainautobench.SuiteItemStatusCleaning, domainautobench.SuiteItemStatusValidating:
			return domainautobench.SuiteStatusRunning
		}
	}

	if hasFailedOrSkipped && hasSuccess {
		return domainautobench.SuiteStatusPartialSuccess
	}
	if hasFailedOrSkipped {
		return domainautobench.SuiteStatusFailed
	}
	return domainautobench.SuiteStatusSuccess
}
