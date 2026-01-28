// Package pages provides GUI pages for DB-BenchMind.
package pages

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// ConnectionPage provides the connection management GUI.
type ConnectionPage struct {
	connUC   *usecase.ConnectionUseCase
	win      fyne.Window
	list     *widget.List
	conns    []connection.Connection
	selected int
}

// NewConnectionPage creates a new connection management page.
func NewConnectionPage(connUC *usecase.ConnectionUseCase, win fyne.Window) fyne.CanvasObject {
	page := &ConnectionPage{
		connUC:   connUC,
		win:     win,
		selected: -1,
	}

	// Create connection list
	page.list = widget.NewList(
		func() int {
			return len(page.conns)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Connection Name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < widget.ListItemID(len(page.conns)) {
				conn := page.conns[id]
				label.SetText(fmt.Sprintf("%s (%s)", conn.GetName(), conn.GetType()))
			}
		},
	)

	// Handle selection
	page.list.OnSelected = func(id widget.ListItemID) {
		page.selected = int(id)
	}

	// Handle unselection - Fyne 2.x doesn't have OnUnselected, use OnSelected with -1
	// page.list.OnUnselected = func() {}

	// Create toolbar
	btnAdd := widget.NewButton("Add Connection", func() { page.onAddConnection() })
	btnDelete := widget.NewButton("Delete", func() { page.onDeleteConnection() })
	btnTest := widget.NewButton("Test", func() { page.onTestConnection() })
	btnRefresh := widget.NewButton("Refresh", func() { page.loadConnections() })

	toolbar := container.NewHBox(btnAdd, btnDelete, btnTest, btnRefresh)

	content := container.NewVBox(
		toolbar,
		widget.NewSeparator(),
		container.NewPadded(page.list),
	)

	// Load initial connections
	page.loadConnections()

	return content
}

// loadConnections loads connections from the use case.
func (p *ConnectionPage) loadConnections() {
	conns, err := p.connUC.ListConnections(context.Background())
	if err != nil {
		dialog.ShowError(err, nil)
		return
	}
	p.conns = conns
	p.list.Refresh()
}

// onAddConnection handles the "Add Connection" button click.
func (p *ConnectionPage) onAddConnection() {
	showConnectionDialog(p.connUC, p.win, p.loadConnections)
}

// onDeleteConnection handles the "Delete" button click.
func (p *ConnectionPage) onDeleteConnection() {
	if p.selected < 0 || p.selected >= len(p.conns) {
		return
	}

	conn := p.conns[p.selected]
	dialog.ShowConfirm(
		"Delete Connection",
		fmt.Sprintf("Delete connection '%s'?", conn.GetName()),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			err := p.connUC.DeleteConnection(context.Background(), conn.GetID())
			if err != nil {
				dialog.ShowError(err, p.win)
				return
			}
			dialog.ShowInformation("Success", "Connection deleted", p.win)
			p.loadConnections()
		},
		p.win,
	)
}

// onTestConnection handles the "Test Connection" button click.
func (p *ConnectionPage) onTestConnection() {
	if p.selected < 0 || p.selected >= len(p.conns) {
		return
	}

	conn := p.conns[p.selected]
	win := p.win // Capture for goroutine

	// Test in background
	go func() {
		result, err := p.connUC.TestConnection(context.Background(), conn.GetID())
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		if result.Success {
			msg := fmt.Sprintf("Success! Latency: %dms\nVersion: %s",
				result.LatencyMs, result.DatabaseVersion)
			dialog.ShowInformation("Connection Test", msg, win)
		} else {
			dialog.ShowError(fmt.Errorf("failed: %s", result.Error), win)
		}
	}()
}

// =============================================================================
// Connection Dialog
// =============================================================================

