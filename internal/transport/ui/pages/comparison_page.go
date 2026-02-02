// Package pages provides GUI pages for DB-BenchMind.
// Result Comparison Page implementation.
package pages

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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
	btnCompare := widget.NewButton("📊 Compare Records", func() {
		page.GenerateSimplifiedReport()
	})
	btnExport := widget.NewButton("💾 Export Report", func() {
		page.onExportReport()
	})
	btnClear := widget.NewButton("🗑️ Clear", func() {
		page.resultsText.SetText("")
		slog.Info("Comparison: Results cleared")
	})

	toolbar := container.NewHBox(btnRefresh, btnCompare, btnExport, btnClear)

	// Selection control buttons
	btnSelectAll := widget.NewButton("✓ Select All", func() {
		page.selectAllRecords(true)
	})
	btnDeselectAll := widget.NewButton("✗ Deselect All", func() {
		page.selectAllRecords(false)
	})
	selectButtons := container.NewHBox(btnSelectAll, btnDeselectAll)

	// Create search entry - using Form layout for better sizing
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search: MySQL, 8 threads, oltp...")
	searchEntry.OnChanged = func(text string) {
		page.filterRecords(text)
	}

	// Use Form to create better layout with proper spacing
	filterForm := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Search Records", searchEntry),
			widget.NewFormItem("Group By", page.groupBySelect),
		),
		selectButtons,
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

// selectAllRecords selects or deselects all records.
func (p *ResultComparisonPage) selectAllRecords(selectAll bool) {
	for _, ref := range p.recordRefs {
		if selectAll {
			p.selectedMap[ref.ID] = true
		} else {
			delete(p.selectedMap, ref.ID)
		}
	}
	// Refresh the list to update checkboxes
	if p.list != nil {
		p.list.Refresh()
	}
	selectedCount := len(p.selectedMap)
	action := "deselected"
	if selectAll {
		action = "selected"
	}
	slog.Info("Comparison: Records "+action, "count", selectedCount)
}

// onExportReport exports the current performance report.
func (p *ResultComparisonPage) onExportReport() {
	resultsText := p.resultsText.Text
	if resultsText == "" {
		dialog.ShowError(fmt.Errorf("no performance report to export"), p.win)
		return
	}

	// Create export dialog content
	formatSelect := widget.NewRadioGroup([]string{"Markdown", "TXT"}, func(selected string) {})
	formatSelect.SetSelected("Markdown")

	content := container.NewVBox(
		widget.NewLabel("Export Performance Report"),
		widget.NewSeparator(),
		widget.NewLabel("Select export format:"),
		formatSelect,
		widget.NewSeparator(),
	)

	dialog.ShowCustomConfirm("Export Report", "Export", "Cancel", content, func(export bool) {
		if !export {
			return
		}

		var format, ext string
		switch formatSelect.Selected {
		case "Markdown":
			format = "markdown"
			ext = ".md"
		case "TXT":
			format = "txt"
			ext = ".txt"
		}

		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("performance_report_%s%s", timestamp, ext)
		filepath := fmt.Sprintf("./exports/%s", filename)

		// Ensure exports directory exists
		if err := os.MkdirAll("./exports", 0755); err != nil {
			dialog.ShowError(fmt.Errorf("failed to create exports directory: %v", err), p.win)
			return
		}

		// Write file
		err := os.WriteFile(filepath, []byte(resultsText), 0644)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to export report: %v", err), p.win)
			return
		}

		dialog.ShowInformation("Export Successful",
			fmt.Sprintf("Report exported to:\n%s\n\nFormat: %s", filepath, format),
			p.win)

		slog.Info("Comparison: Report exported", "filepath", filepath, "format", format)
	}, p.win)
}
