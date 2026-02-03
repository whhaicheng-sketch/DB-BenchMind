// Package pages provides GUI pages for DB-BenchMind.
package pages

import (
	"context"
	"fmt"
	"log/slog"
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
	connUC *usecase.ConnectionUseCase
	win    fyne.Window
	conns  []connection.Connection
	// Group containers
	groupContainers map[string]*fyne.Container // DB type -> container
	listContainer   *fyne.Container
}

// NewConnectionPage creates a new connection management page.
func NewConnectionPage(connUC *usecase.ConnectionUseCase, win fyne.Window) fyne.CanvasObject {
	page := &ConnectionPage{
		connUC:          connUC,
		win:             win,
		groupContainers: make(map[string]*fyne.Container),
		listContainer:   container.NewVBox(),
	}

	// Create toolbar with only Add button
	btnAdd := widget.NewButton("➕ Add", func() {
		slog.Info("Connections: Add button clicked")
		page.onAddConnection()
	})
	toolbar := container.NewVBox(
		container.NewHBox(btnAdd),
	)

	// Load connections to populate the list
	page.loadConnections()

	// Create top area with toolbar
	topArea := container.NewVBox(
		toolbar,
		widget.NewSeparator(),
	)

	content := container.NewBorder(
		topArea,                                 // top - toolbar
		nil,                                     // bottom
		nil,                                     // left
		nil,                                     // right
		container.NewScroll(page.listContainer), // center - fills available space
	)

	return content
}

// loadConnections loads connections from the use case and groups them by database type.
func (p *ConnectionPage) loadConnections() {
	slog.Info("Connections: Loading connections")
	conns, err := p.connUC.ListConnections(context.Background())
	if err != nil {
		slog.Error("Connections: Failed to load", "error", err)
		dialog.ShowError(err, nil)
		return
	}
	p.conns = conns

	// Group connections by database type
	groups := make(map[string][]connection.Connection)
	for _, conn := range conns {
		dbType := string(conn.GetType())
		// Normalize to capitalized name for display and grouping
		displayType := dbType
		switch dbType {
		case "mysql":
			displayType = "MySQL"
		case "postgresql":
			displayType = "PostgreSQL"
		case "oracle":
			displayType = "Oracle"
		case "sqlserver":
			displayType = "SQL Server"
		}
		slog.Info("Connections: Found connection", "name", conn.GetName(), "db_type", dbType, "display_type", displayType)
		groups[displayType] = append(groups[displayType], conn)
	}

	slog.Info("Connections: Groups created", "total_groups", len(groups), "group_keys", fmt.Sprintf("%v", groups))

	// Clear list container
	p.listContainer.Objects = nil
	p.groupContainers = make(map[string]*fyne.Container)

	// Define order of database types
	dbOrder := []string{"MySQL", "PostgreSQL", "Oracle", "SQL Server"}

	// Create collapsible groups
	for _, dbType := range dbOrder {
		conns := groups[dbType]
		if len(conns) == 0 {
			continue
		}

		slog.Info("Connections: Creating group", "db_type", dbType, "count", len(conns))
		p.createConnectionGroup(dbType, conns)
	}

	p.listContainer.Refresh()
	slog.Info("Connections: List refreshed", "total_connections", len(conns))
}

