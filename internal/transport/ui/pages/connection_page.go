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
		var sshEnabled bool
		var winrmEnabled bool

		// Get database icon
		var dbIcon string
		switch c := conn.(type) {
		case *connection.MySQLConnection:
			dbIcon = "🐬"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
			sshEnabled = c.SSH != nil && c.SSH.Enabled
		case *connection.PostgreSQLConnection:
			dbIcon = "🐘"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
			sshEnabled = c.SSH != nil && c.SSH.Enabled
		case *connection.OracleConnection:
			dbIcon = "🔴"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
			sshEnabled = c.SSH != nil && c.SSH.Enabled
		case *connection.SQLServerConnection:
			dbIcon = "🔷"
			host = c.Host
			portStr = fmt.Sprintf("%d", c.Port)
			username = c.Username
			winrmEnabled = c.WinRM != nil && c.WinRM.Enabled
		}

		// Connection info label with SSH/WinRM status
		tunnelIndicator := ""
		if sshEnabled {
			tunnelIndicator = " | 🔒 SSH"
		}
		if winrmEnabled {
			tunnelIndicator = " | 🖥️ WinRM"
		}
		infoText := fmt.Sprintf("%s %s  |  %s@%s:%s%s", dbIcon, connName, username, host, portStr, tunnelIndicator)
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

		ctx := context.Background()

		// First, load connection with passwords from keyring to get SSH config
		connWithPasswords, err := p.connUC.GetConnectionByID(ctx, conn.GetID())
		if err != nil {
			slog.Error("Connections: Failed to load connection with passwords", "error", err)
			dialog.ShowError(fmt.Errorf("failed to load connection: %w", err), win)
			return
		}

		// Check if connection has SSH configured (from loaded connection)
		var sshConfig *connection.SSHTunnelConfig
		switch c := connWithPasswords.(type) {
		case *connection.MySQLConnection:
			sshConfig = c.SSH
		case *connection.PostgreSQLConnection:
			sshConfig = c.SSH
		case *connection.OracleConnection:
			sshConfig = c.SSH
		}

		// Test results
		var sshSuccess bool
		var sshError error
		var sshLatency int64
		var dbSuccess bool
		var dbError error
		var dbResult *connection.TestResult
		var dbConnectedDirectly bool // Whether we connected without SSH

		// Test SSH tunnel if configured
		if sshConfig != nil && sshConfig.Enabled {
			slog.Info("Connections: Testing SSH tunnel",
				"ssh_host", sshConfig.Host,
				"ssh_port", sshConfig.Port,
				"ssh_user", sshConfig.Username,
				"has_password", sshConfig.Password != "")

			start := time.Now()

			// Test SSH connection
			tunnel, err := connection.NewSSHTunnel(ctx, sshConfig, "localhost", 22)
			if err != nil {
				slog.Error("Connections: SSH test failed", "error", err)
				sshError = err
				sshSuccess = false
			} else {
				tunnel.Close()
				sshLatency = time.Since(start).Milliseconds()
				sshSuccess = true
				slog.Info("Connections: SSH test successful",
					"ssh_host", sshConfig.Host,
					"latency_ms", sshLatency)
			}
		}

		// Test database connection
		// If SSH succeeded, use SSH tunnel
		// If SSH failed or not configured, test direct connection
		if sshSuccess {
			slog.Info("Connections: Testing database through SSH tunnel")
			result, err := p.connUC.TestConnection(ctx, conn.GetID())
			if err != nil {
				dbSuccess = false
				dbError = err
				slog.Error("Connections: Database test failed", "error", err)
			} else if result.Success {
				dbSuccess = true
				dbResult = result
				slog.Info("Connections: Database test successful",
					"latency_ms", result.LatencyMs,
					"version", result.DatabaseVersion)
			} else {
				dbSuccess = false
				dbError = fmt.Errorf("%s", result.Error)
				slog.Warn("Connections: Database test failed", "error", result.Error)
			}
		} else {
			// SSH failed or not configured, test direct database connection
			if sshConfig != nil && sshConfig.Enabled {
				slog.Info("Connections: SSH failed, testing direct database connection",
					"reason", "SSH tunnel not available")
			} else {
				slog.Info("Connections: Testing direct database connection",
					"reason", "SSH not configured")
			}

			// Create a connection copy without SSH for direct testing
			connWithoutSSH := p.createConnectionWithoutSSH(connWithPasswords)
			result, err := connWithoutSSH.Test(ctx)
			dbConnectedDirectly = true

			if err != nil {
				dbSuccess = false
				dbError = err
				slog.Error("Connections: Direct database test failed", "error", err)
			} else if result.Success {
				dbSuccess = true
				dbResult = result
				slog.Info("Connections: Direct database test successful",
					"latency_ms", result.LatencyMs,
					"version", result.DatabaseVersion)
			} else {
				dbSuccess = false
				dbError = fmt.Errorf("%s", result.Error)
				slog.Warn("Connections: Direct database test failed", "error", result.Error)
			}
		}

		// Build comprehensive test result message
		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("Connection Test Results: %s\n\n", conn.GetName()))

		// SSH Tunnel section
		if sshConfig != nil && sshConfig.Enabled {
			msg.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			msg.WriteString("📡 SSH TUNNEL\n")
			if sshSuccess {
				msg.WriteString(fmt.Sprintf("  Status: ✓ Connected\n  Host: %s\n  Port: %d\n  User: %s\n  Latency: %dms\n",
					sshConfig.Host, sshConfig.Port, sshConfig.Username, sshLatency))
			} else {
				msg.WriteString(fmt.Sprintf("  Status: ✗ Failed\n  Host: %s\n  Port: %d\n  User: %s\n  Error: %v\n",
					sshConfig.Host, sshConfig.Port, sshConfig.Username, sshError))
			}
		}

		// Database section
		msg.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		msg.WriteString("💾 DATABASE\n")
		if dbSuccess {
			if dbConnectedDirectly && (sshConfig != nil && sshConfig.Enabled) {
				msg.WriteString(fmt.Sprintf("  Status: ✓ Connected (Direct, without SSH)\n  Version: %s\n  Latency: %dms\n  ⚠️  SSH tunnel was not used\n",
					dbResult.DatabaseVersion, dbResult.LatencyMs))
			} else {
				msg.WriteString(fmt.Sprintf("  Status: ✓ Connected\n  Version: %s\n  Latency: %dms\n",
					dbResult.DatabaseVersion, dbResult.LatencyMs))
			}
		} else {
			if dbConnectedDirectly {
				msg.WriteString(fmt.Sprintf("  Status: ✗ Failed (Direct connection)\n  Error: %v\n", dbError))
			} else {
				msg.WriteString(fmt.Sprintf("  Status: ✗ Failed\n  Error: %v\n", dbError))
			}
		}

		// Add helpful note based on results
		hasSSH := sshConfig != nil && sshConfig.Enabled
		if hasSSH && !sshSuccess && dbSuccess && dbConnectedDirectly {
			msg.WriteString("\n💡 Note: Database is directly accessible without SSH tunnel.\n")
		} else if hasSSH && !sshSuccess && !dbSuccess {
			msg.WriteString("\n💡 Note: SSH tunnel failed. Direct database connection also failed.\n")
		}

		// Always show the detailed test results
		dialog.ShowInformation("Connection Test", msg.String(), win)

		// Show error dialog only if both failed
		if !dbSuccess {
			dialog.ShowError(fmt.Errorf("database connection failed"), win)
		}
	}()
}

