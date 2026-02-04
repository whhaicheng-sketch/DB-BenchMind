// Package pages provides GUI pages for DB-BenchMind.
// Template Management Page implementation (with add/delete/default template features).
package pages

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Global storage for custom templates (persists across page recreations)
var (
	customTemplates      []templateInfo
	customTemplatesMutex sync.RWMutex
	// Default template IDs for each database type (persists Set Default operations)
	defaultTemplateIDs   = map[string]string{
		"MySQL":      "sysbench-mysql-test",
		"PostgreSQL": "sysbench-postgresql-test",
		"Oracle":     "swingbench-oracle-test",
		"SQL Server": "", // No SQL Server templates yet
	}
)

// TemplateManagementPage provides the template management GUI.
type TemplateManagementPage struct {
	win             fyne.Window
	templates       []templateInfo
	defaultIndex    int                        // Index of default template
	listContainer   *fyne.Container            // Use VBox for dynamic list (like Connections)
	groupContainers map[string]*fyne.Container // DB type -> container
}

// templateInfo represents display info for a template.
type templateInfo struct {
	ID          string
	Name        string
	Description string
	Tool        string
	DBType      string // Database type: MySQL, PostgreSQL, Oracle, SQL Server
	IsBuiltin   bool
	IsDefault   bool
	Parameters  *OLTPParameters // OLTP parameters for sysbench
}

// OLTPParameters represents sysbench OLTP test parameters.
// Only includes parameters that are actually used by the sysbench adapter.
type OLTPParameters struct {
	Tables    int `json:"tables"`     // Number of tables to create
	TableSize int `json:"table_size"` // Number of rows per table
}

// NewTemplateManagementPage creates a new template management page.
func NewTemplateManagementPage(win fyne.Window) fyne.CanvasObject {
	slog.Info("Templates: NewTemplateManagementPage called - creating new page instance")

	page := &TemplateManagementPage{
		win:             win,
		defaultIndex:    0,
		templates:       []templateInfo{},
		groupContainers: make(map[string]*fyne.Container),
		listContainer:   container.NewVBox(),
	}

	// Load templates to populate the list
	page.loadTemplates()

	// Create toolbar with only Add button
	btnAdd := widget.NewButton("➕ Add Template", func() {
		slog.Info("Templates: Add Template button clicked")
		page.onAddTemplate()
	})

	toolbar := container.NewVBox(
		container.NewHBox(btnAdd),
	)

	// Create top area with toolbar
	topArea := container.NewVBox(
		toolbar,
		widget.NewSeparator(),
	)

	// Use Border layout
	content := container.NewBorder(
		topArea,                                 // top - toolbar
		nil,                                     // bottom
		nil,                                     // left
		nil,                                     // right
		container.NewScroll(page.listContainer), // center - fills available space
	)

	return content
}

