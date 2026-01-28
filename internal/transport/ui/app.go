// Package ui provides the GUI implementation using Fyne.
// Implements: Transport layer (Clean Architecture)
// - Only handles I/O and user interaction
// - All business logic delegated to use cases
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/transport/ui/pages"
)

// Application represents the Fyne GUI application.
type Application struct {
	app    fyne.App
	connUC *usecase.ConnectionUseCase
}

// NewApplication creates a new Fyne application.
func NewApplication(connUC *usecase.ConnectionUseCase) *Application {
	return &Application{
		app:    app.NewWithID("com.db-benchmind.app"),
		connUC: connUC,
	}
}

// Run starts the application.
func (a *Application) Run() {
	// Create main window
	window := a.app.NewWindow("DB-BenchMind")
	window.Resize(fyne.NewSize(1024, 768))
	window.SetMaster()

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Connections", pages.NewConnectionPage(a.connUC, window)),
		container.NewTabItem("Templates", pages.NewTemplatePage(window)),
		container.NewTabItem("Tasks", pages.NewTaskPage(window)),
		container.NewTabItem("Monitor", pages.NewMonitorPage(window)),
		container.NewTabItem("History", pages.NewHistoryPage(window)),
		container.NewTabItem("Comparison", pages.NewComparisonPage(window)),
		container.NewTabItem("Reports", pages.NewReportPage(window)),
		container.NewTabItem("Settings", pages.NewSettingsPage(window, a.connUC)),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	window.SetContent(tabs)
	window.ShowAndRun()
}
