package usecase

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
)

var ErrAutoBenchSuiteNotFound = errors.New("autobench suite not found")
var ErrAutoBenchConnectionRequired = errors.New("autobench suite requires at least one connection")

type CreateSuiteInput struct {
	Name           string
	ConnectionIDs  []string
	Profiles       []domainautobench.ProfileType
	CleanupEnabled *bool
}

type AutoBenchExecutionPlan struct {
	SuiteID        string
	Mode           domainautobench.ExecutionMode
	Sequential     bool
	FailurePolicy  domainautobench.FailurePolicy
	CleanupEnabled bool
	Items          []AutoBenchExecutionPlanItem
}

type AutoBenchExecutionPlanItem struct {
	Sequence     int
	SuiteItemID  string
	ConnectionID string
	ProfileType  domainautobench.ProfileType
}

type AutoBenchSuiteStatus struct {
	SuiteID          string
	Name             string
	Status           domainautobench.SuiteStatus
	SelectedProfiles []domainautobench.ProfileType
	TotalItems       int
	PendingItems     int
	RunningItems     int
	CompletedItems   int
	ExecutionPolicy  domainautobench.ExecutionPolicy
	Items            []domainautobench.SuiteItem
	StartedAt        *time.Time
	EndedAt          *time.Time
}

type AutoBenchSuiteUseCase struct {
	mu     sync.RWMutex
	suites map[string]domainautobench.Suite
	repo   SuiteRepository
}

// AutoBenchSuiteUseCaseOption configures the use case.
type AutoBenchSuiteUseCaseOption func(*AutoBenchSuiteUseCase)

// WithSuiteRepository sets the repository for persisting suites.
func WithSuiteRepository(repo SuiteRepository) AutoBenchSuiteUseCaseOption {
	return func(uc *AutoBenchSuiteUseCase) {
		uc.repo = repo
	}
}

