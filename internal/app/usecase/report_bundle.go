// Package usecase provides report bundle generation logic.
package usecase

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

const (
	// MaxBundleCompressedBytes is the hard limit for compressed bundle size (1MB).
	MaxBundleCompressedBytes = 1 * 1024 * 1024

	// DefaultDownsampleBuckets is the default number of buckets for downsampling.
	DefaultDownsampleBuckets = 60

	// WindowFrontPct is the front window start percentage.
	WindowFrontPct = 0.0
	// WindowFrontEndPct is the front window end percentage.
	WindowFrontEndPct = 0.15
	// WindowMiddleStartPct is the middle window start percentage.
	WindowMiddleStartPct = 0.45
	// WindowMiddleEndPct is the middle window end percentage.
	WindowMiddleEndPct = 0.55
	// WindowTailStartPct is the tail window start percentage.
	WindowTailStartPct = 0.85
	// WindowTailEndPct is the tail window end percentage.
	WindowTailEndPct = 1.0

	// MaxRawSamplesNormal is the max normal samples to keep.
	MaxRawSamplesNormal = 20
	// MaxRawSamplesAnomaly is the max anomaly samples to keep.
	MaxRawSamplesAnomaly = 50
)

// BundleGenerator creates AI analysis bundles from benchmark data.
type BundleGenerator struct{}

// NewBundleGenerator creates a new BundleGenerator.
func NewBundleGenerator() *BundleGenerator {
	return &BundleGenerator{}
}

// BundleInput contains all input data needed for bundle generation.
type BundleInput struct {
	Report       *report.Report
	Run          *execution.Run
	Samples      []execution.MetricSample
	AdapterType  string
	Monitoring   *report.MonitoringData
	Comparison   *report.BundleComparison
	Bottleneck   report.BottleneckJudgment
}

// Generate creates an AI Bundle from the input data.
func (g *BundleGenerator) Generate(input *BundleInput) (*report.Bundle, error) {
	if input.Report == nil {
		return nil, fmt.Errorf("report is required")
	}

	rpt := input.Report
	run := input.Run

	bundle := &report.Bundle{
		SchemaVersion: report.BundleSchemaVersion,
		RunMeta: report.BundleRunMeta{
			RunID:         rpt.ID,
			BenchmarkType: input.AdapterType,
			TemplateName:  rpt.TemplateName,
			TargetName:    rpt.ConnectionName,
			DatabaseType:  rpt.DatabaseType,
			DurationMs:    rpt.DurationMs,
		},
	}

	if !rpt.StartedAt.IsZero() {
		bundle.RunMeta.StartTime = rpt.StartedAt.Format(time.RFC3339)
	}
	if rpt.EndedAt != nil {
		bundle.RunMeta.EndTime = rpt.EndedAt.Format(time.RFC3339)
	}

	// Build benchmark summary
	bundle.BenchmarkSummary = g.buildBenchmarkSummary(rpt, run, input.Samples)

	// Build resource summary
	bundle.ResourceSummary = g.buildResourceSummary(input.Monitoring, input.Samples)

	// Downsample time series
	bundle.TimeseriesDownsampled = g.buildTimeseries(input.Samples, input.Monitoring)

	// Retained windows
	bundle.RetainedWindows = g.buildRetainedWindows(input.Samples)

	// Anomaly detection
	bundle.AnomalyWindows = g.detectAnomalies(input.Samples, input.Monitoring)

	// Comparison
	bundle.Comparison = input.Comparison

	// Raw samples (high-value only)
	bundle.RawSamples = g.selectRawSamples(input.Samples, bundle.AnomalyWindows)

	// Phase breakdown
	if run != nil {
		bundle.PhaseBreakdown = g.buildPhaseBreakdown(run)
	}

	// AI meta (will be updated after compression)
	bundle.AIMeta = report.BundleAIMeta{
		SamplingPolicy: "front_middle_tail_anomaly",
		GeneratedAt:    time.Now().Format(time.RFC3339),
	}

	return bundle, nil
}

