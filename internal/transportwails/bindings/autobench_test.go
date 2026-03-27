package bindings

import (
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
)

func TestAutoBenchBinding_CreateSuite(t *testing.T) {
	suitesUC := usecase.NewAutoBenchSuiteUseCase()
	binding := NewAutoBenchBinding(suitesUC, nil)

	tests := []struct {
		name    string
		req     AutoBenchCreateSuiteRequest
		wantErr bool
	}{
		{
			name: "valid suite creation",
			req: AutoBenchCreateSuiteRequest{
				Name:          "Test Suite",
				ConnectionIDs: []string{"conn-1", "conn-2"},
				ProfileTypes:  []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "empty connection IDs",
			req: AutoBenchCreateSuiteRequest{
				Name:          "Test Suite",
				ConnectionIDs: []string{},
				ProfileTypes:  []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "default profiles",
			req: AutoBenchCreateSuiteRequest{
				Name:          "Test Suite",
				ConnectionIDs: []string{"conn-1"},
				ProfileTypes:  []string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binding.CreateSuite(tt.req)
			if (result.Error != "") != tt.wantErr {
				t.Errorf("CreateSuite() error = %v, wantErr %v", result.Error, tt.wantErr)
			}
			if !tt.wantErr && result.SuiteID == "" {
				t.Error("CreateSuite() returned empty suite ID")
			}
		})
	}
}

func TestAutoBenchBinding_GetSuiteStatus(t *testing.T) {
	suitesUC := usecase.NewAutoBenchSuiteUseCase()
	binding := NewAutoBenchBinding(suitesUC, nil)

	// Create a suite first
	createResult := binding.CreateSuite(AutoBenchCreateSuiteRequest{
		Name:          "Test Suite",
		ConnectionIDs: []string{"conn-1"},
		ProfileTypes:  []string{"test"},
	})
	if createResult.Error != "" {
		t.Fatalf("CreateSuite failed: %s", createResult.Error)
	}

	// Get status
	statusResult := binding.GetSuiteStatus(createResult.SuiteID)
	if statusResult.Error != "" {
		t.Errorf("GetSuiteStatus() error = %v", statusResult.Error)
	}
	if statusResult.Status != "draft" {
		t.Errorf("GetSuiteStatus() status = %v, want draft", statusResult.Status)
	}
	if statusResult.TotalItems != 1 {
		t.Errorf("GetSuiteStatus() totalItems = %v, want 1", statusResult.TotalItems)
	}
}

func TestAutoBenchBinding_GetSuiteStatus_NotFound(t *testing.T) {
	suitesUC := usecase.NewAutoBenchSuiteUseCase()
	binding := NewAutoBenchBinding(suitesUC, nil)

	result := binding.GetSuiteStatus("non-existent-id")
	if result.Error == "" {
		t.Error("GetSuiteStatus() expected error for non-existent suite")
	}
}

func TestAutoBenchBinding_GetExecutionPlan(t *testing.T) {
	suitesUC := usecase.NewAutoBenchSuiteUseCase()
	binding := NewAutoBenchBinding(suitesUC, nil)

	// Create a suite first
	createResult := binding.CreateSuite(AutoBenchCreateSuiteRequest{
		Name:          "Test Suite",
		ConnectionIDs: []string{"conn-1", "conn-2"},
		ProfileTypes:  []string{"test", "cpu_bound"},
	})
	if createResult.Error != "" {
		t.Fatalf("CreateSuite failed: %s", createResult.Error)
	}

	// Get execution plan
	planResult := binding.GetExecutionPlan(createResult.SuiteID)
	if planResult.Error != "" {
		t.Errorf("GetExecutionPlan() error = %v", planResult.Error)
	}
	if !planResult.Sequential {
		t.Error("GetExecutionPlan() expected sequential to be true")
	}
	if len(planResult.Items) != 4 { // 2 connections x 2 profiles
		t.Errorf("GetExecutionPlan() items = %v, want 4", len(planResult.Items))
	}
}

func TestAutoBenchBinding_ListProfiles(t *testing.T) {
	suitesUC := usecase.NewAutoBenchSuiteUseCase()
	binding := NewAutoBenchBinding(suitesUC, nil)

	result := binding.ListProfiles()
	if len(result.Profiles) != 3 {
		t.Errorf("ListProfiles() profiles = %v, want 3", len(result.Profiles))
	}
}

func TestAutoBenchBinding_StartSuite_NoRunner(t *testing.T) {
	suitesUC := usecase.NewAutoBenchSuiteUseCase()
	binding := NewAutoBenchBinding(suitesUC, nil) // No runner

	// Create a suite first
	createResult := binding.CreateSuite(AutoBenchCreateSuiteRequest{
		Name:          "Test Suite",
		ConnectionIDs: []string{"conn-1"},
		ProfileTypes:  []string{"test"},
	})
	if createResult.Error != "" {
		t.Fatalf("CreateSuite failed: %s", createResult.Error)
	}

	// Start suite - should succeed (runner check is done in RunSuite)
	result := binding.StartSuite(createResult.SuiteID)
	if result.Error != "" {
		t.Errorf("StartSuite() error = %v", result.Error)
	}
	if !result.Success {
		t.Error("StartSuite() expected success")
	}
}