// loadTemplatesData loads and returns template information.
func (p *TemplateManagementPage) loadTemplatesData() []templateInfo {
	// Built-in templates (cannot be deleted)
	// Test templates (default for each database type)
	testParams := &OLTPParameters{
		Tables:    10,
		TableSize: 10000,
	}

	// CPU Bound templates (data fits in memory)
	cpuBoundParams := &OLTPParameters{
		Tables:    10,
		TableSize: 10000000,
	}

	// Disk Bound templates (data exceeds memory)
	diskBoundParams := &OLTPParameters{
		Tables:    50,
		TableSize: 10000000,
	}

	// Create builtin templates (initially all IsDefault=false, will be set below)
	builtinTemplates := []templateInfo{
		// MySQL templates
		{
			ID:          "sysbench-mysql-test",
			Name:        "Test (Sysbench)",
			Description: "Lightweight test template for quick MySQL testing (10 tables, 10K rows each)",
			Tool:        "sysbench",
			DBType:      "MySQL",
			IsBuiltin:   true,
			IsDefault:   false, // Will be set based on defaultTemplateIDs
			Parameters:  testParams,
		},
		{
			ID:          "sysbench-mysql-cpu-bound",
			Name:        "CPU Bound (Sysbench)",
			Description: "CPU-bound test template for MySQL (10 tables, 10M rows each - fits in memory)",
			Tool:        "sysbench",
			DBType:      "MySQL",
			IsBuiltin:   true,
			IsDefault:   false,
			Parameters:  cpuBoundParams,
		},
		{
			ID:          "sysbench-mysql-disk-bound",
			Name:        "Disk Bound (Sysbench)",
			Description: "Disk-bound test template for MySQL (50 tables, 10M rows each - exceeds memory)",
			Tool:        "sysbench",
			DBType:      "MySQL",
			IsBuiltin:   true,
			IsDefault:   false,
			Parameters:  diskBoundParams,
		},
		// PostgreSQL templates
		{
			ID:          "sysbench-postgresql-test",
			Name:        "Test (Sysbench)",
			Description: "Lightweight test template for quick PostgreSQL testing (10 tables, 10K rows each)",
			Tool:        "sysbench",
			DBType:      "PostgreSQL",
			IsBuiltin:   true,
			IsDefault:   false, // Will be set based on defaultTemplateIDs
			Parameters:  testParams,
		},
		{
			ID:          "sysbench-postgresql-cpu-bound",
			Name:        "CPU Bound (Sysbench)",
			Description: "CPU-bound test template for PostgreSQL (10 tables, 10M rows each - fits in memory)",
			Tool:        "sysbench",
			DBType:      "PostgreSQL",
			IsBuiltin:   true,
			IsDefault:   false,
			Parameters:  cpuBoundParams,
		},
		{
			ID:          "sysbench-postgresql-disk-bound",
			Name:        "Disk Bound (Sysbench)",
			Description: "Disk-bound test template for PostgreSQL (50 tables, 10M rows each - exceeds memory)",
			Tool:        "sysbench",
			DBType:      "PostgreSQL",
			IsBuiltin:   true,
			IsDefault:   false,
			Parameters:  diskBoundParams,
		},
		// Oracle templates
		{
			ID:          "swingbench-oracle-test",
			Name:        "Test (Swingbench)",
			Description: "Lightweight test template for quick Oracle testing (1GB data, balanced read/write mix)",
			Tool:        "swingbench",
			DBType:      "Oracle",
			IsBuiltin:   true,
			IsDefault:   false, // Will be set based on defaultTemplateIDs
			Parameters:  nil, // Swingbench uses different parameters
		},
		{
			ID:          "swingbench-oracle-cpu-bound",
			Name:        "CPU Bound (Swingbench)",
			Description: "CPU-bound test template for Oracle - 85% read (Browse Products), 15% write operations. Uses 1GB data size.",
			Tool:        "swingbench",
			DBType:      "Oracle",
			IsBuiltin:   true,
			IsDefault:   false,
			Parameters:  nil, // Swingbench uses different parameters
		},
		{
			ID:          "swingbench-oracle-disk-bound",
			Name:        "Disk Bound (Swingbench)",
			Description: "Disk-bound test template for Oracle - Balanced read/write mix (35% Order Products, 35% Browse Products, 10% each for Customer operations). Uses 1GB data size.",
			Tool:        "swingbench",
			DBType:      "Oracle",
			IsBuiltin:   true,
			IsDefault:   false,
			Parameters:  nil, // Swingbench uses different parameters
		},
	}

	// Load custom templates from global storage
	customTemplatesMutex.RLock()
	defer customTemplatesMutex.RUnlock()
	slog.Info("Templates: Loading custom templates from global storage", "count", len(customTemplates))

	// Set default flag for builtin templates based on defaultTemplateIDs map
	// and clear default flag for custom templates that are NOT the default
	for i := range builtinTemplates {
		dbType := builtinTemplates[i].DBType
		defaultID := defaultTemplateIDs[dbType]
		if builtinTemplates[i].ID == defaultID {
			builtinTemplates[i].IsDefault = true
		} else {
			builtinTemplates[i].IsDefault = false
		}
	}

	// Update custom templates to match defaultTemplateIDs (if custom template is default)
	for i := range customTemplates {
		dbType := customTemplates[i].DBType
		defaultID := defaultTemplateIDs[dbType]
		customTemplates[i].IsDefault = (customTemplates[i].ID == defaultID)
	}

	// Combine all templates
	allTemplates := append([]templateInfo{}, builtinTemplates...)
	allTemplates = append(allTemplates, customTemplates...)

	slog.Info("Templates: Total templates loaded", "builtin", len(builtinTemplates), "custom", len(customTemplates), "total", len(allTemplates))
	return allTemplates
}

