package usecase

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
)

func TestSuiteManifestWriter_WriteManifest(t *testing.T) {
	tests := []struct {
		name        string
		suite       *domainautobench.Suite
		wantErr     bool
		errContains string
	}{
		{
			name:    "nil suite returns error",
			suite:   nil,
			wantErr: true,
		},
		{
			name: "valid suite writes manifest",
			suite: &domainautobench.Suite{
				ID:        "suite-001",
				Name:      "Test Suite",
				Status:    domainautobench.SuiteStatusRunning,
				SelectedConnectionIDs: []string{"conn-1", "conn-2"},
				ExecutionPolicy: domainautobench.ExecutionPolicy{
					Mode:           domainautobench.ExecutionModeSerial,
					FailurePolicy:  domainautobench.FailurePolicyContinueByConnection,
					CleanupEnabled: true,
				},
				Items: []domainautobench.SuiteItem{
					{
						ID:           "item-1",
						SuiteID:      "suite-001",
						ConnectionID: "conn-1",
						ProfileType:  domainautobench.ProfileTest,
						Status:       domainautobench.SuiteItemStatusPending,
					},
					{
						ID:           "item-2",
						SuiteID:      "suite-001",
						ConnectionID: "conn-2",
						ProfileType:  domainautobench.ProfileCPU,
						Status:       domainautobench.SuiteItemStatusPending,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "suite with completed items",
			suite: &domainautobench.Suite{
				ID:        "suite-002",
				Name:      "Partial Suite",
				Status:    domainautobench.SuiteStatusRunning,
				SelectedConnectionIDs: []string{"conn-1"},
				ExecutionPolicy: domainautobench.ExecutionPolicy{
					Mode:           domainautobench.ExecutionModeSerial,
					FailurePolicy:  domainautobench.FailurePolicyContinueByConnection,
					CleanupEnabled: false,
				},
				Items: []domainautobench.SuiteItem{
					{
						ID:           "item-1",
						SuiteID:      "suite-002",
						ConnectionID: "conn-1",
						ProfileType:  domainautobench.ProfileTest,
						Status:       domainautobench.SuiteItemStatusSuccess,
						ReportID:     "report-001",
					},
					{
						ID:           "item-2",
						SuiteID:      "suite-002",
						ConnectionID: "conn-1",
						ProfileType:  domainautobench.ProfileCPU,
						Status:       domainautobench.SuiteItemStatusFailed,
						ErrorSummary: "connection timeout",
					},
					{
						ID:           "item-3",
						SuiteID:      "suite-002",
						ConnectionID: "conn-1",
						ProfileType:  domainautobench.ProfileIO,
						Status:       domainautobench.SuiteItemStatusPending,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "manifest-test-*")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			writer := NewSuiteManifestWriter(tmpDir)
			ctx := context.Background()

			path, err := writer.WriteManifest(ctx, tt.suite)

			if tt.wantErr {
				if err == nil {
					t.Errorf("WriteManifest() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("WriteManifest() unexpected error: %v", err)
				return
			}

			// Verify file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("manifest file not created at %s", path)
				return
			}

			// Verify file content is valid JSON
			data, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("read manifest file: %v", err)
				return
			}

			var manifest domainautobench.SuiteManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Errorf("unmarshal manifest: %v", err)
				return
			}

			// Verify basic structure
			if manifest.SchemaVersion != "v1" {
				t.Errorf("SchemaVersion = %q, want v1", manifest.SchemaVersion)
			}
			if manifest.SuiteID != tt.suite.ID {
				t.Errorf("SuiteID = %q, want %q", manifest.SuiteID, tt.suite.ID)
			}
			if len(manifest.Items) != len(tt.suite.Items) {
				t.Errorf("Items count = %d, want %d", len(manifest.Items), len(tt.suite.Items))
			}
		})
	}
}

func TestSuiteManifestWriter_ReadManifest(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(dir string) error
		suiteID     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "read non-existent manifest returns error",
			suiteID: "non-existent",
			wantErr: true,
		},
		{
			name: "read valid manifest",
			setupFunc: func(dir string) error {
				manifest := domainautobench.SuiteManifest{
					SchemaVersion: "v1",
					SuiteID:       "suite-001",
					GeneratedAt:   time.Now().Format(time.RFC3339),
					SuiteInfo: domainautobench.SuiteManifestInfo{
						Name:           "Test Suite",
						ExecutionMode:  domainautobench.ExecutionModeSerial,
						FailurePolicy:  domainautobench.FailurePolicyContinueByConnection,
						CleanupEnabled: true,
					},
					SelectedConnectionIDs: []string{"conn-1"},
					Items: []domainautobench.SuiteManifestItem{
						{
							ID:           "item-1",
							ConnectionID: "conn-1",
							ProfileType:  domainautobench.ProfileTest,
							Status:       "pending",
						},
					},
					Statistics: domainautobench.SuiteManifestStatistics{
						TotalItems: 1,
						Pending:    1,
					},
				}
				data, _ := json.MarshalIndent(manifest, "", "  ")
				suiteDir := filepath.Join(dir, "suite-001")
				os.MkdirAll(suiteDir, 0755)
				return os.WriteFile(filepath.Join(suiteDir, "suite_manifest.json"), data, 0644)
			},
			suiteID: "suite-001",
			wantErr: false,
		},
		{
			name: "read invalid JSON returns error",
			setupFunc: func(dir string) error {
				suiteDir := filepath.Join(dir, "suite-002")
				os.MkdirAll(suiteDir, 0755)
				return os.WriteFile(filepath.Join(suiteDir, "suite_manifest.json"), []byte("invalid json"), 0644)
			},
			suiteID: "suite-002",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "manifest-read-test-*")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(tmpDir); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}

			writer := NewSuiteManifestWriter(tmpDir)
			ctx := context.Background()

			manifest, err := writer.ReadManifest(ctx, tt.suiteID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReadManifest() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ReadManifest() unexpected error: %v", err)
				return
			}

			if manifest == nil {
				t.Error("ReadManifest() returned nil manifest")
				return
			}

			if manifest.SuiteID != tt.suiteID {
				t.Errorf("SuiteID = %q, want %q", manifest.SuiteID, tt.suiteID)
			}
		})
	}
}

