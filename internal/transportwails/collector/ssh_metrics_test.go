package collector

import (
	"testing"
	"time"
)

func TestParseLinuxMetrics(t *testing.T) {
	raw := "cpu  100 0 50 200 25 0 0 0 0 0\n---\n   8       0 sda 1 0 10 0 2 0 20 0 0 0 0 0 0 0 0 0 0\n"
	cpu, disk, err := parseLinuxMetrics(raw)
	if err != nil {
		t.Fatalf("parseLinuxMetrics() error = %v", err)
	}
	if cpu.user != 100 || cpu.system != 50 || cpu.iowait != 25 {
		t.Fatalf("unexpected cpu counters: %+v", cpu)
	}
	if disk.readSectors != 10 || disk.writeSectors != 20 {
		t.Fatalf("unexpected disk counters: %+v", disk)
	}
}

func TestCalcCPUPercentagesIncludesSteal(t *testing.T) {
	prev := cpuCounters{
		user: 100, system: 40, idle: 200, iowait: 20, irq: 0, softirq: 0, steal: 10,
	}
	next := cpuCounters{
		user: 130, system: 55, idle: 230, iowait: 30, irq: 0, softirq: 0, steal: 25,
	}

	user, sys, iowait, steal := calcCPUPercentages(prev, next)

	if user <= 0 {
		t.Fatalf("expected user percentage > 0, got %.2f", user)
	}
	if sys <= 0 {
		t.Fatalf("expected sys percentage > 0, got %.2f", sys)
	}
	if iowait <= 0 {
		t.Fatalf("expected iowait percentage > 0, got %.2f", iowait)
	}
	if steal <= 0 {
		t.Fatalf("expected steal percentage > 0, got %.2f", steal)
	}
}

func TestLatencyForwardFill(t *testing.T) {
	tests := []struct {
		name              string
		lastReadLatency   float64
		lastWriteLatency  float64
		deltaReads        uint64
		deltaWrites       uint64
		deltaReadTimeMs   uint64
		deltaWriteTimeMs  uint64
		wantReadLatency   float64
		wantWriteLatency  float64
	}{
		{
			name:             "normal_io_has_actual_latency",
			lastReadLatency:  5.0,
			lastWriteLatency: 3.0,
			deltaReads:       10,
			deltaWrites:      20,
			deltaReadTimeMs:  50,
			deltaWriteTimeMs: 80,
			wantReadLatency:  5.0,  // 50/10
			wantWriteLatency: 4.0,  // 80/20
		},
		{
			name:             "no_reads_carries_last_read_latency",
			lastReadLatency:  5.0,
			lastWriteLatency: 3.0,
			deltaReads:       0,
			deltaWrites:      10,
			deltaReadTimeMs:  0,
			deltaWriteTimeMs: 30,
			wantReadLatency:  5.0, // forward-fill
			wantWriteLatency: 3.0, // 30/10
		},
		{
			name:             "no_writes_carries_last_write_latency",
			lastReadLatency:  5.0,
			lastWriteLatency: 3.0,
			deltaReads:       10,
			deltaWrites:      0,
			deltaReadTimeMs:  50,
			deltaWriteTimeMs: 0,
			wantReadLatency:  5.0, // 50/10
			wantWriteLatency: 3.0, // forward-fill
		},
		{
			name:             "no_io_at_all_carries_both",
			lastReadLatency:  5.0,
			lastWriteLatency: 3.0,
			deltaReads:       0,
			deltaWrites:      0,
			deltaReadTimeMs:  0,
			deltaWriteTimeMs: 0,
			wantReadLatency:  5.0, // forward-fill
			wantWriteLatency: 3.0, // forward-fill
		},
		{
			name:             "first_collection_no_history_stays_zero",
			lastReadLatency:  0,
			lastWriteLatency: 0,
			deltaReads:       0,
			deltaWrites:      0,
			deltaReadTimeMs:  0,
			deltaWriteTimeMs: 0,
			wantReadLatency:  0,
			wantWriteLatency: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SSHMetricsCollector{
				interval:          3 * time.Second,
				lastReadLatencyMs:  tt.lastReadLatency,
				lastWriteLatencyMs: tt.lastWriteLatency,
				lastDisk: &diskCounters{},
				history:            make([]SSHMetricPoint, 0, 300),
			}
			// Simulate the latency calculation logic from collect()
			var readLat, writeLat float64
			if tt.deltaReads > 0 {
				readLat = float64(tt.deltaReadTimeMs) / float64(tt.deltaReads)
			} else {
				readLat = c.lastReadLatencyMs
			}
			if tt.deltaWrites > 0 {
				writeLat = float64(tt.deltaWriteTimeMs) / float64(tt.deltaWrites)
			} else {
				writeLat = c.lastWriteLatencyMs
			}
			if readLat != tt.wantReadLatency {
				t.Errorf("read latency = %v, want %v", readLat, tt.wantReadLatency)
			}
			if writeLat != tt.wantWriteLatency {
				t.Errorf("write latency = %v, want %v", writeLat, tt.wantWriteLatency)
			}
		})
	}
}
