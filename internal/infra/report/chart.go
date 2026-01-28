// Package report provides chart generation utilities for reports.
// Implements: Phase 5 - Chart Generation
package report

import (
	"fmt"
	"math"
	"strings"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// ChartGenerator generates text-based charts for reports.
type ChartGenerator struct{}

// NewChartGenerator creates a new chart generator.
func NewChartGenerator() *ChartGenerator {
	return &ChartGenerator{}
}

// GenerateTPSSparkline generates a TPS sparkline chart.
func (g *ChartGenerator) GenerateTPSSparkline(samples []report.MetricSample, width, height int) string {
	if len(samples) == 0 {
		return ""
	}

	// Extract TPS values
	values := make([]float64, len(samples))
	for i, s := range samples {
		values[i] = s.TPS
	}

	return g.generateSparkline(values, width, height, "TPS")
}

// GenerateLatencyDistribution generates a latency distribution chart.
func (g *ChartGenerator) GenerateLatencyDistribution(samples []report.MetricSample, width int) string {
	if len(samples) == 0 {
		return ""
	}

	// Extract avg latency values
	values := make([]float64, len(samples))
	for i, s := range samples {
		values[i] = s.LatencyAvg
	}

	// Create histogram
	histogram := g.createHistogram(values, 20)

	// Find max for scaling
	max := 0
	for _, count := range histogram {
		if count > max {
			max = count
		}
	}

	// Build chart
	var sb strings.Builder
	barWidth := (width - 15) / 20
	if barWidth < 1 {
		barWidth = 1
	}

	for i, count := range histogram {
		label := g.formatBinLabel(i, 20, values)
		barLength := 0
		if max > 0 {
			barLength = int(float64(count) / float64(max) * float64(width-15))
		}
		bar := strings.Repeat("█", barLength)
		sb.WriteString(fmt.Sprintf("%s │%s %d\n", label, bar, count))
	}

	return sb.String()
}

// generateSparkline generates a sparkline chart for a series of values.
func (g *ChartGenerator) generateSparkline(values []float64, width, height int, label string) string {
	if len(values) == 0 {
		return ""
	}

	// Find min and max
	min, max := g.minMax(values)
	rangeVal := max - min
	if rangeVal == 0 {
		rangeVal = 1
	}

	// Downsample to fit width
	sampled := g.downsample(values, width)

	// Build sparkline
	lines := make([]string, height)
	for i := range lines {
		lines[i] = strings.Repeat(" ", width)
	}

	// Plot points
	for i, val := range sampled {
		if i >= width {
			break
		}
		normalized := (val - min) / rangeVal
		y := height - 1 - int(normalized*float64(height-1))
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}
		line := []rune(lines[y])
		line[i] = '█'
		lines[y] = string(line)
	}

	// Add y-axis labels
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n", label))
	for i, line := range lines {
		labelVal := max - (float64(i) / float64(height-1)) * (max - min)
		prefix := fmt.Sprintf("%8.2f │", labelVal)
		sb.WriteString(prefix)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// downsample reduces the number of data points to fit the width.
func (g *ChartGenerator) downsample(values []float64, width int) []float64 {
	if len(values) <= width {
		return values
	}

	step := float64(len(values)-1) / float64(width-1)
	result := make([]float64, width)

	for i := 0; i < width; i++ {
		pos := int(float64(i) * step)
		if pos >= len(values) {
			pos = len(values) - 1
		}
		result[i] = values[pos]
	}

	return result
}

// minMax finds the minimum and maximum values in a slice.
func (g *ChartGenerator) minMax(values []float64) (float64, float64) {
	min := math.Inf(1)
	max := math.Inf(-1)

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// If all values are the same or NaN
	if math.IsInf(min, 1) || math.IsInf(max, -1) {
		return 0, 1
	}

	return min, max
}

// createHistogram creates a histogram from values.
func (g *ChartGenerator) createHistogram(values []float64, bins int) []int {
	if len(values) == 0 {
		return make([]int, bins)
	}

	min, max := g.minMax(values)
	rangeVal := max - min
	if rangeVal == 0 {
		rangeVal = 1
	}

	histogram := make([]int, bins)
	for _, v := range values {
		bin := int((v - min) / rangeVal * float64(bins))
		if bin >= bins {
			bin = bins - 1
		}
		if bin < 0 {
			bin = 0
		}
		histogram[bin]++
	}

	return histogram
}

// formatBinLabel formats a histogram bin label.
func (g *ChartGenerator) formatBinLabel(index, bins int, values []float64) string {
	min, max := g.minMax(values)
	binMin := min + (max-min)/float64(bins)*float64(index)
	binMax := min + (max-min)/float64(bins)*float64(index+1)

	if index == 0 {
		return fmt.Sprintf("<%.1f", binMax)
	}
	if index == bins-1 {
		return fmt.Sprintf("≥%.1f", binMin)
	}
	return fmt.Sprintf("%.1f-%.1f", binMin, binMax)
}

// GenerateBarChart generates a simple horizontal bar chart.
func (g *ChartGenerator) GenerateBarChart(labels []string, values []float64, width int) string {
	if len(labels) != len(values) || len(labels) == 0 {
		return ""
	}

	// Find max for scaling
	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}

	// Find max label length
	maxLabelLen := 0
	for _, l := range labels {
		if len(l) > maxLabelLen {
			maxLabelLen = len(l)
		}
	}

	var sb strings.Builder
	barWidth := width - maxLabelLen - 10
	if barWidth < 10 {
		barWidth = 10
	}

	for i, label := range labels {
		value := values[i]
		barLength := int(value / max * float64(barWidth))
		bar := strings.Repeat("█", barLength)
		sb.WriteString(fmt.Sprintf("%*s │%-*s %.2f\n", maxLabelLen, label, barWidth, bar, value))
	}

	return sb.String()
}
