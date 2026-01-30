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
	app         fyne.App
	connUC      *usecase.ConnectionUseCase
	benchmarkUC *usecase.BenchmarkUseCase
	templateUC  *usecase.TemplateUseCase
	historyUC   *usecase.HistoryUseCase
	exportUC    *usecase.ExportUseCase
	comparisonUC *usecase.ComparisonUseCase
}

// NewApplication creates a new Fyne application.
func NewApplication(connUC *usecase.ConnectionUseCase, benchmarkUC *usecase.BenchmarkUseCase, templateUC *usecase.TemplateUseCase, historyUC *usecase.HistoryUseCase, exportUC *usecase.ExportUseCase, comparisonUC *usecase.ComparisonUseCase) *Application {
	return &Application{
		app:         app.NewWithID("com.db-benchmind.app"),
		connUC:      connUC,
		benchmarkUC: benchmarkUC,
		templateUC:  templateUC,
		historyUC:   historyUC,
		exportUC:    exportUC,
		comparisonUC: comparisonUC,
	}
}

// Run starts the application.
func (a *Application) Run() {
	// Create main window
	window := a.app.NewWindow("DB-BenchMind")
	window.Resize(fyne.NewSize(1024, 900)) // Increased from 768 to 900 for more log display space
	window.SetMaster()

	// Set close interceptor when main window closes
	window.SetCloseIntercept(func() {
		a.app.Quit()
	})

	// Create history page and save reference
	historyPage, historyPageContent := pages.NewHistoryRecordPage(window, a.historyUC, a.exportUC)

	// Create comparison page and save reference
	comparisonPage, comparisonPageContent := pages.NewResultComparisonPage(window, a.comparisonUC)

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Connections", pages.NewConnectionPage(a.connUC, window)),
		container.NewTabItem("Templates", pages.NewTemplatePage(window)),
		container.NewTabItem("Tasks & Monitor", pages.NewTaskMonitorPageWithUC(window, a.connUC, a.benchmarkUC, a.templateUC, a.historyUC)),
		container.NewTabItem("History", historyPageContent),
		container.NewTabItem("Comparison", comparisonPageContent),
		container.NewTabItem("Reports", pages.NewReportPage(window)),
		container.NewTabItem("Settings", pages.NewSettingsPage(window, a.connUC)),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	// Add tab change listener to auto-refresh pages when selected
	tabs.OnSelected = func(tab *container.TabItem) {
		// Auto-refresh History when selected
		if tab.Text == "History" {
			historyPage.Refresh()
		}
		// Auto-refresh Comparison when selected
		if tab.Text == "Comparison" {
			comparisonPage.Refresh()
		}
	}

	window.SetContent(tabs)

	// Run main window (blocks until window is closed)
	window.ShowAndRun()
}
