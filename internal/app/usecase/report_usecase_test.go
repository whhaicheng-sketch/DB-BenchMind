// Package usecase provides unit tests for report use case.
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// TestReportUseCase_GenerateReport tests report generation.
func TestReportUseCase_GenerateReport(t *testing.T) {
	ctx := context.Background()

	// Setup
	runRepo := newMockRunRepositoryForReport()
	uc := NewReportUseCase(runRepo, nil, nil)

	// Create a test run
	now := time.Now()
	duration := 5 * time.Minute

	testRun := &execution.Run{
		ID:        "test-run-1",
		TaskID:    "task-1",
		State:     execution.StateCompleted,
		CreatedAt: now,
		StartedAt: &now,
		CompletedAt: &now,
		Duration:   &duration,
		Result: &execution.BenchmarkResult{
			TPSCalculated:      1234.56,
			LatencyAvg:         5.25,
			LatencyP95:         12.34,
			LatencyP99:         23.45,
			TotalTransactions:  10000,
			TotalQueries:       50000,
			ErrorCount:         0,
			ErrorRate:          0.0,
			TimeSeries: []execution.MetricSample{
				{
					Timestamp:   now,
					TPS:         1000,
					LatencyAvg:  5.0,
					LatencyP95:  10.0,
					LatencyP99:  20.0,
					ErrorRate:   0.0,
				},
			},
		},
	}
	runRepo.Save(ctx, testRun)

	// Test Markdown generation
	mdConfig := report.DefaultConfig(report.FormatMarkdown)
	mdReport, err := uc.GenerateReport(ctx, "test-run-1", report.FormatMarkdown, mdConfig)
	if err != nil {
		t.Fatalf("GenerateReport(Markdown) failed: %v", err)
	}

	if mdReport.Format != report.FormatMarkdown {
		t.Errorf("Format = %v, want %v", mdReport.Format, report.FormatMarkdown)
	}
	if len(mdReport.Content) == 0 {
		t.Error("Content should not be empty")
	}

	// Test JSON generation
	jsonReport, err := uc.GenerateReport(ctx, "test-run-1", report.FormatJSON, nil)
	if err != nil {
		t.Fatalf("GenerateReport(JSON) failed: %v", err)
	}

	if jsonReport.Format != report.FormatJSON {
		t.Errorf("Format = %v, want %v", jsonReport.Format, report.FormatJSON)
	}
	if len(jsonReport.Content) == 0 {
		t.Error("Content should not be empty")
	}

	// Test HTML generation
	htmlReport, err := uc.GenerateReport(ctx, "test-run-1", report.FormatHTML, nil)
	if err != nil {
		t.Fatalf("GenerateReport(HTML) failed: %v", err)
	}

	if htmlReport.Format != report.FormatHTML {
		t.Errorf("Format = %v, want %v", htmlReport.Format, report.FormatHTML)
	}
	if len(htmlReport.Content) == 0 {
		t.Error("Content should not be empty")
	}
}

// TestReportUseCase_GenerateReportErrors tests error handling.
func TestReportUseCase_GenerateReportErrors(t *testing.T) {
	ctx := context.Background()

	// Setup
	runRepo := newMockRunRepositoryForReport()
	uc := NewReportUseCase(runRepo, nil, nil)

	// Test invalid format
	_, err := uc.GenerateReport(ctx, "test-run-1", report.ReportFormat("invalid"), nil)
	if err == nil {
		t.Error("GenerateReport should fail with invalid format")
	}

	// Test run not found
	_, err = uc.GenerateReport(ctx, "nonexistent", report.FormatMarkdown, nil)
	if err == nil {
		t.Error("GenerateReport should fail with nonexistent run")
	}
}

// TestReportUseCase_ListSupportedFormats tests format listing.
func TestReportUseCase_ListSupportedFormats(t *testing.T) {
	uc := NewReportUseCase(nil, nil, nil)

	formats := uc.ListSupportedFormats()
	if len(formats) != 3 {
		t.Errorf("ListSupportedFormats() count = %d, want 3", len(formats))
	}
}

// TestReportUseCase_IsFormatSupported tests format support check.
func TestReportUseCase_IsFormatSupported(t *testing.T) {
	uc := NewReportUseCase(nil, nil, nil)

	if !uc.IsFormatSupported(report.FormatMarkdown) {
		t.Error("Markdown should be supported")
	}
	if !uc.IsFormatSupported(report.FormatJSON) {
		t.Error("JSON should be supported")
	}
	if !uc.IsFormatSupported(report.FormatHTML) {
		t.Error("HTML should be supported")
	}
	if uc.IsFormatSupported(report.ReportFormat("invalid")) {
		t.Error("Invalid format should not be supported")
	}
	if uc.IsFormatSupported(report.FormatPDF) {
		t.Error("PDF should not be supported yet")
	}
}

// TestReportUseCase_RegisterGenerator tests custom generator registration.
func TestReportUseCase_RegisterGenerator(t *testing.T) {
	uc := NewReportUseCase(nil, nil, nil)

	// Create a mock generator
	mockGen := &mockReportGenerator{format: report.FormatPDF}

	// Register it
	uc.RegisterGenerator(mockGen)

	// Verify it's registered
	if !uc.IsFormatSupported(report.FormatPDF) {
		t.Error("Custom generator should be registered")
	}

	formats := uc.ListSupportedFormats()
	if len(formats) != 4 {
		t.Errorf("After registration, should have 4 formats, got %d", len(formats))
	}
}

// Mock run repository for testing
type mockRunRepositoryForReport struct {
	runs map[string]*execution.Run
}

func newMockRunRepositoryForReport() *mockRunRepositoryForReport {
	return &mockRunRepositoryForReport{
		runs: make(map[string]*execution.Run),
	}
}

func (m *mockRunRepositoryForReport) Save(ctx context.Context, run *execution.Run) error {
	m.runs[run.ID] = run
	return nil
}

func (m *mockRunRepositoryForReport) FindByID(ctx context.Context, id string) (*execution.Run, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, ErrRunNotFound
	}
	return run, nil
}

func (m *mockRunRepositoryForReport) FindAll(ctx context.Context, opts FindOptions) ([]*execution.Run, error) {
	var result []*execution.Run
	for _, run := range m.runs {
		result = append(result, run)
	}
	return result, nil
}

func (m *mockRunRepositoryForReport) UpdateState(ctx context.Context, id string, state execution.RunState) error {
	run, ok := m.runs[id]
	if !ok {
		return ErrRunNotFound
	}
	if err := run.SetState(state); err != nil {
		return err
	}
	return nil
}

func (m *mockRunRepositoryForReport) SaveMetricSample(ctx context.Context, runID string, sample execution.MetricSample) error {
	return nil
}

func (m *mockRunRepositoryForReport) SaveLogEntry(ctx context.Context, runID string, entry LogEntry) error {
	return nil
}

func (m *mockRunRepositoryForReport) Delete(ctx context.Context, id string) error {
	delete(m.runs, id)
	return nil
}

// Mock report generator for testing
type mockReportGenerator struct {
	format report.ReportFormat
}

func (m *mockReportGenerator) Generate(data *report.GenerateContext) (*report.Report, error) {
	return &report.Report{
		Format:      m.format,
		Content:     []byte("mock report"),
		GeneratedAt: time.Now(),
		RunID:       data.RunID,
	}, nil
}

func (m *mockReportGenerator) Format() report.ReportFormat {
	return m.format
}