// createConnectionGroup creates a collapsible group for a database type.
func (p *ConnectionPage) createConnectionGroup(dbType string, conns []connection.Connection) {
	slog.Info("Connections: createConnectionGroup START", "db_type", dbType, "count", len(conns))

	// Group header with expand/collapse button
	headerText := fmt.Sprintf("📊 %s (%d)", dbType, len(conns))
	headerBtn := widget.NewButton(headerText, nil)

	// Container for this group's connections
	groupContainer := container.NewVBox()
	p.groupContainers[dbType] = groupContainer

	// Initially expanded
	isExpanded := true

	// Toggle expand/collapse
	headerBtn.OnTapped = func() {
		isExpanded = !isExpanded
		slog.Info("Connections: Group header clicked", "db_type", dbType, "expanded", isExpanded)
		if isExpanded {
			groupContainer.Show()
		} else {
			groupContainer.Hide()
		}
	}

	// Add connections to this group
	for _, conn := range conns {
		// Get connection info for display
		connName := conn.GetName()
		var host, portStr, username string

		switch c := conn.(type) {
		case *connection.MySQLConnection:
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		case *connection.PostgreSQLConnection:
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		case *connection.OracleConnection:
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		case *connection.SQLServerConnection:
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		}

		// Connection info label
		infoText := fmt.Sprintf("    🔗 %s  |  %s@%s:%s", connName, username, host, portStr)
		infoLabel := widget.NewLabel(infoText)

		// Buttons for this connection: Test, Edit, Delete
		btnTest := widget.NewButton("🔌 Test", func() {
			slog.Info("Connections: Test button clicked", "connection", connName)
			p.onTestConnection(conn)
		})
		btnEdit := widget.NewButton("✏️ Edit", func() {
			slog.Info("Connections: Edit button clicked", "connection", connName)
			p.onEditConnection(conn)
		})
		btnDelete := widget.NewButton("🗑️ Delete", func() {
			slog.Info("Connections: Delete button clicked", "connection", connName)
			p.onDeleteConnection(conn)
		})
		buttonBox := container.NewHBox(btnTest, btnEdit, btnDelete)

		// Use Border layout to align info left, buttons right
		connRow := container.NewBorder(nil, nil, infoLabel, buttonBox)
		groupContainer.Add(connRow)
	}

	// Add header and group to main list
	p.listContainer.Add(headerBtn)
	p.listContainer.Add(groupContainer)
}

// normalizeDBType converts raw DB type to capitalized display name.
func normalizeDBType(dbType string) string {
	switch dbType {
	case "mysql":
		return "MySQL"
	case "postgresql":
		return "PostgreSQL"
	case "oracle":
		return "Oracle"
	case "sqlserver":
		return "SQL Server"
	}
	return dbType
}

// onAddConnection handles the "Add Connection" button click.
func (p *ConnectionPage) onAddConnection() {
	slog.Info("Connections: Add button clicked")
	showConnectionDialog(p.connUC, p.win, nil, p.loadConnections)
}

// onEditConnection handles the "Edit" button click.
func (p *ConnectionPage) onEditConnection(conn connection.Connection) {
	showConnectionDialog(p.connUC, p.win, conn, p.loadConnections)
}

