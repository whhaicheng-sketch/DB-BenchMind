// Package report provides unit tests for chart generator.
package report

import (
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// TestChartGenerator_GenerateTPSSparkline tests TPS sparkline generation.
func TestChartGenerator_GenerateTPSSparkline(t *testing.T) {
	gen := NewChartGenerator()

	// Empty samples
	result := gen.GenerateTPSSparkline([]report.MetricSample{}, 60, 10)
	if result != "" {
		t.Errorf("Empty samples should return empty string, got: %s", result)
	}

	// With samples
	now := time.Now()
	samples := []report.MetricSample{
		{Timestamp: now, TPS: 1000, LatencyAvg: 5.0},
		{Timestamp: now.Add(time.Second), TPS: 1200, LatencyAvg: 5.5},
		{Timestamp: now.Add(2 * time.Second), TPS: 1100, LatencyAvg: 5.2},
		{Timestamp: now.Add(3 * time.Second), TPS: 1300, LatencyAvg: 5.8},
		{Timestamp: now.Add(4 * time.Second), TPS: 1250, LatencyAvg: 5.6},
	}

	result = gen.GenerateTPSSparkline(samples, 40, 5)
	if result == "" {
		t.Error("Non-empty samples should generate a chart")
	}
	if len(result) < 10 {
		t.Errorf("Chart should have some content, got length %d", len(result))
	}
}

// TestChartGenerator_GenerateLatencyDistribution tests latency distribution chart.
func TestChartGenerator_GenerateLatencyDistribution(t *testing.T) {
	gen := NewChartGenerator()

	// Empty samples
	result := gen.GenerateLatencyDistribution([]report.MetricSample{}, 60)
	if result != "" {
		t.Errorf("Empty samples should return empty string, got: %s", result)
	}

	// With samples
	now := time.Now()
	samples := []report.MetricSample{
		{Timestamp: now, LatencyAvg: 5.0},
		{Timestamp: now.Add(time.Second), LatencyAvg: 5.5},
		{Timestamp: now.Add(2 * time.Second), LatencyAvg: 5.2},
		{Timestamp: now.Add(3 * time.Second), LatencyAvg: 5.8},
		{Timestamp: now.Add(4 * time.Second), LatencyAvg: 5.6},
	}

	result = gen.GenerateLatencyDistribution(samples, 60)
	if result == "" {
		t.Error("Non-empty samples should generate a chart")
	}
	if len(result) < 10 {
		t.Errorf("Chart should have some content, got length %d", len(result))
	}
}

// TestChartGenerator_GenerateBarChart tests bar chart generation.
func TestChartGenerator_GenerateBarChart(t *testing.T) {
	gen := NewChartGenerator()

	// Mismatched lengths
	result := gen.GenerateBarChart([]string{"A", "B"}, []float64{1.0}, 40)
	if result != "" {
		t.Error("Mismatched lengths should return empty string")
	}

	// Empty data
	result = gen.GenerateBarChart([]string{}, []float64{}, 40)
	if result != "" {
		t.Error("Empty data should return empty string")
	}

	// Valid data
	labels := []string{"TPS", "QPS", "RPS"}
	values := []float64{1000.5, 5000.25, 2500.75}

	result = gen.GenerateBarChart(labels, values, 40)
	if result == "" {
		t.Error("Valid data should generate a chart")
	}
	if len(result) < 10 {
		t.Errorf("Chart should have some content, got length %d", len(result))
	}
}

// TestChartGenerator_downsample tests downsampling.
func TestChartGenerator_downsample(t *testing.T) {
	gen := &ChartGenerator{}

	// No downsampling needed
	values := []float64{1, 2, 3, 4, 5}
	result := gen.downsample(values, 10)
	if len(result) != 5 {
		t.Errorf("No downsampling needed, should return same length, got %d", len(result))
	}

	// Downsampling needed
	values = make([]float64, 100)
	for i := range values {
		values[i] = float64(i)
	}
	result = gen.downsample(values, 10)
	if len(result) != 10 {
		t.Errorf("Should downsample to 10, got %d", len(result))
	}
}

// TestChartGenerator_minMax tests min/max calculation.
func TestChartGenerator_minMax(t *testing.T) {
	gen := &ChartGenerator{}

	// Normal values
	values := []float64{1.5, 2.0, 0.5, 3.0, 2.5}
	min, max := gen.minMax(values)
	if min != 0.5 {
		t.Errorf("min = %v, want 0.5", min)
	}
	if max != 3.0 {
		t.Errorf("max = %v, want 3.0", max)
	}

	// All same values
	values = []float64{5.0, 5.0, 5.0}
	min, max = gen.minMax(values)
	if min != 5.0 {
		t.Errorf("min = %v, want 5.0", min)
	}
	if max != 5.0 {
		t.Errorf("max = %v, want 5.0", max)
	}

	// Empty values
	values = []float64{}
	min, max = gen.minMax(values)
	if min != 0 {
		t.Errorf("empty should return min=0, got %v", min)
	}
	if max != 1 {
		t.Errorf("empty should return max=1, got %v", max)
	}
}

// TestChartGenerator_createHistogram tests histogram creation.
func TestChartGenerator_createHistogram(t *testing.T) {
	gen := &ChartGenerator{}

	// Empty values
	hist := gen.createHistogram([]float64{}, 10)
	if len(hist) != 10 {
		t.Errorf("Histogram should have 10 bins, got %d", len(hist))
	}
	for _, count := range hist {
		if count != 0 {
			t.Error("Empty values should result in empty histogram")
		}
	}

	// Normal distribution
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	hist = gen.createHistogram(values, 5)
	if len(hist) != 5 {
		t.Errorf("Histogram should have 5 bins, got %d", len(hist))
	}

	total := 0
	for _, count := range hist {
		total += count
	}
	if total != 10 {
		t.Errorf("Histogram should contain all 10 values, got %d", total)
	}
}
