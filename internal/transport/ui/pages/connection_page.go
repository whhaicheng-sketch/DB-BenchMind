// Package pages provides GUI pages for DB-BenchMind.
//
// Connections Page - Completion: 100%
//
// Features Implemented:
// - ✅ List connections grouped by database type (MySQL, PostgreSQL, Oracle, SQL Server)
// - ✅ Add new connections with database-specific field labels and defaults
// - ✅ Edit existing connections
// - ✅ Delete connections with confirmation
// - ✅ Test connections with intelligent SSL/encryption detection
// - ✅ Database-specific icons (🐬 MySQL, 🐘 PostgreSQL, 🔴 Oracle, 🔷 SQL Server)
// - ✅ Dynamic labels: "Database" for MySQL/PostgreSQL/SQL Server, "SID" for Oracle
// - ✅ Field validation: PostgreSQL Database and Oracle SID are required
// - ✅ Auto-refresh when switching to Connections tab
// - ✅ Dialog remains open on save failure (name conflict, etc.)
// - ✅ Database-specific defaults:
//   - MySQL: Database can be empty
//   - PostgreSQL: Database defaults to "postgres"
//   - Oracle: SID defaults to "orcl"
//   - SQL Server: Database can be empty
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
	content         *fyne.Container // Main content container for Refresh()
}

// NewConnectionPage creates a new connection management page.
func NewConnectionPage(connUC *usecase.ConnectionUseCase, win fyne.Window) (*ConnectionPage, fyne.CanvasObject) {
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

	page.content = content

	return page, content
}

