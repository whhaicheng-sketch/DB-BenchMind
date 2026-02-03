// Package pages provides GUI pages for DB-BenchMind.
// Task Configuration and Monitor Page (Combined Interface with Template support).
package pages

import (
	"context"
	"fmt"
	"image/color"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/google/uuid"
	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// minSizeWidget is a custom widget that wraps a child and enforces a minimum size.
type minSizeWidget struct {
	widget.BaseWidget
	child   fyne.CanvasObject
	minSize fyne.Size
}

func newMinSizeWidget(child fyne.CanvasObject, minHeight float32) *minSizeWidget {
	m := &minSizeWidget{
		child:   child,
		minSize: fyne.NewSize(0, minHeight),
	}
	m.ExtendBaseWidget(m)
	return m
}

func (m *minSizeWidget) CreateRenderer() fyne.WidgetRenderer {
	return &minSizeWidgetRenderer{
		widget: m,
		rect:   canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0}),
	}
}

func (m *minSizeWidget) MinSize() fyne.Size {
	childMin := m.child.MinSize()
	return fyne.NewSize(
		childMin.Width,
		fyne.Max(childMin.Height, m.minSize.Height),
	)
}

type minSizeWidgetRenderer struct {
	widget *minSizeWidget
	rect   *canvas.Rectangle
}

func (r *minSizeWidgetRenderer) Layout(size fyne.Size) {
	r.widget.child.Resize(size)
	r.rect.Resize(size)
	r.rect.Move(fyne.Position{})
}

func (r *minSizeWidgetRenderer) MinSize() fyne.Size {
	return r.widget.MinSize()
}

func (r *minSizeWidgetRenderer) Refresh() {
	r.rect.Refresh()
}

func (r *minSizeWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.widget.child, r.rect}
}

func (r *minSizeWidgetRenderer) Destroy() {}

// TaskMonitorPage provides combined task configuration and real-time monitoring GUI.
type TaskMonitorPage struct {
	win          fyne.Window
	isRunning    bool
	currentRunID string // Current benchmark run ID
	// Use cases
	connUC      *usecase.ConnectionUseCase
	benchmarkUC *usecase.BenchmarkUseCase
	templateUC  *usecase.TemplateUseCase
	historyUC   *usecase.HistoryUseCase
	// Task configuration widgets
	connSelect     *widget.Select
	templateSelect *widget.Select
	// General parameters
	threadsEntry  *widget.Entry
	durationEntry *widget.Entry
	dbNameEntry   *widget.Entry
	// Monitor widgets
	statusLabel     *widget.Label
	tpsLabel        *widget.Label
	qpsLabel        *widget.Label
	latencyP95Label *widget.Label
	errorsLabel     *widget.Label
	threadsLabel    *widget.Label
	progressBar     *widget.ProgressBar
	// Real-time log for sysbench output
	logEntry     *widget.Entry
	maxLogLines  int
	lastLogCount int             // Track number of samples already added to log
	addedSeconds map[string]bool // Track which seconds have been added to prevent duplicates
	// Control buttons
	btnPrepare *widget.Button
	btnRun     *widget.Button
	btnCleanup *widget.Button
	btnStop    *widget.Button
	// Template data
	templates []templateInfo
	// Connection data by ID
	connections map[string]connection.Connection // ID -> Connection
}

// NewTaskMonitorPage creates a new combined task configuration and monitor page.
func NewTaskMonitorPage(win fyne.Window) fyne.CanvasObject {
	return NewTaskMonitorPageWithUC(win, nil, nil, nil, nil)
}