// createConnectionWithoutSSH creates a copy of connection without SSH configuration for direct testing
func (p *ConnectionPage) createConnectionWithoutSSH(conn connection.Connection) connection.Connection {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return &connection.MySQLConnection{
			BaseConnection: c.BaseConnection,
			Host:           c.Host,
			Port:           c.Port,
			Database:       c.Database,
			Username:       c.Username,
			Password:       c.Password,
			SSLMode:        c.SSLMode,
			SSH:            nil, // Remove SSH
		}
	case *connection.PostgreSQLConnection:
		return &connection.PostgreSQLConnection{
			BaseConnection: c.BaseConnection,
			Host:           c.Host,
			Port:           c.Port,
			Database:       c.Database,
			Username:       c.Username,
			Password:       c.Password,
			SSLMode:        c.SSLMode,
			SSH:            nil, // Remove SSH
		}
	case *connection.OracleConnection:
		return &connection.OracleConnection{
			BaseConnection: c.BaseConnection,
			Host:           c.Host,
			Port:           c.Port,
			ServiceName:    c.ServiceName,
			SID:            c.SID,
			Username:       c.Username,
			Password:       c.Password,
			SSH:            nil, // Remove SSH
		}
	case *connection.SQLServerConnection:
		return c // SQL Server doesn't support SSH
	}
	return conn
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

	// Declare button variable that will be used in both SSH config and dialog buttons
	var btnTestSSH *widget.Button
	var btnTestWinRM *widget.Button

	// Create form fields
	d.nameEntry = widget.NewEntry()
	d.hostEntry = widget.NewEntry()
	// Don't set default host - let user enter it manually
	d.portEntry = widget.NewEntry()
	d.portEntry.SetText("3306")
	d.dbEntry = widget.NewEntry()
	d.dbLabel = widget.NewLabel("Database") // Dynamic label, will be updated
	d.userEntry = widget.NewEntry()
	d.passEntry = widget.NewEntry()
	d.passEntry.Password = true
	d.trustServerCertCheck = widget.NewCheck("Trust Server Certificate", func(checked bool) {
		// Handle trust server certificate change
	})
	d.trustServerCertCheck.SetChecked(true) // Default to true for SQL Server (recommended)
	d.trustServerCertCheck.Hide()          // Initially hidden, only show for SQL Server

	// Create SSH configuration fields
	d.sshEnabledCheck = widget.NewCheck("Enable SSH Tunnel", func(checked bool) {
		// Show/hide SSH fields and update test buttons based on checkbox
		if checked {
			d.sshContainer.Show()
		} else {
			d.sshContainer.Hide()
		}
	})
	d.sshPortEntry = widget.NewEntry()
	d.sshPortEntry.SetText("22")
	d.sshPortEntry.SetPlaceHolder("Port")
	d.sshUserEntry = widget.NewEntry()
	d.sshUserEntry.SetText("root") // Default SSH username
	d.sshUserEntry.SetPlaceHolder("SSH username")
	d.sshPassEntry = widget.NewEntry()
	d.sshPassEntry.Password = true
	d.sshPassEntry.SetPlaceHolder("SSH password")

	// Create SSH container (initially hidden)
	sshHeader := container.NewHBox(
		widget.NewLabel("SSH Configuration"),
	)
	sshForm := widget.NewForm(
		widget.NewFormItem("SSH Port", d.sshPortEntry),
		widget.NewFormItem("SSH Username", d.sshUserEntry),
		widget.NewFormItem("SSH Password", d.sshPassEntry),
	)
	// Add help text
	sshHelpText := widget.NewLabel("💡 SSH Host uses Database Host")
	sshHelpText.Importance = widget.LowImportance
	d.sshContainer = container.NewVBox(widget.NewSeparator(), sshHeader, sshForm, sshHelpText)
	d.sshContainer.Hide() // Initially hidden

	// Create WinRM configuration fields (only for SQL Server)
	d.winrmEnabledCheck = widget.NewCheck("Enable WinRM (Windows Remote Management)", func(checked bool) {
		// Show/hide WinRM fields based on checkbox
		if checked {
			d.winrmContainer.Show()
		} else {
			d.winrmContainer.Hide()
		}
	})
	d.winrmPortEntry = widget.NewEntry()
	d.winrmPortEntry.SetText("5985")
	d.winrmPortEntry.SetPlaceHolder("Port")
	d.winrmHTTPSCheck = widget.NewCheck("Use HTTPS", func(checked bool) {
		// Auto-update port based on HTTPS selection
		if checked {
			d.winrmPortEntry.SetText("5986")
		} else {
			d.winrmPortEntry.SetText("5985")
		}
	})
	d.winrmUserEntry = widget.NewEntry()
	d.winrmUserEntry.SetPlaceHolder("WinRM username (empty = integrated Windows auth)")
	d.winrmPassEntry = widget.NewEntry()
	d.winrmPassEntry.Password = true
	d.winrmPassEntry.SetPlaceHolder("WinRM password")

	// Create WinRM container (initially hidden)
	winrmHeader := container.NewHBox(
		widget.NewLabel("WinRM Configuration"),
		widget.NewButton("❓ 配置帮助", func() {
			d.showWinRMHelpDialog()
		}),
	)
	winrmForm := widget.NewForm(
		widget.NewFormItem("WinRM Port", d.winrmPortEntry),
		widget.NewFormItem("", d.winrmHTTPSCheck),
		widget.NewFormItem("WinRM Username", d.winrmUserEntry),
		widget.NewFormItem("WinRM Password", d.winrmPassEntry),
	)
	// Add help text
	winrmHelpText := widget.NewLabel("💡 WinRM Host uses Database Host")
	winrmHelpText.Importance = widget.LowImportance
	d.winrmContainer = container.NewVBox(widget.NewSeparator(), winrmHeader, winrmForm, winrmHelpText)
	d.winrmContainer.Hide() // Initially hidden

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

	// Variable to store SSH config for loading after updateTestButtons is defined
	var loadedSSHConfig *connection.SSHTunnelConfig
	var loadedWinRMConfig *connection.WinRMConfig

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
			// Store SSH config for loading after UI is fully set up
			if c.SSH != nil {
				loadedSSHConfig = c.SSH
			}
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
			// Store SSH config for loading after UI is fully set up
			if c.SSH != nil {
				loadedSSHConfig = c.SSH
			}
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
			// Store SSH config for loading after UI is fully set up
			if c.SSH != nil {
				loadedSSHConfig = c.SSH
			}
		case *connection.SQLServerConnection:
			d.hostEntry.SetText(c.Host)
			d.portEntry.SetText(fmt.Sprintf("%d", c.Port))
			d.dbEntry.SetText(c.Database)
			d.userEntry.SetText(c.Username)
			d.passEntry.SetText(c.Password)
			d.trustServerCertCheck.SetChecked(c.TrustServerCertificate)
			// Store WinRM config for loading after UI is fully set up
			if c.WinRM != nil {
				loadedWinRMConfig = c.WinRM
			}
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

		// Show/hide SSH configuration based on database type
		// SSH is not supported for SQL Server
		if s == "SQL Server" {
			d.sshEnabledCheck.Hide()
			d.sshContainer.Hide()
		} else {
			d.sshEnabledCheck.Show()
			if d.sshEnabledCheck.Checked {
				d.sshContainer.Show()
			}
		}

		// Show/hide WinRM configuration based on database type
		// WinRM is only supported for SQL Server
		if s == "SQL Server" {
			d.winrmEnabledCheck.Show()
		} else {
			d.winrmEnabledCheck.Hide()
			d.winrmContainer.Hide()
		}

		form.Refresh() // Refresh the form to show updated label
	}

	// Create buttons first (before dialog)
	// When SSH is enabled, show two test buttons: "Test SSH" and "Test Database"
	// When SSH is disabled, show only "Test" button
	btnTestSSH = widget.NewButton("Test SSH", func() {
		slog.Info("Connections: Dialog Test SSH button clicked", "name", d.nameEntry.Text)
		d.onTestSSHConnection()
	})
	btnTestSSH.Importance = widget.MediumImportance

	btnTestWinRM = widget.NewButton("Test WinRM", func() {
		slog.Info("Connections: Dialog Test WinRM button clicked", "name", d.nameEntry.Text)
		d.onTestWinRMConnection()
	})
	btnTestWinRM.Importance = widget.MediumImportance

	btnTestDatabase := widget.NewButton("Test Database", func() {
		slog.Info("Connections: Dialog Test Database button clicked", "name", d.nameEntry.Text, "type", d.dbTypeSelect.Selected)
		d.onTestInDialog()
	})
	btnTestDatabase.Importance = widget.MediumImportance

	// Container for test buttons (will show either single Test or multiple buttons)
	testButtonsContainer := container.NewHBox(btnTestDatabase)

	// Function to update test buttons based on SSH/WinRM state
	updateTestButtons := func() {
		sshChecked := d.sshEnabledCheck.Checked
		winrmChecked := d.winrmEnabledCheck.Checked
		dbType := d.dbTypeSelect.Selected

		// Update SSH container visibility
		if sshChecked {
			d.sshContainer.Show()
		} else {
			d.sshContainer.Hide()
		}

		// Update WinRM container visibility
		if winrmChecked {
			d.winrmContainer.Show()
		} else {
			d.winrmContainer.Hide()
		}

		// Update test buttons - Test Database first, then Test SSH or Test WinRM
		testButtonsContainer.Objects = nil
		testButtonsContainer.Add(btnTestDatabase)

		// SSH is only for MySQL, PostgreSQL, Oracle
		if sshChecked && dbType != "SQL Server" {
			testButtonsContainer.Add(btnTestSSH)
		}

		// WinRM is only for SQL Server
		if winrmChecked && dbType == "SQL Server" {
			testButtonsContainer.Add(btnTestWinRM)
		}

		testButtonsContainer.Refresh()
	}

	// Watch SSH checkbox changes to update test buttons
	d.sshEnabledCheck.OnChanged = func(checked bool) {
		updateTestButtons()
	}

	// Watch WinRM checkbox changes to update test buttons
	d.winrmEnabledCheck.OnChanged = func(checked bool) {
		updateTestButtons()
	}

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

	buttonContainer := container.NewHBox(btnSave, btnCancel)

	// Create SSH Tunnel checkbox row (shown before SSH container)
	sshCheckboxRow := container.NewVBox(
		d.sshEnabledCheck,
	)

	// Create WinRM checkbox row (shown before WinRM container)
	winrmCheckboxRow := container.NewVBox(
		d.winrmEnabledCheck,
	)

	// Create dialog content with buttons at bottom
	// Layout:
	// 1. Form (database fields)
	// 2. Test button(s)
	// 3. SSH Tunnel checkbox and container (for MySQL, PostgreSQL, Oracle)
	// 4. WinRM checkbox and container (for SQL Server)
	// 5. Separator
	// 6. Save/Cancel buttons
	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		testButtonsContainer,
		widget.NewSeparator(),
		sshCheckboxRow,
		d.sshContainer,
		winrmCheckboxRow,
		d.winrmContainer,
		widget.NewSeparator(),
		buttonContainer,
	)

	// Create custom dialog without buttons
	dlg := dialog.NewCustomWithoutButtons(title, content, win)
	dlg.Resize(fyne.NewSize(500, 750)) // Increased height for SSH layout
	d.dialog = dlg // Store dialog reference

	// Update Cancel button to close dialog
	btnCancel.OnTapped = func() {
		dlg.Hide()
	}

	// Initialize SSH and WinRM visibility based on current database type
	if displayType == "SQL Server" {
		d.sshEnabledCheck.Hide()
		d.sshContainer.Hide()
		d.winrmEnabledCheck.Show() // Show WinRM for SQL Server
	} else {
		// Make sure SSH checkbox is visible for MySQL, PostgreSQL, Oracle
		d.sshEnabledCheck.Show()
		d.winrmEnabledCheck.Hide() // Hide WinRM for non-SQL Server
		d.winrmContainer.Hide()
	}

	// Load SSH configuration if it was stored earlier (after UI is fully set up)
	if loadedSSHConfig != nil {
		d.sshEnabledCheck.SetChecked(loadedSSHConfig.Enabled)
		if loadedSSHConfig.Port > 0 {
			d.sshPortEntry.SetText(fmt.Sprintf("%d", loadedSSHConfig.Port))
		}
		d.sshUserEntry.SetText(loadedSSHConfig.Username)

		// Try to load SSH password from keyring for edit mode
		if d.isEditMode && d.conn != nil {
			ctx := context.Background()
			sshKey := d.conn.GetID() + ":ssh"
			sshPassword, err := d.connUC.GetKeyring().Get(ctx, sshKey)
			if err == nil && sshPassword != "" {
				d.sshPassEntry.SetText(sshPassword)
				slog.Info("Connections: Loaded SSH password from keyring",
					"conn_id", d.conn.GetID())
			} else {
				slog.Info("Connections: SSH password not in keyring",
					"conn_id", d.conn.GetID(),
					"error", err)
			}
		}

		slog.Info("Connections: Loaded SSH config into UI",
			"enabled", loadedSSHConfig.Enabled,
			"port", loadedSSHConfig.Port,
			"username", loadedSSHConfig.Username)
		// Manually trigger updateTestButtons to show/hide Test SSH button
		updateTestButtons()
	}

	// Load WinRM configuration if it was stored earlier (after UI is fully set up)
	if loadedWinRMConfig != nil {
		d.winrmEnabledCheck.SetChecked(loadedWinRMConfig.Enabled)
		if loadedWinRMConfig.Port > 0 {
			d.winrmPortEntry.SetText(fmt.Sprintf("%d", loadedWinRMConfig.Port))
		}
		d.winrmHTTPSCheck.SetChecked(loadedWinRMConfig.UseHTTPS)
		d.winrmUserEntry.SetText(loadedWinRMConfig.Username)

		// Try to load WinRM password from keyring for edit mode
		if d.isEditMode && d.conn != nil {
			ctx := context.Background()
			winrmKey := d.conn.GetID() + ":winrm"
			winrmPassword, err := d.connUC.GetKeyring().Get(ctx, winrmKey)
			if err == nil && winrmPassword != "" {
				d.winrmPassEntry.SetText(winrmPassword)
				slog.Info("Connections: Loaded WinRM password from keyring",
					"conn_id", d.conn.GetID())
			} else {
				slog.Info("Connections: WinRM password not in keyring",
					"conn_id", d.conn.GetID(),
					"error", err)
			}
		}

		slog.Info("Connections: Loaded WinRM config into UI",
			"enabled", loadedWinRMConfig.Enabled,
			"port", loadedWinRMConfig.Port,
			"use_https", loadedWinRMConfig.UseHTTPS,
			"username", loadedWinRMConfig.Username)
		// Manually trigger updateTestButtons to show/hide Test WinRM button
		updateTestButtons()
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

	// Parse SSH configuration
	var sshConfig *connection.SSHTunnelConfig
	if d.sshEnabledCheck.Checked && dbType != "SQL Server" {
		sshPortStr := strings.TrimSpace(d.sshPortEntry.Text)
		sshPort, sshPortErr := strconv.Atoi(sshPortStr)
		if sshPortStr == "" || sshPortErr != nil || sshPort <= 0 {
			sshPort = 22 // Default SSH port
		}
		sshUser := strings.TrimSpace(d.sshUserEntry.Text)
		sshPass := d.sshPassEntry.Text

		// SSH Host uses the same host as database
		if host != "" && sshUser != "" {
			sshConfig = &connection.SSHTunnelConfig{
				Enabled:  true,
				Host:     host, // Use database host
				Port:     sshPort,
				Username: sshUser,
				Password: sshPass,
				LocalPort: 0, // Always auto-assign
			}
			slog.Info("Connections: SSH tunnel enabled",
				"ssh_host", host,
				"ssh_port", sshPort,
				"ssh_user", sshUser)
		}
	}

	// Parse WinRM configuration (only for SQL Server)
	var winrmConfig *connection.WinRMConfig
	if d.winrmEnabledCheck.Checked && dbType == "SQL Server" {
		winrmPortStr := strings.TrimSpace(d.winrmPortEntry.Text)
		winrmPort, winrmPortErr := strconv.Atoi(winrmPortStr)
		if winrmPortStr == "" || winrmPortErr != nil || winrmPort <= 0 {
			// Default WinRM port based on HTTPS setting
			if d.winrmHTTPSCheck.Checked {
				winrmPort = 5986 // HTTPS
			} else {
				winrmPort = 5985 // HTTP
			}
		}
		winrmUser := strings.TrimSpace(d.winrmUserEntry.Text)
		winrmPass := d.winrmPassEntry.Text
		useHTTPS := d.winrmHTTPSCheck.Checked

		// WinRM Host uses the same host as database
		if host != "" {
			winrmConfig = &connection.WinRMConfig{
				Enabled:  true,
				Host:     host, // Use database host
				Port:     winrmPort,
				Username: winrmUser,
				Password: winrmPass,
				UseHTTPS: useHTTPS,
			}
			slog.Info("Connections: WinRM enabled",
				"winrm_host", host,
				"winrm_port", winrmPort,
				"use_https", useHTTPS,
				"winrm_user", winrmUser)
		}
	}

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
			SSH:      sshConfig,
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
			SSH:      sshConfig,
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
			SSH:      sshConfig,
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
			WinRM:                  winrmConfig,
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

	// NOTE: Test Database button does NOT test SSH tunnel
	// It tests the database connection directly
	// SSH tunnel is tested separately by Test SSH button

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
			"username", username,
			"ssh_enabled", d.sshEnabledCheck.Checked,
			"note", "Test Database tests direct DB connection, SSH not used")

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

		// Create temporary connection from form values WITHOUT SSH config
		// Test Database button tests direct database connection
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
				SSH:      nil, // No SSH for Test Database button
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
				SSH:      nil, // No SSH for Test Database button
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
				SSH:      nil, // No SSH for Test Database button
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

	// SSH fields
	sshEnabledCheck *widget.Check
	sshPortEntry    *widget.Entry
	sshUserEntry    *widget.Entry
	sshPassEntry    *widget.Entry
	sshContainer    *fyne.Container // Container for SSH fields

	// WinRM fields (only for SQL Server)
	winrmEnabledCheck *widget.Check
	winrmPortEntry    *widget.Entry
	winrmHTTPSCheck   *widget.Check
	winrmUserEntry    *widget.Entry
	winrmPassEntry    *widget.Entry
	winrmContainer    *fyne.Container // Container for WinRM fields
}

