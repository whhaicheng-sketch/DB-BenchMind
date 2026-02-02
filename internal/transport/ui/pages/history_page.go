// Package pages provides GUI pages for DB-BenchMind.
// History Records Page implementation.
package pages

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/history"
)

// HistoryRecordPage provides the history records GUI.
type HistoryRecordPage struct {
	win         fyne.Window
	historyUC   *usecase.HistoryUseCase
	exportUC    *usecase.ExportUseCase
	list        *widget.List
	records     []*history.Record
	selected    int
	ctx         context.Context
	summaryLabel *widget.Label  // Need to keep reference to update
}

// historyRecordListItem represents a list item for display.
type historyRecordListItem struct {
	ID            string
	Connection    string
	Template      string
	DatabaseType  string
	Threads       int
	Duration      time.Duration
	TPS           float64
	StartTime     time.Time
}

// NewHistoryRecordPage creates a new history page.
// Returns both the canvas object and the page instance for external refresh control.
func NewHistoryRecordPage(win fyne.Window, historyUC *usecase.HistoryUseCase, exportUC *usecase.ExportUseCase) (*HistoryRecordPage, fyne.CanvasObject) {
	page := &HistoryRecordPage{
		win:       win,
		historyUC: historyUC,
		exportUC:  exportUC,
		selected:  -1,
		ctx:       context.Background(),
	}

	// Load history records from database
	page.loadHistory()

	// Create history list with inline action buttons
	page.list = widget.NewList(
		func() int {
			return len(page.records)
		},
		func() fyne.CanvasObject {
			// Create label and buttons for each row
			label := widget.NewLabel("Run Record")

			// Details button - blue theme color symbol
			btnView := widget.NewButton("🔍 Details", nil)
			btnView.Importance = widget.LowImportance

			// Delete button - red danger symbol
			btnDelete := widget.NewButton("❌ Delete", nil)
			btnDelete.Importance = widget.LowImportance

			// Export button - green action symbol
			btnExport := widget.NewButton("📥 Export", nil)
			btnExport.Importance = widget.LowImportance

			// Create HBox with label (left) and buttons (right)
			content := container.NewHBox(
				label,
				layout.NewSpacer(),
				btnView,
				btnDelete,
				btnExport,
			)

			return content
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= widget.ListItemID(len(page.records)) {
				return
			}
			record := page.records[id]

			// Get the HBox container
			if hbox, ok := obj.(*fyne.Container); ok {
				objects := hbox.Objects
				if len(objects) >= 5 {
					// First object is the label
					if label, ok := objects[0].(*widget.Label); ok {
						label.SetText(fmt.Sprintf("%s | %s | %s | %d threads | %.2f TPS | %s",
							record.ConnectionName,
							record.TemplateName,
							record.DatabaseType,
							record.Threads,
							record.TPSCalculated,
							record.StartTime.Format("2006-01-02 15:04")))
					}

					// Update button handlers
					recordIndex := int(id)

					// Third object (index 2) is View Details button
					if btnView, ok := objects[2].(*widget.Button); ok {
						btnView.OnTapped = func() {
							page.selected = recordIndex
							page.onViewDetails()
						}
					}

					// Fourth object (index 3) is Delete button
					if btnDelete, ok := objects[3].(*widget.Button); ok {
						btnDelete.OnTapped = func() {
							page.selected = recordIndex
							page.onDelete()
						}
					}

					// Fifth object (index 4) is Export button
					if btnExport, ok := objects[4].(*widget.Button); ok {
						btnExport.OnTapped = func() {
							page.selected = recordIndex
							page.onExport()
						}
					}
				}
			}
		},
	)

	// Create toolbar - Refresh, Delete All, Export All
	btnRefresh := widget.NewButton("🔄 Refresh", func() {
		page.Refresh()
	})
	btnDeleteAll := widget.NewButton("🗑️ Delete All", func() {
		page.onDeleteAll()
	})
	btnExportAll := widget.NewButton("💾 Export All", func() {
		page.onExportAll()
	})

	toolbar := container.NewHBox(btnRefresh, btnDeleteAll, btnExportAll)

	// Create summary label
	page.summaryLabel = widget.NewLabel(fmt.Sprintf("Total Runs: %d", len(page.records)))
	content := container.NewBorder(
		container.NewVBox(toolbar, widget.NewSeparator(), page.summaryLabel, widget.NewSeparator()), // top
		nil, // bottom
		nil, // left
		nil, // right
		page.list, // center - will expand to fill available space
	)
	return page, content
}

// loadHistory loads history records from database.
func (p *HistoryRecordPage) loadHistory() {
	if p.historyUC == nil {
		slog.Warn("History: historyUC is nil, using mock data")
		p.loadMockHistory()
		return
	}

	records, err := p.historyUC.GetAllRecords(p.ctx)
	if err != nil {
		slog.Error("History: Failed to load records", "error", err)
		dialog.ShowError(fmt.Errorf("failed to load history: %v", err), p.win)
		return
	}

	p.records = records
	if p.list != nil {
		p.list.Refresh()
	}

	// Update summary label
	if p.summaryLabel != nil {
		p.summaryLabel.SetText(fmt.Sprintf("Total Runs: %d", len(records)))
	}

	slog.Info("History: Loaded records", "count", len(records))
}