func TestSuiteManifestWriter_UpdateManifestItem(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(dir string) error
		suiteID     string
		itemID      string
		updateFn    func(*domainautobench.SuiteManifestItem) error
		wantErr     bool
		verifyFunc  func(t *testing.T, manifest *domainautobench.SuiteManifest)
	}{
		{
			name:    "update non-existent manifest returns error",
			suiteID: "non-existent",
			itemID:  "item-1",
			updateFn: func(item *domainautobench.SuiteManifestItem) error {
				item.Status = "success"
				return nil
			},
			wantErr: true,
		},
		{
			name: "update non-existent item returns error",
			setupFunc: func(dir string) error {
				manifest := domainautobench.SuiteManifest{
					SchemaVersion: "v1",
					SuiteID:       "suite-001",
					GeneratedAt:   time.Now().Format(time.RFC3339),
					Items: []domainautobench.SuiteManifestItem{
						{ID: "item-1", Status: "pending"},
					},
				}
				data, _ := json.MarshalIndent(manifest, "", "  ")
				suiteDir := filepath.Join(dir, "suite-001")
				os.MkdirAll(suiteDir, 0755)
				return os.WriteFile(filepath.Join(suiteDir, "suite_manifest.json"), data, 0644)
			},
			suiteID: "suite-001",
			itemID:  "non-existent-item",
			updateFn: func(item *domainautobench.SuiteManifestItem) error {
				item.Status = "success"
				return nil
			},
			wantErr: true,
		},
		{
			name: "update item status to success",
			setupFunc: func(dir string) error {
				manifest := domainautobench.SuiteManifest{
					SchemaVersion: "v1",
					SuiteID:       "suite-002",
					GeneratedAt:   time.Now().Format(time.RFC3339),
					Items: []domainautobench.SuiteManifestItem{
						{ID: "item-1", ConnectionID: "conn-1", ProfileType: domainautobench.ProfileTest, Status: "running"},
						{ID: "item-2", ConnectionID: "conn-1", ProfileType: domainautobench.ProfileCPU, Status: "pending"},
					},
					Statistics: domainautobench.SuiteManifestStatistics{
						TotalItems: 2,
						Pending:    1,
						Running:    1,
					},
				}
				data, _ := json.MarshalIndent(manifest, "", "  ")
				suiteDir := filepath.Join(dir, "suite-002")
				os.MkdirAll(suiteDir, 0755)
				return os.WriteFile(filepath.Join(suiteDir, "suite_manifest.json"), data, 0644)
			},
			suiteID: "suite-002",
			itemID:  "item-1",
			updateFn: func(item *domainautobench.SuiteManifestItem) error {
				item.Status = "success"
				item.ReportID = "report-001"
				item.EndedAt = time.Now().Format(time.RFC3339)
				return nil
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, manifest *domainautobench.SuiteManifest) {
				// Find the updated item
				var found bool
				for _, item := range manifest.Items {
					if item.ID == "item-1" {
						found = true
						if item.Status != "success" {
							t.Errorf("item status = %q, want success", item.Status)
						}
						if item.ReportID != "report-001" {
							t.Errorf("item report_id = %q, want report-001", item.ReportID)
						}
						break
					}
				}
				if !found {
					t.Error("item-1 not found in manifest")
				}
				// Verify statistics updated
				if manifest.Statistics.Success != 1 {
					t.Errorf("Statistics.Success = %d, want 1", manifest.Statistics.Success)
				}
				if manifest.Statistics.Running != 0 {
					t.Errorf("Statistics.Running = %d, want 0", manifest.Statistics.Running)
				}
			},
		},
		{
			name: "update item status to failed with error",
			setupFunc: func(dir string) error {
				manifest := domainautobench.SuiteManifest{
					SchemaVersion: "v1",
					SuiteID:       "suite-003",
					GeneratedAt:   time.Now().Format(time.RFC3339),
					Items: []domainautobench.SuiteManifestItem{
						{ID: "item-1", ConnectionID: "conn-1", ProfileType: domainautobench.ProfileTest, Status: "running"},
					},
					Statistics: domainautobench.SuiteManifestStatistics{
						TotalItems: 1,
						Running:    1,
					},
				}
				data, _ := json.MarshalIndent(manifest, "", "  ")
				suiteDir := filepath.Join(dir, "suite-003")
				os.MkdirAll(suiteDir, 0755)
				return os.WriteFile(filepath.Join(suiteDir, "suite_manifest.json"), data, 0644)
			},
			suiteID: "suite-003",
			itemID:  "item-1",
			updateFn: func(item *domainautobench.SuiteManifestItem) error {
				item.Status = "failed"
				item.ErrorMessage = "connection timeout"
				item.EndedAt = time.Now().Format(time.RFC3339)
				return nil
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, manifest *domainautobench.SuiteManifest) {
				if manifest.Statistics.Failed != 1 {
					t.Errorf("Statistics.Failed = %d, want 1", manifest.Statistics.Failed)
				}
				for _, item := range manifest.Items {
					if item.ID == "item-1" {
						if item.ErrorMessage != "connection timeout" {
							t.Errorf("ErrorMessage = %q, want 'connection timeout'", item.ErrorMessage)
						}
						break
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "manifest-update-test-*")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(tmpDir); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}

			writer := NewSuiteManifestWriter(tmpDir)
			ctx := context.Background()

			err = writer.UpdateManifestItem(ctx, tt.suiteID, tt.itemID, tt.updateFn)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateManifestItem() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateManifestItem() unexpected error: %v", err)
				return
			}

			// Read back and verify
			manifest, err := writer.ReadManifest(ctx, tt.suiteID)
			if err != nil {
				t.Fatalf("read manifest for verification: %v", err)
			}

			if tt.verifyFunc != nil {
				tt.verifyFunc(t, manifest)
			}
		})
	}
}