// Refresh reloads the connection list when switching to the Connections tab.
func (p *ConnectionPage) Refresh() {
	slog.Info("Connections: Refreshing page")
	p.loadConnections()
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

		// Get database icon
		var dbIcon string
		switch c := conn.(type) {
		case *connection.MySQLConnection:
			dbIcon = "🐬"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		case *connection.PostgreSQLConnection:
			dbIcon = "🐘"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		case *connection.OracleConnection:
			dbIcon = "🔴"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		case *connection.SQLServerConnection:
			dbIcon = "🔷"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
		}

		// Connection info label
		infoText := fmt.Sprintf("%s %s  |  %s@%s:%s", dbIcon, connName, username, host, portStr)
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

	// Test in background
	go func() {
		slog.Info("Connections: Testing connection", "name", conn.GetName())
		result, err := p.connUC.TestConnection(context.Background(), conn.GetID())

		// Show result dialog (safe to call from goroutine in Fyne)
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
	// Don't set default host - let user enter it manually
	d.portEntry = widget.NewEntry()
	d.portEntry.SetText("3306")
	d.dbEntry = widget.NewEntry()
	d.dbLabel = widget.NewLabel("Database") // Dynamic label, will be updated
	d.userEntry = widget.NewEntry()
	d.passEntry = widget.NewPasswordEntry()
	d.trustServerCertCheck = widget.NewCheck("Trust Server Certificate", func(checked bool) {
		// Handle trust server certificate change
	})
	d.trustServerCertCheck.SetChecked(true) // Default to true for SQL Server (recommended)
	d.trustServerCertCheck.Hide()          // Initially hidden, only show for SQL Server

	// updateDBLabel updates the Database/SID label and default value based on database type.
	updateDBLabel := func(dbType string, isAddMode bool) {
		switch dbType {
		case "MySQL":
			d.dbLabel.SetText("Database")
			if isAddMode {
				d.dbEntry.SetText("")
			}
		case "PostgreSQL":
			d.dbLabel.SetText("Database")
			if isAddMode {
				d.dbEntry.SetText("postgres")
			}
		case "Oracle":
			d.dbLabel.SetText("SID")
			if isAddMode {
				d.dbEntry.SetText("orcl")
			}
		case "SQL Server":
			d.dbLabel.SetText("Database")
			if isAddMode {
				d.dbEntry.SetText("")
			}
		}
	}

	// Determine the initial database type for Edit mode
	var displayType string
	if d.isEditMode && d.conn != nil {
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
	} else {
		displayType = "MySQL" // Default
	}

	// Determine initial label text
	initialLabelText := "Database"
	if displayType == "Oracle" {
		initialLabelText = "SID"
	}

	// Create database type selector (will be populated with callback later)
	d.dbTypeSelect = widget.NewSelect([]string{"MySQL", "PostgreSQL", "Oracle", "SQL Server"}, nil)
	d.dbTypeSelect.SetSelected(displayType) // Set initial selection

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

		// Set other fields based on connection type
		switch c := d.conn.(type) {
		case *connection.MySQLConnection:
			d.hostEntry.SetText(c.Host)
			if c.Port > 0 {
				d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			} else {
				d.portEntry.SetText("3306")
			}
			d.dbEntry.SetText(c.Database)
			d.userEntry.SetText(c.Username)
			d.passEntry.SetText(c.Password)
		case *connection.PostgreSQLConnection:
			d.hostEntry.SetText(c.Host)
			if c.Port > 0 {
				d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			} else {
				d.portEntry.SetText("5432")
			}
			d.dbEntry.SetText(c.Database)
			d.userEntry.SetText(c.Username)
			d.passEntry.SetText(c.Password)
		case *connection.OracleConnection:
			d.hostEntry.SetText(c.Host)
			if c.Port > 0 {
				d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			} else {
				d.portEntry.SetText("1521")
			}
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
		// New connection - load default config if available (but NOT host)
		defaultConfig, err := connection.LoadDefaultConnectionConfig()
		if err == nil && defaultConfig != nil {
			// Load defaults for MySQL (default selection) - but don't load host
			if defaultConfig.MySQL != nil {
				// Don't load host - user should enter it manually
				if defaultConfig.MySQL.Port > 0 {
					d.portEntry.SetText(fmt.Sprintf("%d", defaultConfig.MySQL.Port))
				}
				d.dbEntry.SetText(defaultConfig.MySQL.Database)
				d.userEntry.SetText(defaultConfig.MySQL.Username)
			}
			slog.Info("Connections: Loaded default config", "db_type", "MySQL")
		}
	}

	// Determine dialog title
	title := "Add Connection"
	if d.isEditMode {
		title = "Edit Connection"
	}

	// Create form items with dynamic Database/SID label
	formItems := []*widget.FormItem{
		widget.NewFormItem("Database Type", d.dbTypeSelect),
		widget.NewFormItem("Name", d.nameEntry),
		widget.NewFormItem("Host", d.hostEntry),
		widget.NewFormItem("Port", d.portEntry),
		widget.NewFormItem(initialLabelText, d.dbEntry),
		widget.NewFormItem("Username", d.userEntry),
		widget.NewFormItem("Password", d.passEntry),
	}

	// Store reference to the Database/SID FormItem so we can update its label
	dbFormItem := formItems[4] // Index 4 is the Database/SID field

	// Create form
	form := widget.NewForm(formItems...)

	// Set the callback for dbTypeSelect now that we have dbFormItem and form
	d.dbTypeSelect.OnChanged = func(s string) {
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

		// Update label and default database/SID based on database type
		isAddMode := !d.isEditMode
		updateDBLabel(s, isAddMode)

		// Update FormItem label text
		switch s {
		case "MySQL", "PostgreSQL", "SQL Server":
			dbFormItem.Text = "Database"
		case "Oracle":
			dbFormItem.Text = "SID"
		}
		form.Refresh() // Refresh the form to show updated label
	}

	// Create buttons first (before dialog)
	btnTest := widget.NewButton("Test", func() {
		slog.Info("Connections: Dialog Test button clicked", "name", d.nameEntry.Text, "type", d.dbTypeSelect.Selected)
		d.onTestInDialog()
		// Note: dialog remains open after Test
	})
	btnSave := widget.NewButton("Save", func() {
		slog.Info("Connections: Dialog Save button clicked", "name", d.nameEntry.Text, "type", d.dbTypeSelect.Selected, "mode", map[bool]string{true: "edit", false: "add"}[d.isEditMode])
		success := d.onSave(win)
		if success {
			d.dialog.Hide() // Only close dialog if save was successful
		}
	})
	btnSave.Importance = widget.HighImportance
	btnCancel := widget.NewButton("Cancel", func() {
		slog.Info("Connections: Dialog Cancel button clicked", "name", d.nameEntry.Text, "type", d.dbTypeSelect.Selected)
		// Will be set to close dialog after dialog is created
	})

	buttonContainer := container.NewHBox(btnTest, btnSave, btnCancel)

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
// Returns true if save was successful, false otherwise.
func (d *connectionDialog) onSave(win fyne.Window) bool {
	ctx := context.Background()
	now := time.Now()

	dbType := d.dbTypeSelect.Selected
	name := strings.TrimSpace(d.nameEntry.Text)
	host := strings.TrimSpace(d.hostEntry.Text)
	portStr := strings.TrimSpace(d.portEntry.Text)
	port, err := strconv.Atoi(portStr)
	if portStr == "" || err != nil || port <= 0 {
		// Set default port based on database type
		switch dbType {
		case "MySQL":
			port = 3306
		case "PostgreSQL":
			port = 5432
		case "Oracle":
			port = 1521
		case "SQL Server":
			port = 1433
		}
		slog.Info("Connections: Using default port", "db_type", dbType, "port", port)
	}
	database := strings.TrimSpace(d.dbEntry.Text)
	username := strings.TrimSpace(d.userEntry.Text)
	password := d.passEntry.Text

	// Set default TrustServerCertificate for SQL Server
	trustServerCert := true // Default to true for SQL Server

	// In edit mode, if password field is empty, reload from keyring
	if d.isEditMode && d.conn != nil && password == "" {
		slog.Info("Connections: Loading password from keyring for save",
			"conn_id", d.conn.GetID())
		connWithPassword, err := d.connUC.GetConnectionByID(ctx, d.conn.GetID())
		if err != nil {
			slog.Error("Connections: Failed to load password from keyring", "error", err)
			dialog.ShowError(fmt.Errorf("failed to load password: %w", err), win)
			return false
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
		return false
	}

	// Delete old connection in edit mode before creating new one
	if d.isEditMode && d.conn != nil {
		if err := d.connUC.DeleteConnection(ctx, d.conn.GetID()); err != nil {
			dialog.ShowError(fmt.Errorf("failed to update connection: %w", err), win)
			return false
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
		return false
	}
	// Validate
	if err := conn.Validate(); err != nil {
		slog.Warn("Connections: Save validation failed", "name", name, "error", err)
		dialog.ShowError(fmt.Errorf("validation: %w", err), win)
		return false
	}
	// Save
	if err := d.connUC.CreateConnection(ctx, conn); err != nil {
		slog.Error("Connections: Failed to save", "name", name, "error", err)
		dialog.ShowError(fmt.Errorf("save: %w", err), win)
		return false
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

	if d.onSuccess != nil {
		d.onSuccess()
	}

	return true // Save was successful
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

	// Validate required fields before testing
	host := strings.TrimSpace(d.hostEntry.Text)
	database := strings.TrimSpace(d.dbEntry.Text)
	username := strings.TrimSpace(d.userEntry.Text)
	password := d.passEntry.Text
	dbType := d.dbTypeSelect.Selected

	if host == "" {
		dialog.ShowError(fmt.Errorf("host required"), d.win)
		return
	}
	// Validate database/SID based on database type
	// MySQL and SQL Server: Database can be empty
	// PostgreSQL: Database is required
	// Oracle: SID is required
	if database == "" && (dbType == "PostgreSQL" || dbType == "Oracle") {
		fieldName := "Database"
		if dbType == "Oracle" {
			fieldName = "SID"
		}
		dialog.ShowError(fmt.Errorf("%s is required", fieldName), d.win)
		return
	}
	if username == "" {
		dialog.ShowError(fmt.Errorf("username required"), d.win)
		return
	}
	if password == "" {
		dialog.ShowError(fmt.Errorf("password required"), d.win)
		return
	}

	// Test connection in background - always use form values (both ADD and EDIT modes)
	go func() {
		var result *connection.TestResult
		var err error

		// Both ADD and EDIT modes use form values for testing
		mode := "ADD"
		if d.isEditMode {
			mode = "EDIT"
		}

		slog.Info("Connections: Testing in dialog",
			"mode", mode,
			"name", name,
			"host", host,
			"database", database,
			"username", username)

		portStr := strings.TrimSpace(d.portEntry.Text)
		port, portErr := strconv.Atoi(portStr)
		dbType := d.dbTypeSelect.Selected
		trustServerCert := d.trustServerCertCheck.Checked

		// Set default port if empty or invalid
		if portStr == "" || portErr != nil || port <= 0 {
			switch dbType {
			case "MySQL":
				port = 3306
			case "PostgreSQL":
				port = 5432
			case "Oracle":
				port = 1521
			case "SQL Server":
				port = 1433
			}
			slog.Info("Connections: Using default port for test", "db_type", dbType, "port", port)
		}

		// Create temporary connection from form values
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
			dialog.ShowError(fmt.Errorf("unsupported type: %s", dbType), d.win)
			return
		}

		// Validate
		if err := conn.Validate(); err != nil {
			slog.Warn("Connections: Dialog test validation failed", "name", name, "error", err)
			dialog.ShowError(fmt.Errorf("validation: %w", err), d.win)
			return
		}

		// Test
		result, err = conn.Test(ctx)

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
	dbLabel              *widget.Label // Dynamic label for Database/SID field
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