func NewAutoBenchSuiteUseCase(opts ...AutoBenchSuiteUseCaseOption) *AutoBenchSuiteUseCase {
	uc := &AutoBenchSuiteUseCase{
		suites: map[string]domainautobench.Suite{},
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

func (uc *AutoBenchSuiteUseCase) ListSupportedProfiles() []domainautobench.ProfileType {
	return append([]domainautobench.ProfileType(nil), domainautobench.DefaultProfileOrder...)
}

func (uc *AutoBenchSuiteUseCase) CreateSuite(ctx context.Context, input CreateSuiteInput) (domainautobench.Suite, error) {
	connectionIDs := normalizeStringSlice(input.ConnectionIDs)
	if len(connectionIDs) == 0 {
		return domainautobench.Suite{}, ErrAutoBenchConnectionRequired
	}
	profiles := uc.normalizeProfiles(input.Profiles)
	suite := domainautobench.NewSuite(strings.TrimSpace(input.Name), connectionIDs, profiles)
	suite.ID = uuid.New().String()
	if input.CleanupEnabled != nil {
		suite.ExecutionPolicy.CleanupEnabled = *input.CleanupEnabled
	}
	suite.ExecutionPolicy.ProfileOrder = append([]domainautobench.ProfileType(nil), profiles...)
	suite.Items = buildSuiteItems(suite.ID, connectionIDs, profiles)

	uc.mu.Lock()
	uc.suites[suite.ID] = cloneSuite(suite)
	uc.mu.Unlock()

	// Persist to repository if available
	if uc.repo != nil {
		if err := uc.repo.Save(ctx, cloneSuite(suite)); err != nil {
			// Log error but don't fail (per D006/D014)
			// Persistence failure should not break main result
		}
	}

	return cloneSuite(suite), nil
}

func (uc *AutoBenchSuiteUseCase) BuildExecutionPlan(ctx context.Context, suiteID string) (AutoBenchExecutionPlan, error) {
	_ = ctx

	suite, err := uc.getSuite(suiteID)
	if err != nil {
		return AutoBenchExecutionPlan{}, err
	}

	plan := AutoBenchExecutionPlan{
		SuiteID:        suite.ID,
		Mode:           suite.ExecutionPolicy.Mode,
		Sequential:     suite.ExecutionPolicy.Mode == domainautobench.ExecutionModeSerial,
		FailurePolicy:  suite.ExecutionPolicy.FailurePolicy,
		CleanupEnabled: suite.ExecutionPolicy.CleanupEnabled,
		Items:          make([]AutoBenchExecutionPlanItem, 0, len(suite.Items)),
	}
	for i, item := range suite.Items {
		plan.Items = append(plan.Items, AutoBenchExecutionPlanItem{
			Sequence:     i + 1,
			SuiteItemID:  item.ID,
			ConnectionID: item.ConnectionID,
			ProfileType:  item.ProfileType,
		})
	}
	return plan, nil
}

func (uc *AutoBenchSuiteUseCase) GetSuiteStatus(ctx context.Context, suiteID string) (AutoBenchSuiteStatus, error) {
	_ = ctx

	suite, err := uc.getSuite(suiteID)
	if err != nil {
		return AutoBenchSuiteStatus{}, err
	}

	status := AutoBenchSuiteStatus{
		SuiteID:          suite.ID,
		Name:             suite.Name,
		Status:           suite.Status,
		SelectedProfiles: append([]domainautobench.ProfileType(nil), suite.SelectedProfiles...),
		TotalItems:       len(suite.Items),
		ExecutionPolicy:  cloneExecutionPolicy(suite.ExecutionPolicy),
		Items:            append([]domainautobench.SuiteItem(nil), suite.Items...),
		StartedAt:        suite.StartedAt,
		EndedAt:          suite.EndedAt,
	}
	for _, item := range suite.Items {
		switch item.Status {
		case domainautobench.SuiteItemStatusPending:
			status.PendingItems++
		case domainautobench.SuiteItemStatusRunning, domainautobench.SuiteItemStatusPreparing, domainautobench.SuiteItemStatusCleaning, domainautobench.SuiteItemStatusValidating:
			status.RunningItems++
		case domainautobench.SuiteItemStatusSuccess, domainautobench.SuiteItemStatusFailed, domainautobench.SuiteItemStatusSkipped:
			status.CompletedItems++
		}
	}
	return status, nil
}

func (uc *AutoBenchSuiteUseCase) BuildSuiteReport(ctx context.Context, suiteID string) (domainautobench.SuiteReport, error) {
	_ = ctx

	suite, err := uc.getSuite(suiteID)
	if err != nil {
		return domainautobench.SuiteReport{}, err
	}

	report := domainautobench.SuiteReport{
		SuiteID:     suite.ID,
		GeneratedAt: time.Now().UTC(),
		Summary: domainautobench.SuiteReportSummary{
			Status:     suite.Status,
			TotalItems: len(suite.Items),
		},
		ConnectionRows: make([]domainautobench.SuiteReportConnectionRow, 0),
		Failures:       make([]domainautobench.SuiteReportFailure, 0),
		ArtifactPaths: domainautobench.SuiteReportArtifactPaths{
			JSON: suite.ReportJSONPath,
			HTML: suite.ReportPath,
		},
	}

	type connectionAccumulator struct {
		connectionID string
		databaseType string
		profiles     []domainautobench.ProfileType
		items        []domainautobench.SuiteItem
	}

	byConnection := map[string]*connectionAccumulator{}
	for _, item := range suite.Items {
		switch item.Status {
		case domainautobench.SuiteItemStatusSuccess:
			report.Summary.SuccessItemCount++
			report.Summary.CompletedItemCount++
		case domainautobench.SuiteItemStatusFailed:
			report.Summary.FailedItemCount++
			report.Summary.CompletedItemCount++
		case domainautobench.SuiteItemStatusSkipped:
			report.Summary.SkippedItemCount++
			report.Summary.CompletedItemCount++
		}

		acc, ok := byConnection[item.ConnectionID]
		if !ok {
			acc = &connectionAccumulator{connectionID: item.ConnectionID}
			byConnection[item.ConnectionID] = acc
		}
		if acc.databaseType == "" {
			acc.databaseType = item.DatabaseType
		}
		if !containsProfileType(acc.profiles, item.ProfileType) {
			acc.profiles = append(acc.profiles, item.ProfileType)
		}
		acc.items = append(acc.items, item)

		if item.Status == domainautobench.SuiteItemStatusFailed || item.Status == domainautobench.SuiteItemStatusSkipped {
			report.Failures = append(report.Failures, domainautobench.SuiteReportFailure{
				SuiteItemID:  item.ID,
				ConnectionID: item.ConnectionID,
				ProfileType:  item.ProfileType,
				LinkedTaskID: item.LinkedTaskID,
				ErrorSummary: item.ErrorSummary,
			})
		}
	}

	connectionIDs := make([]string, 0, len(byConnection))
	for connectionID := range byConnection {
		connectionIDs = append(connectionIDs, connectionID)
	}
	sort.Strings(connectionIDs)

	for _, connectionID := range connectionIDs {
		acc := byConnection[connectionID]
		sort.Slice(acc.profiles, func(i, j int) bool {
			return profileTypeRank(acc.profiles[i]) < profileTypeRank(acc.profiles[j])
		})

		row := domainautobench.SuiteReportConnectionRow{
			ConnectionID: connectionID,
			DatabaseType: acc.databaseType,
			Status:       summarizeSuiteStatus(acc.items),
			ProfileTypes: append([]domainautobench.ProfileType(nil), acc.profiles...),
		}
		for _, item := range acc.items {
			switch item.Status {
			case domainautobench.SuiteItemStatusSuccess:
				row.SuccessItemCount++
				row.CompletedItemCount++
			case domainautobench.SuiteItemStatusFailed:
				row.FailedItemCount++
				row.CompletedItemCount++
			case domainautobench.SuiteItemStatusSkipped:
				row.SkippedItemCount++
				row.CompletedItemCount++
			}
		}
		report.ConnectionRows = append(report.ConnectionRows, row)
	}

	if report.Summary.FailedItemCount > 0 {
		report.Recommendations = append(report.Recommendations, "Review failed suite items before generating the HTML report.")
	}
	if report.Summary.SkippedItemCount > 0 {
		report.Recommendations = append(report.Recommendations, "Investigate skipped items caused by continue_by_connection isolation.")
	}

	return report, nil
}

func (uc *AutoBenchSuiteUseCase) BuildSuiteReportJSON(ctx context.Context, suiteID string) ([]byte, error) {
	report, err := uc.BuildSuiteReport(ctx, suiteID)
	if err != nil {
		return nil, err
	}
	return report.ToJSON()
}

func (uc *AutoBenchSuiteUseCase) BuildSuiteReportHTML(ctx context.Context, suiteID string) ([]byte, error) {
	report, err := uc.BuildSuiteReport(ctx, suiteID)
	if err != nil {
		return nil, err
	}

	viewModel := struct {
		Report               domainautobench.SuiteReport
		GeneratedAtFormatted string
	}{
		Report:               report,
		GeneratedAtFormatted: report.GeneratedAt.Format(time.RFC3339),
	}

	var builder strings.Builder
	if err := autoBenchHTMLReportTemplate.Execute(&builder, viewModel); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

func (uc *AutoBenchSuiteUseCase) getSuite(suiteID string) (domainautobench.Suite, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	suite, ok := uc.suites[strings.TrimSpace(suiteID)]
	if !ok {
		return domainautobench.Suite{}, fmt.Errorf("%w: %s", ErrAutoBenchSuiteNotFound, suiteID)
	}
	return cloneSuite(suite), nil
}

func (uc *AutoBenchSuiteUseCase) normalizeProfiles(profiles []domainautobench.ProfileType) []domainautobench.ProfileType {
	if len(profiles) == 0 {
		return uc.ListSupportedProfiles()
	}

	allowedOrder := domainautobench.DefaultProfileOrder
	selected := map[domainautobench.ProfileType]struct{}{}
	for _, profile := range profiles {
		selected[profile] = struct{}{}
	}

	normalized := make([]domainautobench.ProfileType, 0, len(selected))
	for _, profile := range allowedOrder {
		if _, ok := selected[profile]; ok {
			normalized = append(normalized, profile)
		}
	}
	return normalized
}

func buildSuiteItems(suiteID string, connectionIDs []string, profiles []domainautobench.ProfileType) []domainautobench.SuiteItem {
	items := make([]domainautobench.SuiteItem, 0, len(connectionIDs)*len(profiles))
	for _, connectionID := range connectionIDs {
		for _, profile := range profiles {
			items = append(items, domainautobench.SuiteItem{
				ID:           uuid.New().String(),
				SuiteID:      suiteID,
				ConnectionID: connectionID,
				ProfileType:  profile,
				Status:       domainautobench.SuiteItemStatusPending,
			})
		}
	}
	return items
}

func normalizeStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func cloneSuite(suite domainautobench.Suite) domainautobench.Suite {
	suite.SelectedConnectionIDs = append([]string(nil), suite.SelectedConnectionIDs...)
	suite.SelectedProfiles = append([]domainautobench.ProfileType(nil), suite.SelectedProfiles...)
	suite.ExecutionPolicy = cloneExecutionPolicy(suite.ExecutionPolicy)
	suite.Items = append([]domainautobench.SuiteItem(nil), suite.Items...)
	return suite
}

func cloneExecutionPolicy(policy domainautobench.ExecutionPolicy) domainautobench.ExecutionPolicy {
	policy.ProfileOrder = append([]domainautobench.ProfileType(nil), policy.ProfileOrder...)
	return policy
}

func containsProfileType(values []domainautobench.ProfileType, target domainautobench.ProfileType) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func profileTypeRank(profile domainautobench.ProfileType) int {
	for index, candidate := range domainautobench.DefaultProfileOrder {
		if candidate == profile {
			return index
		}
	}
	return len(domainautobench.DefaultProfileOrder)
}

var autoBenchHTMLReportTemplate = template.Must(template.New("autobench-suite-report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>AutoBench Suite Report - {{ .Report.SuiteID }}</title>
  <style>
    :root { color-scheme: light; }
    body { margin: 0; padding: 32px; font-family: "Helvetica Neue", Arial, sans-serif; background: #f5f7fb; color: #1f2937; }
    main { max-width: 1080px; margin: 0 auto; }
    h1, h2 { margin: 0 0 12px; }
    p { margin: 0 0 8px; }
    section { background: #fff; border: 1px solid #d7deea; border-radius: 16px; padding: 20px 24px; margin-bottom: 18px; box-shadow: 0 10px 30px rgba(15, 23, 42, 0.06); }
    .meta { color: #526075; font-size: 14px; }
    .summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 12px; margin-top: 16px; }
    .summary-card { background: #eef3fb; border-radius: 12px; padding: 14px; }
    .summary-card strong { display: block; font-size: 28px; margin-bottom: 4px; }
    table { width: 100%; border-collapse: collapse; margin-top: 12px; }
    th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid #e5e7eb; vertical-align: top; }
    th { font-size: 13px; text-transform: uppercase; letter-spacing: 0.04em; color: #526075; }
    ul { margin: 12px 0 0; padding-left: 20px; }
    code { background: #eef3fb; padding: 2px 6px; border-radius: 6px; }
  </style>
</head>
<body>
  <main>
    <section>
      <h1>AutoBench Suite Report</h1>
      <p class="meta">Suite ID: <code>{{ .Report.SuiteID }}</code></p>
      <p class="meta">Generated at: {{ .GeneratedAtFormatted }}</p>
      {{ if .Report.ArtifactPaths.JSON }}<p class="meta">JSON Artifact: <code>{{ .Report.ArtifactPaths.JSON }}</code></p>{{ end }}
      {{ if .Report.ArtifactPaths.HTML }}<p class="meta">HTML Artifact: <code>{{ .Report.ArtifactPaths.HTML }}</code></p>{{ end }}
    </section>

    <section>
      <h2>Executive Summary</h2>
      <p>Status: <strong>{{ .Report.Summary.Status }}</strong></p>
      <div class="summary-grid">
        <div class="summary-card"><strong>{{ .Report.Summary.TotalItems }}</strong><span>Total Items</span></div>
        <div class="summary-card"><strong>{{ .Report.Summary.CompletedItemCount }}</strong><span>Completed</span></div>
        <div class="summary-card"><strong>{{ .Report.Summary.SuccessItemCount }}</strong><span>Successful</span></div>
        <div class="summary-card"><strong>{{ .Report.Summary.FailedItemCount }}</strong><span>Failed</span></div>
        <div class="summary-card"><strong>{{ .Report.Summary.SkippedItemCount }}</strong><span>Skipped</span></div>
      </div>
    </section>

    <section>
      <h2>Connection Summary</h2>
      <table>
        <thead>
          <tr>
            <th>Connection</th>
            <th>Database</th>
            <th>Status</th>
            <th>Profiles</th>
            <th>Completed</th>
            <th>Success</th>
            <th>Failed</th>
            <th>Skipped</th>
          </tr>
        </thead>
        <tbody>
          {{ range .Report.ConnectionRows }}
          <tr>
            <td>{{ .ConnectionID }}</td>
            <td>{{ .DatabaseType }}</td>
            <td>{{ .Status }}</td>
            <td>{{ range $index, $profile := .ProfileTypes }}{{ if $index }}, {{ end }}{{ $profile }}{{ end }}</td>
            <td>{{ .CompletedItemCount }}</td>
            <td>{{ .SuccessItemCount }}</td>
            <td>{{ .FailedItemCount }}</td>
            <td>{{ .SkippedItemCount }}</td>
          </tr>
          {{ end }}
        </tbody>
      </table>
    </section>

    <section>
      <h2>Failure Analysis</h2>
      {{ if .Report.Failures }}
      <table>
        <thead>
          <tr>
            <th>Connection</th>
            <th>Profile</th>
            <th>Item</th>
            <th>Task</th>
            <th>Summary</th>
          </tr>
        </thead>
        <tbody>
          {{ range .Report.Failures }}
          <tr>
            <td>{{ .ConnectionID }}</td>
            <td>{{ .ProfileType }}</td>
            <td>{{ .SuiteItemID }}</td>
            <td>{{ .LinkedTaskID }}</td>
            <td>{{ .ErrorSummary }}</td>
          </tr>
          {{ end }}
        </tbody>
      </table>
      {{ else }}
      <p>No failures were recorded for this suite.</p>
      {{ end }}
    </section>

    <section>
      <h2>Recommendations</h2>
      {{ if .Report.Recommendations }}
      <ul>
        {{ range .Report.Recommendations }}
        <li>{{ . }}</li>
        {{ end }}
      </ul>
      {{ else }}
      <p>No recommendations generated.</p>
      {{ end }}
    </section>
  </main>
</body>
</html>
`))
