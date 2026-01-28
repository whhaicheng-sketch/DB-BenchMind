// Package pages provides GUI pages for DB-BenchMind.
// Report Export Page implementation.
package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ReportExportPage provides the report export GUI.
type ReportExportPage struct {
	win             fyne.Window
	runSelect       *widget.Select
	formatSelect    *widget.Select
	includeSections *multiselectWidget
	outputPath      *widget.Entry
}

// multiselectWidget provides multi-selection capability.
type multiselectWidget struct {
	checkBox *widget.CheckGroup
	options  []string
}

// NewReportExportPage creates a new report page.
func NewReportExportPage(win fyne.Window) fyne.CanvasObject {
	page := &ReportExportPage{
		win: win,
	}
	// Create run selector
	page.runSelect = widget.NewSelect([]string{}, nil)
	page.loadRuns()
	// Create format selector
	page.formatSelect = widget.NewSelect([]string{
		"Markdown (.md)",
		"HTML (.html)",
		"JSON (.json)",
		"PDF (.pdf)",
	}, nil)
	page.formatSelect.SetSelected("Markdown (.md)")
	// Create section multi-select
	sectionChecks := widget.NewCheckGroup([]string{
		"Summary",
		"Metrics",
		"Charts",
		"Raw Data",
		"Configuration",
	}, nil)
	sectionChecks.SetSelected([]string{"Summary", "Metrics", "Charts"})
	page.includeSections = &multiselectWidget{
		checkBox: sectionChecks,
		options:  []string{"Summary", "Metrics", "Charts", "Raw Data", "Configuration"},
	}
	// Create output path entry
	page.outputPath = widget.NewEntry()
	page.outputPath.SetText(fmt.Sprintf("./reports/report-%s.md", getCurrentTimestamp()))
	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Run to Export", page.runSelect),
			widget.NewFormItem("Format", page.formatSelect),
			widget.NewFormItem("Output Path", page.outputPath),
		},
	}
	// Section selection
	sectionLabel := widget.NewLabel("Include Sections:")
	sectionContainer := container.NewVBox(
		sectionLabel,
		page.includeSections.checkBox,
	)
	// Create buttons
	btnGenerate := widget.NewButton("Generate Report", func() {
		page.onGenerateReport()
	})
	btnPreview := widget.NewButton("Preview", func() {
		page.onPreview()
	})
	btnBrowse := widget.NewButton("Browse...", func() {
		page.onBrowsePath()
	})
	toolbar := container.NewHBox(btnGenerate, btnPreview, btnBrowse)
	// Help text
	helpLabel := widget.NewLabel("Generate detailed benchmark reports in various formats.\nSelect a run, choose format, and specify which sections to include.")
	content := container.NewVBox(
		widget.NewCard("Report Configuration", "", container.NewPadded(form)),
		widget.NewSeparator(),
		container.NewPadded(sectionContainer),
		widget.NewSeparator(),
		helpLabel,
		widget.NewSeparator(),
		toolbar,
	)
	return content
}

// loadRuns loads available runs for export.
func (p *ReportExportPage) loadRuns() {
	// Mock data - in production, load from database
	runs := []string{
		"run-001: MySQL OLTP Test (2026-01-28 06:00)",
		"run-002: PostgreSQL TPCC (2026-01-28 07:00)",
		"run-003: Oracle SOE (2026-01-28 08:00)",
	}
	p.runSelect.Options = runs
}

// onGenerateReport generates the report.
func (p *ReportExportPage) onGenerateReport() {
	if p.runSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select a run to export"), p.win)
		return
	}
	if p.outputPath.Text == "" {
		dialog.ShowError(fmt.Errorf("please specify output path"), p.win)
		return
	}
	// Check selected sections
	sections := p.includeSections.checkBox.Selected
	if len(sections) == 0 {
		dialog.ShowError(fmt.Errorf("please select at least one section"), p.win)
		return
	}
	// Mock report generation
	message := fmt.Sprintf("Report generated successfully!\n\n")
	message += fmt.Sprintf("Run: %s\n", p.runSelect.Selected)
	message += fmt.Sprintf("Format: %s\n", p.formatSelect.Selected)
	message += fmt.Sprintf("Output: %s\n", p.outputPath.Text)
	message += fmt.Sprintf("Sections: %v\n", sections)
	dialog.ShowInformation("Report Generated", message, p.win)
}

// onPreview previews the report.
func (p *ReportExportPage) onPreview() {
	if p.runSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select a run to preview"), p.win)
		return
	}
	// Mock preview content
	preview := fmt.Sprintf("# Benchmark Report\n\n")
	preview += fmt.Sprintf("## Run: %s\n\n", p.runSelect.Selected)
	preview += fmt.Sprintf("### Summary\n")
	preview += fmt.Sprintf("- Tool: Sysbench\n")
	preview += fmt.Sprintf("- Duration: 60s\n")
	preview += fmt.Sprintf("- Threads: 4\n")
	preview += fmt.Sprintf("- TPS: 1,250.5\n")
	preview += fmt.Sprintf("- Avg Latency: 8.5ms\n")
	preview += fmt.Sprintf("- Errors: 0\n\n")
	preview += fmt.Sprintf("*(Preview shows partial content)*\n")
	dialog.ShowCustomConfirm(
		"Report Preview",
		"Close",
		"",
		widget.NewRichTextFromMarkdown(preview),
		func(bool) {},
		p.win,
	)
}

// onBrowsePath opens file browser dialog.
func (p *ReportExportPage) onBrowsePath() {
	dialog.ShowInformation("Browse", "File browser will be implemented soon", p.win)
}

// getCurrentTimestamp returns current timestamp in format YYYYMMDD-HHMMSS.
func getCurrentTimestamp() string {
	return "20260128-120000" // Simplified for now
}