// GenerateAndCompress generates a bundle and compresses it, enforcing the 1MB limit.
func (g *BundleGenerator) GenerateAndCompress(input *BundleInput) ([]byte, *report.Bundle, error) {
	bundle, err := g.Generate(input)
	if err != nil {
		return nil, nil, fmt.Errorf("generate bundle: %w", err)
	}

	// Try compressing
	compressed, err := compressBundle(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("compress bundle: %w", err)
	}

	// If within limit, we're done
	if len(compressed) <= MaxBundleCompressedBytes {
		bundle.AIMeta.CompressedSizeBytes = int64(len(compressed))
		bundle.AIMeta.TruncationApplied = false
		return compressed, bundle, nil
	}

	// Apply truncation layers
	samplingPolicy := "front_middle_tail_anomaly"

	// Layer 1: Reduce timeseries bucket count
	bundle.TimeseriesDownsampled = g.buildTimeseries(input.Samples, input.Monitoring)
	for i := 0; i < len(bundle.TimeseriesDownsampled.Metrics); i++ {
		m := &bundle.TimeseriesDownsampled.Metrics[i]
		if len(m.Buckets) > 20 {
			m.Buckets = m.Buckets[:20]
		}
	}

	compressed, err = compressBundle(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("compress bundle (layer 1): %w", err)
	}
	if len(compressed) <= MaxBundleCompressedBytes {
		bundle.AIMeta.CompressedSizeBytes = int64(len(compressed))
		bundle.AIMeta.TruncationApplied = true
		bundle.AIMeta.TruncationReason = "reduced_timeseries_buckets"
		bundle.AIMeta.SamplingPolicy = samplingPolicy
		return compressed, bundle, nil
	}

	// Layer 2: Reduce raw samples
	if len(bundle.RawSamples) > 10 {
		anomalySamples := make([]report.BundleRawSample, 0)
		normalSamples := make([]report.BundleRawSample, 0)
		for _, s := range bundle.RawSamples {
			if s.Type == "anomaly" {
				anomalySamples = append(anomalySamples, s)
			} else {
				normalSamples = append(normalSamples, s)
			}
		}
		// Keep all anomaly + fewer normal
		keepNormal := 10 - len(anomalySamples)
		if keepNormal < 0 {
			keepNormal = 0
		}
		if keepNormal > len(normalSamples) {
			keepNormal = len(normalSamples)
		}
		bundle.RawSamples = append(anomalySamples, normalSamples[:keepNormal]...)
		samplingPolicy = "reduced_raw_samples"
	}

	compressed, err = compressBundle(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("compress bundle (layer 2): %w", err)
	}
	if len(compressed) <= MaxBundleCompressedBytes {
		bundle.AIMeta.CompressedSizeBytes = int64(len(compressed))
		bundle.AIMeta.TruncationApplied = true
		bundle.AIMeta.TruncationReason = "reduced_raw_samples"
		bundle.AIMeta.SamplingPolicy = samplingPolicy
		return compressed, bundle, nil
	}

	// Layer 3: Remove tool_specific and phase breakdown
	bundle.BenchmarkSummary.ToolSpecific = nil
	bundle.PhaseBreakdown = report.BundlePhaseBreakdown{}

	compressed, err = compressBundle(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("compress bundle (layer 3): %w", err)
	}
	if len(compressed) <= MaxBundleCompressedBytes {
		bundle.AIMeta.CompressedSizeBytes = int64(len(compressed))
		bundle.AIMeta.TruncationApplied = true
		bundle.AIMeta.TruncationReason = "removed_tool_specific_and_phases"
		bundle.AIMeta.SamplingPolicy = samplingPolicy
		return compressed, bundle, nil
	}

	// Layer 4: Aggressive - keep only summary + anomalies + minimal windows
	survivingWindows := make([]report.BundleWindow, 0)
	for _, w := range bundle.RetainedWindows {
		w.Summary.SampleCount = 0
		survivingWindows = append(survivingWindows, w)
	}
	bundle.RetainedWindows = survivingWindows
	bundle.RawSamples = bundle.RawSamples[:0]

	compressed, err = compressBundle(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("compress bundle (layer 4): %w", err)
	}

	// Final check - if still too large, something is fundamentally wrong
	bundle.AIMeta.CompressedSizeBytes = int64(len(compressed))
	bundle.AIMeta.TruncationApplied = true
	bundle.AIMeta.TruncationReason = "aggressive_truncation"
	bundle.AIMeta.SamplingPolicy = samplingPolicy

	return compressed, bundle, nil
}