// NewTaskMonitorPageWithUC creates a new combined task configuration and monitor page with use cases.
func NewTaskMonitorPageWithUC(win fyne.Window, connUC *usecase.ConnectionUseCase, benchmarkUC *usecase.BenchmarkUseCase, templateUC *usecase.TemplateUseCase, historyUC *usecase.HistoryUseCase) fyne.CanvasObject {
	slog.Info("Tasks: NewTaskMonitorPageWithUC called", "has_connUC", connUC != nil, "has_benchmarkUC", benchmarkUC != nil, "has_templateUC", templateUC != nil, "has_historyUC", historyUC != nil)
	page := &TaskMonitorPage{
		win:          win,
		isRunning:    false,
		currentRunID: "",
		connUC:       connUC,
		benchmarkUC:  benchmarkUC,
		templateUC:   templateUC,
		historyUC:    historyUC,
		connections:  make(map[string]connection.Connection),
	}

	// Create connection selector
	page.connSelect = widget.NewSelect([]string{}, nil)
	page.connSelect.OnChanged = func(s string) {
		slog.Info("Tasks: Connection selected", "connection", s)
		page.onConnectionChanged()
	}

	// Load connections from database
	if page.connUC != nil {
		page.loadConnections()
	}

	// Initialize template selector (will be populated when connection is selected)
	page.templateSelect = widget.NewSelect([]string{}, func(selected string) {
		if selected != "" {
			slog.Info("Tasks: Template changed", "template", selected)
		} else {
			slog.Info("Tasks: Template cleared")
		}
	})

	// Create general parameter entries
	page.threadsEntry = widget.NewEntry()
	page.threadsEntry.SetText("1")

	page.durationEntry = widget.NewEntry()
	page.durationEntry.SetText("60")

	page.dbNameEntry = widget.NewEntry()
	page.dbNameEntry.SetText("sbtest")

	// Create refresh button for templates
	btnRefreshTemplate := widget.NewButton("🔄 Refresh Templates", func() {
		slog.Info("Tasks: Refresh templates button clicked")
		if page.connSelect.Selected != "" {
			// Save current selection
			currentSelection := page.templateSelect.Selected
			slog.Info("Tasks: Reloading templates for connection", "connection", page.connSelect.Selected, "current_selection", currentSelection)
			page.onConnectionChanged()

			// Try to restore selection
			if currentSelection != "" {
				// Check if the previously selected template still exists
				selectionExists := false
				for _, opt := range page.templateSelect.Options {
					if opt == currentSelection {
						selectionExists = true
						break
					}
				}
				if selectionExists {
					page.templateSelect.SetSelected(currentSelection)
					slog.Info("Tasks: Restored template selection", "selection", currentSelection)
				} else {
					slog.Info("Tasks: Previous selection no longer available", "previous_selection", currentSelection)
				}
			}
		} else {
			slog.Info("Tasks: No connection selected, cannot refresh templates")
		}
	})

	// Template selector with refresh button
	// Use Border layout to stretch templateSelect to fill available space before the button
	templateRow := container.NewBorder(nil, nil, nil, btnRefreshTemplate, page.templateSelect)

	// Create simplified form with general parameters
	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Connection", page.connSelect),
			widget.NewFormItem("Template", templateRow),
			widget.NewFormItem("Threads", page.threadsEntry),
			widget.NewFormItem("Duration (seconds)", page.durationEntry),
			widget.NewFormItem("Database Name", page.dbNameEntry),
		},
	}

	// Create monitor widgets
	page.statusLabel = widget.NewLabel("Idle")
	page.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	page.tpsLabel = widget.NewLabel("--")
	page.qpsLabel = widget.NewLabel("--")
	page.latencyP95Label = widget.NewLabel("--")
	page.errorsLabel = widget.NewLabel("0.00")
	page.threadsLabel = widget.NewLabel("--")

	page.progressBar = widget.NewProgressBar()
	page.progressBar.SetValue(0)

	// Initialize log entry for sysbench output
	page.maxLogLines = 60 // Keep max 60 lines history
	page.logEntry = widget.NewMultiLineEntry()
	page.logEntry.Disable()
	page.logEntry.SetText("Waiting for benchmark data...\n")

	// Create control buttons for each phase
	page.btnPrepare = widget.NewButton("📦 Prepare", func() {
		page.onPreparePhase()
	})
	page.btnPrepare.Importance = widget.MediumImportance

	page.btnRun = widget.NewButton("▶ Run", func() {
		page.onRunPhase()
	})
	page.btnRun.Importance = widget.HighImportance

	page.btnCleanup = widget.NewButton("🧹 Cleanup", func() {
		page.onCleanupPhase()
	})
	page.btnCleanup.Importance = widget.MediumImportance

	page.btnStop = widget.NewButton("■ Stop", func() {
		page.onStopTask()
	})
	page.btnStop.Disable() // Disabled initially

	// Toolbar with Prepare, Run, Cleanup and Stop buttons
	toolbar := container.NewHBox(page.btnPrepare, page.btnRun, page.btnCleanup, page.btnStop)

	// Task configuration card (top section)
	taskCard := widget.NewCard("Task Configuration", "", container.NewPadded(form))

	// Monitor metrics card (middle section)
	metricsGrid := container.NewGridWithColumns(4,
		widget.NewLabel("TPS:"),
		page.tpsLabel,
		widget.NewLabel("QPS:"),
		page.qpsLabel,
		widget.NewLabel("95% Latency:"),
		page.latencyP95Label,
		widget.NewLabel("Threads:"),
		page.threadsLabel,
		widget.NewLabel("Errors/s:"),
		page.errorsLabel,
	)

	statusRow := container.NewHBox(page.statusLabel)

	// Create a VBox for metrics and progress (top section of monitor)
	topSection := container.NewVBox(
		statusRow,
		widget.NewSeparator(),
		metricsGrid,
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("Progress:"),
			page.progressBar,
		),
		widget.NewSeparator(),
		widget.NewLabel("Real-time Output:"),
	)

	// Wrap logEntry in custom widget that enforces minimum height of 240px (10 lines)
	logWrapper := newMinSizeWidget(page.logEntry, 240)

	// Use Border: top=topSection, center=logWrapper
	// The center object in Border fills all available space
	logContainer := container.NewBorder(topSection, nil, nil, nil, logWrapper)

	monitorCard := widget.NewCard("Real-time Monitor", "", logContainer)

	monitorToolbar := container.NewHBox(page.btnStop)

	// Add stop button to monitor card
	monitorToolbar.Objects = []fyne.CanvasObject{page.btnStop}

	// Main layout: Task on top, Monitor in middle
	topContent := container.NewVBox(
		taskCard,
		widget.NewSeparator(),
		toolbar,
		widget.NewSeparator(),
		monitorCard,
	)

	return topContent
}

// loadConnections loads connections from the database.
func (p *TaskMonitorPage) loadConnections() {
	if p.connUC == nil {
		slog.Warn("Tasks: ConnectionUseCase not available, cannot load connections")
		return
	}

	ctx := context.Background()
	conns, err := p.connUC.ListConnections(ctx)
	if err != nil {
		slog.Error("Tasks: Failed to load connections", "err", err)
		dialog.ShowError(fmt.Errorf("failed to load connections: %w", err), p.win)
		return
	}

	// Clear and populate connections map
	p.connections = make(map[string]connection.Connection)
	connectionNames := make([]string, 0, len(conns))

	for _, conn := range conns {
		p.connections[conn.GetName()] = conn
		connectionNames = append(connectionNames, conn.GetName())
	}

	p.connSelect.Options = connectionNames

	slog.Info("Tasks: Connections loaded", "count", len(connectionNames))
}

