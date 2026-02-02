// Package pages provides GUI integration for performance reports.
// This file extends the existing comparison page with report generation capabilities.
package pages

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
)

// GeneratePerformanceReport generates and displays a performance report
// using all history records.
// This is an extension method that can be called from the existing ComparisonPage.
func (p *ResultComparisonPage) GenerateComprehensiveReport() {
	if p.comparisonUC == nil {
		dialog.ShowError(fmt.Errorf("comparison use case not available"), p.win)
		return
	}

	ctx := context.Background()

	// Get all record IDs (use all available records for comprehensive analysis)
	refs, err := p.comparisonUC.GetAllRecords(ctx)
	if err != nil {
		slog.Error("Comparison: Failed to get records", "error", err)
		dialog.ShowError(fmt.Errorf("failed to get records: %v", err), p.win)
		return
	}

	if len(refs) < 2 {
		dialog.ShowInformation("Insufficient Data",
			fmt.Sprintf("Need at least 2 records for comparison, found %d.\n\nPlease run more benchmarks first.", len(refs)),
			p.win)
		return
	}

	// Extract record IDs
	recordIDs := make([]string, len(refs))
	for i, ref := range refs {
		recordIDs[i] = ref.ID
	}

	// Show progress
	progress := dialog.NewInformation("Generating Report",
		"Analyzing benchmark data...\n\nPlease wait.", p.win)
	progress.Show()

	// Generate report in background
	go func() {
		// Determine group by field
		groupBy := comparison.GroupByThreads
		if p.groupBySelect != nil {
			selected := p.groupBySelect.Selected
			switch selected {
			case "Threads":
				groupBy = comparison.GroupByThreads
			case "Database":
				groupBy = comparison.GroupByDatabaseType
			case "Template":
				groupBy = comparison.GroupByTemplate
			}
		}

		// Use default similarity config
		similarityConfig := comparison.DefaultSimilarityConfig()

		// Generate comprehensive report
		report, err := p.comparisonUC.GenerateComprehensiveReport(
			ctx, recordIDs, groupBy, similarityConfig)
		if err != nil {
			slog.Error("Comparison: Failed to generate report", "error", err)
			progress.Hide()
			dialog.ShowError(fmt.Errorf("failed to generate report: %v", err), p.win)
			return
		}

		// Hide progress
		progress.Hide()

		// Display results
		p.displayComprehensiveReport(report)

		slog.Info("Comparison: Comprehensive report generated",
			"report_id", report.ReportID,
			"groups", len(report.ConfigGroups))
	}()
}

// displayComprehensiveReport formats and displays the comprehensive report.
func (p *ResultComparisonPage) displayComprehensiveReport(report *comparison.ComparisonReport) {
	// Generate Markdown format (primary format)
	markdown := report.FormatMarkdown()

	// Update results text
	if p.resultsText != nil {
		p.resultsText.SetText(markdown)
	}

	// Show summary dialog
	summary := fmt.Sprintf(
		"✅ Comprehensive Report Generated!\n\n"+
			"Report ID: %s\n"+
			"Config Groups: %d\n"+
			"Grouped by: %s\n\n"+
			"Sanity Checks: ",
		report.ReportID,
		len(report.ConfigGroups),
		report.GroupBy)

	if report.SanityChecks != nil {
		if report.SanityChecks.AllPassed {
			summary += "✅ ALL PASSED\n"
		} else {
			passed := 0
			for _, check := range report.SanityChecks.Checks {
				if check.Passed {
					passed++
				}
			}
			summary += fmt.Sprintf("⚠️  %d/%d passed\n",
				passed, len(report.SanityChecks.Checks))
		}
	}

	summary += "\nFull report is displayed below.\n\n"+
		"You can export this report to Markdown or TXT format."

	dialog.ShowInformation("Report Generated", summary, p.win)
}