// showConnectionDialog shows the connection add/edit dialog.
func showConnectionDialog(connUC *usecase.ConnectionUseCase, win fyne.Window, onSuccess func()) {
	d := &connectionDialog{
		connUC:    connUC,
		onSuccess: onSuccess,
	}

	// Create form fields
	d.nameEntry = widget.NewEntry()
	d.nameEntry.SetPlaceHolder("My Connection")

	d.hostEntry = widget.NewEntry()
	d.hostEntry.SetText("localhost")

	d.portEntry = widget.NewEntry()
	d.portEntry.SetText("3306")

	d.dbEntry = widget.NewEntry()

	d.userEntry = widget.NewEntry()

	d.passEntry = widget.NewPasswordEntry()

	d.sslSelect = widget.NewSelect([]string{"disabled", "preferred", "required"}, nil)

	d.dbTypeSelect = widget.NewSelect([]string{"MySQL", "PostgreSQL", "Oracle", "SQL Server"}, func(s string) {
		switch s {
		case "MySQL":
			d.portEntry.SetText("3306")
		case "PostgreSQL":
			d.portEntry.SetText("5432")
		case "Oracle":
			d.portEntry.SetText("1521")
		case "SQL Server":
			d.portEntry.SetText("1433")
		}
	})

	formItems := []*widget.FormItem{
		widget.NewFormItem("Database Type", d.dbTypeSelect),
		widget.NewFormItem("Name", d.nameEntry),
		widget.NewFormItem("Host", d.hostEntry),
		widget.NewFormItem("Port", d.portEntry),
		widget.NewFormItem("Database", d.dbEntry),
		widget.NewFormItem("Username", d.userEntry),
		widget.NewFormItem("Password", d.passEntry),
		widget.NewFormItem("SSL", d.sslSelect),
	}

	// Show form dialog
	dialog.ShowForm("Add Connection", "Save", "Cancel", formItems, func(save bool) {
		if save {
			d.onSave(win)
		}
	}, win)
}

// onSave handles the save button click.
func (d *connectionDialog) onSave(win fyne.Window) {
	ctx := context.Background()
	now := time.Now()
	id := fmt.Sprintf("conn-%d", now.UnixNano())

	dbType := d.dbTypeSelect.Selected
	name := strings.TrimSpace(d.nameEntry.Text)
	host := strings.TrimSpace(d.hostEntry.Text)
	port, _ := strconv.Atoi(d.portEntry.Text)
	database := strings.TrimSpace(d.dbEntry.Text)
	username := strings.TrimSpace(d.userEntry.Text)
	password := d.passEntry.Text
	sslMode := d.sslSelect.Selected

	if name == "" {
		dialog.ShowError(fmt.Errorf("name required"), win)
		return
	}

	// Create connection based on type
	var conn connection.Connection

	switch dbType {
	case "MySQL":
		conn = &connection.MySQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:        id,
				Name:      name,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Host:     host,
			Port:     port,
			Database: database,
			Username: username,
			Password: password,
			SSLMode:  sslMode,
		}
	case "PostgreSQL":
		conn = &connection.PostgreSQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:        id,
				Name:      name,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Host:     host,
			Port:     port,
			Database: database,
			Username: username,
			Password: password,
			SSLMode:  sslMode,
		}
	case "Oracle":
		conn = &connection.OracleConnection{
			BaseConnection: connection.BaseConnection{
				ID:        id,
				Name:      name,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Host:     host,
			Port:     port,
			SID:      database,
			Username: username,
			Password: password,
		}
	case "SQL Server":
		conn = &connection.SQLServerConnection{
			BaseConnection: connection.BaseConnection{
				ID:        id,
				Name:      name,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Host:     host,
			Port:     port,
			Database: database,
			Username: username,
			Password: password,
		}
	default:
		dialog.ShowError(fmt.Errorf("unsupported type: %s", dbType), win)
		return
	}

	// Validate
	if err := conn.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("validation: %w", err), win)
		return
	}

	// Save
	if err := d.connUC.CreateConnection(ctx, conn); err != nil {
		dialog.ShowError(fmt.Errorf("save: %w", err), win)
		return
	}

	dialog.ShowInformation("Success", "Connection saved", win)

	if d.onSuccess != nil {
		d.onSuccess()
	}
}

// connectionDialog represents the connection dialog.
type connectionDialog struct {
	connUC       *usecase.ConnectionUseCase
	onSuccess    func()
	nameEntry    *widget.Entry
	hostEntry    *widget.Entry
	portEntry    *widget.Entry
	dbEntry      *widget.Entry
	userEntry    *widget.Entry
	passEntry    *widget.Entry
	sslSelect    *widget.Select
	dbTypeSelect *widget.Select
}

// =============================================================================
// Other Pages
// =============================================================================

// NewTemplatePage creates the template management page.
func NewTemplatePage() fyne.CanvasObject {
	return NewTemplateManagementPage()
}

// NewTaskPage creates the task configuration page.
func NewTaskPage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Task Configuration - Coming Soon"))
}

// NewMonitorPage creates the run monitoring page.
func NewMonitorPage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Run Monitoring - Coming Soon"))
}

// NewHistoryPage creates the history page.
func NewHistoryPage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("History Records - Coming Soon"))
}

// NewComparisonPage creates the result comparison page.
func NewComparisonPage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Result Comparison - Coming Soon"))
}

// NewReportPage creates the report export page.
func NewReportPage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Report Export - Coming Soon"))
}

// NewSettingsPage creates the settings page.
func NewSettingsPage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Settings - Coming Soon"))
}