// loadTemplates loads template information and refreshes the list.
func (p *TemplateManagementPage) loadTemplates() {
	slog.Info("Templates: loadTemplates called")
	p.templates = p.loadTemplatesData()

	// Group templates by database type
	groups := make(map[string][]templateInfo)
	for _, tmpl := range p.templates {
		dbType := tmpl.DBType
		if dbType == "" {
			dbType = "MySQL" // Default to MySQL if not specified
		}
		groups[dbType] = append(groups[dbType], tmpl)
	}

	// Clear list container
	p.listContainer.Objects = nil
	p.groupContainers = make(map[string]*fyne.Container)

	// Define order of database types
	dbOrder := []string{"MySQL", "PostgreSQL", "Oracle", "SQL Server"}

	// Create collapsible groups
	for _, dbType := range dbOrder {
		templates := groups[dbType]
		if len(templates) == 0 {
			continue
		}

		slog.Info("Templates: Creating group", "db_type", dbType, "count", len(templates))
		p.createTemplateGroup(dbType, templates)
	}

	p.listContainer.Refresh()
	slog.Info("Templates: List refreshed", "total_templates", len(p.templates))
}

// createTemplateGroup creates a collapsible group for a database type.
func (p *TemplateManagementPage) createTemplateGroup(dbType string, templates []templateInfo) {
	slog.Info("Templates: Creating group", "db_type", dbType, "count", len(templates))

	// Group header with expand/collapse button
	headerText := fmt.Sprintf("📊 %s (%d)", dbType, len(templates))
	headerBtn := widget.NewButton(headerText, nil)

	// Container for this group's templates
	groupContainer := container.NewVBox()
	p.groupContainers[dbType] = groupContainer

	// Initially expanded
	isExpanded := true

	// Toggle expand/collapse
	headerBtn.OnTapped = func() {
		isExpanded = !isExpanded
		slog.Info("Templates: Group toggled", "db_type", dbType, "expanded", isExpanded)
		if isExpanded {
			groupContainer.Show()
		} else {
			groupContainer.Hide()
		}
	}

	// Add templates to this group
	for _, tmpl := range templates {
		// Create icon
		icon := "📄"
		if tmpl.IsBuiltin {
			icon = "📦"
		}
		if tmpl.IsDefault {
			icon += " ⭐"
		}

		// Template info label
		text := fmt.Sprintf("    %s %s", icon, tmpl.Name)
		infoLabel := widget.NewLabel(text)

		// Buttons for this template
		var buttons []fyne.CanvasObject

		// Built-in templates: Details, Set Default
		if tmpl.IsBuiltin {
			// Details button (first for built-in templates)
			btnDetails := widget.NewButton("📋 Details", func() {
				slog.Info("Templates: Details button clicked", "template", tmpl.Name)
				p.showTemplateDetails(tmpl)
			})
			buttons = append(buttons, btnDetails)

			// Set Default button (second for built-in templates)
			btnSetDefault := widget.NewButton("⭐ Set Default", func() {
				slog.Info("Templates: Set Default button clicked", "template", tmpl.Name, "db_type", tmpl.DBType)
				p.onSetDefault(tmpl, dbType)
			})
			buttons = append(buttons, btnSetDefault)
		} else {
			// Custom templates: Edit, Delete, Set Default
			// Edit button (first for custom templates)
			btnEdit := widget.NewButton("✏️ Edit", func() {
				slog.Info("Templates: Edit button clicked", "template", tmpl.Name)
				p.onEditTemplate(tmpl)
			})
			buttons = append(buttons, btnEdit)

			// Delete button (second for custom templates) - RED WARNING to stand out
			btnDelete := widget.NewButton("⚠️ 🗑️ Delete", func() {
				slog.Info("Templates: Delete button clicked", "template", tmpl.Name)
				p.onDeleteTemplate(tmpl)
			})
			// High importance makes it more prominent (usually red/orange in most themes)
			btnDelete.Importance = widget.HighImportance
			buttons = append(buttons, btnDelete)

			// Set Default button (third for custom templates)
			btnSetDefault := widget.NewButton("⭐ Set Default", func() {
				slog.Info("Templates: Set Default button clicked", "template", tmpl.Name, "db_type", tmpl.DBType)
				p.onSetDefault(tmpl, dbType)
			})
			buttons = append(buttons, btnSetDefault)
		}

		// Use Border layout to align info left, buttons right
		buttonBox := container.NewHBox(buttons...)
		templateRow := container.NewBorder(nil, nil, infoLabel, buttonBox)
		groupContainer.Add(templateRow)
	}

	// Add header and group to main list
	p.listContainer.Add(headerBtn)
	p.listContainer.Add(groupContainer)
}