// onConnectionChanged handles connection selection changes.
func (p *TaskMonitorPage) onConnectionChanged() {
	selectedName := p.connSelect.Selected
	if selectedName == "" {
		// Clear template selector
		p.templateSelect.Options = []string{}
		p.templateSelect.SetSelected("")
		slog.Info("Tasks: Connection cleared, templates reset")
		return
	}

	// Get selected connection
	conn, ok := p.connections[selectedName]
	if !ok {
		slog.Warn("Tasks: Connection not found in map", "connection", selectedName)
		return
	}

	// Get database type from connection
	dbType := string(conn.GetType())
	// Normalize DB type (mysql -> MySQL)
	normalizedDBType := normalizeDBType(dbType)

	slog.Info("Tasks: Connection changed", "connection", selectedName, "db_type", normalizedDBType)

	// Load templates for this database type
	p.loadTemplatesForDBType(normalizedDBType)
}

// loadTemplatesForDBType loads templates for a specific database type.
func (p *TaskMonitorPage) loadTemplatesForDBType(dbType string) {
	slog.Info("Tasks: loadTemplatesForDBType called", "db_type", dbType)

	// Load all templates (built-in + custom)
	templates := p.loadTemplatesData()
	slog.Info("Tasks: All templates loaded", "total", len(templates))

	// Filter templates by DB type
	var filteredTemplates []templateInfo
	var defaultTemplate *templateInfo

	for i := range templates {
		slog.Info("Tasks: Checking template", "index", i, "name", templates[i].Name, "template_db_type", templates[i].DBType, "target_db_type", dbType, "match", templates[i].DBType == dbType)
		if templates[i].DBType == dbType {
			filteredTemplates = append(filteredTemplates, templates[i])
			if templates[i].IsDefault {
				defaultTemplate = &templates[i]
			}
		}
	}

	slog.Info("Tasks: Filtered templates", "db_type", dbType, "count", len(filteredTemplates))

	// Populate template selector
	templateNames := make([]string, len(filteredTemplates))
	for i, tmpl := range filteredTemplates {
		templateNames[i] = tmpl.Name
		slog.Info("Tasks: Adding to selector", "index", i, "name", tmpl.Name, "is_default", tmpl.IsDefault)
	}

	p.templateSelect.Options = templateNames
	p.templates = filteredTemplates

	// Select default template
	if defaultTemplate != nil {
		p.templateSelect.SetSelected(defaultTemplate.Name)
		slog.Info("Tasks: Default template selected", "template", defaultTemplate.Name, "db_type", dbType)
	} else if len(templateNames) > 0 {
		p.templateSelect.SetSelected(templateNames[0])
		slog.Info("Tasks: First template selected (no default found)", "template", templateNames[0])
	}

	slog.Info("Tasks: Template selector updated", "db_type", dbType, "options_count", len(p.templateSelect.Options))
}

