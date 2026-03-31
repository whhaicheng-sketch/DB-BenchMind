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
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
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
	manifestWriter  *SuiteManifestWriter
	reportUsecase   *ReportUsecase
	waitInterval    time.Duration
	waitForNextPoll func(ctx context.Context, interval time.Duration) error
}

// AutoBenchSuiteRunnerOption configures the runner.
type AutoBenchSuiteRunnerOption func(*AutoBenchSuiteRunner)

// WithManifestWriter sets the manifest writer for persisting suite manifests.
func WithManifestWriter(w *SuiteManifestWriter) AutoBenchSuiteRunnerOption {
	return func(r *AutoBenchSuiteRunner) {
		r.manifestWriter = w
	}
}

// WithReportUsecase sets the report use case for inserting running reports.
func WithReportUsecase(uc *ReportUsecase) AutoBenchSuiteRunnerOption {
	return func(r *AutoBenchSuiteRunner) {
		r.reportUsecase = uc
	}
}

func NewAutoBenchSuiteRunner(
	suites *AutoBenchSuiteUseCase,
	benchmark autoBenchBenchmarkRunner,
	connections autoBenchConnectionProvider,
	templates autoBenchTemplateProvider,
	opts ...AutoBenchSuiteRunnerOption,
) *AutoBenchSuiteRunner {
	r := &AutoBenchSuiteRunner{
		suites:          suites,
		benchmark:       benchmark,
		connections:     connections,
		templates:       templates,
		waitInterval:    100 * time.Millisecond,
		waitForNextPoll: waitForNextAutoBenchPoll,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
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
		now := time.Now()
		suite.StartedAt = &now
		return nil
	}); err != nil {
		return err
	}

	// Write initial manifest
	r.writeManifestAsync(ctx, suiteID)
	// Persist suite status=running to DB
	r.suites.PersistSuite(ctx, suiteID)

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
		tmpl, err := r.selectTemplateForItem(ctx, dbType, item.ProfileType)
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
			target.TemplateID = tmpl.ID
			target.Status = domainautobench.SuiteItemStatusRunning
			target.PhaseStatus = execution.StatePending.String()
			now := time.Now()
			target.StartedAt = &now
			return nil
		}); err != nil {
			return err
		}

		// Insert running report row (best-effort)
		if r.reportUsecase != nil {
			rpt := &report.Report{
				ID:             uuid.New().String(),
				SuiteID:        suiteID,
				SuiteItemID:    item.ID,
				SourceType:     report.SourceTypeAutoBench,
				ConnectionID:   item.ConnectionID,
				DatabaseType:   dbType,
				Status:         report.StatusRunning,
				StartedAt:      time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			_ = r.reportUsecase.InsertRunningReport(ctx, rpt)
		}

		task := &execution.BenchmarkTask{
			ID:           uuid.New().String(),
			Name:         fmt.Sprintf("AutoBench %s %s %s", strings.TrimSpace(suite.Name), item.ConnectionID, item.ProfileType),
			ConnectionID: item.ConnectionID,
			TemplateID:   tmpl.ID,
			Parameters:   resolveDefaultRunParams(tmpl),
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

		finalRun, err := r.waitForRunCompletionWithPhaseTracking(ctx, run.ID, suiteID, item.ID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				_ = r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
					suite.Status = domainautobench.SuiteStatusCancelled
					target, findErr := findSuiteItemByID(suite.Items, item.ID)
					if findErr != nil {
						return findErr
					}
					target.PhaseStatus = "cancelled"
					now := time.Now()
					target.EndedAt = &now
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
				now := time.Now()
				target.EndedAt = &now
				return nil
			}); err != nil {
				return err
			}
			// Look up the report by suite_item_id and set ReportID on the suite item
			r.setReportIDOnSuiteItem(ctx, suiteID, item.ID)
			// Write manifest after item completes
			r.writeManifestAsync(ctx, suiteID)
			// Persist suite state to DB after each item
			r.suites.PersistSuite(ctx, suiteID)
			continue
		}

		err = fmt.Errorf("benchmark run ended with state %s", finalRun.State)
		connectionFailure[item.ConnectionID] = true
		if failErr := r.markSuiteItemFailed(suiteID, item.ID, err.Error()); failErr != nil {
			return failErr
		}
		// Look up the report (may have been created and updated to failed) and set ReportID
		r.setReportIDOnSuiteItem(ctx, suiteID, item.ID)
		if suite.ExecutionPolicy.FailurePolicy != domainautobench.FailurePolicyContinueByConnection {
			return err
		}
	}

	// Final suite status update
	if err := r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
		suite.Status = summarizeSuiteStatus(suite.Items)
		now := time.Now()
		suite.EndedAt = &now
		return nil
	}); err != nil {
		return err
	}
	// Final manifest write
	r.writeManifestAsync(ctx, suiteID)
	// Final persist to DB
	r.suites.PersistSuite(ctx, suiteID)
	return nil
}

func (r *AutoBenchSuiteRunner) waitForRunCompletion(ctx context.Context, runID string) (*execution.Run, error) {
	return r.waitForRunCompletionWithPhaseTracking(ctx, runID, "", "")
}