// onDeleteConnection handles the "Delete" button click.
func (p *ConnectionPage) onDeleteConnection(conn connection.Connection) {
	dialog.ShowConfirm(
		"Delete Connection",
		fmt.Sprintf("Delete connection '%s'?", conn.GetName()),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			slog.Info("Connections: Deleting connection", "name", conn.GetName())
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
func (p *ConnectionPage) onTestConnection(conn connection.Connection) {
	win := p.win // Capture for goroutine

	// Show progress dialog
	progressDlg := dialog.NewInformation("Testing Connection",
		fmt.Sprintf("Testing connection '%s'...\n\nPlease wait.", conn.GetName()), win)
	progressDlg.Show()

	// Test in background
	go func() {
		slog.Info("Connections: Testing connection", "name", conn.GetName())
		result, err := p.connUC.TestConnection(context.Background(), conn.GetID())

		// Hide progress dialog
		progressDlg.Hide()

		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if result.Success {
			slog.Info("Connections: Test successful", "name", conn.GetName(), "latency_ms", result.LatencyMs)
			msg := fmt.Sprintf("Success! Latency: %dms\nVersion: %s",
				result.LatencyMs, result.DatabaseVersion)
			dialog.ShowInformation("Connection Test", msg, win)
		} else {
			slog.Warn("Connections: Test failed", "name", conn.GetName(), "error", result.Error)
			dialog.ShowError(fmt.Errorf("failed: %s", result.Error), win)
		}
	}()
}

// =============================================================================
// Connection Dialog
// =============================================================================
// showConnectionDialog shows the connection add/edit dialog.
func showConnectionDialog(connUC *usecase.ConnectionUseCase, win fyne.Window, conn connection.Connection, onSuccess func()) {
	d := &connectionDialog{
		connUC:     connUC,
		onSuccess:  onSuccess,
		conn:       conn,
		isEditMode: conn != nil,
		win:        win,
	}
	// Create form fields
	d.nameEntry = widget.NewEntry()
	d.hostEntry = widget.NewEntry()
	d.hostEntry.SetText("localhost")
	d.portEntry = widget.NewEntry()
	d.portEntry.SetText("3306")
	d.dbEntry = widget.NewEntry()
	d.userEntry = widget.NewEntry()
	d.passEntry = widget.NewPasswordEntry()
	d.trustServerCertCheck = widget.NewCheck("Trust Server Certificate", func(checked bool) {
		// Handle trust server certificate change
	})
	d.trustServerCertCheck.SetChecked(true) // Default to true for SQL Server (recommended)
	d.trustServerCertCheck.Hide()          // Initially hidden, only show for SQL Server
	d.dbTypeSelect = widget.NewSelect([]string{"MySQL", "PostgreSQL", "Oracle", "SQL Server"}, func(s string) {
		// Set default port based on database type
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

		// Show/hide Trust Server Certificate based on database type
		if s == "SQL Server" {
			d.trustServerCertCheck.Show()
			d.trustServerCertCheck.SetChecked(true) // Enable by default for SQL Server
		} else {
			d.trustServerCertCheck.Hide()
		}

		// Load default configuration for this database type (only in add mode)
		if !d.isEditMode {
			defaultConfig, err := connection.LoadDefaultConnectionConfig()
			if err == nil && defaultConfig != nil {
				switch s {
				case "MySQL":
					if defaultConfig.MySQL != nil {
						if defaultConfig.MySQL.Host != "" {
							d.hostEntry.SetText(defaultConfig.MySQL.Host)
						}
						if defaultConfig.MySQL.Port > 0 {
							d.portEntry.SetText(fmt.Sprintf("%d", defaultConfig.MySQL.Port))
						}
						d.dbEntry.SetText(defaultConfig.MySQL.Database)
						d.userEntry.SetText(defaultConfig.MySQL.Username)
						slog.Info("Connections: Loaded default config for MySQL")
					}
				case "PostgreSQL":
					if defaultConfig.PostgreSQL != nil {
						if defaultConfig.PostgreSQL.Host != "" {
							d.hostEntry.SetText(defaultConfig.PostgreSQL.Host)
						}
						if defaultConfig.PostgreSQL.Port > 0 {
							d.portEntry.SetText(fmt.Sprintf("%d", defaultConfig.PostgreSQL.Port))
						}
						d.dbEntry.SetText(defaultConfig.PostgreSQL.Database)
						d.userEntry.SetText(defaultConfig.PostgreSQL.Username)
						slog.Info("Connections: Loaded default config for PostgreSQL")
					}
				case "Oracle":
					if defaultConfig.Oracle != nil {
						if defaultConfig.Oracle.Host != "" {
							d.hostEntry.SetText(defaultConfig.Oracle.Host)
						}
						if defaultConfig.Oracle.Port > 0 {
							d.portEntry.SetText(fmt.Sprintf("%d", defaultConfig.Oracle.Port))
						}
						// Oracle uses ServiceName or SID
						if defaultConfig.Oracle.ServiceName != "" {
							d.dbEntry.SetText(defaultConfig.Oracle.ServiceName)
						} else if defaultConfig.Oracle.SID != "" {
							d.dbEntry.SetText(defaultConfig.Oracle.SID)
						}
						d.userEntry.SetText(defaultConfig.Oracle.Username)
						slog.Info("Connections: Loaded default config for Oracle")
					}
				case "SQL Server":
					if defaultConfig.SQLServer != nil {
						if defaultConfig.SQLServer.Host != "" {
							d.hostEntry.SetText(defaultConfig.SQLServer.Host)
						}
						if defaultConfig.SQLServer.Port > 0 {
							d.portEntry.SetText(fmt.Sprintf("%d", defaultConfig.SQLServer.Port))
						}
						d.dbEntry.SetText(defaultConfig.SQLServer.Database)
						d.userEntry.SetText(defaultConfig.SQLServer.Username)
						d.trustServerCertCheck.SetChecked(defaultConfig.SQLServer.TrustServerCertificate)
						slog.Info("Connections: Loaded default config for SQL Server")
					}
				}
			}
		}
	})

	// If editing, populate with existing values
	if d.isEditMode && d.conn != nil {
		// For edit mode, load connection with password from keyring
		connWithPassword, err := connUC.GetConnectionByID(context.Background(), d.conn.GetID())
		if err != nil {
			slog.Warn("Connections: Failed to load password from keyring", "error", err)
			// Continue without password, user will need to re-enter
		} else {
			d.conn = connWithPassword // Replace with connection that has password
			slog.Info("Connections: Loaded connection with password from keyring",
				"id", d.conn.GetID(),
				"has_password", d.conn != nil)
		}

		d.nameEntry.SetText(d.conn.GetName())
		// Convert DatabaseType to display name
		displayType := ""
		switch d.conn.GetType() {
		case connection.DatabaseTypeMySQL:
			displayType = "MySQL"
		case connection.DatabaseTypePostgreSQL:
			displayType = "PostgreSQL"
		case connection.DatabaseTypeOracle:
			displayType = "Oracle"
		case connection.DatabaseTypeSQLServer:
			displayType = "SQL Server"
		}
		d.dbTypeSelect.SetSelected(displayType)

		// Set other fields based on connection type
		switch c := d.conn.(type) {
		case *connection.MySQLConnection:
			d.hostEntry.SetText(c.Host)
			d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			d.dbEntry.SetText(c.Database)
			d.userEntry.SetText(c.Username)
			d.passEntry.SetText(c.Password)
		case *connection.PostgreSQLConnection:
			d.hostEntry.SetText(c.Host)
			d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			d.dbEntry.SetText(c.Database)
			d.userEntry.SetText(c.Username)
			d.passEntry.SetText(c.Password)
		case *connection.OracleConnection:
			d.hostEntry.SetText(c.Host)
			d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			d.dbEntry.SetText(c.SID)
			d.userEntry.SetText(c.Username)
			d.passEntry.SetText(c.Password)
		case *connection.SQLServerConnection:
			d.hostEntry.SetText(c.Host)
			d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			d.dbEntry.SetText(c.Database)
			d.userEntry.SetText(c.Username)
			d.passEntry.SetText(c.Password)
			d.trustServerCertCheck.SetChecked(c.TrustServerCertificate)
			slog.Info("Connections: Loaded SQL Server connection in edit mode",
				"host", c.Host,
				"port", c.Port,
				"username", c.Username,
				"password_length", len(c.Password),
				"trust_server_cert", c.TrustServerCertificate)
		}
	} else {
		// New connection - load default config if available
		defaultConfig, err := connection.LoadDefaultConnectionConfig()
		if err == nil && defaultConfig != nil {
			// Load defaults for MySQL (default selection)
			if defaultConfig.MySQL != nil {
				d.hostEntry.SetText(defaultConfig.MySQL.Host)
				if defaultConfig.MySQL.Port > 0 {
					d.portEntry.SetText(fmt.Sprintf("%d", defaultConfig.MySQL.Port))
				}
				d.dbEntry.SetText(defaultConfig.MySQL.Database)
				d.userEntry.SetText(defaultConfig.MySQL.Username)
			}
			slog.Info("Connections: Loaded default config", "db_type", "MySQL")
		}
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Database Type", d.dbTypeSelect),
		widget.NewFormItem("Name", d.nameEntry),
		widget.NewFormItem("Host", d.hostEntry),
		widget.NewFormItem("Port", d.portEntry),
		widget.NewFormItem("Database", d.dbEntry),
		widget.NewFormItem("Username", d.userEntry),
		widget.NewFormItem("Password", d.passEntry),
		widget.NewFormItem("Encryption", d.trustServerCertCheck),
	}

	// Initially hide Trust Server Certificate (only show for SQL Server)
	d.trustServerCertCheck.Hide()

	// Determine dialog title
	title := "Add Connection"
	if d.isEditMode {
		title = "Edit Connection"
	}

	// Create buttons first (before dialog)
	btnTest := widget.NewButton("Test", func() {
		slog.Info("Connections: Dialog Test button clicked", "name", d.nameEntry.Text, "type", d.dbTypeSelect.Selected)
		d.onTestInDialog()
		// Note: dialog remains open after Test
	})
	btnSave := widget.NewButton("Save", func() {
		slog.Info("Connections: Dialog Save button clicked", "name", d.nameEntry.Text, "type", d.dbTypeSelect.Selected, "mode", map[bool]string{true: "edit", false: "add"}[d.isEditMode])
		d.onSave(win)
		d.dialog.Hide() // Close dialog after save
	})
	btnSave.Importance = widget.HighImportance
	btnCancel := widget.NewButton("Cancel", func() {
		slog.Info("Connections: Dialog Cancel button clicked", "name", d.nameEntry.Text, "type", d.dbTypeSelect.Selected)
		// Will be set to close dialog after dialog is created
	})

	buttonContainer := container.NewHBox(btnTest, btnSave, btnCancel)

	// Create form
	form := widget.NewForm(formItems...)

	// Create dialog content with buttons at bottom
	content := container.NewVBox(form, widget.NewSeparator(), buttonContainer)

	// Create custom dialog without buttons
	dlg := dialog.NewCustomWithoutButtons(title, content, win)
	dlg.Resize(fyne.NewSize(500, 600))
	d.dialog = dlg // Store dialog reference

	// Update Cancel button to close dialog
	btnCancel.OnTapped = func() {
		dlg.Hide()
	}

	dlg.Show()
}

// onSave handles the save button click.
func (d *connectionDialog) onSave(win fyne.Window) {
	ctx := context.Background()
	now := time.Now()

	dbType := d.dbTypeSelect.Selected
	name := strings.TrimSpace(d.nameEntry.Text)
	host := strings.TrimSpace(d.hostEntry.Text)
	port, _ := strconv.Atoi(d.portEntry.Text)
	database := strings.TrimSpace(d.dbEntry.Text)
	username := strings.TrimSpace(d.userEntry.Text)
	password := d.passEntry.Text
	trustServerCert := d.trustServerCertCheck.Checked

	// In edit mode, if password field is empty, reload from keyring
	if d.isEditMode && d.conn != nil && password == "" {
		slog.Info("Connections: Loading password from keyring for save",
			"conn_id", d.conn.GetID())
		connWithPassword, err := d.connUC.GetConnectionByID(ctx, d.conn.GetID())
		if err != nil {
			slog.Error("Connections: Failed to load password from keyring", "error", err)
			dialog.ShowError(fmt.Errorf("failed to load password: %w", err), win)
			return
		}
		switch c := connWithPassword.(type) {
		case *connection.MySQLConnection:
			password = c.Password
		case *connection.PostgreSQLConnection:
			password = c.Password
		case *connection.OracleConnection:
			password = c.Password
		case *connection.SQLServerConnection:
			password = c.Password
		}
		slog.Info("Connections: Loaded password from keyring for save",
			"password_length", len(password))
	}

	mode := "add"
	if d.isEditMode {
		mode = "edit"
	}

	slog.Info("Connections: Saving connection",
		"mode", mode,
		"name", name,
		"type", dbType,
		"host", host,
		"port", port,
		"database", database,
		"username", username,
		"trust_server_cert", trustServerCert)

	// In edit mode, use the existing connection's ID to avoid duplicate name error
	// In add mode, generate a new ID
	var id string
	var createdAt time.Time
	if d.isEditMode && d.conn != nil {
		id = d.conn.GetID()
		// Get original creation time from the connection
		// We need to extract BaseConnection, but each type embeds it
		// For simplicity, we'll use current time in edit mode too
		createdAt = now
	} else {
		id = fmt.Sprintf("conn-%d", now.UnixNano())
		createdAt = now
	}

	if name == "" {
		slog.Warn("Connections: Save validation failed", "error", "name required")
		dialog.ShowError(fmt.Errorf("name required"), win)
		return
	}

	// Delete old connection in edit mode before creating new one
	if d.isEditMode && d.conn != nil {
		if err := d.connUC.DeleteConnection(ctx, d.conn.GetID()); err != nil {
			dialog.ShowError(fmt.Errorf("failed to update connection: %w", err), win)
			return
		}
	}

	// Create connection based on type
	var conn connection.Connection
	switch dbType {
	case "MySQL":
		conn = &connection.MySQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:        id,
				Name:      name,
				CreatedAt: createdAt,
				UpdatedAt: time.Now(),
			},
			Host:     host,
			Port:     port,
			Database: database,
			Username: username,
			Password: password,
			SSLMode:  "disable", // Default value
		}
	case "PostgreSQL":
		conn = &connection.PostgreSQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:        id,
				Name:      name,
				CreatedAt: createdAt,
				UpdatedAt: time.Now(),
			},
			Host:     host,
			Port:     port,
			Database: database,
			Username: username,
			Password: password,
			SSLMode:  "disable", // Default value
		}
	case "Oracle":
		conn = &connection.OracleConnection{
			BaseConnection: connection.BaseConnection{
				ID:        id,
				Name:      name,
				CreatedAt: createdAt,
				UpdatedAt: time.Now(),
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
				CreatedAt: createdAt,
				UpdatedAt: time.Now(),
			},
			Host:                   host,
			Port:                   port,
			Database:               database,
			Username:               username,
			Password:               password,
			TrustServerCertificate: trustServerCert,
		}
	default:
		dialog.ShowError(fmt.Errorf("unsupported type: %s", dbType), win)
		return
	}
	// Validate
	if err := conn.Validate(); err != nil {
		slog.Warn("Connections: Save validation failed", "name", name, "error", err)
		dialog.ShowError(fmt.Errorf("validation: %w", err), win)
		return
	}
	// Save
	if err := d.connUC.CreateConnection(ctx, conn); err != nil {
		slog.Error("Connections: Failed to save", "name", name, "error", err)
		dialog.ShowError(fmt.Errorf("save: %w", err), win)
		return
	}

	slog.Info("Connections: Connection saved successfully",
		"id", id,
		"name", name,
		"type", dbType,
		"mode", mode)

	// Save as default configuration (automatic, no prompt)
	if err := connection.SaveConnectionAsDefault(conn); err != nil {
		slog.Warn("Connections: Failed to save default config", "error", err)
		// Don't fail the save operation if default config save fails
	} else {
		slog.Info("Connections: Saved as default config", "db_type", dbType, "connection", name)
	}

	dialog.ShowInformation("Success", "Connection saved", win)
	// Dialog will be closed by button callback

	if d.onSuccess != nil {
		d.onSuccess()
	}
}

