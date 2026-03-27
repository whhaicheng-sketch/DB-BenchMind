// Package usecase defines repository interfaces for the application.
package usecase

import (
	"context"

	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
)

// SuiteRepository defines the interface for persisting suites.
type SuiteRepository interface {
	// Save saves a suite to the database.
	Save(ctx context.Context, suite domainautobench.Suite) error

	// FindByID finds a suite by its ID.
	FindByID(ctx context.Context, id string) (domainautobench.Suite, error)

	// FindAll finds all suites in the database.
	FindAll(ctx context.Context) ([]domainautobench.Suite, error)

	// UpdateStatus updates the status of a suite.
	UpdateStatus(ctx context.Context, id string, status domainautobench.SuiteStatus, manifestPath string) error

	// Delete deletes a suite by its ID.
	Delete(ctx context.Context, id string) error
}