// ExportComprehensiveReport exports the current comprehensive report.
func (p *ResultComparisonPage) ExportComprehensiveReport(report *comparison.ComparisonReport) {
	if report == nil {
		dialog.ShowError(fmt.Errorf("no report to export"), p.win)
		return
	}

	// Ask for format
	formatSelect := widget.NewRadioGroup([]string{"Markdown", "TXT"}, func(selected string) {})
	formatSelect.SetSelected("Markdown")

	content := container.NewVBox(
		widget.NewLabel("Export Comprehensive Report"),
		widget.NewSeparator(),
		widget.NewLabel("Select export format:"),
		formatSelect,
		widget.NewSeparator(),
	)

	dialog.ShowCustomConfirm("Export Report", "Export", "Cancel", content, func(export bool) {
		if !export {
			return
		}

		// Determine format
		format := "markdown"
		if formatSelect.Selected == "TXT" {
			format = "txt"
		}

		// Generate filename
		timestamp := time.Now().Format("20060102_150405")
		var ext string
		if format == "markdown" {
			ext = ".md"
		} else {
			ext = ".txt"
		}
		filename := fmt.Sprintf("comparison_report_%s%s", timestamp, ext)
		filepath := fmt.Sprintf("./exports/%s", filename)

		// Export via usecase
		ctx := context.Background()
		err := p.comparisonUC.ExportReport(ctx, report, format, filepath)
		if err != nil {
			slog.Error("Comparison: Failed to export report", "error", err)
			dialog.ShowError(fmt.Errorf("export failed: %v", err), p.win)
			return
		}

		dialog.ShowInformation("Export Successful",
			fmt.Sprintf("Report exported to:\n%s\n\nFormat: %s", filepath, format),
			p.win)

		slog.Info("Comparison: Report exported", "filepath", filepath, "format", format)
	}, p.win)
}

// AddComprehensiveReportButton adds a button to generate comprehensive reports.
// This can be called from the comparison page initialization to add the new feature.
func (p *ResultComparisonPage) AddComprehensiveReportButton(toolbar *fyne.Container) {
	btnComprehensive := widget.NewButton("📊 Full Report", func() {
		p.GenerateComprehensiveReport()
	})
	btnComprehensive.Importance = widget.MediumImportance

	// Add to toolbar
	if objs := toolbar.Objects; len(objs) > 0 {
		toolbar.Objects = append(objs, btnComprehensive)
		toolbar.Refresh()
	}
}

// GeneratePerformanceReport generates and displays a performance report.
// Automatically uses ALL history records from the database.
// This is the main reporting feature for performance analysis.
func (p *ResultComparisonPage) GenerateSimplifiedReport() {
	if p.comparisonUC == nil {
		dialog.ShowError(fmt.Errorf("comparison use case not available"), p.win)
		return
	}

	ctx := context.Background()

	// Get all record IDs (use all available records for simplified analysis)
	refs, err := p.comparisonUC.GetAllRecords(ctx)
	if err != nil {
		slog.Error("Comparison: Failed to get records", "error", err)
		dialog.ShowError(fmt.Errorf("failed to get records: %v", err), p.win)
		return
	}

	if len(refs) < 2 {
		dialog.ShowInformation("Insufficient Data",
			fmt.Sprintf("Need at least 2 records for comparison, found %d.\n\nPlease run more benchmarks first.", len(refs)),
			p.win)
		return
	}

	// Extract record IDs
	recordIDs := make([]string, len(refs))
	for i, ref := range refs {
		recordIDs[i] = ref.ID
	}

	// Show progress
	progress := dialog.NewInformation("Generating Simplified Report",
		"Analyzing benchmark data...\n\nPlease wait.", p.win)
	progress.Show()

	// Determine group by field
	groupBy := comparison.GroupByThreads
	if p.groupBySelect != nil {
		selected := p.groupBySelect.Selected
		switch selected {
		case "Threads":
			groupBy = comparison.GroupByThreads
		case "Database":
			groupBy = comparison.GroupByDatabaseType
		case "Template":
			groupBy = comparison.GroupByTemplate
		}
	}

	// Generate simplified report (synchronous for simplicity)
	report, err := p.comparisonUC.GenerateSimplifiedReport(ctx, recordIDs, groupBy)
	if err != nil {
		slog.Error("Comparison: Failed to generate simplified report", "error", err)
		progress.Hide()
		dialog.ShowError(fmt.Errorf("failed to generate simplified report: %v", err), p.win)
		return
	}

	// Hide progress
	progress.Hide()

	// Display results
	p.displaySimplifiedReport(report)

	slog.Info("Comparison: Simplified report generated",
		"report_id", report.ReportID,
		"groups", len(report.ConfigGroups))
}