// onTestSSHConnection tests the SSH connection only (without database).
func (d *connectionDialog) onTestSSHConnection() {
	ctx := context.Background()
	// SSH Host uses the database host
	host := strings.TrimSpace(d.hostEntry.Text)
	if host == "" {
		dialog.ShowError(fmt.Errorf("Database host is required (used as SSH host)"), d.win)
		return
	}
	sshPortStr := strings.TrimSpace(d.sshPortEntry.Text)
	sshPort, err := strconv.Atoi(sshPortStr)
	sshUser := strings.TrimSpace(d.sshUserEntry.Text)
	sshPass := d.sshPassEntry.Text

	// Validate SSH fields
	if sshPortStr == "" || err != nil || sshPort <= 0 || sshPort > 65535 {
		dialog.ShowError(fmt.Errorf("SSH port must be between 1 and 65535"), d.win)
		return
	}
	if sshUser == "" {
		dialog.ShowError(fmt.Errorf("SSH username is required"), d.win)
		return
	}

	// Test SSH connection in background
	go func() {
		slog.Info("Connections: Testing SSH connection",
			"ssh_host", host,
			"ssh_port", sshPort,
			"ssh_user", sshUser)

		start := time.Now()

		// Create SSH config and test connection
		sshConfig := &connection.SSHTunnelConfig{
			Enabled:  true,
			Host:     host, // Use database host
			Port:     sshPort,
			Username: sshUser,
			Password: sshPass,
			LocalPort: 0, // Auto-assign for testing
		}

		// Try to connect to SSH server
		// We'll use a dummy target (localhost:22) to just test SSH connection
		// The actual tunnel won't be used, we just want to verify SSH auth works
		tunnel, err := connection.NewSSHTunnel(ctx, sshConfig, "localhost", 22)
		if err != nil {
			slog.Error("Connections: SSH test failed", "error", err)
			dialog.ShowError(fmt.Errorf("SSH connection failed: %w", err), d.win)
			return
		}
		defer tunnel.Close()

		latency := time.Since(start).Milliseconds()
		localPort := tunnel.GetLocalPort()

		slog.Info("Connections: SSH test successful",
			"ssh_host", host,
			"ssh_port", sshPort,
			"latency_ms", latency,
			"local_port", localPort)

		msg := fmt.Sprintf("SSH connection successful!\n\nLatency: %dms\nLocal Port: %d (auto-assigned)\n\nYou can now test the database connection.",
			latency, localPort)
		dialog.ShowInformation("SSH Test", msg, d.win)
	}()
}

