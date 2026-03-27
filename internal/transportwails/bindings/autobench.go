package bindings

import (
	"context"
	"sync"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
)

// AutoBenchBinding exposes AutoBench operations to the frontend.
type AutoBenchBinding struct {
	ctx    context.Context
	suites *usecase.AutoBenchSuiteUseCase
	runner *usecase.AutoBenchSuiteRunner

	mu          sync.RWMutex
	activeSuite string
}

// NewAutoBenchBinding creates a new AutoBenchBinding.
func NewAutoBenchBinding(
	suites *usecase.AutoBenchSuiteUseCase,
	runner *usecase.AutoBenchSuiteRunner,
) *AutoBenchBinding {
	return &AutoBenchBinding{
		suites: suites,
		runner: runner,
	}
}

// SetContext sets the context for the binding.
func (b *AutoBenchBinding) SetContext(ctx context.Context) {
	b.ctx = ctx
}

// AutoBenchCreateSuiteRequest contains the parameters for creating a suite.
type AutoBenchCreateSuiteRequest struct {
	Name            string   `json:"name"`
	ConnectionIDs   []string `json:"connection_ids"`
	ProfileTypes    []string `json:"profile_types"`
	TemplateIDs     []string `json:"template_ids,omitempty"`
	ExecutionMode   string   `json:"execution_mode,omitempty"`
	FailurePolicy   string   `json:"failure_policy,omitempty"`
	CleanupEnabled  *bool    `json:"cleanup_enabled,omitempty"`
}

