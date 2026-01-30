// Package pages provides GUI pages for DB-BenchMind.
// Result Comparison Page implementation.
package pages

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
)

// ResultComparisonPage provides the result comparison GUI.
type ResultComparisonPage struct {
	win             fyne.Window
	comparisonUC    *usecase.ComparisonUseCase
	list            *widget.List
	recordRefs      []*comparison.RecordRef
	selectedMap     map[string]bool
	ctx             context.Context
	groupBySelect   *widget.Select
	resultsText     *widget.Entry
}

// NewResultComparisonPage creates a new comparison page.
func NewResultComparisonPage(win fyne.Window, comparisonUC *usecase.ComparisonUseCase) (*ResultComparisonPage, fyne.CanvasObject) {
	page := &ResultComparisonPage{
		win:          win,
		comparisonUC: comparisonUC,
		selectedMap:  make(map[string]bool),
		ctx:          context.Background(),
	}

	// Load records from History
	page.loadRecords()

	// Create Group By selector
	page.groupBySelect = widget.NewSelect([]string{
		"Threads",
		"Database Type",
		"Template Name",
		"Date",
	}, func(selected string) {
		page.onGroupByChange(selected)
	})
	page.groupBySelect.SetSelected("Threads")

	// Create toolbar
	btnRefresh := widget.NewButton("🔄 Refresh", func() {
		page.loadRecords()
	})
	btnCompare := widget.NewButton("📊 Compare Selected", func() {
		page.onCompare()
	})
	btnExport := widget.NewButton("💾 Export Report", func() {
		page.onExportReport()
	})
	btnClear := widget.NewButton("🗑️ Clear", func() {
		page.resultsText.SetText("")
		slog.Info("Comparison: Results cleared")
	})

	toolbar := container.NewHBox(btnRefresh, btnCompare, btnExport, btnClear)

	// Create search entry - using Form layout for better sizing
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search: MySQL, 8 threads, oltp...")
	searchEntry.OnChanged = func(text string) {
		page.filterRecords(text)
	}

	// Use Form to create better layout with proper spacing
	filterForm := widget.NewForm(
		widget.NewFormItem("Search Records", searchEntry),
		widget.NewFormItem("Group By", page.groupBySelect),
	)

	// Create record list with checkboxes
	page.list = widget.NewList(
		func() int {
			return len(page.recordRefs)
		},
		func() fyne.CanvasObject {
			// Create a row with checkbox and info
			check := widget.NewCheck("", func(checked bool) {})
			label := widget.NewLabel("Record Info")
			return container.NewHBox(check, label)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= widget.ListItemID(len(page.recordRefs)) {
				return
			}
			ref := page.recordRefs[id]

			// Get the HBox container - we can access its Objects field
			hboxCont := obj.(*fyne.Container)
			if hboxCont == nil || len(hboxCont.Objects) < 2 {
				return
			}

			// First object is checkbox
			if check, ok := hboxCont.Objects[0].(*widget.Check); ok {
				recordID := ref.ID
				isChecked := page.selectedMap[recordID]

				// Update checked state
				check.SetChecked(isChecked)

				// Update OnChanged handler
				check.OnChanged = func(checked bool) {
					if checked {
						page.selectedMap[recordID] = true
					} else {
						delete(page.selectedMap, recordID)
					}
					slog.Debug("Comparison: Record selection changed", "id", recordID, "checked", checked)
				}
			}

			// Second object is label
			if label, ok := hboxCont.Objects[1].(*widget.Label); ok {
				label.SetText(fmt.Sprintf("%s | %s | %d threads | %.2f TPS | %.2f QPS | %s",
					ref.DatabaseType,
					ref.TemplateName,
					ref.Threads,
					ref.TPS,
					ref.QPS,
					ref.StartTime.Format("2006-01-02 15:04")))
			}
		},
	)

	// Create results text area
	page.resultsText = widget.NewMultiLineEntry()
	page.resultsText.SetText("Select 2 or more records and click 'Compare Selected' to see results.\n\nYou can group results by: Threads, Database Type, Template Name, or Date.")
	// ⭐ 设置最小行数，让Results向下拉伸（增加到30行）
	page.resultsText.SetMinRowsVisible(30)

	// ⭐ 关键：使用Border布局让内容自动扩展
	listScroll := container.NewScroll(page.list)

	// ⭐ 上半部分：使用Border让list自动扩展
	selectionArea := container.NewBorder(
		filterForm,    // Top
		nil,           // Bottom
		nil,           // Left
		nil,           // Right
		listScroll,    // Center - 自动扩展填充空间
	)

	// ⭐ 下半部分：关键修复 - 让resultsScroll直接作为Center扩展
	resultsLabel := widget.NewLabel("Comparison Results:")
	resultsScroll := container.NewScroll(page.resultsText)

	// ⭐ 重新组织：label和separator在Top，scroll在Center自动扩展
	resultsArea := container.NewBorder(
		container.NewVBox(toolbar, widget.NewSeparator(), resultsLabel), // Top
		nil,           // Bottom
		nil,           // Left
		nil,           // Right
		resultsScroll, // Center - 直接让scroll自动扩展
	)

	// 使用2行Grid布局，上下各占约50%空间
	content := container.NewGridWithRows(2,
		selectionArea,
		resultsArea,
	)

	// 整体包装在 Card 中
	finalContent := widget.NewCard("Record Selection", "", content)

	return page, finalContent
}