func (g *BundleGenerator) buildBenchmarkSummary(rpt *report.Report, run *execution.Run, samples []execution.MetricSample) report.BundleBenchmarkSummary {
	summary := report.BundleBenchmarkSummary{
		Throughput: report.BundleThroughputData{
			TPS: rpt.TPS,
			TPM: rpt.TPM,
		},
		Latency: report.BundleLatencyData{
			AvgMs: rpt.LatencyAvgMs,
			P95Ms: rpt.LatencyP95Ms,
			P99Ms: rpt.LatencyP99Ms,
		},
		Errors: report.BundleErrorData{
			ErrorCount: rpt.ErrorCount,
		},
	}

	if rpt.ErrorCount > 0 && rpt.TPM > 0 {
		summary.Errors.ErrorRate = float64(rpt.ErrorCount) / (rpt.TPM * float64(max(rpt.DurationMs, 1)) / 60000.0)
	}

	if run != nil && run.Result != nil {
		summary.Latency.MinMs = run.Result.LatencyMin
		summary.Latency.MaxMs = run.Result.LatencyMax
		if run.Result.Duration > 0 {
			summary.Throughput.QPS = float64(run.Result.TotalQueries) / run.Result.Duration.Seconds()
		}
	}

	return summary
}

func (g *BundleGenerator) buildResourceSummary(monitoring *report.MonitoringData, samples []execution.MetricSample) report.BundleResourceSummary {
	summary := report.BundleResourceSummary{}

	// Default verdicts
	summary.CPU.Verdict = "normal"
	summary.Memory.Verdict = "normal"
	summary.Disk.Verdict = "normal"
	summary.Network.Verdict = "normal"

	if monitoring != nil {
		if monitoring.CPU != nil {
			summary.CPU.Avg = monitoring.CPU.UsageAvg
			summary.CPU.Peak = monitoring.CPU.UsageMax
			summary.CPU.Verdict = resourceVerdict(monitoring.CPU.UsageMax, 70, 85, 95)
		}
		if monitoring.Memory != nil {
			summary.Memory.AvgMB = monitoring.Memory.UsedAvgMB
			summary.Memory.PeakMB = monitoring.Memory.UsedMaxMB
			summary.Memory.UsedPct = monitoring.Memory.UsedPercent
			summary.Memory.Verdict = resourceVerdict(monitoring.Memory.UsedPercent, 70, 85, 95)
		}
		if monitoring.Disk != nil {
			summary.Disk.UtilPeak = max(monitoring.Disk.ReadBytesPerSec, monitoring.Disk.WriteBytesPerSec)
			summary.Disk.ReadIOPSPeak = monitoring.Disk.ReadBytesPerSec
			summary.Disk.WriteIOPSPeak = monitoring.Disk.WriteBytesPerSec
			summary.Disk.Verdict = "normal"
		}
		if monitoring.Network != nil {
			summary.Network.RxPeak = monitoring.Network.RxBytesPerSec
			summary.Network.TxPeak = monitoring.Network.TxBytesPerSec
			summary.Network.Verdict = "normal"
		}
	}

	return summary
}