// loadTemplatesData loads and returns template information (shared with template_page).
func (p *TaskMonitorPage) loadTemplatesData() []templateInfo {
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

	builtinTemplates := []templateInfo{
		// MySQL templates
		{
			ID:          "sysbench-mysql-test",
			Name:        "Test (Sysbench)",
			Description: "Lightweight test template for quick MySQL testing (10 tables, 10K rows each)",
			Tool:        "sysbench",
			DBType:      "MySQL",
			IsBuiltin:   true,
			IsDefault:   true,
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
			IsDefault:   true,
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
	}

	// Load custom templates from global storage
	customTemplatesMutex.RLock()
	customCount := len(customTemplates)
	copiedTemplates := make([]templateInfo, customCount)
	copy(copiedTemplates, customTemplates)
	customTemplatesMutex.RUnlock()

	slog.Info("Tasks: Loading custom templates from global storage", "count", customCount)
	for i, ct := range copiedTemplates {
		slog.Info("Tasks: Custom template", "index", i, "name", ct.Name, "db_type", ct.DBType, "is_default", ct.IsDefault)
	}

	// Check which DB types have custom default templates
	dbTypesWithCustomDefault := make(map[string]bool)
	for _, ct := range customTemplates {
		if ct.IsDefault {
			dbTypesWithCustomDefault[ct.DBType] = true
		}
	}

	// Adjust built-in templates' default flag
	for i := range builtinTemplates {
		dbType := builtinTemplates[i].DBType
		if dbTypesWithCustomDefault[dbType] {
			builtinTemplates[i].IsDefault = false
		} else {
			builtinTemplates[i].IsDefault = true
		}
	}

	// Combine built-in and custom templates
	allTemplates := append(builtinTemplates, customTemplates...)
	slog.Info("Tasks: Total templates loaded", "builtin", len(builtinTemplates), "custom", len(customTemplates), "total", len(allTemplates))

	// Sync custom templates to repository if templateUC is available (run in background to avoid UI blocking)
	if p.templateUC != nil && len(customTemplates) > 0 {
		go p.syncCustomTemplatesToRepository(customTemplates)
	}

	return allTemplates
}

// syncCustomTemplatesToRepository saves custom templates to the TemplateRepository.
// This ensures that custom templates created in the GUI can be used by BenchmarkUseCase.
func (p *TaskMonitorPage) syncCustomTemplatesToRepository(customTemplates []templateInfo) {
	ctx := context.Background()

	for _, ct := range customTemplates {
		// Check if template already exists in repository
		existing, err := p.templateUC.GetTemplate(ctx, ct.ID)
		if err == nil && existing != nil {
			// Template already exists, skip
			slog.Debug("Tasks: Template already in repository", "id", ct.ID, "name", ct.Name)
			continue
		}

		// Create template.Template from templateInfo
		// For custom templates, we'll create a basic sysbench template
		tmpl := &domaintemplate.Template{
			ID:            ct.ID,
			Name:          ct.Name,
			Description:   ct.Description,
			Tool:          ct.Tool,
			DatabaseTypes: []string{strings.ToLower(ct.DBType)},
			Version:       "1.0.0",
			Parameters:    make(map[string]domaintemplate.Parameter),
			CommandTemplate: domaintemplate.CommandTemplate{
				Prepare: "sysbench {db_type} --tables={tables} --table-size={table_size} {connection_string} prepare",
				Run:     "sysbench {db_type} --threads={threads} --time={time} --tables={tables} --report-interval=1 {rate_arg} {connection_string} run",
				Cleanup: "sysbench {db_type} --tables={tables} {connection_string} cleanup",
			},
			OutputParser: domaintemplate.OutputParser{
				Type: domaintemplate.ParserTypeRegex,
				Patterns: map[string]string{
					"tps":             `transactions:\s*\(\s*(\d+\.?\d*)\s*per sec\.`,
					"latency_avg":     `latency:\s*\(ms\).*?avg=\s*(\d+\.?\d*)`,
					"latency_min":     `latency:\s*\(ms\).*?min=\s*(\d+\.?\d*)`,
					"latency_max":     `latency:\s*\(ms\).*?max=\s*(\d+\.?\d*)`,
					"95th_percentile": `latency:\s*\(ms\).*?95th percentile=\s*(\d+\.?\d*)`,
				},
			},
		}

		// Add parameters
		if ct.Parameters != nil {
			tmpl.Parameters["threads"] = domaintemplate.Parameter{
				Type:    domaintemplate.ParameterTypeInteger,
				Label:   "Thread count",
				Default: 1,
				Min:     intPtr(1),
				Max:     intPtr(1024),
			}
			tmpl.Parameters["time"] = domaintemplate.Parameter{
				Type:    domaintemplate.ParameterTypeInteger,
				Label:   "Runtime (seconds)",
				Default: 60,
				Min:     intPtr(10),
				Max:     intPtr(86400),
			}
			tmpl.Parameters["tables"] = domaintemplate.Parameter{
				Type:    domaintemplate.ParameterTypeInteger,
				Label:   "Number of tables",
				Default: ct.Parameters.Tables,
				Min:     intPtr(1),
				Max:     intPtr(1000),
			}
			tmpl.Parameters["table_size"] = domaintemplate.Parameter{
				Type:    domaintemplate.ParameterTypeInteger,
				Label:   "Rows per table",
				Default: ct.Parameters.TableSize,
				Min:     intPtr(1000),
				Max:     intPtr(100000000),
			}
			tmpl.Parameters["rate"] = domaintemplate.Parameter{
				Type:    domaintemplate.ParameterTypeInteger,
				Label:   "Transaction rate (0 = unlimited)",
				Default: 0,
				Min:     intPtr(0),
				Max:     intPtr(100000),
			}
		}

		// Save to repository
		if err := p.templateUC.CreateTemplate(ctx, tmpl); err != nil {
			slog.Error("Tasks: Failed to save custom template to repository", "id", ct.ID, "name", ct.Name, "error", err)
		} else {
			slog.Info("Tasks: Saved custom template to repository", "id", ct.ID, "name", ct.Name)
		}
	}
}

// intPtr returns a pointer to an int.
func intPtr(i int) *int {
	return &i
}

// onRunTask starts the benchmark task.
// onPreparePhase executes the prepare phase.
func (p *TaskMonitorPage) onPreparePhase() {
	slog.Info("Tasks: onPreparePhase called")
	p.validateAndExecutePhase("prepare")
}

// onRunPhase executes the run phase.
func (p *TaskMonitorPage) onRunPhase() {
	slog.Info("Tasks: onRunPhase called")
	p.validateAndExecutePhase("run")
}

// onCleanupPhase executes the cleanup phase.
func (p *TaskMonitorPage) onCleanupPhase() {
	slog.Info("Tasks: onCleanupPhase called")
	p.validateAndExecutePhase("cleanup")
}

// validateAndExecutePhase validates inputs and executes a specific phase.
func (p *TaskMonitorPage) validateAndExecutePhase(phase string) {
	// Validate
	if p.connSelect.Selected == "" {
		slog.Warn("Tasks: No connection selected")
		dialog.ShowError(fmt.Errorf("please select a connection"), p.win)
		return
	}

	if p.templateSelect.Selected == "" {
		slog.Warn("Tasks: No template selected")
		dialog.ShowError(fmt.Errorf("please select a template"), p.win)
		return
	}

	// ⭐ 关键改进：在执行前先测试数据库连接（仅失败时弹窗）
	if p.connUC != nil {
		// Get connection object
		connName := p.connSelect.Selected
		conn, ok := p.connections[connName]
		if !ok {
			slog.Error("Tasks: Connection not found", "name", connName)
			dialog.ShowError(fmt.Errorf("connection not found: %s", connName), p.win)
			return
		}

		slog.Info("Tasks: Testing connection before benchmark execution", "connection", connName, "connection_id", conn.GetID())

		// Test connection（静默测试，不弹窗）
		testResult, err := p.connUC.TestConnection(context.Background(), conn.GetID())
		if err != nil {
			slog.Error("Tasks: Connection test failed", "connection", connName, "error", err)
			dialog.ShowError(fmt.Errorf("connection test failed for %s: %w\n\nTask execution cancelled.", connName, err), p.win)
			return
		}

		if !testResult.Success {
			slog.Error("Tasks: Connection test unsuccessful", "connection", connName, "error", testResult.Error)
			dialog.ShowError(fmt.Errorf("connection test failed for %s:\n%s\n\nPlease check your connection settings and database availability.\n\nTask execution cancelled.", connName, testResult.Error), p.win)
			return
		}

		// 成功时只记录日志，不弹窗
		slog.Info("Tasks: Connection test successful", "connection", connName, "latency_ms", testResult.LatencyMs, "db_version", testResult.DatabaseVersion)
	}

	slog.Info("Tasks: Building benchmark task", "connection", p.connSelect.Selected, "template", p.templateSelect.Selected, "phase", phase)
	// Build benchmark task from UI inputs
	task, err := p.buildBenchmarkTask()
	if err != nil {
		slog.Error("Tasks: Failed to build task", "error", err)
		dialog.ShowError(fmt.Errorf("failed to build task: %w", err), p.win)
		return
	}

	slog.Info("Tasks: Task built successfully", "task_id", task.ID, "connection_id", task.ConnectionID, "template_id", task.TemplateID)
	// Check if BenchmarkUseCase is available
	if p.benchmarkUC == nil {
		slog.Error("Tasks: BenchmarkUseCase is nil")
		dialog.ShowError(fmt.Errorf("benchmark use case not available - please check application configuration"), p.win)
		return
	}

	// Execute the specific phase
	p.startBenchmarkPhase(task, phase)
}

// onRunTask is deprecated - use onPreparePhase, onRunPhase, or onCleanupPhase instead.
func (p *TaskMonitorPage) onRunTask() {
	slog.Info("Tasks: onRunTask called (deprecated, using executePhase instead)")
	p.validateAndExecutePhase("run")
}

// startSimulatedBenchmark starts a simulated benchmark (for debugging).
func (p *TaskMonitorPage) startSimulatedBenchmark(task *execution.BenchmarkTask) {
	slog.Info("Tasks: Starting SIMULATED benchmark (debug mode)", "task_id", task.ID)

	// Lock task form during execution
	p.setTaskFormEnabled(false)

	// Start monitoring
	p.isRunning = true
	p.statusLabel.SetText("Status: Running (Simulated)")
	p.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	p.btnRun.Disable()
	p.btnStop.Enable()

	// Get parameters
	threads, _ := task.Parameters["threads"].(int)
	duration, _ := task.Parameters["time"].(int)
	rateLimit, _ := task.Parameters["rate"].(int)

	// Start simulated execution
	go p.simulateExecution(threads, duration, rateLimit)
}

// simulateExecution simulates running a benchmark task.
func (p *TaskMonitorPage) simulateExecution(threads, duration, rateLimit int) {
	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		// Check if stopped at the beginning of each iteration
		if !p.isRunning {
			slog.Info("Tasks: SimulateExecution detected stop signal, exiting")
			return
		}

		select {
		case <-ticker.C:
			// Check again before processing
			if !p.isRunning {
				slog.Info("Tasks: SimulateExecution detected stop signal in ticker, exiting")
				return
			}

			elapsed := int(time.Since(startTime).Seconds())
			if elapsed >= duration {
				break
			}

			progress := float64(elapsed) / float64(duration)
			p.progressBar.SetValue(progress)

			// Simulate metrics
			tps := 1000 + int(progress*500)
			qps := tps * 20 // Simulate QPS as 20x TPS
			if qps < 1 {
				qps = 1
			}
			errors := int(progress * 2)

			p.tpsLabel.SetText(fmt.Sprintf("%d", tps))
			p.qpsLabel.SetText(fmt.Sprintf("%d", qps))
			p.errorsLabel.SetText(fmt.Sprintf("%d", errors))
		}
	}

	// Task completed
	if p.isRunning {
		p.isRunning = false
		p.statusLabel.SetText("Status: Completed (Simulated)")
		p.progressBar.SetValue(1.0)

		p.btnRun.Enable()
		p.btnStop.Disable()
		p.setTaskFormEnabled(true)

		slog.Info("Tasks: Simulated benchmark completed")
	}
}