// Refresh refreshes the history list and summary.
func (p *HistoryRecordPage) Refresh() {
	p.loadHistory()
}

// loadMockHistory loads mock history records (fallback).
func (p *HistoryRecordPage) loadMockHistory() {
	now := time.Now()
	p.records = []*history.Record{
		{
			ID:            "run-001",
			CreatedAt:     now.Add(-2 * time.Hour),
			ConnectionName: "MySQL Test",
			TemplateName:  "Sysbench OLTP Read-Write",
			DatabaseType:  "MySQL",
			Threads:       8,
			StartTime:     now.Add(-2 * time.Hour),
			Duration:      5 * time.Minute,
			TPSCalculated: 1250.5,
		},
		{
			ID:            "run-002",
			CreatedAt:     now.Add(-24 * time.Hour),
			ConnectionName: "PostgreSQL Test",
			TemplateName:  "Sysbench OLTP Read-Write",
			DatabaseType:  "PostgreSQL",
			Threads:       16,
			StartTime:     now.Add(-24 * time.Hour),
			Duration:      10 * time.Minute,
			TPSCalculated: 980.2,
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

	// Calculate per-second rates
	durationSec := record.Duration.Seconds()
	qps := 0.0
	if durationSec > 0 && record.TotalQueries > 0 {
		qps = float64(record.TotalQueries) / durationSec
	}
	ignoredErrorsPerSec := 0.0
	if durationSec > 0 {
		ignoredErrorsPerSec = float64(record.IgnoredErrors) / durationSec
	}
	reconnectsPerSec := 0.0
	if durationSec > 0 {
		reconnectsPerSec = float64(record.Reconnects) / durationSec
	}

	// Build detailed statistics message in sysbench format
	details := fmt.Sprintf(
		"Connection: %s\n"+
		"Template: %s\n"+
		"Database Type: %s\n"+
		"Threads: %d\n"+
		"Start Time: %s\n"+
		"Duration: %v\n\n"+
		"SQL statistics:\n"+
		"    queries performed:\n"+
		"        read:                            %d\n"+
		"        write:                           %d\n"+
		"        other:                           %d\n"+
		"        total:                           %d\n"+
		"    transactions:                        %d  (%.2f per sec.)\n"+
		"    queries:                             %d (%.2f per sec.)\n"+
		"    ignored errors:                      %d      (%.2f per sec.)\n"+
		"    reconnects:                          %d      (%.2f per sec.)\n\n"+
		"General statistics:\n"+
		"    total time:                          %.4fs\n"+
		"    total number of events:              %d\n\n"+
		"Latency (ms):\n"+
		"         min:                                    %.2f\n"+
		"         avg:                                   %.2f\n"+
		"         max:                                   %.2f\n"+
		"         95th percentile:                       %.2f\n"+
		"         99th percentile:                       %.2f\n\n"+
		"Threads fairness:\n"+
		"    events (avg/stddev):           %.4f/%.2f\n"+
		"    execution time (avg/stddev):   %.4f/%.2f",
		record.ConnectionName,
		record.TemplateName,
		record.DatabaseType,
		record.Threads,
		record.StartTime.Format("2006-01-02 15:04:05"),
		record.Duration,
		record.ReadQueries,
		record.WriteQueries,
		record.OtherQueries,
		record.TotalQueries,
		record.TotalTransactions,
		record.TPSCalculated,
		record.TotalQueries,
		qps,
		record.IgnoredErrors,
		ignoredErrorsPerSec,
		record.Reconnects,
		reconnectsPerSec,
		record.TotalTime,
		record.TotalEvents,
		record.LatencyMin,
		record.LatencyAvg,
		record.LatencyMax,
		record.LatencyP95,
		record.LatencyP99,
		record.EventsAvg,
		record.EventsStddev,
		record.ExecTimeAvg,
		record.ExecTimeStddev,
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
		fmt.Sprintf("Delete run '%s' from %s?", record.TemplateName, record.StartTime.Format("2006-01-02 15:04")),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			// Delete from database
			if p.historyUC != nil {
				if err := p.historyUC.DeleteRecord(p.ctx, record.ID); err != nil {
					slog.Error("History: Failed to delete record", "id", record.ID, "error", err)
					dialog.ShowError(fmt.Errorf("failed to delete: %v", err), p.win)
					return
				}
			}
			// Remove from list
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

	if p.exportUC == nil {
		dialog.ShowError(fmt.Errorf("export functionality not available"), p.win)
		return
	}

	record := p.records[p.selected]

	// Create format selection dialog
	formatSelect := widget.NewRadioGroup([]string{"TXT", "Markdown"}, func(selected string) {})
	formatSelect.SetSelected("TXT") // Default to TXT

	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Export selected record: %s", record.TemplateName)),
		widget.NewLabel(fmt.Sprintf("Run at: %s", record.StartTime.Format("2006-01-02 15:04"))),
		widget.NewSeparator(),
		widget.NewLabel("Select export format:"),
		formatSelect,
	)

	dialog.ShowCustomConfirm("Export One Record", "Export", "Cancel", form, func(export bool) {
		if !export {
			return
		}

		// Map radio selection to ExportFormat
		var format usecase.ExportFormat
		switch formatSelect.Selected {
		case "TXT":
			format = usecase.FormatTXT
		case "Markdown":
			format = usecase.FormatMarkdown
		default:
			format = usecase.FormatTXT
		}

		// Export immediately (in goroutine to avoid blocking UI)
		go func() {
			filepath, err := p.exportUC.ExportRecord(p.ctx, record, format)
			if err != nil {
				slog.Error("History: Failed to export record", "id", record.ID, "error", err)
				dialog.ShowError(fmt.Errorf("export failed: %v", err), p.win)
				return
			}

			slog.Info("History: Exported record", "id", record.ID, "format", format, "filepath", filepath)
			dialog.ShowInformation("Export Successful",
				fmt.Sprintf("Record exported to:\n%s\n\nFormat: %s", filepath, format),
				p.win)
		}()
	}, p.win)
}

// onExportAll exports all history records.
func (p *HistoryRecordPage) onExportAll() {
	if p.exportUC == nil {
		dialog.ShowError(fmt.Errorf("export functionality not available"), p.win)
		return
	}

	if len(p.records) == 0 {
		dialog.ShowError(fmt.Errorf("no records to export"), p.win)
		return
	}

	// Create format selection dialog
	formatSelect := widget.NewRadioGroup([]string{"TXT", "Markdown"}, func(selected string) {})
	formatSelect.SetSelected("TXT") // Default to TXT

	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Export ALL history records (%d records)", len(p.records))),
		widget.NewLabel("All records will be exported to the exports directory."),
		widget.NewSeparator(),
		widget.NewLabel("Select export format:"),
		formatSelect,
	)

	dialog.ShowCustomConfirm("Export All Records", "Export", "Cancel", form, func(export bool) {
		if !export {
			return
		}

		// Map radio selection to ExportFormat
		var format usecase.ExportFormat
		switch formatSelect.Selected {
		case "TXT":
			format = usecase.FormatTXT
		case "Markdown":
			format = usecase.FormatMarkdown
		default:
			format = usecase.FormatTXT
		}

		// Export all records immediately (in goroutine to avoid blocking UI)
		go func() {
			count, exportDir, err := p.exportUC.ExportAllRecords(p.ctx, p.records, format)
			if err != nil {
				slog.Error("History: Failed to export all records", "error", err)
				// Show partial success message
				if count > 0 {
					dialog.ShowInformation("Export Partially Completed",
						fmt.Sprintf("Successfully exported %d out of %d records to:\n%s\n\n%d records failed.\n\nCheck logs for details.",
							count, len(p.records), exportDir, len(p.records)-count),
						p.win)
				} else {
					dialog.ShowError(fmt.Errorf("export failed: %v", err), p.win)
				}
				return
			}

			slog.Info("History: Exported all records", "count", count, "format", format, "directory", exportDir)
			dialog.ShowInformation("Export All Successful",
				fmt.Sprintf("Successfully exported %d records to:\n%s\n\nFormat: %s", count, exportDir, format),
				p.win)
		}()
	}, p.win)
}

// onDeleteAll deletes all history records after confirmation.
func (p *HistoryRecordPage) onDeleteAll() {
	if len(p.records) == 0 {
		dialog.ShowInformation("Delete All", "No records to delete", p.win)
		return
	}

	dialog.ShowConfirm(
		"Delete All Records",
		fmt.Sprintf("Are you sure you want to delete ALL %d history records?\n\nThis action cannot be undone!", len(p.records)),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			recordCount := len(p.records)
			slog.Info("History: Deleting all records", "count", recordCount)

			// Delete all records from database
			if p.historyUC != nil {
				for _, record := range p.records {
					if err := p.historyUC.DeleteRecord(p.ctx, record.ID); err != nil {
						slog.Error("History: Failed to delete record", "id", record.ID, "error", err)
					}
				}
			}

			// Clear the list
			p.records = []*history.Record{}
			p.selected = -1
			p.list.Refresh()

			// Update summary
			p.summaryLabel.SetText("Total Runs: 0")

			slog.Info("History: All records deleted successfully", "count", recordCount)
			dialog.ShowInformation("Delete All Successful",
				fmt.Sprintf("Successfully deleted all %d records", recordCount),
				p.win)
		},
		p.win,
	)
}