// onTestWinRMConnection tests the WinRM connection only (without database).
func (d *connectionDialog) onTestWinRMConnection() {
	ctx := context.Background()
	// WinRM Host uses the database host
	host := strings.TrimSpace(d.hostEntry.Text)
	if host == "" {
		dialog.ShowError(fmt.Errorf("Database host is required (used as WinRM host)"), d.win)
		return
	}
	winrmPortStr := strings.TrimSpace(d.winrmPortEntry.Text)
	winrmPort, err := strconv.Atoi(winrmPortStr)
	useHTTPS := d.winrmHTTPSCheck.Checked
	winrmUser := strings.TrimSpace(d.winrmUserEntry.Text)
	winrmPass := d.winrmPassEntry.Text

	// Validate WinRM fields
	if winrmPortStr == "" || err != nil || winrmPort <= 0 || winrmPort > 65535 {
		dialog.ShowError(fmt.Errorf("WinRM port must be between 1 and 65535"), d.win)
		return
	}

	// Validate standard ports
	if useHTTPS && winrmPort != 5986 {
		dialog.ShowError(fmt.Errorf("HTTPS requires port 5986, got %d", winrmPort), d.win)
		return
	}
	if !useHTTPS && winrmPort != 5985 {
		dialog.ShowError(fmt.Errorf("HTTP requires port 5985, got %d", winrmPort), d.win)
		return
	}

	// Test WinRM connection in background
	go func() {
		slog.Info("Connections: Testing WinRM connection",
			"winrm_host", host,
			"winrm_port", winrmPort,
			"use_https", useHTTPS,
			"winrm_user", winrmUser)

		// Create WinRM config and test connection
		winrmConfig := &connection.WinRMConfig{
			Enabled:  true,
			Host:     host, // Use database host
			Port:     winrmPort,
			Username: winrmUser,
			Password: winrmPass,
			UseHTTPS: useHTTPS,
		}

		// Try to connect to WinRM server
		client, err := connection.NewWinRMClient(ctx, winrmConfig)
		if err != nil {
			slog.Error("Connections: WinRM test failed", "error", err)
			// Show error dialog with help button
			d.showWinRMErrorDialog(fmt.Errorf("WinRM connection failed: %w", err), true)
			return
		}
		defer client.Close()

		// Test the connection
		result, err := client.Test(ctx)
		if err != nil {
			slog.Error("Connections: WinRM test error", "error", err)
			d.showWinRMErrorDialog(fmt.Errorf("WinRM test failed: %w", err), true)
			return
		}

		if !result.Success {
			slog.Error("Connections: WinRM test failed", "error", result.Error)
			d.showWinRMErrorDialog(fmt.Errorf("WinRM connection failed: %s", result.Error), true)
			return
		}

		slog.Info("Connections: WinRM test successful",
			"winrm_host", host,
			"winrm_port", winrmPort,
			"latency_ms", result.LatencyMs)

		msg := fmt.Sprintf("WinRM connection successful!\n\nLatency: %dms\n\nYou can now test the database connection.",
			result.LatencyMs)
		dialog.ShowInformation("WinRM Test", msg, d.win)
	}()
}