// onAddTemplate adds a new custom template.
func (p *TemplateManagementPage) onAddTemplate() {
	slog.Info("Templates: Add Template button clicked")
	showTemplateDialog(p.win, "Add Template", nil, "", func(params *OLTPParameters, name string, dbType string) {
		slog.Info("Templates: Creating new template", "name", name, "db_type", dbType)

		// Create new template
		newTemplate := templateInfo{
			ID:          fmt.Sprintf("custom-%d", time.Now().UnixNano()),
			Name:        name,
			Description: "Custom template",
			Tool:        "sysbench",
			DBType:      dbType, // Set database type
			IsBuiltin:   false,
			IsDefault:   false,
			Parameters:  params,
		}

		// Save to global storage
		customTemplatesMutex.Lock()
		customTemplates = append(customTemplates, newTemplate)
		slog.Info("Templates: Saved to global storage", "name", name, "total_custom", len(customTemplates))
		customTemplatesMutex.Unlock()

		// Reload
		p.loadTemplates()

		slog.Info("Templates: Template added successfully", "name", name, "total_templates", len(p.templates))
		dialog.ShowInformation("Success", "Template added successfully", p.win)
	})
}

// onEditTemplate edits an existing template.
func (p *TemplateManagementPage) onEditTemplate(tmpl templateInfo) {
	// Cannot edit built-in templates
	if tmpl.IsBuiltin {
		slog.Warn("Templates: Attempted to edit built-in template", "name", tmpl.Name)
		dialog.ShowError(
			fmt.Errorf("cannot edit built-in template '%s'", tmpl.Name),
			p.win,
		)
		return
	}

	slog.Info("Templates: Editing template", "name", tmpl.Name, "db_type", tmpl.DBType)

	// Show dialog with existing parameters and DB type
	showTemplateDialogWithDBType(p.win, "Edit Template", tmpl.Parameters, tmpl.Name, tmpl.DBType, func(params *OLTPParameters, newName string, newDBType string) {
		slog.Info("Templates: Updating template", "old_name", tmpl.Name, "new_name", newName, "old_db_type", tmpl.DBType, "new_db_type", newDBType)

		// Update in global storage
		customTemplatesMutex.Lock()
		for i, ct := range customTemplates {
			if ct.ID == tmpl.ID {
				customTemplates[i].Name = newName
				customTemplates[i].Parameters = params
				customTemplates[i].DBType = newDBType // Update DB type
				slog.Info("Templates: Updated in global storage", "id", tmpl.ID, "new_name", newName, "new_db_type", newDBType)
				break
			}
		}
		customTemplatesMutex.Unlock()

		// Reload
		p.loadTemplates()

		slog.Info("Templates: Template updated successfully", "name", newName)
		dialog.ShowInformation("Success", "Template updated successfully", p.win)
	})
}

// onDeleteTemplate deletes a template.
func (p *TemplateManagementPage) onDeleteTemplate(tmpl templateInfo) {
	// Cannot delete built-in templates
	if tmpl.IsBuiltin {
		slog.Warn("Templates: Attempted to delete built-in template", "name", tmpl.Name)
		dialog.ShowError(
			fmt.Errorf("cannot delete built-in template '%s'", tmpl.Name),
			p.win,
		)
		return
	}

	dialog.ShowConfirm(
		"Delete Template",
		fmt.Sprintf("Delete custom template '%s'?", tmpl.Name),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			slog.Info("Templates: Deleting custom template", "name", tmpl.Name)

			// Delete from global storage
			customTemplatesMutex.Lock()
			for i, ct := range customTemplates {
				if ct.ID == tmpl.ID {
					customTemplates = append(customTemplates[:i], customTemplates[i+1:]...)
					break
				}
			}
			customTemplatesMutex.Unlock()

			// Reload
			p.loadTemplates()

			dialog.ShowInformation("Deleted", "Template deleted", p.win)
		},
		p.win,
	)
}

// onSetDefault sets a template as default for its database type.
func (p *TemplateManagementPage) onSetDefault(tmpl templateInfo, dbType string) {
	// Update the global defaultTemplateIDs map (works for both builtin and custom templates)
	defaultTemplateIDs[dbType] = tmpl.ID
	slog.Info("Templates: Default template updated", "db_type", dbType, "template_id", tmpl.ID, "template_name", tmpl.Name)

	// Update custom templates in global storage
	customTemplatesMutex.Lock()
	// Clear default flag for all templates of the same database type
	for i := range customTemplates {
		if customTemplates[i].DBType == dbType {
			customTemplates[i].IsDefault = false
		}
	}

	// Set the selected template as default (only for custom templates)
	for i := range customTemplates {
		if customTemplates[i].ID == tmpl.ID {
			customTemplates[i].IsDefault = true
			customTemplates[i].DBType = dbType // Ensure DB type is set
			break
		}
	}
	customTemplatesMutex.Unlock()

	// Reload UI (must release lock first to avoid deadlock)
	p.loadTemplates()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Default template for %s changed to: %s\n\n", dbType, tmpl.Name))
	sb.WriteString("This template will be auto-selected in Tasks page.")

	dialog.ShowInformation("Default Set", sb.String(), p.win)
}

