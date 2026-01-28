// Package pages provides GUI pages for DB-BenchMind.
// Result Comparison Page implementation.
package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ResultComparisonPage provides the result comparison GUI.
type ResultComparisonPage struct {
	win            fyne.Window
	comparisonType *widget.Select
	baselineSelect *widget.Select
	compareSelect  *widget.Select
	resultsText    *widget.Entry
}

// NewResultComparisonPage creates a new comparison page.
func NewResultComparisonPage(win fyne.Window) fyne.CanvasObject {
	page := &ResultComparisonPage{
		win: win,
	}
	// Create comparison type selector
	page.comparisonType = widget.NewSelect([]string{
		"Baseline Comparison",
		"Trend Analysis",
		"Multi-Run Comparison",
	}, func(s string) {
		page.onComparisonTypeChange(s)
	})
	// Create run selectors
	page.baselineSelect = widget.NewSelect([]string{}, nil)
	page.compareSelect = widget.NewSelect([]string{}, nil)
	// Load available runs
	page.loadRuns()
	// Create results text area
	page.resultsText = widget.NewMultiLineEntry()
	page.resultsText.SetText("Select runs and comparison type to see results.\n")
	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Comparison Type", page.comparisonType),
			widget.NewFormItem("Baseline Run", page.baselineSelect),
			widget.NewFormItem("Comparison Run", page.compareSelect),
		},
	}
	// Create buttons
	btnCompare := widget.NewButton("Compare", func() {
		page.onCompare()
	})
	btnExport := widget.NewButton("Export Report", func() {
		page.onExportReport()
	})
	btnClear := widget.NewButton("Clear", func() {
		page.resultsText.SetText("")
	})
	toolbar := container.NewHBox(btnCompare, btnExport, btnClear)
	// Help text
	helpLabel := widget.NewLabel("Compare multiple benchmark runs to analyze performance trends and differences.")
	content := container.NewVBox(
		widget.NewCard("Comparison Configuration", "", container.NewPadded(form)),
		widget.NewSeparator(),
		helpLabel,
		widget.NewSeparator(),
		toolbar,
		widget.NewSeparator(),
		widget.NewLabel("Comparison Results:"),
		container.NewPadded(
			container.NewScroll(page.resultsText),
		),
	)
	return content
}

// loadRuns loads available runs for comparison.
func (p *ResultComparisonPage) loadRuns() {
	// Mock data - in production, load from database
	runs := []string{
		"run-001: MySQL OLTP Test (2026-01-28 06:00)",
		"run-002: PostgreSQL TPCC (2026-01-28 07:00)",
		"run-003: MySQL OLTP Test (2026-01-28 08:00)",
		"run-004: Oracle SOE (2026-01-27 14:00)",
	}
	p.baselineSelect.Options = runs
	p.compareSelect.Options = runs
}

// onComparisonTypeChange handles comparison type change.
func (p *ResultComparisonPage) onComparisonTypeChange(comparisonType string) {
	// Update UI based on comparison type
	// Label updates are not directly supported, using dialog instead
	dialog.ShowInformation("Comparison Type", fmt.Sprintf("Selected: %s", comparisonType), p.win)
}

// onCompare performs the comparison.
func (p *ResultComparisonPage) onCompare() {
	if p.baselineSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select baseline run"), p.win)
		return
	}
	if p.comparisonType.Selected != "Multi-Run Comparison" && p.compareSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select comparison run"), p.win)
		return
	}
	// Mock comparison results
	results := fmt.Sprintf("Comparison Results\n")
	results += fmt.Sprintf("==================\n\n")
	results += fmt.Sprintf("Comparison Type: %s\n", p.comparisonType.Selected)
	results += fmt.Sprintf("Baseline: %s\n", p.baselineSelect.Selected)
	results += fmt.Sprintf("Comparison: %s\n\n", p.compareSelect.Selected)
	results += fmt.Sprintf("TPS Comparison:\n")
	results += fmt.Sprintf("  Baseline TPS: 1,250.5\n")
	results += fmt.Sprintf("  Comparison TPS: 1,380.2\n")
	results += fmt.Sprintf("  Difference: +129.7 (+10.4%%)\n\n")
	results += fmt.Sprintf("Latency Comparison:\n")
	results += fmt.Sprintf("  Baseline Avg: 8.5ms\n")
	results += fmt.Sprintf("  Comparison Avg: 7.2ms\n")
	results += fmt.Sprintf("  Difference: -1.3ms (-15.3%%)\n\n")
	results += fmt.Sprintf("Conclusion: Comparison run shows 10.4%% better TPS\n")
	results += fmt.Sprintf("with 15.3%% lower latency.\n")
	p.resultsText.SetText(results)
}

// onExportReport exports the comparison report.
func (p *ResultComparisonPage) onExportReport() {
	if p.resultsText.Text == "" || p.resultsText.Text == "Select runs and comparison type to see results.\n" {
		dialog.ShowError(fmt.Errorf("no comparison results to export"), p.win)
		return
	}
	dialog.ShowInformation("Export", "Report export will be implemented soon", p.win)
}