// displaySimplifiedReport formats and displays the performance report.
func (p *ResultComparisonPage) displaySimplifiedReport(report *comparison.SimplifiedReport) {
	// Generate Markdown format (primary format)
	markdown := report.FormatMarkdown()

	// Update results text
	if p.resultsText != nil {
		p.resultsText.SetText(markdown)
	}

	// Show summary dialog
	passed := 0
	for _, check := range report.SanityChecks {
		if check.Passed {
			passed++
		}
	}

	summary := fmt.Sprintf(
		"✅ Simplified Report Generated!\n\n"+
			"Report ID: %s\n"+
			"Config Groups: %d\n"+
			"Grouped by: %s\n"+
			"Records: %d\n\n"+
			"Sanity Checks: %d/%d passed\n\n"+
			"Full report is displayed below.\n\n"+
			"You can export this report to Markdown or TXT format.",
		report.ReportID,
		len(report.ConfigGroups),
		report.GroupBy,
		report.SelectedRecords,
		passed, len(report.SanityChecks))

	dialog.ShowInformation("Report Generated", summary, p.win)
}

// ExportPerformanceReport exports the current performance report.
func (p *ResultComparisonPage) ExportSimplifiedReport(report *comparison.SimplifiedReport) {
	if report == nil {
		dialog.ShowError(fmt.Errorf("no report to export"), p.win)
		return
	}

	// Ask for format
	formatSelect := widget.NewRadioGroup([]string{"Markdown", "TXT"}, func(selected string) {})
	formatSelect.SetSelected("Markdown")

	content := container.NewVBox(
		widget.NewLabel("Export Simplified Report"),
		widget.NewSeparator(),
		widget.NewLabel("Select export format:"),
		formatSelect,
		widget.NewSeparator(),
	)

	dialog.ShowCustomConfirm("Export Report", "Export", "Cancel", content, func(export bool) {
		if !export {
			return
		}

		// Determine format
		format := "markdown"
		if formatSelect.Selected == "TXT" {
			format = "txt"
		}

		// Generate filename
		timestamp := time.Now().Format("20060102_150405")
		var ext string
		if format == "markdown" {
			ext = ".md"
		} else {
			ext = ".txt"
		}
		filename := fmt.Sprintf("simplified_report_%s%s", timestamp, ext)
		filepath := fmt.Sprintf("./exports/%s", filename)

		// Export via usecase
		ctx := context.Background()
		err := p.comparisonUC.ExportSimplifiedReport(ctx, report, format, filepath)
		if err != nil {
			slog.Error("Comparison: Failed to export simplified report", "error", err)
			dialog.ShowError(fmt.Errorf("export failed: %v", err), p.win)
			return
		}

		dialog.ShowInformation("Export Successful",
			fmt.Sprintf("Report exported to:\n%s\n\nFormat: %s", filepath, format),
			p.win)

		slog.Info("Comparison: Simplified report exported", "filepath", filepath, "format", format)
	}, p.win)
}

// AddSimplifiedReportButton adds a button to generate simplified reports.
// This can be called from the comparison page initialization to add the new feature.
func (p *ResultComparisonPage) AddSimplifiedReportButton(toolbar *fyne.Container) {
	btnSimplified := widget.NewButton("📋 Simple Report", func() {
		p.GenerateSimplifiedReport()
	})
	btnSimplified.Importance = widget.MediumImportance

	// Add to toolbar
	if objs := toolbar.Objects; len(objs) > 0 {
		toolbar.Objects = append(objs, btnSimplified)
		toolbar.Refresh()
	}
}