// showWinRMHelpDialog 显示 WinRM 配置帮助对话框
func (d *connectionDialog) showWinRMHelpDialog() {
	helpText := `WinRM 配置（数据库宿主机开启远程采集用）
适用：Windows Server 2012/2016/2019/2022

【方案1：HTTP（最简单，测试/内网）】
宿主机（管理员 PowerShell）执行：
  Enable-PSRemoting -Force
验证：
  Test-WSMan localhost
说明：端口 5985；多数情况下会自动放行防火墙。

【方案2：HTTPS（更安全，生产）】
宿主机（管理员 PowerShell）执行：
  Enable-PSRemoting -Force
  $cert = New-SelfSignedCertificate -CertStoreLocation Cert:\LocalMachine\My -DnsName $env:COMPUTERNAME
  New-Item -Path WSMan:\localhost\Listener -Transport HTTPS -Address * -CertificateThumbprint $cert.Thumbprint -Port 5986 -Force
验证：
  Test-WSMan localhost -UseSSL

【可选：工作组/非域时，客户端设置 TrustedHosts（在压测机上执行，不是宿主机）】
  Set-Item WSMan:\localhost\Client\TrustedHosts -Value "宿主机IP或主机名" -Force

【查看监听】
  winrm enumerate winrm/config/listener

【关闭 WinRM】
  Disable-PSRemoting -Force
`

	// 创建可选择和复制的文本框（自动换行，支持 Ctrl+A）
	helpEntry := widget.NewMultiLineEntry()
	helpEntry.SetText(helpText)
	helpEntry.Wrapping = fyne.TextWrapWord // 自动按单词换行

	// 创建对话框（不需要滚动容器，Entry 自带滚动）
	dlg := dialog.NewCustom("WinRM 配置帮助", "关闭", helpEntry, d.win)
	dlg.Resize(fyne.NewSize(650, 450))
	dlg.Show()
}

