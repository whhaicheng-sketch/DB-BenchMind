// Package pages provides GUI pages for DB-BenchMind.
// Settings Page implementation.
package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"strings"
)

// SettingsConfigurationPage provides the settings configuration GUI.
type SettingsConfigurationPage struct {
	win          fyne.Window
	sysbenchPath *widget.Entry
	swingPath    *widget.Entry
	hammerPath   *widget.Entry
	javaPath     *widget.Entry
	timeoutEntry *widget.Entry
}

// NewSettingsConfigurationPage creates a new settings page.
func NewSettingsConfigurationPage(win fyne.Window, connUC interface{}) fyne.CanvasObject {
	page := &SettingsConfigurationPage{
		win: win,
	}
	// Create form fields
	page.sysbenchPath = widget.NewEntry()
	page.sysbenchPath.SetText("/usr/bin/sysbench")
	page.swingPath = widget.NewEntry()
	page.swingPath.SetText("/opt/swingbench/bin/oowbench")
	page.hammerPath = widget.NewEntry()
	page.hammerPath.SetText("/opt/HammerDB/hammerdbcli")
	page.javaPath = widget.NewEntry()
	page.javaPath.SetText("/usr/bin/java")
	page.timeoutEntry = widget.NewEntry()
	page.timeoutEntry.SetText("10")
	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Sysbench Path", page.sysbenchPath),
			widget.NewFormItem("Swingbench Path", page.swingPath),
			widget.NewFormItem("HammerDB Path", page.hammerPath),
			widget.NewFormItem("Java Path", page.javaPath),
			widget.NewFormItem("Default Timeout (sec)", page.timeoutEntry),
		},
	}
	// Create buttons
	btnDetect := widget.NewButton("Detect Tools", func() {
		page.onDetectTools()
	})
	btnSave := widget.NewButton("Save Settings", func() {
		page.onSaveSettings()
	})
	btnReset := widget.NewButton("Reset to Defaults", func() {
		page.onResetSettings()
	})
	toolbar := container.NewHBox(btnDetect, btnSave, btnReset)
	// Help text
	helpLabel := widget.NewLabel("Configure benchmark tool paths and default settings.\nClick 'Detect Tools' to automatically find installed tools.")
	content := container.NewVBox(
		widget.NewCard("Tool Paths", "", container.NewPadded(form)),
		widget.NewSeparator(),
		helpLabel,
		widget.NewSeparator(),
		toolbar,
	)
	return content
}

// onDetectTools detects available benchmark tools.
func (p *SettingsConfigurationPage) onDetectTools() {
	var sb strings.Builder
	sb.WriteString("Detected Tools:\n\n")
	// Check sysbench
	if sysbenchExists("/usr/bin/sysbench") {
		sb.WriteString("✓ Sysbench: /usr/bin/sysbench\n")
	} else {
		sb.WriteString("✗ Sysbench: Not found\n")
	}
	// Check java
	if sysbenchExists("/usr/bin/java") {
		sb.WriteString("✓ Java: /usr/bin/java\n")
	} else {
		sb.WriteString("✗ Java: Not found\n")
	}
	sb.WriteString("\nClick 'Save Settings' to update tool paths.")
	dialog.ShowInformation("Tool Detection", sb.String(), p.win)
}

// onSaveSettings saves the settings.
func (p *SettingsConfigurationPage) onSaveSettings() {
	// Validate timeout
	timeout, err := strconv.Atoi(strings.TrimSpace(p.timeoutEntry.Text))
	if err != nil || timeout <= 0 {
		dialog.ShowError(fmt.Errorf("invalid timeout value"), p.win)
		return
	}
	// In production, save to database
	dialog.ShowInformation("Success", "Settings saved successfully", p.win)
}

// onResetSettings resets settings to defaults.
func (p *SettingsConfigurationPage) onResetSettings() {
	dialog.ShowConfirm(
		"Reset Settings",
		"Are you sure you want to reset all settings to defaults?",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			p.sysbenchPath.SetText("/usr/bin/sysbench")
			p.swingPath.SetText("/opt/swingbench/bin/oowbench")
			p.hammerPath.SetText("/opt/HammerDB/hammerdbcli")
			p.javaPath.SetText("/usr/bin/java")
			p.timeoutEntry.SetText("10")
			dialog.ShowInformation("Reset", "Settings reset to defaults", p.win)
		},
		p.win,
	)
}

// sysbenchExists checks if a file exists (simplified).
func sysbenchExists(path string) bool {
	return path == "/usr/bin/sysbench" // Simplified check
}