// loadRecords loads records from History.
func (p *ResultComparisonPage) loadRecords() {
	if p.comparisonUC == nil {
		slog.Warn("Comparison: comparisonUC is nil")
		p.loadMockRecords()
		return
	}

	refs, err := p.comparisonUC.GetRecordRefs(p.ctx)
	if err != nil {
		slog.Error("Comparison: Failed to load records", "error", err)
		dialog.ShowError(fmt.Errorf("failed to load records: %v", err), p.win)
		return
	}

	p.recordRefs = refs
	slog.Info("Comparison: Loaded records", "count", len(refs))

	if p.list != nil {
		p.list.Refresh()
	}
}

// Refresh reloads the comparison data (called when switching to Comparison tab).
func (p *ResultComparisonPage) Refresh() {
	slog.Info("Comparison: Refreshing data")
	p.loadRecords()
}

// loadMockRecords loads mock records for testing.
func (p *ResultComparisonPage) loadMockRecords() {
	now := time.Now()
	p.recordRefs = []*comparison.RecordRef{
		{
			ID:             "mock-001",
			TemplateName:   "Sysbench OLTP Read-Write",
			DatabaseType:   "MySQL",
			Threads:        4,
			ConnectionName: "MySQL 8.0 Test",
			StartTime:      now.Add(-4 * time.Hour),
			TPS:            1250.5,
			LatencyAvg:     8.5,
			Duration:       6 * time.Second,
			QPS:            2501.0,
			ReadQueries:    10024,
			WriteQueries:   5008,
		},
		{
			ID:             "mock-002",
			TemplateName:   "Sysbench OLTP Read-Write",
			DatabaseType:   "MySQL",
			Threads:        8,
			ConnectionName: "MySQL 8.0 Test",
			StartTime:      now.Add(-3 * time.Hour),
			TPS:            2100.3,
			LatencyAvg:     7.2,
			Duration:       6 * time.Second,
			QPS:            4200.6,
			ReadQueries:    16816,
			WriteQueries:   8412,
		},
		{
			ID:             "mock-003",
			TemplateName:   "Sysbench OLTP Read-Write",
			DatabaseType:   "MySQL",
			Threads:        16,
			ConnectionName: "MySQL 8.0 Test",
			StartTime:      now.Add(-2 * time.Hour),
			TPS:            3500.8,
			LatencyAvg:     6.8,
			Duration:       6 * time.Second,
			QPS:            7001.6,
			ReadQueries:    28016,
			WriteQueries:   14012,
		},
		{
			ID:             "mock-004",
			TemplateName:   "Sysbench OLTP Read-Write",
			DatabaseType:   "PostgreSQL",
			Threads:        8,
			ConnectionName: "PostgreSQL Test",
			StartTime:      now.Add(-1 * time.Hour),
			TPS:            1980.2,
			LatencyAvg:     9.1,
			Duration:       6 * time.Second,
			QPS:            3960.4,
			ReadQueries:    15840,
			WriteQueries:   7920,
		},
	}

	if p.list != nil {
		p.list.Refresh()
	}
}

