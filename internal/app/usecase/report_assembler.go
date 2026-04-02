package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// ReportAssembler orchestrates all report generators to produce the final human, AI, and detailed data.
type ReportAssembler struct {
	bundleGen    *BundleGenerator
	markdownGen *MarkdownGenerator
	bottleneck   *BottleneckRulesEngine
	db           *sql.DB
	reportsDir  string
}

// NewReportAssembler creates a new ReportAssembler.
func NewReportAssembler(db *sql.DB, reportsDir string) *ReportAssembler {
	return &ReportAssembler{
		bundleGen:    NewBundleGenerator(),
		markdownGen: NewMarkdownGenerator(),
		bottleneck:   NewBottleneckRulesEngine(),
		db:         db,
		reportsDir:  reportsDir,
	}
}

// AssembleAndPersist generates all report artifacts: markdown, bundle, detailed data preview.
func (a *ReportAssembler) AssembleAndPersist(
	ctx context.Context,
	rpt *report.Report,
	run *execution.Run,
	samples []execution.MetricSample,
	adapterType string,
	monitoring *report.MonitoringData,
	prevReport *report.Report,
) (*report.HumanReportData, []byte, *report.Bundle, error) {
	// Build core KPIs
	coreKPIs := a.buildCoreKPIs(rpt, samples)

	// Build resource bottleneck section from monitoring data
	resourceBottleneck := a.buildResourceBottleneck(monitoring)

	// Run bottleneck judgment
	bottleneckInput := &BottleneckInput{
		CoreKPIs:           coreKPIs,
		ResourceBottleneck: resourceBottleneck,
		SampleCount:        len(samples),
		ErrorRate:          coreKPIs.ErrorRate,
		AnomalyCount:       0, // will be computed during bundle generation
	}
	bj := a.bottleneck.Judge(bottleneckInput)

	// Build comparison
	var comp *report.ComparisonResult
	if prevReport != nil {
		comp = CompareReports(rpt, prevReport)
	}

	// Determine overall status
	overallStatus := DetermineOverallStatus(coreKPIs, resourceBottleneck, bj)
	oneLineConclusion := GenerateOneLineConclusion(overallStatus, coreKPIs, bj)
	recommendation := GenerateRecommendations(bj)
	trendHighlights := generateTrendHighlights(comp, bj)

	// Generate AI bundle
	bundleInput := &BundleInput{
		Report:      rpt,
		Run:         run,
		Samples:     samples,
		AdapterType: adapterType,
		Monitoring:  monitoring,
		Comparison: nil, // built separately
		Bottleneck:  bj,
	}

	compressedBundle, bundle, err := a.bundleGen.GenerateAndCompress(bundleInput)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate bundle: %w", err)
	}

	// Update anomaly count in bottleneck input
	anomalyCount := len(bundle.AnomalyWindows)

	// Re-judge with anomaly count
	if anomalyCount > 0 {
		bottleneckInput.AnomalyCount = anomalyCount
		bj = a.bottleneck.Judge(bottleneckInput)
	}

	// Build detailed data preview from bundle
	detailedPreview := BuildPreviewFromBundle(bundle, compressedBundle, rpt.ID)

	// Assemble human report data
	humanData := &report.HumanReportData{
		RunMeta: report.RunMetaSection{
			RunID:         rpt.ID,
			BenchmarkType: adapterType,
			TemplateName:  rpt.TemplateName,
			TargetName:    rpt.ConnectionName,
			StartTime:     formatTime(rpt.StartedAt),
			EndTime:       formatTimePtr(rpt.EndedAt),
			DurationMs:    rpt.DurationMs,
		},
		OverallStatus:      overallStatus,
		OneLineConclusion:  oneLineConclusion,
		CoreKPIs:           coreKPIs,
		ResourceBottleneck: resourceBottleneck,
		BottleneckJudgment: bj,
		Comparison:         comp,
		TrendHighlights:   trendHighlights,
		Recommendation:    recommendation,
		DetailedDataPreview: detailedPreview,
	}

	// Generate markdown
	markdown := a.markdownGen.Generate(humanData)

	// Determine report directory
	reportDir := filepath.Join(a.reportsDir, rpt.SuiteID, rpt.ID)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return nil, nil, nil, fmt.Errorf("create report directory: %w", err)
	}

	// Persist report.md
	markdownPath := filepath.Join(reportDir, "report.md")
	if err := os.WriteFile(markdownPath, []byte(markdown), 0644); err != nil {
		return nil, nil, nil, fmt.Errorf("write report.md: %w", err)
	}

	// Persist report_bundle.json.gz (using gzip compression)
	bundlePath := filepath.Join(reportDir, "report_bundle.json.gz")
	if err := os.WriteFile(bundlePath, compressedBundle, 0644); err != nil {
		return nil, nil, nil, fmt.Errorf("write report_bundle.json.gz: %w", err)
	}

	return humanData, compressedBundle, bundle, nil
}