// buildBenchmarkTask creates a BenchmarkTask from UI inputs.
func (p *TaskMonitorPage) buildBenchmarkTask() (*execution.BenchmarkTask, error) {
	// Get selected connection
	connName := p.connSelect.Selected
	conn, ok := p.connections[connName]
	if !ok {
		return nil, fmt.Errorf("connection not found: %s", connName)
	}

	// Parse and validate general parameters
	threads, err := strconv.Atoi(strings.TrimSpace(p.threadsEntry.Text))
	if err != nil || threads < 1 {
		return nil, fmt.Errorf("invalid threads value (must be >= 1)")
	}

	duration, err := strconv.Atoi(strings.TrimSpace(p.durationEntry.Text))
	if err != nil || duration <= 0 {
		return nil, fmt.Errorf("invalid duration value")
	}

	dbName := strings.TrimSpace(p.dbNameEntry.Text)

	// Get OLTP parameters and template ID from selected template
	var tables, tableSize int
	var templateID string
	for _, tmpl := range p.templates {
		if tmpl.Name == p.templateSelect.Selected {
			templateID = tmpl.ID
			if tmpl.Parameters != nil {
				tables = tmpl.Parameters.Tables
				tableSize = tmpl.Parameters.TableSize
			}
			break
		}
	}
	// Default values if no template selected
	if tables == 0 {
		tables = 10
	}
	if tableSize == 0 {
		tableSize = 10000
	}
	// Default template ID
	if templateID == "" {
		templateID = "sysbench-oltp-read-write"
	}

	// Build parameters map for sysbench
	parameters := map[string]interface{}{
		"threads":    threads,
		"time":       duration,
		"tables":     tables,
		"table_size": tableSize,
		"db_name":    dbName,
	}

	// Build task options
	options := execution.TaskOptions{
		SkipPrepare:    false,
		SkipCleanup:    false,
		WarmupTime:     0,
		SampleInterval: 10 * time.Second, // Default 10 seconds
		DryRun:         false,            // Set to true for testing without actually running
		PrepareTimeout: 30 * time.Minute,
		// Set timeout to 2x duration as a safety net to prevent hangs
		// Sysbench will control its own execution time via --time parameter
		// We should wait for it to complete naturally, not force kill it
		RunTimeout: time.Duration(duration*2) * time.Second,
	}

	// Create task
	task := &execution.BenchmarkTask{
		ID:           uuid.New().String(),
		Name:         fmt.Sprintf("%s Benchmark", connName),
		ConnectionID: conn.GetID(),
		TemplateID:   templateID,
		Parameters:   parameters,
		Options:      options,
		Tags:         []string{"gui", string(conn.GetType())},
		CreatedAt:    time.Now(),
	}

	slog.Info("Tasks: Built benchmark task",
		"task_id", task.ID,
		"connection_id", task.ConnectionID,
		"threads", threads,
		"duration", duration,
		"db_name", dbName)

	return task, nil
}

