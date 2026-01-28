// Package pages provides GUI pages for DB-BenchMind.
// Template Management Page implementation.
package pages

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// TemplateManagementPage provides the template management GUI.
type TemplateManagementPage struct {
	list     *widget.List
	templates []templateInfo
	selected int
}

// templateInfo represents display info for a template.
type templateInfo struct {
	ID          string
	Name        string
	Description string
	Tool        string
	IsBuiltin   bool
}

// NewTemplateManagementPage creates a new template management page.
func NewTemplateManagementPage() fyne.CanvasObject {
	page := &TemplateManagementPage{
		selected: -1,
	}

	// Create template list
	page.list = widget.NewList(
		func() int {
			return len(page.templates)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template Name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < widget.ListItemID(len(page.templates)) {
				tmpl := page.templates[id]
				icon := "📄"
				if tmpl.IsBuiltin {
					icon = "📦"
				}
				label.SetText(fmt.Sprintf("%s %s", icon, tmpl.Name))
			}
		},
	)

	page.list.OnSelected = func(id widget.ListItemID) {
		page.selected = int(id)
		tmpl := page.templates[id]
		page.showTemplateDetails(tmpl)
	}

	// Load built-in templates
	page.loadTemplates()

	// Create toolbar
	btnRefresh := widget.NewButton("Refresh", func() { page.loadTemplates() })
	btnDetails := widget.NewButton("View Details", func() {
		if page.selected >= 0 && page.selected < len(page.templates) {
			page.showTemplateDetails(page.templates[page.selected])
		}
	})

	toolbar := container.NewHBox(btnRefresh, btnDetails)

	content := container.NewVBox(
		toolbar,
		widget.NewSeparator(),
		container.NewPadded(page.list),
	)

	return content
}

// loadTemplates loads template information.
func (p *TemplateManagementPage) loadTemplates() {
	// Built-in templates
	p.templates = []templateInfo{
		{
			ID:          "sysbench-oltp-mixed",
			Name:        "Sysbench OLTP Mixed",
			Description: "OLTP read-write mixed workload",
			Tool:        "sysbench",
			IsBuiltin:   true,
		},
		{
			ID:          "sysbench-oltp-read",
			Name:        "Sysbench OLTP Read Only",
			Description: "OLTP read-only workload",
			Tool:        "sysbench",
			IsBuiltin:   true,
		},
		{
			ID:          "swingbench-soe",
			Name:        "Swingbench SOE",
			Description: "Sales Order Entry benchmark for Oracle",
			Tool:        "swingbench",
			IsBuiltin:   true,
		},
		{
			ID:          "hammerdb-tpcc",
			Name:        "HammerDB TPCC",
			Description: "TPC-C benchmark for multiple databases",
			Tool:        "hammerdb",
			IsBuiltin:   true,
		},
	}

	if p.list != nil {
		p.list.Refresh()
	}
}

// showTemplateDetails shows template details.
func (p *TemplateManagementPage) showTemplateDetails(tmpl templateInfo) {
	var sb strings.Builder

	sb.WriteString("# ")
	sb.WriteString(tmpl.Name)
	sb.WriteString("\n\n")

	if tmpl.Description != "" {
		sb.WriteString("**Description:** ")
		sb.WriteString(tmpl.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("**ID:** `")
	sb.WriteString(tmpl.ID)
	sb.WriteString("`\n\n")

	sb.WriteString("**Tool:** `")
	sb.WriteString(tmpl.Tool)
	sb.WriteString("`\n\n")

	if tmpl.IsBuiltin {
		sb.WriteString("**Type:** Built-in Template\n\n")
	} else {
		sb.WriteString("**Type:** Custom Template\n\n")
	}

	sb.WriteString("**Supported Databases:**\n")
	switch tmpl.Tool {
	case "sysbench":
		sb.WriteString("- MySQL\n")
		sb.WriteString("- PostgreSQL\n")
	case "swingbench":
		sb.WriteString("- Oracle\n")
	case "hammerdb":
		sb.WriteString("- MySQL\n")
		sb.WriteString("- Oracle\n")
		sb.WriteString("- SQL Server\n")
		sb.WriteString("- PostgreSQL\n")
	}

	content := widget.NewRichTextFromMarkdown(sb.String())

	dlg := dialog.NewCustomConfirm(
		"Template Details",
		"Close",
		"",
		content,
		func(bool) {},
		nil,
	)
	dlg.Resize(fyne.NewSize(600, 500))
	dlg.Show()
}
