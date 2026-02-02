// Package sysbench provides sysbench log parsing functionality.
// This file implements parsers for sysbench raw output including
// per-second time series data and summary statistics.
package sysbench

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ParsedRun represents a fully parsed sysbench run.
type ParsedRun struct {
	RunID       string
	Timestamp   time.Time

	// Raw output
	RawOutput   string

	// Summary statistics
	TPS         float64
	QPS         float64
	SQLStats    SQLStatistics
	Latency     LatencyStats
	Reliability ReliabilityMetrics

	// Time series data (per-second samples)
	TimeSeries  []TimeSeriesSample

	// Metadata
	Threads     int
	Duration    float64 // seconds
}

// TimeSeriesSample represents per-second metrics.
type TimeSeriesSample struct {
	Second      int
	TPS         float64
	QPS         float64
	LatencyP95  float64
	ErrorRate   float64
}

// SQLStatistics contains SQL statistics from summary.
type SQLStatistics struct {
	ReadQueries    int64
	WriteQueries   int64
	OtherQueries   int64
	TotalQueries   int64
	TotalTransactions int64
}

// LatencyStats contains latency statistics.
type LatencyStats struct {
	Avg  float64
	Min  float64
	Max  float64
	P95  float64
	P99  float64
}

// ReliabilityMetrics contains reliability metrics.
type ReliabilityMetrics struct {
	Errors      int64
	Reconnects  int64
}

// ParseSysbenchOutput parses raw sysbench output.
func ParseSysbenchOutput(runID string, rawOutput string) (*ParsedRun, error) {
	run := &ParsedRun{
		RunID:     runID,
		Timestamp: time.Now(),
		RawOutput: rawOutput,
	}

	// Parse per-second time series
	run.TimeSeries = extractTimeSeries(rawOutput)

	// Parse summary statistics
	parseSummaryStatistics(rawOutput, run)

	return run, nil
}

// extractTimeSeries extracts per-second time series data.
// Input format:
// [ 1s ] thds: 1 tps: 74.14 qps: 1482.87 (r/w/o: 1043.00/439.00/0.87) lat (ms,99%): 12.00
func extractTimeSeries(rawOutput string) []TimeSeriesSample {
	var samples []TimeSeriesSample

	// Regex pattern for per-second lines
	// Pattern: [ Ns ] thds: X tps: Y.YY qps: Z.ZZ (r/w/o: ...) lat (ms,99%): LL.LL
	pattern := regexp.MustCompile(`\[\s*(\d+)s\s*\]\s*thds:\s*(\d+)\s*tps:\s*(\d+\.\d+)\s*qps:\s*(\d+\.\d+)`)

	lines := strings.Split(rawOutput, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "[") || !strings.Contains(line, "]") {
			continue
		}

		matches := pattern.FindStringSubmatch(line)
		if len(matches) < 4 {
			continue
		}

		second, _ := strconv.Atoi(matches[1])
		tps, _ := strconv.ParseFloat(matches[3], 64)
		qps, _ := strconv.ParseFloat(matches[4], 64)

		sample := TimeSeriesSample{
			Second:    second,
			TPS:       tps,
			QPS:       qps,
		}

		samples = append(samples, sample)
	}

	return samples
}

// parseSummaryStatistics parses the summary section from sysbench output.
func parseSummaryStatistics(rawOutput string, run *ParsedRun) {
	// Parse TPS
	if tps := extractMetric(rawOutput, `transactions:\s*(\d+)\s*\((\d+\.\d+)\s*per sec\.\)`); tps > 0 {
		run.TPS = tps
	}

	// Parse QPS
	if qps := extractMetric(rawOutput, `queries:\s*(\d+)\s*\((\d+\.\d+)\s*per sec\.\)`); qps > 0 {
		run.QPS = qps
	}

	// Parse SQL statistics
	run.SQLStats = extractSQLStatistics(rawOutput)

	// Parse latency
	run.Latency = extractLatency(rawOutput)

	// Parse reliability
	run.Reliability = extractReliability(rawOutput)

	// Parse threads
	if threads := extractMetric(rawOutput, `Number of threads:\s*(\d+)`); threads > 0 {
		run.Threads = int(threads)
	}

	// Parse duration
	if duration := extractMetric(rawOutput, `total time:\s*(\d+\.\d+)s`); duration > 0 {
		run.Duration = duration
	}
}

// extractMetric extracts a numeric value using regex.
func extractMetric(rawOutput, pattern string) float64 {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(rawOutput)
	if len(matches) < 2 {
		return 0
	}

	val, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	return val
}