// startBenchmarkPhase starts a specific benchmark phase (prepare/run/cleanup).
func (p *TaskMonitorPage) startBenchmarkPhase(task *execution.BenchmarkTask, phase string) {
	ctx := context.Background()

	// Configure task options based on phase
	switch phase {
	case "prepare":
		task.Options.SkipPrepare = false
		task.Options.SkipCleanup = true
		task.Options.WarmupTime = 0
		// Set a very short run time to avoid running
		duration, _ := task.Parameters["time"].(int)
		task.Parameters["time"] = 0                  // Don't run
		task.Parameters["_original_time"] = duration // Save original

	case "run":
		task.Options.SkipPrepare = true
		task.Options.SkipCleanup = true
		// Restore original duration if saved
		if originalTime, ok := task.Parameters["_original_time"].(int); ok {
			task.Parameters["time"] = originalTime
		}

	case "cleanup":
		task.Options.SkipPrepare = true
		task.Options.SkipCleanup = false
		task.Options.WarmupTime = 0
		// Set time=0 to signal cleanup-only mode
		task.Parameters["time"] = 0
		// Don't save _original_time for cleanup - this signals cleanup-only mode
	}

	// Start benchmark with configured options
	run, err := p.benchmarkUC.StartBenchmark(ctx, task)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to start %s phase: %w", phase, err), p.win)
		return
	}

	// Store run ID for later reference
	p.currentRunID = run.ID
	slog.Info("Tasks: Benchmark phase started", "phase", phase, "run_id", run.ID, "task_id", task.ID)

	// Lock task form during execution
	p.setTaskFormEnabled(false)

	// Start monitoring
	p.isRunning = true
	p.statusLabel.SetText(fmt.Sprintf("Status: %s (Running)", strings.Title(phase)))
	p.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	p.btnPrepare.Disable()
	p.btnRun.Disable()
	p.btnCleanup.Disable()
	p.btnStop.Enable()

	// Reset log counter and map for new run
	p.lastLogCount = 0
	p.addedSeconds = make(map[string]bool)

	// Set realtime callback to receive samples directly (streaming, no polling)
	// This provides zero-delay UI updates compared to database polling
	if phase == "run" {
		p.benchmarkUC.SetRealtimeCallback(func(runID string, sample execution.MetricSample) {
			// Update UI in main thread using fyne.Do
			fyne.Do(func() {
				if !p.isRunning {
					return // Don't update if benchmark stopped
				}

				// Update metrics labels
				if sample.TPS > 0 {
					p.tpsLabel.SetText(fmt.Sprintf("%.0f", sample.TPS))
				}
				if sample.QPS > 0 {
					p.qpsLabel.SetText(fmt.Sprintf("%.0f", sample.QPS))
				}
				if sample.LatencyP95 > 0 {
					p.latencyP95Label.SetText(fmt.Sprintf("%.2fms", sample.LatencyP95))
				}
				p.errorsLabel.SetText(fmt.Sprintf("%.2f", sample.ErrorRate))

				// Update thread count from form
				threads := p.threadsEntry.Text
				if threads != "" {
					p.threadsLabel.SetText(threads)
				}

				// Update log with raw output line (with deduplication)
				if sample.RawLine != "" {
					// Extract second from raw line to prevent duplicates
					// Format: "[ 28s ] thds: 1 tps: ..."
					re := regexp.MustCompile(`\[\s*(\d+)s\s*\]`)
					matches := re.FindStringSubmatch(sample.RawLine)
					if len(matches) > 1 {
						secondKey := matches[1] + "s"
						if !p.addedSeconds[secondKey] {
							p.appendLogLine(sample.RawLine)
							p.addedSeconds[secondKey] = true
							slog.Info("Tasks: Realtime sample added", "second", secondKey, "run_id", runID)
						}
					} else {
						// No second marker, just add it
						p.appendLogLine(sample.RawLine)
					}
				}
			})
		})
	} else {
		// Clear callback for non-run phases
		p.benchmarkUC.SetRealtimeCallback(nil)
	}

	// Start monitoring goroutine (only for status tracking, not metrics)
	slog.Info("Tasks: Starting monitor goroutine", "run_id", run.ID, "phase", phase)
	go p.monitorBenchmarkProgress(ctx, run.ID, phase)
}

// startRealBenchmark starts the actual benchmark execution (all phases).
// Deprecated: Use startBenchmarkPhase for individual phase control.
func (p *TaskMonitorPage) startRealBenchmark(task *execution.BenchmarkTask) {
	p.startBenchmarkPhase(task, "run")
}