// AutoBenchCreateSuiteResult contains the result of creating a suite.
type AutoBenchCreateSuiteResult struct {
	SuiteID string `json:"suite_id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// CreateSuite creates a new AutoBench suite.
func (b *AutoBenchBinding) CreateSuite(req AutoBenchCreateSuiteRequest) AutoBenchCreateSuiteResult {
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	profiles := make([]domainautobench.ProfileType, 0, len(req.ProfileTypes))
	for _, p := range req.ProfileTypes {
		profiles = append(profiles, domainautobench.ProfileType(p))
	}

	input := usecase.CreateSuiteInput{
		Name:          req.Name,
		ConnectionIDs: req.ConnectionIDs,
		Profiles:      profiles,
		CleanupEnabled: req.CleanupEnabled,
	}

	suite, err := b.suites.CreateSuite(ctx, input)
	if err != nil {
		return AutoBenchCreateSuiteResult{Error: err.Error()}
	}

	return AutoBenchCreateSuiteResult{
		SuiteID: suite.ID,
		Name:    suite.Name,
		Status:  string(suite.Status),
	}
}

// AutoBenchStartSuiteResult contains the result of starting a suite.
type AutoBenchStartSuiteResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// StartSuite starts execution of an existing suite.
func (b *AutoBenchBinding) StartSuite(suiteID string) AutoBenchStartSuiteResult {
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	b.mu.Lock()
	b.activeSuite = suiteID
	b.mu.Unlock()

	go func() {
		// Run in background to not block the UI
		// If runner is nil, RunSuite will handle the error internally
		if b.runner != nil {
			_ = b.runner.RunSuite(context.Background(), suiteID)
		}

		b.mu.Lock()
		b.activeSuite = ""
		b.mu.Unlock()
	}()

	return AutoBenchStartSuiteResult{Success: true}
}

// AutoBenchSuiteStatusResult contains the status of a suite.
type AutoBenchSuiteStatusResult struct {
	SuiteID        string                      `json:"suite_id"`
	Name           string                      `json:"name"`
	Status         string                      `json:"status"`
	TotalItems     int                         `json:"total_items"`
	PendingItems   int                         `json:"pending_items"`
	RunningItems   int                         `json:"running_items"`
	CompletedItems int                         `json:"completed_items"`
	Items          []AutoBenchItemStatusResult `json:"items"`
	Error          string                      `json:"error,omitempty"`
}

// AutoBenchItemStatusResult contains the status of a suite item.
type AutoBenchItemStatusResult struct {
	ID           string `json:"id"`
	ConnectionID string `json:"connection_id"`
	DatabaseType string `json:"database_type,omitempty"`
	ProfileType  string `json:"profile_type"`
	TemplateID   string `json:"template_id,omitempty"`
	Status       string `json:"status"`
	ReportID     string `json:"report_id,omitempty"`
	StartedAt    string `json:"started_at,omitempty"`
	EndedAt      string `json:"ended_at,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GetSuiteStatus returns the current status of a suite.
func (b *AutoBenchBinding) GetSuiteStatus(suiteID string) AutoBenchSuiteStatusResult {
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	status, err := b.suites.GetSuiteStatus(ctx, suiteID)
	if err != nil {
		return AutoBenchSuiteStatusResult{Error: err.Error()}
	}

	items := make([]AutoBenchItemStatusResult, 0, len(status.Items))
	for _, item := range status.Items {
		result := AutoBenchItemStatusResult{
			ID:           item.ID,
			ConnectionID: item.ConnectionID,
			DatabaseType: item.DatabaseType,
			ProfileType:  string(item.ProfileType),
			TemplateID:   item.TemplateID,
			Status:       string(item.Status),
			ReportID:     item.ReportID,
			ErrorMessage: item.ErrorSummary,
		}
		if item.StartedAt != nil {
			result.StartedAt = item.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if item.EndedAt != nil {
			result.EndedAt = item.EndedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, result)
	}

	return AutoBenchSuiteStatusResult{
		SuiteID:        status.SuiteID,
		Name:           status.Name,
		Status:         string(status.Status),
		TotalItems:     status.TotalItems,
		PendingItems:   status.PendingItems,
		RunningItems:   status.RunningItems,
		CompletedItems: status.CompletedItems,
		Items:          items,
	}
}

// AutoBenchSuiteListResult contains a list of suites.
type AutoBenchSuiteListResult struct {
	Suites []AutoBenchSuiteSummaryResult `json:"suites"`
	Error  string                        `json:"error,omitempty"`
}

// AutoBenchSuiteSummaryResult contains a summary of a suite.
type AutoBenchSuiteSummaryResult struct {
	SuiteID string `json:"suite_id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
}

// ListSuites returns a list of all suites.
func (b *AutoBenchBinding) ListSuites() AutoBenchSuiteListResult {
	// For now, return empty list as we don't have a persistent store
	// This will be enhanced when we add SQLite persistence
	return AutoBenchSuiteListResult{Suites: []AutoBenchSuiteSummaryResult{}}
}

// AutoBenchExecutionPlanResult contains the execution plan for a suite.
type AutoBenchExecutionPlanResult struct {
	SuiteID        string                          `json:"suite_id"`
	Sequential     bool                            `json:"sequential"`
	FailurePolicy  string                          `json:"failure_policy"`
	CleanupEnabled bool                            `json:"cleanup_enabled"`
	Items          []AutoBenchExecutionPlanItem    `json:"items"`
	Error          string                          `json:"error,omitempty"`
}

// AutoBenchExecutionPlanItem represents an item in the execution plan.
type AutoBenchExecutionPlanItem struct {
	Sequence     int    `json:"sequence"`
	SuiteItemID  string `json:"suite_item_id"`
	ConnectionID string `json:"connection_id"`
	ProfileType  string `json:"profile_type"`
}

// GetExecutionPlan returns the execution plan for a suite.
func (b *AutoBenchBinding) GetExecutionPlan(suiteID string) AutoBenchExecutionPlanResult {
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	plan, err := b.suites.BuildExecutionPlan(ctx, suiteID)
	if err != nil {
		return AutoBenchExecutionPlanResult{Error: err.Error()}
	}

	items := make([]AutoBenchExecutionPlanItem, 0, len(plan.Items))
	for _, item := range plan.Items {
		items = append(items, AutoBenchExecutionPlanItem{
			Sequence:     item.Sequence,
			SuiteItemID:  item.SuiteItemID,
			ConnectionID: item.ConnectionID,
			ProfileType:  string(item.ProfileType),
		})
	}

	return AutoBenchExecutionPlanResult{
		SuiteID:        plan.SuiteID,
		Sequential:     plan.Sequential,
		FailurePolicy:  string(plan.FailurePolicy),
		CleanupEnabled: plan.CleanupEnabled,
		Items:          items,
	}
}

// AutoBenchProfileListResult contains the list of supported profiles.
type AutoBenchProfileListResult struct {
	Profiles []string `json:"profiles"`
}

// ListProfiles returns the list of supported profile types.
func (b *AutoBenchBinding) ListProfiles() AutoBenchProfileListResult {
	profiles := b.suites.ListSupportedProfiles()
	result := make([]string, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, string(p))
	}
	return AutoBenchProfileListResult{Profiles: result}
}
