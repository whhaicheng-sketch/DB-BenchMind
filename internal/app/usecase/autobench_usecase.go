package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

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
}

type AutoBenchSuiteUseCase struct {
	mu     sync.RWMutex
	suites map[string]domainautobench.Suite
}

func NewAutoBenchSuiteUseCase() *AutoBenchSuiteUseCase {
	return &AutoBenchSuiteUseCase{
		suites: map[string]domainautobench.Suite{},
	}
}

func (uc *AutoBenchSuiteUseCase) ListSupportedProfiles() []domainautobench.ProfileType {
	return append([]domainautobench.ProfileType(nil), domainautobench.DefaultProfileOrder...)
}

func (uc *AutoBenchSuiteUseCase) CreateSuite(ctx context.Context, input CreateSuiteInput) (domainautobench.Suite, error) {
	_ = ctx

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