func TestSuiteManifestWriter_CalculateStatistics(t *testing.T) {
	tests := []struct {
		name     string
		items    []domainautobench.SuiteManifestItem
		expected domainautobench.SuiteManifestStatistics
	}{
		{
			name:     "empty items",
			items:    []domainautobench.SuiteManifestItem{},
			expected: domainautobench.SuiteManifestStatistics{TotalItems: 0},
		},
		{
			name: "all pending",
			items: []domainautobench.SuiteManifestItem{
				{Status: "pending"},
				{Status: "pending"},
				{Status: "pending"},
			},
			expected: domainautobench.SuiteManifestStatistics{
				TotalItems: 3,
				Pending:    3,
			},
		},
		{
			name: "mixed statuses",
			items: []domainautobench.SuiteManifestItem{
				{Status: "success"},
				{Status: "failed"},
				{Status: "skipped"},
				{Status: "running"},
				{Status: "pending"},
			},
			expected: domainautobench.SuiteManifestStatistics{
				TotalItems: 5,
				Success:    1,
				Failed:     1,
				Skipped:    1,
				Running:    1,
				Pending:    1,
			},
		},
		{
			name: "all success",
			items: []domainautobench.SuiteManifestItem{
				{Status: "success"},
				{Status: "success"},
			},
			expected: domainautobench.SuiteManifestStatistics{
				TotalItems: 2,
				Success:    2,
			},
		},
		{
			name: "running phases count as running",
			items: []domainautobench.SuiteManifestItem{
				{Status: "running"},
				{Status: "validating"},
				{Status: "preparing"},
				{Status: "cleaning"},
			},
			expected: domainautobench.SuiteManifestStatistics{
				TotalItems: 4,
				Running:    4,
			},
		},
	}

	writer := &SuiteManifestWriter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.calculateStatistics(tt.items)

			if result.TotalItems != tt.expected.TotalItems {
				t.Errorf("TotalItems = %d, want %d", result.TotalItems, tt.expected.TotalItems)
			}
			if result.Pending != tt.expected.Pending {
				t.Errorf("Pending = %d, want %d", result.Pending, tt.expected.Pending)
			}
			if result.Running != tt.expected.Running {
				t.Errorf("Running = %d, want %d", result.Running, tt.expected.Running)
			}
			if result.Success != tt.expected.Success {
				t.Errorf("Success = %d, want %d", result.Success, tt.expected.Success)
			}
			if result.Failed != tt.expected.Failed {
				t.Errorf("Failed = %d, want %d", result.Failed, tt.expected.Failed)
			}
			if result.Skipped != tt.expected.Skipped {
				t.Errorf("Skipped = %d, want %d", result.Skipped, tt.expected.Skipped)
			}
		})
	}
}