func (g *BundleGenerator) buildTimeseries(samples []execution.MetricSample, monitoring *report.MonitoringData) report.BundleTimeseries {
	ts := report.BundleTimeseries{}

	if len(samples) == 0 {
		return ts
	}

	// Build TPS series
	tpsBuckets := downsample(samplesToTPSPoints(samples), DefaultDownsampleBuckets)
	if len(tpsBuckets) > 0 {
		ts.Metrics = append(ts.Metrics, report.BundleTimeseriesMetric{
			Name:       "tps",
			Buckets:    tpsBuckets,
			PointCount: len(samples),
		})
	}

	// Build latency series
	latBuckets := downsample(samplesToLatencyPoints(samples), DefaultDownsampleBuckets)
	if len(latBuckets) > 0 {
		ts.Metrics = append(ts.Metrics, report.BundleTimeseriesMetric{
			Name:       "latency_p95",
			Buckets:    latBuckets,
			PointCount: len(samples),
		})
	}

	// Build CPU series from monitoring
	if monitoring != nil && len(monitoring.TimeSeries) > 0 {
		cpuBuckets := downsample(monitoringToCPUPoints(monitoring.TimeSeries), DefaultDownsampleBuckets)
		if len(cpuBuckets) > 0 {
			ts.Metrics = append(ts.Metrics, report.BundleTimeseriesMetric{
				Name:       "cpu_usage",
				Buckets:    cpuBuckets,
				PointCount: len(monitoring.TimeSeries),
			})
		}
	}

	return ts
}

func (g *BundleGenerator) buildRetainedWindows(samples []execution.MetricSample) []report.BundleWindow {
	if len(samples) < 3 {
		return nil
	}

	n := len(samples)
	windows := make([]report.BundleWindow, 0, 3)

	// Front window: 0%-15%
	frontEnd := max(1, int(float64(n)*WindowFrontEndPct))
	windows = append(windows, report.BundleWindow{
		Name:     "front",
		StartPct: WindowFrontPct,
		EndPct:   WindowFrontEndPct,
		Summary:  computeWindowSummary(samples[:frontEnd]),
	})

	// Middle window: 45%-55%
	midStart := max(frontEnd, int(float64(n)*WindowMiddleStartPct))
	midEnd := max(midStart+1, int(float64(n)*WindowMiddleEndPct))
	windows = append(windows, report.BundleWindow{
		Name:     "middle",
		StartPct: WindowMiddleStartPct,
		EndPct:   WindowMiddleEndPct,
		Summary:  computeWindowSummary(samples[midStart:midEnd]),
	})

	// Tail window: 85%-100%
	tailStart := max(midEnd, int(float64(n)*WindowTailStartPct))
	windows = append(windows, report.BundleWindow{
		Name:     "tail",
		StartPct: WindowTailStartPct,
		EndPct:   WindowTailEndPct,
		Summary:  computeWindowSummary(samples[tailStart:]),
	})

	return windows
}