// showWinRMErrorDialog 显示 WinRM 错误对话框，带查看帮助按钮
func (d *connectionDialog) showWinRMErrorDialog(err error, showHelp bool) {
	errorMsg := fmt.Sprintf("WinRM 连接失败：%v\n\n可能的原因：\n1. WinRM 服务未在 Windows Server 上启用\n2. 防火墙阻止了连接\n3. 端口配置错误（HTTP: 5985, HTTPS: 5986）\n4. 用户名或密码错误", err)

	// 创建错误标签
	errorLabel := widget.NewLabel(errorMsg)
	errorLabel.Importance = widget.MediumImportance

	// 创建按钮
	btnHelp := widget.NewButton("查看配置帮助", func() {
		d.showWinRMHelpDialog()
	})
	btnHelp.Importance = widget.MediumImportance

	btnOK := widget.NewButton("关闭", func() {
		// Dialog will be closed
	})
	btnOK.Importance = widget.HighImportance

	buttonContainer := container.NewHBox(btnHelp, btnOK)

	// 创建对话框内容
	content := container.NewVBox(
		errorLabel,
		widget.NewSeparator(),
		buttonContainer,
	)

	// 创建自定义对话框
	dlg := dialog.NewCustomWithoutButtons("WinRM 测试失败", content, d.win)
	dlg.Resize(fyne.NewSize(500, 200))

	// 设置关闭按钮动作
	btnOK.OnTapped = func() {
		dlg.Hide()
	}

	dlg.Show()
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