func (r *AutoBenchSuiteRunner) waitForRunCompletionWithPhaseTracking(ctx context.Context, runID string, suiteID string, itemID string) (*execution.Run, error) {
	var lastPhase string
	var phaseStart time.Time

	for {
		run, err := r.benchmark.GetBenchmarkStatus(ctx, runID)
		if err != nil {
			return nil, err
		}
		if run.IsCompleted() {
			// Record final phase timing if tracking
			if lastPhase != "" && suiteID != "" {
				duration := time.Since(phaseStart).Milliseconds()
				r.recordPhaseTiming(suiteID, itemID, lastPhase, duration)
			}
			return run, nil
		}

		currentPhase := run.State.String()
		if currentPhase != lastPhase {
			if lastPhase != "" && suiteID != "" {
				duration := time.Since(phaseStart).Milliseconds()
				r.recordPhaseTiming(suiteID, itemID, lastPhase, duration)
			}
			lastPhase = currentPhase
			phaseStart = time.Now()
			// Update PhaseStatus on item
			if suiteID != "" {
				_ = r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
					target, _ := findSuiteItemByID(suite.Items, itemID)
					if target != nil {
						target.PhaseStatus = currentPhase
					}
					return nil
				})
			}
		}
		if err := r.waitForNextPoll(ctx, r.waitInterval); err != nil {
			return nil, err
		}
	}
}

func (r *AutoBenchSuiteRunner) recordPhaseTiming(suiteID, itemID, phase string, durationMs int64) {
	_ = r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
		target, _ := findSuiteItemByID(suite.Items, itemID)
		if target != nil {
			target.PhaseTimings = append(target.PhaseTimings, domainautobench.PhaseTiming{
				Phase: phase, DurationMs: durationMs,
			})
		}
		return nil
	})
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

func (r *AutoBenchSuiteRunner) selectTemplateForItem(ctx context.Context, dbType string, profile domainautobench.ProfileType) (*domaintemplate.Template, error) {
	templates, err := r.templates.ListTemplates(ctx)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("%w: %s/%s", ErrAutoBenchTemplateNotFound, dbType, profile)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].IsBuiltin != candidates[j].IsBuiltin {
			return candidates[i].IsBuiltin
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0], nil
}

// resolveDefaultRunParams extracts default run parameters from a template's
// Runtime and ToolConfig fields, matching the same logic used by the
// task.go resolveParams function for GUI-initiated benchmarks.
func resolveDefaultRunParams(tmpl *domaintemplate.Template) map[string]interface{} {
	params := make(map[string]interface{})

	// Template-level parameter defaults (e.g., scale, warehouses)
	for key, def := range tmpl.Parameters {
		if def.Default != nil {
			params[key] = def.Default
		}
	}

	// Duration (mapped to "time" for sysbench, "duration" for hammerdb)
	if tmpl.Runtime.DurationSeconds > 0 {
		params["time"] = tmpl.Runtime.DurationSeconds
	}

	// Tool-specific parameters
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
	case domaintemplate.ToolHammerDB:
		if tmpl.ToolConfig.HammerDB.VirtualUsers > 0 {
			params["virtual_users"] = tmpl.ToolConfig.HammerDB.VirtualUsers
		}
		if tmpl.WorkloadFamily == "tproc-c" && tmpl.ToolConfig.HammerDB.Warehouses > 0 {
			params["warehouses"] = tmpl.ToolConfig.HammerDB.Warehouses
		}
		if tmpl.WorkloadFamily == "tproc-h" && tmpl.ToolConfig.HammerDB.ScaleFactor > 0 {
			params["scale"] = tmpl.ToolConfig.HammerDB.ScaleFactor
		}
		if tmpl.Runtime.DurationSeconds > 0 {
			params["duration"] = tmpl.Runtime.DurationSeconds
		}
		if tmpl.Runtime.RampUpSeconds > 0 {
			params["rampup"] = tmpl.Runtime.RampUpSeconds
		}
		if tmpl.Runtime.Iterations > 0 {
			params["iterations"] = tmpl.Runtime.Iterations
		}
	}

	return params
}

func (r *AutoBenchSuiteRunner) markSuiteItemFailed(suiteID, itemID, summary string) error {
	err := r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
		suite.Status = domainautobench.SuiteStatusFailed
		target, findErr := findSuiteItemByID(suite.Items, itemID)
		if findErr != nil {
			return findErr
		}
		target.Status = domainautobench.SuiteItemStatusFailed
		target.ErrorSummary = summary
		now := time.Now()
		target.EndedAt = &now
		return nil
	})
	if err != nil {
		return err
	}
	// Write manifest after item failure
	r.writeManifestAsync(context.Background(), suiteID)
	return nil
}

// setReportIDOnSuiteItem looks up the report by suite_item_id and sets
// the ReportID field on the corresponding suite item. Errors are silently
// ignored since this is a best-effort enrichment.
func (r *AutoBenchSuiteRunner) setReportIDOnSuiteItem(ctx context.Context, suiteID, itemID string) {
	if r.reportUsecase == nil {
		return
	}
	rptID, rptErr := r.reportUsecase.GetReportIDBySuiteItemID(ctx, itemID)
	if rptErr != nil || rptID == "" {
		return
	}
	_ = r.suites.mutateSuite(suiteID, func(suite *domainautobench.Suite) error {
		target, findErr := findSuiteItemByID(suite.Items, itemID)
		if findErr != nil {
			return findErr
		}
		target.ReportID = rptID
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

// writeManifestAsync writes the suite manifest in a goroutine.
// Errors are logged but do not block execution.
func (r *AutoBenchSuiteRunner) writeManifestAsync(ctx context.Context, suiteID string) {
	if r.manifestWriter == nil {
		return
	}
	go func() {
		suite, err := r.suites.getSuite(suiteID)
		if err != nil {
			return
		}
		manifestPath, err := r.manifestWriter.WriteManifest(ctx, &suite)
		if err != nil {
			// Log error but don't fail the execution (per D006/D014)
			return
		}
		// Update the suite with the manifest path
		_ = r.suites.mutateSuite(suiteID, func(s *domainautobench.Suite) error {
			s.SuiteManifestJSONPath = manifestPath
			return nil
		})
	}()
}