// onStopTask stops the running task.
func (p *TaskMonitorPage) onStopTask() {
	if !p.isRunning {
		return
	}

	slog.Info("Tasks: Stop button clicked, stopping task")

	// Stop the actual benchmark if running
	if p.currentRunID != "" && p.benchmarkUC != nil {
		ctx := context.Background()
		err := p.benchmarkUC.StopBenchmark(ctx, p.currentRunID, false)
		if err != nil {
			slog.Error("Tasks: Failed to stop benchmark", "error", err)
		} else {
			slog.Info("Tasks: Benchmark stopped", "run_id", p.currentRunID)
		}
	}

	// Reset UI state immediately
	p.isRunning = false
	p.statusLabel.SetText("Status: Stopped")
	p.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Reset all metrics and progress
	p.resetTaskMetrics()

	// Re-enable all phase buttons, disable Stop button
	p.btnPrepare.Enable()
	p.btnRun.Enable()
	p.btnCleanup.Enable()
	p.btnStop.Disable()
	p.setTaskFormEnabled(true)

	slog.Info("Tasks: Task stopped, UI reset completed")
}

// monitorBenchmarkProgress monitors the progress of a running benchmark phase.
// Note: Realtime metrics are now updated via callback, this only tracks status and progress bar.
func (p *TaskMonitorPage) monitorBenchmarkProgress(ctx context.Context, runID string, phase string) {
	slog.Info("Tasks: monitorBenchmarkProgress started", "run_id", runID, "phase", phase)
	defer slog.Info("Tasks: monitorBenchmarkProgress exiting", "run_id", runID, "phase", phase)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// For prepare and cleanup, only set progress once to avoid Fyne warnings
	progressSet := false

	for p.isRunning {
		select {
		case <-ticker.C:
			// Get current run status
			run, err := p.benchmarkUC.GetBenchmarkStatus(ctx, runID)
			if err != nil {
				slog.Error("Tasks: Failed to get benchmark status", "error", err)
				p.handleBenchmarkError(ctx, runID, fmt.Errorf("failed to get status: %w", err), phase)
				return
			}

			slog.Info("Tasks: Monitoring tick", "run_id", runID, "phase", phase, "state", run.State)

			// Check if run completed or failed
			if run.State == execution.StateCompleted {
				slog.Info("Tasks: State completed detected, calling handleBenchmarkCompleted", "run_id", runID)
				p.handleBenchmarkCompleted(ctx, run, phase)
				return
			}
			if run.State == execution.StateFailed || run.State == execution.StateCancelled || run.State == execution.StateForceStopped {
				slog.Info("Tasks: State stopped detected", "run_id", runID, "state", run.State)
				p.handleBenchmarkStopped(ctx, run, phase)
				return
			}

			// Update progress bar based on time (only for run phase)
			// Note: Metrics are updated via realtime callback, not here
			fyne.Do(func() {
				if phase == "run" && run.StartedAt != nil {
					elapsed := time.Since(*run.StartedAt).Seconds()
					duration := 60.0 // Default
					if dur, err := strconv.Atoi(p.durationEntry.Text); err == nil {
						duration = float64(dur)
					}
					progress := elapsed / duration
					if progress > 0.95 {
						progress = 0.95
					}
					p.progressBar.SetValue(progress)
				} else if phase != "run" && !progressSet {
					// For prepare and cleanup, only set progress once
					p.progressBar.SetValue(0.5) // Halfway to show activity
					progressSet = true
				}
			})

		case <-ctx.Done():
			return
		}
	}
}

// handleBenchmarkCompleted handles benchmark phase completion.
func (p *TaskMonitorPage) handleBenchmarkCompleted(ctx context.Context, run *execution.Run, phase string) {
	// Update UI state safely on main thread
	p.isRunning = false

	// Clear realtime callback to free resources
	if p.benchmarkUC != nil {
		p.benchmarkUC.SetRealtimeCallback(nil)
	}

	slog.Info("Tasks: handleBenchmarkCompleted called",
		"phase", phase,
		"run_id", run.ID,
		"has_result", run.Result != nil,
		"state", run.State)

	duration := "unknown"
	if run.Duration != nil {
		duration = fmt.Sprintf("%.1f seconds", run.Duration.Seconds())
	} else if run.StartedAt != nil && run.CompletedAt != nil {
		duration = fmt.Sprintf("%.1f seconds", run.CompletedAt.Sub(*run.StartedAt).Seconds())
	}

	// Update UI elements on main thread
	fyne.DoAndWait(func() {
		p.statusLabel.SetText(fmt.Sprintf("Status: %s Completed", strings.Title(phase)))
		p.progressBar.SetValue(1.0) // Show completion

		// Build completion message with detailed statistics
		var message string
		if run.Message != "" {
			message = run.Message
		} else if phase == "run" {
			if run.Result != nil {
				// Show detailed final statistics
				result := run.Result
				qps := 0.0
				if result.TotalTransactions > 0 {
					qps = result.TPSCalculated * float64(result.TotalQueries) / float64(result.TotalTransactions)
				}
				latencySumMs := 0.0
				if result.Duration > 0 {
					latencySumMs = result.Duration.Seconds() * 1000
				}

				message = fmt.Sprintf("Benchmark completed successfully!\n\n"+
					"Duration: %s\n\n"+
					"Transactions: %20d  (%.2f per sec.)\n"+
					"Queries:      %20d  (%.2f per sec.)\n\n"+
					"Latency (ms):\n"+
					"     min:      %25.2f\n"+
					"     avg:      %25.2f\n"+
					"     max:      %25.2f\n"+
					"     95th percentile: %15.2f\n"+
					"     sum:      %25.2f",
					duration,
					result.TotalTransactions,
					result.TPSCalculated,
					result.TotalQueries,
					qps,
					result.LatencyMin,
					result.LatencyAvg,
					result.LatencyMax,
					result.LatencyP95,
					latencySumMs)
			} else {
				// No result available, show simple message
				message = fmt.Sprintf("Benchmark completed successfully!\n\nDuration: %s\n\n(Note: Final statistics not available)", duration)
				slog.Warn("Tasks: Run completed but Result is nil")
			}
		} else {
			message = fmt.Sprintf("%s phase completed successfully!\n\nDuration: %s",
				strings.Title(phase), duration)
		}

		// Show Save/OK dialog for successful run completion
		if phase == "run" && run.Result != nil && p.historyUC != nil {
			p.showCompletionDialog(ctx, run, message)
		} else {
			// For prepare/cleanup phases or no history use case, show simple dialog
			dialog.ShowInformation(strings.Title(phase)+" Completed", message, p.win)
		}

		// Re-enable all phase buttons, disable stop
		p.btnPrepare.Enable()
		p.btnRun.Enable()
		p.btnCleanup.Enable()
		p.btnStop.Disable()
		p.setTaskFormEnabled(true)
	})

	slog.Info("Tasks: Benchmark phase completed", "phase", phase, "run_id", run.ID, "duration", duration)

	// Don't reset metrics - keep final TPS/QPS displayed
}