func (a *ReportAssembler) buildCoreKPIs(rpt *report.Report, samples []execution.MetricSample) report.CoreKPIsSection {
	tpm := rpt.TPM
	tps := rpt.TPS
	avgLatMs := rpt.LatencyAvgMs
	p95Ms := rpt.LatencyP95Ms
	p99Ms := rpt.LatencyP99Ms
	errorRate := float64(0)

	// Compute error rate from error count
	if rpt.ErrorCount > 0 && tpm > 0 {
		durationMin := float64(max(rpt.DurationMs, 1)) / 60000.0
		if durationMin > 0 {
			errorRate = float64(rpt.ErrorCount) / (tpm * durationMin)
		}
	}

	return report.CoreKPIsSection{
		TPS:           tps,
		TPM:           tpm,
		AvgLatencyMs:  avgLatMs,
		P95LatencyMs: p95Ms,
		P99LatencyMs:  p99Ms,
		ErrorRate:     errorRate,
		CPUPeak:       0, // populated from monitoring
		DiskUtilPeak: 0, // populated from monitoring
	}
}

func (a *ReportAssembler) buildResourceBottleneck(monitoring *report.MonitoringData) report.ResourceBottleneckSection {
	cpu := report.ResourceVerdict{Resource: "CPU"}
	memory := report.ResourceVerdict{Resource: "Memory"}
	disk := report.ResourceVerdict{Resource: "Disk"}
	network := report.ResourceVerdict{Resource: "Network"}

	if monitoring != nil {
		if monitoring.CPU != nil {
			cpu.Avg = monitoring.CPU.UsageAvg
			cpu.Peak = monitoring.CPU.UsageMax
			cpu.Verdict = resourceVerdict(monitoring.CPU.UsageMax, 70, 85, 95)
		}
		if monitoring.Memory != nil {
			memory.AvgMB = monitoring.Memory.UsedAvgMB
			memory.PeakMB = monitoring.Memory.UsedMaxMB
			memory.UsedPct = monitoring.Memory.UsedPercent
			memory.Verdict = resourceVerdict(monitoring.Memory.UsedPercent, 70, 85, 95)
		}
		if monitoring.Disk != nil {
			disk.Peak = max(monitoring.Disk.ReadBytesPerSec, monitoring.Disk.WriteBytesPerSec)
			disk.Verdict = resourceVerdict(disk.Peak, 70, 85, 95)
		}
		if monitoring.Network != nil {
			network.RxPeak = monitoring.Network.RxBytesPerSec
			network.TxPeak = monitoring.Network.TxBytesPerSec
			network.Verdict = "normal"
		}
	}

	return report.ResourceBottleneckSection{
		CPU: cpu, Memory: memory, Disk: disk, Network: network,
	}
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func generateTrendHighlights(comp *report.ComparisonResult, bj report.BottleneckJudgment) []string {
	var highlights []string
	if comp != nil && len(comp.Deltas) > 0 {
		for _, d := range comp.Deltas {
			if d.PctChange > 10 {
				highlights = append(highlights, fmt.Sprintf("%s increased by %.1f%%", d.Metric, d.PctChange))
			} else if d.PctChange < -10 {
				highlights = append(highlights, fmt.Sprintf("%s decreased by %.1f%%", d.Metric, -d.PctChange))
			}
		}
		if comp.TrendDirection != "stable" {
			highlights = append(highlights, fmt.Sprintf("Overall trend: %s", comp.TrendDirection))
		}
	}
	if bj.PrimaryBottleneck != report.BottleneckUnknown {
		highlights = append(highlights, fmt.Sprintf("Primary bottleneck: %s (confidence: %.0f%%)", bj.PrimaryBottleneck, bj.Confidence*100))
		}
	return highlights
}