// showTemplateDetails shows template details with all parameters.
func (p *TemplateManagementPage) showTemplateDetails(tmpl templateInfo) {
	var sb strings.Builder

	sb.WriteString("## ")
	sb.WriteString(tmpl.Name)
	if tmpl.IsDefault {
		sb.WriteString(" ⭐ (Default)")
	}
	sb.WriteString("\n\n")

	if tmpl.Description != "" {
		sb.WriteString("**Description:** ")
		sb.WriteString(tmpl.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("**Tool:** `")
	sb.WriteString(tmpl.Tool)
	sb.WriteString("`\n\n")
	sb.WriteString("**Database Type:** `")
	sb.WriteString(tmpl.DBType)
	sb.WriteString("`\n\n")

	sb.WriteString("**Type:** 📦 Built-in Template\n")
	sb.WriteString("**Actions:** Can be set as default\n\n")

	// Show parameters
	if tmpl.Parameters != nil {
		sb.WriteString("---\n\n")
		sb.WriteString("### Parameters\n\n")

		sb.WriteString("**General Parameters:**\n\n")
		sb.WriteString(fmt.Sprintf("- `--tables=%d` - Number of tables\n", tmpl.Parameters.Tables))
		sb.WriteString(fmt.Sprintf("- `--table-size=%d` - Rows per table\n", tmpl.Parameters.TableSize))

		sb.WriteString("\n**OLTP Test Parameters** (for reference, currently not used in execution):\n\n")
		sb.WriteString("The following OLTP parameters can be configured in the Add/Edit dialog,\n")
		sb.WriteString("but are currently not passed to sysbench. The benchmark uses sysbench defaults.\n\n")
		sb.WriteString("- `--db-ps-mode` - Prepared statement mode (disable/auto/no_ps)\n")
		sb.WriteString("- `--oltp-test-mode` - Test mode (complex/simple/nontrx/specific)\n")
		sb.WriteString("- `--oltp-point-selects` - Point select ratio\n")
		sb.WriteString("- `--oltp-simple-ranges` - Simple range ratio\n")
		sb.WriteString("- `--oltp-sum-ranges` - Sum range ratio\n")
		sb.WriteString("- `--oltp-order-ranges` - Order range ratio\n")
		sb.WriteString("- `--oltp-distinct-ranges` - Distinct range ratio\n")
		sb.WriteString("- `--oltp-index-updates` - Index update ratio\n")
		sb.WriteString("- `--oltp-non-index-updates` - Non-index update ratio\n")
		sb.WriteString("- `--oltp-delete-inserts` - Delete-insert ratio\n")

		sb.WriteString("\n**Note:** Additional parameters (threads, time, rate) are configured in the Tasks page when running the benchmark.\n")
	}

	// Show transaction weights for Oracle Swingbench templates
	if tmpl.Tool == "swingbench" && tmpl.DBType == "Oracle" {
		sb.WriteString("---\n\n")
		sb.WriteString("### Transaction Mix (Proportions)\n\n")

		// Load the actual template to get transaction weights
		weights := p.getTransactionWeights(tmpl.ID)
		if weights != nil {
			sb.WriteString("**Transaction Distribution:**\n\n")
			for name, weight := range weights {
				sb.WriteString(fmt.Sprintf("- %s：**%d**\n", name, weight))
			}
		}
	}

	content := widget.NewRichTextFromMarkdown(sb.String())

	dlg := dialog.NewCustomConfirm(
		"Template Details",
		"Close",
		"",
		content,
		func(bool) {},
		p.win,
	)
	dlg.Resize(fyne.NewSize(700, 600))
	dlg.Show()
}

// getTransactionWeights loads transaction weights from a template file.
func (p *TemplateManagementPage) getTransactionWeights(templateID string) map[string]int {
	// Mapping of transaction weights for each Oracle template
	weightMap := map[string]map[string]int{
		"swingbench-oracle-test": {
			"Customer_Registration": 10,
			"Update_Customer_Details": 10,
			"Browse_Products":       35,
			"Order_Products":        35,
			"Process_Orders":        5,
			"Browse_Orders":          5,
		},
		"swingbench-oracle-cpu-bound": {
			"Customer_Registration": 1,
			"Update_Customer_Details": 1,
			"Browse_Products":       85,
			"Order_Products":        5,
			"Process_Orders":        3,
			"Browse_Orders":          5,
		},
		"swingbench-oracle-disk-bound": {
			"Customer_Registration": 10,
			"Update_Customer_Details": 10,
			"Browse_Products":       35,
			"Order_Products":        35,
			"Process_Orders":        5,
			"Browse_Orders":          5,
		},
	}

	return weightMap[templateID]
}

// GetDefaultTemplate returns the default template for use in Tasks page.
func (p *TemplateManagementPage) GetDefaultTemplate() *templateInfo {
	if p.defaultIndex >= 0 && p.defaultIndex < len(p.templates) {
		return &p.templates[p.defaultIndex]
	}
	return nil
}

// =============================================================================
// Template Add/Edit Dialog
// =============================================================================

// templateDialog represents the template add/edit dialog.
type templateDialog struct {
	win                 fyne.Window
	onSuccess           func(*OLTPParameters, string, string) // Added dbType parameter
	isEditMode          bool
	originalName        string // For edit mode - original template name
	templateID          string // For edit mode - template ID
	dialog              *dialog.CustomDialog
	nameEntry           *widget.Entry
	dbTypeSelect        *widget.Select // Added database type selection
	formContainer       *fyne.Container // Container for dynamic form fields

	// Sysbench parameters
	tablesEntry         *widget.Entry
	tableSizeEntry      *widget.Entry
	dbPSModeEntry       *widget.Select
	oltpTestModeEntry   *widget.Select
	oltpPointSelects    *widget.Entry
	oltpSimpleRanges    *widget.Entry
	oltpSumRanges       *widget.Entry
	oltpOrderRanges     *widget.Entry
	oltpDistinctRanges  *widget.Entry
	oltpIndexUpdates    *widget.Entry
	oltpNonIndexUpdates *widget.Entry
	oltpDeleteInserts   *widget.Entry

	// Swingbench parameters (for Oracle)
	usersEntry          *widget.Entry
	timeEntry           *widget.Entry
	scaleEntry          *widget.Entry
	usernameEntry       *widget.Entry
	passwordEntry       *widget.Entry
	dbaUsernameEntry    *widget.Entry
	dbaPasswordEntry    *widget.Entry
	configFileEntry     *widget.Entry
	threadsEntry        *widget.Entry
}

// showTemplateDialog shows the template add/edit dialog.
func showTemplateDialog(win fyne.Window, title string, existingParams *OLTPParameters, existingName string, onSuccess func(*OLTPParameters, string, string)) {
	showTemplateDialogWithDBType(win, title, existingParams, existingName, "MySQL", onSuccess)
}

// showTemplateDialogWithDBType shows the template add/edit dialog with initial DB type.
func showTemplateDialogWithDBType(win fyne.Window, title string, existingParams *OLTPParameters, existingName string, initialDBType string, onSuccess func(*OLTPParameters, string, string)) {
	slog.Info("Templates: Showing template dialog", "title", title, "is_edit_mode", existingParams != nil, "existing_name", existingName, "initial_db_type", initialDBType)
	d := &templateDialog{
		win:          win,
		onSuccess:    onSuccess,
		isEditMode:   existingParams != nil,
		originalName: existingName, // Store original name for edit mode
	}

	// Default values
	defaultParams := &OLTPParameters{
		Tables:    10,
		TableSize: 10000,
	}

	if existingParams != nil {
		defaultParams = existingParams
	}

	// Default OLTP parameters (for display only - not currently used in execution)
	defaultDBPSMode := "disable"
	defaultOLTPTestMode := "complex"
	defaultOLTPPointSelects := 10
	defaultOLTPSimpleRanges := 1
	defaultOLTPSumRanges := 1
	defaultOLTPOrderRanges := 1
	defaultOLTPDistinctRanges := 1
	defaultOLTPIndexUpdates := 1
	defaultOLTPNonIndexUpdates := 1
	defaultOLTPDeleteInserts := 1

	// Default Swingbench parameters
	defaultUsers := 8
	defaultTime := 10
	defaultScale := 1
	defaultUsername := "soe"
	defaultDBAUsername := "sys as sysdba"
	defaultConfigFile := "/opt/benchtools/swingbench/configs/SOE_TEST.xml"
	defaultThreads := 32

	// Create common form fields
	d.nameEntry = widget.NewEntry()
	d.nameEntry.SetPlaceHolder("My Custom Template")
	if existingName != "" {
		d.nameEntry.SetText(existingName)
	}

	// Database type selection
	d.dbTypeSelect = widget.NewSelect([]string{"MySQL", "PostgreSQL", "Oracle", "SQL Server"}, nil)
	d.dbTypeSelect.SetSelected(initialDBType) // Use initial DB type

	// ============ Create Sysbench parameters ============
	d.tablesEntry = widget.NewEntry()
	d.tablesEntry.SetText(fmt.Sprintf("%d", defaultParams.Tables))

	d.tableSizeEntry = widget.NewEntry()
	d.tableSizeEntry.SetText(fmt.Sprintf("%d", defaultParams.TableSize))

	d.dbPSModeEntry = widget.NewSelect([]string{"disable", "auto", "no_ps"}, nil)
	d.dbPSModeEntry.SetSelected(defaultDBPSMode)

	d.oltpTestModeEntry = widget.NewSelect([]string{"complex", "simple", "nontrx", "specific"}, nil)
	d.oltpTestModeEntry.SetSelected(defaultOLTPTestMode)

	d.oltpPointSelects = widget.NewEntry()
	d.oltpPointSelects.SetText(fmt.Sprintf("%d", defaultOLTPPointSelects))

	d.oltpSimpleRanges = widget.NewEntry()
	d.oltpSimpleRanges.SetText(fmt.Sprintf("%d", defaultOLTPSimpleRanges))

	d.oltpSumRanges = widget.NewEntry()
	d.oltpSumRanges.SetText(fmt.Sprintf("%d", defaultOLTPSumRanges))

	d.oltpOrderRanges = widget.NewEntry()
	d.oltpOrderRanges.SetText(fmt.Sprintf("%d", defaultOLTPOrderRanges))

	d.oltpDistinctRanges = widget.NewEntry()
	d.oltpDistinctRanges.SetText(fmt.Sprintf("%d", defaultOLTPDistinctRanges))

	d.oltpIndexUpdates = widget.NewEntry()
	d.oltpIndexUpdates.SetText(fmt.Sprintf("%d", defaultOLTPIndexUpdates))

	d.oltpNonIndexUpdates = widget.NewEntry()
	d.oltpNonIndexUpdates.SetText(fmt.Sprintf("%d", defaultOLTPNonIndexUpdates))

	d.oltpDeleteInserts = widget.NewEntry()
	d.oltpDeleteInserts.SetText(fmt.Sprintf("%d", defaultOLTPDeleteInserts))

	// ============ Create Swingbench parameters ============
	d.usersEntry = widget.NewEntry()
	d.usersEntry.SetText(fmt.Sprintf("%d", defaultUsers))

	d.timeEntry = widget.NewEntry()
	d.timeEntry.SetText(fmt.Sprintf("%d", defaultTime))

	d.scaleEntry = widget.NewEntry()
	d.scaleEntry.SetText(fmt.Sprintf("%d", defaultScale))

	d.usernameEntry = widget.NewEntry()
	d.usernameEntry.SetText(defaultUsername)

	d.passwordEntry = widget.NewEntry()
	d.passwordEntry.SetPlaceHolder("Schema password")

	d.dbaUsernameEntry = widget.NewEntry()
	d.dbaUsernameEntry.SetText(defaultDBAUsername)

	d.dbaPasswordEntry = widget.NewEntry()
	d.dbaPasswordEntry.SetPlaceHolder("DBA password")

	d.configFileEntry = widget.NewEntry()
	d.configFileEntry.SetText(defaultConfigFile)

	d.threadsEntry = widget.NewEntry()
	d.threadsEntry.SetText(fmt.Sprintf("%d", defaultThreads))

	// ============ Create dynamic form container ============
	d.formContainer = container.NewVBox()

	// Function to update form fields based on database type
	updateFormFields := func(dbType string) {
		slog.Info("Templates: Updating form fields for DB type", "db_type", dbType)
		d.formContainer.Objects = nil

		if dbType == "Oracle" {
			// Show message: Oracle custom templates not supported yet
			msgLabel := widget.NewLabel("Oracle templates use Swingbench with different parameters.\n\nCurrently, only built-in Oracle templates are supported.\n\nPlease use the built-in Oracle templates:\n- Test (Swingbench)\n- CPU Bound (Swingbench)\n- Disk Bound (Swingbench)")
			d.formContainer.Add(container.NewVBox(msgLabel))
		} else {
			// Show Sysbench parameters
			formItems := []*widget.FormItem{
				widget.NewFormItem("Tables (N)", d.tablesEntry),
				widget.NewFormItem("Table Size (N)", d.tableSizeEntry),
				widget.NewFormItem("DB PS Mode", d.dbPSModeEntry),
				widget.NewFormItem("OLTP Test Mode", d.oltpTestModeEntry),
				widget.NewFormItem("Point Selects", d.oltpPointSelects),
				widget.NewFormItem("Simple Ranges", d.oltpSimpleRanges),
				widget.NewFormItem("Sum Ranges", d.oltpSumRanges),
				widget.NewFormItem("Order Ranges", d.oltpOrderRanges),
				widget.NewFormItem("Distinct Ranges", d.oltpDistinctRanges),
				widget.NewFormItem("Index Updates", d.oltpIndexUpdates),
				widget.NewFormItem("Non-Index Updates", d.oltpNonIndexUpdates),
				widget.NewFormItem("Delete Inserts", d.oltpDeleteInserts),
			}
			form := widget.NewForm(formItems...)
			d.formContainer.Add(form)
		}
		d.formContainer.Refresh()
	}

	// Set up callback for database type change
	d.dbTypeSelect.OnChanged = func(dbType string) {
		slog.Info("Templates: DB type changed", "db_type", dbType)
		updateFormFields(dbType)
	}

	// Initialize form with initial DB type
	updateFormFields(initialDBType)

	// Create buttons
	btnSave := widget.NewButton("Save", func() {
		if d.onSave() {
			d.dialog.Hide()
		}
		// If onSave returns false, dialog stays open
	})
	btnSave.Importance = widget.HighImportance

	btnCancel := widget.NewButton("Cancel", func() {
		// Will be set to close dialog after dialog is created
	})

	buttonContainer := container.NewHBox(btnSave, btnCancel)

	// Create static form items (Database Type and Template Name)
	staticForm := widget.NewForm(
		widget.NewFormItem("Database Type", d.dbTypeSelect),
		widget.NewFormItem("Template Name", d.nameEntry),
	)

	// Create dialog content with buttons at bottom
	content := container.NewVBox(staticForm, d.formContainer, widget.NewSeparator(), buttonContainer)

	// Create custom dialog without buttons
	dlg := dialog.NewCustomWithoutButtons(title, content, win)
	dlg.Resize(fyne.NewSize(500, 700))
	d.dialog = dlg

	// Update Cancel button to close dialog
	btnCancel.OnTapped = func() {
		dlg.Hide()
	}

	dlg.Show()
}

// onSave handles the save button click.
// Returns true if save was successful (dialog should close), false otherwise (dialog stays open).
func (d *templateDialog) onSave() bool {
	slog.Info("Templates: Save button clicked in dialog")

	dbType := d.dbTypeSelect.Selected

	// Check if Oracle is selected (not supported for custom templates yet)
	if dbType == "Oracle" {
		slog.Warn("Templates: Cannot create custom Oracle templates")
		dialog.ShowError(fmt.Errorf("custom Oracle templates are not supported yet\n\nPlease use the built-in Oracle templates"), d.win)
		return false
	}

	// Parse and validate parameters
	name := strings.TrimSpace(d.nameEntry.Text)
	if name == "" {
		slog.Warn("Templates: Template name is empty")
		dialog.ShowError(fmt.Errorf("template name is required"), d.win)
		return false
	}

	// Check for duplicate names
	customTemplatesMutex.RLock()
	for _, tmpl := range customTemplates {
		// Skip self in edit mode if name hasn't changed
		if d.isEditMode && tmpl.Name == d.originalName && name == d.originalName {
			continue
		}
		// Check for duplicate
		if tmpl.Name == name {
			customTemplatesMutex.RUnlock()
			slog.Warn("Templates: Template name already exists", "name", name)
			dialog.ShowError(fmt.Errorf("template name '%s' already exists", name), d.win)
			return false
		}
	}
	customTemplatesMutex.RUnlock()

	// Also check built-in templates
	if name == "OLTP Read-Write (Sysbench)" {
		slog.Warn("Templates: Template name conflicts with built-in template", "name", name)
		dialog.ShowError(fmt.Errorf("template name '%s' conflicts with built-in template", name), d.win)
		return false
	}

	slog.Info("Templates: Template validated", "name", name)

	// Parse numeric values (simplified - no strict validation)
	tables := parseIntOrDefault(d.tablesEntry.Text, 10)
	tableSize := parseIntOrDefault(d.tableSizeEntry.Text, 10000)

	params := &OLTPParameters{
		Tables:    tables,
		TableSize: tableSize,
	}

	slog.Info("Templates: DB Type from selector", "db_type", dbType, "selected", d.dbTypeSelect.Selected, "options", d.dbTypeSelect.Options)

	if d.onSuccess != nil {
		d.onSuccess(params, name, dbType)
	}

	return true
}

// parseIntOrDefault parses an integer or returns default value.
func parseIntOrDefault(s string, defaultValue int) int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		return defaultValue
	}
	return val
}