// extractSQLStatistics extracts SQL statistics from sysbench output.
func extractSQLStatistics(rawOutput string) SQLStatistics {
	stats := SQLStatistics{}

	// Try to find SQL statistics section
	lines := strings.Split(rawOutput, "\n")
	inSQLSection := false

	for _, line := range lines {
		if strings.Contains(line, "SQL statistics:") {
			inSQLSection = true
			continue
		}

		if inSQLSection {
			// Parse individual lines
			if strings.Contains(line, "queries performed:") {
				inSQLSection = false
				break
			}

			// read:  3136
			if matches := regexp.MustCompile(`read:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				stats.ReadQueries, _ = strconv.ParseInt(matches[1], 10, 64)
			}
			// write:  896
			if matches := regexp.MustCompile(`write:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				stats.WriteQueries, _ = strconv.ParseInt(matches[1], 10, 64)
			}
			// other:  448
			if matches := regexp.MustCompile(`other:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				stats.OtherQueries, _ = strconv.ParseInt(matches[1], 10, 64)
			}
			// total:  4480
			if matches := regexp.MustCompile(`total:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				stats.TotalQueries, _ = strconv.ParseInt(matches[1], 10, 64)
			}
		}
	}

	// transactions count
	if matches := regexp.MustCompile(`transactions:\s*(\d+)\s*\(`).FindStringSubmatch(rawOutput); len(matches) > 1 {
		stats.TotalTransactions, _ = strconv.ParseInt(matches[1], 10, 64)
	}

	return stats
}

// extractLatency extracts latency statistics.
func extractLatency(rawOutput string) LatencyStats {
	stats := LatencyStats{}

	// avg: 13.39
	if val := extractMetric(rawOutput, `avg=\s*(\d+\.\d+)`); val > 0 {
		stats.Avg = val
	}

	// min: 6.06
	if val := extractMetric(rawOutput, `min=\s*(\d+\.\d+)`); val > 0 {
		stats.Min = val
	}

	// max: 48.64
	if val := extractMetric(rawOutput, `max=\s*(\d+\.\d+)`); val > 0 {
		stats.Max = val
	}

	// 95th percentile:  28.67
	if val := extractMetric(rawOutput, `95th\s*percentile:\s*(\d+\.\d+)`); val > 0 {
		stats.P95 = val
	}

	// 99th percentile: 45.23
	if val := extractMetric(rawOutput, `99th\s*percentile:\s*(\d+\.\d+)`); val > 0 {
		stats.P99 = val
	}

	return stats
}

// extractReliability extracts errors and reconnects.
func extractReliability(rawOutput string) ReliabilityMetrics {
	metrics := ReliabilityMetrics{}

	// Errors: total: 0
	if matches := regexp.MustCompile(`Errors:\s*total:\s*(\d+)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
		metrics.Errors, _ = strconv.ParseInt(matches[1], 10, 64)
	}

	// Reconnects: total: 0
	if matches := regexp.MustCompile(`reconnects:\s*total:\s*(\d+)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
		metrics.Reconnects, _ = strconv.ParseInt(matches[1], 10, 64)
	}

	return metrics
}

// CalculateVariantID computes variant ID from template parameters.
func CalculateVariantID(params map[string]string) string {
	if len(params) == 0 {
		return "default"
	}

	// Sort keys for canonical representation
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build canonical string: key1=value1&key2=value2&...
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	canonical := strings.Join(pairs, "&")

	// Compute SHA-1 hash
	h := fnv.New32a()
	h.Write([]byte(canonical))
	hash := hex.EncodeToString(h.Sum(nil))

	// Return first 12 characters
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

// GetSteadyStateTimeSeries extracts steady-state time series (discard first N seconds).
// Returns samples with second >= steadyStart.
func GetSteadyStateTimeSeries(samples []TimeSeriesSample, steadyStart int) []TimeSeriesSample {
	var steady []TimeSeriesSample
	for _, s := range samples {
		if s.Second >= steadyStart {
			steady = append(steady, s)
		}
	}
	return steady
}

// CalculateMean calculates mean of values.
func CalculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// CalculateStdDev calculates standard deviation.
func CalculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	var varianceSum float64
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}
	return varianceSum / float64(len(values)-1)
}

// CalculateCV calculates coefficient of variation.
func CalculateCV(stddev, mean float64) float64 {
	if mean == 0 {
		return 0
	}
	return (stddev / mean) * 100
}