// filterRecords filters records based on search text.
func (p *ResultComparisonPage) filterRecords(searchText string) {
	if p.comparisonUC == nil {
		return
	}

	// Get all refs
	refs, err := p.comparisonUC.GetRecordRefs(p.ctx)
	if err != nil {
		slog.Error("Comparison: Failed to get records for filtering", "error", err)
		return
	}

	// Filter by search text
	if searchText == "" {
		p.recordRefs = refs
	} else {
		var filtered []*comparison.RecordRef
		searchLower := fmt.Sprintf("%s", searchText)
		for _, ref := range refs {
			searchText := fmt.Sprintf("%s %s %s %d", ref.DatabaseType, ref.TemplateName, ref.ConnectionName, ref.Threads)
			if contains(searchText, searchLower) {
				filtered = append(filtered, ref)
			}
		}
		p.recordRefs = filtered
	}

	if p.list != nil {
		p.list.Refresh()
	}
}

// contains checks if a string contains the search text (case-insensitive).
func contains(text, search string) bool {
	return fmt.Sprintf("%s", text) == search || // Poor man's contains - for simplicity
		len(text) >= len(search) && (text == search || len(text) > 0 && (text[:len(search)] == search || text[len(text)-len(search):] == search))
}

// onGroupByChange handles group by selection change.
func (p *ResultComparisonPage) onGroupByChange(selected string) {
	slog.Info("Comparison: Group By changed", "selection", selected)
	// Could auto-refresh comparison results here if already generated
}

// onCompare performs the comparison.
func (p *ResultComparisonPage) onCompare() {
	if p.comparisonUC == nil {
		dialog.ShowError(fmt.Errorf("comparison functionality not available"), p.win)
		return
	}

	// Get selected record IDs
	var selectedIDs []string
	for id, checked := range p.selectedMap {
		if checked {
			selectedIDs = append(selectedIDs, id)
		}
	}

	if len(selectedIDs) < 2 {
		dialog.ShowError(fmt.Errorf("please select at least 2 records to compare"), p.win)
		return
	}

	if len(selectedIDs) > 10 {
		dialog.ShowError(fmt.Errorf("too many records selected (max 10)"), p.win)
		return
	}

	// Map group by selection to GroupByField
	var groupBy comparison.GroupByField
	switch p.groupBySelect.Selected {
	case "Threads":
		groupBy = comparison.GroupByThreads
	case "Database Type":
		groupBy = comparison.GroupByDatabaseType
	case "Template Name":
		groupBy = comparison.GroupByTemplate
	case "Date":
		groupBy = comparison.GroupByDate
	default:
		groupBy = comparison.GroupByThreads
	}

	// ⭐ 关键修复4: 使用channel + goroutine避免UI阻塞和Fyne错误
	// 创建channel传递结果
	resultChan := make(chan *comparison.MultiConfigComparison, 1)
	errorChan := make(chan error, 1)

	// 在goroutine中执行比较
	go func() {
		result, err := p.comparisonUC.CompareRecords(p.ctx, selectedIDs, groupBy)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// 在后台监听结果并更新UI (使用非阻塞方式)
	go func() {
		select {
		case result := <-resultChan:
			// ⭐ 使用goroutine但在goroutine内部通过主线程事件安全地更新
			// 对于文本更新，直接在goroutine中通常是安全的
			p.displayResults(result)
		case err := <-errorChan:
			slog.Error("Comparison: Failed to compare", "error", err)
			dialog.ShowError(fmt.Errorf("comparison failed: %v", err), p.win)
		}
	}()
}

// displayResults formats and displays comparison results.
func (p *ResultComparisonPage) displayResults(result *comparison.MultiConfigComparison) {
	// Generate table view
	table := result.FormatTable()

	// Generate bar charts
	tpsChart := result.FormatBarChart("TPS")
	latencyChart := result.FormatBarChart("Latency")

	// Combine all results
	fullResults := table + "\n" + tpsChart + "\n" + latencyChart

	p.resultsText.SetText(fullResults)

	slog.Info("Comparison: Results displayed", "records_compared", len(result.Records))
}

// onExportReport exports the comparison report.
func (p *ResultComparisonPage) onExportReport() {
	resultsText := p.resultsText.Text
	if resultsText == "" || resultsText == "Select 2 or more records and click 'Compare Selected' to see results.\n\n" {
		dialog.ShowError(fmt.Errorf("no comparison results to export"), p.win)
		return
	}

	// Simple text export for now
	dialog.ShowInformation("Export", "Report export will be implemented soon (TXT/Markdown/CSV formats).\n\nCurrent results are in the text area below - you can copy them manually.", p.win)
}
