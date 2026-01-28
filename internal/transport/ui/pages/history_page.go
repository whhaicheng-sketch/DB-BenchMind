// Package pages provides GUI pages for DB-BenchMind.
// History Records Page implementation.
package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"time"
)

// HistoryRecordPage provides the history records GUI.
type HistoryRecordPage struct {
	win      fyne.Window
	list     *widget.List
	records  []historyRecord
	selected int
}

// historyRecord represents a run history record.
type historyRecord struct {
	ID        string
	Name      string
	Tool      string
	StartTime time.Time
	Duration  time.Duration
	Status    string
	TPS       float64
}

// NewHistoryRecordPage creates a new history page.
func NewHistoryRecordPage(win fyne.Window) fyne.CanvasObject {
	page := &HistoryRecordPage{
		win:      win,
		selected: -1,
	}
	// Load mock history records
	page.loadHistory()
	// Create history list
	page.list = widget.NewList(
		func() int {
			return len(page.records)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Run Record")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= widget.ListItemID(len(page.records)) {
				return
			}
			record := page.records[id]
			label := obj.(*widget.Label)
			label.SetText(fmt.Sprintf("%s | %s | %s | %s",
				record.Name,
				record.Tool,
				record.StartTime.Format("2006-01-02 15:04"),
				record.Status))
		},
	)
	page.list.OnSelected = func(id widget.ListItemID) {
		page.selected = int(id)
	}
	// Create toolbar
	btnRefresh := widget.NewButton("Refresh", func() {
		page.loadHistory()
	})
	btnView := widget.NewButton("View Details", func() {
		page.onViewDetails()
	})
	btnDelete := widget.NewButton("Delete", func() {
		page.onDelete()
	})
	btnExport := widget.NewButton("Export Results", func() {
		page.onExport()
	})
	toolbar := container.NewHBox(btnRefresh, btnView, btnDelete, btnExport)
	// Create summary label
	summaryLabel := widget.NewLabel(fmt.Sprintf("Total Runs: %d", len(page.records)))
	content := container.NewVBox(
		toolbar,
		widget.NewSeparator(),
		summaryLabel,
		widget.NewSeparator(),
		container.NewPadded(page.list),
	)
	return content
}

// loadHistory loads history records.
func (p *HistoryRecordPage) loadHistory() {
	// Mock data - in production, load from database
	now := time.Now()
	p.records = []historyRecord{
		{
			ID:        "run-001",
			Name:      "MySQL OLTP Test",
			Tool:      "Sysbench",
			StartTime: now.Add(-2 * time.Hour),
			Duration:  5 * time.Minute,
			Status:    "Completed",
			TPS:       1250.5,
		},
		{
			ID:        "run-002",
			Name:      "PostgreSQL TPCC",
			Tool:      "HammerDB",
			StartTime: now.Add(-24 * time.Hour),
			Duration:  10 * time.Minute,
			Status:    "Completed",
			TPS:       980.2,
		},
		{
			ID:        "run-003",
			Name:      "Oracle SOE",
			Tool:      "Swingbench",
			StartTime: now.Add(-48 * time.Hour),
			Duration:  15 * time.Minute,
			Status:    "Failed",
			TPS:       0,
		},
	}
	if p.list != nil {
		p.list.Refresh()
	}
}

// onViewDetails shows record details.
func (p *HistoryRecordPage) onViewDetails() {
	if p.selected < 0 || p.selected >= len(p.records) {
		dialog.ShowError(fmt.Errorf("please select a record"), p.win)
		return
	}
	record := p.records[p.selected]
	details := fmt.Sprintf(
		"Run ID: %s\n\nName: %s\nTool: %s\nStart Time: %s\nDuration: %v\nStatus: %s\nTPS: %.2f",
		record.ID,
		record.Name,
		record.Tool,
		record.StartTime.Format("2006-01-02 15:04:05"),
		record.Duration,
		record.Status,
		record.TPS,
	)
	dialog.ShowInformation("Run Details", details, p.win)
}

// onDelete deletes a record.
func (p *HistoryRecordPage) onDelete() {
	if p.selected < 0 || p.selected >= len(p.records) {
		dialog.ShowError(fmt.Errorf("please select a record"), p.win)
		return
	}
	record := p.records[p.selected]
	dialog.ShowConfirm(
		"Delete Record",
		fmt.Sprintf("Delete run '%s'?", record.Name),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			// Remove from list (in production, delete from database)
			p.records = append(p.records[:p.selected], p.records[p.selected+1:]...)
			p.selected = -1
			p.list.Refresh()
			dialog.ShowInformation("Deleted", "Record deleted successfully", p.win)
		},
		p.win,
	)
}

// onExport exports results.
func (p *HistoryRecordPage) onExport() {
	if p.selected < 0 || p.selected >= len(p.records) {
		dialog.ShowError(fmt.Errorf("please select a record"), p.win)
		return
	}
	dialog.ShowInformation("Export", "Export functionality will be implemented soon", p.win)
}
