// Package pages provides GUI pages for DB-BenchMind.
// Task Configuration Page implementation (Simplified for beginners).
package pages

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// TaskConfigurationPage provides the simplified task configuration GUI.
type TaskConfigurationPage struct {
	win            fyne.Window
	connSelect     *widget.Select
	toolSelect     *widget.Select
	templateSelect  *widget.Select
	durationEntry  *widget.Entry
	rateLimitEntry *widget.Entry
}

// NewTaskConfigurationPage creates a new task configuration page.
func NewTaskConfigurationPage(win fyne.Window) fyne.CanvasObject {
	page := &TaskConfigurationPage{
		win: win,
		// Will initialize connUC later or skip it for mock data
	}

	// Create connection selector
	page.connSelect = widget.NewSelect([]string{}, nil)

	// Create tool selector
	page.toolSelect = widget.NewSelect([]string{"Sysbench", "Swingbench", "HammerDB"}, func(s string) {
		// Update templates based on tool
		page.updateTemplates(s)
	})

	// Create template selector
	page.templateSelect = widget.NewSelect([]string{}, nil)

	// Create simplified form fields
	page.durationEntry = widget.NewEntry()
	page.durationEntry.SetText("60")

	page.rateLimitEntry = widget.NewEntry()
	page.rateLimitEntry.SetText("0")

	// Load connections after all widgets are created
	page.loadConnections()

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Connection", page.connSelect),
			widget.NewFormItem("Tool", page.toolSelect),
			widget.NewFormItem("Template", page.templateSelect),
			widget.NewFormItem("Duration (seconds)", page.durationEntry),
			widget.NewFormItem("Rate Limit (0=unlimited)", page.rateLimitEntry),
		},
	}

	// Create buttons
	btnRun := widget.NewButton("Run Task", func() {
		page.onRunTask()
	})

	btnRefresh := widget.NewButton("Refresh Connections", func() {
		page.loadConnections()
	})

	toolbar := container.NewHBox(btnRun, btnRefresh)

	// Help text
	helpLabel := widget.NewLabel("Configure and run a benchmark task.\nSelect a connection, tool, and template, then set duration.")

	content := container.NewVBox(
		widget.NewCard("Task Configuration", "", container.NewPadded(form)),
		widget.NewSeparator(),
		helpLabel,
		widget.NewSeparator(),
		toolbar,
	)

	return content
}

// loadConnections loads connections from the database (synchronized with Connections page).
func (p *TaskConfigurationPage) loadConnections() {
	// For now, use mock connections that match what's in Connections page
	// In production, this would use: conns, err := p.connUC.ListConnections(context.Background())
	connections := []string{
		"MySQL Local Test",
		"PostgreSQL Production",
		"Oracle SOE Test",
	}

	p.connSelect.Options = connections
	if len(connections) > 0 {
		p.connSelect.SetSelected(connections[0])
		// Auto-select templates when connection is selected
		p.updateTemplates(p.toolSelect.Selected)
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
		}
	case "Swingbench":
		p.templateSelect.Options = []string{
			"SOE (Sales Order Entry)",
		}
	case "HammerDB":
		p.templateSelect.Options = []string{
			"TPC-C",
		}
	}

	if len(p.templateSelect.Options) > 0 {
		p.templateSelect.SetSelected(p.templateSelect.Options[0])
	}
}

// onRunTask runs the benchmark task.
func (p *TaskConfigurationPage) onRunTask() {
	// Validate
	if p.connSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select a connection"), p.win)
		return
	}

	if p.toolSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select a tool"), p.win)
		return
	}

	if p.templateSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select a template"), p.win)
		return
	}

	duration, err := strconv.Atoi(strings.TrimSpace(p.durationEntry.Text))
	if err != nil || duration <= 0 {
		dialog.ShowError(fmt.Errorf("invalid duration value"), p.win)
		return
	}

	// Show task summary
	var sb strings.Builder
	sb.WriteString("Task Configuration Summary\n\n")
	sb.WriteString(fmt.Sprintf("Connection: %s\n", p.connSelect.Selected))
	sb.WriteString(fmt.Sprintf("Tool: %s\n", p.toolSelect.Selected))
	sb.WriteString(fmt.Sprintf("Template: %s\n", p.templateSelect.Selected))
	sb.WriteString(fmt.Sprintf("Duration: %d seconds\n", duration))
	sb.WriteString(fmt.Sprintf("Rate Limit: %s\n", p.rateLimitEntry.Text))

	sb.WriteString("\nTask is ready to run!\n")
	sb.WriteString("(Full task execution will be implemented soon)")

	dialog.ShowInformation("Task Ready", sb.String(), p.win)
}
