// Package pages provides GUI pages for DB-BenchMind.
// Task Configuration Page implementation.
package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"strings"
)

// TaskConfigurationPage provides the task configuration GUI.
type TaskConfigurationPage struct {
	win         fyne.Window
	connections []string
	// Form fields
	connSelect     *widget.Select
	toolSelect     *widget.Select
	templateSelect *widget.Select
	threadsEntry   *widget.Entry
	durationEntry  *widget.Entry
	tableSizeEntry *widget.Entry
	rateLimitEntry *widget.Entry
}

// NewTaskConfigurationPage creates a new task configuration page.
func NewTaskConfigurationPage(win fyne.Window) fyne.CanvasObject {
	page := &TaskConfigurationPage{
		win:            win,
		threadsEntry:   widget.NewEntry(),
		durationEntry:  widget.NewEntry(),
		tableSizeEntry: widget.NewEntry(),
		rateLimitEntry: widget.NewEntry(),
	}
	page.loadConnections()
	// Create connection selector
	page.connSelect = widget.NewSelect(page.connections, func(s string) {})
	// Create tool selector
	page.toolSelect = widget.NewSelect([]string{"Sysbench", "Swingbench", "HammerDB"}, func(s string) {
		// Update templates based on tool
		page.updateTemplates(s)
	})
	// Create template selector
	page.templateSelect = widget.NewSelect([]string{}, func(s string) {})
	// Set default values
	page.threadsEntry.SetText("4")
	page.durationEntry.SetText("60")
	page.tableSizeEntry.SetText("10000")
	page.rateLimitEntry.SetText("0")
	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Connection", page.connSelect),
			widget.NewFormItem("Tool", page.toolSelect),
			widget.NewFormItem("Template", page.templateSelect),
			widget.NewFormItem("Threads", page.threadsEntry),
			widget.NewFormItem("Duration (seconds)", page.durationEntry),
			widget.NewFormItem("Table Size", page.tableSizeEntry),
			widget.NewFormItem("Rate Limit (0=unlimited)", page.rateLimitEntry),
		},
	}
	// Create buttons
	btnSave := widget.NewButton("Save Task", func() {
		page.onSaveTask()
	})
	btnSaveAndRun := widget.NewButton("Save and Run", func() {
		page.onSaveAndRun()
	})
	btnReset := widget.NewButton("Reset", func() {
		page.onReset()
	})
	toolbar := container.NewHBox(btnSave, btnSaveAndRun, btnReset)
	// Help text
	helpLabel := widget.NewLabel("Configure a benchmark task to run against a database connection.\n" +
		"Select a connection, tool, and template, then configure parameters.")
	content := container.NewVBox(
		widget.NewCard("Task Configuration", "", container.NewPadded(form)),
		widget.NewSeparator(),
		helpLabel,
		widget.NewSeparator(),
		toolbar,
	)
	return content
}

// loadConnections loads available connections.
func (p *TaskConfigurationPage) loadConnections() {
	// For now, use mock connections
	// In production, this would load from database
	p.connections = []string{
		"MySQL Local Test",
		"PostgreSQL Production",
		"Oracle SOE Test",
	}
}

// updateTemplates updates template list based on selected tool.
func (p *TaskConfigurationPage) updateTemplates(tool string) {
	p.templateSelect.Options = []string{}
	switch tool {
	case "Sysbench":
		p.templateSelect.Options = []string{
			"OLTP Read-Write",
			"OLTP Read Only",
			"OLTP Write Only",
		}
	case "Swingbench":
		p.templateSelect.Options = []string{
			"SOE (Sales Order Entry)",
			"SH (Sales History)",
		}
	case "HammerDB":
		p.templateSelect.Options = []string{
			"TPC-C",
			"TPC-B",
		}
	}
}

// onSaveTask saves the task configuration.
func (p *TaskConfigurationPage) onSaveTask() {
	if err := p.validate(); err != nil {
		dialog.ShowError(err, p.win)
		return
	}
	dialog.ShowInformation("Success", "Task configuration saved", p.win)
}

// onSaveAndRun saves and runs the task.
func (p *TaskConfigurationPage) onSaveAndRun() {
	if err := p.validate(); err != nil {
		dialog.ShowError(err, p.win)
		return
	}
	dialog.ShowInformation("Coming Soon", "Task execution will be implemented soon", p.win)
}

// onReset resets the form.
func (p *TaskConfigurationPage) onReset() {
	p.connSelect.ClearSelected()
	p.toolSelect.ClearSelected()
	p.templateSelect.ClearSelected()
	p.threadsEntry.SetText("4")
	p.durationEntry.SetText("60")
	p.tableSizeEntry.SetText("10000")
	p.rateLimitEntry.SetText("0")
}

// validate validates the form input.
func (p *TaskConfigurationPage) validate() error {
	if p.connSelect.Selected == "" {
		return fmt.Errorf("please select a connection")
	}
	if p.toolSelect.Selected == "" {
		return fmt.Errorf("please select a tool")
	}
	if p.templateSelect.Selected == "" {
		return fmt.Errorf("please select a template")
	}
	// Validate threads
	threads, err := strconv.Atoi(strings.TrimSpace(p.threadsEntry.Text))
	if err != nil || threads <= 0 {
		return fmt.Errorf("invalid threads value")
	}
	// Validate duration
	duration, err := strconv.Atoi(strings.TrimSpace(p.durationEntry.Text))
	if err != nil || duration <= 0 {
		return fmt.Errorf("invalid duration value")
	}
	return nil
}