func TestSuite_ToManifest(t *testing.T) {
	now := time.Now()
	suite := domainautobench.Suite{
		ID:        "suite-001",
		Name:      "Test Suite",
		Status:    domainautobench.SuiteStatusRunning,
		SelectedConnectionIDs: []string{"conn-1", "conn-2"},
		ExecutionPolicy: domainautobench.ExecutionPolicy{
			Mode:           domainautobench.ExecutionModeSerial,
			FailurePolicy:  domainautobench.FailurePolicyContinueByConnection,
			CleanupEnabled: true,
		},
		Items: []domainautobench.SuiteItem{
			{
				ID:           "item-1",
				SuiteID:      "suite-001",
				ConnectionID: "conn-1",
				DatabaseType: "mysql",
				ProfileType:  domainautobench.ProfileTest,
				TemplateID:   "tmpl-001",
				Status:       domainautobench.SuiteItemStatusSuccess,
				ReportID:     "report-001",
				StartedAt:    &now,
				EndedAt:      &now,
			},
			{
				ID:           "item-2",
				SuiteID:      "suite-001",
				ConnectionID: "conn-2",
				ProfileType:  domainautobench.ProfileCPU,
				Status:       domainautobench.SuiteItemStatusFailed,
				ErrorSummary: "timeout",
				EndedAt:      &now,
			},
		},
	}

	manifest := suite.ToManifest()

	// Verify basic fields
	if manifest.SchemaVersion != "v1" {
		t.Errorf("SchemaVersion = %q, want v1", manifest.SchemaVersion)
	}
	if manifest.SuiteID != suite.ID {
		t.Errorf("SuiteID = %q, want %q", manifest.SuiteID, suite.ID)
	}
	if manifest.SuiteInfo.Name != suite.Name {
		t.Errorf("SuiteInfo.Name = %q, want %q", manifest.SuiteInfo.Name, suite.Name)
	}

	// Verify connection IDs copied
	if len(manifest.SelectedConnectionIDs) != len(suite.SelectedConnectionIDs) {
		t.Errorf("SelectedConnectionIDs length = %d, want %d", len(manifest.SelectedConnectionIDs), len(suite.SelectedConnectionIDs))
	}

	// Verify items converted
	if len(manifest.Items) != len(suite.Items) {
		t.Errorf("Items length = %d, want %d", len(manifest.Items), len(suite.Items))
	}

	// Verify first item
	if manifest.Items[0].ID != "item-1" {
		t.Errorf("Items[0].ID = %q, want item-1", manifest.Items[0].ID)
	}
	if manifest.Items[0].Status != "success" {
		t.Errorf("Items[0].Status = %q, want success", manifest.Items[0].Status)
	}
	if manifest.Items[0].ReportID != "report-001" {
		t.Errorf("Items[0].ReportID = %q, want report-001", manifest.Items[0].ReportID)
	}

	// Verify statistics
	if manifest.Statistics.TotalItems != 2 {
		t.Errorf("Statistics.TotalItems = %d, want 2", manifest.Statistics.TotalItems)
	}
	if manifest.Statistics.Success != 1 {
		t.Errorf("Statistics.Success = %d, want 1", manifest.Statistics.Success)
	}
	if manifest.Statistics.Failed != 1 {
		t.Errorf("Statistics.Failed = %d, want 1", manifest.Statistics.Failed)
	}
}

