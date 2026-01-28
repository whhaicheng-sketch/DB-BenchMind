// Package pages provides GUI pages for DB-BenchMind.
// Each page corresponds to a tab in the main window.
package pages

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// ConnectionPage provides the connection management GUI.
// Implements: REQ-CONN-001 (show connection list), REQ-CONN-008 (edit connection)
type ConnectionPage struct {
	connUC *usecase.ConnectionUseCase
	list   *widget.List
	conns  []connection.Connection
}

// NewConnectionPage creates a new connection management page.
func NewConnectionPage(connUC *usecase.ConnectionUseCase) fyne.CanvasObject {
	page := &ConnectionPage{
		connUC: connUC,
	}

	// Create toolbar
	toolbar := container.NewHBox(
		widget.NewButton("Add Connection", page.onAddConnection),
		widget.NewButton("Refresh", page.onRefresh),
	)

	// Create connection list
	page.list = widget.NewList(
		func() int {
			return len(page.conns)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < widget.ListItemID(len(page.conns)) {
				conn := page.conns[id]
				label.SetText(conn.GetName())
			}
		},
	)

	// Load initial connections
	page.loadConnections()

	// Create layout
	return container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewPadded(page.list),
	)
}

// loadConnections loads connections from the use case.
func (p *ConnectionPage) loadConnections() {
	conns, err := p.connUC.ListConnections(context.Background())
	if err != nil {
		showErrorDialog("Failed to load connections", err, p.list)
		return
	}
	p.conns = conns
	p.list.Refresh()
	p.list.UnselectAll()
}

// onAddConnection handles the "Add Connection" button click.
// Implements: REQ-CONN-001
func (p *ConnectionPage) onAddConnection() {
	showConnectionDialog(p.connUC, p.list, nil, p.loadConnections)
}

// onRefresh handles the "Refresh" button click.
func (p *ConnectionPage) onRefresh() {
	p.loadConnections()
}

// =============================================================================
// Connection Dialog
// =============================================================================

// showConnectionDialog shows the connection add/edit dialog.
func showConnectionDialog(connUC *usecase.ConnectionUseCase, parent fyne.CanvasObject, conn connection.Connection, onSuccess func()) {
	d := &connectionDialog{
		connUC:    connUC,
		parent:    parent,
		conn:      conn,
		onSuccess: onSuccess,
	}

	// Create dialog items
	items := d.createDialogItems()

	dialog.ShowForm("Add Connection", "Save", "Cancel", items, func(b bool) {
		if b && d.onSuccess != nil {
			d.onSuccess()
		}
	}, nil)
}

// createDialogItems creates form items for the connection dialog.
func (d *connectionDialog) createDialogItems() []*widget.FormItem {
	// Database type selection
	dbTypeSelect := widget.NewSelect([]string{"MySQL", "PostgreSQL", "Oracle", "SQL Server"}, func(s string) {
		// Update form based on selection
	})

	// Name
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Connection Name")

	// Host
	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("localhost")

	// Port
	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("3306")
	portEntry.Validator = func(s string) error {
		// Port validation will be done by use case
		return nil
	}

	// Database
	dbEntry := widget.NewEntry()
	dbEntry.SetPlaceHolder("Database Name")

	// Username
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Username")

	// Password
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Password")

	// SSL Mode
	sslSelect := widget.NewSelect([]string{"disabled", "preferred", "required"}, nil)

	// Store references
	d.nameEntry = nameEntry
	d.hostEntry = hostEntry
	d.portEntry = portEntry
	d.dbEntry = dbEntry
	d.userEntry = userEntry
	d.passEntry = passEntry
	d.sslSelect = sslSelect
	d.dbTypeSelect = dbTypeSelect

	return []*widget.FormItem{
		{Text: "Database Type", Widget: dbTypeSelect},
		{Text: "Name", Widget: nameEntry},
		{Text: "Host", Widget: hostEntry},
		{Text: "Port", Widget: portEntry},
		{Text: "Database", Widget: dbEntry},
		{Text: "Username", Widget: userEntry},
		{Text: "Password", Widget: passEntry},
		{Text: "SSL Mode", Widget: sslSelect},
	}
}

// connectionDialog represents the connection add/edit dialog.
type connectionDialog struct {
	connUC        *usecase.ConnectionUseCase
	parent       fyne.CanvasObject
	conn         connection.Connection
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

// showErrorDialog shows an error dialog.
func showErrorDialog(title string, err error, parent fyne.CanvasObject) {
	dialog.ShowError(err, nil)
}

// =============================================================================
// Other Page Stubs (to be implemented)
// =============================================================================

// NewTemplatePage creates the template management page.
func NewTemplatePage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Template Management - Coming Soon"))
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
