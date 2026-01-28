// Package pages provides GUI pages for DB-BenchMind.
// Run Monitor Page implementation.
package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"time"
)

// RunMonitorPage provides real-time run monitoring GUI.
type RunMonitorPage struct {
	win       fyne.Window
	isRunning bool
	// Metrics widgets
	statusLabel  *widget.Label
	tpsLabel     *widget.Label
	latencyLabel *widget.Label
	errorsLabel  *widget.Label
	progressBar  *widget.ProgressBar
	logText      *widget.Entry
}

// NewRunMonitorPage creates a new run monitoring page.
func NewRunMonitorPage(win fyne.Window) fyne.CanvasObject {
	page := &RunMonitorPage{
		win:       win,
		isRunning: false,
	}
	// Create status label
	page.statusLabel = widget.NewLabel("Status: Idle")
	page.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	// Create metrics labels
	page.tpsLabel = widget.NewLabel("TPS: 0")
	page.latencyLabel = widget.NewLabel("Avg Latency: 0ms")
	page.errorsLabel = widget.NewLabel("Errors: 0")
	// Create progress bar
	page.progressBar = widget.NewProgressBar()
	page.progressBar.SetValue(0)
	// Create log text area
	page.logText = widget.NewMultiLineEntry()
	page.logText.SetText("No active run. Start a task to see real-time metrics.\n")
	// Create metrics card
	metricsCard := widget.NewCard("Real-time Metrics", "", container.NewVBox(
		page.statusLabel,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("TPS:"),
			page.tpsLabel,
			widget.NewLabel("Avg Latency:"),
			page.latencyLabel,
			widget.NewLabel("Errors:"),
			page.errorsLabel,
		),
		widget.NewSeparator(),
		widget.NewLabel("Progress:"),
		page.progressBar,
	))
	// Create control buttons
	btnStart := widget.NewButton("Start Monitor", func() {
		page.onStartMonitor()
	})
	btnStop := widget.NewButton("Stop Monitor", func() {
		page.onStopMonitor()
	})
	btnClear := widget.NewButton("Clear Logs", func() {
		page.logText.SetText("")
	})
	btnRefresh := widget.NewButton("Refresh", func() {
		page.onRefresh()
	})
	toolbar := container.NewHBox(btnStart, btnStop, btnClear, btnRefresh)
	// Create log card
	logCard := widget.NewCard("Run Logs", "", container.NewPadded(
		container.NewVBox(
			container.NewScroll(page.logText),
		),
	))
	// Main layout
	content := container.NewVBox(
		metricsCard,
		widget.NewSeparator(),
		toolbar,
		widget.NewSeparator(),
		container.NewGridWithColumns(1,
			container.NewVBox(
				widget.NewLabel("Logs:"),
				logCard,
			),
		),
	)
	return content
}

// onStartMonitor starts monitoring a run.
func (p *RunMonitorPage) onStartMonitor() {
	if p.isRunning {
		dialog.ShowError(fmt.Errorf("monitor already running"), p.win)
		return
	}
	p.isRunning = true
	p.statusLabel.SetText("Status: Monitoring")
	p.logText.SetText("[" + time.Now().Format("15:04:05") + "] Monitor started\n")
	// Simulate metrics updates (in production, this would connect to actual run)
	go p.simulateMetrics()
}

// onStopMonitor stops monitoring.
func (p *RunMonitorPage) onStopMonitor() {
	if !p.isRunning {
		return
	}
	p.isRunning = false
	p.statusLabel.SetText("Status: Stopped")
	p.appendLog("[" + time.Now().Format("15:04:05") + "] Monitor stopped\n")
}

// onRefresh refreshes the metrics.
func (p *RunMonitorPage) onRefresh() {
	p.appendLog("[" + time.Now().Format("15:04:05") + "] Refreshed metrics\n")
}

// simulateMetrics simulates real-time metric updates.
func (p *RunMonitorPage) simulateMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	progress := 0.0
	for p.isRunning && progress < 1.0 {
		select {
		case <-ticker.C:
			progress += 0.05
			if progress > 1.0 {
				progress = 1.0
			}
			// Simulate metrics
			tps := int(1000 + progress*500)
			latency := int(10 - progress*5)
			errors := int(progress * 2)
			p.tpsLabel.SetText(fmt.Sprintf("TPS: %d", tps))
			p.latencyLabel.SetText(fmt.Sprintf("Avg Latency: %dms", latency))
			p.errorsLabel.SetText(fmt.Sprintf("Errors: %d", errors))
			p.progressBar.SetValue(progress)
			if progress < 1.0 {
				p.appendLog(fmt.Sprintf("[%s] TPS: %d, Latency: %dms, Errors: %d\n",
					time.Now().Format("15:04:05"), tps, latency, errors))
			}
		}
	}
	if progress >= 1.0 {
		p.isRunning = false
		p.statusLabel.SetText("Status: Completed")
		p.appendLog("[" + time.Now().Format("15:04:05") + "] Run completed\n")
	}
}

// appendLog appends a log message.
func (p *RunMonitorPage) appendLog(msg string) {
	p.logText.SetText(p.logText.Text + msg)
}