func TestSuiteManifest_ToJSON(t *testing.T) {
	manifest := domainautobench.SuiteManifest{
		SchemaVersion: "v1",
		SuiteID:       "suite-001",
		GeneratedAt:   "2026-03-26T10:00:00Z",
		SuiteInfo: domainautobench.SuiteManifestInfo{
			Name:           "Test Suite",
			ExecutionMode:  domainautobench.ExecutionModeSerial,
			FailurePolicy:  domainautobench.FailurePolicyContinueByConnection,
			CleanupEnabled: true,
		},
		SelectedConnectionIDs: []string{"conn-1"},
		Items: []domainautobench.SuiteManifestItem{
			{
				ID:           "item-1",
				ConnectionID: "conn-1",
				ProfileType:  domainautobench.ProfileTest,
				Status:       "pending",
			},
		},
		Statistics: domainautobench.SuiteManifestStatistics{
			TotalItems: 1,
			Pending:    1,
		},
	}

	data, err := manifest.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("unmarshal result: %v", err)
	}

	// Verify schema_version is present
	if v, ok := parsed["schema_version"].(string); !ok || v != "v1" {
		t.Errorf("schema_version not found or wrong value")
	}

	// Verify it's formatted (contains newlines)
	if len(data) > 0 && data[0] == '{' && !contains(string(data), "\n") {
		t.Error("JSON should be formatted with indentation")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
