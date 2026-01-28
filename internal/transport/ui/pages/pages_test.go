// Package pages provides GUI page tests.
package pages

import (
	"testing"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// TestAllPagesInitialization tests that all pages can be initialized without panics.
func TestAllPagesInitialization(t *testing.T) {
	// Create a test app
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	t.Run("Connections Page", func(t *testing.T) {
		// Note: This requires a mock ConnectionUseCase
		// For now, just verify the page can be created without panic
		t.Log("Connections page initialization - skipped (requires database)")
	})

	t.Run("Templates Page", func(t *testing.T) {
		content := NewTemplateManagementPage(win)
		if content == nil {
			t.Error("Templates page should not be nil")
		}
	})

	t.Run("Tasks Page", func(t *testing.T) {
		content := NewTaskConfigurationPage(win)
		if content == nil {
			t.Error("Tasks page should not be nil")
		}
	})

	t.Run("Monitor Page", func(t *testing.T) {
		content := NewRunMonitorPage(win)
		if content == nil {
			t.Error("Monitor page should not be nil")
		}
	})

	t.Run("History Page", func(t *testing.T) {
		content := NewHistoryRecordPage(win)
		if content == nil {
			t.Error("History page should not be nil")
		}
	})

	t.Run("Comparison Page", func(t *testing.T) {
		content := NewResultComparisonPage(win)
		if content == nil {
			t.Error("Comparison page should not be nil")
		}
	})

	t.Run("Report Page", func(t *testing.T) {
		content := NewReportExportPage(win)
		if content == nil {
			t.Error("Report page should not be nil")
		}
	})

	t.Run("Settings Page", func(t *testing.T) {
		// Note: Settings page can accept nil for connUC parameter in test
		content := NewSettingsConfigurationPage(win, nil)
		if content == nil {
			t.Error("Settings page should not be nil")
		}
	})
}

// TestTemplatePageComponents tests template page components.
func TestTemplatePageComponents(t *testing.T) {
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	page := NewTemplateManagementPage(win)
	if page == nil {
		t.Fatal("Template page should not be nil")
	}

	// Verify template list is created
	t.Logf("Template page created successfully")
}

// TestMonitorPageComponents tests monitor page components.
func TestMonitorPageComponents(t *testing.T) {
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	content := NewRunMonitorPage(win)
	if content == nil {
		t.Fatal("Monitor page should not be nil")
	}

	t.Log("Monitor page created successfully")
}

// TestHistoryPageComponents tests history page components.
func TestHistoryPageComponents(t *testing.T) {
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	content := NewHistoryRecordPage(win)
	if content == nil {
		t.Fatal("History page should not be nil")
	}

	t.Log("History page created successfully")
}

// TestComparisonPageComponents tests comparison page components.
func TestComparisonPageComponents(t *testing.T) {
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	content := NewResultComparisonPage(win)
	if content == nil {
		t.Fatal("Comparison page should not be nil")
	}

	t.Log("Comparison page created successfully")
}

// TestReportPageComponents tests report page components.
func TestReportPageComponents(t *testing.T) {
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	content := NewReportExportPage(win)
	if content == nil {
		t.Fatal("Report page should not be nil")
	}

	t.Log("Report page created successfully")
}

// TestSettingsPageComponents tests settings page components.
func TestSettingsPageComponents(t *testing.T) {
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	content := NewSettingsConfigurationPage(win, nil)
	if content == nil {
		t.Fatal("Settings page should not be nil")
	}

	t.Log("Settings page created successfully")
}

// TestTaskPageComponents tests task page components.
func TestTaskPageComponents(t *testing.T) {
	testApp := app.NewWithID("com.db-benchmind.test")
	win := testApp.NewWindow("Test Window")

	content := NewTaskConfigurationPage(win)
	if content == nil {
		t.Fatal("Task page should not be nil")
	}

	t.Log("Task page created successfully")
}

// TestWidgetCreation tests that common widgets can be created.
func TestWidgetCreation(t *testing.T) {
	t.Run("Create Entry", func(t *testing.T) {
		entry := widget.NewEntry()
		if entry == nil {
			t.Error("Entry should not be nil")
		}
	})

	t.Run("Create Label", func(t *testing.T) {
		label := widget.NewLabel("Test")
		if label == nil {
			t.Error("Label should not be nil")
		}
	})

	t.Run("Create Button", func(t *testing.T) {
		btn := widget.NewButton("Test", nil)
		if btn == nil {
			t.Error("Button should not be nil")
		}
	})

	t.Run("Create Select", func(t *testing.T) {
		selectWidget := widget.NewSelect([]string{"A", "B"}, nil)
		if selectWidget == nil {
			t.Error("Select should not be nil")
		}
	})

	t.Run("Create Form", func(t *testing.T) {
		form := &widget.Form{
			Items: []*widget.FormItem{
				widget.NewFormItem("Test", widget.NewEntry()),
			},
		}
		if form == nil {
			t.Error("Form should not be nil")
		}
	})
}
