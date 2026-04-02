// Package usecase provides bottleneck judgment rule engine.
package usecase

import (
	"fmt"
	"math"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// Suppress math import - used by Min function
var _ = math.Pi

// BottleneckRulesEngine evaluates benchmark data to identify bottlenecks.
type BottleneckRulesEngine struct{}

// NewBottleneckRulesEngine creates a new BottleneckRulesEngine.
func NewBottleneckRulesEngine() *BottleneckRulesEngine {
	return &BottleneckRulesEngine{}
}

// BottleneckInput contains the data needed for bottleneck analysis.
type BottleneckInput struct {
	CoreKPIs           report.CoreKPIsSection
	ResourceBottleneck report.ResourceBottleneckSection
	SampleCount        int
	ErrorRate          float64
	AnomalyCount       int
}

// Judge evaluates the input and returns a BottleneckJudgment.
func (e *BottleneckRulesEngine) Judge(input *BottleneckInput) report.BottleneckJudgment {
	if input == nil || input.SampleCount < 3 {
		return report.BottleneckJudgment{
			PrimaryBottleneck: report.BottleneckUnknown,
			Confidence:        0.0,
			Evidence:          []string{"Insufficient data for analysis"},
			Recommendations:   []string{"Run a longer benchmark to gather more data"},
		}
	}

	candidates := []struct {
		bottleneck   report.BottleneckType
		confidence   float64
		evidence     []string
		recommendations []string
	}{
		e.checkCPUBound(input),
		e.checkIOBound(input),
		e.checkLockContention(input),
		e.checkNetwork(input),
		e.checkMisconfiguration(input),
	}

	// Find highest confidence
	bestIdx := -1
	bestConf := 0.0
	for i, c := range candidates {
		if c.confidence > bestConf {
			bestConf = c.confidence
			bestIdx = i
		}
	}

	if bestIdx < 0 || bestConf < 0.3 {
		return report.BottleneckJudgment{
			PrimaryBottleneck: report.BottleneckUnknown,
			Confidence:        bestConf,
			Evidence:          []string{"No strong bottleneck pattern identified"},
			Recommendations:   []string{"Review full metrics for potential issues", "Consider running with different workload parameters"},
		}
	}

	return report.BottleneckJudgment{
		PrimaryBottleneck: candidates[bestIdx].bottleneck,
		Confidence:        candidates[bestIdx].confidence,
		Evidence:          candidates[bestIdx].evidence,
		Recommendations:   candidates[bestIdx].recommendations,
	}
}

func (e *BottleneckRulesEngine) checkCPUBound(input *BottleneckInput) struct {
	bottleneck   report.BottleneckType
	confidence   float64
	evidence     []string
	recommendations []string
} {
	cpu := input.ResourceBottleneck.CPU
	var evidence []string
	confidence := 0.0

	if cpu.Peak >= 90 {
		evidence = append(evidence, fmt.Sprintf("CPU peak usage at %.1f%% (critical)", cpu.Peak))
		confidence += 0.4
	} else if cpu.Peak >= 80 {
		evidence = append(evidence, fmt.Sprintf("CPU peak usage at %.1f%% (high)", cpu.Peak))
		confidence += 0.25
	} else if cpu.Peak >= 70 {
		evidence = append(evidence, fmt.Sprintf("CPU peak usage at %.1f%% (elevated)", cpu.Peak))
		confidence += 0.1
	}

	if cpu.Avg >= 70 {
		evidence = append(evidence, fmt.Sprintf("CPU average usage at %.1f%% (sustained high)", cpu.Avg))
		confidence += 0.2
	}

	// High CPU + good throughput suggests CPU-bound
	if cpu.Peak >= 80 && input.CoreKPIs.TPS > 0 {
		evidence = append(evidence, "High CPU usage correlates with benchmark activity")
		confidence += 0.1
	}

	return struct {
		bottleneck   report.BottleneckType
		confidence   float64
		evidence     []string
		recommendations []string
	}{
		bottleneck: report.BottleneckCPU,
		confidence: math.Min(confidence, 0.95),
		evidence:   evidence,
		recommendations: []string{
			"Check if database queries can be optimized to reduce CPU usage",
			"Consider scaling up CPU resources or adding read replicas",
			"Review query execution plans for full table scans or expensive sorts",
		},
	}
}

func (e *BottleneckRulesEngine) checkIOBound(input *BottleneckInput) struct {
	bottleneck   report.BottleneckType
	confidence   float64
	evidence     []string
	recommendations []string
} {
	disk := input.ResourceBottleneck.Disk
	var evidence []string
	confidence := 0.0

	if disk.Verdict == "critical" || disk.Verdict == "high" {
		evidence = append(evidence, fmt.Sprintf("Disk utilization verdict: %s", disk.Verdict))
		confidence += 0.4
	}

	if disk.Peak > 0 {
		if disk.Peak > 80 {
			evidence = append(evidence, fmt.Sprintf("Disk util peak at %.1f", disk.Peak))
			confidence += 0.3
		}
	}

	// High P99 latency with moderate CPU often indicates IO-bound
	if input.CoreKPIs.P99LatencyMs > 100 && input.ResourceBottleneck.CPU.Peak < 70 {
		evidence = append(evidence, fmt.Sprintf("High P99 latency (%.1fms) with moderate CPU (%.1f%%) suggests IO bottleneck",
			input.CoreKPIs.P99LatencyMs, input.ResourceBottleneck.CPU.Peak))
		confidence += 0.2
	}

	// Large gap between P99 and Avg latency suggests IO wait
	if input.CoreKPIs.P99LatencyMs > 0 && input.CoreKPIs.AvgLatencyMs > 0 {
		ratio := input.CoreKPIs.P99LatencyMs / input.CoreKPIs.AvgLatencyMs
		if ratio > 5 {
			evidence = append(evidence, fmt.Sprintf("P99/Avg latency ratio: %.1fx (high tail latency)", ratio))
			confidence += 0.15
		}
	}

	return struct {
		bottleneck   report.BottleneckType
		confidence   float64
		evidence     []string
		recommendations []string
	}{
		bottleneck: report.BottleneckIO,
		confidence: math.Min(confidence, 0.95),
		evidence:   evidence,
		recommendations: []string{
			"Check disk I/O performance with iostat or similar tools",
			"Consider using faster storage (SSD/NVMe) or optimizing I/O patterns",
			"Review buffer pool size and query index usage",
			"Check for full table scans causing excessive disk reads",
		},
	}
}

func (e *BottleneckRulesEngine) checkLockContention(input *BottleneckInput) struct {
	bottleneck   report.BottleneckType
	confidence   float64
	evidence     []string
	recommendations []string
} {
	var evidence []string
	confidence := 0.0

	// High error rate often correlates with lock contention / deadlocks
	if input.ErrorRate > 0.05 {
		evidence = append(evidence, fmt.Sprintf("Error rate at %.2f%% (high)", input.ErrorRate*100))
		confidence += 0.3
	}

	// High latency with low CPU suggests waiting (likely locks)
	if input.CoreKPIs.P95LatencyMs > 50 && input.ResourceBottleneck.CPU.Peak < 50 {
		evidence = append(evidence, fmt.Sprintf("High P95 latency (%.1fms) with low CPU (%.1f%%) suggests lock contention",
			input.CoreKPIs.P95LatencyMs, input.ResourceBottleneck.CPU.Peak))
		confidence += 0.25
	}

	// Large number of TPS anomalies (sawtooth pattern)
	if input.AnomalyCount > 5 {
		evidence = append(evidence, fmt.Sprintf("%d anomalies detected (possible contention pattern)", input.AnomalyCount))
		confidence += 0.15
	}

	return struct {
		bottleneck   report.BottleneckType
		confidence   float64
		evidence     []string
		recommendations []string
	}{
		bottleneck: report.BottleneckLockContention,
		confidence: math.Min(confidence, 0.95),
		evidence:   evidence,
		recommendations: []string{
			"Check database lock statistics and deadlocks",
			"Review transaction isolation levels",
			"Consider reducing transaction scope or batch size",
			"Check for hot rows or tables causing contention",
		},
	}
}

func (e *BottleneckRulesEngine) checkNetwork(input *BottleneckInput) struct {
	bottleneck   report.BottleneckType
	confidence   float64
	evidence     []string
	recommendations []string
} {
	net := input.ResourceBottleneck.Network
	var evidence []string
	confidence := 0.0

	if net.Verdict == "critical" || net.Verdict == "high" {
		evidence = append(evidence, fmt.Sprintf("Network verdict: %s", net.Verdict))
		confidence += 0.3
	}

	if net.RxPeak > 0 || net.TxPeak > 0 {
		evidence = append(evidence, fmt.Sprintf("Network activity detected - Rx peak: %.0f, Tx peak: %.0f",
			net.RxPeak, net.TxPeak))
	}

	// High avg latency with no resource pressure suggests network latency
	if input.CoreKPIs.AvgLatencyMs > 100 &&
		input.ResourceBottleneck.CPU.Peak < 50 &&
		input.ResourceBottleneck.Disk.Verdict == "normal" {
		evidence = append(evidence, "High latency with no CPU or disk pressure suggests network latency")
		confidence += 0.2
	}

	return struct {
		bottleneck   report.BottleneckType
		confidence   float64
		evidence     []string
		recommendations []string
	}{
		bottleneck: report.BottleneckNetwork,
		confidence: math.Min(confidence, 0.95),
		evidence:   evidence,
		recommendations: []string{
			"Check network latency between client and database server",
			"Verify network configuration and bandwidth",
			"Consider reducing round trips with batched queries",
			"Check for connection pool exhaustion",
		},
	}
}

func (e *BottleneckRulesEngine) checkMisconfiguration(input *BottleneckInput) struct {
	bottleneck   report.BottleneckType
	confidence   float64
	evidence     []string
	recommendations []string
} {
	var evidence []string
	confidence := 0.0

	// Very low throughput with no resource pressure suggests misconfiguration
	if input.CoreKPIs.TPS > 0 && input.CoreKPIs.TPS < 10 &&
		input.ResourceBottleneck.CPU.Peak < 30 &&
		input.ResourceBottleneck.Disk.Verdict == "normal" {
		evidence = append(evidence, fmt.Sprintf("Very low TPS (%.1f) with no resource pressure", input.CoreKPIs.TPS))
		confidence += 0.35
	}

	// High error rate with low throughput
	if input.ErrorRate > 0.1 && input.CoreKPIs.TPS < 100 {
		evidence = append(evidence, "High error rate combined with low throughput")
		confidence += 0.2
	}

	return struct {
		bottleneck   report.BottleneckType
		confidence   float64
		evidence     []string
		recommendations []string
	}{
		bottleneck: report.BottleneckMisconfig,
		confidence: math.Min(confidence, 0.95),
		evidence:   evidence,
		recommendations: []string{
			"Review benchmark configuration parameters (threads, duration, etc.)",
			"Verify database connection settings and pool size",
			"Check if benchmark tool is correctly configured for the target database",
			"Review database server configuration (memory, buffers, connections)",
		},
	}
}

// DetermineOverallStatus determines overall health from metrics.
func DetermineOverallStatus(kpis report.CoreKPIsSection, rb report.ResourceBottleneckSection, bj report.BottleneckJudgment) report.OverallStatus {
	score := 0

	// Error rate check
	if kpis.ErrorRate > 0.05 {
		score += 3
	} else if kpis.ErrorRate > 0.01 {
		score += 1
	}

	// Latency check
	if kpis.P99LatencyMs > 500 {
		score += 2
	} else if kpis.P99LatencyMs > 200 {
		score += 1
	}

	// CPU check
	if rb.CPU.Peak > 90 {
		score += 2
	} else if rb.CPU.Peak > 75 {
		score += 1
	}

	// Bottleneck confidence check
	if bj.Confidence > 0.7 {
		score += 1
	}

	if score >= 4 {
		return report.StatusCritical
	}
	if score >= 2 {
		return report.StatusWarning
	}
	return report.StatusHealthy
}

// GenerateOneLineConclusion generates a one-line conclusion from the data.
func GenerateOneLineConclusion(status report.OverallStatus, kpis report.CoreKPIsSection, bj report.BottleneckJudgment) string {
	switch status {
	case report.StatusHealthy:
		if kpis.TPS > 0 {
			return fmt.Sprintf("Benchmark completed successfully with %.0f TPS and %.1fms avg latency. No significant bottlenecks detected.", kpis.TPS, kpis.AvgLatencyMs)
		}
		return "Benchmark completed successfully with no significant bottlenecks detected."

	case report.StatusWarning:
		if bj.PrimaryBottleneck != report.BottleneckUnknown {
			return fmt.Sprintf("Benchmark completed with warnings. Possible %s detected (confidence: %.0f%%).", bj.PrimaryBottleneck, bj.Confidence*100)
		}
		return "Benchmark completed with warnings. Some metrics are outside normal range."

	case report.StatusCritical:
		if kpis.ErrorRate > 0.05 {
			return fmt.Sprintf("Critical: High error rate (%.1f%%) detected. Benchmark results may be unreliable.", kpis.ErrorRate*100)
		}
		if bj.PrimaryBottleneck != report.BottleneckUnknown {
			return fmt.Sprintf("Critical: %s detected (confidence: %.0f%%). Performance is significantly impacted.", bj.PrimaryBottleneck, bj.Confidence*100)
		}
		return "Critical: Multiple metrics indicate severe performance issues."

	default:
		return "Benchmark status unknown."
	}
}

// GenerateRecommendations generates actionable recommendations.
func GenerateRecommendations(bj report.BottleneckJudgment) string {
	if len(bj.Recommendations) == 0 {
		return "No specific recommendations. Benchmark results look normal."
	}

	result := ""
	for i, r := range bj.Recommendations {
		result += fmt.Sprintf("%d. %s\n", i+1, r)
	}
	return result
}