func (g *BundleGenerator) detectAnomalies(samples []execution.MetricSample, monitoring *report.MonitoringData) []report.BundleAnomalyWindow {
	var anomalies []report.BundleAnomalyWindow

	if len(samples) < 5 {
		return anomalies
	}

	// Compute baseline stats
	tpsVals := make([]float64, len(samples))
	latVals := make([]float64, 0, len(samples))
	for i, s := range samples {
		tpsVals[i] = s.TPS
		if s.LatencyAvg > 0 {
			latVals = append(latVals, s.LatencyAvg)
		}
	}

	tpsAvg, _, _ := minMaxAvg(tpsVals)

	// Detect TPS drops (below 50% of average)
	tpsThreshold := tpsAvg * 0.5
	if tpsAvg > 0 {
		for i := 0; i < len(samples); i++ {
			if samples[i].TPS > 0 && samples[i].TPS < tpsThreshold {
				start := i
				for i < len(samples) && samples[i].TPS < tpsThreshold {
					i++
				}
				severity := "medium"
				minTPS := tpsVals[start]
				for j := start; j < i; j++ {
					if tpsVals[j] < minTPS {
						minTPS = tpsVals[j]
					}
				}
				if minTPS < tpsAvg*0.25 {
					severity = "critical"
				} else if minTPS < tpsAvg*0.35 {
					severity = "high"
				}
				anomalies = append(anomalies, report.BundleAnomalyWindow{
					Type:      "tps_drop",
					StartIdx:  start,
					EndIdx:    i - 1,
					Severity:  severity,
					Value:     minTPS,
					Threshold: tpsThreshold,
					Summary:   fmt.Sprintf("TPS dropped to %.1f (%.0f%% below avg %.1f)", minTPS, (1-minTPS/tpsAvg)*100, tpsAvg),
				})
			}
		}
	}

	// Detect latency spikes (above 5x median, robust to outliers)
	if len(latVals) > 0 {
		sort.Float64s(latVals)
		latMedian := latVals[len(latVals)/2]
		latThreshold := latMedian * 5.0
		if latThreshold > 0 {
			for i := 0; i < len(samples); i++ {
				if samples[i].LatencyAvg > latThreshold {
					start := i
					maxLat := samples[i].LatencyAvg
					for i < len(samples) && samples[i].LatencyAvg > latThreshold {
						if samples[i].LatencyAvg > maxLat {
							maxLat = samples[i].LatencyAvg
						}
						i++
					}
					severity := "medium"
					if maxLat > latMedian*20 {
						severity = "critical"
					} else if maxLat > latMedian*10 {
						severity = "high"
					}
					anomalies = append(anomalies, report.BundleAnomalyWindow{
						Type:      "latency_spike",
						StartIdx:  start,
						EndIdx:    i - 1,
						Severity:  severity,
						Value:     maxLat,
						Threshold: latThreshold,
						Summary:   fmt.Sprintf("Latency spiked to %.1fms (%.1fx above median %.1fms)", maxLat, maxLat/latMedian, latMedian),
					})
				}
			}
		}
	}

	// Detect error bursts
	for i := 0; i < len(samples); i++ {
		if samples[i].Errors > 0 && samples[i].ErrorRate > 0.01 {
			start := i
			maxErrors := samples[i].Errors
			for i < len(samples) && (samples[i].Errors > 0) {
				if samples[i].Errors > maxErrors {
					maxErrors = samples[i].Errors
				}
				i++
			}
			severity := "medium"
			if maxErrors > 100 {
				severity = "high"
			}
			anomalies = append(anomalies, report.BundleAnomalyWindow{
				Type:     "error_burst",
				StartIdx: start,
				EndIdx:   i - 1,
				Severity: severity,
				Value:    float64(maxErrors),
				Summary:  fmt.Sprintf("Error burst: %d errors detected", maxErrors),
			})
		}
	}

	// Detect CPU high pressure from monitoring
	if monitoring != nil && len(monitoring.TimeSeries) > 10 {
		for i := 0; i < len(monitoring.TimeSeries); i++ {
			if monitoring.TimeSeries[i].CPUUsage > 90 {
				start := i
				maxCPU := monitoring.TimeSeries[i].CPUUsage
				for i < len(monitoring.TimeSeries) && monitoring.TimeSeries[i].CPUUsage > 90 {
					if monitoring.TimeSeries[i].CPUUsage > maxCPU {
						maxCPU = monitoring.TimeSeries[i].CPUUsage
					}
					i++
				}
				anomalies = append(anomalies, report.BundleAnomalyWindow{
					Type:     "cpu_pressure",
					StartIdx: start,
					EndIdx:   i - 1,
					Severity: "high",
					Value:    maxCPU,
					Summary:  fmt.Sprintf("CPU usage peaked at %.1f%%", maxCPU),
				})
			}
		}
	}

	return anomalies
}