// showCompletionDialog shows a completion dialog with Save and OK buttons.
func (p *TaskMonitorPage) showCompletionDialog(ctx context.Context, run *execution.Run, message string) {
	// Create custom dialog with Save and OK buttons
	d := dialog.NewCustomConfirm("Benchmark Completed", "Save", "OK",
		widget.NewLabel(message),
		func(save bool) {
			if save && p.historyUC != nil {
				// Save to history
				if err := p.historyUC.SaveRunToHistory(ctx, run); err != nil {
					slog.Error("Tasks: Failed to save to history", "run_id", run.ID, "error", err)
					dialog.ShowError(fmt.Errorf("Failed to save to history: %v", err), p.win)
				} else {
					slog.Info("Tasks: Saved to history", "run_id", run.ID)
					dialog.ShowInformation("Saved", "✅ Run saved to History!\n\nGo to History tab to view details.", p.win)
				}
			}
			// OK button does nothing - just dismisses the dialog
		},
		p.win,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// handleBenchmarkStopped handles benchmark stop/cancellation.
func (p *TaskMonitorPage) handleBenchmarkStopped(ctx context.Context, run *execution.Run, phase string) {
	p.isRunning = false

	// Clear realtime callback
	if p.benchmarkUC != nil {
		p.benchmarkUC.SetRealtimeCallback(nil)
	}

	// Update UI on main thread
	fyne.DoAndWait(func() {
		p.statusLabel.SetText(fmt.Sprintf("Status: %s", run.State))

		// Check if there's a user-friendly message to display
		if run.Message != "" {
			dialog.ShowError(fmt.Errorf("%s", run.Message), p.win)
		}

		// Re-enable all phase buttons, disable stop
		p.btnPrepare.Enable()
		p.btnRun.Enable()
		p.btnCleanup.Enable()
		p.btnStop.Disable()
		p.setTaskFormEnabled(true)
	})
}

// handleBenchmarkError handles benchmark errors.
func (p *TaskMonitorPage) handleBenchmarkError(ctx context.Context, runID string, err error, phase string) {
	p.isRunning = false

	// Clear realtime callback
	if p.benchmarkUC != nil {
		p.benchmarkUC.SetRealtimeCallback(nil)
	}

	p.statusLabel.SetText("Status: Error")

	// Re-enable all phase buttons, disable stop
	fyne.Do(func() {
		p.btnPrepare.Enable()
		p.btnRun.Enable()
		p.btnCleanup.Enable()
		p.btnStop.Disable()
		p.setTaskFormEnabled(true)
	})

	// Show error dialog
	dialog.ShowError(fmt.Errorf("%s phase failed: %v", strings.Title(phase), err), p.win)
	slog.Error("Tasks: Benchmark phase failed", "phase", phase, "error", err)
}

// setTaskFormEnabled enables or disables the task form during execution.
func (p *TaskMonitorPage) setTaskFormEnabled(enabled bool) {
	// Note: Fyne Form doesn't have direct Enable/Disable
	// We'll rely on button states and user feedback
}

// appendLog appends a log message.
// resetTaskMetrics resets all task metrics to initial state.
func (p *TaskMonitorPage) resetTaskMetrics() {
	p.progressBar.SetValue(0)
	p.tpsLabel.SetText("--")
	p.qpsLabel.SetText("--")
	p.latencyP95Label.SetText("--")
	p.errorsLabel.SetText("0.00")
	p.threadsLabel.SetText("--")
	// Clear log
	p.logEntry.SetText("Waiting for benchmark data...\n")
	// Reset log counter
	p.lastLogCount = 0
	// Reset added seconds map
	p.addedSeconds = make(map[string]bool)
}

// appendLogLine appends a new line to the log output.
// Keeps only the last maxLogLines lines.
func (p *TaskMonitorPage) appendLogLine(line string) {
	currentText := p.logEntry.Text

	// If this is the first data line, clear the "Waiting..." message
	if strings.Contains(currentText, "Waiting for benchmark data") {
		currentText = ""
	}

	// Append new line
	newText := currentText
	if newText != "" && !strings.HasSuffix(newText, "\n") {
		newText += "\n"
	}
	newText += line

	// Split into lines and keep only last maxLogLines
	lines := strings.Split(newText, "\n")
	if len(lines) > p.maxLogLines {
		lines = lines[len(lines)-p.maxLogLines:]
	}

	// Join back and set
	p.logEntry.SetText(strings.Join(lines, "\n"))

	// Scroll to bottom
	p.logEntry.CursorRow = len(lines)
}