// onTestInDialog tests the connection from the dialog.
func (d *connectionDialog) onTestInDialog() {
	ctx := context.Background()
	name := strings.TrimSpace(d.nameEntry.Text)

	if name == "" {
		slog.Warn("Connections: Dialog test validation failed", "error", "name required")
		dialog.ShowError(fmt.Errorf("name required"), d.win)
		return
	}

	// Show progress dialog
	progressDlg := dialog.NewInformation("Testing Connection",
		fmt.Sprintf("Testing connection '%s'...\n\nPlease wait.", name), d.win)
	progressDlg.Show()

	// Test connection in background
	go func() {
		var result *connection.TestResult
		var err error

		if d.isEditMode && d.conn != nil {
			// EDIT MODE: Use the saved connection (same logic as list test)
			slog.Info("Connections: Testing in EDIT mode",
				"name", name,
				"conn_id", d.conn.GetID())

			result, err = d.connUC.TestConnection(ctx, d.conn.GetID())
		} else {
			// ADD MODE: Create temporary connection from form fields
			slog.Info("Connections: Testing in ADD mode", "name", name)

			host := strings.TrimSpace(d.hostEntry.Text)
			port, _ := strconv.Atoi(d.portEntry.Text)
			database := strings.TrimSpace(d.dbEntry.Text)
			username := strings.TrimSpace(d.userEntry.Text)
			password := d.passEntry.Text
			trustServerCert := d.trustServerCertCheck.Checked
			dbType := d.dbTypeSelect.Selected

			// Create temporary connection
			var conn connection.Connection
			now := time.Now()
			switch dbType {
			case "MySQL":
				conn = &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:        "temp-test",
						Name:      name,
						CreatedAt: now,
						UpdatedAt: now,
					},
					Host:     host,
					Port:     port,
					Database: database,
					Username: username,
					Password: password,
					SSLMode:  "disable", // Default, will be removed later
				}
			case "PostgreSQL":
				conn = &connection.PostgreSQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:        "temp-test",
						Name:      name,
						CreatedAt: now,
						UpdatedAt: now,
					},
					Host:     host,
					Port:     port,
					Database: database,
					Username: username,
					Password: password,
					SSLMode:  "disable", // Default, will be removed later
				}
			case "Oracle":
				conn = &connection.OracleConnection{
					BaseConnection: connection.BaseConnection{
						ID:        "temp-test",
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
						ID:        "temp-test",
						Name:      name,
						CreatedAt: now,
						UpdatedAt: now,
					},
					Host:                   host,
					Port:                   port,
					Database:               database,
					Username:               username,
					Password:               password,
					TrustServerCertificate: trustServerCert,
				}
			default:
				progressDlg.Hide()
				dialog.ShowError(fmt.Errorf("unsupported type: %s", dbType), d.win)
				return
			}

			// Validate
			if err := conn.Validate(); err != nil {
				progressDlg.Hide()
				slog.Warn("Connections: Dialog test validation failed", "name", name, "error", err)
				dialog.ShowError(fmt.Errorf("validation: %w", err), d.win)
				return
			}

			// Test
			result, err = conn.Test(ctx)
		}

		// Hide progress dialog
		progressDlg.Hide()

		if err != nil {
			slog.Error("Connections: Dialog test error", "name", name, "error", err)
			dialog.ShowError(err, d.win)
			return
		}

		if result.Success {
			slog.Info("Connections: Dialog test successful",
				"name", name,
				"latency_ms", result.LatencyMs,
				"version", result.DatabaseVersion)
			msg := fmt.Sprintf("Success! Latency: %dms\nVersion: %s",
				result.LatencyMs, result.DatabaseVersion)
			dialog.ShowInformation("Connection Test", msg, d.win)
		} else {
			slog.Warn("Connections: Dialog test failed",
				"name", name,
				"error", result.Error)
			dialog.ShowError(fmt.Errorf("failed: %s", result.Error), d.win)
		}
	}()
}