func (g *BundleGenerator) selectRawSamples(samples []execution.MetricSample, anomalies []report.BundleAnomalyWindow) []report.BundleRawSample {
	if len(samples) == 0 {
		return nil
	}

	anomalyIdxSet := make(map[int]bool)
	for _, a := range anomalies {
		for i := a.StartIdx; i <= a.EndIdx && i < len(samples); i++ {
			anomalyIdxSet[i] = true
		}
	}

	result := make([]report.BundleRawSample, 0)
	anomalyCount := 0
	normalCount := 0

	// Prioritize anomaly samples
	for idx := range anomalyIdxSet {
		if anomalyCount >= MaxRawSamplesAnomaly {
			break
		}
		if idx < len(samples) {
			s := samples[idx]
			result = append(result, report.BundleRawSample{
				Index:     idx,
				Timestamp: s.Timestamp.Unix(),
				Type:      "anomaly",
				Data: map[string]interface{}{
					"tps":         s.TPS,
					"latency_avg": s.LatencyAvg,
					"errors":      s.Errors,
				},
			})
			anomalyCount++
		}
	}

	// Add evenly spaced normal samples
	if len(samples) > 0 && normalCount < MaxRawSamplesNormal {
		step := max(1, len(samples)/MaxRawSamplesNormal)
		for i := 0; i < len(samples) && normalCount < MaxRawSamplesNormal; i += step {
			if anomalyIdxSet[i] {
				continue
			}
			s := samples[i]
			result = append(result, report.BundleRawSample{
				Index:     i,
				Timestamp: s.Timestamp.Unix(),
				Type:      "normal",
				Data: map[string]interface{}{
					"tps":         s.TPS,
					"latency_avg": s.LatencyAvg,
				},
			})
			normalCount++
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Index < result[j].Index
	})

	return result
}

func (g *BundleGenerator) buildPhaseBreakdown(run *execution.Run) report.BundlePhaseBreakdown {
	pb := report.BundlePhaseBreakdown{}
	if run.Result != nil && run.Result.Duration > 0 {
		pb.Phases = append(pb.Phases, report.BundlePhase{
			Name:       "run",
			DurationMs: int64(run.Result.Duration * 1000),
			Status:     string(run.State),
		})
	}
	return pb
}

// compressBundle compresses a bundle to gzip bytes.
func compressBundle(bundle *report.Bundle) ([]byte, error) {
	jsonData, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("marshal bundle: %w", err)
	}

	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("create gzip writer: %w", err)
	}
	if _, err := w.Write(jsonData); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressBundle decompresses gzip bytes back to a Bundle.
func DecompressBundle(data []byte) (*report.Bundle, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer r.Close()

	var bundle report.Bundle
	if err := json.NewDecoder(r).Decode(&bundle); err != nil {
		return nil, fmt.Errorf("decode bundle: %w", err)
	}
	return &bundle, nil
}

// Helper types and functions for downsampling

type tsPoint struct {
	timestamp int64
	value     float64
}

func samplesToTPSPoints(samples []execution.MetricSample) []tsPoint {
	points := make([]tsPoint, 0, len(samples))
	for _, s := range samples {
		if s.Timestamp.IsZero() {
			continue
		}
		points = append(points, tsPoint{timestamp: s.Timestamp.Unix(), value: s.TPS})
	}
	return points
}

func samplesToLatencyPoints(samples []execution.MetricSample) []tsPoint {
	points := make([]tsPoint, 0, len(samples))
	for _, s := range samples {
		if s.Timestamp.IsZero() || s.LatencyAvg <= 0 {
			continue
		}
		points = append(points, tsPoint{timestamp: s.Timestamp.Unix(), value: s.LatencyAvg})
	}
	return points
}

func monitoringToCPUPoints(ts []report.MonitoringTimeSeries) []tsPoint {
	points := make([]tsPoint, 0, len(ts))
	for _, p := range ts {
		if p.Timestamp.IsZero() {
			continue
		}
		points = append(points, tsPoint{timestamp: p.Timestamp.Unix(), value: p.CPUUsage})
	}
	return points
}