// connectionDialog represents the connection dialog.
type connectionDialog struct {
	connUC               *usecase.ConnectionUseCase
	onSuccess            func()
	conn                 connection.Connection // For editing
	isEditMode           bool
	win                  fyne.Window
	dialog               *dialog.CustomDialog // Reference to dialog for closing
	nameEntry            *widget.Entry
	hostEntry            *widget.Entry
	portEntry            *widget.Entry
	dbEntry              *widget.Entry
	userEntry            *widget.Entry
	passEntry            *widget.Entry
	trustServerCertCheck *widget.Check // For SQL Server
	dbTypeSelect         *widget.Select
}

// =============================================================================
// Other Pages - Wrapper Functions
// =============================================================================
// NewTemplatePage creates the template management page.
func NewTemplatePage(win fyne.Window) fyne.CanvasObject {
	return NewTemplateManagementPage(win)
}

// NewTaskPage creates the task configuration and monitor page (combined).
func NewTaskPage(win fyne.Window) fyne.CanvasObject {
	return NewTaskMonitorPage(win)
}

// NewMonitorPage creates the run monitoring page (deprecated - now merged with Tasks).
func NewMonitorPage(win fyne.Window) fyne.CanvasObject {
	// Return the combined task/monitor page
	return NewTaskMonitorPage(win)
}

// NewHistoryPage creates the history page.
func NewHistoryPage(win fyne.Window) fyne.CanvasObject {
	_, content := NewHistoryRecordPage(win, nil, nil)
	return content
}

// NewComparisonPage creates the result comparison page.
func NewComparisonPage(win fyne.Window, comparisonUC *usecase.ComparisonUseCase) fyne.CanvasObject {
	_, content := NewResultComparisonPage(win, comparisonUC)
	return content
}

// NewReportPage creates the report export page.
func NewReportPage(win fyne.Window) fyne.CanvasObject {
	return NewReportExportPage(win)
}

// NewSettingsPage creates the settings page.
func NewSettingsPage(win fyne.Window, connUC *usecase.ConnectionUseCase) fyne.CanvasObject {
	return NewSettingsConfigurationPage(win, connUC)
}