// downsample reduces a time series to a fixed number of buckets.
func downsample(points []tsPoint, bucketCount int) []report.BundleTimeseriesBucket {
	if len(points) == 0 {
		return nil
	}
	if len(points) <= bucketCount {
		// No need to downsample, but still compute stats per point
		buckets := make([]report.BundleTimeseriesBucket, len(points))
		for i, p := range points {
			buckets[i] = report.BundleTimeseriesBucket{
				Timestamp: p.timestamp,
				Avg:       p.value,
				Min:       p.value,
				Max:       p.value,
			}
		}
		return buckets
	}

	bucketSize := len(points) / bucketCount
	buckets := make([]report.BundleTimeseriesBucket, 0, bucketCount)

	for i := 0; i < bucketCount; i++ {
		start := i * bucketSize
		end := start + bucketSize
		if i == bucketCount-1 {
			end = len(points)
		}
		if start >= len(points) {
			break
		}

		vals := make([]float64, 0, end-start)
		var tsSum int64
		for _, p := range points[start:end] {
			vals = append(vals, p.value)
			tsSum += p.timestamp
		}

		avg, mx, mn := minMaxAvg(vals)
		sorted := make([]float64, len(vals))
		copy(sorted, vals)
		sort.Float64s(sorted)
		p95 := percentile(sorted, 95)

		buckets = append(buckets, report.BundleTimeseriesBucket{
			Timestamp: tsSum / int64(len(vals)),
			Avg:       avg,
			Min:       mn,
			Max:       mx,
			P95:       p95,
		})
	}

	return buckets
}

func computeWindowSummary(samples []execution.MetricSample) report.BundleWindowSummary {
	if len(samples) == 0 {
		return report.BundleWindowSummary{}
	}

	tpsVals := make([]float64, 0, len(samples))
	latVals := make([]float64, 0, len(samples))
	cpuVals := make([]float64, 0, len(samples))

	for _, s := range samples {
		tpsVals = append(tpsVals, s.TPS)
		if s.LatencyAvg > 0 {
			latVals = append(latVals, s.LatencyAvg)
		}
	}

	tpsAvg, tpsMax, tpsMin := minMaxAvg(tpsVals)
	latAvg, latMax, latMin := minMaxAvg(latVals)
	cpuAvg, cpuMax, cpuMin := minMaxAvg(cpuVals)

	return report.BundleWindowSummary{
		TPS:        report.WindowStat{Avg: tpsAvg, Min: tpsMin, Max: tpsMax},
		Latency:    report.WindowStat{Avg: latAvg, Min: latMin, Max: latMax},
		CPU:        report.WindowStat{Avg: cpuAvg, Min: cpuMin, Max: cpuMax},
		SampleCount: len(samples),
	}
}

func resourceVerdict(value, normalThreshold, elevatedThreshold, highThreshold float64) string {
	if value >= highThreshold {
		return "critical"
	}
	if value >= elevatedThreshold {
		return "high"
	}
	if value >= normalThreshold {
		return "elevated"
	}
	return "normal"
}

// BuildPreviewFromBundle creates a DetailedDataPreview from an already-generated bundle.
func (g *BundleGenerator) BuildPreviewFromBundle(bundle *report.Bundle, compressed []byte, reportID string) *report.DetailedDataPreview {
	windows := make([]report.WindowPreview, 0, len(bundle.RetainedWindows))
	for _, w := range bundle.RetainedWindows {
		windows = append(windows, report.WindowPreview{
			Name:        w.Name,
			SampleCount: w.Summary.SampleCount,
		})
	}

	anomalies := make([]report.AnomalyPreview, 0, len(bundle.AnomalyWindows))
	for _, a := range bundle.AnomalyWindows {
		anomalies = append(anomalies, report.AnomalyPreview{
			Type:     a.Type,
			Severity: a.Severity,
			Summary:  a.Summary,
			Value:    a.Value,
		})
	}

	return &report.DetailedDataPreview{
		BundleFilename: fmt.Sprintf("report_bundle_%s.json.gz", reportID),
		CompressedSize: int64(len(compressed)),
		SamplingPolicy: bundle.AIMeta.SamplingPolicy,
		RetainedWindows: windows,
		AnomalyWindows:  anomalies,
		RawSampleCount:  len(bundle.RawSamples),
	}
}

